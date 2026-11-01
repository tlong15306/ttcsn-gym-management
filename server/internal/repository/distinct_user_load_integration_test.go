package repository_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"sort"
	"sync"
	"testing"
	"time"
)

// Opt in: this uses bcrypt for 100 simultaneous logins and resets only the test DB.
func TestDistinctUserLoginAndMutationLoad(t *testing.T) {
	if os.Getenv("RUN_DISTINCT_USER_LOAD") != "1" {
		t.Skip("set RUN_DISTINCT_USER_LOAD=1 for the isolated load run")
	}
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	defer slog.SetDefault(previousLogger)
	a := newManagementAPI(t)
	a.sql(`INSERT INTO users(username,password_hash,email,full_name,role_id)
		SELECT 'load.member.'||i,u.password_hash,'load.member.'||i||'@test.local','Load member '||i,r.id
		FROM generate_series(1,100)i CROSS JOIN users u CROSS JOIN roles r
		WHERE u.id=$1 AND r.code='MEMBER'`, a.users["ADMIN"])
	type result struct {
		login, read, mutation time.Duration
		err                   error
	}
	results := make(chan result, 100)
	start := make(chan struct{})
	var ready sync.WaitGroup
	ready.Add(100)
	for i := 1; i <= 100; i++ {
		go func(i int) {
			ready.Done()
			<-start
			send := func(method, path, token string, body []byte) (*httptest.ResponseRecorder, time.Duration) {
				req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				if token != "" {
					req.Header.Set("Authorization", "Bearer "+token)
				}
				response := httptest.NewRecorder()
				now := time.Now()
				a.router.ServeHTTP(response, req)
				return response, time.Since(now)
			}
			loginBody, _ := json.Marshal(map[string]string{"username": fmt.Sprintf("load.member.%d", i), "password": "DemoPass123!"})
			response, login := send("POST", "/auth/login", "", loginBody)
			if response.Code != 200 {
				results <- result{err: fmt.Errorf("login %d: HTTP %d", i, response.Code)}
				return
			}
			var body struct {
				Data struct {
					AccessToken string `json:"access_token"`
				} `json:"data"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Data.AccessToken == "" {
				results <- result{err: fmt.Errorf("login %d: invalid token response", i)}
				return
			}
			response, read := send("GET", "/me", body.Data.AccessToken, nil)
			if response.Code != 200 {
				results <- result{err: fmt.Errorf("read %d: HTTP %d", i, response.Code)}
				return
			}
			update, _ := json.Marshal(map[string]string{"full_name": fmt.Sprintf("Updated Load Member %d", i), "email": fmt.Sprintf("load.member.%d@test.local", i), "phone": ""})
			response, mutation := send("PATCH", "/me", body.Data.AccessToken, update)
			if response.Code != 200 {
				results <- result{err: fmt.Errorf("mutation %d: HTTP %d", i, response.Code)}
				return
			}
			results <- result{login: login, read: read, mutation: mutation}
		}(i)
	}
	ready.Wait()
	started := time.Now()
	close(start)
	var logins, reads, mutations []time.Duration
	for i := 0; i < 100; i++ {
		r := <-results
		if r.err != nil {
			t.Error(r.err)
			continue
		}
		logins = append(logins, r.login)
		reads = append(reads, r.read)
		mutations = append(mutations, r.mutation)
	}
	if t.Failed() {
		return
	}
	var updated int
	if err := a.pool.QueryRow(context.Background(), `SELECT count(*) FROM users WHERE username LIKE 'load.member.%' AND full_name LIKE 'Updated Load Member %'`).Scan(&updated); err != nil || updated != 100 {
		t.Fatalf("updated=%d error=%v", updated, err)
	}
	p95 := func(values []time.Duration) time.Duration {
		sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
		return values[94]
	}
	t.Logf("100 distinct identities, 100 logins + 100 reads + 100 profile mutations: total=%s, login p95=%s, read p95=%s, mutation p95=%s", time.Since(started), p95(logins), p95(reads), p95(mutations))
}
