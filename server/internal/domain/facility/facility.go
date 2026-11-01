package facility

type Room struct {
	HasCover         bool   `json:"has_cover"`
	CoverAltText     string `json:"cover_alt_text"`
	CoverVersion     int64  `json:"cover_version"`
	CoverIsDemo      bool   `json:"cover_is_demo"`
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	RoomType         string `json:"room_type"`
	DefaultCapacity  int    `json:"default_capacity"`
	Description      string `json:"description"`
	Status           string `json:"status"`
	CalendarRevision int64  `json:"calendar_revision"`
	CalendarRequired bool   `json:"calendar_required"`
	PreparedThrough  string `json:"calendar_prepared_through"`
	CalendarJobError string `json:"calendar_job_error"`
}
type RoomInput struct {
	Name            string `json:"name"`
	RoomType        string `json:"room_type"`
	DefaultCapacity int    `json:"default_capacity"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	Reason          string `json:"reason"`
}
type Impact struct {
	FutureSessions int `json:"future_sessions"`
	ActiveHolds    int `json:"active_holds"`
}
type TimeSlot struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	IsActive  bool   `json:"is_active"`
}
type SlotInput struct {
	Reason    string `json:"reason"`
	Name      string `json:"name"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}
type SlotStatusInput struct {
	IsActive bool   `json:"is_active"`
	Reason   string `json:"reason"`
}
type Session struct {
	ID            int64  `json:"id"`
	RoomID        int64  `json:"room_id"`
	RoomName      string `json:"room_name"`
	TimeSlotID    int64  `json:"time_slot_id"`
	SlotName      string `json:"slot_name"`
	SessionDate   string `json:"session_date"`
	StartsAt      string `json:"starts_at"`
	EndsAt        string `json:"ends_at"`
	Capacity      int    `json:"capacity"`
	ActiveHolds   int    `json:"active_holds"`
	Status        string `json:"status"`
	ChangeReason  string `json:"change_reason"`
	Origin        string `json:"origin"`
	ClosureSource string `json:"closure_source"`
}
type SessionInput struct {
	RoomID      int64  `json:"room_id"`
	TimeSlotID  int64  `json:"time_slot_id"`
	SessionDate string `json:"session_date"`
	Capacity    int    `json:"capacity"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
}
