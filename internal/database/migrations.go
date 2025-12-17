package database

import (
	"database/sql"

	"github.com/rs/zerolog/log"
)

func applyMigrations(db *sql.DB) error {
	migrations := []func(*sql.DB) error{
		MigrateAddRunnerURLv1,
		MigrateDeploymentRenameRunnerToRunnerV1,
	}

	for i, migrate := range migrations {
		log.Info().Int("migration", i+1).Msg("Applying database migration")
		if err := migrate(db); err != nil {
			return err
		}
	}

	return nil
}
