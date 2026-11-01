package repository

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"strconv"
	"time"
)

type CalendarRepository interface {
	Replay(context.Context, int64, int64, facility.CalendarMutation, string) (facility.CalendarResult, bool, error)
	Snapshot(context.Context, int64) (facility.CalendarSnapshot, error)
	Preview(context.Context, int64, facility.CalendarMutation) (facility.CalendarPreview, string, error)
	Commit(context.Context, int64, int64, facility.CalendarMutation, string, time.Time, string, string) (facility.CalendarResult, error)
	MaterializeAll(context.Context) error
}
type PostgresCalendarRepository struct {
	pool *pgxpool.Pool
	now  func() time.Time
}

func NewCalendarRepository(pool *pgxpool.Pool) *PostgresCalendarRepository {
	return &PostgresCalendarRepository{pool: pool, now: time.Now}
}

// A committed request can be replayed after a restart without trusting an old preview signature.
func (r *PostgresCalendarRepository) Replay(ctx context.Context, id, actor int64, m facility.CalendarMutation, key string) (facility.CalendarResult, bool, error) {
	var result facility.CalendarResult
	var hash string
	var body []byte
	err := r.pool.QueryRow(ctx, `SELECT request_hash,response_body FROM idempotency_keys WHERE actor_user_id=$1 AND request_scope=$2 AND idempotency_key=$3 AND expires_at>clock_timestamp()`, actor, fmt.Sprintf("room-calendar:%d", id), key).Scan(&hash, &body)
	if err == pgx.ErrNoRows {
		return result, false, nil
	}
	if err != nil {
		return result, false, err
	}
	if hash != hashJSON(m) {
		return result, true, identity.Conflict("Idempotency-Key đã dùng cho nội dung khác.")
	}
	err = json.Unmarshal(body, &result)
	return result, true, err
}

const versionsSelect = `SELECT v.id,v.room_id,v.effective_from::text,COALESCE(v.effective_until::text,'') effective_until,v.status,v.reason,v.created_at,COALESCE((SELECT jsonb_agg(jsonb_build_object('weekday',h.weekday,'start_minute',h.start_minute,'end_minute',h.end_minute) ORDER BY h.weekday,h.start_minute) FROM room_weekly_hours h WHERE h.version_id=v.id),'[]') hours FROM room_schedule_versions v`
const exceptionsSelect = `SELECT e.id,e.room_id,e.exception_date::text date,e.mode,e.closure_type,e.reason,e.status,e.updated_at,COALESCE((SELECT jsonb_agg(jsonb_build_object('weekday',0,'start_minute',h.start_minute,'end_minute',h.end_minute) ORDER BY h.start_minute) FROM room_exception_hours h WHERE h.exception_id=e.id),'[]') hours FROM room_calendar_exceptions e`

// jsonOne wraps a row in to_jsonb; scalar arrays must be scanned directly instead.
func arrayJSON[T any](ctx context.Context, db queryRower, query string, args ...any) ([]T, error) {
	var data []byte
	err := db.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(q)),'[]') FROM (`+query+`) q`, args...).Scan(&data)
	if err != nil {
		return nil, err
	}
	var out []T
	err = json.Unmarshal(data, &out)
	return out, err
}
func calendarSnapshot(ctx context.Context, db queryRower, id int64) (facility.CalendarSnapshot, error) {
	s := facility.CalendarSnapshot{}
	var err error
	if s.Room, err = jsonOne[facility.Room](ctx, db, roomSelect+` WHERE id=$1`, id); err != nil {
		return s, err
	}
	if s.Versions, err = arrayJSON[facility.ScheduleVersion](ctx, db, versionsSelect+` WHERE v.room_id=$1 ORDER BY v.effective_from,v.id`, id); err != nil {
		return s, err
	}
	if s.Exceptions, err = arrayJSON[facility.CalendarException](ctx, db, exceptionsSelect+` WHERE e.room_id=$1 ORDER BY e.exception_date,e.id`, id); err != nil {
		return s, err
	}
	if s.Slots, err = arrayJSON[facility.TimeSlot](ctx, db, slotSelect+` ORDER BY id`); err != nil {
		return s, err
	}
	s.Sessions, err = arrayJSON[facility.Session](ctx, db, sessionSelect+` WHERE s.room_id=$1 AND s.ends_at>now() ORDER BY s.id`, id)
	return s, err
}
func (r *PostgresCalendarRepository) Snapshot(ctx context.Context, id int64) (facility.CalendarSnapshot, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (facility.CalendarSnapshot, error) {
		if err := lockCalendar(ctx, tx, id); err != nil {
			return facility.CalendarSnapshot{}, err
		}
		return calendarSnapshot(ctx, tx, id)
	})
}
func lockCalendar(ctx context.Context, tx pgx.Tx, id int64) error {
	var exists int64
	if err := tx.QueryRow(ctx, `SELECT id FROM rooms WHERE id=$1 FOR UPDATE`, id).Scan(&exists); err != nil {
		return err
	}
	// Every slot writer takes this lock as well, including inserts, avoiding phantoms.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(841017)`); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT id FROM room_sessions WHERE room_id=$1 AND ends_at>now() ORDER BY id FOR UPDATE`, id)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var locked int64
		if err = rows.Scan(&locked); err != nil {
			return err
		}
	}
	return rows.Err()
}
func hashJSON(value any) string {
	b, _ := json.Marshal(value)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}
func prospective(s facility.CalendarSnapshot, m facility.CalendarMutation) (facility.CalendarSnapshot, error) {
	// Copy slices: previews cannot mutate loaded snapshots or reuse another actor's draft.
	s.Versions = append([]facility.ScheduleVersion{}, s.Versions...)
	s.Exceptions = append([]facility.CalendarException{}, s.Exceptions...)
	switch m.Kind {
	case "WEEKLY":
		until := ""
		for i, v := range s.Versions {
			if v.Status != "ACTIVE" {
				continue
			}
			if v.EffectiveFrom > m.EffectiveFrom && (until == "" || v.EffectiveFrom < until) {
				until = v.EffectiveFrom
			}
			if v.EffectiveFrom == m.EffectiveFrom {
				s.Versions[i].Status = "SUPERSEDED"
			} else if v.EffectiveFrom < m.EffectiveFrom && (v.EffectiveUntil == "" || v.EffectiveUntil > m.EffectiveFrom) {
				s.Versions[i].EffectiveUntil = m.EffectiveFrom
			}
		}
		s.Versions = append(s.Versions, facility.ScheduleVersion{ID: -1, RoomID: s.Room.ID, EffectiveFrom: m.EffectiveFrom, EffectiveUntil: until, Hours: m.Hours, Status: "ACTIVE", Reason: m.Reason})
	case "EXCEPTION":
		found := false
		for i, e := range s.Exceptions {
			if e.ID == m.ExceptionID && e.Status == "ACTIVE" {
				if e.Date != m.Date {
					return s, identity.Conflict("Ngày ngoại lệ không được đổi; hãy rút và tạo ngoại lệ mới.")
				}
				s.Exceptions[i].Status = "WITHDRAWN"
				found = true
			} else if e.Date == m.Date && e.Status == "ACTIVE" {
				return s, identity.Conflict("Ngày này đã có ngoại lệ; hãy chỉnh sửa ngoại lệ hiện có.")
			}
		}
		if m.ExceptionID > 0 && !found {
			return s, identity.ErrNotFound
		}
		s.Exceptions = append(s.Exceptions, facility.CalendarException{ID: -1, RoomID: s.Room.ID, Date: m.Date, Mode: m.Mode, ClosureType: m.ClosureType, Reason: m.Reason, Status: "ACTIVE", Hours: m.Hours})
	case "WITHDRAW":
		found := false
		for i, e := range s.Exceptions {
			if e.ID == m.ExceptionID && e.Status == "ACTIVE" {
				s.Exceptions[i].Status = "WITHDRAWN"
				found = true
			}
		}
		if !found {
			return s, identity.ErrNotFound
		}
	}
	return s, nil
}
func slotMinutes(value string) int {
	if value == "24:00" {
		return 1440
	}
	t, err := time.Parse("15:04", value)
	if err != nil {
		return -1
	}
	return t.Hour()*60 + t.Minute()
}
func sessionMinutes(s facility.Session) (int, int, error) {
	a, err := time.Parse(time.RFC3339, s.StartsAt)
	if err != nil {
		return 0, 0, err
	}
	b, err := time.Parse(time.RFC3339, s.EndsAt)
	if err != nil {
		return 0, 0, err
	}
	start := facility.MinuteTime(s.SessionDate, 0)
	return int(a.Sub(start) / time.Minute), int(b.Sub(start) / time.Minute), nil
}

type candidate struct {
	date string
	slot facility.TimeSlot
	day  facility.CalendarDay
}
type calendarPlan struct {
	preview         facility.CalendarPreview
	creates         []candidate
	closes, reopens []int64
}

func planCalendar(s facility.CalendarSnapshot, m facility.CalendarMutation, now time.Time, fillOnly bool) (calendarPlan, error) {
	p := calendarPlan{preview: facility.CalendarPreview{Revision: s.Room.CalendarRevision, Changes: []facility.CalendarChange{}, Conflicts: []facility.CalendarChange{}}}
	proposed, err := prospective(s, m)
	if err != nil {
		return p, err
	}
	if m.Kind == "WEEKLY" && m.EffectiveFrom == facility.Today(now) && len(s.Versions) > 0 {
		return p, identity.Conflict("Lịch tuần đã có chỉ đổi từ ngày mai; hôm nay dùng ngoại lệ.")
	}
	for _, e := range s.Exceptions {
		if m.Kind == "WITHDRAW" && e.ID == m.ExceptionID && e.Date < facility.Today(now) {
			return p, identity.Conflict("Không thể rút ngoại lệ đã qua.")
		}
	}
	occupied := map[string]bool{}
	openRanges := map[string][][2]int{}
	for _, session := range s.Sessions {
		occupied[session.SessionDate+"/"+strconv.FormatInt(session.TimeSlotID, 10)] = true
		a, b, e := sessionMinutes(session)
		if e != nil {
			return p, e
		}
		start := facility.MinuteTime(session.SessionDate, a)
		fits := facility.Contains(proposed.Day(session.SessionDate).Hours, a, b)
		action := ""
		affected := m.Kind == "WEEKLY" && session.SessionDate >= m.EffectiveFrom
		if m.Kind == "EXCEPTION" {
			affected = session.SessionDate == m.Date
		}
		if m.Kind == "WITHDRAW" {
			for _, exception := range s.Exceptions {
				if exception.ID == m.ExceptionID {
					affected = session.SessionDate == exception.Date
				}
			}
		}
		// Existing future versions delimit the scope of a newly inserted weekly version.
		if m.Kind == "WEEKLY" {
			for _, version := range s.Versions {
				if version.Status == "ACTIVE" && version.EffectiveFrom > m.EffectiveFrom && session.SessionDate >= version.EffectiveFrom {
					affected = false
				}
			}
		}
		if !fillOnly && affected && start.After(now) {
			if session.Status == "OPEN" && !fits {
				action = "CLOSE"
				p.closes = append(p.closes, session.ID)
			} else if session.Status != "OPEN" && session.ClosureSource == "CALENDAR" && fits {
				action = "REOPEN"
				p.reopens = append(p.reopens, session.ID)
			}
		}
		if action != "" {
			change := facility.CalendarChange{SessionID: session.ID, Date: session.SessionDate, Start: session.StartsAt, End: session.EndsAt, Holds: session.ActiveHolds, Action: action}
			p.preview.Changes = append(p.preview.Changes, change)
			if session.ActiveHolds > 0 && action == "CLOSE" {
				p.preview.Conflicts = append(p.preview.Conflicts, change)
			}
		}
		if (session.Status == "OPEN" && action != "CLOSE") || action == "REOPEN" {
			for _, span := range openRanges[session.SessionDate] {
				if a < span[1] && b > span[0] {
					return p, identity.Conflict("Mở lại phiên sẽ chồng thời gian phiên đang mở.")
				}
			}
			openRanges[session.SessionDate] = append(openRanges[session.SessionDate], [2]int{a, b})
		}
	}
	days := m.Days
	if days == 0 {
		days = 28
	}
	from := facility.Today(now)
	if m.Kind == "WEEKLY" && m.EffectiveFrom > from {
		from = m.EffectiveFrom
	}
	base := facility.MinuteTime(from, 0)
	for offset := 0; offset < days; offset++ {
		date := base.AddDate(0, 0, offset).Format("2006-01-02")
		day := proposed.Day(date)
		for _, slot := range s.Slots {
			a, b := slotMinutes(slot.StartTime), slotMinutes(slot.EndTime)
			if !slot.IsActive || occupied[date+"/"+strconv.FormatInt(slot.ID, 10)] || !facility.Contains(day.Hours, a, b) || !facility.MinuteTime(date, a).After(now) {
				continue
			}
			overlap := false
			for _, span := range openRanges[date] {
				if a < span[1] && b > span[0] {
					overlap = true
				}
			}
			if overlap {
				continue
			}
			p.creates = append(p.creates, candidate{date, slot, day})
			openRanges[date] = append(openRanges[date], [2]int{a, b})
		}
		p.preview.PreparedThrough = date
	}
	p.preview.Creates = len(p.creates)
	return p, nil
}
func (r *PostgresCalendarRepository) Preview(ctx context.Context, id int64, m facility.CalendarMutation) (facility.CalendarPreview, string, error) {
	type result struct {
		p      facility.CalendarPreview
		finger string
	}
	v, err := inTransaction(ctx, r.pool, func(tx pgx.Tx) (result, error) {
		if err := lockCalendar(ctx, tx, id); err != nil {
			return result{}, err
		}
		s, err := calendarSnapshot(ctx, tx, id)
		if err != nil {
			return result{}, err
		}
		if s.Room.CalendarRevision != m.BaseRevision {
			return result{}, identity.Conflict("Lịch đã thay đổi. Tải lại trước khi xem trước.")
		}
		p, err := planCalendar(s, m, r.now(), false)
		return result{p.preview, hashJSON(s)}, err
	})
	return v.p, v.finger, err
}
func writeCalendar(ctx context.Context, tx pgx.Tx, id, actor int64, m facility.CalendarMutation) error {
	switch m.Kind {
	case "WEEKLY":
		var until *string
		if err := tx.QueryRow(ctx, `SELECT min(effective_from)::text FROM room_schedule_versions WHERE room_id=$1 AND status='ACTIVE' AND effective_from>$2::date`, id, m.EffectiveFrom).Scan(&until); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE room_schedule_versions SET status='SUPERSEDED' WHERE room_id=$1 AND status='ACTIVE' AND effective_from=$2::date`, id, m.EffectiveFrom); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE room_schedule_versions SET effective_until=$2::date WHERE room_id=$1 AND status='ACTIVE' AND effective_from<$2::date AND (effective_until IS NULL OR effective_until>$2::date)`, id, m.EffectiveFrom); err != nil {
			return err
		}
		var version int64
		if err := tx.QueryRow(ctx, `INSERT INTO room_schedule_versions(room_id,effective_from,effective_until,published_by_user_id,reason,payload_hash) VALUES($1,$2::date,$3::date,$4,$5,$6) RETURNING id`, id, m.EffectiveFrom, until, actor, m.Reason, hashJSON(m)).Scan(&version); err != nil {
			return err
		}
		for _, h := range m.Hours {
			if _, err := tx.Exec(ctx, `INSERT INTO room_weekly_hours(version_id,weekday,start_minute,end_minute) VALUES($1,$2,$3,$4)`, version, h.Weekday, h.Start, h.End); err != nil {
				return err
			}
		}
	case "EXCEPTION":
		if m.ExceptionID > 0 {
			if _, err := tx.Exec(ctx, `UPDATE room_calendar_exceptions SET status='WITHDRAWN',updated_by_user_id=$2,updated_at=now() WHERE id=$1 AND room_id=$3 AND status='ACTIVE'`, m.ExceptionID, actor, id); err != nil {
				return err
			}
		}
		var exception int64
		if err := tx.QueryRow(ctx, `INSERT INTO room_calendar_exceptions(room_id,exception_date,mode,closure_type,reason,created_by_user_id,updated_by_user_id) VALUES($1,$2::date,$3,$4,$5,$6,$6) RETURNING id`, id, m.Date, m.Mode, m.ClosureType, m.Reason, actor).Scan(&exception); err != nil {
			return err
		}
		for _, h := range m.Hours {
			if _, err := tx.Exec(ctx, `INSERT INTO room_exception_hours(exception_id,start_minute,end_minute) VALUES($1,$2,$3)`, exception, h.Start, h.End); err != nil {
				return err
			}
		}
	case "WITHDRAW":
		if _, err := tx.Exec(ctx, `UPDATE room_calendar_exceptions SET status='WITHDRAWN',updated_by_user_id=$2,updated_at=now() WHERE id=$1 AND room_id=$3 AND status='ACTIVE'`, m.ExceptionID, actor, id); err != nil {
			return err
		}
	}
	return nil
}
func applyPlan(ctx context.Context, tx pgx.Tx, id int64, p calendarPlan, s facility.CalendarSnapshot, now time.Time) (facility.CalendarResult, error) {
	result := facility.CalendarResult{Revision: s.Room.CalendarRevision, PreparedThrough: p.preview.PreparedThrough}
	for _, sessionID := range p.closes {
		if _, err := tx.Exec(ctx, `UPDATE room_sessions SET status='CLOSED',closure_source='CALENDAR',change_reason='Lịch hoạt động phòng',updated_at=now() WHERE id=$1`, sessionID); err != nil {
			return result, err
		}
		result.Closed++
	}
	for _, sessionID := range p.reopens {
		if _, err := tx.Exec(ctx, `UPDATE room_sessions SET status='OPEN',closure_source=NULL,change_reason='Lịch hoạt động phòng',updated_at=now() WHERE id=$1`, sessionID); err != nil {
			return result, err
		}
		result.Reopened++
	}
	for _, c := range p.creates {
		day := s.Day(c.date)
		_, err := tx.Exec(ctx, `INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,origin,schedule_version_id,calendar_exception_id) VALUES($1,$2,$3::date,$4,$5,$6,'CALENDAR',NULLIF($7,0),NULLIF($8,0)) ON CONFLICT (room_id,session_date,time_slot_id) DO NOTHING`, id, c.slot.ID, c.date, facility.MinuteTime(c.date, slotMinutes(c.slot.StartTime)), facility.MinuteTime(c.date, slotMinutes(c.slot.EndTime)), s.Room.DefaultCapacity, day.VersionID, day.ExceptionID)
		if err != nil {
			return result, err
		}
		result.Created++
	}
	// A future-only publish does not falsely mark an unconfigured gap as prepared.
	through := p.preview.PreparedThrough
	for date := facility.Today(now); date <= through; date = facility.MinuteTime(date, 0).AddDate(0, 0, 1).Format("2006-01-02") {
		if s.Day(date).Source == "UNCONFIGURED" {
			through = ""
			break
		}
	}
	result.PreparedThrough = through
	_, err := tx.Exec(ctx, `UPDATE rooms SET calendar_prepared_through=CASE WHEN $2='' THEN calendar_prepared_through ELSE GREATEST(calendar_prepared_through,$2::date) END,calendar_job_error=NULL WHERE id=$1`, id, through)
	return result, err
}
func (r *PostgresCalendarRepository) Commit(ctx context.Context, id, actor int64, m facility.CalendarMutation, fingerprint string, expires time.Time, key, rid string) (facility.CalendarResult, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (facility.CalendarResult, error) {
		empty := facility.CalendarResult{}
		if err := lockCalendar(ctx, tx, id); err != nil {
			return empty, err
		}
		scope := fmt.Sprintf("room-calendar:%d", id)
		hash := hashJSON(m)
		if err := deleteExpiredIdempotencyKey(ctx, tx, actor, scope, key); err != nil {
			return empty, err
		}
		var oldHash string
		var oldBody []byte
		err := tx.QueryRow(ctx, `SELECT request_hash,response_body FROM idempotency_keys WHERE actor_user_id=$1 AND request_scope=$2 AND idempotency_key=$3`, actor, scope, key).Scan(&oldHash, &oldBody)
		if err == nil {
			if oldHash != hash {
				return empty, identity.Conflict("Idempotency-Key đã dùng cho nội dung khác.")
			}
			err = json.Unmarshal(oldBody, &empty)
			return empty, err
		}
		if err != pgx.ErrNoRows {
			return empty, err
		}
		if !r.now().Before(expires) {
			return empty, identity.Conflict("Preview đã hết hạn. Hãy xem trước lại.")
		}
		s, err := calendarSnapshot(ctx, tx, id)
		if err != nil {
			return empty, err
		}
		if s.Room.CalendarRevision != m.BaseRevision || hashJSON(s) != fingerprint {
			return empty, identity.Conflict("Lịch, ca hoặc booking đã thay đổi. Tải lại và xem trước lại.")
		}
		p, err := planCalendar(s, m, r.now(), false)
		if err != nil {
			return empty, err
		}
		if len(p.preview.Conflicts) > 0 {
			return empty, identity.Conflict("Có booking đang giữ chỗ trong phiên bị ảnh hưởng. Xử lý booking rồi xem trước lại.")
		}
		if err = writeCalendar(ctx, tx, id, actor, m); err != nil {
			return empty, err
		}
		if _, err = tx.Exec(ctx, `UPDATE rooms SET calendar_revision=calendar_revision+1,calendar_required=CASE WHEN $2='WEEKLY' THEN TRUE ELSE calendar_required END WHERE id=$1`, id, m.Kind); err != nil {
			return empty, err
		}
		after, err := calendarSnapshot(ctx, tx, id)
		if err != nil {
			return empty, err
		}
		result, err := applyPlan(ctx, tx, id, p, after, r.now())
		if err != nil {
			return empty, err
		}
		if err = writeAudit(ctx, tx, actor, "room_calendar."+m.Kind, "room", id, m.Reason, s, struct {
			Mutation facility.CalendarMutation
			Result   facility.CalendarResult
		}{m, result}, rid); err != nil {
			return empty, err
		}
		body, _ := json.Marshal(result)
		_, err = tx.Exec(ctx, `INSERT INTO idempotency_keys(actor_user_id,request_scope,idempotency_key,request_hash,response_status,response_body,committed_at,expires_at) VALUES($1,$2,$3,$4,200,$5,now(),now()+interval '24 hours')`, actor, scope, key, hash, body)
		return result, err
	})
}
func (r *PostgresCalendarRepository) MaterializeAll(ctx context.Context) error {
	rows, err := r.pool.Query(ctx, `SELECT id FROM rooms WHERE status='AVAILABLE' AND EXISTS(SELECT 1 FROM room_schedule_versions v WHERE v.room_id=rooms.id AND v.status='ACTIVE') ORDER BY id`)
	if err != nil {
		return err
	}
	ids := []int64{}
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
	var first error
	for _, id := range ids {
		_, err = inTransaction(ctx, r.pool, func(tx pgx.Tx) (facility.CalendarResult, error) {
			if err := lockCalendar(ctx, tx, id); err != nil {
				return facility.CalendarResult{}, err
			}
			s, err := calendarSnapshot(ctx, tx, id)
			if err != nil {
				return facility.CalendarResult{}, err
			}
			p, err := planCalendar(s, facility.CalendarMutation{Kind: "EXTEND", Days: 28}, r.now(), true)
			if err != nil {
				return facility.CalendarResult{}, err
			}
			result, err := applyPlan(ctx, tx, id, p, s, r.now())
			if err != nil {
				return result, err
			}
			if result.Created > 0 {
				err = writeAudit(ctx, tx, 0, "room_calendar.SYSTEM_materialized", "room", id, "Tự bổ sung phiên", nil, result, "calendar-job")
			}
			return result, err
		})
		if err != nil {
			if first == nil {
				first = err
			}
			errorCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			_, _ = r.pool.Exec(errorCtx, `UPDATE rooms SET calendar_job_error='Không thể chuẩn bị phiên. Kiểm tra API log và chạy lại sinh phiên.' WHERE id=$1`, id)
			cancel()
		}
	}
	return first
}

// Used by manual creation/reopening inside the same room lock as publishing.
func requireCalendarSession(ctx context.Context, tx pgx.Tx, roomID int64, date string, start, end int) error {
	s, err := calendarSnapshot(ctx, tx, roomID)
	if err != nil {
		return err
	}
	if !s.Room.CalendarRequired {
		return nil
	}
	if !facility.Contains(s.Day(date).Hours, start, end) {
		return identity.Conflict("Phiên phải nằm hoàn toàn trong giờ mở cửa đã công bố của phòng.")
	}
	return nil
}
