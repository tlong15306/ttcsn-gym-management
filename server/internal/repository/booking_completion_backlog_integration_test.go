package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
)

func TestBookingCompletionBacklogOverOneThousand(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	a.sql(`INSERT INTO users(username,password_hash,email,full_name,role_id)
		SELECT 'completion.backlog.'||i,u.password_hash,'completion.backlog.'||i||'@test.local','Completion backlog '||i,r.id
		FROM generate_series(1,1001)i CROSS JOIN roles r CROSS JOIN users u
		WHERE r.code='MEMBER' AND u.id=$1`, a.users["ADMIN"])
	a.sql(`INSERT INTO members(user_id,member_code,join_date,membership_expiry)
		SELECT id,username,current_date,current_date+365 FROM users WHERE username LIKE 'completion.backlog.%'`)
	var room, slot, session int64
	if err := a.pool.QueryRow(ctx, `INSERT INTO rooms(name,room_type,default_capacity) VALUES('Completion backlog room','GYM',1200) RETURNING id`).Scan(&room); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO time_slots(name,start_time,end_time) VALUES('Completion backlog slot','00:00','24:00') RETURNING id`).Scan(&slot); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date,now()+interval '1 hour',now()+interval '2 hours',1200,$3) RETURNING id`, room, slot, a.users["ADMIN"]).Scan(&session); err != nil {
		t.Fatal(err)
	}
	a.sql(`INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,status,created_by_user_id)
		SELECT m.id,s.id,s.session_date,s.time_slot_id,'CONFIRMED',$2 FROM members m CROSS JOIN room_sessions s
		WHERE m.member_code LIKE 'completion.backlog.%' AND s.id=$1`, session, a.users["ADMIN"])
	a.sql(`INSERT INTO booking_attendance(booking_id,participant_type,participant_user_id,recorded_by_user_id)
		SELECT b.id,'MEMBER',m.user_id,$2 FROM bookings b JOIN members m ON m.id=b.member_id
		WHERE b.room_session_id=$1 AND b.id % 2 = 0`, session, a.users["ADMIN"])
	a.sql(`UPDATE room_sessions SET starts_at=now()-interval '1 hour',ends_at=now()-interval '1 second' WHERE id=$1`, session)
	repo := repository.NewBookingRepository(a.pool)
	previous := 1001
	for cycle := 1; cycle <= 8; cycle++ {
		stage, cancel := context.WithTimeout(ctx, 12*time.Second)
		started := time.Now()
		err := repo.Complete(stage)
		elapsed := time.Since(started)
		cancel()
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
		var remaining, finished, audits, wrong int
		err = a.pool.QueryRow(ctx, `SELECT
			count(*) FILTER (WHERE b.status='CONFIRMED'),
			count(*) FILTER (WHERE b.status IN ('COMPLETED','NO_SHOW')),
			(SELECT count(*) FROM audit_events WHERE action IN ('booking.COMPLETED','booking.NO_SHOW')),
			count(*) FILTER (WHERE (b.status='COMPLETED' AND NOT EXISTS (SELECT 1 FROM booking_attendance ba WHERE ba.booking_id=b.id AND ba.participant_type='MEMBER')) OR (b.status='NO_SHOW' AND EXISTS (SELECT 1 FROM booking_attendance ba WHERE ba.booking_id=b.id AND ba.participant_type='MEMBER')))
			FROM bookings b WHERE b.room_session_id=$1`, session).Scan(&remaining, &finished, &audits, &wrong)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("cycle %d: elapsed=%s remaining=%d finished=%d audits=%d", cycle, elapsed, remaining, finished, audits)
		if remaining > previous || previous-remaining > 1000 || finished != 1001-remaining || audits != finished || wrong != 0 {
			t.Fatalf("cycle %d: remaining=%d finished=%d audits=%d wrong=%d", cycle, remaining, finished, audits, wrong)
		}
		if remaining == 0 && previous == 0 {
			return
		}
		previous = remaining
	}
	t.Fatal("completion backlog did not drain and remain idempotent within eight cycles")
}
