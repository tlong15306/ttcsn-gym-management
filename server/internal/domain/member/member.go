package member

import "github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"

type Member struct {
	ID               int64         `json:"id"`
	UserID           int64         `json:"user_id"`
	User             identity.User `json:"user"`
	MemberCode       string        `json:"member_code"`
	DateOfBirth      string        `json:"date_of_birth"`
	Gender           string        `json:"gender"`
	Address          string        `json:"address"`
	JoinDate         string        `json:"join_date"`
	MembershipExpiry string        `json:"membership_expiry"`
	Status           string        `json:"status"`
	Expired          bool          `json:"expired"`
}
type AccountInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}
type Input struct {
	UserID           int64         `json:"user_id"`
	Account          *AccountInput `json:"account,omitempty"`
	MemberCode       string        `json:"member_code"`
	DateOfBirth      string        `json:"date_of_birth"`
	Gender           string        `json:"gender"`
	Address          string        `json:"address"`
	JoinDate         string        `json:"join_date"`
	MembershipExpiry string        `json:"membership_expiry"`
	FullName         string        `json:"full_name"`
	Email            string        `json:"email"`
	Phone            string        `json:"phone"`
}
type MembershipInput struct {
	MembershipExpiry string `json:"membership_expiry"`
	Status           string `json:"status"`
	Reason           string `json:"reason"`
}
