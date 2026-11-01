package identity

// PublicError carries a deliberate, safe business message rather than driver details.
type PublicError struct {
	Message string
	Cause   error
}

func (e *PublicError) Error() string { return e.Message }
func (e *PublicError) Unwrap() error { return e.Cause }
func Conflict(message string) error  { return &PublicError{Message: message, Cause: ErrConflict} }
