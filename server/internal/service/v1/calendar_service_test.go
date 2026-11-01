package v1service

import (
	"context"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"strings"
	"testing"
	"time"
)

type calendarStub struct {
	commits int
	expires time.Time
}

func (s *calendarStub) Replay(context.Context, int64, int64, facility.CalendarMutation, string) (facility.CalendarResult, bool, error) {
	return facility.CalendarResult{}, false, nil
}

func (s *calendarStub) Snapshot(context.Context, int64) (facility.CalendarSnapshot, error) {
	return facility.CalendarSnapshot{}, nil
}
func (s *calendarStub) Preview(context.Context, int64, facility.CalendarMutation) (facility.CalendarPreview, string, error) {
	return facility.CalendarPreview{}, "fingerprint", nil
}
func (s *calendarStub) Commit(_ context.Context, _, _ int64, _ facility.CalendarMutation, _ string, expires time.Time, _, _ string) (facility.CalendarResult, error) {
	s.commits++
	s.expires = expires
	return facility.CalendarResult{}, nil
}
func (s *calendarStub) MaterializeAll(context.Context) error { return nil }
func TestCalendarPreviewBinding(t *testing.T) {
	repo := &calendarStub{}
	s := NewCalendarService(repo)
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	ctx := context.Background()
	actor := identity.User{ID: 1, Role: identity.RoleAdmin}
	m := facility.CalendarMutation{Kind: "WEEKLY", EffectiveFrom: "2026-10-12", Hours: []facility.Hours{{Weekday: 1, Start: 420, End: 1260}}, Reason: "publish"}
	p, err := s.Preview(ctx, actor, 2, m)
	if err != nil || !p.ExpiresAt.Equal(now.Add(5*time.Minute)) {
		t.Fatal(p, err)
	}
	if _, err = s.Confirm(ctx, actor, 2, facility.CalendarConfirm{CalendarMutation: m, PreviewID: p.ID}, "request-key", "rid"); err != nil {
		t.Fatal(err)
	}
	if repo.commits != 1 || !repo.expires.Equal(p.ExpiresAt) {
		t.Fatal(repo)
	}
	for _, tc := range []struct {
		actor    identity.User
		room     int64
		mutation facility.CalendarMutation
		token    string
	}{{identity.User{ID: 2, Role: identity.RoleManager}, 2, m, p.ID}, {actor, 3, m, p.ID}, {actor, 2, facility.CalendarMutation{Kind: "WEEKLY", EffectiveFrom: "2026-10-12", Reason: "changed"}, p.ID}, {actor, 2, m, p.ID + "tamper"}} {
		if _, err = s.Confirm(ctx, tc.actor, tc.room, facility.CalendarConfirm{CalendarMutation: tc.mutation, PreviewID: tc.token}, "request-key", "rid"); err == nil {
			t.Fatal("unbound preview accepted")
		}
	}
	if repo.commits != 1 {
		t.Fatal("invalid previews reached repository")
	}
	for _, role := range []identity.Role{identity.RoleMember, identity.RolePT, identity.RoleReceptionist} {
		actor.Role = role
		if _, err = s.Preview(ctx, actor, 2, m); err != identity.ErrForbidden {
			t.Fatal("RBAC", err)
		}
	}
	if _, err = s.Days(ctx, 2, "2026-10-11", "2026-12-11"); err == nil {
		t.Fatal("unbounded date range")
	}
	m.Reason = strings.Repeat("x", 1001)
	if _, err = s.normalize(m); err == nil {
		t.Fatal("reason length")
	}
}
