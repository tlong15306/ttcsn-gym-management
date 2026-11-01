package job

import (
	"context"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"log/slog"
	"time"
)

type CalendarMaterializer interface{ MaterializeAll(context.Context) error }

// Run performs startup catch-up, then prepares the next 28 days at 00:05 Vietnam.
func RunCalendar(ctx context.Context, repo CalendarMaterializer) {
	for {
		runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		err := repo.MaterializeAll(runCtx)
		cancel()
		if err != nil && ctx.Err() == nil {
			slog.Error("calendar materialization failed", "code", "CALENDAR_JOB_FAILED")
		}
		now := time.Now().In(facility.Vietnam)
		next := time.Date(now.Year(), now.Month(), now.Day(), 0, 5, 0, 0, facility.Vietnam)
		if !next.After(now) {
			next = next.AddDate(0, 0, 1)
		}
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
