package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/review"
)

type ReviewRepository interface {
	Create(context.Context, identity.User, int64, review.Input, string) (review.Review, error)
	Get(context.Context, identity.User, int64) (*review.Review, error)
	ListPT(context.Context, identity.User, int64, catalog.Filter) ([]review.Review, int, error)
}

type PostgresReviewRepository struct{ pool *pgxpool.Pool }

func NewReviewRepository(pool *pgxpool.Pool) *PostgresReviewRepository {
	return &PostgresReviewRepository{pool: pool}
}

const reviewSelect = `SELECT rating,COALESCE(comment,'') comment,created_at,'Hội viên' reviewer_label FROM pt_reviews`

// Keep public lists, profile summaries, and reports on one definition of a valid review.
const validReviewProof = `b.status='COMPLETED' AND b.pt_id IS NOT NULL
	AND EXISTS(SELECT 1 FROM booking_attendance a WHERE a.booking_id=b.id AND a.participant_type='MEMBER')
	AND EXISTS(SELECT 1 FROM booking_attendance a WHERE a.booking_id=b.id AND a.participant_type='PT')`

const validReviewScope = validReviewProof + ` AND rs.ends_at<=clock_timestamp()`

func (r *PostgresReviewRepository) Create(ctx context.Context, actor identity.User, bookingID int64, in review.Input, requestID string) (review.Review, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (review.Review, error) {
		var result review.Review
		before, err := jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1 AND m.user_id=$2`, bookingID, actor.ID)
		if err != nil {
			return result, err
		}
		if err = bookingLocks(ctx, tx, before.RoomID, before.MemberID, before.PTID, before.RoomSessionID); err != nil {
			return result, mapDBError(err)
		}
		var locked int64
		if err = tx.QueryRow(ctx, `SELECT id FROM bookings WHERE id=$1 FOR UPDATE`, bookingID).Scan(&locked); err != nil {
			return result, mapDBError(err)
		}
		before, err = jsonOne[booking.Booking](ctx, tx, bookingSelect+` WHERE b.id=$1 AND m.user_id=$2`, bookingID, actor.ID)
		if err != nil {
			return result, err
		}
		var exists bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pt_reviews WHERE booking_id=$1)`, bookingID).Scan(&exists); err != nil {
			return result, err
		}
		if exists {
			return result, review.Fail("REVIEW_EXISTS", "Buổi tập này đã có đánh giá.")
		}
		var eligible bool
		if err = tx.QueryRow(ctx, `SELECT b.status='COMPLETED' AND s.ends_at<=clock_timestamp() AND b.pt_id IS NOT NULL
			AND EXISTS(SELECT 1 FROM booking_attendance a WHERE a.booking_id=b.id AND a.participant_type='MEMBER')
			AND EXISTS(SELECT 1 FROM booking_attendance a WHERE a.booking_id=b.id AND a.participant_type='PT')
			FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id WHERE b.id=$1 AND b.member_id=$2`, bookingID, before.MemberID).Scan(&eligible); err != nil {
			return result, err
		}
		if !eligible {
			return result, review.Fail("REVIEW_NOT_ELIGIBLE", "Chỉ đánh giá được buổi đã hoàn tất, có PT và cả hai người đã điểm danh.")
		}
		if err = tx.QueryRow(ctx, `INSERT INTO pt_reviews(booking_id,rating,comment) VALUES($1,$2,NULLIF($3,'')) RETURNING rating,COALESCE(comment,'') comment,created_at,'Hội viên' reviewer_label`, bookingID, in.Rating, in.Comment).Scan(&result.Rating, &result.Comment, &result.CreatedAt, &result.ReviewerLabel); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "pt_reviews_booking_id_key" {
				return result, review.Fail("REVIEW_EXISTS", "Buổi tập này đã có đánh giá.")
			}
			return result, err
		}
		if err = writeAudit(ctx, tx, actor.ID, "review.created", "booking", bookingID, "", nil, result, requestID); err != nil {
			return result, err
		}
		return result, nil
	})
}

func (r *PostgresReviewRepository) Get(ctx context.Context, actor identity.User, bookingID int64) (*review.Review, error) {
	if _, err := jsonOne[booking.Booking](ctx, r.pool, bookingSelect+` WHERE b.id=$1 AND `+bookingScope(actor), bookingID); err != nil {
		return nil, err
	}
	value, err := jsonOne[review.Review](ctx, r.pool, reviewSelect+` WHERE booking_id=$1`, bookingID)
	if errors.Is(err, identity.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (r *PostgresReviewRepository) ListPT(ctx context.Context, actor identity.User, ptID int64, f catalog.Filter) ([]review.Review, int, error) {
	var allowed bool
	query := `SELECT EXISTS(SELECT 1 FROM personal_trainers p JOIN users u ON u.id=p.user_id JOIN roles rr ON rr.id=u.role_id WHERE p.id=$1 AND (
		($2='PT' AND p.user_id=$3) OR
		($2='MEMBER' AND rr.code='PT' AND u.status='ACTIVE' AND p.approval_status='APPROVED') OR
		$2 IN ('ADMIN','MANAGER','RECEPTIONIST')
	))`
	if err := r.pool.QueryRow(ctx, query, ptID, actor.Role, actor.ID).Scan(&allowed); err != nil {
		return nil, 0, err
	}
	if !allowed {
		return nil, 0, identity.ErrNotFound
	}
	base := `SELECT rv.rating,COALESCE(rv.comment,'') comment,rv.created_at,'Hội viên' reviewer_label FROM pt_reviews rv JOIN bookings b ON b.id=rv.booking_id JOIN room_sessions rs ON rs.id=b.room_session_id WHERE b.pt_id=$1 AND ` + validReviewScope
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM (`+base+`) items`, ptID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, base+` ORDER BY rv.created_at DESC,rv.id DESC LIMIT $2 OFFSET $3`, ptID, f.PageSize, (f.Page-1)*f.PageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]review.Review, 0)
	for rows.Next() {
		var item review.Review
		if err = rows.Scan(&item.Rating, &item.Comment, &item.CreatedAt, &item.ReviewerLabel); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}
