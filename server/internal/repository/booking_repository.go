package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"sort"
	"time"
)

type BookingRepository interface {
	Create(context.Context, identity.User, int64, booking.CreateInput, string, string) (booking.Booking, error)
	CreateBatch(context.Context, identity.User, int64, booking.BatchCreateInput, string, string) ([]booking.Booking, error)
	List(context.Context, identity.User, booking.Filter) ([]booking.Booking, int, error)
	Get(context.Context, identity.User, int64) (booking.Booking, error)
	Act(context.Context, identity.User, int64, booking.ActionInput, bool, string) (booking.Booking, error)
	Sessions(context.Context, identity.User, catalog.Filter) ([]booking.SessionOption, int, error)
	PTs(context.Context, int64, catalog.Filter) ([]booking.PTOption, int, error)
	Expire(context.Context) error
	PurgeExpiredIdempotency(context.Context) error
}
type PostgresBookingRepository struct{ pool *pgxpool.Pool }

func NewBookingRepository(pool *pgxpool.Pool) *PostgresBookingRepository {
	return &PostgresBookingRepository{pool}
}

func deleteExpiredIdempotencyKey(ctx context.Context, tx pgx.Tx, actorID int64, scope, key string) error {
	_, err := tx.Exec(ctx, `DELETE FROM idempotency_keys WHERE actor_user_id=$1 AND request_scope=$2 AND idempotency_key=$3 AND expires_at<=clock_timestamp()`, actorID, scope, key)
	return err
}

// PurgeExpiredIdempotency removes records after their documented 24-hour
// replay window. It is safe to run repeatedly and covers booking and calendar
// keys because both scopes share the same table.
func (r *PostgresBookingRepository) PurgeExpiredIdempotency(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM idempotency_keys WHERE expires_at<=clock_timestamp()`)
	return err
}

const bookingSelect = `SELECT b.id,b.member_id,mu.full_name member_name,b.room_session_id,s.room_id,r.name room_name,COALESCE(b.pt_id,0) pt_id,COALESCE(pu.full_name,'') pt_name,b.status,s.starts_at,s.ends_at,b.confirmed_at,b.respond_by,b.created_at,COALESCE(b.cancellation_reason,'') cancellation_reason,b.cancellation_source,b.cancelled_actor_role,EXISTS(SELECT 1 FROM booking_attendance ba WHERE ba.booking_id=b.id) has_attendance,clock_timestamp() server_time FROM bookings b JOIN members m ON m.id=b.member_id JOIN users mu ON mu.id=m.user_id JOIN room_sessions s ON s.id=b.room_session_id JOIN rooms r ON r.id=s.room_id LEFT JOIN personal_trainers p ON p.id=b.pt_id LEFT JOIN users pu ON pu.id=p.user_id`

func bookingScope(a identity.User) string {
	switch a.Role {
	case identity.RoleMember:
		return "m.user_id=" + fmt.Sprint(a.ID)
	case identity.RolePT:
		return "p.user_id=" + fmt.Sprint(a.ID)
	default:
		return "TRUE"
	}
}
func (r *PostgresBookingRepository) Get(ctx context.Context, a identity.User, id int64) (booking.Booking, error) {
	return jsonOne[booking.Booking](ctx, r.pool, bookingSelect+` WHERE b.id=$1 AND `+bookingScope(a), id)
}
func (r *PostgresBookingRepository) List(ctx context.Context, a identity.User, f booking.Filter) ([]booking.Booking, int, error) {
	return jsonList[booking.Booking](ctx, r.pool, bookingSelect+` WHERE `+bookingScope(a)+` AND ($1='' OR b.status=$1) AND ($2='' OR b.session_date>=NULLIF($2,'')::date) AND ($3='' OR b.session_date<=NULLIF($3,'')::date) AND ($4::bigint=0 OR s.room_id=$4) AND ($5::bigint=0 OR b.pt_id=$5) AND ($6='' OR b.id::text=$6 OR m.member_code ILIKE '%'||$6||'%' OR mu.full_name ILIKE '%'||$6||'%' OR ($7::boolean AND mu.phone ILIKE '%'||$6||'%')) AND (NOT $8::boolean OR s.ends_at<=clock_timestamp() OR b.status IN ('CANCELLED','REJECTED','EXPIRED','COMPLETED','NO_SHOW'))`, []any{f.Status, f.From, f.To, f.RoomID, f.PTID, f.Query, a.Role != identity.RoleMember && a.Role != identity.RolePT, f.History}, f.Filter)
}

const ptAvailable = `p.approval_status='APPROVED' AND u.status='ACTIVE' AND pr.code='PT'
 AND EXISTS(SELECT 1 FROM pt_working_hours h WHERE h.pt_id=p.id AND h.is_active AND h.day_of_week=EXTRACT(ISODOW FROM s.session_date) AND h.start_time<=(s.starts_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::time AND ((s.ends_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date=s.session_date AND h.end_time>=(s.ends_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::time OR (s.ends_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date=s.session_date+1 AND (s.ends_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::time='00:00'::time AND h.end_time='24:00'::time))
 AND NOT EXISTS(SELECT 1 FROM pt_unavailability n WHERE n.pt_id=p.id AND tstzrange(n.starts_at,n.ends_at,'[)') && tstzrange(s.starts_at,s.ends_at,'[)'))
 AND NOT EXISTS(SELECT 1 FROM bookings x JOIN room_sessions xs ON xs.id=x.room_session_id WHERE x.pt_id=p.id AND x.status IN ('PENDING_PT','CONFIRMED') AND tstzrange(xs.starts_at,xs.ends_at,'[)') && tstzrange(s.starts_at,s.ends_at,'[)'))`

func (r *PostgresBookingRepository) PTs(ctx context.Context, id int64, f catalog.Filter) ([]booking.PTOption, int, error) {
	query := `SELECT p.id,u.full_name,COALESCE(p.specialty,'') specialty,
		(SELECT AVG(rv.rating)::float8 FROM pt_reviews rv JOIN bookings b ON b.id=rv.booking_id JOIN room_sessions rs ON rs.id=b.room_session_id WHERE b.pt_id=p.id AND ` + validReviewScope + `) rating_average,
		(SELECT COUNT(*)::int FROM pt_reviews rv JOIN bookings b ON b.id=rv.booking_id JOIN room_sessions rs ON rs.id=b.room_session_id WHERE b.pt_id=p.id AND ` + validReviewScope + `) review_count
		FROM personal_trainers p JOIN users u ON u.id=p.user_id JOIN roles pr ON pr.id=u.role_id CROSS JOIN room_sessions s JOIN rooms room ON room.id=s.room_id WHERE s.id=$1 AND s.status='OPEN' AND room.status='AVAILABLE' AND s.starts_at>now() AND ` + ptAvailable
	return jsonList[booking.PTOption](ctx, r.pool, query, []any{id}, f)
}
func (r *PostgresBookingRepository) Sessions(ctx context.Context, a identity.User, f catalog.Filter) ([]booking.SessionOption, int, error) {
	return jsonList[booking.SessionOption](ctx, r.pool, `SELECT q.*,GREATEST(q.capacity-q.active_holds,0) remaining,CASE WHEN q.status<>'OPEN' OR r.status<>'AVAILABLE' THEN 'Phòng hoặc ca không nhận đặt lịch.' WHEN q.starts_at<=now() THEN 'Ca đã bắt đầu.' WHEN q.active_holds>=q.capacity THEN 'Ca đã đầy.' WHEN $3::boolean AND NOT EXISTS(SELECT 1 FROM members m JOIN users u ON u.id=m.user_id JOIN roles rr ON rr.id=u.role_id WHERE m.user_id=$4 AND m.status='ACTIVE' AND u.status='ACTIVE' AND rr.code='MEMBER' AND m.membership_expiry>=q.session_date::date) THEN 'Hồ sơ hoặc hạn thẻ chưa đủ điều kiện. Liên hệ lễ tân.' ELSE '' END unavailable_reason FROM (`+sessionSelect+`) q JOIN rooms r ON r.id=q.room_id WHERE q.room_id=$1 AND q.session_date=$2 AND q.ends_at>now()`, []any{f.RoomID, f.Date, a.Role == identity.RoleMember, a.ID}, f)
}

// Common order: room -> member -> optional PT -> slot gate -> session -> booking.
// User rows are read, never locked after a profile row (Identity/profile writers
// lock user first). The member/PT row serializes membership/availability writers.
func bookingLocks(ctx context.Context, tx pgx.Tx, room, member, pt, session int64) error {
	return bookingBatchLocks(ctx, tx, room, member, pt, []int64{session})
}

// bookingBatchLocks preserves the booking lock order while acquiring every
// selected session in chronological order. The slot gate makes booking
// capacity checks serializable with the existing single-booking path.
func bookingBatchLocks(ctx context.Context, tx pgx.Tx, room, member, pt int64, sessions []int64) error {
	for _, item := range []struct {
		table string
		id    int64
	}{{"rooms", room}, {"members", member}, {"personal_trainers", pt}} {
		if item.id == 0 {
			continue
		}
		var id int64
		if err := tx.QueryRow(ctx, `SELECT id FROM `+item.table+` WHERE id=$1 FOR UPDATE`, item.id).Scan(&id); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(841017)`); err != nil {
		return err
	}
	for _, session := range sessions {
		var id int64
		if err := tx.QueryRow(ctx, `SELECT id FROM room_sessions WHERE id=$1 FOR UPDATE`, session).Scan(&id); err != nil {
			return err
		}
	}
	return nil
}
func (r *PostgresBookingRepository) Create(ctx context.Context, a identity.User, member int64, in booking.CreateInput, key, rid string) (booking.Booking, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (booking.Booking, error) {
		var zero booking.Booking
		scope := "booking:create"
		hash := hashJSON(struct {
			Member int64
			Input  booking.CreateInput
		}{member, in})
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, fmt.Sprintf("booking-key:%d:%s", a.ID, key)); err != nil {
			return zero, err
		}
		if err := deleteExpiredIdempotencyKey(ctx, tx, a.ID, scope, key); err != nil {
			return zero, err
		}
		var storedHash string
		var raw []byte
		err := tx.QueryRow(ctx, `SELECT request_hash,response_body FROM idempotency_keys WHERE actor_user_id=$1 AND request_scope=$2 AND idempotency_key=$3`, a.ID, scope, key).Scan(&storedHash, &raw)
		if err == nil {
			if storedHash != hash {
				return zero, booking.Fail("IDEMPOTENCY_CONFLICT", "Khóa yêu cầu đã dùng cho nội dung khác.")
			}
			err = json.Unmarshal(raw, &zero)
			zero.ServerTime = time.Now()
			return zero, err
		}
		if err != pgx.ErrNoRows {
			return zero, err
		}
		if member == 0 {
			if err = tx.QueryRow(ctx, `SELECT id FROM members WHERE user_id=$1`, a.ID).Scan(&member); err != nil {
				return zero, booking.Fail("MEMBERSHIP_INVALID", "Chưa có hồ sơ hội viên. Liên hệ lễ tân.")
			}
		}
		var room int64
		if err = tx.QueryRow(ctx, `SELECT room_id FROM room_sessions WHERE id=$1`, in.RoomSessionID).Scan(&room); err != nil {
			return zero, err
		}
		if err = bookingLocks(ctx, tx, room, member, in.PTID, in.RoomSessionID); err != nil {
			return zero, err
		}
		session, err := jsonOne[facility.Session](ctx, tx, sessionSelect+` WHERE s.id=$1`, in.RoomSessionID)
		if err != nil {
			return zero, err
		}
		start, err := time.Parse(time.RFC3339, session.StartsAt)
		if err != nil {
			return zero, err
		}
		end, err := time.Parse(time.RFC3339, session.EndsAt)
		if err != nil {
			return zero, err
		}
		var now time.Time
		var eligible bool
		err = tx.QueryRow(ctx, `SELECT clock_timestamp(),EXISTS(SELECT 1 FROM members m JOIN users u ON u.id=m.user_id JOIN roles rr ON rr.id=u.role_id WHERE m.id=$1 AND m.status='ACTIVE' AND m.membership_expiry>=$2::date AND u.status='ACTIVE' AND rr.code='MEMBER')`, member, session.SessionDate).Scan(&now, &eligible)
		if err != nil {
			return zero, err
		}
		if !eligible {
			return zero, booking.Fail("MEMBERSHIP_INVALID", "Hồ sơ hoặc hạn thẻ không đủ điều kiện đặt lịch.")
		}
		var status string
		if err = tx.QueryRow(ctx, `SELECT status FROM rooms WHERE id=$1`, room).Scan(&status); err != nil {
			return zero, err
		}
		if session.Status != "OPEN" || status != "AVAILABLE" || !start.After(now) {
			return zero, booking.Fail("SESSION_UNAVAILABLE", "Phòng hoặc ca không còn nhận đặt lịch.")
		}
		if session.ActiveHolds >= session.Capacity {
			return zero, booking.Fail("CAPACITY_FULL", "Ca vừa hết chỗ. Hãy chọn ca khác.")
		}
		var overlap bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id WHERE b.member_id=$1 AND b.status IN ('PENDING_PT','CONFIRMED') AND tstzrange(s.starts_at,s.ends_at,'[)') && tstzrange($2::timestamptz,$3::timestamptz,'[)'))`, member, start, end).Scan(&overlap); err != nil {
			return zero, err
		}
		if overlap {
			return zero, booking.Fail("MEMBER_OVERLAP", "Bạn đã có lịch trùng giờ này.")
		}
		bookingStatus := "CONFIRMED"
		var deadline *time.Time
		if in.PTID > 0 {
			var available bool
			err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM personal_trainers p JOIN users u ON u.id=p.user_id JOIN roles pr ON pr.id=u.role_id CROSS JOIN room_sessions s WHERE p.id=$1 AND s.id=$2 AND `+ptAvailable+`)`, in.PTID, in.RoomSessionID).Scan(&available)
			if err != nil {
				return zero, err
			}
			if !available {
				return zero, booking.Fail("PT_UNAVAILABLE", "PT không còn rảnh toàn bộ ca này. Hãy chọn PT hoặc ca khác.")
			}
			bookingStatus = "PENDING_PT"
			d := booking.Deadline(now, start)
			deadline = &d
		}
		var confirmedAt *time.Time
		if bookingStatus == "CONFIRMED" {
			confirmedAt = &now
		}
		var id int64
		err = tx.QueryRow(ctx, `INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,pt_id,status,respond_by,created_by_user_id,created_at,confirmed_at) VALUES($1,$2,$3::date,$4,NULLIF($5,0),$6,$7,$8,$9,$10) RETURNING id`, member, in.RoomSessionID, session.SessionDate, session.TimeSlotID, in.PTID, bookingStatus, deadline, a.ID, now, confirmedAt).Scan(&id)
		if err != nil {
			return zero, err
		}
		result, err := jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1`, id)
		if err != nil {
			return result, err
		}
		if err = writeAudit(ctx, tx, a.ID, "booking.created", "booking", id, "", nil, result, rid); err != nil {
			return result, err
		}
		data, _ := json.Marshal(result)
		_, err = tx.Exec(ctx, `INSERT INTO idempotency_keys(actor_user_id,request_scope,idempotency_key,request_hash,response_status,response_body,committed_at,expires_at) VALUES($1,$2,$3,$4,201,$5,clock_timestamp(),clock_timestamp()+interval '24 hours')`, a.ID, scope, key, hash, data)
		return result, err
	})
}

type batchSessionStamp struct {
	ID          int64
	RoomID      int64
	SessionDate string
	StartsAt    time.Time
	EndsAt      time.Time
}

type lockedBatchSession struct {
	Session facility.Session
	Start   time.Time
	End     time.Time
}

func loadBatchSessionStamps(ctx context.Context, tx pgx.Tx, ids []int64) ([]batchSessionStamp, error) {
	rows, err := tx.Query(ctx, `SELECT id,room_id,session_date::text,starts_at,ends_at FROM room_sessions WHERE id=ANY($1::bigint[])`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stamps := make([]batchSessionStamp, 0, len(ids))
	for rows.Next() {
		var stamp batchSessionStamp
		if err := rows.Scan(&stamp.ID, &stamp.RoomID, &stamp.SessionDate, &stamp.StartsAt, &stamp.EndsAt); err != nil {
			return nil, err
		}
		stamps = append(stamps, stamp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(stamps) != len(ids) {
		return nil, booking.Fail("BATCH_SESSION_INVALID", "Một hoặc nhiều ca không tồn tại hoặc bị lặp.")
	}
	sort.Slice(stamps, func(i, j int) bool {
		if stamps[i].StartsAt.Equal(stamps[j].StartsAt) {
			return stamps[i].ID < stamps[j].ID
		}
		return stamps[i].StartsAt.Before(stamps[j].StartsAt)
	})
	return stamps, nil
}

func validateBatchSessionStamps(stamps []batchSessionStamp) error {
	if len(stamps) == 0 {
		return booking.Fail("BATCH_SESSION_INVALID", "Cần chọn ít nhất một ca.")
	}
	first := stamps[0]
	for i, stamp := range stamps {
		if stamp.RoomID != first.RoomID || stamp.SessionDate != first.SessionDate {
			return booking.Fail("BATCH_SESSION_INVALID", "Các ca phải thuộc cùng phòng và cùng ngày.")
		}
		if i > 0 && !stamps[i-1].EndsAt.Equal(stamp.StartsAt) {
			return booking.Fail("BATCH_SESSIONS_NOT_CONSECUTIVE", "Các ca phải liền kề theo thời gian.")
		}
	}
	return nil
}

// CreateBatch creates an entire contiguous selection in one transaction. No
// booking is committed if one selected session is no longer eligible, full, or
// unavailable for the requested PT.
func (r *PostgresBookingRepository) CreateBatch(ctx context.Context, a identity.User, member int64, in booking.BatchCreateInput, key, rid string) ([]booking.Booking, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) ([]booking.Booking, error) {
		result := make([]booking.Booking, 0, len(in.RoomSessionIDs))
		scope := "booking:create-batch"
		hash := hashJSON(struct {
			Member int64
			Input  booking.BatchCreateInput
		}{member, in})
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, fmt.Sprintf("booking-batch-key:%d:%s", a.ID, key)); err != nil {
			return nil, err
		}
		if err := deleteExpiredIdempotencyKey(ctx, tx, a.ID, scope, key); err != nil {
			return nil, err
		}
		var storedHash string
		var raw []byte
		err := tx.QueryRow(ctx, `SELECT request_hash,response_body FROM idempotency_keys WHERE actor_user_id=$1 AND request_scope=$2 AND idempotency_key=$3`, a.ID, scope, key).Scan(&storedHash, &raw)
		if err == nil {
			if storedHash != hash {
				return nil, booking.Fail("IDEMPOTENCY_CONFLICT", "Khóa yêu cầu đã dùng cho nội dung khác.")
			}
			if err = json.Unmarshal(raw, &result); err != nil {
				return nil, err
			}
			serverTime := time.Now()
			for i := range result {
				result[i].ServerTime = serverTime
			}
			return result, nil
		}
		if err != pgx.ErrNoRows {
			return nil, err
		}
		if member == 0 {
			if err = tx.QueryRow(ctx, `SELECT id FROM members WHERE user_id=$1`, a.ID).Scan(&member); err != nil {
				return nil, booking.Fail("MEMBERSHIP_INVALID", "Chưa có hồ sơ hội viên. Liên hệ lễ tân.")
			}
		}

		preflight, err := loadBatchSessionStamps(ctx, tx, in.RoomSessionIDs)
		if err != nil {
			return nil, err
		}
		if err = validateBatchSessionStamps(preflight); err != nil {
			return nil, err
		}
		sessionIDs := make([]int64, len(preflight))
		for i, stamp := range preflight {
			sessionIDs[i] = stamp.ID
		}
		if err = bookingBatchLocks(ctx, tx, preflight[0].RoomID, member, in.PTID, sessionIDs); err != nil {
			return nil, err
		}

		locked := make([]lockedBatchSession, 0, len(sessionIDs))
		for _, sessionID := range sessionIDs {
			session, err := jsonOne[facility.Session](ctx, tx, sessionSelect+` WHERE s.id=$1`, sessionID)
			if err != nil {
				return nil, err
			}
			start, err := time.Parse(time.RFC3339, session.StartsAt)
			if err != nil {
				return nil, err
			}
			end, err := time.Parse(time.RFC3339, session.EndsAt)
			if err != nil {
				return nil, err
			}
			locked = append(locked, lockedBatchSession{Session: session, Start: start, End: end})
		}
		sort.Slice(locked, func(i, j int) bool {
			if locked[i].Start.Equal(locked[j].Start) {
				return locked[i].Session.ID < locked[j].Session.ID
			}
			return locked[i].Start.Before(locked[j].Start)
		})
		lockedStamps := make([]batchSessionStamp, len(locked))
		for i, item := range locked {
			lockedStamps[i] = batchSessionStamp{ID: item.Session.ID, RoomID: item.Session.RoomID, SessionDate: item.Session.SessionDate, StartsAt: item.Start, EndsAt: item.End}
		}
		if err = validateBatchSessionStamps(lockedStamps); err != nil {
			return nil, err
		}

		var now time.Time
		var eligible bool
		err = tx.QueryRow(ctx, `SELECT clock_timestamp(),EXISTS(SELECT 1 FROM members m JOIN users u ON u.id=m.user_id JOIN roles rr ON rr.id=u.role_id WHERE m.id=$1 AND m.status='ACTIVE' AND m.membership_expiry>=$2::date AND u.status='ACTIVE' AND rr.code='MEMBER')`, member, locked[0].Session.SessionDate).Scan(&now, &eligible)
		if err != nil {
			return nil, err
		}
		if !eligible {
			return nil, booking.Fail("MEMBERSHIP_INVALID", "Hồ sơ hoặc hạn thẻ không đủ điều kiện đặt lịch.")
		}
		var roomStatus string
		if err = tx.QueryRow(ctx, `SELECT status FROM rooms WHERE id=$1`, locked[0].Session.RoomID).Scan(&roomStatus); err != nil {
			return nil, err
		}
		for _, item := range locked {
			if item.Session.Status != "OPEN" || roomStatus != "AVAILABLE" || !item.Start.After(now) {
				return nil, booking.Fail("SESSION_UNAVAILABLE", "Phòng hoặc một trong các ca không còn nhận đặt lịch.")
			}
			if item.Session.ActiveHolds >= item.Session.Capacity {
				return nil, booking.Fail("CAPACITY_FULL", "Một trong các ca vừa hết chỗ. Hãy chọn ca khác.")
			}
		}
		var overlap bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id WHERE b.member_id=$1 AND b.status IN ('PENDING_PT','CONFIRMED') AND tstzrange(s.starts_at,s.ends_at,'[)') && tstzrange($2::timestamptz,$3::timestamptz,'[)'))`, member, locked[0].Start, locked[len(locked)-1].End).Scan(&overlap); err != nil {
			return nil, err
		}
		if overlap {
			return nil, booking.Fail("MEMBER_OVERLAP", "Bạn đã có lịch trùng giờ trong dải ca đã chọn.")
		}
		if in.PTID > 0 {
			for _, item := range locked {
				var available bool
				err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM personal_trainers p JOIN users u ON u.id=p.user_id JOIN roles pr ON pr.id=u.role_id CROSS JOIN room_sessions s WHERE p.id=$1 AND s.id=$2 AND `+ptAvailable+`)`, in.PTID, item.Session.ID).Scan(&available)
				if err != nil {
					return nil, err
				}
				if !available {
					return nil, booking.Fail("PT_UNAVAILABLE", "PT không còn rảnh toàn bộ dải ca đã chọn. Hãy chọn PT hoặc ca khác.")
				}
			}
		}

		bookingStatus := "CONFIRMED"
		var confirmedAt *time.Time
		if in.PTID > 0 {
			bookingStatus = "PENDING_PT"
		} else {
			confirmedAt = &now
		}
		for _, item := range locked {
			var deadline *time.Time
			if in.PTID > 0 {
				d := booking.Deadline(now, item.Start)
				deadline = &d
			}
			var id int64
			err = tx.QueryRow(ctx, `INSERT INTO bookings(member_id,room_session_id,session_date,time_slot_id,pt_id,status,respond_by,created_by_user_id,created_at,confirmed_at) VALUES($1,$2,$3::date,$4,NULLIF($5,0),$6,$7,$8,$9,$10) RETURNING id`, member, item.Session.ID, item.Session.SessionDate, item.Session.TimeSlotID, in.PTID, bookingStatus, deadline, a.ID, now, confirmedAt).Scan(&id)
			if err != nil {
				return nil, err
			}
			created, err := jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1`, id)
			if err != nil {
				return nil, err
			}
			if err = writeAudit(ctx, tx, a.ID, "booking.created", "booking", id, "", nil, created, rid); err != nil {
				return nil, err
			}
			result = append(result, created)
		}
		data, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx, `INSERT INTO idempotency_keys(actor_user_id,request_scope,idempotency_key,request_hash,response_status,response_body,committed_at,expires_at) VALUES($1,$2,$3,$4,201,$5,clock_timestamp(),clock_timestamp()+interval '24 hours')`, a.ID, scope, key, hash, data)
		return result, err
	})
}
func (r *PostgresBookingRepository) Act(ctx context.Context, a identity.User, id int64, in booking.ActionInput, cancel bool, rid string) (booking.Booking, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (booking.Booking, error) {
		before, err := jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1 AND `+bookingScope(a), id)
		if err != nil {
			return before, err
		}
		if err = bookingLocks(ctx, tx, before.RoomID, before.MemberID, before.PTID, before.RoomSessionID); err != nil {
			return before, err
		}
		var locked int64
		if err = tx.QueryRow(ctx, `SELECT id FROM bookings WHERE id=$1 FOR UPDATE`, id).Scan(&locked); err != nil {
			return before, err
		}
		before, err = jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1 AND `+bookingScope(a), id)
		if err != nil {
			return before, err
		}
		now := before.ServerTime
		if before.Status != "PENDING_PT" && before.Status != "CONFIRMED" {
			return before, booking.Fail("BOOKING_INVALID_STATE", "Lịch đã được xử lý ở nơi khác. Hãy tải lại.")
		}
		if before.Status == "PENDING_PT" && before.RespondBy != nil && !now.Before(*before.RespondBy) {
			return before, booking.Fail("BOOKING_EXPIRED", "Yêu cầu đã hết hạn phản hồi.")
		}
		next := in.Action
		if cancel {
			if !now.Before(before.EndsAt) {
				return before, booking.Fail("BOOKING_ENDED", "Buổi tập đã kết thúc.")
			}
			var attended bool
			if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM booking_attendance WHERE booking_id=$1)`, id).Scan(&attended); err != nil {
				return before, err
			}
			if attended {
				return before, booking.Fail("BOOKING_ATTENDED", "Buổi đã điểm danh không thể hủy bằng thao tác thường.")
			}
			if a.Role == identity.RoleMember && !booking.CanCancel(now, before.StartsAt) {
				return before, booking.Fail("CANCEL_TOO_LATE", "Chỉ được tự hủy trước giờ tập ít nhất 2 giờ.")
			}
			if (a.Role != identity.RoleMember) && in.Reason == "" {
				return before, identity.ErrValidation
			}
			next = "CANCELLED"
			if in.Reason == "" {
				in.Reason = "Hội viên tự hủy"
			}
			source, validRole := booking.CancellationProvenance(a.Role)
			if !validRole {
				return before, identity.ErrForbidden
			}
			actorRole := a.Role
			_, err = tx.Exec(ctx, `UPDATE bookings SET status=$2,cancellation_reason=$3,cancelled_by_user_id=$4,cancelled_at=$5,updated_at=$5,cancellation_source=$6,cancelled_actor_role=$7 WHERE id=$1`, id, next, in.Reason, a.ID, now, source, actorRole)
		} else {
			if before.Status != "PENDING_PT" {
				return before, booking.Fail("BOOKING_INVALID_STATE", "Chỉ được phản hồi yêu cầu đang chờ PT.")
			}
			next = "CONFIRMED"
			if in.Action == "REJECT" {
				next = "REJECTED"
			}
			var confirmedAt *time.Time
			if next == "CONFIRMED" {
				confirmedAt = &now
			}
			_, err = tx.Exec(ctx, `UPDATE bookings SET status=$2,cancellation_reason=NULLIF($3,''),confirmed_at=$4,updated_at=$5 WHERE id=$1`, id, next, in.Reason, confirmedAt, now)
		}
		if err != nil {
			return before, err
		}
		after, err := jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1`, id)
		if err != nil {
			return after, err
		}
		err = writeAudit(ctx, tx, a.ID, "booking."+next, "booking", id, in.Reason, before, after, rid)
		return after, err
	})
}
func (r *PostgresBookingRepository) Expire(ctx context.Context) error {
	rows, err := r.pool.Query(ctx, `SELECT id FROM bookings WHERE status='PENDING_PT' AND respond_by<=clock_timestamp() ORDER BY id LIMIT 1000`)
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
			b, err := jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1`, id)
			if err != nil {
				return false, err
			}
			if err = bookingLocks(ctx, tx, b.RoomID, b.MemberID, b.PTID, b.RoomSessionID); err != nil {
				return false, err
			}
			var changed int64
			err = tx.QueryRow(ctx, `UPDATE bookings SET status='EXPIRED',updated_at=clock_timestamp() WHERE id=$1 AND status='PENDING_PT' AND respond_by<=clock_timestamp() RETURNING id`, id).Scan(&changed)
			if err == pgx.ErrNoRows {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			after, err := jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1`, id)
			if err != nil {
				return false, err
			}
			return true, writeAudit(ctx, tx, 0, "booking.EXPIRED", "booking", id, "Hết hạn phản hồi", b, after, "booking-expiration")
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
