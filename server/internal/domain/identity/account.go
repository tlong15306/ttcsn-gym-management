package identity

// ProfileInput contains the only profile fields an account owner may change.
type ProfileInput struct {
	FullName string
	Email    string
	Phone    string
}

// CreateInput contains the account fields accepted by the account service.
type CreateInput struct {
	Username string
	Password string
	Email    string
	Phone    string
	FullName string
	Role     Role
}

// ListFilter describes a bounded administrator account-list request.
type ListFilter struct {
	Query    string
	Role     Role
	Status   Status
	Page     int
	PageSize int
}
