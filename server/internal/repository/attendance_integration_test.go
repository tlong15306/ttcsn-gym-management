package repository_test

import (
	"context"
	"fmt"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/attendance"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	"sync"
	"testing"
)

func TestAttendanceAndCompletionIntegration(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	id := func(q string, args ...any) int64 {
		var v int64
		if err := a.pool.QueryRow(ctx, q, args...).Scan(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	member := id(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'ATT-M',current_date,current_date+365) RETURNING id`, a.users["MEMBER"])
	pt := id(`INSERT INTO personal_trainers(user_id,approval_status,decided_at,decided_by_user_id) VALUES($1,'APPROVED',now(),$2) RETURNING id`, a.users["PT"], a.users["ADMIN"])
	room := id(`INSERT INTO rooms(name,room_type,default_capacity) VALUES('Attendance test','GYM',10) RETURNING id`)
	slot := id(`INSERT INTO time_slots(name,start_time,end_time) VALUES('Attendance test','00:00','24:00') RETURNING id`)
	session := id(`INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,(now() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date,now()-interval '10 minutes',now()+interval '50 minutes',10,$3) RETURNING id`, room, slot, a.users["ADMIN"])
	makeBooking := func(withPT bool) int64 {
		var p any
		if withPT {
			p = pt
		}
		return id(`INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,pt_id,status,created_by_user_id) SELECT $1,id,session_date,time_slot_id,$2,'CONFIRMED',$3 FROM room_sessions WHERE id=$4 RETURNING bookings.id`, member, p, a.users["ADMIN"], session)
	}
	b := makeBooking(true)
	path := fmt.Sprintf("/bookings/%d/attendance", b)
	input := map[string]string{"participant_type": "MEMBER"}
	for _, role := range []string{"MEMBER", "PT"} {
		a.call("POST", path, role, input, 403)
		a.call("GET", path, role, nil, 200)
	}
	a.call("POST", path, "RECEPTIONIST", map[string]any{"participant_type": "MEMBER", "participant_user_id": 999}, 422)
	a.sql(`CREATE FUNCTION fail_attendance_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.action='attendance.RECORDED' THEN RAISE EXCEPTION 'test audit failure'; END IF; RETURN NEW; END $$`)
	a.sql(`CREATE TRIGGER fail_attendance_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION fail_attendance_audit()`)
	a.call("POST", path, "RECEPTIONIST", input, 500)
	var rollbackCount int
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM booking_attendance WHERE booking_id=$1`, b).Scan(&rollbackCount); err != nil || rollbackCount != 0 {
		t.Fatal("attendance audit rollback", err, rollbackCount)
	}
	a.sql(`DROP TRIGGER fail_attendance_audit ON audit_events`)
	a.sql(`DROP FUNCTION fail_attendance_audit()`)
	a.call("POST", path, "RECEPTIONIST", input, 201)
	a.call("POST", path, "MANAGER", input, 200)
	a.call("POST", fmt.Sprintf("/bookings/%d/cancel", b), "ADMIN", map[string]string{"reason": "cancel"}, 409)
	repo := repository.NewAttendanceRepository(a.pool)
	operator := identity.User{ID: a.users["RECEPTIONIST"], Role: identity.RoleReceptionist}
	var wg sync.WaitGroup
	out := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.Record(ctx, operator, b, attendance.Input{ParticipantType: "PT"}, "race")
			out <- err
		}()
	}
	wg.Wait()
	close(out)
	for err := range out {
		if err != nil {
			t.Fatal(err)
		}
	}
	var n int
	a.pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action='attendance.RECORDED' AND entity_id=$1`, fmt.Sprint(b)).Scan(&n)
	if n != 2 {
		t.Fatal("duplicate attendance audit", n)
	}
	a.sql(`UPDATE room_sessions SET ends_at=now()-interval '1 second' WHERE id=$1`, session)
	completion := repository.NewBookingRepository(a.pool)
	a.sql(`CREATE FUNCTION fail_completion_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.action='booking.COMPLETED' THEN RAISE EXCEPTION 'test audit failure'; END IF; RETURN NEW; END $$`)
	a.sql(`CREATE TRIGGER fail_completion_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION fail_completion_audit()`)
	if err := completion.Complete(ctx); err == nil {
		t.Fatal("expected completion audit failure")
	}
	var rollbackStatus string
	if err := a.pool.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1`, b).Scan(&rollbackStatus); err != nil || rollbackStatus != "CONFIRMED" {
		t.Fatal("completion audit rollback", err, rollbackStatus)
	}
	a.sql(`DROP TRIGGER fail_completion_audit ON audit_events`)
	a.sql(`DROP FUNCTION fail_completion_audit()`)
	if err := completion.Complete(ctx); err != nil {
		t.Fatal(err)
	}
	if err := completion.Complete(ctx); err != nil {
		t.Fatal(err)
	}
	var status string
	a.pool.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1`, b).Scan(&status)
	if status != "COMPLETED" {
		t.Fatal(status)
	}
	// Completed replay preserves the original record; no post-session additions.
	a.call("POST", path, "ADMIN", input, 200)
	a.sql(`UPDATE room_sessions SET starts_at=now()+interval '1 hour',ends_at=now()+interval '2 hours' WHERE id=$1`, session)
	b2 := makeBooking(false)
	path2 := fmt.Sprintf("/bookings/%d/attendance", b2)
	a.call("GET", path2, "PT", nil, 404)
	a.call("POST", path2, "ADMIN", map[string]string{"participant_type": "PT"}, 422)
	a.call("POST", path2, "ADMIN", input, 422)
	a.call("POST", path2, "ADMIN", map[string]string{"participant_type": "MEMBER", "override_reason": "Synthetic early arrival"}, 201)
	a.sql(`UPDATE bookings SET status='COMPLETED' WHERE id=$1`, b2)
	b3 := makeBooking(false)
	a.sql(`UPDATE room_sessions SET starts_at=now()-interval '1 hour',ends_at=now()-interval '1 second' WHERE id=$1`, session)
	a.call("POST", fmt.Sprintf("/bookings/%d/attendance", b3), "ADMIN", input, 409)
	a.call("POST", fmt.Sprintf("/bookings/%d/cancel", b3), "ADMIN", map[string]string{"reason": "late"}, 409)
	if err := completion.Complete(ctx); err != nil {
		t.Fatal(err)
	}
	a.pool.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1`, b3).Scan(&status)
	if status != "NO_SHOW" {
		t.Fatal(status)
	}
	// PT presence alone must never complete the member's booking.
	a.sql(`UPDATE room_sessions SET starts_at=now()-interval '10 minutes',ends_at=now()+interval '50 minutes' WHERE id=$1`, session)
	b4 := makeBooking(true)
	a.call("POST", fmt.Sprintf("/bookings/%d/attendance", b4), "MANAGER", map[string]string{"participant_type": "PT"}, 201)
	a.sql(`UPDATE room_sessions SET ends_at=now()-interval '1 second' WHERE id=$1`, session)
	if err := completion.Complete(ctx); err != nil {
		t.Fatal(err)
	}
	a.pool.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1`, b4).Scan(&status)
	if status != "NO_SHOW" {
		t.Fatal("PT-only result", status)
	}
	// Concurrent operator cancellation/check-in serialize to exactly one outcome.
	a.sql(`UPDATE room_sessions SET starts_at=now()-interval '10 minutes',ends_at=now()+interval '50 minutes' WHERE id=$1`, session)
	b5 := makeBooking(false)
	raceErrs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := repo.Record(ctx, operator, b5, attendance.Input{ParticipantType: "MEMBER"}, "checkin-race")
		raceErrs <- err
	}()
	go func() {
		defer wg.Done()
		_, err := completion.Act(ctx, operator, b5, booking.ActionInput{Reason: "Synthetic race"}, true, "cancel-race")
		raceErrs <- err
	}()
	wg.Wait()
	close(raceErrs)
	success := 0
	for err := range raceErrs {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatal("race winners", success)
	}
	a.pool.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1`, b5).Scan(&status)
	a.pool.QueryRow(ctx, `SELECT count(*) FROM booking_attendance WHERE booking_id=$1`, b5).Scan(&n)
	if status == "CANCELLED" && n != 0 || status == "CONFIRMED" && n != 1 || status != "CANCELLED" && status != "CONFIRMED" {
		t.Fatal("invalid race outcome", status, n)
	}
	a.sql(`UPDATE users SET phone='0350012345' WHERE id=$1`, a.users["MEMBER"])
	var found []booking.Booking
	decodeData(t, a.call("GET", "/bookings?q=0350012345", "ADMIN", nil, 200), &found)
	if len(found) == 0 {
		t.Fatal("operator phone search")
	}
	decodeData(t, a.call("GET", "/bookings?q=0350012345", "MEMBER", nil, 200), &found)
	if len(found) != 0 {
		t.Fatal("member must not search phone")
	}
	decodeData(t, a.call("GET", "/bookings?q=ATT-M&history=true", "MEMBER", nil, 200), &found)
	if len(found) == 0 {
		t.Fatal("member code history search")
	}
	// A stale CONFIRMED state after end cannot accept check-in, even concurrently
	// with the completion worker. Both outcomes converge to NO_SHOW without rows.
	a.sql(`UPDATE bookings SET status='COMPLETED' WHERE id=$1 AND status='CONFIRMED'`, b5)
	b6 := makeBooking(false)
	a.sql(`UPDATE room_sessions SET ends_at=now()-interval '1 second' WHERE id=$1`, session)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = repo.Record(ctx, operator, b6, attendance.Input{ParticipantType: "MEMBER"}, "completion-checkin-race")
	}()
	var completionErr error
	go func() { defer wg.Done(); completionErr = completion.Complete(ctx) }()
	wg.Wait()
	if completionErr != nil {
		t.Fatal(completionErr)
	}
	a.pool.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1`, b6).Scan(&status)
	a.pool.QueryRow(ctx, `SELECT count(*) FROM booking_attendance WHERE booking_id=$1`, b6).Scan(&n)
	if status != "NO_SHOW" || n != 0 {
		t.Fatal("completion/checkin outcome", status, n)
	}
}
