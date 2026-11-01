package repository

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/trainer"
)

type TrainerRepository interface {
	List(context.Context, identity.User, catalog.Filter) ([]trainer.Profile, int, error)
	Get(context.Context, int64) (trainer.Profile, error)
	Own(context.Context, int64) (trainer.Profile, error)
	SaveProfile(context.Context, int64, trainer.ProfileInput, string) (trainer.Profile, error)
	Submit(context.Context, int64, string) (trainer.Profile, error)
	Decision(context.Context, int64, trainer.DecisionInput, int64, string) (trainer.Profile, error)
	Certificate(context.Context, int64) (string, error)
	SetCertificate(context.Context, int64, string, string) (trainer.Profile, error)
	Hours(context.Context, int64) ([]trainer.WorkingHour, error)
	SaveHours(context.Context, int64, []trainer.WorkingHour, int64, string) ([]trainer.WorkingHour, error)
	Absences(context.Context, int64, catalog.Filter) ([]trainer.Unavailability, int, error)
	SaveAbsence(context.Context, int64, trainer.Unavailability, int64, string) (trainer.Unavailability, error)
	DeleteAbsence(context.Context, int64, int64, int64, string) error
}
type PostgresTrainerRepository struct{ pool *pgxpool.Pool }

func NewTrainerRepository(pool *pgxpool.Pool) *PostgresTrainerRepository {
	return &PostgresTrainerRepository{pool}
}

const trainerSelect = `SELECT p.id,p.user_id,COALESCE(p.specialty,'') specialty,COALESCE(p.experience_text,'') experience_text,(p.certificate_path IS NOT NULL) has_certificate,p.approval_status,COALESCE(p.decision_reason,'') decision_reason,p.decided_at,
	(SELECT AVG(rv.rating)::float8 FROM pt_reviews rv JOIN bookings b ON b.id=rv.booking_id JOIN room_sessions rs ON rs.id=b.room_session_id WHERE b.pt_id=p.id AND ` + validReviewScope + `) rating_average,
	(SELECT COUNT(*)::int FROM pt_reviews rv JOIN bookings b ON b.id=rv.booking_id JOIN room_sessions rs ON rs.id=b.room_session_id WHERE b.pt_id=p.id AND ` + validReviewScope + `) review_count,
	jsonb_build_object('id',u.id,'username',u.username,'full_name',u.full_name,'role',r.code,'status',u.status) "user" FROM personal_trainers p JOIN users u ON u.id=p.user_id JOIN roles r ON r.id=u.role_id`
const hoursSelect = `SELECT id,day_of_week,to_char(start_time,'HH24:MI') start_time,to_char(end_time,'HH24:MI') end_time FROM pt_working_hours WHERE is_active AND pt_id=$1`
const absenceSelect = `SELECT id,starts_at,ends_at,COALESCE(reason,'') reason FROM pt_unavailability WHERE pt_id=$1`

func (r *PostgresTrainerRepository) List(ctx context.Context, a identity.User, f catalog.Filter) ([]trainer.Profile, int, error) {
	return jsonList[trainer.Profile](ctx, r.pool, trainerSelect+` WHERE (u.full_name ILIKE $1 OR p.specialty ILIKE $1) AND ($2='' OR p.approval_status=$2) AND ($3::boolean OR (p.approval_status='APPROVED' AND r.code='PT' AND u.status='ACTIVE') OR p.user_id=$4)
 AND ($5='' OR (p.approval_status='APPROVED' AND r.code='PT' AND u.status='ACTIVE'
 AND EXISTS(SELECT 1 FROM pt_working_hours h WHERE h.pt_id=p.id AND h.is_active AND h.day_of_week=EXTRACT(ISODOW FROM NULLIF($5,'')::timestamptz AT TIME ZONE 'Asia/Ho_Chi_Minh') AND h.start_time<=(NULLIF($5,'')::timestamptz AT TIME ZONE 'Asia/Ho_Chi_Minh')::time AND h.end_time>(NULLIF($5,'')::timestamptz AT TIME ZONE 'Asia/Ho_Chi_Minh')::time)
 AND NOT EXISTS(SELECT 1 FROM pt_unavailability n WHERE n.pt_id=p.id AND tstzrange(n.starts_at,n.ends_at,'[)') @> NULLIF($5,'')::timestamptz)
 AND NOT EXISTS(SELECT 1 FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id WHERE b.pt_id=p.id AND b.status IN ('PENDING_PT','CONFIRMED') AND tstzrange(s.starts_at,s.ends_at,'[)') @> NULLIF($5,'')::timestamptz)))`, []any{"%" + f.Query + "%", f.Status, a.Role == identity.RoleAdmin || a.Role == identity.RoleManager, a.ID, f.AvailableAt}, f)
}
func (r *PostgresTrainerRepository) Get(ctx context.Context, id int64) (trainer.Profile, error) {
	return jsonOne[trainer.Profile](ctx, r.pool, trainerSelect+` WHERE p.id=$1`, id)
}
func (r *PostgresTrainerRepository) Own(ctx context.Context, userID int64) (trainer.Profile, error) {
	return jsonOne[trainer.Profile](ctx, r.pool, trainerSelect+` WHERE p.user_id=$1`, userID)
}
func lockTrainer(ctx context.Context, tx pgx.Tx, id int64) (trainer.Profile, error) {
	var locked int64
	if err := tx.QueryRow(ctx, `SELECT id FROM personal_trainers WHERE id=$1 FOR UPDATE`, id).Scan(&locked); err != nil {
		return trainer.Profile{}, mapDBError(err)
	}
	return jsonOne[trainer.Profile](ctx, tx, trainerSelect+` WHERE p.id=$1`, id)
}
func (r *PostgresTrainerRepository) SaveProfile(ctx context.Context, userID int64, in trainer.ProfileInput, rid string) (trainer.Profile, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (trainer.Profile, error) {
		// Serialize create/update with role changes for this user.
		u, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id=$1 FOR UPDATE`, userID)
		if err != nil {
			return trainer.Profile{}, err
		}
		if u.Role != identity.RolePT {
			return trainer.Profile{}, identity.ErrForbidden
		}
		var id int64
		if err := tx.QueryRow(ctx, `INSERT INTO personal_trainers(user_id) VALUES($1) ON CONFLICT(user_id) DO UPDATE SET user_id=EXCLUDED.user_id RETURNING id`, userID).Scan(&id); err != nil {
			return trainer.Profile{}, err
		}
		before, err := lockTrainer(ctx, tx, id)
		if err != nil {
			return before, err
		}
		if before.ApprovalStatus != "DRAFT" && before.ApprovalStatus != "REJECTED" {
			return before, identity.Conflict("Trạng thái hồ sơ hoặc booking liên quan không cho phép thao tác này.")
		}
		if _, err = tx.Exec(ctx, `UPDATE personal_trainers SET specialty=$2,experience_text=$3,updated_at=now() WHERE id=$1`, id, in.Specialty, in.ExperienceText); err != nil {
			return before, err
		}
		after, err := jsonOne[trainer.Profile](ctx, tx, trainerSelect+` WHERE p.id=$1`, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, userID, "pt.profile_updated", "personal_trainer", id, "", before, after, rid); err != nil {
			return after, err
		}
		return after, nil
	})
}
func (r *PostgresTrainerRepository) Submit(ctx context.Context, userID int64, rid string) (trainer.Profile, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (trainer.Profile, error) {
		var id int64
		if err := tx.QueryRow(ctx, `SELECT id FROM personal_trainers WHERE user_id=$1`, userID).Scan(&id); err != nil {
			return trainer.Profile{}, err
		}
		before, err := lockTrainer(ctx, tx, id)
		if err != nil {
			return before, err
		}
		if !before.HasCertificate || before.Specialty == "" || before.ExperienceText == "" || (before.ApprovalStatus != "DRAFT" && before.ApprovalStatus != "REJECTED") {
			return before, identity.Conflict("Trạng thái hồ sơ hoặc booking liên quan không cho phép thao tác này.")
		}
		if _, err = tx.Exec(ctx, `UPDATE personal_trainers SET approval_status='PENDING',decision_reason=NULL,decided_by_user_id=NULL,decided_at=NULL,updated_at=now() WHERE id=$1`, id); err != nil {
			return before, err
		}
		after, err := jsonOne[trainer.Profile](ctx, tx, trainerSelect+` WHERE p.id=$1`, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, userID, "pt.submitted", "personal_trainer", id, "", before, after, rid); err != nil {
			return after, err
		}
		return after, nil
	})
}
func (r *PostgresTrainerRepository) Decision(ctx context.Context, id int64, in trainer.DecisionInput, actorID int64, rid string) (trainer.Profile, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (trainer.Profile, error) {
		var userID int64
		if err := tx.QueryRow(ctx, `SELECT user_id FROM personal_trainers WHERE id=$1`, id).Scan(&userID); err != nil {
			return trainer.Profile{}, err
		}
		user, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id=$1 FOR UPDATE`, userID)
		if err != nil {
			return trainer.Profile{}, err
		}
		before, err := lockTrainer(ctx, tx, id)
		if err != nil {
			return before, err
		}
		status := ""
		switch in.Action {
		case "APPROVE":
			if before.ApprovalStatus == "PENDING" && user.Role == identity.RolePT {
				status = "APPROVED"
			}
		case "REJECT":
			if before.ApprovalStatus == "PENDING" {
				status = "REJECTED"
			}
		case "SUSPEND":
			if before.ApprovalStatus == "APPROVED" {
				status = "SUSPENDED"
			}
		}
		if status == "" {
			return before, identity.Conflict("Trạng thái hồ sơ hoặc booking liên quan không cho phép thao tác này.")
		}
		if status == "SUSPENDED" {
			var holds int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id WHERE b.pt_id=$1 AND s.ends_at>now() AND b.status IN ('PENDING_PT','CONFIRMED')`, id).Scan(&holds); err != nil {
				return before, err
			}
			if holds > 0 {
				return before, identity.Conflict("Trạng thái hồ sơ hoặc booking liên quan không cho phép thao tác này.")
			}
		}
		if _, err = tx.Exec(ctx, `UPDATE personal_trainers SET approval_status=$2,decision_reason=$3,decided_by_user_id=$4,decided_at=now(),updated_at=now() WHERE id=$1`, id, status, in.Reason, actorID); err != nil {
			return before, err
		}
		after, err := jsonOne[trainer.Profile](ctx, tx, trainerSelect+` WHERE p.id=$1`, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, actorID, "pt.decision", "personal_trainer", id, in.Reason, before, after, rid); err != nil {
			return after, err
		}
		return after, nil
	})
}
func (r *PostgresTrainerRepository) Certificate(ctx context.Context, id int64) (string, error) {
	var path string
	err := r.pool.QueryRow(ctx, `SELECT certificate_path FROM personal_trainers WHERE id=$1 AND certificate_path IS NOT NULL`, id).Scan(&path)
	return path, mapDBError(err)
}
func (r *PostgresTrainerRepository) SetCertificate(ctx context.Context, userID int64, path, rid string) (trainer.Profile, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (trainer.Profile, error) {
		var id int64
		if err := tx.QueryRow(ctx, `SELECT id FROM personal_trainers WHERE user_id=$1`, userID).Scan(&id); err != nil {
			return trainer.Profile{}, err
		}
		before, err := lockTrainer(ctx, tx, id)
		if err != nil {
			return before, err
		}
		if before.ApprovalStatus != "DRAFT" && before.ApprovalStatus != "REJECTED" {
			return before, identity.Conflict("Trạng thái hồ sơ hoặc booking liên quan không cho phép thao tác này.")
		}
		if _, err = tx.Exec(ctx, `UPDATE personal_trainers SET certificate_path=$2,updated_at=now() WHERE id=$1`, id, path); err != nil {
			return before, err
		}
		after, err := jsonOne[trainer.Profile](ctx, tx, trainerSelect+` WHERE p.id=$1`, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, userID, "pt.certificate_uploaded", "personal_trainer", id, "", before, after, rid); err != nil {
			return after, err
		}
		return after, nil
	})
}
func (r *PostgresTrainerRepository) Hours(ctx context.Context, id int64) ([]trainer.WorkingHour, error) {
	data, _, err := jsonList[trainer.WorkingHour](ctx, r.pool, hoursSelect, []any{id}, catalog.Filter{Page: 1, PageSize: 100})
	return data, err
}
func (r *PostgresTrainerRepository) SaveHours(ctx context.Context, id int64, hours []trainer.WorkingHour, actorID int64, rid string) ([]trainer.WorkingHour, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) ([]trainer.WorkingHour, error) {
		if _, err := lockTrainer(ctx, tx, id); err != nil {
			return nil, err
		}
		// Read prior rows inside the transaction for an accurate audit snapshot.
		var oldJSON []byte
		if err := tx.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(h)),'[]'::jsonb) FROM (`+hoursSelect+`) h`, id).Scan(&oldJSON); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `UPDATE pt_working_hours SET is_active=false WHERE pt_id=$1`, id); err != nil {
			return nil, err
		}
		for _, h := range hours {
			if _, err := tx.Exec(ctx, `INSERT INTO pt_working_hours(pt_id,day_of_week,start_time,end_time) VALUES($1,$2,$3::time,$4::time) ON CONFLICT(pt_id,day_of_week,start_time,end_time) DO UPDATE SET is_active=true`, id, h.DayOfWeek, h.StartTime, h.EndTime); err != nil {
				return nil, err
			}
		}
		var invalid int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id WHERE b.pt_id=$1 AND s.ends_at>now() AND b.status IN ('PENDING_PT','CONFIRMED') AND NOT EXISTS(SELECT 1 FROM pt_working_hours h WHERE h.pt_id=$1 AND h.is_active AND h.day_of_week=EXTRACT(ISODOW FROM s.session_date) AND h.start_time<=(s.starts_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::time AND ((s.ends_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date=s.session_date AND h.end_time>=(s.ends_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::time OR (s.ends_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date=s.session_date+1 AND (s.ends_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::time='00:00'::time AND h.end_time='24:00'::time))`, id).Scan(&invalid); err != nil {
			return nil, err
		}
		if invalid > 0 {
			return nil, identity.Conflict("Lịch mới không bao phủ booking đang giữ chỗ. Hãy xử lý booking trước.")
		}
		if err := writeAudit(ctx, tx, actorID, "pt.availability_changed", "personal_trainer", id, "", json.RawMessage(oldJSON), hours, rid); err != nil {
			return nil, err
		}
		return hours, nil
	})
}
func (r *PostgresTrainerRepository) Absences(ctx context.Context, id int64, f catalog.Filter) ([]trainer.Unavailability, int, error) {
	return jsonList[trainer.Unavailability](ctx, r.pool, absenceSelect, []any{id}, f)
}
func (r *PostgresTrainerRepository) SaveAbsence(ctx context.Context, id int64, in trainer.Unavailability, actorID int64, rid string) (trainer.Unavailability, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (trainer.Unavailability, error) {
		if _, err := lockTrainer(ctx, tx, id); err != nil {
			return in, err
		}
		var holds int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id WHERE b.pt_id=$1 AND b.status IN ('PENDING_PT','CONFIRMED') AND tstzrange(s.starts_at,s.ends_at,'[)') && tstzrange($2::timestamptz,$3::timestamptz,'[)')`, id, in.StartsAt, in.EndsAt).Scan(&holds); err != nil {
			return in, err
		}
		if holds > 0 {
			return in, identity.Conflict("Khoảng nghỉ ảnh hưởng booking đang giữ chỗ. Hãy chọn thời gian khác.")
		}
		if err := tx.QueryRow(ctx, `INSERT INTO pt_unavailability(pt_id,starts_at,ends_at,reason) VALUES($1,$2::timestamptz,$3::timestamptz,$4) RETURNING id`, id, in.StartsAt, in.EndsAt, in.Reason).Scan(&in.ID); err != nil {
			return in, err
		}
		if err := writeAudit(ctx, tx, actorID, "pt.unavailability_created", "personal_trainer", id, in.Reason, nil, in, rid); err != nil {
			return in, err
		}
		return in, nil
	})
}
func (r *PostgresTrainerRepository) DeleteAbsence(ctx context.Context, ptID, id, actorID int64, rid string) error {
	_, err := inTransaction(ctx, r.pool, func(tx pgx.Tx) (bool, error) {
		if _, err := lockTrainer(ctx, tx, ptID); err != nil {
			return false, err
		}
		before, err := jsonOne[trainer.Unavailability](ctx, tx, absenceSelect+` AND id=$2 FOR UPDATE`, ptID, id)
		if err != nil {
			return false, err
		}
		if _, err = tx.Exec(ctx, `DELETE FROM pt_unavailability WHERE pt_id=$1 AND id=$2`, ptID, id); err != nil {
			return false, err
		}
		return true, writeAudit(ctx, tx, actorID, "pt.unavailability_removed", "personal_trainer", ptID, before.Reason, before, nil, rid)
	})
	return err
}
