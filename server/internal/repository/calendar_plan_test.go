package repository

import (
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"testing"
	"time"
)

func TestCalendarPlanProtectsSnapshotsAndScope(t *testing.T) {
	now := facility.MinuteTime("2026-10-11", 600)
	mk := func(id int64, date string, start int, closure string) facility.Session {
		status := "OPEN"
		if closure != "" {
			status = "CLOSED"
		}
		return facility.Session{ID: id, RoomID: 1, TimeSlotID: id, SessionDate: date, StartsAt: facility.MinuteTime(date, start).Format(time.RFC3339), EndsAt: facility.MinuteTime(date, start+60).Format(time.RFC3339), Status: status, ClosureSource: closure}
	}
	s := facility.CalendarSnapshot{Room: facility.Room{ID: 1, Status: "AVAILABLE"}, Versions: []facility.ScheduleVersion{{ID: 1, Status: "ACTIVE", EffectiveFrom: "2026-10-01", Hours: []facility.Hours{{Weekday: 7, Start: 0, End: 1440}, {Weekday: 1, Start: 0, End: 1440}}}}, Sessions: []facility.Session{mk(1, "2026-10-11", 540, ""), mk(2, "2026-10-11", 660, ""), mk(3, "2026-10-12", 660, ""), mk(4, "2026-10-12", 780, "MANUAL")}}
	p, err := planCalendar(s, facility.CalendarMutation{Kind: "WEEKLY", EffectiveFrom: "2026-10-12"}, now, false)
	if err != nil || len(p.closes) != 1 || p.closes[0] != 3 {
		t.Fatal("ongoing or outside scope changed", p, err)
	}
	p, err = planCalendar(s, facility.CalendarMutation{Kind: "EXTEND"}, now, true)
	if err != nil || len(p.closes) != 0 || len(p.reopens) != 0 {
		t.Fatal("job changes snapshots", p, err)
	}
	p, err = planCalendar(s, facility.CalendarMutation{Kind: "EXCEPTION", Date: "2026-10-11", Mode: "CLOSED"}, now, false)
	if err != nil || len(p.closes) != 1 || p.closes[0] != 2 {
		t.Fatal("exception scope", p, err)
	}
}
