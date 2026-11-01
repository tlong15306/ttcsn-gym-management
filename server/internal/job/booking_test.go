package job

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeBookings struct {
	expired, completed chan struct{}
	cancel             context.CancelFunc
}

func (f fakeBookings) Expire(context.Context) error {
	close(f.expired)
	return errors.New("synthetic failure")
}
func (f fakeBookings) Complete(context.Context) error { close(f.completed); f.cancel(); return nil }

type timeoutBookings struct{ completed chan struct{} }

func (timeoutBookings) Expire(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
func (f timeoutBookings) Complete(context.Context) error {
	close(f.completed)
	return nil
}

type pruningBookings struct{ pruned chan struct{} }

func (pruningBookings) Expire(context.Context) error   { return nil }
func (pruningBookings) Complete(context.Context) error { return nil }
func (f pruningBookings) PurgeExpiredIdempotency(context.Context) error {
	close(f.pruned)
	return nil
}

func TestBookingStartupCompletionDespiteExpirationFailureAndShutdown(t *testing.T) {
	for i := 0; i < 2; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		f := fakeBookings{make(chan struct{}), make(chan struct{}), cancel}
		done := make(chan struct{})
		go func() { RunBooking(ctx, f); close(done) }()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("startup/shutdown stalled")
		}
		select {
		case <-f.expired:
		default:
			t.Fatal("expiration not called on startup")
		}
		select {
		case <-f.completed:
		default:
			t.Fatal("completion not called")
		}
	}
}

func TestBookingCompletionGetsFreshBudgetAfterExpirationTimeout(t *testing.T) {
	ctx := context.Background()
	f := timeoutBookings{completed: make(chan struct{})}
	runBookingCycle(ctx, f, 5*time.Millisecond)
	select {
	case <-f.completed:
	default:
		t.Fatal("completion was skipped after expiration timed out")
	}
}

func TestBookingStartupPrunesExpiredIdempotencyKeys(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f := pruningBookings{pruned: make(chan struct{})}
	done := make(chan struct{})
	go func() { RunBooking(ctx, f); close(done) }()
	select {
	case <-f.pruned:
	case <-time.After(time.Second):
		t.Fatal("idempotency cleanup was not called on startup")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("booking job did not stop after cleanup")
	}
}
