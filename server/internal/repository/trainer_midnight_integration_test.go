package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
)

func TestTrainerAvailabilityMidnightProtectsBookings(t *testing.T) {
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
	getID(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'MIDNIGHT-M',current_date,current_date+365) RETURNING id`, a.users["MEMBER"])
	pt := getID(`INSERT INTO personal_trainers(user_id,approval_status,decided_at,decided_by_user_id) VALUES($1,'APPROVED',clock_timestamp(),$2) RETURNING id`, a.users["PT"], a.users["ADMIN"])
	room := getID(`INSERT INTO rooms(name,room_type,default_capacity) VALUES('Midnight room','GYM',4) RETURNING id`)
	slot := getID(`INSERT INTO time_slots(name,start_time,end_time) VALUES('Midnight slot','23:30','24:00') RETURNING id`)
	session := getID(`INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date+2,((current_date+2)+time '23:30') AT TIME ZONE 'Asia/Ho_Chi_Minh',((current_date+2)+time '24:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',4,$3) RETURNING id`, room, slot, a.users["ADMIN"])
	var day int
	if err := a.pool.QueryRow(ctx, `SELECT extract(isodow from current_date+2)::int`).Scan(&day); err != nil {
		t.Fatal(err)
	}
	hours := func(end string) map[string]any {
		return map[string]any{"hours": []map[string]any{{"day_of_week": day, "start_time": "23:00", "end_time": end}}}
	}
	a.call("PUT", "/pts/me/availability", "PT", hours("24:00"), 200)
	created, err := repository.NewBookingRepository(a.pool).Create(ctx, identity.User{ID: a.users["MEMBER"], Role: identity.RoleMember}, 0, booking.CreateInput{RoomSessionID: session, PTID: pt}, "midnight-booking-test", "midnight-test")
	if err != nil || created.Status != "PENDING_PT" {
		t.Fatalf("create midnight booking: %v / %+v", err, created)
	}
	a.call("PUT", "/pts/me/availability", "PT", hours("23:30"), 409)
	var end string
	if err := a.pool.QueryRow(ctx, `SELECT to_char(end_time,'HH24:MI') FROM pt_working_hours WHERE pt_id=$1 AND is_active`, pt).Scan(&end); err != nil || end != "24:00" {
		t.Fatal(fmt.Sprintf("held session lost midnight coverage: end=%s err=%v", end, err))
	}
}
