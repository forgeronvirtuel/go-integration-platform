package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// InitDB initialise la connexion à la base de données SQLite et crée les tables
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Activer les foreign keys pour SQLite
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	// Vérifier la connexion
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Créer les tables
	if err := createTables(db); err != nil {
		return nil, err
	}

	log.Info().Str("path", dbPath).Msg("Base de données SQLite connectée")
	return db, nil
}

// createTables crée les tables nécessaires
func createTables(db *sql.DB) error {
	// Table projects
	if err := CreateProjectsTable(db); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la création de la table projects")
		return err
	}

	// Table builds
	if err := CreateBuildsTable(db); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la création de la table builds")
		return err
	}

	return nil
}
