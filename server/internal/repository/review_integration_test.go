package repository_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/review"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/trainer"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	"github.com/tlong15306/ttcsn-gym-management/server/pkg/auth"
)

type reviewFixture struct {
	bookingID int64
	ptID      int64
}

func seedReviewFixture(t *testing.T, a *testAPI) (int64, func(int, string, bool, bool, bool) reviewFixture) {
	t.Helper()
	ctx := context.Background()
	var memberID, ptID, roomID, slotID, otherMemberID, otherPTID int64
	if err := a.pool.QueryRow(ctx, `INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'REVIEW-M',current_date,current_date+365) RETURNING id`, a.users["MEMBER"]).Scan(&memberID); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO personal_trainers(user_id,approval_status,decided_at,decided_by_user_id) VALUES($1,'APPROVED',clock_timestamp(),$2) RETURNING id`, a.users["PT"], a.users["ADMIN"]).Scan(&ptID); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO rooms(name,room_type,default_capacity) VALUES('Review room','GYM',10) RETURNING id`).Scan(&roomID); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO time_slots(name,start_time,end_time) VALUES('Review slot','10:00','11:00') RETURNING id`).Scan(&slotID); err != nil {
		t.Fatal(err)
	}
	makeUser := func(username, role string) int64 {
		var id int64
		if err := a.pool.QueryRow(ctx, `INSERT INTO users(username,password_hash,email,full_name,role_id) SELECT $1,'unused',$2,$3,id FROM roles WHERE code=$4 RETURNING users.id`, username, username+"@test.local", username, role).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	otherMemberUser := makeUser("review.outsider", "MEMBER")
	if err := a.pool.QueryRow(ctx, `INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'REVIEW-OUT',current_date,current_date+365) RETURNING id`, otherMemberUser).Scan(&otherMemberID); err != nil {
		t.Fatal(err)
	}
	otherPTUser := makeUser("review.otherpt", "PT")
	if err := a.pool.QueryRow(ctx, `INSERT INTO personal_trainers(user_id,approval_status,decided_at,decided_by_user_id) VALUES($1,'APPROVED',clock_timestamp(),$2) RETURNING id`, otherPTUser, a.users["ADMIN"]).Scan(&otherPTID); err != nil {
		t.Fatal(err)
	}
	tokens := auth.NewTokenService("01234567890123456789012345678901", 15*time.Minute)
	outsiderMemberToken, err := tokens.Issue(identity.User{ID: otherMemberUser, Role: identity.RoleMember, AuthVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	otherPTToken, err := tokens.Issue(identity.User{ID: otherPTUser, Role: identity.RolePT, AuthVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	a.tokens["OUTSIDER_MEMBER"] = outsiderMemberToken
	a.tokens["OTHER_PT"] = otherPTToken

	makeBooking := func(days int, status string, withPT, memberAttended, ptAttended bool) reviewFixture {
		var sessionID int64
		err := a.pool.QueryRow(ctx, `INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id)
			VALUES($1,$2,current_date+$3::int,((current_date+$3::int)+time '10:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',((current_date+$3::int)+time '11:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',10,$4) RETURNING id`, roomID, slotID, days, a.users["ADMIN"]).Scan(&sessionID)
		if err != nil {
			t.Fatal(err)
		}
		var bookingID int64
		var assignedPT any
		var selectedPT int64
		if withPT {
			assignedPT, selectedPT = ptID, ptID
		}
		err = a.pool.QueryRow(ctx, `INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,pt_id,status,created_by_user_id,confirmed_at)
			SELECT $1,$2,s.session_date,s.time_slot_id,$3,$4,$5,clock_timestamp() FROM room_sessions s WHERE s.id=$2 RETURNING id`, memberID, sessionID, assignedPT, status, a.users["MEMBER"]).Scan(&bookingID)
		if err != nil {
			t.Fatal(err)
		}
		if memberAttended {
			a.sql(`INSERT INTO booking_attendance(booking_id,participant_type,participant_user_id,recorded_by_user_id) VALUES($1,'MEMBER',$2,$3)`, bookingID, a.users["MEMBER"], a.users["ADMIN"])
		}
		if ptAttended && withPT {
			a.sql(`INSERT INTO booking_attendance(booking_id,participant_type,participant_user_id,recorded_by_user_id) VALUES($1,'PT',$2,$3)`, bookingID, a.users["PT"], a.users["ADMIN"])
		}
		return reviewFixture{bookingID: bookingID, ptID: selectedPT}
	}
	return otherPTID, makeBooking
}

func postReviewRaw(a *testAPI, bookingID int64, rating int) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]any{"rating": rating, "comment": "race review"})
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/bookings/%d/review", bookingID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+a.tokens["MEMBER"])
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "review-race")
	w := httptest.NewRecorder()
	a.router.ServeHTTP(w, req)
	return w
}

func requireReviewError(t *testing.T, w *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("status = %d, want %d: %s", w.Code, status, w.Body.String())
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Error.Code != code {
		t.Fatalf("error code = %q, want %q: %s", payload.Error.Code, code, w.Body.String())
	}
}

func TestReviewAPIEligibilityPrivacyAndRatingIntegration(t *testing.T) {
	a := newManagementAPI(t)
	otherPTID, makeBooking := seedReviewFixture(t, a)
	eligible := makeBooking(-1, "COMPLETED", true, true, true)
	missingMember := makeBooking(-2, "COMPLETED", true, false, true)
	missingPT := makeBooking(-3, "COMPLETED", true, true, false)
	selfTraining := makeBooking(-4, "COMPLETED", false, true, false)
	noShow := makeBooking(-5, "NO_SHOW", true, true, true)
	futureEnd := makeBooking(1, "COMPLETED", true, true, true)
	a.sql(`INSERT INTO pt_working_hours(pt_id,day_of_week,start_time,end_time)
		SELECT $1,EXTRACT(ISODOW FROM s.session_date)::smallint,'09:00','12:00' FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id WHERE b.id=$2`, eligible.ptID, futureEnd.bookingID)

	for _, fixture := range []reviewFixture{missingMember, missingPT, selfTraining, noShow, futureEnd} {
		requireReviewError(t, a.call("POST", fmt.Sprintf("/bookings/%d/review", fixture.bookingID), "MEMBER", map[string]any{"rating": 4}, 422), 422, "REVIEW_NOT_ELIGIBLE")
	}
	requireReviewError(t, a.call("POST", fmt.Sprintf("/bookings/%d/review", eligible.bookingID), "PT", map[string]any{"rating": 4}, 403), 403, "FORBIDDEN")
	requireReviewError(t, a.call("POST", fmt.Sprintf("/bookings/%d/review", eligible.bookingID), "MEMBER", map[string]any{"rating": 6}, 422), 422, "VALIDATION_FAILED")
	requireReviewError(t, a.call("POST", fmt.Sprintf("/bookings/%d/review", eligible.bookingID), "MEMBER", map[string]any{"rating": 5, "member_id": 1}, 422), 422, "VALIDATION_FAILED")
	requireReviewError(t, a.call("GET", fmt.Sprintf("/bookings/%d/review", eligible.bookingID), "OUTSIDER_MEMBER", nil, 404), 404, "NOT_FOUND")
	requireReviewError(t, a.call("GET", fmt.Sprintf("/bookings/%d/review", eligible.bookingID), "OTHER_PT", nil, 404), 404, "NOT_FOUND")
	noReview := a.call("GET", fmt.Sprintf("/bookings/%d/review", eligible.bookingID), "MEMBER", nil, 200)
	if !strings.Contains(noReview.Body.String(), `"data":null`) {
		t.Fatalf("missing review should return null data: %s", noReview.Body.String())
	}

	// An audit storage error must roll back the review insert.
	rollback := repository.NewReviewRepository(a.pool)
	_, err := rollback.Create(context.Background(), identity.User{ID: a.users["MEMBER"], Role: identity.RoleMember}, eligible.bookingID, review.Input{Rating: 4}, strings.Repeat("r", 101))
	if err == nil {
		t.Fatal("expected overlong audit request id to fail")
	}
	var count int
	if err = a.pool.QueryRow(context.Background(), `SELECT count(*) FROM pt_reviews WHERE booking_id=$1`, eligible.bookingID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("review survived audit rollback: count=%d err=%v", count, err)
	}

	created := a.call("POST", fmt.Sprintf("/bookings/%d/review", eligible.bookingID), "MEMBER", map[string]any{"rating": 4, "comment": "  Rất tốt  "}, 201)
	var first review.Review
	decodeData(t, created, &first)
	if first.Rating != 4 || first.Comment != "Rất tốt" || first.ReviewerLabel != "Hội viên" {
		t.Fatalf("unexpected review: %+v", first)
	}
	requireReviewError(t, a.call("POST", fmt.Sprintf("/bookings/%d/review", eligible.bookingID), "MEMBER", map[string]any{"rating": 2}, 409), 409, "REVIEW_EXISTS")
	for _, role := range []string{"MEMBER", "PT", "RECEPTIONIST", "MANAGER", "ADMIN"} {
		a.call("GET", fmt.Sprintf("/bookings/%d/review", eligible.bookingID), role, nil, 200)
	}
	public := a.call("GET", fmt.Sprintf("/pts/%d/reviews", eligible.ptID), "MEMBER", nil, 200)
	for _, private := range []string{"member_id", "user_id", "booking_id", "username", "email", "phone", "REVIEW-M", "MEMBER Test"} {
		if strings.Contains(public.Body.String(), private) {
			t.Fatalf("private review data %q leaked: %s", private, public.Body.String())
		}
	}
	for _, role := range []string{"PT", "RECEPTIONIST", "MANAGER", "ADMIN"} {
		a.call("GET", fmt.Sprintf("/pts/%d/reviews", eligible.ptID), role, nil, 200)
	}
	requireReviewError(t, a.call("GET", fmt.Sprintf("/pts/%d/reviews", eligible.ptID), "OTHER_PT", nil, 404), 404, "NOT_FOUND")
	otherPTReviews := a.call("GET", fmt.Sprintf("/pts/%d/reviews", otherPTID), "MEMBER", nil, 200)
	if !strings.Contains(otherPTReviews.Body.String(), `"total":0`) {
		t.Fatalf("new public PT should have no reviews: %s", otherPTReviews.Body.String())
	}

	// Concurrent duplicate submissions have exactly one winner and one audit.
	raceBooking := makeBooking(-6, "COMPLETED", true, true, true)
	var wg sync.WaitGroup
	results := make(chan *httptest.ResponseRecorder, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- postReviewRaw(a, raceBooking.bookingID, 2) }()
	}
	wg.Wait()
	close(results)
	createdCount, existsCount := 0, 0
	for w := range results {
		switch w.Code {
		case 201:
			createdCount++
		case 409:
			requireReviewError(t, w, 409, "REVIEW_EXISTS")
			existsCount++
		default:
			t.Fatalf("unexpected concurrent response: %d %s", w.Code, w.Body.String())
		}
	}
	if createdCount != 1 || existsCount != 11 {
		t.Fatalf("concurrent outcomes created=%d duplicate=%d", createdCount, existsCount)
	}
	if err = a.pool.QueryRow(context.Background(), `SELECT count(*) FROM audit_events WHERE action='review.created'`).Scan(&count); err != nil || count != 2 {
		t.Fatalf("review audit count=%d err=%v", count, err)
	}
	paged := a.call("GET", fmt.Sprintf("/pts/%d/reviews?page=1&page_size=1", eligible.ptID), "MEMBER", nil, 200)
	var page struct {
		Meta struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
			Total    int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(paged.Body.Bytes(), &page); err != nil || page.Meta.Page != 1 || page.Meta.PageSize != 1 || page.Meta.Total != 2 {
		t.Fatalf("review pagination mismatch: %+v err=%v body=%s", page.Meta, err, paged.Body.String())
	}

	profileResponse := a.call("GET", fmt.Sprintf("/pts/%d", eligible.ptID), "ADMIN", nil, 200)
	var profile trainer.Profile
	decodeData(t, profileResponse, &profile)
	if profile.RatingAverage == nil || *profile.RatingAverage != 3 || profile.ReviewCount != 2 {
		t.Fatalf("PT profile rating summary mismatch: avg=%v count=%d", profile.RatingAverage, profile.ReviewCount)
	}
	var optionSessionID int64
	if err = a.pool.QueryRow(context.Background(), `SELECT room_session_id FROM bookings WHERE id=$1`, futureEnd.bookingID).Scan(&optionSessionID); err != nil {
		t.Fatal(err)
	}
	options := a.call("GET", fmt.Sprintf("/booking-options/pts?room_session_id=%d", optionSessionID), "MEMBER", nil, 200)
	var ptOptions []booking.PTOption
	decodeData(t, options, &ptOptions)
	if len(ptOptions) != 1 || ptOptions[0].RatingAverage == nil || *ptOptions[0].RatingAverage != 3 || ptOptions[0].ReviewCount != 2 {
		t.Fatalf("booking PT options rating mismatch: %+v", ptOptions)
	}
}
