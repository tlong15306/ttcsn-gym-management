package review

import "time"

// Review deliberately exposes no reviewer, member, PT or booking identifiers.
type Review struct {
	Rating        int       `json:"rating"`
	Comment       string    `json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
	ReviewerLabel string    `json:"reviewer_label"`
}

type Input struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

func Fail(code, message string) error { return &Error{Code: code, Message: message} }
