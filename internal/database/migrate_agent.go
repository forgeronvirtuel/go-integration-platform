package database

import (
	"database/sql"
	"strings"
)

func MigrateAddAgentURLv1(db *sql.DB) error {
	// Try to add the column; if it already exists, SQLite will error.
	_, err := db.Exec(`ALTER TABLE agents ADD COLUMN url TEXT NOT NULL DEFAULT '';`)
	if err != nil {
		// "duplicate column name: url" => migration already applied
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "duplicate column name") && strings.Contains(msg, "url") {
			return nil
		}
		return err
	}
	return nil
}
