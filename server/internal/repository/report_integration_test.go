package repository_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/report"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
)

func TestReportAPIUsesVerifiedPeriodMetricsAndScopes(t *testing.T) {
	a := newManagementAPI(t)
	ctx := context.Background()
	var memberID, ptID, roomID, slotID int64
	if err := a.pool.QueryRow(ctx, `INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'REPORT-M',current_date,current_date+365) RETURNING id`, a.users["MEMBER"]).Scan(&memberID); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO personal_trainers(user_id,approval_status,decided_at,decided_by_user_id) VALUES($1,'APPROVED',clock_timestamp(),$2) RETURNING id`, a.users["PT"], a.users["ADMIN"]).Scan(&ptID); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO rooms(name,room_type,default_capacity) VALUES('Report fixture room','GYM',10) RETURNING id`).Scan(&roomID); err != nil {
		t.Fatal(err)
	}
	if err := a.pool.QueryRow(ctx, `INSERT INTO time_slots(name,start_time,end_time) VALUES('Report fixture slot','08:00','09:00') RETURNING id`).Scan(&slotID); err != nil {
		t.Fatal(err)
	}
	var localDate string
	if err := a.pool.QueryRow(ctx, `SELECT (clock_timestamp() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date::text`).Scan(&localDate); err != nil {
		t.Fatal(err)
	}
	today, err := time.Parse("2006-01-02", localDate)
	if err != nil {
		t.Fatal(err)
	}
	from := today.AddDate(0, 0, -30).Format("2006-01-02")
	to := today.Format("2006-01-02")
	type fixture struct {
		dayOffset     int
		status        string
		withPT        bool
		confirmed     bool
		source        string
		memberPresent bool
		ptPresent     bool
	}
	fixtures := []fixture{
		{1, "COMPLETED", true, true, "", true, true}, // eligible review and both attendance types
		{2, "NO_SHOW", true, true, "", false, true},  // PT attended, member did not
		{3, "NO_SHOW", true, true, "", false, false},
		{4, "CANCELLED", true, true, "MEMBER", false, true}, // BR-22 includes member cancel
		{5, "CANCELLED", true, true, "PT", false, false},
		{6, "CANCELLED", true, false, "FACILITY", false, false}, // known facility cancellation stays excluded without confirmation evidence
		{7, "CANCELLED", true, true, "UNKNOWN", false, false},
		{8, "COMPLETED", true, false, "", false, false}, // legacy confirmation missing
		{9, "REJECTED", true, false, "", false, false},
		{10, "EXPIRED", true, false, "", false, false},
		{11, "CONFIRMED", true, true, "", false, false}, // completion job pending
		{12, "COMPLETED", false, true, "", true, false}, // self practice
		{13, "CONFIRMED", true, true, "", false, false}, // future session
		{14, "PENDING_PT", true, false, "", false, false},
	}
	var reviewedBooking int64
	for i, f := range fixtures {
		day := today.AddDate(0, 0, -f.dayOffset)
		if f.dayOffset == 13 { // future fixture
			day = today.AddDate(0, 0, 1)
		}
		var sessionID int64
		if err = a.pool.QueryRow(ctx, `INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id)
			VALUES($1,$2,$3::date,($3::date+time '08:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',($3::date+time '09:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',10,$4) RETURNING id`, roomID, slotID, day.Format("2006-01-02"), a.users["ADMIN"]).Scan(&sessionID); err != nil {
			t.Fatal(err)
		}
		var selectedPT any
		if f.withPT {
			selectedPT = ptID
		}
		var confirmedAt any
		if f.confirmed {
			confirmedAt = time.Now().Add(-24 * time.Hour)
		}
		var respondBy any
		if f.status == "PENDING_PT" {
			respondBy = time.Now().Add(24 * time.Hour)
		}
		var cancelledBy, cancelledAt, reason, source, actorRole any
		if f.status == "CANCELLED" {
			role := map[string]string{"MEMBER": "MEMBER", "PT": "PT", "FACILITY": "MANAGER", "UNKNOWN": ""}[f.source]
			if role != "" {
				cancelledBy = a.users[role]
				actorRole = role
			} else {
				cancelledBy = a.users["MEMBER"]
			}
			cancelledAt = time.Now().Add(-24 * time.Hour)
			reason, source = "synthetic fixture", f.source
		}
		var bookingID int64
		if err = a.pool.QueryRow(ctx, `INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,pt_id,status,respond_by,created_by_user_id,confirmed_at,cancelled_by_user_id,cancellation_reason,cancelled_at,cancellation_source,cancelled_actor_role)
			SELECT $1,$2,session_date,time_slot_id,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12 FROM room_sessions WHERE id=$2 RETURNING bookings.id`, memberID, sessionID, selectedPT, f.status, respondBy, a.users["MEMBER"], confirmedAt, cancelledBy, reason, cancelledAt, source, actorRole).Scan(&bookingID); err != nil {
			t.Fatalf("fixture %d booking: %v", i+1, err)
		}
		if f.memberPresent {
			if _, err = a.pool.Exec(ctx, `INSERT INTO booking_attendance(booking_id,participant_type,participant_user_id,recorded_by_user_id) VALUES($1,'MEMBER',$2,$3)`, bookingID, a.users["MEMBER"], a.users["MANAGER"]); err != nil {
				t.Fatal(err)
			}
		}
		if f.ptPresent {
			if _, err = a.pool.Exec(ctx, `INSERT INTO booking_attendance(booking_id,participant_type,participant_user_id,recorded_by_user_id) VALUES($1,'PT',$2,$3)`, bookingID, a.users["PT"], a.users["MANAGER"]); err != nil {
				t.Fatal(err)
			}
		}
		if i == 0 {
			reviewedBooking = bookingID
		}
	}
	if _, err = a.pool.Exec(ctx, `INSERT INTO pt_reviews(booking_id,rating,comment) VALUES($1,5,'verified report fixture')`, reviewedBooking); err != nil {
		t.Fatal(err)
	}
	var otherRoomID, otherSessionID int64
	if err = a.pool.QueryRow(ctx, `INSERT INTO rooms(name,room_type,default_capacity) VALUES('Other report room','GYM',2) RETURNING id`).Scan(&otherRoomID); err != nil {
		t.Fatal(err)
	}
	otherDay := today.AddDate(0, 0, -1).Format("2006-01-02")
	if err = a.pool.QueryRow(ctx, `INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id)
		VALUES($1,$2,$3::date,($3::date+time '08:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',($3::date+time '09:00') AT TIME ZONE 'Asia/Ho_Chi_Minh',2,$4) RETURNING id`, otherRoomID, slotID, otherDay, a.users["ADMIN"]).Scan(&otherSessionID); err != nil {
		t.Fatal(err)
	}
	if _, err = a.pool.Exec(ctx, `INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,status,created_by_user_id)
		SELECT $1,$2,session_date,time_slot_id,'REJECTED',$3 FROM room_sessions WHERE id=$2`, memberID, otherSessionID, a.users["MEMBER"]); err != nil {
		t.Fatal(err)
	}
	directOps, err := repository.NewReportRepository(a.pool).Operations(ctx, report.Range{From: today.AddDate(0, 0, -30), To: today}, roomID)
	if err != nil {
		t.Fatalf("report repository query: %v", err)
	}
	if directOps.TotalBookings != 13 {
		t.Fatalf("report repository returned an unexpected total: %+v", directOps)
	}

	ops := a.call("GET", fmt.Sprintf("/reports/operations?from=%s&to=%s&room_id=%d", url.QueryEscape(from), url.QueryEscape(to), roomID), "MANAGER", nil, 200)
	var operations report.Operations
	decodeData(t, ops, &operations)
	if operations.TotalBookings != 13 || operations.StatusCounts["NO_SHOW"] != 2 || operations.MemberAttendanceCount != 2 || operations.PTAttendanceCount != 3 {
		t.Fatalf("operations counts mismatch: %+v", operations)
	}
	if operations.NoShow.Numerator != 2 || operations.NoShow.Denominator != 5 || operations.NoShow.Percent == nil || *operations.NoShow.Percent != 40 {
		t.Fatalf("no-show formula mismatch: %+v", operations.NoShow)
	}
	if operations.PendingCompletionCount != 1 || operations.UnclassifiedBookingCount != 2 {
		t.Fatalf("pending/legacy classification mismatch: %+v", operations)
	}
	if operations.Utilization.Booked != 360 || operations.Utilization.Capacity != 7800 || operations.Utilization.Percent == nil || *operations.Utilization.Percent != 4.62 {
		t.Fatalf("reservation utilization formula mismatch: %+v", operations.Utilization)
	}
	if operations.CancellationSourceCounts["UNKNOWN"] != 1 || operations.CancellationSourceCounts["FACILITY"] != 1 {
		t.Fatalf("cancellation sources missing: %+v", operations.CancellationSourceCounts)
	}
	allRoomsResponse := a.call("GET", fmt.Sprintf("/reports/operations?from=%s&to=%s", from, to), "ADMIN", nil, 200)
	var allRooms report.Operations
	decodeData(t, allRoomsResponse, &allRooms)
	if allRooms.TotalBookings != 14 || allRooms.Utilization.Capacity != 7920 || allRooms.Utilization.Booked != 360 {
		t.Fatalf("room filter or facility total mismatch: %+v", allRooms)
	}

	ptResponse := a.call("GET", fmt.Sprintf("/reports/pts/%d?from=%s&to=%s", ptID, from, to), "PT", nil, 200)
	var trainer report.Trainer
	decodeData(t, ptResponse, &trainer)
	if trainer.Attendance.Numerator != 3 || trainer.Attendance.Denominator != 5 || trainer.Attendance.Percent == nil || *trainer.Attendance.Percent != 60 {
		t.Fatalf("PT attendance metric mismatch: %+v", trainer.Attendance)
	}
	if trainer.ReviewCount != 1 || trainer.RatingAverage == nil || *trainer.RatingAverage != 5 || trainer.ExcludedCount != 4 || trainer.UnclassifiedCount != 2 || trainer.PendingCompletionCount != 1 {
		t.Fatalf("PT rating/classification mismatch: %+v", trainer)
	}
	future := today.AddDate(0, 0, 1).Format("2006-01-02")
	emptyOpsResponse := a.call("GET", fmt.Sprintf("/reports/operations?from=%s&to=%s&room_id=%d", future, future, roomID), "MANAGER", nil, 200)
	var emptyOps report.Operations
	decodeData(t, emptyOpsResponse, &emptyOps)
	if emptyOps.TotalBookings != 1 || emptyOps.NoShow.Denominator != 0 || emptyOps.NoShow.Percent != nil || emptyOps.Utilization.Capacity != 0 || emptyOps.Utilization.Percent != nil {
		t.Fatalf("zero-denominator metrics should be N/A: %+v", emptyOps)
	}
	emptyPTResponse := a.call("GET", fmt.Sprintf("/reports/pts/%d?from=%s&to=%s", ptID, future, future), "PT", nil, 200)
	var emptyPT report.Trainer
	decodeData(t, emptyPTResponse, &emptyPT)
	if emptyPT.ReviewCount != 0 || emptyPT.RatingAverage != nil || emptyPT.Attendance.Denominator != 0 || emptyPT.Attendance.Percent != nil {
		t.Fatalf("empty PT report should expose N/A, not zero: %+v", emptyPT)
	}
	a.call("GET", fmt.Sprintf("/reports/operations?from=%s&to=%s&room_id=%d", from, to, roomID), "ADMIN", nil, 200)
	a.call("GET", fmt.Sprintf("/reports/pts/%d?from=%s&to=%s", ptID, from, to), "MANAGER", nil, 200)
	a.call("GET", fmt.Sprintf("/reports/operations?room_id=%d", roomID), "MEMBER", nil, 403)
	a.call("GET", fmt.Sprintf("/reports/operations?room_id=%d", roomID), "RECEPTIONIST", nil, 403)
	a.call("GET", fmt.Sprintf("/reports/pts/%d", ptID), "MEMBER", nil, 403)
	a.call("GET", "/reports/operations?from=2026-02-30", "ADMIN", nil, 422)
	a.call("GET", "/reports/operations?from=2026-10-15&to=2026-10-14", "ADMIN", nil, 422)
	a.call("GET", "/reports/operations?room_id=0", "ADMIN", nil, 422)
	a.call("GET", fmt.Sprintf("/reports/operations?room_id=%d", roomID+99999), "ADMIN", nil, 404)
	var otherUserID int64
	if err = a.pool.QueryRow(ctx, `INSERT INTO users(username,password_hash,email,full_name,role_id) SELECT 'report.other.pt','unused','report.other.pt@test.local','Other PT',id FROM roles WHERE code='PT' RETURNING users.id`).Scan(&otherUserID); err != nil {
		t.Fatal(err)
	}
	var otherPTID int64
	if err = a.pool.QueryRow(ctx, `INSERT INTO personal_trainers(user_id,approval_status) VALUES($1,'DRAFT') RETURNING id`, otherUserID).Scan(&otherPTID); err != nil {
		t.Fatal(err)
	}
	a.call("GET", fmt.Sprintf("/reports/pts/%d", otherPTID), "PT", nil, 404)
}
