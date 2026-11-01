package repository_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
)

func TestBookingBatchIntegration(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	mustID := func(query string, args ...any) int64 {
		t.Helper()
		var id int64
		if err := a.pool.QueryRow(ctx, query, args...).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	member := mustID(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'BATCH-M',current_date,current_date+365) RETURNING id`, a.users["MEMBER"])
	pt := mustID(`INSERT INTO personal_trainers(user_id,approval_status,decided_at,decided_by_user_id) VALUES($1,'APPROVED',clock_timestamp(),$2) RETURNING id`, a.users["PT"], a.users["ADMIN"])
	room := mustID(`INSERT INTO rooms(name,room_type,default_capacity) VALUES('Batch booking room','GYM',1) RETURNING id`)
	intervals := []struct{ start, end string }{{"08:00", "09:00"}, {"09:00", "10:00"}, {"10:00", "11:00"}, {"11:00", "12:00"}}
	sessions := make([]int64, 0, len(intervals))
	for i, interval := range intervals {
		slot := mustID(`INSERT INTO time_slots(name,start_time,end_time) VALUES($1,$2,$3) RETURNING id`, fmt.Sprintf("Batch slot %d", i+1), interval.start, interval.end)
		session := mustID(`INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date+2,((current_date+2)+$3::time) AT TIME ZONE 'Asia/Ho_Chi_Minh',((current_date+2)+$4::time) AT TIME ZONE 'Asia/Ho_Chi_Minh',1,$5) RETURNING id`, room, slot, interval.start, interval.end, a.users["ADMIN"])
		sessions = append(sessions, session)
	}
	a.sql(`INSERT INTO pt_working_hours(pt_id,day_of_week,start_time,end_time) VALUES($1,extract(isodow from current_date+2),'07:00','21:00')`, pt)

	createBatch := func(role, path, key string, input booking.BatchCreateInput) *httptest.ResponseRecorder {
		t.Helper()
		raw, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("POST", "/api/v1"+path, bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer "+a.tokens[role])
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		w := httptest.NewRecorder()
		a.router.ServeHTTP(w, req)
		return w
	}

	input := booking.BatchCreateInput{RoomSessionIDs: []int64{sessions[1], sessions[0]}, PTID: pt}
	created := createBatch("MEMBER", "/bookings/batch", "booking-batch-first", input)
	if created.Code != 201 {
		t.Fatal(created.Code, created.Body.String())
	}
	var batch []booking.Booking
	decodeData(t, created, &batch)
	if len(batch) != 2 || batch[0].RoomSessionID != sessions[0] || batch[1].RoomSessionID != sessions[1] {
		t.Fatalf("batch was not normalized chronologically: %+v", batch)
	}
	for _, createdBooking := range batch {
		if createdBooking.PTID != pt || createdBooking.Status != "PENDING_PT" || createdBooking.RespondBy == nil {
			t.Fatalf("unexpected pending PT booking: %+v", createdBooking)
		}
	}
	replay := createBatch("MEMBER", "/bookings/batch", "booking-batch-first", input)
	var replayed []booking.Booking
	decodeData(t, replay, &replayed)
	if replay.Code != 201 || len(replayed) != 2 || replayed[0].ID != batch[0].ID || replayed[1].ID != batch[1].ID {
		t.Fatal("batch replay created new bookings", replay.Code, replay.Body.String())
	}
	if changed := createBatch("MEMBER", "/bookings/batch", "booking-batch-first", booking.BatchCreateInput{RoomSessionIDs: input.RoomSessionIDs}); changed.Code != 409 || !strings.Contains(changed.Body.String(), "IDEMPOTENCY_CONFLICT") {
		t.Fatal("batch idempotency conflict", changed.Code, changed.Body.String())
	}
	if invalid := createBatch("MEMBER", "/bookings/batch", "booking-batch-gap", booking.BatchCreateInput{RoomSessionIDs: []int64{sessions[0], sessions[2]}}); invalid.Code != 422 || !strings.Contains(invalid.Body.String(), "BATCH_SESSIONS_NOT_CONSECUTIVE") {
		t.Fatal("non-consecutive batch was accepted", invalid.Code, invalid.Body.String())
	}
	if duplicate := createBatch("MEMBER", "/bookings/batch", "booking-batch-duplicate", booking.BatchCreateInput{RoomSessionIDs: []int64{sessions[2], sessions[2]}}); duplicate.Code != 422 {
		t.Fatal("duplicate session batch was accepted", duplicate.Code, duplicate.Body.String())
	}
	if forbidden := createBatch("MEMBER", fmt.Sprintf("/members/%d/bookings/batch", member), "booking-batch-forbidden", booking.BatchCreateInput{RoomSessionIDs: []int64{sessions[2], sessions[3]}}); forbidden.Code != 403 {
		t.Fatal("member was allowed to create on behalf", forbidden.Code, forbidden.Body.String())
	}

	newMember := func(username, code string) (identity.User, int64) {
		uid := mustID(`INSERT INTO users(username,password_hash,email,full_name,role_id) SELECT $1,'test-unused-hash',$2,$3,id FROM roles WHERE code='MEMBER' RETURNING users.id`, username, username+"@test.local", username)
		memberID := mustID(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,$2,current_date,current_date+365) RETURNING id`, uid, code)
		return identity.User{ID: uid, Role: identity.RoleMember}, memberID
	}
	repo := repository.NewBookingRepository(a.pool)
	seatHolder, _ := newMember("batch-seat-holder", "BATCH-SEAT")
	if _, err := repo.Create(ctx, seatHolder, 0, booking.CreateInput{RoomSessionID: sessions[3]}, "booking-batch-seat-holder", "batch-test"); err != nil {
		t.Fatal(err)
	}
	failingActor, failingMember := newMember("batch-atomic-failure", "BATCH-ROLLBACK")
	_, err := repo.CreateBatch(ctx, failingActor, 0, booking.BatchCreateInput{RoomSessionIDs: []int64{sessions[2], sessions[3]}}, "booking-batch-atomic", "batch-test")
	var conflict *booking.Error
	if !errors.As(err, &conflict) || conflict.Code != "CAPACITY_FULL" {
		t.Fatal("expected capacity conflict for group", err)
	}
	var count int
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM bookings WHERE member_id=$1 AND room_session_id=ANY($2::bigint[])`, failingMember, []int64{sessions[2], sessions[3]}).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("partial batch survived a later capacity failure", count)
	}
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM idempotency_keys WHERE actor_user_id=$1 AND request_scope='booking:create-batch' AND idempotency_key='booking-batch-first'`, a.users["MEMBER"]).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("batch idempotency result was not retained", count)
	}
}

func TestBookingBatchConcurrentLastSeat(t *testing.T) {
	for _, competitor := range []string{"batch", "single"} {
		t.Run(competitor, func(t *testing.T) {
			a := newManagementAPI(t)
			ctx := context.Background()
			getID := func(query string, args ...any) int64 {
				t.Helper()
				var id int64
				if err := a.pool.QueryRow(ctx, query, args...).Scan(&id); err != nil {
					t.Fatal(err)
				}
				return id
			}
			getID(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'RACE-M1',current_date,current_date+365) RETURNING id`, a.users["MEMBER"])
			secondUser := getID(`INSERT INTO users(username,password_hash,email,full_name,role_id) SELECT 'batch-race-member','test-hash','batch-race@test.local','Batch Race',id FROM roles WHERE code='MEMBER' RETURNING users.id`)
			getID(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'RACE-M2',current_date,current_date+365) RETURNING id`, secondUser)
			room := getID(`INSERT INTO rooms(name,room_type,default_capacity) VALUES('Batch race room','GYM',1) RETURNING id`)
			var sessions []int64
			for i, interval := range [][2]string{{"08:00", "09:00"}, {"09:00", "10:00"}} {
				slot := getID(`INSERT INTO time_slots(name,start_time,end_time) VALUES($1,$2,$3) RETURNING id`, fmt.Sprintf("Race slot %d", i), interval[0], interval[1])
				sessions = append(sessions, getID(`INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date+2,((current_date+2)+$3::time) AT TIME ZONE 'Asia/Ho_Chi_Minh',((current_date+2)+$4::time) AT TIME ZONE 'Asia/Ho_Chi_Minh',1,$5) RETURNING id`, room, slot, interval[0], interval[1], a.users["ADMIN"]))
			}
			repo := repository.NewBookingRepository(a.pool)
			start := make(chan struct{})
			results := make(chan error, 2)
			var workers sync.WaitGroup
			workers.Add(2)
			go func() {
				defer workers.Done()
				<-start
				_, err := repo.CreateBatch(ctx, identity.User{ID: a.users["MEMBER"], Role: identity.RoleMember}, 0, booking.BatchCreateInput{RoomSessionIDs: sessions}, "batch-race-first", "batch-race")
				results <- err
			}()
			go func() {
				defer workers.Done()
				<-start
				actor := identity.User{ID: secondUser, Role: identity.RoleMember}
				var err error
				if competitor == "batch" {
					_, err = repo.CreateBatch(ctx, actor, 0, booking.BatchCreateInput{RoomSessionIDs: sessions}, "batch-race-second", "batch-race")
				} else {
					_, err = repo.Create(ctx, actor, 0, booking.CreateInput{RoomSessionID: sessions[0]}, "single-race-second", "batch-race")
				}
				results <- err
			}()
			close(start)
			workers.Wait()
			close(results)
			successes, conflicts := 0, 0
			for err := range results {
				if err == nil {
					successes++
					continue
				}
				var business *booking.Error
				if errors.As(err, &business) && business.Code == "CAPACITY_FULL" {
					conflicts++
					continue
				}
				t.Fatalf("unexpected concurrent error: %v", err)
			}
			if successes != 1 || conflicts != 1 {
				t.Fatalf("success=%d conflict=%d", successes, conflicts)
			}
			var firstHolds, secondHolds int
			if err := a.pool.QueryRow(ctx, `SELECT count(*) FILTER (WHERE room_session_id=$1), count(*) FILTER (WHERE room_session_id=$2) FROM bookings WHERE status IN ('PENDING_PT','CONFIRMED')`, sessions[0], sessions[1]).Scan(&firstHolds, &secondHolds); err != nil {
				t.Fatal(err)
			}
			if firstHolds != 1 || (competitor == "batch" && secondHolds != 1) || secondHolds > 1 {
				t.Fatalf("first holds=%d second holds=%d", firstHolds, secondHolds)
			}
		})
	}
}
