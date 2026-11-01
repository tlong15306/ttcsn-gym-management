package repository_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"
)

// Opt in because it resets ttcsn_gym_test and creates a simultaneous login burst.
func TestNetworkBookingCapacityLoad(t *testing.T) {
	if os.Getenv("RUN_NETWORK_BOOKING_LOAD") != "1" {
		t.Skip("set RUN_NETWORK_BOOKING_LOAD=1")
	}
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer slog.SetDefault(previousLogger)
	a := newManagementAPI(t)
	ctx := context.Background()
	a.sql(`INSERT INTO users(username,password_hash,email,full_name,role_id)
		SELECT 'network.load.'||i,u.password_hash,'network.load.'||i||'@test.local','Network load '||i,r.id
		FROM generate_series(1,100)i CROSS JOIN users u CROSS JOIN roles r
		WHERE u.id=$1 AND r.code='MEMBER'`, a.users["ADMIN"])
	a.sql(`INSERT INTO members(user_id,member_code,join_date,membership_expiry)
		SELECT id,username,current_date,current_date+365 FROM users WHERE username LIKE 'network.load.%'`)
	var room, slot, session int64
	if err := a.pool.QueryRow(ctx, `INSERT INTO rooms(name,room_type,default_capacity) VALUES('Network load room','GYM',80) RETURNING id`).Scan(&room); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO time_slots(name,start_time,end_time) VALUES('Network load slot','09:00','10:00') RETURNING id`).Scan(&slot); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,current_date+2,((current_date+2)+time '09:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',((current_date+2)+time '10:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',80,$3) RETURNING id`, room, slot, a.users["ADMIN"]).Scan(&session); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(a.router)
	defer server.Close()
	client := &http.Client{Timeout: 60 * time.Second}
	defer client.CloseIdleConnections()
	type result struct {
		login, read, booking time.Duration
		status               int
		err                  error
	}
	authResults := make(chan error, 100)
	results := make(chan result, 100)
	loginStart, bookingStart := make(chan struct{}), make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(100)
	for i := 1; i <= 100; i++ {
		go func(i int) {
			ready.Done()
			<-loginStart
			send := func(method, path, token, key string, body []byte) (int, []byte, time.Duration, error) {
				req, err := http.NewRequest(method, server.URL+"/api/v1"+path, bytes.NewReader(body))
				if err != nil {
					return 0, nil, 0, err
				}
				req.Header.Set("Content-Type", "application/json")
				if token != "" {
					req.Header.Set("Authorization", "Bearer "+token)
				}
				if key != "" {
					req.Header.Set("Idempotency-Key", key)
				}
				start := time.Now()
				response, err := client.Do(req)
				elapsed := time.Since(start)
				if err != nil {
					return 0, nil, elapsed, err
				}
				defer response.Body.Close()
				data, err := io.ReadAll(response.Body)
				return response.StatusCode, data, elapsed, err
			}
			loginBody, _ := json.Marshal(map[string]string{"username": fmt.Sprintf("network.load.%d", i), "password": "DemoPass123!"})
			status, body, login, err := send("POST", "/auth/login", "", "", loginBody)
			if err != nil || status != 200 {
				authResults <- fmt.Errorf("login %d: HTTP %d: %v", i, status, err)
				return
			}
			var envelope struct {
				Data struct {
					AccessToken string `json:"access_token"`
				} `json:"data"`
			}
			if err = json.Unmarshal(body, &envelope); err != nil || envelope.Data.AccessToken == "" {
				authResults <- fmt.Errorf("login %d: invalid token", i)
				return
			}
			authResults <- nil
			<-bookingStart
			status, _, read, err := send("GET", "/me", envelope.Data.AccessToken, "", nil)
			if err != nil || status != 200 {
				results <- result{err: fmt.Errorf("read %d: HTTP %d: %v", i, status, err)}
				return
			}
			bookingBody, _ := json.Marshal(map[string]int64{"room_session_id": session})
			status, body, booking, err := send("POST", "/bookings", envelope.Data.AccessToken, fmt.Sprintf("network-load-%d", i), bookingBody)
			if err != nil {
				results <- result{err: err}
				return
			}
			if status != 201 && !(status == 409 && bytes.Contains(body, []byte("CAPACITY_FULL"))) {
				results <- result{err: fmt.Errorf("booking %d: HTTP %d: %s", i, status, string(body))}
				return
			}
			results <- result{login: login, read: read, booking: booking, status: status}
		}(i)
	}
	ready.Wait()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	close(loginStart)
	failures := 0
	for i := 0; i < 100; i++ {
		if err := <-authResults; err != nil {
			t.Error(err)
			failures++
		}
	}
	bookingAt := time.Now()
	close(bookingStart)
	var loginTimes, readTimes, bookingTimes []time.Duration
	created, full := 0, 0
	for i := 0; i < 100-failures; i++ {
		r := <-results
		if r.err != nil {
			t.Error(r.err)
			continue
		}
		loginTimes = append(loginTimes, r.login)
		readTimes = append(readTimes, r.read)
		bookingTimes = append(bookingTimes, r.booking)
		if r.status == 201 {
			created++
		} else {
			full++
		}
	}
	runtime.ReadMemStats(&after)
	if t.Failed() {
		return
	}
	var holds, audits int
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM bookings WHERE room_session_id=$1 AND status='CONFIRMED'`, session).Scan(&holds); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action='booking.created' AND entity_id IN (SELECT id::text FROM bookings WHERE room_session_id=$1)`, session).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if created != 80 || full != 20 || holds != 80 || audits != 80 {
		t.Fatalf("created=%d full=%d holds=%d audits=%d", created, full, holds, audits)
	}
	percentiles := func(v []time.Duration) string {
		sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
		return fmt.Sprintf("p50=%s p95=%s p99=%s", v[49], v[94], v[98])
	}
	stats := a.pool.Stat()
	t.Logf("100 distinct users via loopback HTTP; 80 created, 20 capacity conflicts; login %s; read %s; booking %s; total=%s booking_phase=%s; heap_before=%d heap_after=%d; pool_total=%d pool_max=%d", percentiles(loginTimes), percentiles(readTimes), percentiles(bookingTimes), time.Since(start), time.Since(bookingAt), before.HeapAlloc, after.HeapAlloc, stats.TotalConns(), stats.MaxConns())
}
