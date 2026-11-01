package repository_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/member"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/trainer"
	v1handler "github.com/tlong15306/ttcsn-gym-management/server/internal/handler/v1"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/middleware"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	v1routes "github.com/tlong15306/ttcsn-gym-management/server/internal/routes/v1"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"github.com/tlong15306/ttcsn-gym-management/server/pkg/auth"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type testAPI struct {
	t      *testing.T
	router http.Handler
	pool   *pgxpool.Pool
	tokens map[string]string
	users  map[string]int64
}

func newManagementAPI(t *testing.T) *testAPI {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL for ttcsn_gym_test")
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil || strings.TrimPrefix(parsed.Path, "/") != "ttcsn_gym_test" {
		t.Fatal("TEST_DATABASE_URL must point exactly to ttcsn_gym_test")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	goose.SetDialect("postgres")
	migrations := filepath.Join("..", "db", "migrations")
	if err = goose.DownTo(db, migrations, 0); err != nil && !strings.Contains(err.Error(), "no current version") {
		t.Fatal(err)
	}
	if err = goose.Up(db, migrations); err != nil {
		t.Fatal(err)
	}
	if err = goose.Up(db, migrations); err != nil {
		t.Fatal("repeat migration", err)
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	passwords := auth.NewPasswordService(10)
	hash, err := passwords.Hash("DemoPass123!")
	if err != nil {
		t.Fatal(err)
	}
	api := &testAPI{t: t, pool: pool, tokens: map[string]string{}, users: map[string]int64{}}
	for _, role := range []string{"ADMIN", "MANAGER", "RECEPTIONIST", "MEMBER", "PT"} {
		var id int64
		if err = pool.QueryRow(context.Background(), `INSERT INTO users(username,password_hash,email,full_name,role_id) SELECT $1,$2,$3,$4,id FROM roles WHERE code=$5 RETURNING users.id`, strings.ToLower(role)+".test", hash, strings.ToLower(role)+"@test.local", role+" Test", role).Scan(&id); err != nil {
			t.Fatal(err)
		}
		api.users[role] = id
	}
	users := repository.NewIdentityRepository(pool)
	tokens := auth.NewTokenService("01234567890123456789012345678901", 15*time.Minute)
	identityService := v1service.NewService(users, passwords, tokens)
	management := v1handler.NewManagementHandler(v1service.NewMemberService(repository.NewMemberRepository(pool), passwords), v1service.NewFacilityService(repository.NewFacilityRepository(pool)), v1service.NewTrainerService(repository.NewTrainerRepository(pool), t.TempDir()))
	gin.SetMode(gin.TestMode)
	router, err := v1routes.New(v1routes.Dependencies{Readiness: pool, IdentityHandler: v1handler.NewIdentityHandler(identityService, 900), ManagementHandler: management, CalendarHandler: v1handler.NewCalendarHandler(v1service.NewCalendarService(repository.NewCalendarRepository(pool))), AttendanceHandler: v1handler.NewAttendanceHandler(v1service.NewAttendanceService(repository.NewAttendanceRepository(pool))), BookingHandler: v1handler.NewBookingHandler(v1service.NewBookingService(repository.NewBookingRepository(pool))), ReviewHandler: v1handler.NewReviewHandler(v1service.NewReviewService(repository.NewReviewRepository(pool))), ReportHandler: v1handler.NewReportHandler(v1service.NewReportService(repository.NewReportRepository(pool))), RoomCoverHandler: v1handler.NewRoomCoverHandler(v1service.NewRoomCoverService(repository.NewRoomCoverRepository(pool))), Authenticator: middleware.NewAuthenticator(tokens, users)})
	if err != nil {
		t.Fatal(err)
	}
	api.router = router
	for role := range api.users {
		result := api.call("POST", "/auth/login", "", map[string]any{"username": strings.ToLower(role) + ".test", "password": "DemoPass123!"}, 200)
		var data v1service.LoginResult
		decodeData(t, result, &data)
		if data.User.Role != identity.Role(role) {
			t.Fatal("wrong login role")
		}
		api.tokens[role] = data.AccessToken
	}
	return api
}
func (a *testAPI) call(method, path, role string, input any, want int) *httptest.ResponseRecorder {
	a.t.Helper()
	var body []byte
	if input != nil {
		body, _ = json.Marshal(input)
	}
	req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "management-integration")
	if role != "" {
		req.Header.Set("Authorization", "Bearer "+a.tokens[role])
	}
	w := httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	if w.Code != want {
		a.t.Fatalf("%s %s => %d, want %d: %s", method, path, w.Code, want, w.Body.String())
	}
	if want != 204 && strings.Contains(w.Header().Get("Content-Type"), "json") && !strings.Contains(w.Body.String(), "request_id") {
		a.t.Fatal("missing request_id")
	}
	if strings.Contains(w.Body.String(), "password_hash") || strings.Contains(w.Body.String(), "certificate_path") {
		a.t.Fatal("private field leaked")
	}
	return w
}
func decodeData(t *testing.T, w *httptest.ResponseRecorder, dest any) {
	t.Helper()
	var envelope struct{ Data json.RawMessage }
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(envelope.Data, dest); err != nil {
		t.Fatal(err)
	}
}
func (a *testAPI) sql(query string, args ...any) {
	a.t.Helper()
	if _, err := a.pool.Exec(context.Background(), query, args...); err != nil {
		a.t.Fatal(err)
	}
}
func TestManagementAPIIntegration(t *testing.T) {
	a := newManagementAPI(t)
	t.Run("account permissions and audit", func(t *testing.T) {
		for _, role := range []string{"MANAGER", "RECEPTIONIST", "MEMBER", "PT"} {
			a.call("GET", "/users", ""+role, nil, 403)
			a.call("PATCH", fmt.Sprint("/users/", a.users["MEMBER"]), role, map[string]string{"full_name": "Other", "email": "other@test.local"}, 403)
		}
		path := fmt.Sprint("/users/", a.users["MEMBER"])
		a.call("GET", path, "ADMIN", nil, 200)
		a.call("PATCH", path, "ADMIN", map[string]string{"full_name": "Member Edited", "email": "edited@test.local", "phone": "0123456789"}, 200)
		var count int
		a.pool.QueryRow(context.Background(), `SELECT count(*) FROM audit_events WHERE action='account.profile_updated' AND request_id='management-integration'`).Scan(&count)
		if count != 1 {
			t.Fatal("missing account audit")
		}
	})
	var m member.Member
	t.Run("member transactions and object scope", func(t *testing.T) {
		input := map[string]any{"member_code": "M-TEST", "join_date": "2026-10-01", "membership_expiry": "2027-10-01", "user_id": a.users["MEMBER"]}
		a.call("POST", "/members", "MEMBER", input, 403)
		decodeData(t, a.call("POST", "/members", "RECEPTIONIST", input, 201), &m)
		if m.ID == 0 || m.User.FullName != "Member Edited" {
			t.Fatal("incorrect member projection")
		}
		a.call("POST", "/members", "MANAGER", input, 409)
		a.call("GET", fmt.Sprint("/members/", m.ID), "MEMBER", nil, 200)
		a.call("PATCH", fmt.Sprint("/members/", m.ID), "MEMBER", map[string]string{"full_name": "Self", "email": "self@test.local"}, 403)
		a.call("PATCH", fmt.Sprint("/members/", m.ID, "/membership"), "MANAGER", map[string]string{"membership_expiry": "2026-09-01", "status": "INACTIVE", "reason": "invalid"}, 422)
		a.call("PATCH", fmt.Sprint("/members/", m.ID, "/membership"), "RECEPTIONIST", map[string]string{"membership_expiry": "2028-10-01", "status": "INACTIVE", "reason": "Membership updated"}, 200)
		a.call("PATCH", fmt.Sprint("/members/", m.ID, "/membership"), "MANAGER", map[string]string{"membership_expiry": "2028-10-01", "status": "ACTIVE", "reason": "Membership resumed"}, 200)
		create := map[string]any{"member_code": "M-NEW", "join_date": "2026-10-01", "membership_expiry": "2027-10-01", "account": map[string]string{"username": "member.new", "password": "DemoPass123!", "email": "new@test.local", "full_name": "New Member"}}
		var created member.Member
		decodeData(t, a.call("POST", "/members", "RECEPTIONIST", create, 201), &created)
		if created.User.Role != identity.RoleMember {
			t.Fatal("member role not forced")
		}
		a.call("GET", fmt.Sprint("/members/", created.ID), "MEMBER", nil, 404)
		account := create["account"].(map[string]string)
		account["username"] = "rollback.member"
		account["email"] = "rollback@test.local"
		create["member_code"] = "M-TEST"
		a.call("POST", "/members", "MANAGER", create, 409)
		var count int
		a.pool.QueryRow(context.Background(), `SELECT count(*) FROM users WHERE username='rollback.member'`).Scan(&count)
		if count != 0 {
			t.Fatal("account creation survived profile rollback")
		}
		create["member_code"] = "M-AUDIT"
		a.sql(`CREATE FUNCTION fail_member_audit() RETURNS trigger LANGUAGE plpgsql AS $$BEGIN IF NEW.action='member.created' THEN RAISE EXCEPTION 'fixture audit failure'; END IF; RETURN NEW; END$$`)
		a.sql(`CREATE TRIGGER fail_member_audit BEFORE INSERT ON audit_events FOR EACH ROW EXECUTE FUNCTION fail_member_audit()`)
		a.call("POST", "/members", "MANAGER", create, 500)
		a.sql(`DROP TRIGGER fail_member_audit ON audit_events`)
		a.sql(`DROP FUNCTION fail_member_audit()`)
		a.pool.QueryRow(context.Background(), `SELECT count(*) FROM users WHERE username='rollback.member'`).Scan(&count)
		if count != 0 {
			t.Fatal("account survived audit failure")
		}
		a.call("GET", "/members/eligible-accounts", "MANAGER", nil, 200)
		a.call("GET", "/members/eligible-accounts", "PT", nil, 403)
	})
	var room facility.Room
	var slot facility.TimeSlot
	var session facility.Session
	t.Run("facilities and concurrent slot exclusion", func(t *testing.T) {
		input := map[string]any{"name": "Gym Test", "room_type": "Gym", "default_capacity": 3}
		a.call("POST", "/rooms", "RECEPTIONIST", input, 403)
		decodeData(t, a.call("POST", "/rooms", "MANAGER", input, 201), &room)
		input["name"] = "gym test"
		a.call("POST", "/rooms", "ADMIN", input, 409)
		decodeData(t, a.call("POST", "/time-slots", "MANAGER", map[string]string{"name": "Morning", "start_time": "08:00", "end_time": "09:00"}, 201), &slot)
		a.call("POST", "/time-slots", "ADMIN", map[string]string{"name": "Overlap", "start_time": "08:59", "end_time": "10:00"}, 409)
		a.call("POST", "/time-slots", "ADMIN", map[string]string{"name": "Adjacent", "start_time": "09:00", "end_time": "10:00"}, 201)
		codes := make(chan int, 2)
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				b, _ := json.Marshal(map[string]string{"name": fmt.Sprint("Concurrent", i), "start_time": "10:00", "end_time": "11:00"})
				req := httptest.NewRequest("POST", "/api/v1/time-slots", bytes.NewReader(b))
				req.Header.Set("Authorization", "Bearer "+a.tokens["ADMIN"])
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				a.router.ServeHTTP(w, req)
				codes <- w.Code
			}(i)
		}
		wg.Wait()
		close(codes)
		success, conflict := 0, 0
		for code := range codes {
			if code == 201 {
				success++
			} else if code == 409 {
				conflict++
			} else {
				t.Fatalf("concurrent slot status=%d", code)
			}
		}
		if success != 1 || conflict != 1 {
			t.Fatal("concurrent exclusion failed")
		}
		day := time.Now().AddDate(0, 0, 14).Format("2006-01-02")
		// Explicit full-day fixture; legacy CRUD tests do not infer a schedule from sessions.
		a.sql(`WITH v AS (INSERT INTO room_schedule_versions(room_id,effective_from,published_by_user_id,reason,payload_hash) VALUES($1,CURRENT_DATE,$2,'CRUD test fixture',repeat('0',64)) RETURNING id) INSERT INTO room_weekly_hours(version_id,weekday,start_minute,end_minute) SELECT v.id,d,0,1440 FROM v CROSS JOIN generate_series(1,7) d`, room.ID, a.users["ADMIN"])
		decodeData(t, a.call("POST", "/room-sessions", "ADMIN", map[string]any{"room_id": room.ID, "time_slot_id": slot.ID, "session_date": day}, 201), &session)
		if session.Capacity != 3 || session.ID == 0 {
			t.Fatal("invalid capacity snapshot")
		}
		start, _ := time.Parse(time.RFC3339, session.StartsAt)
		if start.In(time.FixedZone("Vietnam", 7*3600)).Hour() != 8 {
			t.Fatal("invalid session timezone")
		}
		a.call("PATCH", fmt.Sprint("/time-slots/", slot.ID), "ADMIN", map[string]string{"name": "Edited", "start_time": "07:00", "end_time": "08:00", "reason": "in use"}, 409)
		a.call("PATCH", fmt.Sprint("/time-slots/", slot.ID, "/status"), "ADMIN", map[string]any{"is_active": false, "reason": "retired"}, 200)
		a.call("POST", "/room-sessions", "ADMIN", map[string]any{"room_id": room.ID, "time_slot_id": slot.ID, "session_date": time.Now().AddDate(0, 0, 15).Format("2006-01-02")}, 409)
	})
	var pt trainer.Profile
	t.Run("PT submission certificates and decisions", func(t *testing.T) {
		a.call("PUT", "/pts/me/profile", "ADMIN", map[string]string{"specialty": "Gym", "experience_text": "Experience"}, 403)
		decodeData(t, a.call("PUT", "/pts/me/profile", "PT", map[string]string{"specialty": "Strength", "experience_text": "3 years"}, 200), &pt)
		a.call("GET", fmt.Sprint("/pts/", pt.ID), "MEMBER", nil, 404)
		a.call("POST", "/pts/me/submit", "PT", nil, 409)
		var pngData bytes.Buffer
		png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 1, 1)))
		upload := func(content []byte, want int) {
			t.Helper()
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			part, _ := writer.CreateFormFile("file", "test.png")
			part.Write(content)
			writer.Close()
			req := httptest.NewRequest("POST", "/api/v1/pts/me/certificates", &body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer "+a.tokens["PT"])
			w := httptest.NewRecorder()
			a.router.ServeHTTP(w, req)
			if w.Code != want {
				t.Fatalf("upload=%d want=%d %s", w.Code, want, w.Body.String())
			}
		}
		upload([]byte("<html>fake certificate</html>"), 422)
		upload(bytes.Repeat([]byte("x"), v1service.MaxCertificateBytes+1), 422)
		upload(pngData.Bytes(), 200)
		a.call("GET", fmt.Sprint("/pts/", pt.ID, "/certificate"), "MEMBER", nil, 403)
		a.call("GET", fmt.Sprint("/pts/", pt.ID, "/certificate"), "", nil, 401)
		a.call("GET", fmt.Sprint("/pts/", pt.ID, "/certificate"), "MANAGER", nil, 200)
		a.call("POST", "/pts/me/submit", "PT", nil, 200)
		a.call("PUT", "/pts/me/profile", "PT", map[string]string{"specialty": "No", "experience_text": "No"}, 409)
		a.call("POST", fmt.Sprint("/pts/", pt.ID, "/decision"), "RECEPTIONIST", map[string]string{"action": "APPROVE", "reason": "no permission"}, 403)
		a.call("POST", fmt.Sprint("/pts/", pt.ID, "/decision"), "MANAGER", map[string]string{"action": "REJECT", "reason": "Please revise"}, 200)
		a.call("PUT", "/pts/me/profile", "PT", map[string]string{"specialty": "Strength revised", "experience_text": "Updated experience"}, 200)
		a.call("POST", "/pts/me/submit", "PT", nil, 200)
		a.call("POST", fmt.Sprint("/pts/", pt.ID, "/decision"), "ADMIN", map[string]string{"action": "APPROVE", "reason": "Checked"}, 200)
		a.call("GET", fmt.Sprint("/pts/", pt.ID), "MEMBER", nil, 200)
		var publicProfile trainer.Profile
		decodeData(t, a.call("GET", fmt.Sprint("/pts/", pt.ID), "MEMBER", nil, 200), &publicProfile)
		if publicProfile.DecisionReason != "" || publicProfile.DecidedAt != nil {
			t.Fatal("internal decision leaked")
		}
		a.call("GET", "/pts", "ADMIN", nil, 200)
		a.call("GET", "/pts?approved=true", "MEMBER", nil, 200)
		a.call("GET", "/pts?available_at=invalid", "ADMIN", nil, 422)
	})
	t.Run("holds protect capacity and PT schedule", func(t *testing.T) {
		a.sql(`INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,pt_id,status,created_by_user_id) VALUES($1,$2,$3::date,$4,$5,'CONFIRMED',$6)`, m.ID, session.ID, session.SessionDate, slot.ID, pt.ID, a.users["ADMIN"])
		// A second member fixture holds another seat.
		var otherMember int64
		a.pool.QueryRow(context.Background(), `SELECT id FROM members WHERE member_code='M-NEW'`).Scan(&otherMember)
		a.sql(`INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,status,created_by_user_id) VALUES($1,$2,$3::date,$4,'CONFIRMED',$5)`, otherMember, session.ID, session.SessionDate, slot.ID, a.users["ADMIN"])
		a.call("PATCH", fmt.Sprint("/room-sessions/", session.ID), "ADMIN", map[string]any{"capacity": 1, "status": "OPEN", "reason": "too small"}, 409)
		a.call("PATCH", fmt.Sprint("/room-sessions/", session.ID), "MANAGER", map[string]any{"capacity": 3, "status": "CLOSED", "reason": "holds"}, 409)
		impact := a.call("GET", fmt.Sprint("/rooms/", room.ID, "/impact"), "ADMIN", nil, 200)
		var value facility.Impact
		decodeData(t, impact, &value)
		if value.ActiveHolds != 2 {
			t.Fatal("incorrect impact")
		}
		a.call("PATCH", fmt.Sprint("/rooms/", room.ID), "ADMIN", map[string]any{"name": room.Name, "room_type": room.RoomType, "default_capacity": 3, "status": "MAINTENANCE", "reason": "holds"}, 409)
		a.call("GET", fmt.Sprint("/members/", m.ID), "PT", nil, 200)
		var assigned []member.Member
		decodeData(t, a.call("GET", "/members?q="+url.QueryEscape(m.MemberCode), "PT", nil, 200), &assigned)
		if len(assigned) != 1 || assigned[0].ID != m.ID || assigned[0].User.Email != "" || assigned[0].User.Phone != "" {
			t.Fatal("assigned member search must preserve PT scope and redact contacts")
		}
		for _, query := range []string{"edited@test.local", "0123456789"} {
			decodeData(t, a.call("GET", "/members?q="+url.QueryEscape(query), "ADMIN", nil, 200), &assigned)
			if len(assigned) != 1 || assigned[0].ID != m.ID {
				t.Fatal("operator contact search must find member")
			}
			decodeData(t, a.call("GET", "/members?q="+url.QueryEscape(query), "PT", nil, 200), &assigned)
			if len(assigned) != 0 {
				t.Fatal("PT search must not infer redacted member contacts")
			}
		}
		start, _ := time.Parse(time.RFC3339, session.StartsAt)
		dow := int(start.In(time.FixedZone("Vietnam", 7*3600)).Weekday())
		if dow == 0 {
			dow = 7
		}
		hours := map[string]any{"hours": []map[string]any{{"day_of_week": dow, "start_time": "08:00", "end_time": "09:00"}}}
		a.call("PUT", "/pts/me/availability", "PT", hours, 200)
		// The held hour is excluded; the adjacent end boundary is also unavailable.
		var available []trainer.Profile
		decodeData(t, a.call("GET", "/pts?available_at="+url.QueryEscape(session.StartsAt), "MEMBER", nil, 200), &available)
		if len(available) != 0 {
			t.Fatal("booked PT offered as available")
		}
		decodeData(t, a.call("GET", "/pts?available_at="+url.QueryEscape(session.EndsAt), "MEMBER", nil, 200), &available)
		if len(available) != 0 {
			t.Fatal("end boundary must be exclusive")
		}
		a.call("PUT", "/pts/me/availability", "PT", map[string]any{"hours": []any{}}, 409)
		a.call("PUT", fmt.Sprint("/pts/", pt.ID, "/availability"), "MEMBER", hours, 403)
		a.call("POST", "/pts/me/unavailability", "PT", map[string]string{"starts_at": session.StartsAt, "ends_at": session.EndsAt, "reason": "conflicts"}, 409)
		a.call("POST", fmt.Sprint("/pts/", pt.ID, "/decision"), "MANAGER", map[string]string{"action": "SUSPEND", "reason": "has holds"}, 409)
		absenceInput := map[string]string{"starts_at": "2026-12-01T12:00:00+07:00", "ends_at": "2026-12-01T13:00:00+07:00", "reason": "leave"}
		var absence trainer.Unavailability
		decodeData(t, a.call("POST", "/pts/me/unavailability", "PT", absenceInput, 201), &absence)
		a.call("GET", "/pts/me/unavailability?page=1&page_size=20", "PT", nil, 200)
		a.call("GET", "/pts/me/unavailability?page_size=101", "PT", nil, 422)
		a.call("POST", "/pts/me/unavailability", "PT", absenceInput, 409)
		a.call("DELETE", fmt.Sprint("/pts/", pt.ID, "/unavailability/", absence.ID), "MEMBER", nil, 403)
		a.call("DELETE", fmt.Sprint("/pts/me/unavailability/", absence.ID), "PT", nil, 204)
		a.sql(`UPDATE bookings SET status='COMPLETED'`)
		decodeData(t, a.call("GET", "/pts?available_at="+url.QueryEscape(session.StartsAt), "MEMBER", nil, 200), &available)
		if len(available) != 1 || available[0].ID != pt.ID {
			t.Fatal("available PT missing")
		}
		a.call("POST", fmt.Sprint("/pts/", pt.ID, "/decision"), "MANAGER", map[string]string{"action": "SUSPEND", "reason": "Checked"}, 200)
		a.call("PATCH", fmt.Sprint("/rooms/", room.ID), "ADMIN", map[string]any{"name": room.Name, "room_type": room.RoomType, "default_capacity": 3, "status": "MAINTENANCE", "reason": "maintenance"}, 200)
	})
	t.Run("old tokens revoked by account lifecycle", func(t *testing.T) {
		path := fmt.Sprint("/users/", a.users["MEMBER"])
		a.call("PATCH", path+"/status", "ADMIN", map[string]string{"status": "LOCKED", "reason": "test"}, 200)
		a.call("GET", "/me", "MEMBER", nil, 401)
		a.call("PATCH", path+"/status", "ADMIN", map[string]string{"status": "ACTIVE", "reason": "test"}, 200)
		a.call("GET", "/me", "MEMBER", nil, 401)
		var loggedIn v1service.LoginResult
		decodeData(t, a.call("POST", "/auth/login", "", map[string]string{"username": "member.test", "password": "DemoPass123!"}, 200), &loggedIn)
		a.tokens["MEMBER"] = loggedIn.AccessToken
		a.call("POST", path+"/reset-password", "ADMIN", map[string]string{"password": "FreshPass123!"}, 200)
		a.call("GET", "/me", "MEMBER", nil, 401)
		decodeData(t, a.call("POST", "/auth/login", "", map[string]string{"username": "member.test", "password": "FreshPass123!"}, 200), &loggedIn)
		a.tokens["MEMBER"] = loggedIn.AccessToken
		a.call("PATCH", path+"/role", "ADMIN", map[string]string{"role": "RECEPTIONIST", "reason": "reassigned"}, 200)
		a.call("GET", "/me", "MEMBER", nil, 401)
		a.call("PATCH", fmt.Sprint("/users/", a.users["ADMIN"], "/role"), "ADMIN", map[string]string{"role": "MANAGER", "reason": "last admin"}, 409)
		a.call("POST", "/auth/logout", "PT", nil, 204)
		a.call("GET", "/me", "PT", nil, 401)
	})
}

func TestLastAdminConcurrentIntegration(t *testing.T) {
	a := newManagementAPI(t)
	var second identity.User
	decodeData(t, a.call("POST", "/users", "ADMIN", map[string]string{"username": "admin.second", "password": "DemoPass123!", "email": "admin.second@test.local", "full_name": "Second Admin", "role": "ADMIN"}, 201), &second)
	repo := repository.NewIdentityRepository(a.pool)
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, id := range []int64{a.users["ADMIN"], second.ID} {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			_, err := repo.SetStatus(context.Background(), id, identity.StatusLocked, "concurrent last admin", a.users["ADMIN"], "concurrent-test")
			results <- err
		}(id)
	}
	wg.Wait()
	close(results)
	success, blocked := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, identity.ErrLastActiveAdmin) {
			blocked++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || blocked != 1 {
		t.Fatal("last-admin invariant failed", success, blocked)
	}
	var remaining int
	if err := a.pool.QueryRow(context.Background(), `SELECT count(*) FROM users u JOIN roles r ON r.id=u.role_id WHERE r.code='ADMIN' AND u.status='ACTIVE'`).Scan(&remaining); err != nil || remaining != 1 {
		t.Fatal("active admin count", remaining, err)
	}
}
