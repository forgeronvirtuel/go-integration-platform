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
	// Table users
	usersQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(usersQuery); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la création de la table users")
		return err
	}

	log.Info().Msg("Table 'users' créée ou déjà existante")

	// Table projects
	if err := CreateProjectsTable(db); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la création de la table projects")
		return err
	}

	log.Info().Msg("Table 'projects' créée ou déjà existante")

	// Insérer des données de test si la table est vide
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}

	if count == 0 {
		insertQuery := `
		INSERT INTO users (name, email) VALUES
			('Alice Dupont', 'alice@example.com'),
			('Bob Martin', 'bob@example.com'),
			('Charlie Bernard', 'charlie@example.com');
		`
		if _, err := db.Exec(insertQuery); err != nil {
			log.Error().Err(err).Msg("Erreur lors de l'insertion des données de test")
			return err
		}
		log.Info().Msg("Données de test insérées dans la table 'users'")
	}

	return nil
}

// ExampleQuery montre comment faire une requête SQL
func ExampleQuery(db *sql.DB) error {
	rows, err := db.Query("SELECT id, name, email FROM users WHERE id = ?", 1)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			return err
		}
		log.Info().Int("id", id).Str("name", name).Str("email", email).Msg("Utilisateur trouvé")
	}

	return rows.Err()
}
