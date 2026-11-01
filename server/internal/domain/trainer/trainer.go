package trainer

import "github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"

type Profile struct {
	ID             int64         `json:"id"`
	UserID         int64         `json:"user_id"`
	User           identity.User `json:"user"`
	Specialty      string        `json:"specialty"`
	ExperienceText string        `json:"experience_text"`
	HasCertificate bool          `json:"has_certificate"`
	ApprovalStatus string        `json:"approval_status"`
	DecisionReason string        `json:"decision_reason"`
	DecidedAt      *string       `json:"decided_at"`
	RatingAverage  *float64      `json:"rating_average"`
	ReviewCount    int           `json:"review_count"`
}
type ProfileInput struct {
	Specialty      string `json:"specialty"`
	ExperienceText string `json:"experience_text"`
}
type DecisionInput struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}
type WorkingHour struct {
	ID        int64  `json:"id"`
	DayOfWeek int    `json:"day_of_week"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}
type AvailabilityInput struct {
	Hours []WorkingHour `json:"hours"`
}
type Unavailability struct {
	ID       int64  `json:"id"`
	StartsAt string `json:"starts_at"`
	EndsAt   string `json:"ends_at"`
	Reason   string `json:"reason"`
}
