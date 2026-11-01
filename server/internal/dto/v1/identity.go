package v1dto

import (
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
)

// LoginRequest is the transport contract for POST /api/v1/auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UpdateProfileRequest deliberately contains only fields a user may update.
type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

func (r UpdateProfileRequest) ToServiceInput() identity.ProfileInput {
	return identity.ProfileInput{FullName: r.FullName, Email: r.Email, Phone: r.Phone}
}

// CreateUserRequest is restricted to fields accepted by the account API.
type CreateUserRequest struct {
	Username string        `json:"username"`
	Password string        `json:"password"`
	Email    string        `json:"email"`
	Phone    string        `json:"phone"`
	FullName string        `json:"full_name"`
	Role     identity.Role `json:"role"`
}

func (r CreateUserRequest) ToServiceInput() identity.CreateInput {
	return identity.CreateInput{
		Username: r.Username,
		Password: r.Password,
		Email:    r.Email,
		Phone:    r.Phone,
		FullName: r.FullName,
		Role:     r.Role,
	}
}

type UpdateStatusRequest struct {
	Status identity.Status `json:"status"`
	Reason string          `json:"reason"`
}

type UpdateRoleRequest struct {
	Role   identity.Role `json:"role"`
	Reason string        `json:"reason"`
}

type ResetPasswordRequest struct {
	Password string `json:"password"`
}
