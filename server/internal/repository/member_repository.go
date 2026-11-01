package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/member"
)

type MemberRepository interface {
	EligibleAccounts(context.Context, catalog.Filter) ([]identity.User, int, error)
	List(context.Context, identity.User, catalog.Filter) ([]member.Member, int, error)
	Get(context.Context, identity.User, int64) (member.Member, error)
	Create(context.Context, member.Input, string, int64, string) (member.Member, error)
	Update(context.Context, int64, member.Input, int64, string) (member.Member, error)
	Membership(context.Context, int64, member.MembershipInput, int64, string) (member.Member, error)
}

func (r *PostgresMemberRepository) EligibleAccounts(ctx context.Context, f catalog.Filter) ([]identity.User, int, error) {
	return jsonList[identity.User](ctx, r.pool, `SELECT u.id,u.username,u.full_name,u.email,COALESCE(u.phone,'') phone,r.code role,u.status`+userFrom+` WHERE r.code='MEMBER' AND NOT EXISTS(SELECT 1 FROM members m WHERE m.user_id=u.id) AND (u.username ILIKE $1 OR u.full_name ILIKE $1 OR u.email ILIKE $1)`, []any{"%" + f.Query + "%"}, f)
}

type PostgresMemberRepository struct{ pool *pgxpool.Pool }

func NewMemberRepository(pool *pgxpool.Pool) *PostgresMemberRepository {
	return &PostgresMemberRepository{pool}
}

const memberSelect = `SELECT m.id,m.user_id,m.member_code,COALESCE(m.date_of_birth::text,'') date_of_birth,COALESCE(m.gender,'') gender,COALESCE(m.address,'') address,m.join_date::text,m.membership_expiry::text,m.status,(m.membership_expiry < (now() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date) expired,jsonb_build_object('id',u.id,'username',u.username,'email',u.email,'phone',COALESCE(u.phone,''),'full_name',u.full_name,'role',r.code,'status',u.status) "user" FROM members m JOIN users u ON u.id=m.user_id JOIN roles r ON r.id=u.role_id`

func memberScope(actor identity.User) string {
	if actor.Role == identity.RoleMember {
		return ` AND m.user_id=$3`
	}
	if actor.Role == identity.RolePT {
		return ` AND EXISTS(SELECT 1 FROM bookings b JOIN personal_trainers p ON p.id=b.pt_id WHERE b.member_id=m.id AND p.user_id=$3 AND b.status IN ('PENDING_PT','CONFIRMED') AND b.session_date >= (now() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date)`
	}
	return ` AND $3::bigint>0`
}
func memberPublic(m member.Member, actor identity.User) member.Member {
	if actor.Role == identity.RolePT {
		m.User.Email = ""
		m.User.Phone = ""
		m.Address = ""
		m.DateOfBirth = ""
		m.Gender = ""
	}
	return m
}
func (r *PostgresMemberRepository) List(ctx context.Context, actor identity.User, f catalog.Filter) ([]member.Member, int, error) {
	// PT search must not reveal redacted contact fields through result counts.
	list, total, err := jsonList[member.Member](ctx, r.pool, memberSelect+` WHERE (m.member_code ILIKE $1 OR u.full_name ILIKE $1 OR ($4::boolean AND (u.email ILIKE $1 OR COALESCE(u.phone,'') ILIKE $1))) AND ($2='' OR m.status=$2)`+memberScope(actor), []any{"%" + f.Query + "%", f.Status, actor.ID, actor.Role != identity.RolePT}, f)
	for i := range list {
		list[i] = memberPublic(list[i], actor)
	}
	return list, total, err
}
func (r *PostgresMemberRepository) Get(ctx context.Context, actor identity.User, id int64) (member.Member, error) {
	m, err := jsonOne[member.Member](ctx, r.pool, memberSelect+` WHERE m.id=$1 AND $2::text=''`+memberScope(actor), id, "", actor.ID)
	return memberPublic(m, actor), err
}
func memberByID(ctx context.Context, tx pgx.Tx, id int64) (member.Member, error) {
	return jsonOne[member.Member](ctx, tx, memberSelect+` WHERE m.id=$1`, id)
}
func (r *PostgresMemberRepository) Create(ctx context.Context, in member.Input, hash string, actorID int64, requestID string) (member.Member, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (member.Member, error) {
		userID := in.UserID
		if in.Account != nil {
			a := in.Account
			if err := tx.QueryRow(ctx, `INSERT INTO users(username,password_hash,email,phone,full_name,role_id) SELECT $1,$2,$3,NULLIF($4,''),$5,id FROM roles WHERE code='MEMBER' RETURNING users.id`, a.Username, hash, a.Email, a.Phone, a.FullName).Scan(&userID); err != nil {
				return member.Member{}, err
			}
		} else {
			u, err := findUser(ctx, tx, `SELECT `+userColumns+userFrom+` WHERE u.id=$1 FOR UPDATE`, userID)
			if err != nil {
				return member.Member{}, err
			}
			if u.Role != identity.RoleMember {
				return member.Member{}, identity.Conflict("Chỉ tài khoản MEMBER chưa có hồ sơ mới được liên kết.")
			}
		}
		var id int64
		if err := tx.QueryRow(ctx, `INSERT INTO members(user_id,member_code,date_of_birth,gender,address,join_date,membership_expiry) VALUES($1,$2,NULLIF($3,'')::date,NULLIF($4,''),NULLIF($5,''),$6::date,$7::date) RETURNING id`, userID, in.MemberCode, in.DateOfBirth, in.Gender, in.Address, in.JoinDate, in.MembershipExpiry).Scan(&id); err != nil {
			return member.Member{}, err
		}
		after, err := memberByID(ctx, tx, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, actorID, "member.created", "member", id, "", nil, after, requestID); err != nil {
			return after, err
		}
		return after, nil
	})
}
func (r *PostgresMemberRepository) Update(ctx context.Context, id int64, in member.Input, actorID int64, requestID string) (member.Member, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (member.Member, error) {
		var userID int64
		if err := tx.QueryRow(ctx, `SELECT user_id FROM members WHERE id=$1 FOR UPDATE`, id).Scan(&userID); err != nil {
			return member.Member{}, err
		}
		if _, err := tx.Exec(ctx, `SELECT id FROM users WHERE id=$1 FOR UPDATE`, userID); err != nil {
			return member.Member{}, err
		}
		before, err := memberByID(ctx, tx, id)
		if err != nil {
			return before, err
		}
		if _, err = tx.Exec(ctx, `UPDATE users SET full_name=$2,email=$3,phone=NULLIF($4,''),updated_at=now() WHERE id=$1`, userID, in.FullName, in.Email, in.Phone); err != nil {
			return before, err
		}
		if _, err = tx.Exec(ctx, `UPDATE members SET date_of_birth=NULLIF($2,'')::date,gender=NULLIF($3,''),address=NULLIF($4,''),updated_at=now() WHERE id=$1`, id, in.DateOfBirth, in.Gender, in.Address); err != nil {
			return before, err
		}
		after, err := memberByID(ctx, tx, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, actorID, "member.updated", "member", id, "", before, after, requestID); err != nil {
			return after, err
		}
		return after, nil
	})
}
func (r *PostgresMemberRepository) Membership(ctx context.Context, id int64, in member.MembershipInput, actorID int64, requestID string) (member.Member, error) {
	return inTransaction(ctx, r.pool, func(tx pgx.Tx) (member.Member, error) {
		var locked int64
		if err := tx.QueryRow(ctx, `SELECT id FROM members WHERE id=$1 FOR UPDATE`, id).Scan(&locked); err != nil {
			return member.Member{}, err
		}
		before, err := memberByID(ctx, tx, id)
		if err != nil {
			return before, err
		}
		if _, err = tx.Exec(ctx, `UPDATE members SET membership_expiry=$2::date,status=$3,updated_at=now() WHERE id=$1`, id, in.MembershipExpiry, in.Status); err != nil {
			return before, err
		}
		after, err := memberByID(ctx, tx, id)
		if err != nil {
			return after, err
		}
		if err = writeAudit(ctx, tx, actorID, "member.membership_changed", "member", id, in.Reason, before, after, requestID); err != nil {
			return after, err
		}
		return after, nil
	})
}
