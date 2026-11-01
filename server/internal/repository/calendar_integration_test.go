package repository_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func (a *testAPI) calendarCommit(room int64, m facility.CalendarMutation, p facility.CalendarPreview, key string, want int) *httptest.ResponseRecorder {
	a.t.Helper()
	path := fmt.Sprintf("/rooms/%d/weekly-schedule", room)
	method := "PUT"
	switch m.Kind {
	case "EXCEPTION":
		path = fmt.Sprintf("/rooms/%d/calendar/exceptions", room)
		method = "POST"
		if m.ExceptionID > 0 {
			path += fmt.Sprint("/", m.ExceptionID)
			method = "PATCH"
		}
	case "WITHDRAW":
		path = fmt.Sprintf("/rooms/%d/calendar/exceptions/%d/withdraw", room, m.ExceptionID)
		method = "POST"
	case "EXTEND":
		path = fmt.Sprintf("/rooms/%d/calendar/materialize", room)
		method = "POST"
	}
	b, _ := json.Marshal(facility.CalendarConfirm{CalendarMutation: m, PreviewID: p.ID})
	req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+a.tokens["ADMIN"])
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", key)
	w := httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	if w.Code != want {
		a.t.Fatalf("calendar confirm => %d want %d %s", w.Code, want, w.Body.String())
	}
	return w
}
func (a *testAPI) preview(room int64, m facility.CalendarMutation) facility.CalendarPreview {
	a.t.Helper()
	var p facility.CalendarPreview
	decodeData(a.t, a.call("POST", fmt.Sprintf("/rooms/%d/calendar/preview", room), "ADMIN", m, 200), &p)
	return p
}
func (a *testAPI) snapshot(room int64) facility.CalendarSnapshot {
	a.t.Helper()
	var s facility.CalendarSnapshot
	decodeData(a.t, a.call("GET", fmt.Sprintf("/rooms/%d/weekly-schedule", room), "ADMIN", nil, 200), &s)
	return s
}

func TestCalendarExceptionReplacementKeepsHistory(t *testing.T) {
	a := newManagementAPI(t)
	var room facility.Room
	decodeData(t, a.call("POST", "/rooms", "MANAGER", facility.RoomInput{Name: "Exception history fixture", RoomType: "Gym", DefaultCapacity: 10, Status: "AVAILABLE"}, 201), &room)
	date := time.Now().In(facility.Vietnam).AddDate(0, 0, 3).Format("2006-01-02")
	m := facility.CalendarMutation{Kind: "EXCEPTION", Date: date, Mode: "CLOSED", ClosureType: "HOLIDAY", Reason: "Initial closure"}
	a.calendarCommit(room.ID, m, a.preview(room.ID, m), "exception-history-create", 200)
	first := a.snapshot(room.ID).Exceptions[0]
	m.BaseRevision = 1
	m.ExceptionID = first.ID
	m.Mode = "REPLACE"
	m.Hours = []facility.Hours{{Start: 480, End: 600}}
	m.Reason = "Changed opening hours"
	a.calendarCommit(room.ID, m, a.preview(room.ID, m), "exception-history-edit", 200)
	s := a.snapshot(room.ID)
	if len(s.Exceptions) != 2 || s.Exceptions[0].ID != first.ID || s.Exceptions[0].Status != "WITHDRAWN" || s.Exceptions[1].ID == first.ID || s.Exceptions[1].Status != "ACTIVE" {
		t.Fatal("exception history was overwritten", s.Exceptions)
	}
	day := s.Day(date)
	if day.Source != "REPLACE" || len(day.Hours) != 1 || day.Hours[0].Start != 480 || day.Hours[0].End != 600 {
		t.Fatal("replacement hours were not published", day)
	}
	withdrawn := facility.CalendarMutation{Kind: "WITHDRAW", BaseRevision: 2, ExceptionID: first.ID, Reason: "Old version cannot be withdrawn again"}
	a.call("POST", fmt.Sprintf("/rooms/%d/calendar/preview", room.ID), "ADMIN", withdrawn, 404)
	m.BaseRevision = 2
	m.ExceptionID = s.Exceptions[1].ID
	m.Date = time.Now().In(facility.Vietnam).AddDate(0, 0, 4).Format("2006-01-02")
	a.call("POST", fmt.Sprintf("/rooms/%d/calendar/preview", room.ID), "ADMIN", m, 409)
	if a.snapshot(room.ID).Room.CalendarRevision != 2 {
		t.Fatal("rejected edit changed the revision")
	}
}

func TestRoomCalendarPostgresAPI(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	var room facility.Room
	decodeData(t, a.call("POST", "/rooms", "MANAGER", facility.RoomInput{Name: "Calendar test", RoomType: "Gym", DefaultCapacity: 5}, 201), &room)
	var slot facility.TimeSlot
	decodeData(t, a.call("POST", "/time-slots", "ADMIN", facility.SlotInput{Name: "Morning", StartTime: "08:00", EndTime: "09:00"}, 201), &slot)
	day := time.Now().In(facility.Vietnam).AddDate(0, 0, 1).Format("2006-01-02")
	a.call("POST", "/room-sessions", "ADMIN", facility.SessionInput{RoomID: room.ID, TimeSlotID: slot.ID, SessionDate: day}, 409)
	hours := []facility.Hours{}
	for d := 1; d <= 7; d++ {
		hours = append(hours, facility.Hours{Weekday: d, Start: 420, End: 1260})
	}
	m := facility.CalendarMutation{Kind: "WEEKLY", EffectiveFrom: facility.Today(time.Now()), Hours: hours, Reason: "Initial schedule"}
	p := a.preview(room.ID, m)
	if p.Creates < 27 || len(p.Conflicts) != 0 {
		t.Fatal(p)
	}
	// All three read-only roles receive effective hours, never draft/history/reasons.
	for _, role := range []string{"MEMBER", "PT", "RECEPTIONIST"} {
		a.call("POST", fmt.Sprintf("/rooms/%d/calendar/preview", room.ID), role, m, 403)
		a.call("GET", fmt.Sprintf("/rooms/%d/weekly-schedule", room.ID), role, nil, 403)
	}
	saved := a.calendarCommit(room.ID, m, p, "initial-calendar", 200)
	replay := a.calendarCommit(room.ID, m, p, "initial-calendar", 200)
	if saved.Body.String() != replay.Body.String() {
		var x, y facility.CalendarResult
		decodeData(t, saved, &x)
		decodeData(t, replay, &y)
		if x != y {
			t.Fatal("replay changed result")
		}
	}
	s := a.snapshot(room.ID)
	// A restarted process must replay committed keys despite losing its preview signing key.
	restarted := v1service.NewCalendarService(repository.NewCalendarRepository(a.pool))
	if _, err := restarted.Confirm(ctx, identity.User{ID: a.users["ADMIN"], Role: identity.RoleAdmin}, room.ID, facility.CalendarConfirm{CalendarMutation: m, PreviewID: "old-process-preview"}, "initial-calendar", "restart-replay"); err != nil {
		t.Fatal("restart idempotency", err)
	}
	if _, err := repository.NewCalendarRepository(a.pool).Commit(ctx, room.ID, a.users["ADMIN"], m, "irrelevant", time.Now().Add(-time.Minute), "expired-preview-key", "expired-test"); err == nil {
		t.Fatal("expired preview accepted")
	}
	if s.Room.CalendarRevision != 1 || len(s.Versions) != 1 {
		t.Fatal(s.Room)
	}
	for _, role := range []string{"MEMBER", "PT", "RECEPTIONIST"} {
		w := a.call("GET", fmt.Sprintf("/rooms/%d/calendar?from=%s&to=%s", room.ID, day, day), role, nil, 200)
		if bytes.Contains(w.Body.Bytes(), []byte("reason")) {
			t.Fatal("reason leaked")
		}
	}
	future := s.Sessions[0]
	for _, session := range s.Sessions {
		if session.SessionDate == day {
			future = session
		}
	}
	// A manually closed session remains closed through calendar extension and jobs.
	a.call("PATCH", fmt.Sprint("/room-sessions/", future.ID), "ADMIN", facility.SessionInput{Capacity: 5, Status: "CLOSED", Reason: "manual closure"}, 200)
	extend := facility.CalendarMutation{Kind: "EXTEND", BaseRevision: 1, Reason: "extend", Days: 28}
	p = a.preview(room.ID, extend)
	a.calendarCommit(room.ID, extend, p, "extend-calendar", 200)
	if err := repository.NewCalendarRepository(a.pool).MaterializeAll(ctx); err != nil {
		t.Fatal(err)
	}
	var source, status string
	if err := a.pool.QueryRow(ctx, `SELECT status,closure_source FROM room_sessions WHERE id=$1`, future.ID).Scan(&status, &source); err != nil || status != "CLOSED" || source != "MANUAL" {
		t.Fatal(status, source, err)
	}
	// Date closure, replacement, withdrawal restore weekly hours without deleting history.
	closeDay := time.Now().In(facility.Vietnam).AddDate(0, 0, 2).Format("2006-01-02")
	exception := facility.CalendarMutation{Kind: "EXCEPTION", BaseRevision: 2, Date: closeDay, Mode: "CLOSED", ClosureType: "HOLIDAY", Reason: "holiday"}
	p = a.preview(room.ID, exception)
	if len(p.Changes) != 1 {
		t.Fatal("expected closure", p)
	}
	a.calendarCommit(room.ID, exception, p, "close-exception", 200)
	s = a.snapshot(room.ID)
	id := s.Exceptions[0].ID
	withdraw := facility.CalendarMutation{Kind: "WITHDRAW", BaseRevision: 3, ExceptionID: id, Reason: "withdraw"}
	p = a.preview(room.ID, withdraw)
	if len(p.Changes) != 1 || p.Changes[0].Action != "REOPEN" {
		t.Fatal(p)
	}
	a.calendarCommit(room.ID, withdraw, p, "withdraw-exception", 200)
	// Midnight generation snapshots end at the following date and serialize 24:00.
	var late facility.TimeSlot
	decodeData(t, a.call("POST", "/time-slots", "ADMIN", facility.SlotInput{Name: "Last hour", StartTime: "23:00", EndTime: "24:00"}, 201), &late)
	if late.EndTime != "24:00" {
		t.Fatal(late)
	}
	replaceDate := time.Now().In(facility.Vietnam).AddDate(0, 0, 3).Format("2006-01-02")
	replacement := facility.CalendarMutation{Kind: "EXCEPTION", BaseRevision: 4, Date: replaceDate, Mode: "REPLACE", ClosureType: "OTHER", Hours: []facility.Hours{{Start: 1380, End: 1440}}, Reason: "late special"}
	p = a.preview(room.ID, replacement)
	a.calendarCommit(room.ID, replacement, p, "late-exception", 200)
	var end time.Time
	if err := a.pool.QueryRow(ctx, `SELECT ends_at FROM room_sessions WHERE room_id=$1 AND session_date=$2::date AND time_slot_id=$3`, room.ID, replaceDate, late.ID).Scan(&end); err != nil || end.In(facility.Vietnam).Format("2006-01-02 15:04") != facility.MinuteTime(replaceDate, 1440).Format("2006-01-02 15:04") {
		t.Fatal(end, err)
	}
	// Preview invalidation detects slot changes and new booking holds in addition to revision.
	blockDate := time.Now().In(facility.Vietnam).AddDate(0, 0, 4).Format("2006-01-02")
	block := facility.CalendarMutation{Kind: "EXCEPTION", BaseRevision: 5, Date: blockDate, Mode: "CLOSED", ClosureType: "OTHER", Reason: "hold block"}
	p = a.preview(room.ID, block)
	var memberID, sessionID int64
	a.pool.QueryRow(ctx, `INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'CAL-TEST',CURRENT_DATE,CURRENT_DATE+365) RETURNING id`, a.users["MEMBER"]).Scan(&memberID)
	a.pool.QueryRow(ctx, `SELECT id FROM room_sessions WHERE room_id=$1 AND session_date=$2::date AND time_slot_id=$3`, room.ID, blockDate, slot.ID).Scan(&sessionID)
	a.sql(`INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,status,created_by_user_id) VALUES($1,$2,$3::date,$4,'CONFIRMED',$5)`, memberID, sessionID, blockDate, slot.ID, a.users["ADMIN"])
	a.calendarCommit(room.ID, block, p, "stale-hold-key", 409)
	p = a.preview(room.ID, block)
	if len(p.Conflicts) != 1 {
		t.Fatal("missing hold conflict", p)
	}
	a.calendarCommit(room.ID, block, p, "blocked-hold-key", 409)
	if a.snapshot(room.ID).Room.CalendarRevision != 5 {
		t.Fatal("failed change committed")
	}
	a.sql(`DELETE FROM bookings WHERE room_session_id=$1`, sessionID)
	var ptID int64
	if err := a.pool.QueryRow(ctx, `INSERT INTO personal_trainers(user_id) VALUES($1) RETURNING id`, a.users["PT"]).Scan(&ptID); err != nil {
		t.Fatal(err)
	}
	a.sql(`INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,pt_id,status,respond_by,created_by_user_id) VALUES($1,$2,$3::date,$4,$5,'PENDING_PT',now()+interval '30 minutes',$6)`, memberID, sessionID, blockDate, slot.ID, ptID, a.users["ADMIN"])
	p = a.preview(room.ID, block)
	if len(p.Conflicts) != 1 {
		t.Fatal("PENDING_PT must hold", p)
	}
	a.calendarCommit(room.ID, block, p, "pending-hold-key", 409)
	a.sql(`UPDATE bookings SET status='EXPIRED' WHERE room_session_id=$1`, sessionID)
	p = a.preview(room.ID, block)
	if len(p.Conflicts) != 0 {
		t.Fatal("terminal booking still holds", p)
	}
	p = a.preview(room.ID, block)
	a.call("PATCH", fmt.Sprint("/time-slots/", late.ID, "/status"), "ADMIN", facility.SlotStatusInput{IsActive: false, Reason: "retire"}, 200)
	a.calendarCommit(room.ID, block, p, "stale-slot-key", 409)
	// Audit rejection rolls back version/exception/session/revision/idempotency writes together.
	a.sql(`CREATE FUNCTION fail_calendar_audit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.action LIKE 'room_calendar.%' THEN RAISE EXCEPTION 'fixture reject'; END IF; RETURN NEW; END $$`)
	a.sql(`CREATE TRIGGER reject_calendar BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION fail_calendar_audit()`)
	p = a.preview(room.ID, block)
	a.calendarCommit(room.ID, block, p, "audit-rollback-key", 500)
	if a.snapshot(room.ID).Room.CalendarRevision != 5 {
		t.Fatal("audit rollback failed")
	}
	a.sql(`DROP TRIGGER reject_calendar ON audit_events`)
	a.sql(`DROP FUNCTION fail_calendar_audit()`)
	// Concurrent managers cannot overwrite one another. Exactly one revision commits.
	p = a.preview(room.ID, block)
	var wg sync.WaitGroup
	codes := make(chan int, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body, _ := json.Marshal(facility.CalendarConfirm{CalendarMutation: block, PreviewID: p.ID})
			req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/rooms/%d/calendar/exceptions", room.ID), bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+a.tokens["ADMIN"])
			req.Header.Set("Idempotency-Key", fmt.Sprint("concurrent-key-", i))
			w := httptest.NewRecorder()
			a.router.ServeHTTP(w, req)
			codes <- w.Code
		}(i)
	}
	wg.Wait()
	close(codes)
	ok, conflict := 0, 0
	for code := range codes {
		if code == 200 {
			ok++
		} else if code == 409 {
			conflict++
		} else {
			t.Fatal(code)
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatal(ok, conflict)
	}
	if err := repository.NewCalendarRepository(a.pool).MaterializeAll(ctx); err != nil {
		t.Fatal(err)
	}
	before := len(a.snapshot(room.ID).Sessions)
	if err := repository.NewCalendarRepository(a.pool).MaterializeAll(ctx); err != nil {
		t.Fatal(err)
	}
	if len(a.snapshot(room.ID).Sessions) != before {
		t.Fatal("job duplicated sessions")
	}
}
