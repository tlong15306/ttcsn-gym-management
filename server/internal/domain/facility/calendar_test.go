package facility

import (
	"testing"
	"time"
)

func TestHoursBoundaries(t *testing.T) {
	h, err := NormalizeHours([]Hours{{1, 420, 720}, {1, 720, 1260}}, true)
	if err != nil || len(h) != 1 || h[0].End != 1260 {
		t.Fatal("touching hours must merge", h, err)
	}
	for _, bad := range [][]Hours{{{1, 420, 800}}, {{0, 0, 60}}, {{7, 1380, 1500}}, {{1, 60, 180}, {1, 120, 240}}} {
		if _, err := NormalizeHours(bad, true); err == nil {
			t.Fatal("accepted invalid hours", bad)
		}
	}
	if Contains(h, 720, 1320) || !Contains(h, 480, 540) {
		t.Fatal("containment boundary")
	}
	if MinuteTime("2026-10-11", 1440).Format("2006-01-02 15:04") != "2026-10-12 00:00" {
		t.Fatal("24:00 must advance date")
	}
	now := time.Date(2026, 10, 11, 18, 0, 0, 0, time.UTC)
	if Today(now) != "2026-10-12" {
		t.Fatal("Vietnam date")
	}
}
func TestEffectiveCalendarPriority(t *testing.T) {
	s := CalendarSnapshot{Room: Room{Status: "AVAILABLE"}, Versions: []ScheduleVersion{{ID: 1, Status: "ACTIVE", EffectiveFrom: "2026-10-01", Hours: []Hours{{7, 420, 1260}}}}, Exceptions: []CalendarException{{ID: 2, Status: "ACTIVE", Date: "2026-10-11", Mode: "REPLACE", Hours: []Hours{{0, 480, 600}}}}}
	if d := s.Day("2026-10-11"); d.ExceptionID != 2 || len(d.Hours) != 1 || d.Hours[0].Start != 480 {
		t.Fatal(d)
	}
	s.Exceptions[0].Status = "WITHDRAWN"
	if d := s.Day("2026-10-11"); d.VersionID != 1 || len(d.Hours) != 1 {
		t.Fatal(d)
	}
	s.Room.Status = "MAINTENANCE"
	if d := s.Day("2026-10-11"); len(d.Hours) != 0 || d.Source != "MAINTENANCE" {
		t.Fatal(d)
	}
}
