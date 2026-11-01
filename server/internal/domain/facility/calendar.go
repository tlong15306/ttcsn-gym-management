package facility

import (
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"sort"
	"time"
)

var Vietnam = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)

func Today(now time.Time) string { return now.In(Vietnam).Format("2006-01-02") }

type Hours struct {
	Weekday int `json:"weekday"`
	Start   int `json:"start_minute"`
	End     int `json:"end_minute"`
}

// NormalizeHours merges touching intervals but rejects overlaps and non-hour boundaries.
func NormalizeHours(hours []Hours, weekly bool) ([]Hours, error) {
	out := append([]Hours{}, hours...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Weekday != out[j].Weekday {
			return out[i].Weekday < out[j].Weekday
		}
		return out[i].Start < out[j].Start
	})
	merged := []Hours{}
	for _, h := range out {
		if h.Start < 0 || h.End > 1440 || h.Start >= h.End || h.Start%60 != 0 || h.End%60 != 0 || (weekly && (h.Weekday < 1 || h.Weekday > 7)) || (!weekly && h.Weekday != 0) {
			return nil, identity.ErrValidation
		}
		if len(merged) > 0 {
			p := &merged[len(merged)-1]
			if p.Weekday == h.Weekday {
				if h.Start < p.End {
					return nil, identity.ErrValidation
				}
				if h.Start == p.End {
					p.End = h.End
					continue
				}
			}
		}
		merged = append(merged, h)
	}
	return merged, nil
}
func Contains(hours []Hours, start, end int) bool {
	for _, h := range hours {
		if h.Start <= start && h.End >= end {
			return true
		}
	}
	return false
}
func MinuteTime(date string, minute int) time.Time {
	d, _ := time.ParseInLocation("2006-01-02", date, Vietnam)
	return d.Add(time.Duration(minute) * time.Minute)
}

type CalendarMutation struct {
	Kind          string  `json:"kind"`
	BaseRevision  int64   `json:"base_revision"`
	EffectiveFrom string  `json:"effective_from,omitempty"`
	Date          string  `json:"date,omitempty"`
	Mode          string  `json:"mode,omitempty"`
	ClosureType   string  `json:"closure_type,omitempty"`
	ExceptionID   int64   `json:"exception_id,omitempty"`
	Hours         []Hours `json:"hours"`
	Reason        string  `json:"reason"`
	Days          int     `json:"days,omitempty"`
}
type CalendarConfirm struct {
	CalendarMutation
	PreviewID string `json:"preview_id"`
}
type ScheduleVersion struct {
	ID             int64   `json:"id"`
	RoomID         int64   `json:"room_id"`
	EffectiveFrom  string  `json:"effective_from"`
	EffectiveUntil string  `json:"effective_until"`
	Hours          []Hours `json:"hours"`
	Reason         string  `json:"reason"`
	CreatedAt      string  `json:"created_at"`
	Status         string  `json:"status"`
}
type CalendarException struct {
	ID          int64   `json:"id"`
	RoomID      int64   `json:"room_id"`
	Date        string  `json:"date"`
	Mode        string  `json:"mode"`
	ClosureType string  `json:"closure_type"`
	Reason      string  `json:"reason"`
	Status      string  `json:"status"`
	Hours       []Hours `json:"hours"`
	UpdatedAt   string  `json:"updated_at"`
}
type CalendarSnapshot struct {
	Room       Room                `json:"room"`
	Versions   []ScheduleVersion   `json:"versions"`
	Exceptions []CalendarException `json:"exceptions"`
	Slots      []TimeSlot          `json:"slots"`
	Sessions   []Session           `json:"sessions"`
}
type CalendarDay struct {
	Date        string  `json:"date"`
	Source      string  `json:"source"`
	VersionID   int64   `json:"version_id"`
	ExceptionID int64   `json:"exception_id"`
	Hours       []Hours `json:"hours"`
}

func (s CalendarSnapshot) Day(date string) CalendarDay {
	d := CalendarDay{Date: date, Source: "UNCONFIGURED", Hours: []Hours{}}
	if s.Room.Status != "AVAILABLE" {
		d.Source = s.Room.Status
		return d
	}
	for _, e := range s.Exceptions {
		if e.Date == date && e.Status == "ACTIVE" {
			d.Source = e.Mode
			d.ExceptionID = e.ID
			if e.Mode == "REPLACE" {
				d.Hours = e.Hours
			}
			return d
		}
	}
	parsed, _ := time.Parse("2006-01-02", date)
	weekday := int(parsed.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	for _, v := range s.Versions {
		if v.Status == "ACTIVE" && v.EffectiveFrom <= date && (v.EffectiveUntil == "" || date < v.EffectiveUntil) {
			d.Source = "WEEKLY"
			d.VersionID = v.ID
			for _, h := range v.Hours {
				if h.Weekday == weekday {
					d.Hours = append(d.Hours, h)
				}
			}
			return d
		}
	}
	return d
}

type CalendarChange struct {
	SessionID int64  `json:"session_id"`
	Date      string `json:"date"`
	Start     string `json:"starts_at"`
	End       string `json:"ends_at"`
	Holds     int    `json:"active_holds"`
	Action    string `json:"action"`
}
type CalendarPreview struct {
	ID              string           `json:"preview_id"`
	ExpiresAt       time.Time        `json:"expires_at"`
	Revision        int64            `json:"base_revision"`
	Changes         []CalendarChange `json:"changes"`
	Conflicts       []CalendarChange `json:"conflicts"`
	Creates         int              `json:"creates"`
	PreparedThrough string           `json:"prepared_through"`
}
type CalendarResult struct {
	Revision        int64  `json:"calendar_revision"`
	Created         int    `json:"created"`
	Closed          int    `json:"closed"`
	Reopened        int    `json:"reopened"`
	PreparedThrough string `json:"prepared_through"`
}
