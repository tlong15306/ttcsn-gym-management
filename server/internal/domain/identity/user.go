package identity

import (
	"errors"
	"strings"
	"time"
)

type Role string

const (
	RoleMember       Role = "MEMBER"
	RolePT           Role = "PT"
	RoleReceptionist Role = "RECEPTIONIST"
	RoleManager      Role = "MANAGER"
	RoleAdmin        Role = "ADMIN"
)

func (r Role) Valid() bool {
	switch r {
	case RoleMember, RolePT, RoleReceptionist, RoleManager, RoleAdmin:
		return true
	default:
		return false
	}
}

type Status string

const (
	StatusActive Status = "ACTIVE"
	StatusLocked Status = "LOCKED"
)

func (s Status) Valid() bool {
	return s == StatusActive || s == StatusLocked
}

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone,omitempty"`
	FullName     string    `json:"full_name"`
	Role         Role      `json:"role"`
	Status       Status    `json:"status"`
	AuthVersion  int       `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func NormalizeIdentifier(value string) string {
	return strings.TrimSpace(value)
}

var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("conflict")
	ErrValidation         = errors.New("validation failed")
	ErrLastActiveAdmin    = errors.New("last active admin")
	ErrProfileRequirement = errors.New("profile requirement")
)
