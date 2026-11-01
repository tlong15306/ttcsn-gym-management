package repository

import (
	"context"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
)

// UserRepository is the persistence contract used by service/v1. It contains
// no HTTP or Gin dependency, so service tests can use a small in-memory fake.
type UserRepository interface {
	FindByUsername(context.Context, string) (identity.User, error)
	FindByID(context.Context, int64) (identity.User, error)
	List(context.Context, identity.ListFilter) ([]identity.User, int, error)
	Create(context.Context, identity.CreateInput, string, int64, string) (identity.User, error)
	UpdateProfile(context.Context, int64, identity.ProfileInput) (identity.User, error)
	UpdateAccount(context.Context, int64, identity.ProfileInput, int64, string) (identity.User, error)
	SetStatus(context.Context, int64, identity.Status, string, int64, string) (identity.User, error)
	SetRole(context.Context, int64, identity.Role, string, int64, string) (identity.User, error)
	ResetPassword(context.Context, int64, string, int64, string) (identity.User, error)
	BumpAuthVersion(context.Context, int64, int64, string) error
}
