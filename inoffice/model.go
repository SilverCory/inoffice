package inoffice

import (
	"time"
)

type Day string

const (
	DayMonday    Day = "MONDAY"
	DayTuesday   Day = "TUESDAY"
	DayWednesday Day = "WEDNESDAY"
	DayThursday  Day = "THURSDAY"
	DayFriday    Day = "FRIDAY"
)

type InOffice struct {
	UserID    string    `sql:"user_id"`
	Username  string    `sql:"user_name"`
	InOn      Day       `sql:"in_on"`
	WeekStart time.Time `sql:"week_start"`

	CreatedAt time.Time `sql:"created_at"`
	Deleted   bool      `sql:"deleted"`
}

func StartOfWeek(t time.Time) time.Time {
	weekday := time.Duration(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	year, month, day := t.Date()
	currentZeroDay := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return currentZeroDay.Add(-1 * (weekday - 1) * 24 * time.Hour)
}

func NextWeek(t time.Time) time.Time {
	return StartOfWeek(StartOfWeek(t).Add(7 * (25 * time.Hour))) // Overrun to come back.
}

func IsInPast(startOfWeek time.Time, day Day) bool {
	var daysAdd time.Duration = 0
	switch day {
	case DayMonday:
		daysAdd = 0
	case DayTuesday:
		daysAdd = 1
	case DayWednesday:
		daysAdd = 2
	case DayThursday:
		daysAdd = 3
	case DayFriday:
		daysAdd = 4
	default:
		panic("invalid day: " + day)
	}

	dayIn := startOfWeek.Add(daysAdd * (time.Hour * 24))
	now := time.Now()

	// End of day is 15:00-16:00 (CBA to deal with DST)
	var todayEnd = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	var fromDay = time.Date(dayIn.Year(), dayIn.Month(), dayIn.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), time.UTC)

	// add special condition for today to end at 1600
	if dayIn.Equal(todayEnd) {
		return !fromDay.Before(todayEnd.Add(15 * time.Hour))
	}

	return fromDay.Before(todayEnd)
}
