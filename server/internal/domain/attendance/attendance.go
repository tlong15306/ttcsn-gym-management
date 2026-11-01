package attendance

import "time"

type Record struct {
	ID                int64     `json:"id"`
	BookingID         int64     `json:"booking_id"`
	ParticipantType   string    `json:"participant_type"`
	ParticipantUserID int64     `json:"participant_user_id"`
	CheckedInAt       time.Time `json:"checked_in_at"`
	RecordedByUserID  int64     `json:"recorded_by_user_id"`
	RecorderName      string    `json:"recorder_name"`
	OverrideReason    string    `json:"override_reason"`
}
type Participant struct {
	Type       string  `json:"participant_type"`
	Name       string  `json:"name"`
	Attendance *Record `json:"attendance"`
}
type Summary struct {
	Participants []Participant `json:"participants"`
	ServerTime   time.Time     `json:"server_time"`
}
type Input struct {
	ParticipantType string `json:"participant_type"`
	OverrideReason  string `json:"override_reason"`
}
type Result struct {
	Record  Record
	Created bool
}
type Error struct{ Code, Message string }

func (e *Error) Error() string { return e.Message }
func InWindow(now, start, end time.Time) bool {
	return !now.Before(start.Add(-15*time.Minute)) && now.Before(end)
}
