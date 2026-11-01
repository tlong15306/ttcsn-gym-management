package report

import "time"

const Timezone = "Asia/Ho_Chi_Minh"

type Period struct {
	From     string    `json:"from"`
	To       string    `json:"to"`
	Timezone string    `json:"timezone"`
	AsOf     time.Time `json:"as_of"`
}

type Rate struct {
	Numerator   int64    `json:"numerator"`
	Denominator int64    `json:"denominator"`
	Percent     *float64 `json:"rate_percent"`
}

type SeatMinutes struct {
	Booked   int64    `json:"booked"`
	Capacity int64    `json:"capacity"`
	Percent  *float64 `json:"rate_percent"`
}

type Operations struct {
	Period
	TotalBookings            int64            `json:"total_bookings"`
	StatusCounts             map[string]int64 `json:"status_counts"`
	MemberAttendanceCount    int64            `json:"member_attendance_count"`
	PTAttendanceCount        int64            `json:"pt_attendance_count"`
	NoShow                   Rate             `json:"no_show"`
	PendingCompletionCount   int64            `json:"pending_completion_count"`
	Utilization              SeatMinutes      `json:"utilization"`
	CancellationSourceCounts map[string]int64 `json:"cancellation_source_counts"`
	UnclassifiedBookingCount int64            `json:"unclassified_booking_count"`
}

type Trainer struct {
	Period
	PTID                   int64    `json:"pt_id"`
	RatingAverage          *float64 `json:"rating_average"`
	ReviewCount            int64    `json:"review_count"`
	Attendance             Rate     `json:"attendance"`
	ExcludedCount          int64    `json:"excluded_count"`
	UnclassifiedCount      int64    `json:"unclassified_count"`
	PendingCompletionCount int64    `json:"pending_completion_count"`
}

type PeriodInput struct {
	From string
	To   string
}

type Range struct {
	From time.Time
	To   time.Time
}

// ResolveRange applies the current Vietnam calendar month when either bound is omitted.
// The inclusive range is limited to 366 local calendar days.
func ResolveRange(input PeriodInput, now time.Time) (Range, error) {
	loc, err := time.LoadLocation(Timezone)
	if err != nil {
		return Range{}, err
	}
	localNow := now.In(loc)
	from := time.Date(localNow.Year(), localNow.Month(), 1, 0, 0, 0, 0, loc)
	to := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc)
	if input.From != "" {
		from, err = time.ParseInLocation("2006-01-02", input.From, loc)
		if err != nil || from.Format("2006-01-02") != input.From {
			return Range{}, errInvalidRange
		}
	}
	if input.To != "" {
		to, err = time.ParseInLocation("2006-01-02", input.To, loc)
		if err != nil || to.Format("2006-01-02") != input.To {
			return Range{}, errInvalidRange
		}
	}
	if to.Before(from) || int(to.Sub(from).Hours()/24) > 365 {
		return Range{}, errInvalidRange
	}
	return Range{From: from, To: to}, nil
}

var errInvalidRange = invalidRangeError{}

type invalidRangeError struct{}

func (invalidRangeError) Error() string { return "invalid report date range" }
