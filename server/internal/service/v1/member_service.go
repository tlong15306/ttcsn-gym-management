package v1service

import (
	"context"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/member"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	"strings"
	"time"
)

type MemberService struct {
	repo   repository.MemberRepository
	hasher PasswordHasher
}

func (s *MemberService) EligibleAccounts(ctx context.Context, a identity.User, f catalog.Filter) ([]identity.User, int, error) {
	if !managesMembers(a) {
		return nil, 0, identity.ErrForbidden
	}
	f, err := normalizeFilter(f)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.EligibleAccounts(ctx, f)
}

func NewMemberService(repo repository.MemberRepository, hasher PasswordHasher) *MemberService {
	return &MemberService{repo, hasher}
}
func managesMembers(a identity.User) bool {
	return a.Role == identity.RoleAdmin || a.Role == identity.RoleManager || a.Role == identity.RoleReceptionist
}
func dateValid(s string) bool { _, err := time.Parse("2006-01-02", s); return err == nil }
func normalizeFilter(f catalog.Filter) (catalog.Filter, error) {
	if f.Page < 1 || f.PageSize < 1 || f.PageSize > 100 || f.Page > 1000000 {
		return f, identity.ErrValidation
	}
	f.Query = strings.TrimSpace(f.Query)
	return f, nil
}
func (s *MemberService) List(ctx context.Context, a identity.User, f catalog.Filter) ([]member.Member, int, error) {
	f, err := normalizeFilter(f)
	if err != nil {
		return nil, 0, err
	}
	if f.Status != "" && f.Status != "ACTIVE" && f.Status != "INACTIVE" {
		return nil, 0, identity.ErrValidation
	}
	return s.repo.List(ctx, a, f)
}
func (s *MemberService) Get(ctx context.Context, a identity.User, id int64) (member.Member, error) {
	return s.repo.Get(ctx, a, id)
}
func validMember(in member.Input, create bool) bool {
	if len(in.Address) > 2000 || len(in.Gender) > 20 {
		return false
	}
	if in.DateOfBirth != "" && (!dateValid(in.DateOfBirth) || in.DateOfBirth > time.Now().Format("2006-01-02")) {
		return false
	}
	if create {
		return strings.TrimSpace(in.MemberCode) != "" && len(in.MemberCode) <= 30 && dateValid(in.JoinDate) && dateValid(in.MembershipExpiry) && in.MembershipExpiry >= in.JoinDate
	}
	return validContact(in.FullName, in.Email, in.Phone)
}
func (s *MemberService) Create(ctx context.Context, a identity.User, in member.Input, requestID string) (member.Member, error) {
	if !managesMembers(a) {
		return member.Member{}, identity.ErrForbidden
	}
	if !validMember(in, true) || (in.UserID > 0) == (in.Account != nil) {
		return member.Member{}, identity.ErrValidation
	}
	hash := ""
	var err error
	if in.Account != nil {
		account := in.Account
		v := identity.CreateInput{Username: account.Username, Password: account.Password, FullName: account.FullName, Email: account.Email, Phone: account.Phone, Role: identity.RoleMember}
		if err = validateCreate(v); err != nil {
			return member.Member{}, err
		}
		hash, err = s.hasher.Hash(account.Password)
		if err != nil {
			return member.Member{}, err
		}
		account.Username = identity.NormalizeIdentifier(account.Username)
		account.Email = strings.ToLower(strings.TrimSpace(account.Email))
		account.FullName = strings.TrimSpace(account.FullName)
	}
	in.MemberCode = strings.TrimSpace(in.MemberCode)
	return s.repo.Create(ctx, in, hash, a.ID, requestID)
}
func (s *MemberService) Update(ctx context.Context, a identity.User, id int64, in member.Input, requestID string) (member.Member, error) {
	if !managesMembers(a) {
		return member.Member{}, identity.ErrForbidden
	}
	if !validMember(in, false) {
		return member.Member{}, identity.ErrValidation
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.FullName = strings.TrimSpace(in.FullName)
	return s.repo.Update(ctx, id, in, a.ID, requestID)
}
func (s *MemberService) Membership(ctx context.Context, a identity.User, id int64, in member.MembershipInput, requestID string) (member.Member, error) {
	if !managesMembers(a) {
		return member.Member{}, identity.ErrForbidden
	}
	if !dateValid(in.MembershipExpiry) || (in.Status != "ACTIVE" && in.Status != "INACTIVE") || strings.TrimSpace(in.Reason) == "" {
		return member.Member{}, identity.ErrValidation
	}
	return s.repo.Membership(ctx, id, in, a.ID, requestID)
}
