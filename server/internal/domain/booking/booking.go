package booking

import (
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"time"
)

type CancellationSource string

const (
	CancellationSourceMember   CancellationSource = "MEMBER"
	CancellationSourcePT       CancellationSource = "PT"
	CancellationSourceFacility CancellationSource = "FACILITY"
	CancellationSourceUnknown  CancellationSource = "UNKNOWN"
)

func CancellationProvenance(role identity.Role) (CancellationSource, bool) {
	switch role {
	case identity.RoleMember:
		return CancellationSourceMember, true
	case identity.RolePT:
		return CancellationSourcePT, true
	case identity.RoleAdmin, identity.RoleManager, identity.RoleReceptionist:
		return CancellationSourceFacility, true
	default:
		return "", false
	}
}

type Booking struct {
	HasAttendance      bool                `json:"has_attendance"`
	ID                 int64               `json:"id"`
	MemberID           int64               `json:"member_id"`
	MemberName         string              `json:"member_name"`
	RoomSessionID      int64               `json:"room_session_id"`
	RoomID             int64               `json:"room_id"`
	RoomName           string              `json:"room_name"`
	PTID               int64               `json:"pt_id"`
	PTName             string              `json:"pt_name"`
	Status             string              `json:"status"`
	StartsAt           time.Time           `json:"starts_at"`
	EndsAt             time.Time           `json:"ends_at"`
	ConfirmedAt        *time.Time          `json:"confirmed_at"`
	RespondBy          *time.Time          `json:"respond_by"`
	CreatedAt          time.Time           `json:"created_at"`
	CancellationReason string              `json:"cancellation_reason"`
	CancellationSource *CancellationSource `json:"cancellation_source"`
	CancelledActorRole *identity.Role      `json:"cancelled_actor_role"`
	ServerTime         time.Time           `json:"server_time"`
}
type CreateInput struct {
	RoomSessionID int64 `json:"room_session_id"`
	PTID          int64 `json:"pt_id,omitempty"`
}

// BatchCreateInput creates bookings for one contiguous selection of room
// sessions. The server owns the ordering and rejects sessions that do not
// belong to the same room and date or are not immediately consecutive.
type BatchCreateInput struct {
	RoomSessionIDs []int64 `json:"room_session_ids"`
	PTID           int64   `json:"pt_id,omitempty"`
}
type ActionInput struct {
	Action string `json:"action,omitempty"`
	Reason string `json:"reason,omitempty"`
}
type Filter struct {
	History bool
	catalog.Filter
	From string
	To   string
	PTID int64
}
type SessionOption struct {
	facility.Session
	Remaining         int    `json:"remaining"`
	UnavailableReason string `json:"unavailable_reason"`
}
type PTOption struct {
	ID            int64    `json:"id"`
	FullName      string   `json:"full_name"`
	Specialty     string   `json:"specialty"`
	RatingAverage *float64 `json:"rating_average"`
	ReviewCount   int      `json:"review_count"`
}
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string        { return e.Message }
func Fail(code, message string) error { return &Error{code, message} }
func Deadline(created, start time.Time) time.Time {
	deadline := created.Add(30 * time.Minute)
	if start.Before(deadline) {
		return start
	}
	return deadline
}
func CanCancel(now, start time.Time) bool { return !now.After(start.Add(-2 * time.Hour)) }
