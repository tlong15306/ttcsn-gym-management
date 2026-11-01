package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/attendance"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
)

type AttendanceRepository interface {
	Get(context.Context, identity.User, int64) (attendance.Summary, error)
	Record(context.Context, identity.User, int64, attendance.Input, string) (attendance.Result, error)
}
type PostgresAttendanceRepository struct{ pool *pgxpool.Pool }

func NewAttendanceRepository(pool *pgxpool.Pool) *PostgresAttendanceRepository {
	return &PostgresAttendanceRepository{pool}
}

const attendanceSelect = `SELECT a.id,a.booking_id,a.participant_type,a.participant_user_id,a.checked_in_at,a.recorded_by_user_id,u.full_name recorder_name,COALESCE(a.override_reason,'') override_reason FROM booking_attendance a JOIN users u ON u.id=a.recorded_by_user_id`

func (r *PostgresAttendanceRepository) Get(ctx context.Context, a identity.User, id int64) (attendance.Summary, error) {
	// One SQL statement gives a consistent booking/attendance snapshot.
	return jsonOne[attendance.Summary](ctx, r.pool, `SELECT clock_timestamp() server_time,
 jsonb_build_array(jsonb_build_object('participant_type','MEMBER','name',mu.full_name,'attendance',(SELECT to_jsonb(x) FROM (`+attendanceSelect+` WHERE a.booking_id=b.id AND a.participant_type='MEMBER') x))) ||
 CASE WHEN b.pt_id IS NULL THEN '[]'::jsonb ELSE jsonb_build_array(jsonb_build_object('participant_type','PT','name',pu.full_name,'attendance',(SELECT to_jsonb(x) FROM (`+attendanceSelect+` WHERE a.booking_id=b.id AND a.participant_type='PT') x))) END participants
 FROM bookings b JOIN members m ON m.id=b.member_id JOIN users mu ON mu.id=m.user_id LEFT JOIN personal_trainers p ON p.id=b.pt_id LEFT JOIN users pu ON pu.id=p.user_id WHERE b.id=$1 AND `+bookingScope(a), id)
}
func (r *PostgresAttendanceRepository) Record(ctx context.Context, a identity.User, id int64, in attendance.Input, rid string) (attendance.Result, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (attendance.Result, error) {
		var result attendance.Result
		b, err := jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1 AND `+bookingScope(a), id)
		if err != nil {
			return result, err
		}
		if err = bookingLocks(ctx, tx, b.RoomID, b.MemberID, b.PTID, b.RoomSessionID); err != nil {
			return result, err
		}
		var locked int64
		if err = tx.QueryRow(ctx, `SELECT id FROM bookings WHERE id=$1 FOR UPDATE`, id).Scan(&locked); err != nil {
			return result, err
		}
		b, err = jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1`, id)
		if err != nil {
			return result, err
		}
		if in.ParticipantType == "PT" && b.PTID == 0 {
			return result, &attendance.Error{Code: "ATTENDANCE_PARTICIPANT_INVALID", Message: "Buổi tự tập không có PT để điểm danh."}
		}
		existing, err := jsonOne[attendance.Record](ctx, tx, attendanceSelect+` WHERE a.booking_id=$1 AND a.participant_type=$2`, id, in.ParticipantType)
		if err == nil {
			result.Record = existing
			return result, nil
		}
		if !errors.Is(err, identity.ErrNotFound) {
			return result, err
		}
		if b.Status != "CONFIRMED" || !b.ServerTime.Before(b.EndsAt) {
			return result, booking.Fail("ATTENDANCE_INVALID_STATE", "Chỉ điểm danh buổi đã xác nhận và chưa kết thúc.")
		}
		if !attendance.InWindow(b.ServerTime, b.StartsAt, b.EndsAt) && in.OverrideReason == "" {
			return result, &attendance.Error{Code: "ATTENDANCE_WINDOW", Message: "Điểm danh mở trước giờ tập 15 phút. Ghi sớm cần lý do ngoại lệ."}
		}
		var userID int64
		if in.ParticipantType == "MEMBER" {
			err = tx.QueryRow(ctx, `SELECT user_id FROM members WHERE id=$1`, b.MemberID).Scan(&userID)
		} else {
			err = tx.QueryRow(ctx, `SELECT user_id FROM personal_trainers WHERE id=$1`, b.PTID).Scan(&userID)
		}
		if err != nil {
			return result, err
		}
		var attendanceID int64
		err = tx.QueryRow(ctx, `INSERT INTO booking_attendance(booking_id,participant_type,participant_user_id,recorded_by_user_id,override_reason,checked_in_at) VALUES($1,$2,$3,$4,NULLIF($5,''),$6) RETURNING id`, id, in.ParticipantType, userID, a.ID, in.OverrideReason, b.ServerTime).Scan(&attendanceID)
		if err != nil {
			return result, err
		}
		result.Record, err = jsonOne[attendance.Record](ctx, tx, attendanceSelect+` WHERE a.id=$1`, attendanceID)
		if err != nil {
			return result, err
		}
		result.Created = true
		return result, writeAudit(ctx, tx, a.ID, "attendance.RECORDED", "booking", id, in.OverrideReason, nil, result.Record, rid)
	})
}

// Completion uses the same locks as cancellation/check-in. Only MEMBER presence
// determines the booking result; PT attendance remains independently visible.
func (r *PostgresBookingRepository) Complete(ctx context.Context) error {
	rows, err := r.pool.Query(ctx, `SELECT b.id FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id WHERE b.status='CONFIRMED' AND s.ends_at<=clock_timestamp() ORDER BY b.id LIMIT 1000`)
	if err != nil {
		return err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	var failures []error
	for _, id := range ids {
		_, err = inTransaction(ctx, r.pool, func(tx pgx.Tx) (bool, error) {
			b, e := jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1`, id)
			if e != nil {
				return false, e
			}
			if e = bookingLocks(ctx, tx, b.RoomID, b.MemberID, b.PTID, b.RoomSessionID); e != nil {
				return false, e
			}
			var changed int64
			e = tx.QueryRow(ctx, `UPDATE bookings b SET status=CASE WHEN EXISTS(SELECT 1 FROM booking_attendance a WHERE a.booking_id=b.id AND a.participant_type='MEMBER') THEN 'COMPLETED' ELSE 'NO_SHOW' END,updated_at=clock_timestamp() FROM room_sessions s WHERE b.id=$1 AND s.id=b.room_session_id AND b.status='CONFIRMED' AND s.ends_at<=clock_timestamp() RETURNING b.id`, id).Scan(&changed)
			if e == pgx.ErrNoRows {
				return false, nil
			}
			if e != nil {
				return false, e
			}
			after, e := jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1`, id)
			if e != nil {
				return false, e
			}
			return true, writeAudit(ctx, tx, 0, "booking."+after.Status, "booking", id, "Kết thúc buổi theo điểm danh Hội viên", b, after, "booking-completion")
		})
		if err != nil {
			failures = append(failures, err)
		}
		if ctx.Err() != nil {
			break
		}
	}
	return errors.Join(failures...)
}
