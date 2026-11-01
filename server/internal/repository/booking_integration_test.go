package repository_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	"image"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBookingAndCoverIntegration(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	var member, pt, room, slot, session int64
	mustID := func(q string, args ...any) int64 {
		var id int64
		if err := a.pool.QueryRow(ctx, q, args...).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	member = mustID(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'BOOK-M',current_date,current_date+365) RETURNING id`, a.users["MEMBER"])
	pt = mustID(`INSERT INTO personal_trainers(user_id,approval_status,decided_at,decided_by_user_id) VALUES($1,'APPROVED',now(),$2) RETURNING id`, a.users["PT"], a.users["ADMIN"])
	room = mustID(`INSERT INTO rooms(name,room_type,default_capacity) VALUES('Booking room','GYM',1) RETURNING id`)
	slot = mustID(`INSERT INTO time_slots(name,start_time,end_time) VALUES('Booking slot','09:00','10:00') RETURNING id`)
	session = mustID(`INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date+2,(current_date+2+time '09:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',(current_date+2+time '10:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',1,$3) RETURNING id`, room, slot, a.users["ADMIN"])
	a.sql(`INSERT INTO pt_working_hours(pt_id,day_of_week,start_time,end_time) VALUES($1,extract(isodow from current_date+2),'07:00','21:00')`, pt)
	var date string
	a.pool.QueryRow(ctx, `SELECT session_date::text FROM room_sessions WHERE id=$1`, session).Scan(&date)
	a.call("GET", fmt.Sprintf("/booking-options/sessions?room_id=%d&date=%s", room, date), "MEMBER", nil, 200)
	a.call("GET", fmt.Sprintf("/booking-options/pts?room_session_id=%d", session), "MEMBER", nil, 200)
	a.call("GET", fmt.Sprintf("/booking-options/pts?room_session_id=%d", session), "PT", nil, 403)
	create := func(role, key string, ptid int64) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(booking.CreateInput{RoomSessionID: session, PTID: ptid})
		req := httptest.NewRequest("POST", "/api/v1/bookings", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer "+a.tokens[role])
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		w := httptest.NewRecorder()
		a.router.ServeHTTP(w, req)
		return w
	}
	if w := create("PT", "forbidden-key", 0); w.Code != 403 {
		t.Fatalf("PT create: %d", w.Code)
	}
	w := create("MEMBER", "booking-first", pt)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var b booking.Booking
	decodeData(t, w, &b)
	if b.Status != "PENDING_PT" || b.RespondBy == nil || b.ConfirmedAt != nil {
		t.Fatal("pending deadline missing")
	}
	again := create("MEMBER", "booking-first", pt)
	var replay booking.Booking
	decodeData(t, again, &replay)
	if again.Code != 201 || replay.ID != b.ID {
		t.Fatal("replay duplicated booking")
	}
	if w := create("MEMBER", "booking-first", 0); w.Code != 409 {
		t.Fatal("idempotency conflict", w.Code, w.Body.String())
	}
	if w := create("MEMBER", "booking-second", 0); w.Code != 409 {
		t.Fatal("capacity conflict", w.Code, w.Body.String())
	}
	// Retention is 24 hours: once the committed key expires, the same request
	// must be processed anew rather than replaying the previous 201 response.
	// This session is full, so a fresh attempt must fail with CAPACITY_FULL.
	a.sql(`UPDATE idempotency_keys SET expires_at=clock_timestamp()-interval '1 second' WHERE actor_user_id=$1 AND request_scope='booking:create' AND idempotency_key='booking-first'`, a.users["MEMBER"])
	if w := create("MEMBER", "booking-first", pt); w.Code != 409 || !strings.Contains(w.Body.String(), "CAPACITY_FULL") {
		t.Fatal("expired booking idempotency key replayed or bypassed capacity", w.Code, w.Body.String())
	}
	a.call("GET", fmt.Sprintf("/bookings/%d", b.ID), "PT", nil, 200)
	a.call("POST", fmt.Sprintf("/bookings/%d/pt-response", b.ID), "MEMBER", map[string]string{"action": "ACCEPT"}, 403)
	accepted := a.call("POST", fmt.Sprintf("/bookings/%d/pt-response", b.ID), "PT", map[string]string{"action": "ACCEPT"}, 200)
	var acceptedBooking booking.Booking
	decodeData(t, accepted, &acceptedBooking)
	if acceptedBooking.ConfirmedAt == nil {
		t.Fatal("PT acceptance did not record confirmation time")
	}
	a.call("POST", fmt.Sprintf("/bookings/%d/pt-response", b.ID), "PT", map[string]string{"action": "ACCEPT"}, 409)
	memberCancelled := a.call("POST", fmt.Sprintf("/bookings/%d/cancel", b.ID), "MEMBER", map[string]string{}, 200)
	var memberCancelledBooking booking.Booking
	decodeData(t, memberCancelled, &memberCancelledBooking)
	if memberCancelledBooking.CancellationSource == nil || *memberCancelledBooking.CancellationSource != booking.CancellationSourceMember || memberCancelledBooking.CancelledActorRole == nil || *memberCancelledBooking.CancelledActorRole != identity.RoleMember {
		t.Fatal("member cancellation provenance was not captured")
	}

	selfTraining := create("MEMBER", "booking-self-training", 0)
	var selfTrainingBooking booking.Booking
	decodeData(t, selfTraining, &selfTrainingBooking)
	if selfTraining.Code != 201 || selfTrainingBooking.ConfirmedAt == nil {
		t.Fatal("self-training confirmation time missing", selfTraining.Code)
	}
	operatorCancelled := a.call("POST", fmt.Sprintf("/bookings/%d/cancel", selfTrainingBooking.ID), "ADMIN", map[string]string{"reason": "Facility schedule change"}, 200)
	var operatorCancelledBooking booking.Booking
	decodeData(t, operatorCancelled, &operatorCancelledBooking)
	if operatorCancelledBooking.CancellationSource == nil || *operatorCancelledBooking.CancellationSource != booking.CancellationSourceFacility || operatorCancelledBooking.CancelledActorRole == nil || *operatorCancelledBooking.CancelledActorRole != identity.RoleAdmin {
		t.Fatal("facility cancellation provenance was not captured")
	}

	ptCancelled := create("MEMBER", "booking-pt-cancel", pt)
	var ptCancelledBooking booking.Booking
	decodeData(t, ptCancelled, &ptCancelledBooking)
	if ptCancelled.Code != 201 || ptCancelledBooking.ConfirmedAt != nil {
		t.Fatal("pending PT booking must not be marked confirmed", ptCancelled.Code)
	}
	ptAccepted := a.call("POST", fmt.Sprintf("/bookings/%d/pt-response", ptCancelledBooking.ID), "PT", map[string]string{"action": "ACCEPT"}, 200)
	decodeData(t, ptAccepted, &ptCancelledBooking)
	ptCancellation := a.call("POST", fmt.Sprintf("/bookings/%d/cancel", ptCancelledBooking.ID), "PT", map[string]string{"reason": "Coach schedule change"}, 200)
	decodeData(t, ptCancellation, &ptCancelledBooking)
	if ptCancelledBooking.CancellationSource == nil || *ptCancelledBooking.CancellationSource != booking.CancellationSourcePT || ptCancelledBooking.CancelledActorRole == nil || *ptCancelledBooking.CancelledActorRole != identity.RolePT {
		t.Fatal("PT cancellation provenance was not captured")
	}

	w = create("MEMBER", "booking-expire", pt)
	decodeData(t, w, &b)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	a.sql(`UPDATE bookings SET respond_by=now()-interval '1 minute' WHERE id=$1`, b.ID)
	repo := repository.NewBookingRepository(a.pool)
	if err := repo.Expire(ctx); err != nil {
		t.Fatal(err)
	}
	if err := repo.Expire(ctx); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action='booking.EXPIRED' AND entity_id=$1`, fmt.Sprint(b.ID)).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatal("expiry audit not idempotent", n)
	}
	expired, err := repo.Get(ctx, identity.User{ID: a.users["MEMBER"], Role: identity.RoleMember}, b.ID)
	if err != nil || expired.ConfirmedAt != nil {
		t.Fatal("expired PT request must not have a confirmation timestamp", err)
	}
	a.call("POST", fmt.Sprintf("/bookings/%d/pt-response", b.ID), "PT", map[string]string{"action": "ACCEPT"}, 409)
	// Concurrent identical requests must return one record, one hold and one audit.
	var wg sync.WaitGroup
	out := make(chan *httptest.ResponseRecorder, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); out <- create("MEMBER", "booking-race", 0) }()
	}
	wg.Wait()
	close(out)
	var shared int64
	for w := range out {
		if w.Code != 201 {
			t.Fatal(w.Code, w.Body.String())
		}
		decodeData(t, w, &b)
		if shared == 0 {
			shared = b.ID
		}
		if b.ID != shared {
			t.Fatal("duplicate race")
		}
	}
	a.call("GET", "/bookings?status=CONFIRMED&from="+date+"&to="+date, "MEMBER", nil, 200)
	a.call("GET", fmt.Sprintf("/bookings/%d", shared), "PT", nil, 404)
	// An audit failure must roll back cancellation and retain its capacity hold.
	a.sql(`CREATE FUNCTION fail_booking_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.action='booking.CANCELLED' THEN RAISE EXCEPTION 'test audit failure'; END IF; RETURN NEW; END $$`)
	a.sql(`CREATE TRIGGER fail_booking_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION fail_booking_audit()`)
	a.call("POST", fmt.Sprintf("/bookings/%d/cancel", shared), "MEMBER", map[string]string{}, 500)
	a.sql(`DROP TRIGGER fail_booking_audit ON audit_events`)
	a.sql(`DROP FUNCTION fail_booking_audit()`)
	got, err := repo.Get(ctx, identity.User{ID: a.users["MEMBER"], Role: identity.RoleMember}, shared)
	if err != nil || got.Status != "CONFIRMED" || got.CancellationSource != nil || got.CancelledActorRole != nil {
		t.Fatal("audit rollback failed", err)
	}
	_ = member
	a.call("POST", fmt.Sprintf("/bookings/%d/cancel", shared), "MEMBER", map[string]string{}, 200)
	// Twenty distinct members compete for the final seat, not just the same retry key.
	actors := make([]identity.User, 20)
	for i := range actors {
		uid := mustID(`INSERT INTO users(username,password_hash,email,full_name,role_id) SELECT $1,'test-unused-hash',$2,'Race member',id FROM roles WHERE code='MEMBER' RETURNING users.id`, fmt.Sprintf("race%d", i), fmt.Sprintf("race%d@test.local", i))
		mustID(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,$2,current_date,current_date+365) RETURNING id`, uid, fmt.Sprintf("RACE-%d", i))
		actors[i] = identity.User{ID: uid, Role: identity.RoleMember}
	}
	results := make(chan error, 20)
	for _, actor := range actors {
		wg.Add(1)
		go func(actor identity.User) {
			defer wg.Done()
			_, err := repo.Create(ctx, actor, 0, booking.CreateInput{RoomSessionID: session}, fmt.Sprintf("race-seat-%d", actor.ID), "seat-race")
			results <- err
		}(actor)
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else {
			var conflict *booking.Error
			if !errors.As(err, &conflict) || conflict.Code != "CAPACITY_FULL" {
				t.Fatal("unexpected race error", err)
			}
		}
	}
	if success != 1 {
		t.Fatal("last seat successes", success)
	}
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM bookings WHERE room_session_id=$1 AND status IN ('PENDING_PT','CONFIRMED')`, session).Scan(&n); err != nil || n != 1 {
		t.Fatal("capacity exceeded", n, err)
	}
	// Distinct slot IDs with overlapping snapshot times must still conflict across rooms.
	room2 := mustID(`INSERT INTO rooms(name,room_type,default_capacity) VALUES('Overlap room','GYM',10) RETURNING id`)
	slot2 := mustID(`INSERT INTO time_slots(name,start_time,end_time,is_active) VALUES('Legacy overlap','09:30','10:30',false) RETURNING id`)
	session2 := mustID(`INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date+2,(current_date+2+time '09:30') AT TIME ZONE 'Asia/Ho_Chi_Minh',(current_date+2+time '10:30') AT TIME ZONE 'Asia/Ho_Chi_Minh',10,$3) RETURNING id`, room2, slot2, a.users["ADMIN"])
	first, err := repo.Create(ctx, identity.User{ID: a.users["MEMBER"], Role: identity.RoleMember}, 0, booking.CreateInput{RoomSessionID: session2, PTID: pt}, "overlap-first", "overlap")
	if err != nil {
		t.Fatal(err)
	}
	// A third room avoids the filled first room masking the overlap guard.
	room3 := mustID(`INSERT INTO rooms(name,room_type,default_capacity) VALUES('Third room','GYM',10) RETURNING id`)
	session3 := mustID(`INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) SELECT $1,$2,session_date,starts_at,ends_at,10,$3 FROM room_sessions WHERE id=$4 RETURNING id`, room3, slot, a.users["ADMIN"], session)
	_, err = repo.Create(ctx, identity.User{ID: a.users["MEMBER"], Role: identity.RoleMember}, 0, booking.CreateInput{RoomSessionID: session3}, "overlap-member", "overlap")
	var conflict *booking.Error
	if !errors.As(err, &conflict) || conflict.Code != "MEMBER_OVERLAP" {
		t.Fatal("cross-room member overlap", err)
	}
	var freeUser int64
	if err = a.pool.QueryRow(ctx, `SELECT m.user_id FROM members m WHERE m.member_code LIKE 'RACE-%' AND NOT EXISTS(SELECT 1 FROM bookings b WHERE b.member_id=m.id AND b.status IN ('PENDING_PT','CONFIRMED')) LIMIT 1`).Scan(&freeUser); err != nil {
		t.Fatal(err)
	}
	_, err = repo.Create(ctx, identity.User{ID: freeUser, Role: identity.RoleMember}, 0, booking.CreateInput{RoomSessionID: session3, PTID: pt}, "overlap-pt-key", "overlap")
	if !errors.As(err, &conflict) || conflict.Code != "PT_UNAVAILABLE" {
		t.Fatal("cross-slot PT overlap", err)
	}
	a.call("POST", fmt.Sprintf("/bookings/%d/cancel", first.ID), "MEMBER", map[string]string{}, 200)
	for _, role := range []string{"MEMBER", "PT", "MANAGER"} {
		a.call("POST", fmt.Sprintf("/members/%d/bookings", member), role, booking.CreateInput{RoomSessionID: session2}, 403)
	}
	raw, _ := json.Marshal(booking.CreateInput{RoomSessionID: session2})
	onBehalf := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/members/%d/bookings", member), bytes.NewReader(raw))
	onBehalf.Header.Set("Authorization", "Bearer "+a.tokens["RECEPTIONIST"])
	onBehalf.Header.Set("Idempotency-Key", "onbehalf-demo-key")
	onBehalf.Header.Set("Content-Type", "application/json")
	onBehalfResponse := httptest.NewRecorder()
	a.router.ServeHTTP(onBehalfResponse, onBehalf)
	if onBehalfResponse.Code != 201 {
		t.Fatal("reception create", onBehalfResponse.Code, onBehalfResponse.Body.String())
	}
	decodeData(t, onBehalfResponse, &b)
	a.call("POST", fmt.Sprintf("/bookings/%d/cancel", b.ID), "MANAGER", map[string]string{}, 422)
	a.call("POST", fmt.Sprintf("/bookings/%d/cancel", b.ID), "MANAGER", map[string]string{"reason": "Synthetic operation override"}, 200)
	// Covers are protected binary endpoints; upload strips PNG metadata into bounded JPEGs.
	var img bytes.Buffer
	png.Encode(&img, image.NewRGBA(image.Rect(0, 0, 80, 40)))
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	f, _ := writer.CreateFormFile("file", "room.png")
	f.Write(img.Bytes())
	writer.WriteField("alt_text", "Phòng tập kiểm thử")
	writer.Close()
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/rooms/%d/cover-image", room), &body)
	req.Header.Set("Authorization", "Bearer "+a.tokens["ADMIN"])
	req.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	a.router.ServeHTTP(response, req)
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/rooms/%d/cover-image?size=card", room), nil)
	req.Header.Set("Authorization", "Bearer "+a.tokens["MEMBER"])
	response = httptest.NewRecorder()
	a.router.ServeHTTP(response, req)
	if response.Code != 200 || response.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatal("cover read", response.Code)
	}
	a.call("DELETE", fmt.Sprintf("/rooms/%d/cover-image", room), "MEMBER", nil, 403)
	a.call("DELETE", fmt.Sprintf("/rooms/%d/cover-image", room), "ADMIN", nil, 200)
	a.call("GET", fmt.Sprintf("/rooms/%d/cover-image", room), "MEMBER", nil, 404)
}

func TestBookingBoundaries(t *testing.T) {
	now := time.Date(2026, 10, 12, 7, 0, 0, 0, time.UTC)
	if !booking.CanCancel(now, now.Add(2*time.Hour)) || booking.CanCancel(now.Add(time.Nanosecond), now.Add(2*time.Hour)) {
		t.Fatal("cancel boundary")
	}
	if booking.Deadline(now, now.Add(10*time.Minute)) != now.Add(10*time.Minute) || booking.Deadline(now, now.Add(time.Hour)) != now.Add(30*time.Minute) {
		t.Fatal("deadline boundary")
	}
}
