package utils

import (
	"time"
)

const DateFormat = "20060102"
const DateSearchFormat = "02.01.2006"

func CheckDate(date string) (time.Time, bool) {
	d, err := time.Parse(DateFormat, date)
	if err != nil {
		return time.Time{}, false
	}
	t := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	return t, t.Year() == d.Year() && t.Month() == d.Month() && t.Day() == d.Day()
}
