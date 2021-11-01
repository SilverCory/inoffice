package inoffice

import "time"

type Store interface {
	GetByDay(weekStart time.Time, day Day) ([]InOffice, error)
	GetAllForWeek(weekStart time.Time) (map[Day][]InOffice, error)
	Save(office InOffice) error
}
