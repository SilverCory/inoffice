package inoffice

import (
	"database/sql"
	"fmt"
	"office"
	_ "office/inoffice/migrations"
	"time"

	"github.com/SilverCory/sqlstruct"

	"github.com/pressly/goose/v3"

	_ "github.com/go-sql-driver/mysql"
)

var _ Store = new(StoreMySQL)

type StoreMySQL struct {
	DB *sql.DB
}

func NewStore(env office.Env) *StoreMySQL {
	db, err := sql.Open("mysql", env.MySQLDSN)
	if err != nil {
		panic(err)
	}

	ensureUpToDate(env)

	return &StoreMySQL{DB: db}
}

func (s *StoreMySQL) GetByDay(weekStart time.Time, day Day) ([]InOffice, error) {
	var query = fmt.Sprintf(`SELECT %s FROM inoffice WHERE deleted = FALSE AND week_start = ? AND in_on = ?`, sqlstruct.Columns(InOffice{}))
	rows, err := s.DB.Query(query, weekStart, day)
	if err != nil {
		return nil, err
	}

	var ret []InOffice
	for rows.Next() {
		var t InOffice
		if err := sqlstruct.Scan(&t, rows); err != nil {
			return nil, err
		}

		ret = append(ret, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (s *StoreMySQL) GetAllForWeek(weekStart time.Time) (map[Day][]InOffice, error) {
	var query = fmt.Sprintf(`SELECT %s FROM inoffice WHERE deleted = FALSE AND week_start = ?`, sqlstruct.Columns(InOffice{}))
	rows, err := s.DB.Query(query, weekStart)
	if err != nil {
		return nil, err
	}

	var ret = make(map[Day][]InOffice)
	for rows.Next() {
		var t InOffice
		if err := sqlstruct.Scan(&t, rows); err != nil {
			return nil, err
		}

		ret[t.InOn] = append(ret[t.InOn], t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (s *StoreMySQL) Save(office InOffice) error {
	const query = `INSERT INTO inoffice (user_id, user_name, in_on, week_start) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE deleted = !deleted`
	_, err := s.DB.Exec(query, office.UserID, office.Username, office.InOn, office.WeekStart)
	return err
}

func ensureUpToDate(env office.Env) {
	db, err := goose.OpenDBWithDriver("mysql", env.MySQLDSN)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()

	if err := goose.Run("up", db, "."); err != nil {
		panic(err)
	}

	return
}
