package discovery

import "time"

// Explicit anonymous projections. Never embed account or management models.
type Room struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	RoomType        string `json:"room_type"`
	DefaultCapacity int    `json:"default_capacity"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	HasCover        bool   `json:"has_cover"`
	CoverAltText    string `json:"cover_alt_text"`
	CoverVersion    int64  `json:"cover_version"`
	CoverIsDemo     bool   `json:"cover_is_demo"`
}

type Trainer struct {
	ID             int64    `json:"id"`
	FullName       string   `json:"full_name"`
	Specialty      string   `json:"specialty"`
	ExperienceText string   `json:"experience_text"`
	RatingAverage  *float64 `json:"rating_average"`
	ReviewCount    int      `json:"review_count"`
}

type Session struct {
	ID          int64     `json:"id"`
	RoomID      int64     `json:"room_id"`
	SessionDate string    `json:"session_date"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	Capacity    int       `json:"capacity"`
	Remaining   int       `json:"remaining"`
	Status      string    `json:"status"`
}
