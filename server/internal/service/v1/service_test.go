package v1service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
)

type fakeRepository struct {
	user identity.User
}

func (r *fakeRepository) FindByUsername(context.Context, string) (identity.User, error) {
	return r.user, nil
}
func (r *fakeRepository) FindByID(context.Context, int64) (identity.User, error) { return r.user, nil }
func (r *fakeRepository) List(context.Context, identity.ListFilter) ([]identity.User, int, error) {
	return []identity.User{r.user}, 1, nil
}
func (r *fakeRepository) Create(context.Context, identity.CreateInput, string, int64, string) (identity.User, error) {
	return r.user, nil
}
func (r *fakeRepository) UpdateProfile(context.Context, int64, identity.ProfileInput) (identity.User, error) {
	return r.user, nil
}
func (r *fakeRepository) UpdateAccount(context.Context, int64, identity.ProfileInput, int64, string) (identity.User, error) {
	return r.user, nil
}
func (r *fakeRepository) SetStatus(context.Context, int64, identity.Status, string, int64, string) (identity.User, error) {
	return r.user, nil
}
func (r *fakeRepository) SetRole(context.Context, int64, identity.Role, string, int64, string) (identity.User, error) {
	return r.user, nil
}
func (r *fakeRepository) ResetPassword(context.Context, int64, string, int64, string) (identity.User, error) {
	return r.user, nil
}
func (r *fakeRepository) BumpAuthVersion(context.Context, int64, int64, string) error { return nil }

type fakeHasher struct{ err error }

func (h fakeHasher) Hash(string) (string, error)  { return "hash", h.err }
func (h fakeHasher) Compare(string, string) error { return h.err }

type fakeTokens struct{ err error }

func (t fakeTokens) Issue(identity.User) (string, error) { return "token", t.err }

func TestLoginRejectsLockedAccount(t *testing.T) {
	repository := &fakeRepository{user: identity.User{ID: 1, Username: "locked", Role: identity.RoleAdmin, Status: identity.StatusLocked, AuthVersion: 1}}
	service := NewService(repository, fakeHasher{}, fakeTokens{})
	_, err := service.Login(context.Background(), "locked", "password", 15*time.Minute)
	if !errors.Is(err, identity.ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want invalid credentials", err)
	}
}

func TestCreateUserRequiresAdminAndValidInput(t *testing.T) {
	repository := &fakeRepository{user: identity.User{ID: 1, Role: identity.RoleAdmin, Status: identity.StatusActive, AuthVersion: 1}}
	service := NewService(repository, fakeHasher{}, fakeTokens{})
	input := identity.CreateInput{Username: "new.account", Password: "Password123!", Email: "new@example.com", FullName: "New Account", Role: identity.RoleReceptionist}
	if _, err := service.CreateUser(context.Background(), identity.User{Role: identity.RoleMember}, input, "request"); !errors.Is(err, identity.ErrForbidden) {
		t.Fatalf("CreateUser() error = %v, want forbidden", err)
	}
	if _, err := service.CreateUser(context.Background(), repository.user, identity.CreateInput{}, "request"); !errors.Is(err, identity.ErrValidation) {
		t.Fatalf("CreateUser() error = %v, want validation", err)
	}
}
