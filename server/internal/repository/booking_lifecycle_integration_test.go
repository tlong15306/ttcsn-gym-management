package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/report"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
)

// Session timestamps are moved in the isolated test DB to exercise the clock
// transitions without waiting for the real session to finish.
func TestSyntheticBookingLifecycleThroughReport(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	id := func(query string, args ...any) int64 {
		t.Helper()
		var value int64
		if err := a.pool.QueryRow(ctx, query, args...).Scan(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	id(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'LIFECYCLE-M',current_date,current_date+365) RETURNING id`, a.users["MEMBER"])
	pt := id(`INSERT INTO personal_trainers(user_id,approval_status,decided_at,decided_by_user_id) VALUES($1,'APPROVED',clock_timestamp(),$2) RETURNING id`, a.users["PT"], a.users["ADMIN"])
	room := id(`INSERT INTO rooms(name,room_type,default_capacity) VALUES('Lifecycle room','GYM',5) RETURNING id`)
	slot := id(`INSERT INTO time_slots(name,start_time,end_time) VALUES('Lifecycle slot','09:00','10:00') RETURNING id`)
	session := id(`INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date+2,((current_date+2)+time '09:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',((current_date+2)+time '10:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',5,$3) RETURNING id`, room, slot, a.users["ADMIN"])
	a.sql(`INSERT INTO pt_working_hours(pt_id,day_of_week,start_time,end_time) VALUES($1,extract(isodow from current_date+2),'08:00','11:00')`, pt)
	repo := repository.NewBookingRepository(a.pool)
	created, err := repo.Create(ctx, identity.User{ID: a.users["MEMBER"], Role: identity.RoleMember}, 0, booking.CreateInput{RoomSessionID: session, PTID: pt}, "lifecycle-demo-key", "lifecycle-test")
	if err != nil || created.Status != "PENDING_PT" {
		t.Fatalf("booking create: %+v: %v", created, err)
	}
	a.call("POST", fmt.Sprintf("/bookings/%d/pt-response", created.ID), "PT", map[string]string{"action": "ACCEPT"}, 200)
	a.sql(`UPDATE room_sessions SET starts_at=clock_timestamp()-interval '10 minutes',ends_at=clock_timestamp()+interval '30 minutes' WHERE id=$1`, session)
	attendancePath := fmt.Sprintf("/bookings/%d/attendance", created.ID)
	a.call("POST", attendancePath, "RECEPTIONIST", map[string]string{"participant_type": "MEMBER"}, 201)
	a.call("POST", attendancePath, "RECEPTIONIST", map[string]string{"participant_type": "PT"}, 201)
	a.sql(`UPDATE room_sessions SET ends_at=clock_timestamp()-interval '1 second' WHERE id=$1`, session)
	if err := repo.Complete(ctx); err != nil {
		t.Fatal(err)
	}
	var status, sessionDate string
	if err := a.pool.QueryRow(ctx, `SELECT b.status,s.session_date::text FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id WHERE b.id=$1`, created.ID).Scan(&status, &sessionDate); err != nil || status != "COMPLETED" {
		t.Fatalf("completion status %s: %v", status, err)
	}
	a.call("POST", fmt.Sprintf("/bookings/%d/review", created.ID), "MEMBER", map[string]any{"rating": 5, "comment": "Synthetic lifecycle"}, 201)
	response := a.call("GET", fmt.Sprintf("/reports/operations?from=%s&to=%s&room_id=%d", sessionDate, sessionDate, room), "MANAGER", nil, 200)
	var result report.Operations
	decodeData(t, response, &result)
	if result.TotalBookings != 1 || result.StatusCounts["COMPLETED"] != 1 || result.MemberAttendanceCount != 1 || result.PTAttendanceCount != 1 {
		t.Fatalf("lifecycle report mismatch: %+v", result)
	}
	trainerResponse := a.call("GET", fmt.Sprintf("/reports/pts/%d?from=%s&to=%s", pt, sessionDate, sessionDate), "PT", nil, 200)
	var trainerResult report.Trainer
	decodeData(t, trainerResponse, &trainerResult)
	if trainerResult.ReviewCount != 1 || trainerResult.RatingAverage == nil || *trainerResult.RatingAverage != 5 {
		t.Fatalf("trainer report mismatch: %+v", trainerResult)
	}
}
