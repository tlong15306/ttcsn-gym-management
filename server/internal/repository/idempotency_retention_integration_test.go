package repository_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
)

func TestExpiredIdempotencyKeysAreNotReplayedAndArePurged(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	actor := a.users["ADMIN"]
	a.sql(`INSERT INTO idempotency_keys(actor_user_id,request_scope,idempotency_key,request_hash,response_status,response_body,expires_at) VALUES
		($1,'room-calendar:901','expired-calendar',repeat('a',64),200,'{}',clock_timestamp()-interval '1 second'),
		($1,'booking:create','expired-booking',repeat('b',64),201,'{}',clock_timestamp()-interval '1 second'),
		($1,'booking:create','active-booking',repeat('c',64),201,'{}',clock_timestamp()+interval '1 hour')`, actor)

	_, found, err := repository.NewCalendarRepository(a.pool).Replay(ctx, 901, actor, facility.CalendarMutation{}, "expired-calendar")
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("expired calendar key must not replay a committed result")
	}

	if err = repository.NewBookingRepository(a.pool).PurgeExpiredIdempotency(ctx); err != nil {
		t.Fatal(err)
	}
	var expired, active int
	if err = a.pool.QueryRow(ctx, `SELECT count(*) FILTER (WHERE expires_at<=clock_timestamp()),count(*) FILTER (WHERE idempotency_key='active-booking') FROM idempotency_keys`).Scan(&expired, &active); err != nil {
		t.Fatal(err)
	}
	if expired != 0 || active != 1 {
		t.Fatalf("purge retained expired=%d active=%d", expired, active)
	}

	// Calendar Commit also queries the table after preview validation. Reusing
	// an expired key with a new valid revision proves that Commit deletes it
	// before its uniqueness/idempotency check, rather than treating it as a
	// permanent conflict.
	var room facility.Room
	decodeData(t, a.call("POST", "/rooms", "ADMIN", facility.RoomInput{Name: "Retention calendar room", RoomType: "GYM", DefaultCapacity: 8, Status: "AVAILABLE"}, 201), &room)
	hours := make([]facility.Hours, 0, 7)
	for weekday := 1; weekday <= 7; weekday++ {
		hours = append(hours, facility.Hours{Weekday: weekday, Start: 420, End: 1260})
	}
	mutation := facility.CalendarMutation{Kind: "WEEKLY", EffectiveFrom: facility.Today(time.Now()), Hours: hours, Reason: "Initial retention schedule"}
	key := "expired-calendar-commit"
	a.calendarCommit(room.ID, mutation, a.preview(room.ID, mutation), key, 200)
	a.sql(`UPDATE idempotency_keys SET expires_at=clock_timestamp()-interval '1 second' WHERE actor_user_id=$1 AND request_scope=$2 AND idempotency_key=$3`, actor, "room-calendar:"+strconv.FormatInt(room.ID, 10), key)
	mutation.BaseRevision = 1
	mutation.EffectiveFrom = time.Now().In(facility.Vietnam).AddDate(0, 0, 1).Format("2006-01-02")
	mutation.Reason = "Replacement after retention"
	a.calendarCommit(room.ID, mutation, a.preview(room.ID, mutation), key, 200)
}
