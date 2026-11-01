package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
)

const userColumns = `u.id, u.username, u.password_hash, u.email, COALESCE(u.phone, ''), u.full_name, r.code, u.status, u.auth_version, u.created_at, u.updated_at`
const userFrom = ` FROM users u JOIN roles r ON r.id = u.role_id`

type PostgresIdentityRepository struct {
	pool *pgxpool.Pool
}

func NewIdentityRepository(pool *pgxpool.Pool) *PostgresIdentityRepository {
	return &PostgresIdentityRepository{pool: pool}
}

func (r *PostgresIdentityRepository) FindByUsername(ctx context.Context, username string) (identity.User, error) {
	return findUser(ctx, r.pool, `SELECT `+userColumns+userFrom+` WHERE LOWER(u.username) = LOWER($1)`, username)
}

func (r *PostgresIdentityRepository) FindByID(ctx context.Context, id int64) (identity.User, error) {
	return findUser(ctx, r.pool, `SELECT `+userColumns+userFrom+` WHERE u.id = $1`, id)
}

func (r *PostgresIdentityRepository) List(ctx context.Context, filter identity.ListFilter) ([]identity.User, int, error) {
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 5)
	if filter.Query != "" {
		args = append(args, "%"+filter.Query+"%")
		placeholder := fmt.Sprintf("$%d", len(args))
		clauses = append(clauses, "(u.username ILIKE "+placeholder+" OR u.email ILIKE "+placeholder+" OR u.full_name ILIKE "+placeholder+")")
	}
	if filter.Role != "" {
		args = append(args, string(filter.Role))
		clauses = append(clauses, fmt.Sprintf("r.code = $%d", len(args)))
	}
	if filter.Status != "" {
		args = append(args, string(filter.Status))
		clauses = append(clauses, fmt.Sprintf("u.status = $%d", len(args)))
	}

	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*)`+userFrom+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	limitArg, offsetArg := len(args)-1, len(args)
	rows, err := r.pool.Query(ctx, `SELECT `+userColumns+userFrom+where+fmt.Sprintf(" ORDER BY u.id ASC LIMIT $%d OFFSET $%d", limitArg, offsetArg), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]identity.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users: %w", err)
	}
	return users, total, nil
}

func (r *PostgresIdentityRepository) Create(ctx context.Context, input identity.CreateInput, passwordHash string, actorID int64, requestID string) (identity.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return identity.User{}, fmt.Errorf("begin create user: %w", err)
	}
	defer tx.Rollback(ctx)

	var roleID int16
	if err := tx.QueryRow(ctx, `SELECT id FROM roles WHERE code = $1`, input.Role).Scan(&roleID); err != nil {
		return identity.User{}, mapDBError(err)
	}
	var id int64
	err = tx.QueryRow(ctx, `INSERT INTO users (username, password_hash, email, phone, full_name, role_id) VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6) RETURNING id`, input.Username, passwordHash, input.Email, input.Phone, input.FullName, roleID).Scan(&id)
	if err != nil {
		return identity.User{}, mapDBError(err)
	}
	user, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id = $1`, id)
	if err != nil {
		return identity.User{}, err
	}
	if err := writeAudit(ctx, tx, actorID, "account.created", "user", id, "", nil, user, requestID); err != nil {
		return identity.User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return identity.User{}, fmt.Errorf("commit create user: %w", err)
	}
	return user, nil
}

func (r *PostgresIdentityRepository) UpdateProfile(ctx context.Context, id int64, input identity.ProfileInput) (identity.User, error) {
	return r.UpdateAccount(ctx, id, input, id, "")
}

func (r *PostgresIdentityRepository) UpdateAccount(ctx context.Context, id int64, input identity.ProfileInput, actorID int64, requestID string) (identity.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return identity.User{}, err
	}
	defer tx.Rollback(ctx)
	before, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id = $1 FOR UPDATE`, id)
	if err != nil {
		return identity.User{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE users SET full_name=$2, email=$3, phone=NULLIF($4,''), updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id, input.FullName, input.Email, input.Phone); err != nil {
		return identity.User{}, mapDBError(err)
	}
	after, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id=$1`, id)
	if err != nil {
		return identity.User{}, err
	}
	if err = writeAudit(ctx, tx, actorID, "account.profile_updated", "user", id, "", before, after, requestID); err != nil {
		return identity.User{}, err
	}
	return after, tx.Commit(ctx)
}

func (r *PostgresIdentityRepository) SetStatus(ctx context.Context, targetID int64, status identity.Status, reason string, actorID int64, requestID string) (identity.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return identity.User{}, fmt.Errorf("begin set status: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(7042026)`); err != nil {
		return identity.User{}, fmt.Errorf("lock admin policy: %w", err)
	}

	before, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id = $1 FOR UPDATE`, targetID)
	if err != nil {
		return identity.User{}, err
	}
	if before.Role == identity.RoleAdmin && before.Status == identity.StatusActive && status != identity.StatusActive {
		if err := ensureAnotherActiveAdmin(ctx, tx, before.ID); err != nil {
			return identity.User{}, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET status = $2, auth_version = auth_version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, targetID, status); err != nil {
		return identity.User{}, mapDBError(err)
	}
	after, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id = $1`, targetID)
	if err != nil {
		return identity.User{}, err
	}
	if err := writeAudit(ctx, tx, actorID, "account.status_changed", "user", targetID, reason, before, after, requestID); err != nil {
		return identity.User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return identity.User{}, fmt.Errorf("commit set status: %w", err)
	}
	return after, nil
}

func (r *PostgresIdentityRepository) SetRole(ctx context.Context, targetID int64, role identity.Role, reason string, actorID int64, requestID string) (identity.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return identity.User{}, fmt.Errorf("begin set role: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(7042026)`); err != nil {
		return identity.User{}, fmt.Errorf("lock admin policy: %w", err)
	}

	before, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id = $1 FOR UPDATE`, targetID)
	if err != nil {
		return identity.User{}, err
	}
	if before.Role == identity.RoleAdmin && before.Status == identity.StatusActive && role != identity.RoleAdmin {
		if err := ensureAnotherActiveAdmin(ctx, tx, before.ID); err != nil {
			return identity.User{}, err
		}
	}
	if role == identity.RoleMember || role == identity.RolePT {
		hasProfile, err := hasRequiredProfile(ctx, tx, targetID, role)
		if err != nil {
			return identity.User{}, err
		}
		if !hasProfile {
			return identity.User{}, identity.ErrProfileRequirement
		}
	}
	var roleID int16
	if err := tx.QueryRow(ctx, `SELECT id FROM roles WHERE code = $1`, role).Scan(&roleID); err != nil {
		return identity.User{}, mapDBError(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET role_id = $2, auth_version = auth_version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, targetID, roleID); err != nil {
		return identity.User{}, mapDBError(err)
	}
	after, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id = $1`, targetID)
	if err != nil {
		return identity.User{}, err
	}
	if err := writeAudit(ctx, tx, actorID, "account.role_changed", "user", targetID, reason, before, after, requestID); err != nil {
		return identity.User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return identity.User{}, fmt.Errorf("commit set role: %w", err)
	}
	return after, nil
}

func (r *PostgresIdentityRepository) ResetPassword(ctx context.Context, targetID int64, passwordHash string, actorID int64, requestID string) (identity.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return identity.User{}, fmt.Errorf("begin password reset: %w", err)
	}
	defer tx.Rollback(ctx)
	before, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id = $1 FOR UPDATE`, targetID)
	if err != nil {
		return identity.User{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET password_hash = $2, auth_version = auth_version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, targetID, passwordHash); err != nil {
		return identity.User{}, mapDBError(err)
	}
	after, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id = $1`, targetID)
	if err != nil {
		return identity.User{}, err
	}
	if err := writeAudit(ctx, tx, actorID, "account.password_reset", "user", targetID, "", before, after, requestID); err != nil {
		return identity.User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return identity.User{}, fmt.Errorf("commit password reset: %w", err)
	}
	return after, nil
}

func (r *PostgresIdentityRepository) BumpAuthVersion(ctx context.Context, targetID, actorID int64, requestID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin logout: %w", err)
	}
	defer tx.Rollback(ctx)
	before, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id = $1 FOR UPDATE`, targetID)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET auth_version = auth_version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, targetID); err != nil {
		return mapDBError(err)
	}
	after, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id = $1`, targetID)
	if err != nil {
		return err
	}
	if err := writeAudit(ctx, tx, actorID, "auth.logout_all", "user", targetID, "", before, after, requestID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type queryRower interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func findUser(ctx context.Context, db queryRower, query string, args ...any) (identity.User, error) {
	user, err := scanUser(db.QueryRow(ctx, query, args...))
	if err != nil {
		return identity.User{}, mapDBError(err)
	}
	return user, nil
}

func scanUser(row pgx.Row) (identity.User, error) {
	var user identity.User
	var role, status string
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.Phone, &user.FullName, &role, &status, &user.AuthVersion, &user.CreatedAt, &user.UpdatedAt)
	user.Role = identity.Role(role)
	user.Status = identity.Status(status)
	return user, err
}

func ensureAnotherActiveAdmin(ctx context.Context, tx pgx.Tx, excludedID int64) error {
	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM users u JOIN roles r ON r.id = u.role_id WHERE r.code = 'ADMIN' AND u.status = 'ACTIVE' AND u.id <> $1`, excludedID).Scan(&count); err != nil {
		return fmt.Errorf("count active admins: %w", err)
	}
	if count == 0 {
		return identity.ErrLastActiveAdmin
	}
	return nil
}

func hasRequiredProfile(ctx context.Context, tx pgx.Tx, userID int64, role identity.Role) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM members WHERE user_id = $1)`
	if role == identity.RolePT {
		query = `SELECT EXISTS (SELECT 1 FROM personal_trainers WHERE user_id = $1)`
	}
	if err := tx.QueryRow(ctx, query, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check role profile: %w", err)
	}
	return exists, nil
}

func writeAudit(ctx context.Context, tx pgx.Tx, actorID int64, action, entityType string, entityID int64, reason string, before, after any, requestID string) error {
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		return fmt.Errorf("marshal audit before data: %w", err)
	}
	afterJSON, err := json.Marshal(after)
	if err != nil {
		return fmt.Errorf("marshal audit after data: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO audit_events (actor_user_id, action, entity_type, entity_id, reason, before_data, after_data, request_id) VALUES (NULLIF($1,0), $2, $3, $4, NULLIF($5, ''), $6, $7, $8)`, actorID, action, entityType, fmt.Sprint(entityID), reason, beforeJSON, afterJSON, requestID)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

func mapDBError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "uq_users_username_ci":
			return identity.Conflict("Tên đăng nhập đã được sử dụng.")
		case "uq_users_email_ci":
			return identity.Conflict("Email đã được sử dụng.")
		case "members_member_code_key":
			return identity.Conflict("Mã hội viên đã tồn tại.")
		case "members_user_id_key":
			return identity.Conflict("Tài khoản này đã có hồ sơ hội viên.")
		case "uq_rooms_name_ci":
			return identity.Conflict("Tên phòng đã tồn tại.")
		}
		return identity.ErrConflict
	}
	if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
		return identity.Conflict("Khoảng thời gian bị chồng với lịch đang hoạt động. Hãy chọn giờ khác.")
	}
	if errors.As(err, &pgErr) && (pgErr.Code == "23514" || pgErr.Code == "22001" || pgErr.Code == "22007" || pgErr.Code == "22008") {
		return identity.ErrValidation
	}
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return identity.ErrConflict
	}
	return err
}
