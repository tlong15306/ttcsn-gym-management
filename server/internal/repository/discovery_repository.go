package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/discovery"
)

type DiscoveryRepository interface {
	Rooms(context.Context, catalog.Filter) ([]discovery.Room, int, error)
	Room(context.Context, int64) (discovery.Room, error)
	Trainers(context.Context, catalog.Filter) ([]discovery.Trainer, int, error)
	Sessions(context.Context, int64, string, string, catalog.Filter) ([]discovery.Session, int, error)
	Cover(context.Context, int64, bool) ([]byte, int64, error)
}
type PostgresDiscoveryRepository struct{ pool *pgxpool.Pool }

func NewDiscoveryRepository(pool *pgxpool.Pool) *PostgresDiscoveryRepository {
	return &PostgresDiscoveryRepository{pool}
}

const publicRoomSelect = `SELECT id,name,room_type,default_capacity,COALESCE(description,'') description,status,
 (cover_image IS NOT NULL) has_cover,cover_alt_text,cover_version,cover_is_demo FROM rooms`
const visibleRoom = `status IN ('AVAILABLE','MAINTENANCE')`

func (r *PostgresDiscoveryRepository) Rooms(ctx context.Context, f catalog.Filter) ([]discovery.Room, int, error) {
	return jsonList[discovery.Room](ctx, r.pool, publicRoomSelect+` WHERE `+visibleRoom+` AND (name ILIKE $1 OR room_type ILIKE $1) AND ($2='' OR room_type=$2)`, []any{"%" + f.Query + "%", f.Status}, f)
}
func (r *PostgresDiscoveryRepository) Room(ctx context.Context, id int64) (discovery.Room, error) {
	return jsonOne[discovery.Room](ctx, r.pool, publicRoomSelect+` WHERE id=$1 AND `+visibleRoom, id)
}
func (r *PostgresDiscoveryRepository) Trainers(ctx context.Context, f catalog.Filter) ([]discovery.Trainer, int, error) {
	return jsonList[discovery.Trainer](ctx, r.pool, `SELECT p.id,u.full_name,COALESCE(p.specialty,'') specialty,COALESCE(p.experience_text,'') experience_text,
 (SELECT AVG(rv.rating)::float8 FROM pt_reviews rv JOIN bookings b ON b.id=rv.booking_id JOIN room_sessions rs ON rs.id=b.room_session_id WHERE b.pt_id=p.id AND `+validReviewScope+`) rating_average,
 (SELECT COUNT(*)::int FROM pt_reviews rv JOIN bookings b ON b.id=rv.booking_id JOIN room_sessions rs ON rs.id=b.room_session_id WHERE b.pt_id=p.id AND `+validReviewScope+`) review_count
 FROM personal_trainers p JOIN users u ON u.id=p.user_id JOIN roles role ON role.id=u.role_id
 WHERE p.approval_status='APPROVED' AND u.status='ACTIVE' AND role.code='PT' AND (u.full_name ILIKE $1 OR p.specialty ILIKE $1)`, []any{"%" + f.Query + "%"}, f)
}
func (r *PostgresDiscoveryRepository) Sessions(ctx context.Context, id int64, from, to string, f catalog.Filter) ([]discovery.Session, int, error) {
	if _, err := r.Room(ctx, id); err != nil {
		return nil, 0, err
	}
	return jsonList[discovery.Session](ctx, r.pool, `SELECT s.id,s.room_id,s.session_date::text,s.starts_at,s.ends_at,s.capacity,
 CASE WHEN r.status='AVAILABLE' AND s.status='OPEN' AND s.starts_at>clock_timestamp() THEN GREATEST(s.capacity-(SELECT COUNT(*) FROM bookings b WHERE b.room_session_id=s.id AND b.status IN ('PENDING_PT','CONFIRMED')),0)::int ELSE 0 END remaining,
 CASE WHEN r.status='MAINTENANCE' THEN 'MAINTENANCE' ELSE s.status END status
 FROM room_sessions s JOIN rooms r ON r.id=s.room_id WHERE s.room_id=$1 AND r.status IN ('AVAILABLE','MAINTENANCE')
 AND s.session_date BETWEEN $2::date AND $3::date AND s.ends_at>clock_timestamp() AND (NOT r.calendar_required OR s.origin='CALENDAR')`, []any{id, from, to}, f)
}
func (r *PostgresDiscoveryRepository) Cover(ctx context.Context, id int64, card bool) ([]byte, int64, error) {
	column := "cover_image"
	if card {
		column = "cover_thumbnail"
	}
	var data []byte
	var version int64
	err := r.pool.QueryRow(ctx, `SELECT `+column+`,cover_version FROM rooms WHERE id=$1 AND `+visibleRoom+` AND cover_image IS NOT NULL`, id).Scan(&data, &version)
	return data, version, mapDBError(err)
}
