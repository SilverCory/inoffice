package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(Up0001, Down0001)
}

func Up0001(tx *sql.Tx) error {
	var query = `
CREATE TABLE inoffice (
  user_id VARCHAR(30) NOT NULL,
  user_name VARCHAR(255) CHARACTER SET 'utf8' COLLATE 'utf8_bin' NOT NULL,
  in_on ENUM('MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY') NOT NULL,
  week_start DATE NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted BOOLEAN NOT NULL DEFAULT FALSE,
  UNIQUE INDEX uq_in_office (user_id ASC, in_on ASC, week_start ASC) VISIBLE);`
	_, err := tx.Exec(query)
	return err
}

func Down0001(tx *sql.Tx) error {
	var query = `DROP TABLE inoffice;`
	_, err := tx.Exec(query)
	return err
}
