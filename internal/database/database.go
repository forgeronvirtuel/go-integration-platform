package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	if err := applyMigrations(db); err != nil {
		return nil, err
	}

	log.Info().Str("path", dbPath).Msg("Base de données SQLite connectée")
	return db, nil
}

func createTables(db *sql.DB) error {
	if err := CreateProjectsTable(db); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la création de la table projects")
		return err
	}

	if err := CreateBuildsTable(db); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la création de la table builds")
		return err
	}

	if err := CreateAgentsTable(db); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la création de la table agents")
		return err
	}

	if err := CreateDeploymentsTable(db); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la création de la table deployments")
		return err
	}

	return nil
}
