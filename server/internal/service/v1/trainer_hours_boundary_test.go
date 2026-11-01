package v1service

import (
	"testing"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/trainer"
)

func TestTrainerHoursMidnightBoundary(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hours []trainer.WorkingHour
		want  bool
	}{
		{"end at midnight", []trainer.WorkingHour{{DayOfWeek: 1, StartTime: "23:30", EndTime: "24:00"}}, true},
		{"start at midnight end", []trainer.WorkingHour{{DayOfWeek: 1, StartTime: "24:00", EndTime: "24:00"}}, false},
		{"invalid late end", []trainer.WorkingHour{{DayOfWeek: 1, StartTime: "23:30", EndTime: "24:30"}}, false},
		{"overlap through midnight", []trainer.WorkingHour{{DayOfWeek: 1, StartTime: "23:00", EndTime: "24:00"}, {DayOfWeek: 1, StartTime: "23:30", EndTime: "24:00"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := validHours(tc.hours); got != tc.want {
				t.Fatalf("validHours() = %v, want %v", got, tc.want)
			}
		})
	}
}
