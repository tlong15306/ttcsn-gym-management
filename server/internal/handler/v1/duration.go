package v1handler

import "time"

func secondsToDuration(seconds int64) time.Duration {
	return time.Duration(seconds) * time.Second
}
