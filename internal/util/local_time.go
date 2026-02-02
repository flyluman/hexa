package util

import "time"

var dhaka *time.Location

func init() {
	dhaka, _ = time.LoadLocation("Asia/Dhaka")
}

func Now() time.Time {
	return time.Now().In(dhaka)
}
