package repository_test

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/discovery"
	v1handler "github.com/tlong15306/ttcsn-gym-management/server/internal/handler/v1"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	v1routes "github.com/tlong15306/ttcsn-gym-management/server/internal/routes/v1"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
)

func TestAnonymousDiscoveryProjectionAndBoundaries(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	router, err := v1routes.New(v1routes.Dependencies{Readiness: a.pool, DiscoveryHandler: v1handler.NewDiscoveryHandler(v1service.NewDiscoveryService(repository.NewDiscoveryRepository(a.pool)))})
	if err != nil {
		t.Fatal(err)
	}
	call := func(method, path string, want int) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(method, "/api/v1/public"+path, nil))
		if w.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		return w
	}
	id := func(query string, args ...any) int64 {
		t.Helper()
		var value int64
		if err := a.pool.QueryRow(ctx, query, args...).Scan(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	room := id(`INSERT INTO rooms(name,room_type,default_capacity,description,calendar_required,calendar_job_error) VALUES('Public strength','STRENGTH',2,'Public description',false,'PRIVATE_JOB_ERROR') RETURNING id`)
	maintenance := id(`INSERT INTO rooms(name,room_type,default_capacity,status) VALUES('Public maintenance','GYM',2,'MAINTENANCE') RETURNING id`)
	hidden := id(`INSERT INTO rooms(name,room_type,default_capacity,status) VALUES('Hidden room','GYM',2,'INACTIVE') RETURNING id`)
	manualRoom := id(`INSERT INTO rooms(name,room_type,default_capacity,calendar_required) VALUES('Unpublished manual room','GYM',2,true) RETURNING id`)
	var cover bytes.Buffer
	if err := jpeg.Encode(&cover, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil); err != nil {
		t.Fatal(err)
	}
	a.sql(`UPDATE rooms SET cover_image=$2,cover_thumbnail=$2,cover_alt_text='Cover',cover_version=1 WHERE id IN ($1,$3)`, room, cover.Bytes(), hidden)
	a.sql(`INSERT INTO personal_trainers(user_id,specialty,experience_text,approval_status,certificate_path,decision_reason,decided_at,decided_by_user_id) VALUES($1,'Strength','Public experience','APPROVED','private/certificate.pdf','PRIVATE_DECISION',now(),$2)`, a.users["PT"], a.users["ADMIN"])
	a.sql(`INSERT INTO personal_trainers(user_id,approval_status) VALUES($1,'DRAFT')`, a.users["MEMBER"])
	slot := id(`INSERT INTO time_slots(name,start_time,end_time) VALUES('Public slot','08:00','09:00') RETURNING id`)
	session := id(`INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,change_reason,created_by_user_id) VALUES($1,$2,current_date+1,((current_date+1)+time '08:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',((current_date+1)+time '09:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',2,'PRIVATE_CHANGE',$3) RETURNING id`, room, slot, a.users["ADMIN"])
	a.sql(`INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) SELECT $1,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id FROM room_sessions WHERE id=$2`, manualRoom, session)
	member := id(`INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'PUBLIC-M',current_date,current_date+365) RETURNING id`, a.users["MEMBER"])
	a.sql(`INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,status,created_by_user_id) SELECT $1,id,session_date,time_slot_id,'CONFIRMED',$3 FROM room_sessions WHERE id=$2`, member, session, a.users["MEMBER"])
	var date, tooLate string
	if err := a.pool.QueryRow(ctx, `SELECT (current_date+1)::text,(current_date+8)::text`).Scan(&date, &tooLate); err != nil {
		t.Fatal(err)
	}
	response := call("GET", "/rooms?page_size=1", 200)
	if !strings.Contains(response.Body.String(), `"total":3`) {
		t.Fatal("incorrect public room count", response.Body.String())
	}
	var listed []discovery.Room
	decodeData(t, response, &listed)
	if len(listed) != 1 {
		t.Fatal("pagination", len(listed))
	}
	for _, field := range []string{"calendar_job_error", "PRIVATE_JOB_ERROR", "calendar_revision", "created_by_user_id"} {
		if strings.Contains(response.Body.String(), field) {
			t.Fatal("management field leaked", field)
		}
	}
	call("GET", fmt.Sprintf("/rooms/%d", hidden), 404)
	call("GET", fmt.Sprintf("/rooms/%d/cover-image", hidden), 404)
	call("GET", fmt.Sprintf("/rooms/%d", maintenance), 200)
	trainers := call("GET", "/trainers", 200)
	var pts []discovery.Trainer
	decodeData(t, trainers, &pts)
	if len(pts) != 1 || pts[0].FullName != "PT Test" {
		t.Fatal("trainer visibility", trainers.Body.String())
	}
	for _, field := range []string{"certificate", "user_id", "username", "email", "phone", "decision", "PRIVATE"} {
		if strings.Contains(trainers.Body.String(), field) {
			t.Fatal("private trainer field leaked", field)
		}
	}
	path := fmt.Sprintf("/rooms/%d/sessions?from=%s&to=%s", room, date, date)
	calendar := call("GET", path, 200)
	var sessions []discovery.Session
	decodeData(t, calendar, &sessions)
	if len(sessions) != 1 || sessions[0].Remaining != 1 {
		t.Fatal("public availability", calendar.Body.String())
	}
	for _, field := range []string{"member", "pt_id", "change_reason", "created_by", "PRIVATE"} {
		if strings.Contains(calendar.Body.String(), field) {
			t.Fatal("private session field leaked", field)
		}
	}
	unpublished := call("GET", fmt.Sprintf("/rooms/%d/sessions?from=%s&to=%s", manualRoom, date, date), 200)
	var absent []discovery.Session
	decodeData(t, unpublished, &absent)
	if len(absent) != 0 {
		t.Fatal("unpublished manual session leaked")
	}
	call("GET", fmt.Sprintf("/rooms/%d/sessions?from=%s&to=%s", room, date, tooLate), 422)
	call("GET", "/rooms?page_size=101", 422)
	call("GET", "/rooms/nope", 422)
	call("POST", "/rooms", 404)
	imageResponse := call("GET", fmt.Sprintf("/rooms/%d/cover-image?size=card", room), 200)
	if imageResponse.Header().Get("Content-Type") != "image/jpeg" {
		t.Fatal("wrong image content type")
	}
	conditionalImage := func(want int) {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/public/rooms/%d/cover-image?size=card", room), nil)
		req.Header.Set("If-None-Match", imageResponse.Header().Get("ETag"))
		router.ServeHTTP(w, req)
		if w.Code != want {
			t.Fatalf("conditional public image: %d %s", w.Code, w.Body.String())
		}
	}
	conditionalImage(304)
	a.sql(`UPDATE rooms SET status='MAINTENANCE' WHERE id=$1`, room)
	maintenanceCalendar := call("GET", path, 200)
	decodeData(t, maintenanceCalendar, &sessions)
	if len(sessions) != 1 || sessions[0].Status != "MAINTENANCE" || sessions[0].Remaining != 0 {
		t.Fatal("maintenance availability", maintenanceCalendar.Body.String())
	}
	a.sql(`UPDATE users SET status='LOCKED' WHERE id=$1`, a.users["PT"])
	locked := call("GET", "/trainers", 200)
	decodeData(t, locked, &pts)
	if len(pts) != 0 {
		t.Fatal("locked PT exposed", locked.Body.String())
	}
	a.sql(`UPDATE rooms SET status='INACTIVE' WHERE id=$1`, room)
	conditionalImage(404)
	call("GET", fmt.Sprintf("/rooms/%d/cover-image?size=card", room), 404)
	call("GET", path, 404)
	for _, privatePath := range []string{"/rooms", "/members", "/bookings", "/reports/operations"} {
		a.call("GET", privatePath, "", nil, 401)
	}
	a.call("POST", "/bookings", "", map[string]any{"room_session_id": session}, 401)
}
