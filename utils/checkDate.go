package utils

import (
	"time"
)

func CheckDate(date string) (time.Time, bool) {
	d, err := time.Parse(dateFormat, date)
	if err != nil {
		return time.Time{}, false
	}
	t := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	return t, t.Year() == d.Year() && t.Month() == d.Month() && t.Day() == d.Day()
}
