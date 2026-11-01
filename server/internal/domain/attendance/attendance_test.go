package attendance

import (
	"testing"
	"time"
)

func TestWindowBoundaries(t *testing.T) {
	start := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	for _, tc := range []struct {
		at   time.Time
		want bool
	}{{start.Add(-15*time.Minute - time.Nanosecond), false}, {start.Add(-15 * time.Minute), true}, {end.Add(-time.Nanosecond), true}, {end, false}} {
		if InWindow(tc.at, start, end) != tc.want {
			t.Fatalf("boundary %s", tc.at)
		}
	}
}
