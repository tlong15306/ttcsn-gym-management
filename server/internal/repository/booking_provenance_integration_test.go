package repository_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestBookingProvenanceMigrationDoesNotGuessLegacyHistory(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	var member, room, slot, session, bookingID int64
	if err := a.pool.QueryRow(ctx, `INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'PROVENANCE-M',current_date,current_date+365) RETURNING id`, a.users["MEMBER"]).Scan(&member); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO rooms(name,room_type,default_capacity) VALUES('Provenance legacy room','GYM',1) RETURNING id`).Scan(&room); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO time_slots(name,start_time,end_time) VALUES('Provenance legacy slot','09:00','10:00') RETURNING id`).Scan(&slot); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date+2,(current_date+2+time '09:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',(current_date+2+time '10:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',1,$3) RETURNING id`, room, slot, a.users["ADMIN"]).Scan(&session); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,status,created_by_user_id) SELECT $1,id,session_date,time_slot_id,'CONFIRMED',$2 FROM room_sessions WHERE id=$3 RETURNING bookings.id`, member, a.users["ADMIN"], session).Scan(&bookingID); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("pgx", os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	migrations := filepath.Join("..", "db", "migrations")
	if err = goose.DownTo(db, migrations, 4); err != nil {
		t.Fatal("roll back only the provenance migration on the isolated test DB:", err)
	}
	if _, err = db.Exec(`UPDATE bookings SET status='CANCELLED',cancelled_by_user_id=$2,cancelled_at=clock_timestamp(),cancellation_reason='Legacy cancellation' WHERE id=$1`, bookingID, a.users["MEMBER"]); err != nil {
		t.Fatal(err)
	}
	if err = goose.Up(db, migrations); err != nil {
		t.Fatal(err)
	}
	var source, actorRole, confirmedAt sql.NullString
	if err = db.QueryRow(`SELECT cancellation_source,cancelled_actor_role,confirmed_at::text FROM bookings WHERE id=$1`, bookingID).Scan(&source, &actorRole, &confirmedAt); err != nil {
		t.Fatal(err)
	}
	if !source.Valid || source.String != "UNKNOWN" || actorRole.Valid || confirmedAt.Valid {
		t.Fatalf("legacy provenance guessed or lost: source=%s actor_role=%s confirmed_at=%s", nullable(source), nullable(actorRole), nullable(confirmedAt))
	}
}

func nullable(value sql.NullString) string {
	if !value.Valid {
		return "NULL"
	}
	return value.String
}
