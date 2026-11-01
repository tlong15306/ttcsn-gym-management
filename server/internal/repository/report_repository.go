package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/report"
)

type PostgresReportRepository struct{ pool *pgxpool.Pool }

func NewReportRepository(pool *pgxpool.Pool) *PostgresReportRepository {
	return &PostgresReportRepository{pool: pool}
}

func reportArgs(period report.Range) (string, string) {
	return period.From.Format("2006-01-02"), period.To.Format("2006-01-02")
}

func (r *PostgresReportRepository) Operations(ctx context.Context, period report.Range, roomID int64) (report.Operations, error) {
	if roomID > 0 {
		var exists bool
		if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM rooms WHERE id=$1)`, roomID).Scan(&exists); err != nil {
			return report.Operations{}, err
		}
		if !exists {
			return report.Operations{}, identity.ErrNotFound
		}
	}
	from, to := reportArgs(period)
	return jsonOne[report.Operations](ctx, r.pool, `WITH bounds AS (
		SELECT $1::date from_date,$2::date to_date,$3::bigint room_id,clock_timestamp() as_of
	), period_sessions AS (
		SELECT s.* FROM room_sessions s,bounds x
		WHERE s.session_date BETWEEN x.from_date AND x.to_date AND (x.room_id=0 OR s.room_id=x.room_id)
	), period_bookings AS (
		SELECT b.id,b.status,b.confirmed_at,b.cancellation_source,s.ends_at,s.id room_session_id
		FROM bookings b JOIN period_sessions s ON s.id=b.room_session_id
	), status_counts AS (
		SELECT status,count(*)::bigint amount FROM period_bookings GROUP BY status
	), cancellation_counts AS (
		SELECT COALESCE(cancellation_source,'UNKNOWN') source,count(*)::bigint amount
		FROM period_bookings WHERE status='CANCELLED' GROUP BY COALESCE(cancellation_source,'UNKNOWN')
	), ended AS (
		SELECT p.* FROM period_bookings p,bounds x WHERE p.ends_at<=x.as_of
	), utilization AS (
		SELECT
			COALESCE((SELECT sum((s.capacity::numeric * EXTRACT(EPOCH FROM (s.ends_at-s.starts_at))/60))::bigint
				FROM period_sessions s,bounds x WHERE s.status='OPEN' AND s.ends_at<=x.as_of),0)::bigint capacity_minutes,
			COALESCE((SELECT sum(EXTRACT(EPOCH FROM (s.ends_at-s.starts_at))/60)::bigint
				FROM period_sessions s JOIN bookings b ON b.room_session_id=s.id,bounds x
				WHERE s.status='OPEN' AND s.ends_at<=x.as_of AND b.status IN ('CONFIRMED','COMPLETED','NO_SHOW')),0)::bigint booked_minutes
	), attendance_counts AS (
		SELECT
			count(DISTINCT a.booking_id) FILTER (WHERE a.participant_type='MEMBER')::bigint member_count,
			count(DISTINCT a.booking_id) FILTER (WHERE a.participant_type='PT')::bigint pt_count
		FROM booking_attendance a JOIN ended e ON e.id=a.booking_id
	), no_show AS (
		SELECT count(*) FILTER (WHERE status='NO_SHOW')::bigint numerator,
			count(*) FILTER (WHERE status IN ('COMPLETED','NO_SHOW'))::bigint denominator
		FROM ended
	), quality AS (
		SELECT count(*) FILTER (WHERE
			(status IN ('COMPLETED','NO_SHOW') AND confirmed_at IS NULL)
			OR (status='CANCELLED' AND (cancellation_source IS NULL OR cancellation_source='UNKNOWN'
				OR (cancellation_source IN ('MEMBER','PT') AND confirmed_at IS NULL)))
		)::bigint unclassified_count
		FROM ended
	)
	SELECT to_char(x.from_date,'YYYY-MM-DD') AS "from",to_char(x.to_date,'YYYY-MM-DD') AS "to",
		'Asia/Ho_Chi_Minh' timezone,x.as_of,
		(SELECT count(*)::bigint FROM period_bookings) total_bookings,
		COALESCE((SELECT jsonb_object_agg(status,amount) FROM status_counts),'{}'::jsonb) status_counts,
		a.member_count member_attendance_count,a.pt_count pt_attendance_count,
		jsonb_build_object('numerator',n.numerator,'denominator',n.denominator,'rate_percent',
			CASE WHEN n.denominator=0 THEN NULL ELSE ROUND(100.0*n.numerator/n.denominator,2)::float8 END) no_show,
		(SELECT count(*)::bigint FROM ended WHERE status='CONFIRMED') pending_completion_count,
		jsonb_build_object('booked',u.booked_minutes,'capacity',u.capacity_minutes,'rate_percent',
			CASE WHEN u.capacity_minutes=0 THEN NULL ELSE ROUND(100.0*u.booked_minutes/u.capacity_minutes,2)::float8 END) utilization,
		COALESCE((SELECT jsonb_object_agg(source,amount) FROM cancellation_counts),'{}'::jsonb) cancellation_source_counts,
		q.unclassified_count unclassified_booking_count
	FROM bounds x CROSS JOIN attendance_counts a CROSS JOIN no_show n CROSS JOIN utilization u CROSS JOIN quality q`, from, to, roomID)
}

func (r *PostgresReportRepository) Trainer(ctx context.Context, actor identity.User, ptID int64, period report.Range) (report.Trainer, error) {
	var exists bool
	if actor.Role == identity.RolePT {
		if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM personal_trainers WHERE id=$1 AND user_id=$2)`, ptID, actor.ID).Scan(&exists); err != nil {
			return report.Trainer{}, err
		}
	} else {
		if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM personal_trainers WHERE id=$1)`, ptID).Scan(&exists); err != nil {
			return report.Trainer{}, err
		}
	}
	if !exists {
		return report.Trainer{}, identity.ErrNotFound
	}
	from, to := reportArgs(period)
	return jsonOne[report.Trainer](ctx, r.pool, `WITH bounds AS (
		SELECT $1::date from_date,$2::date to_date,$3::bigint pt_id,clock_timestamp() as_of
	), ended AS (
		SELECT b.id,b.status,b.confirmed_at,b.cancellation_source,s.ends_at
		FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id,bounds x
		WHERE b.pt_id=x.pt_id AND s.session_date BETWEEN x.from_date AND x.to_date AND s.ends_at<=x.as_of
	), classified AS (
		SELECT e.*,
			((e.status IN ('COMPLETED','NO_SHOW') AND e.confirmed_at IS NOT NULL)
				OR (e.status='CANCELLED' AND e.confirmed_at IS NOT NULL AND e.cancellation_source IN ('MEMBER','PT'))) eligible,
			((e.status IN ('COMPLETED','NO_SHOW') AND e.confirmed_at IS NULL)
				OR (e.status='CANCELLED' AND (e.cancellation_source IS NULL OR e.cancellation_source='UNKNOWN'
					OR (e.cancellation_source IN ('MEMBER','PT') AND e.confirmed_at IS NULL)))) unclassified
		FROM ended e
	), attendance AS (
		SELECT count(*) FILTER (WHERE c.eligible)::bigint eligible_count,
			count(*) FILTER (WHERE c.eligible AND EXISTS(SELECT 1 FROM booking_attendance a WHERE a.booking_id=c.id AND a.participant_type='PT'))::bigint present_count,
			count(*) FILTER (WHERE c.unclassified)::bigint unclassified_count,
			count(*) FILTER (WHERE c.status='CONFIRMED')::bigint pending_completion_count,
			count(*) FILTER (WHERE NOT c.eligible AND NOT c.unclassified AND c.status<>'CONFIRMED')::bigint excluded_count
		FROM classified c
	), ratings AS (
		SELECT avg(rv.rating)::float8 rating_average,count(rv.id)::bigint review_count
		FROM pt_reviews rv JOIN bookings b ON b.id=rv.booking_id JOIN room_sessions rs ON rs.id=b.room_session_id,bounds x
		WHERE b.pt_id=x.pt_id AND rs.session_date BETWEEN x.from_date AND x.to_date AND `+validReviewProof+` AND rs.ends_at<=x.as_of
	)
	SELECT to_char(x.from_date,'YYYY-MM-DD') AS "from",to_char(x.to_date,'YYYY-MM-DD') AS "to",
		'Asia/Ho_Chi_Minh' timezone,x.as_of,x.pt_id,
		r.rating_average,r.review_count,
		jsonb_build_object('numerator',a.present_count,'denominator',a.eligible_count,'rate_percent',
			CASE WHEN a.eligible_count=0 THEN NULL ELSE ROUND(100.0*a.present_count/a.eligible_count,2)::float8 END) attendance,
		a.excluded_count,a.unclassified_count,a.pending_completion_count
	FROM bounds x CROSS JOIN ratings r CROSS JOIN attendance a`, from, to, ptID)
}
