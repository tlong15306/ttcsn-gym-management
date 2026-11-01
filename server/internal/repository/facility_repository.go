package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"time"
)

type FacilityRepository interface {
	Slot(context.Context, int64) (facility.TimeSlot, error)
	UpdateSlot(context.Context, int64, facility.SlotInput, int64, string) (facility.TimeSlot, error)
	Rooms(context.Context, catalog.Filter) ([]facility.Room, int, error)
	Room(context.Context, int64) (facility.Room, error)
	RoomImpact(context.Context, int64) (facility.Impact, error)
	SaveRoom(context.Context, int64, facility.RoomInput, int64, string) (facility.Room, error)
	Slots(context.Context, catalog.Filter) ([]facility.TimeSlot, int, error)
	CreateSlot(context.Context, facility.SlotInput, int64, string) (facility.TimeSlot, error)
	SlotStatus(context.Context, int64, facility.SlotStatusInput, int64, string) (facility.TimeSlot, error)
	Sessions(context.Context, catalog.Filter) ([]facility.Session, int, error)
	Session(context.Context, int64) (facility.Session, error)
	SaveSession(context.Context, int64, facility.SessionInput, int64, string) (facility.Session, error)
}

func (r *PostgresFacilityRepository) Slot(ctx context.Context, id int64) (facility.TimeSlot, error) {
	return jsonOne[facility.TimeSlot](ctx, r.pool, slotSelect+` WHERE id=$1`, id)
}
func (r *PostgresFacilityRepository) UpdateSlot(ctx context.Context, id int64, in facility.SlotInput, actorID int64, rid string) (facility.TimeSlot, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (facility.TimeSlot, error) {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(841017)`); err != nil {
			return facility.TimeSlot{}, err
		}
		var locked int64
		if err := tx.QueryRow(ctx, `SELECT id FROM time_slots WHERE id=$1 FOR UPDATE`, id).Scan(&locked); err != nil {
			return facility.TimeSlot{}, err
		}
		before, err := jsonOne[facility.TimeSlot](ctx, tx, slotSelect+` WHERE id=$1`, id)
		if err != nil {
			return before, err
		}
		var used bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM room_sessions WHERE time_slot_id=$1)`, id).Scan(&used); err != nil {
			return before, err
		}
		if used {
			return before, identity.Conflict("Ca đã được dùng trong phiên. Hãy ngừng ca cũ và tạo ca mới.")
		}
		if _, err = tx.Exec(ctx, `UPDATE time_slots SET name=$2,start_time=$3::time,end_time=$4::time WHERE id=$1`, id, in.Name, in.StartTime, in.EndTime); err != nil {
			return before, err
		}
		after, err := jsonOne[facility.TimeSlot](ctx, tx, slotSelect+` WHERE id=$1`, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, actorID, "time_slot.updated", "time_slot", id, in.Reason, before, after, rid); err != nil {
			return after, err
		}
		return after, nil
	})
}

type PostgresFacilityRepository struct{ pool *pgxpool.Pool }

func NewFacilityRepository(pool *pgxpool.Pool) *PostgresFacilityRepository {
	return &PostgresFacilityRepository{pool}
}

const roomSelect = `SELECT id,name,room_type,default_capacity,COALESCE(description,'') description,status,calendar_revision,calendar_required,COALESCE(calendar_prepared_through::text,'') calendar_prepared_through,COALESCE(calendar_job_error,'') calendar_job_error,(cover_image IS NOT NULL) has_cover,cover_alt_text,cover_version,cover_is_demo FROM rooms`
const slotSelect = `SELECT id,name,to_char(start_time,'HH24:MI') start_time,to_char(end_time,'HH24:MI') end_time,is_active FROM time_slots`
const sessionSelect = `SELECT s.id,s.room_id,r.name room_name,s.time_slot_id,t.name slot_name,s.session_date::text,s.starts_at,s.ends_at,s.capacity,s.status,s.origin,COALESCE(s.closure_source,'') closure_source,COALESCE(s.change_reason,'') change_reason,(SELECT count(*) FROM bookings b WHERE b.room_session_id=s.id AND b.status IN ('PENDING_PT','CONFIRMED')) active_holds FROM room_sessions s JOIN rooms r ON r.id=s.room_id JOIN time_slots t ON t.id=s.time_slot_id`

func (r *PostgresFacilityRepository) Rooms(ctx context.Context, f catalog.Filter) ([]facility.Room, int, error) {
	return jsonList[facility.Room](ctx, r.pool, roomSelect+` WHERE (name ILIKE $1 OR room_type ILIKE $1) AND ($2='' OR status=$2)`, []any{"%" + f.Query + "%", f.Status}, f)
}
func (r *PostgresFacilityRepository) Room(ctx context.Context, id int64) (facility.Room, error) {
	return jsonOne[facility.Room](ctx, r.pool, roomSelect+` WHERE id=$1`, id)
}
func roomImpact(ctx context.Context, db queryRower, id int64) (facility.Impact, error) {
	return jsonOne[facility.Impact](ctx, db, `SELECT (SELECT count(*) FROM room_sessions WHERE room_id=r.id AND ends_at>now()) future_sessions,(SELECT count(*) FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id WHERE s.room_id=r.id AND s.ends_at>now() AND b.status IN ('PENDING_PT','CONFIRMED')) active_holds FROM rooms r WHERE r.id=$1`, id)
}
func (r *PostgresFacilityRepository) RoomImpact(ctx context.Context, id int64) (facility.Impact, error) {
	return roomImpact(ctx, r.pool, id)
}
func (r *PostgresFacilityRepository) SaveRoom(ctx context.Context, id int64, in facility.RoomInput, actorID int64, requestID string) (facility.Room, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (facility.Room, error) {
		var before any
		action := "room.created"
		if id == 0 {
			if err := tx.QueryRow(ctx, `INSERT INTO rooms(name,room_type,default_capacity,description,status) VALUES($1,$2,$3,NULLIF($4,''),$5) RETURNING id`, in.Name, in.RoomType, in.DefaultCapacity, in.Description, in.Status).Scan(&id); err != nil {
				return facility.Room{}, err
			}
		} else {
			if _, err := tx.Exec(ctx, `SELECT id FROM rooms WHERE id=$1 FOR UPDATE`, id); err != nil {
				return facility.Room{}, err
			}
			old, err := jsonOne[facility.Room](ctx, tx, roomSelect+` WHERE id=$1`, id)
			if err != nil {
				return old, err
			}
			before = old
			action = "room.updated"
			if in.Status != old.Status && in.Status != "AVAILABLE" {
				// Lock the affected session rows before inspecting holds. Booking writers must use the same rows.
				rows, err := tx.Query(ctx, `SELECT id FROM room_sessions WHERE room_id=$1 AND ends_at>now() ORDER BY id FOR UPDATE`, id)
				if err != nil {
					return old, err
				}
				for rows.Next() {
					var locked int64
					if err := rows.Scan(&locked); err != nil {
						rows.Close()
						return old, err
					}
				}
				err = rows.Err()
				rows.Close()
				if err != nil {
					return old, err
				}
				impact, err := roomImpact(ctx, tx, id)
				if err != nil {
					return old, err
				}
				if impact.ActiveHolds > 0 {
					return old, identity.Conflict("Phòng còn booking đang giữ chỗ. Hãy xử lý lịch liên quan trước khi đổi trạng thái.")
				}
				newStatus := "CLOSED"
				if in.Status == "MAINTENANCE" {
					newStatus = "MAINTENANCE"
				}
				if _, err = tx.Exec(ctx, `UPDATE room_sessions SET status=$2,closure_source='ROOM',change_reason=$3,updated_at=now() WHERE room_id=$1 AND starts_at>now() AND status='OPEN'`, id, newStatus, in.Reason); err != nil {
					return old, err
				}
			}
			if _, err := tx.Exec(ctx, `UPDATE rooms SET name=$2,room_type=$3,default_capacity=$4,description=NULLIF($5,''),status=$6,calendar_revision=calendar_revision+1,calendar_prepared_through=NULL,updated_at=now() WHERE id=$1`, id, in.Name, in.RoomType, in.DefaultCapacity, in.Description, in.Status); err != nil {
				return old, err
			}
		}
		after, err := jsonOne[facility.Room](ctx, tx, roomSelect+` WHERE id=$1`, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, actorID, action, "room", id, in.Reason, before, after, requestID); err != nil {
			return after, err
		}
		return after, nil
	})
}
func (r *PostgresFacilityRepository) Slots(ctx context.Context, f catalog.Filter) ([]facility.TimeSlot, int, error) {
	return jsonList[facility.TimeSlot](ctx, r.pool, slotSelect+` WHERE name ILIKE $1 AND ($2='' OR is_active=($2='ACTIVE'))`, []any{"%" + f.Query + "%", f.Status}, f)
}
func (r *PostgresFacilityRepository) CreateSlot(ctx context.Context, in facility.SlotInput, actorID int64, requestID string) (facility.TimeSlot, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (facility.TimeSlot, error) {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(841017)`); err != nil {
			return facility.TimeSlot{}, err
		}
		var id int64
		if err := tx.QueryRow(ctx, `INSERT INTO time_slots(name,start_time,end_time) VALUES($1,$2::time,$3::time) RETURNING id`, in.Name, in.StartTime, in.EndTime).Scan(&id); err != nil {
			return facility.TimeSlot{}, err
		}
		after, err := jsonOne[facility.TimeSlot](ctx, tx, slotSelect+` WHERE id=$1`, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, actorID, "time_slot.created", "time_slot", id, "", nil, after, requestID); err != nil {
			return after, err
		}
		return after, nil
	})
}
func (r *PostgresFacilityRepository) SlotStatus(ctx context.Context, id int64, in facility.SlotStatusInput, actorID int64, requestID string) (facility.TimeSlot, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (facility.TimeSlot, error) {
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(841017)`); err != nil {
			return facility.TimeSlot{}, err
		}
		var locked int64
		if err := tx.QueryRow(ctx, `SELECT id FROM time_slots WHERE id=$1 FOR UPDATE`, id).Scan(&locked); err != nil {
			return facility.TimeSlot{}, err
		}
		before, err := jsonOne[facility.TimeSlot](ctx, tx, slotSelect+` WHERE id=$1`, id)
		if err != nil {
			return before, err
		}
		if _, err = tx.Exec(ctx, `UPDATE time_slots SET is_active=$2 WHERE id=$1`, id, in.IsActive); err != nil {
			return before, err
		}
		after, err := jsonOne[facility.TimeSlot](ctx, tx, slotSelect+` WHERE id=$1`, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, actorID, "time_slot.status_changed", "time_slot", id, in.Reason, before, after, requestID); err != nil {
			return after, err
		}
		return after, nil
	})
}
func (r *PostgresFacilityRepository) Sessions(ctx context.Context, f catalog.Filter) ([]facility.Session, int, error) {
	return jsonList[facility.Session](ctx, r.pool, sessionSelect+` WHERE (r.name ILIKE $1 OR t.name ILIKE $1) AND ($2='' OR s.status=$2) AND ($3='' OR s.session_date=NULLIF($3,'')::date) AND ($4::bigint=0 OR s.room_id=$4)`, []any{"%" + f.Query + "%", f.Status, f.Date, f.RoomID}, f)
}
func (r *PostgresFacilityRepository) Session(ctx context.Context, id int64) (facility.Session, error) {
	return jsonOne[facility.Session](ctx, r.pool, sessionSelect+` WHERE s.id=$1`, id)
}
func (r *PostgresFacilityRepository) SaveSession(ctx context.Context, id int64, in facility.SessionInput, actorID int64, requestID string) (facility.Session, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (facility.Session, error) {
		var before any
		action := "room_session.created"
		if id == 0 {
			room, err := jsonOne[facility.Room](ctx, tx, roomSelect+` WHERE id=$1 FOR UPDATE`, in.RoomID)
			if err != nil {
				return facility.Session{}, err
			}
			if room.Status != "AVAILABLE" {
				return facility.Session{}, identity.Conflict("Phòng hoặc ca đang ngừng hoạt động, không thể tạo phiên.")
			}
			if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(841017)`); err != nil {
				return facility.Session{}, err
			}
			slot, err := jsonOne[facility.TimeSlot](ctx, tx, slotSelect+` WHERE id=$1 FOR SHARE`, in.TimeSlotID)
			if err != nil {
				return facility.Session{}, err
			}
			if !slot.IsActive {
				return facility.Session{}, identity.Conflict("Phòng hoặc ca đang ngừng hoạt động, không thể tạo phiên.")
			}
			if !facility.MinuteTime(in.SessionDate, slotMinutes(slot.StartTime)).After(time.Now()) {
				return facility.Session{}, identity.Conflict("Chỉ tạo phiên chưa bắt đầu.")
			}
			if err = requireCalendarSession(ctx, tx, in.RoomID, in.SessionDate, slotMinutes(slot.StartTime), slotMinutes(slot.EndTime)); err != nil {
				return facility.Session{}, err
			}
			capacity := in.Capacity
			if capacity == 0 {
				capacity = room.DefaultCapacity
			}
			if err := tx.QueryRow(ctx, `INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,created_by_user_id) VALUES($1,$2,$3::date,($3::date+$4::time) AT TIME ZONE 'Asia/Ho_Chi_Minh',($3::date+$5::time) AT TIME ZONE 'Asia/Ho_Chi_Minh',$6,$7) RETURNING id`, in.RoomID, in.TimeSlotID, in.SessionDate, slot.StartTime, slot.EndTime, capacity, actorID).Scan(&id); err != nil {
				return facility.Session{}, err
			}
		} else {
			// Lock room first, then session, matching SaveRoom to avoid lock-order cycles.
			var roomID int64
			if err := tx.QueryRow(ctx, `SELECT room_id FROM room_sessions WHERE id=$1`, id).Scan(&roomID); err != nil {
				return facility.Session{}, err
			}
			var roomStatus string
			if err := tx.QueryRow(ctx, `SELECT status FROM rooms WHERE id=$1 FOR UPDATE`, roomID).Scan(&roomStatus); err != nil {
				return facility.Session{}, err
			}
			var locked int64
			if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(841017)`); err != nil {
				return facility.Session{}, err
			}
			if err := tx.QueryRow(ctx, `SELECT id FROM room_sessions WHERE id=$1 FOR UPDATE`, id).Scan(&locked); err != nil {
				return facility.Session{}, err
			}
			old, err := jsonOne[facility.Session](ctx, tx, sessionSelect+` WHERE s.id=$1`, id)
			if err != nil {
				return old, err
			}
			before = old
			starts, err := time.Parse(time.RFC3339, old.StartsAt)
			if err != nil {
				return old, err
			}
			if !starts.After(time.Now()) {
				return old, identity.Conflict("Phiên đã bắt đầu được giữ nguyên snapshot.")
			}
			if in.Status == "OPEN" {
				start, end, err := sessionMinutes(old)
				if err != nil {
					return old, err
				}
				if err = requireCalendarSession(ctx, tx, roomID, old.SessionDate, start, end); err != nil {
					return old, err
				}
			}
			action = "room_session.updated"
			if in.Capacity < old.ActiveHolds || (in.Status != "OPEN" && old.ActiveHolds > 0) || (in.Status == "OPEN" && roomStatus != "AVAILABLE") {
				return old, identity.Conflict("Không thể giảm sức chứa dưới số chỗ giữ hoặc đóng phiên có booking. Phòng phải sẵn sàng để mở phiên.")
			}
			if _, err = tx.Exec(ctx, `UPDATE room_sessions SET capacity=$2,status=$3::varchar,closure_source=CASE WHEN $3::varchar='OPEN' THEN NULL ELSE 'MANUAL' END,change_reason=$4,updated_at=now() WHERE id=$1`, id, in.Capacity, in.Status, in.Reason); err != nil {
				return old, err
			}
		}
		after, err := jsonOne[facility.Session](ctx, tx, sessionSelect+` WHERE s.id=$1`, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, actorID, action, "room_session", id, in.Reason, before, after, requestID); err != nil {
			return after, err
		}
		return after, nil
	})
}
