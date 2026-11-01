package repository_test

import (
	"context"
	"fmt"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	"sync"
	"testing"
)

func TestPendingAcceptanceRaces(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	id := func(q string, args ...any) int64 {
		var v int64
		if err := a.pool.QueryRow(ctx, q, args...).Scan(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	member := id(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'PENDING-RACE',current_date,current_date+365) RETURNING id`, a.users["MEMBER"])
	pt := id(`INSERT INTO personal_trainers(user_id,approval_status,decided_at,decided_by_user_id) VALUES($1,'APPROVED',now(),$2) RETURNING id`, a.users["PT"], a.users["ADMIN"])
	room := id(`INSERT INTO rooms(name,room_type,default_capacity) VALUES('Pending races','GYM',1) RETURNING id`)
	slot := id(`INSERT INTO time_slots(name,start_time,end_time) VALUES('Pending races','09:00','10:00') RETURNING id`)
	session := id(`INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date+2,(current_date+2+time '09:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',(current_date+2+time '10:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',1,$3) RETURNING id`, room, slot, a.users["ADMIN"])
	a.sql(`INSERT INTO pt_working_hours(pt_id,day_of_week,start_time,end_time) VALUES($1,extract(isodow from current_date+2),'07:00','21:00')`, pt)
	repo := repository.NewBookingRepository(a.pool)
	memberActor := identity.User{ID: a.users["MEMBER"], Role: identity.RoleMember}
	ptActor := identity.User{ID: a.users["PT"], Role: identity.RolePT}
	operator := identity.User{ID: a.users["ADMIN"], Role: identity.RoleAdmin}
	for _, kind := range []string{"expire", "cancel"} {
		b, err := repo.Create(ctx, memberActor, 0, booking.CreateInput{RoomSessionID: session, PTID: pt}, "pending-race-"+kind, "pending-create")
		if err != nil {
			t.Fatal(err)
		}
		if b.MemberID != member {
			t.Fatal("wrong member")
		}
		if kind == "expire" {
			a.sql(`UPDATE bookings SET respond_by=clock_timestamp()-interval '1 second' WHERE id=$1`, b.ID)
		}
		var wg sync.WaitGroup
		start := make(chan struct{})
		var acceptErr, otherErr error
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			_, acceptErr = repo.Act(ctx, ptActor, b.ID, booking.ActionInput{Action: "ACCEPT"}, false, "accept-race")
		}()
		go func() {
			defer wg.Done()
			<-start
			if kind == "expire" {
				otherErr = repo.Expire(ctx)
			} else {
				_, otherErr = repo.Act(ctx, operator, b.ID, booking.ActionInput{Reason: "Synthetic cancellation race"}, true, "cancel-race")
			}
		}()
		close(start)
		wg.Wait()
		if otherErr != nil {
			t.Fatal(otherErr)
		}
		var status string
		var holds, audits int
		a.pool.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1`, b.ID).Scan(&status)
		a.pool.QueryRow(ctx, `SELECT count(*) FROM bookings WHERE id=$1 AND status IN ('PENDING_PT','CONFIRMED')`, b.ID).Scan(&holds)
		a.pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE entity_type='booking' AND entity_id=$1`, fmt.Sprint(b.ID)).Scan(&audits)
		expected := "CANCELLED"
		if kind == "expire" {
			expected = "EXPIRED"
			if acceptErr == nil {
				t.Fatal("expired acceptance committed")
			}
		}
		if status != expected || holds != 0 {
			t.Fatal("pending race invariant", status, holds)
		}
		expectedAudits := 2
		if kind == "cancel" && acceptErr == nil {
			expectedAudits = 3
		}
		if audits != expectedAudits {
			t.Fatal("pending race audit", audits, expectedAudits)
		}
	}
}
