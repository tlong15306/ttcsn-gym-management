package repository_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/member"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/trainer"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"sync"
	"testing"
)

func TestBookingConcurrentEligibilityWriters(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	admin := identity.User{ID: a.users["ADMIN"], Role: identity.RoleAdmin}
	repo := repository.NewBookingRepository(a.pool)
	facilities := repository.NewFacilityRepository(a.pool)
	trainers := repository.NewTrainerRepository(a.pool)
	id := func(q string, args ...any) int64 {
		var v int64
		if err := a.pool.QueryRow(ctx, q, args...).Scan(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	slot := id(`INSERT INTO time_slots(name,start_time,end_time) VALUES('Writer race','09:00','10:00') RETURNING id`)
	for _, kind := range []string{"room", "calendar", "capacity", "membership", "availability"} {
		t.Run(kind, func(t *testing.T) {
			uid := id(`INSERT INTO users(username,password_hash,email,full_name,role_id) SELECT $1::text,u.password_hash,$1::text||'@test.local','Writer member',r.id FROM users u CROSS JOIN roles r WHERE u.id=$2 AND r.code='MEMBER' RETURNING users.id`, "writer-member-"+kind, a.users["ADMIN"])
			mid := id(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,$2,current_date,current_date+365) RETURNING id`, uid, "WRITER-"+kind)
			puid := id(`INSERT INTO users(username,password_hash,email,full_name,role_id) SELECT $1::text,u.password_hash,$1::text||'@test.local','Writer PT',r.id FROM users u CROSS JOIN roles r WHERE u.id=$2 AND r.code='PT' RETURNING users.id`, "writer-pt-"+kind, a.users["ADMIN"])
			pt := id(`INSERT INTO personal_trainers(user_id,approval_status,decided_at,decided_by_user_id) VALUES($1,'APPROVED',now(),$2) RETURNING id`, puid, admin.ID)
			a.sql(`INSERT INTO pt_working_hours(pt_id,day_of_week,start_time,end_time) VALUES($1,extract(isodow from current_date+2),'07:00','21:00')`, pt)
			room := id(`INSERT INTO rooms(name,room_type,default_capacity) VALUES($1,'GYM',2) RETURNING id`, "Writer "+kind)
			session := id(`INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date+2,(current_date+2+time '09:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',(current_date+2+time '10:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',2,$3) RETURNING id`, room, slot, admin.ID)
			var date, today string
			if e := a.pool.QueryRow(ctx, `SELECT (current_date+2)::text,((now() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date)::text`).Scan(&date, &today); e != nil {
				t.Fatal(e)
			}
			writer := func() error { return nil }
			switch kind {
			case "room":
				writer = func() error {
					_, e := facilities.SaveRoom(ctx, room, facility.RoomInput{Name: "Writer " + kind, RoomType: "GYM", DefaultCapacity: 2, Status: "MAINTENANCE", Reason: "Synthetic race"}, admin.ID, "writer-room")
					return e
				}
			case "capacity":
				writer = func() error {
					_, e := facilities.SaveSession(ctx, session, facility.SessionInput{RoomID: room, TimeSlotID: slot, SessionDate: date, Capacity: 1, Status: "OPEN", Reason: "Synthetic race"}, admin.ID, "writer-capacity")
					return e
				}
			case "membership":
				writer = func() error {
					_, e := repository.NewMemberRepository(a.pool).Membership(ctx, mid, member.MembershipInput{MembershipExpiry: today, Status: "ACTIVE", Reason: "Synthetic race"}, admin.ID, "writer-member")
					return e
				}
			case "availability":
				writer = func() error {
					_, e := trainers.SaveHours(ctx, pt, []trainer.WorkingHour{}, admin.ID, "writer-hours")
					return e
				}
			case "calendar":
				calendars := service.NewCalendarService(repository.NewCalendarRepository(a.pool))
				m := facility.CalendarMutation{Kind: "WEEKLY", BaseRevision: 0, EffectiveFrom: today, Hours: []facility.Hours{}, Days: 28, Reason: "Synthetic closure race"}
				preview, e := calendars.Preview(ctx, admin, room, m)
				if e != nil {
					t.Fatalf("calendar preview: %v; today=%q session_date=%q", e, today, date)
				}
				writer = func() error {
					_, e := calendars.Confirm(ctx, admin, room, facility.CalendarConfirm{CalendarMutation: m, PreviewID: preview.ID}, "writer-calendar-key", "writer-calendar")
					return e
				}
			}
			var created booking.Booking
			var createErr, writerErr error
			var wg sync.WaitGroup
			start := make(chan struct{})
			wg.Add(2)
			go func() {
				defer wg.Done()
				<-start
				created, createErr = repo.Create(ctx, identity.User{ID: uid, Role: identity.RoleMember}, 0, booking.CreateInput{RoomSessionID: session, PTID: pt}, "writer-create-"+kind, "writer-create")
			}()
			go func() { defer wg.Done(); <-start; writerErr = writer() }()
			close(start)
			wg.Wait()
			checkError := func(err error) {
				if err == nil {
					return
				}
				var b *booking.Error
				var p *identity.PublicError
				if !errors.As(err, &b) && !errors.As(err, &p) && !errors.Is(err, identity.ErrNotFound) && !errors.Is(err, identity.ErrConflict) {
					t.Fatal("unexpected writer failure", err)
				}
			}
			checkError(createErr)
			checkError(writerErr)
			if createErr != nil && writerErr != nil {
				t.Fatal("neither writer committed", createErr, writerErr)
			}
			if (kind == "room" || kind == "calendar" || kind == "availability") && createErr == nil && writerErr == nil {
				t.Fatal("conflicting writers both committed")
			}
			var holds, capacity, audits int
			if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM bookings WHERE room_session_id=$1 AND status IN ('PENDING_PT','CONFIRMED')`, session).Scan(&holds); err != nil {
				t.Fatal(err)
			}
			if kind == "capacity" {
				if err := a.pool.QueryRow(ctx, `SELECT capacity FROM room_sessions WHERE id=$1`, session).Scan(&capacity); err != nil || holds > capacity {
					t.Fatal("capacity invariant", holds, capacity, err)
				}
			}
			if createErr == nil {
				if holds != 1 {
					t.Fatal("lost hold", holds)
				}
				if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE entity_type='booking' AND entity_id=$1`, fmt.Sprint(created.ID)).Scan(&audits); err != nil || audits != 1 {
					t.Fatal("creation audit", audits, err)
				}
			} else if holds != 0 {
				t.Fatal("failed creation left hold", holds)
			}
		})
	}
}
