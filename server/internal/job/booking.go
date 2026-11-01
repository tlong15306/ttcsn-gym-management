package job

import (
	"context"
	"log/slog"
	"time"
)

type BookingExpirer interface {
	Expire(context.Context) error
	Complete(context.Context) error
}

type idempotencyPruner interface {
	PurgeExpiredIdempotency(context.Context) error
}

func RunBooking(ctx context.Context, repo BookingExpirer) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	pruneTicker := time.NewTicker(time.Hour)
	defer pruneTicker.Stop()
	runBookingMaintenance(ctx, repo, 12*time.Second)
	for {
		runBookingCycle(ctx, repo, 12*time.Second)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-pruneTicker.C:
			runBookingMaintenance(ctx, repo, 12*time.Second)
		}
	}
}

func runBookingCycle(ctx context.Context, repo BookingExpirer, stageTimeout time.Duration) {
	if err := runBookingStage(ctx, stageTimeout, repo.Expire); err != nil {
		if ctx.Err() == nil {
			slog.Error("booking expiration failed", "code", "BOOKING_JOB_FAILED")
		}
	}
	// A timed-out expiration pass must not starve completion. A shutdown still
	// cancels the parent context and stops the second stage promptly.
	if ctx.Err() != nil {
		return
	}
	if err := runBookingStage(ctx, stageTimeout, repo.Complete); err != nil && ctx.Err() == nil {
		slog.Error("booking completion failed", "code", "BOOKING_JOB_FAILED")
	}
}

func runBookingStage(ctx context.Context, timeout time.Duration, run func(context.Context) error) error {
	stage, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return run(stage)
}

func runBookingMaintenance(ctx context.Context, repo BookingExpirer, stageTimeout time.Duration) {
	pruner, ok := repo.(idempotencyPruner)
	if !ok {
		return
	}
	if err := runBookingStage(ctx, stageTimeout, pruner.PurgeExpiredIdempotency); err != nil && ctx.Err() == nil {
		slog.Error("idempotency cleanup failed", "code", "IDEMPOTENCY_PURGE_FAILED")
	}
}
