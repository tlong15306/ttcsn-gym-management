package repository_test

import (
	"context"
	"testing"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
)

func TestBookingExpirationBacklogOverOneThousand(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	a.sql(`INSERT INTO users(username,password_hash,email,full_name,role_id)
		SELECT 'expiry.backlog.'||i,u.password_hash,'expiry.backlog.'||i||'@test.local','Expiry backlog '||i,r.id
		FROM generate_series(1,1001)i CROSS JOIN roles r CROSS JOIN users u
		WHERE r.code='MEMBER' AND u.id=$1`, a.users["ADMIN"])
	a.sql(`INSERT INTO members(user_id,member_code,join_date,membership_expiry)
		SELECT id,username,current_date,current_date+365 FROM users WHERE username LIKE 'expiry.backlog.%'`)
	a.sql(`INSERT INTO users(username,password_hash,email,full_name,role_id)
		SELECT 'expiry.backlog.pt.'||i,u.password_hash,'expiry.backlog.pt.'||i||'@test.local','Expiry backlog PT '||i,r.id
		FROM generate_series(1,1001)i CROSS JOIN roles r CROSS JOIN users u
		WHERE r.code='PT' AND u.id=$1`, a.users["ADMIN"])
	a.sql(`INSERT INTO personal_trainers(user_id,approval_status,decided_at,decided_by_user_id)
		SELECT id,'APPROVED',clock_timestamp(),$1 FROM users WHERE username LIKE 'expiry.backlog.pt.%'`, a.users["ADMIN"])

	var room, slot, session int64
	if err := a.pool.QueryRow(ctx, `INSERT INTO rooms(name,room_type,default_capacity) VALUES('Expiry backlog room','GYM',1200) RETURNING id`).Scan(&room); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO time_slots(name,start_time,end_time) VALUES('Expiry backlog slot','00:00','24:00') RETURNING id`).Scan(&slot); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date,now()+interval '1 hour',now()+interval '2 hours',1200,$3) RETURNING id`, room, slot, a.users["ADMIN"]).Scan(&session); err != nil {
		t.Fatal(err)
	}
	a.sql(`INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,pt_id,status,respond_by,created_by_user_id)
		SELECT m.id,s.id,s.session_date,s.time_slot_id,p.id,'PENDING_PT',now()-interval '1 minute',$2
		FROM members m
		JOIN users mu ON mu.id=m.user_id
		JOIN users pu ON pu.username=replace(mu.username,'expiry.backlog.','expiry.backlog.pt.')
		JOIN personal_trainers p ON p.user_id=pu.id
		CROSS JOIN room_sessions s
		WHERE m.member_code LIKE 'expiry.backlog.%' AND s.id=$1`, session, a.users["ADMIN"])

	repo := repository.NewBookingRepository(a.pool)
	if err := repo.Expire(ctx); err != nil {
		t.Fatal(err)
	}
	var expired, pending int
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FILTER (WHERE status='EXPIRED'),count(*) FILTER (WHERE status='PENDING_PT') FROM bookings`).Scan(&expired, &pending); err != nil {
		t.Fatal(err)
	}
	if expired != 1000 || pending != 1 {
		t.Fatalf("first expiration cycle expired=%d pending=%d, want 1000 and 1", expired, pending)
	}

	if err := repo.Expire(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FILTER (WHERE status='EXPIRED'),count(*) FILTER (WHERE status='PENDING_PT') FROM bookings`).Scan(&expired, &pending); err != nil {
		t.Fatal(err)
	}
	if expired != 1001 || pending != 0 {
		t.Fatalf("second expiration cycle expired=%d pending=%d, want 1001 and 0", expired, pending)
	}
	var audits int
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action='booking.EXPIRED'`).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if audits != 1001 {
		t.Fatalf("expiration audit count=%d, want 1001", audits)
	}
}
