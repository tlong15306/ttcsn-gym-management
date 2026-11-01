package repository_test

import (
	"context"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	"testing"
)

func TestBookingJobsBacklogOver100(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	a.sql(`INSERT INTO users(username,password_hash,email,full_name,role_id) SELECT 'backlog.m.'||i,u.password_hash,'backlog.m.'||i||'@test.local','Backlog member '||i,r.id FROM generate_series(1,105)i CROSS JOIN roles r CROSS JOIN users u WHERE r.code='MEMBER' AND u.id=$1`, a.users["ADMIN"])
	a.sql(`INSERT INTO users(username,password_hash,email,full_name,role_id) SELECT 'backlog.pt.'||i,u.password_hash,'backlog.pt.'||i||'@test.local','Backlog PT '||i,r.id FROM generate_series(1,105)i CROSS JOIN roles r CROSS JOIN users u WHERE r.code='PT' AND u.id=$1`, a.users["ADMIN"])
	a.sql(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) SELECT id,username,current_date,current_date+365 FROM users WHERE username LIKE 'backlog.m.%'`)
	a.sql(`INSERT INTO personal_trainers(user_id,approval_status,decided_at,decided_by_user_id) SELECT id,'APPROVED',now(),$1 FROM users WHERE username LIKE 'backlog.pt.%'`, a.users["ADMIN"])
	var room, slot, session int64
	for _, item := range []struct {
		query string
		id    *int64
	}{{`INSERT INTO rooms(name,room_type,default_capacity) VALUES('Backlog demo','GYM',200) RETURNING id`, &room}, {`INSERT INTO time_slots(name,start_time,end_time) VALUES('Backlog demo','00:00','24:00') RETURNING id`, &slot}} {
		if err := a.pool.QueryRow(ctx, item.query).Scan(item.id); err != nil {
			t.Fatal(err)
		}
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date,now()+interval '1 hour',now()+interval '2 hours',200,$3) RETURNING id`, room, slot, a.users["ADMIN"]).Scan(&session); err != nil {
		t.Fatal(err)
	}
	a.sql(`INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,pt_id,status,respond_by,created_by_user_id) SELECT m.id,s.id,s.session_date,s.time_slot_id,p.id,'PENDING_PT',now()-interval '1 minute',$2 FROM members m JOIN users mu ON mu.id=m.user_id JOIN users pu ON pu.username=replace(mu.username,'backlog.m.','backlog.pt.') JOIN personal_trainers p ON p.user_id=pu.id CROSS JOIN room_sessions s WHERE mu.username LIKE 'backlog.m.%' AND s.id=$1`, session, a.users["ADMIN"])
	repo := repository.NewBookingRepository(a.pool)
	if err := repo.Expire(ctx); err != nil {
		t.Fatal(err)
	}
	if err := repo.Expire(ctx); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM bookings WHERE status='EXPIRED'`).Scan(&count); err != nil || count != 105 {
		t.Fatal("expiry backlog", count, err)
	}
	a.sql(`INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,status,created_by_user_id) SELECT m.id,s.id,s.session_date,s.time_slot_id,'CONFIRMED',$2 FROM members m JOIN users u ON u.id=m.user_id CROSS JOIN room_sessions s WHERE u.username LIKE 'backlog.m.%' AND s.id=$1`, session, a.users["ADMIN"])
	a.sql(`UPDATE room_sessions SET starts_at=now()-interval '1 hour',ends_at=now()-interval '1 second' WHERE id=$1`, session)
	if err := repo.Complete(ctx); err != nil {
		t.Fatal(err)
	}
	if err := repo.Complete(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM bookings WHERE status='NO_SHOW'`).Scan(&count); err != nil || count != 105 {
		t.Fatal("completion backlog", count, err)
	}
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action IN ('booking.EXPIRED','booking.NO_SHOW')`).Scan(&count); err != nil || count != 210 {
		t.Fatal("job replay audits", count, err)
	}
}
