package main

import "testing"

func TestValidateLocalDevURL(t *testing.T) {
	for _, raw := range []string{
		"postgres://demo:pass@127.0.0.1:5433/ttcsn_gym_dev?sslmode=disable",
		"postgres://demo:pass@localhost/ttcsn_gym_dev",
	} {
		if err := validateLocalDevURL(raw); err != nil {
			t.Fatalf("expected %q to be accepted: %v", raw, err)
		}
	}
	for _, raw := range []string{
		"postgres://demo:pass@127.0.0.1:5433/ttcsn_gym_test?sslmode=disable",
		"postgres://demo:pass@db.example.com/ttcsn_gym_dev?sslmode=require",
		"postgres://demo:pass@127.0.0.1:5433/postgres?sslmode=disable",
	} {
		if err := validateLocalDevURL(raw); err == nil {
			t.Fatalf("expected %q to be rejected", raw)
		}
	}
}

func TestRefreshPendingOnlyReplacesTerminalDemoRequests(t *testing.T) {
	fixture := bookingFixture{Status: "PENDING_PT"}
	for _, status := range []string{"EXPIRED", "REJECTED", "CANCELLED"} {
		if !canRefreshPending(fixture, status, true) {
			t.Fatalf("%s should allow a fresh request", status)
		}
		if canRefreshPending(fixture, status, false) {
			t.Fatalf("%s must not refresh without the flag", status)
		}
	}
	for _, status := range []string{"PENDING_PT", "CONFIRMED", "COMPLETED", "NO_SHOW"} {
		if canRefreshPending(fixture, status, true) {
			t.Fatalf("%s must be preserved", status)
		}
	}
	if canRefreshPending(bookingFixture{Status: "CONFIRMED"}, "CANCELLED", true) {
		t.Fatal("confirmed scenarios must not be recreated")
	}
}

func TestFixtureCatalogHasPlannedStateCoverage(t *testing.T) {
	if got := len(memberFixtures()); got != 12 {
		t.Fatalf("member fixture count = %d, want 12", got)
	}
	if got := len(trainerFixtures()); got != 4 {
		t.Fatalf("trainer fixture count = %d, want 4", got)
	}
	rooms := roomFixtures()
	if got := len(rooms); got != 6 {
		t.Fatalf("room fixture count = %d, want 6", got)
	}
	available := 0
	for _, room := range rooms {
		if room.Status == "AVAILABLE" {
			available++
		}
	}
	if available != 5 {
		t.Fatalf("available room count = %d, want 5", available)
	}

	want := map[string]int{
		"PENDING_PT": 6,
		"CONFIRMED":  8,
		"COMPLETED":  4,
		"NO_SHOW":    2,
		"CANCELLED":  2,
		"REJECTED":   1,
		"EXPIRED":    1,
	}
	got := make(map[string]int)
	for _, fixture := range bookingFixtures() {
		got[fixture.Status]++
	}
	for status, count := range want {
		if got[status] != count {
			t.Fatalf("fixture count for %s = %d, want %d", status, got[status], count)
		}
	}
}
