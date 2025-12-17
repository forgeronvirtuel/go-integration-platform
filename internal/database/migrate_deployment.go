package database

import (
	"database/sql"
	"strings"
)

func MigrateDeploymentRenameRunnerToRunnerV1(db *sql.DB) error {
	_, err := db.Exec("ALTER TABLE deployments RENAME COLUMN runner_id TO runner_id;")
	if err != nil {
		msg := strings.ToLower(err.Error())
		if !(strings.Contains(msg, "no such column") && strings.Contains(msg, "runner_id")) {
			return err
		}
	}

	_, err = db.Exec(`DROP INDEX IF EXISTS idx_deployments_runner_id;`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	CREATE INDEX IF NOT EXISTS idx_deployments_runner_id ON deployments(runner_id);
	`)
	if err != nil {
		return err
	}

	return nil
}
