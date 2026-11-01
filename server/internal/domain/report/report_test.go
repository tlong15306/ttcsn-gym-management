package report

import (
	"testing"
	"time"
)

func TestResolveRangeDefaultsToVietnamCurrentMonth(t *testing.T) {
	loc, err := time.LoadLocation(Timezone)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 14, 1, 30, 0, 0, time.UTC)
	got, err := ResolveRange(PeriodInput{}, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.From.Format("2006-01-02") != "2026-10-01" || got.To.Format("2006-01-02") != "2026-10-14" || got.From.Location().String() != loc.String() {
		t.Fatalf("unexpected local range: %+v", got)
	}
}

func TestResolveRangeAllows366InclusiveDates(t *testing.T) {
	got, err := ResolveRange(PeriodInput{From: "2025-10-15", To: "2026-10-14"}, time.Now())
	if err != nil || got.From.Format("2006-01-02") != "2025-10-15" {
		t.Fatalf("366-day inclusive range rejected: %+v %v", got, err)
	}
	for _, input := range []PeriodInput{
		{From: "2025-10-13", To: "2026-10-14"},
		{From: "2026-10-15", To: "2026-10-14"},
		{From: "2026-02-30", To: "2026-03-01"},
		{From: "2026-1-01", To: "2026-10-14"},
	} {
		if _, err := ResolveRange(input, time.Now()); err == nil {
			t.Errorf("invalid range accepted: %+v", input)
		}
	}
}
