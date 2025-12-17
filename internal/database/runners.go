package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
)

// Runner représente un runner de build
type Runner struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	URL        string            `json:"url"` // URL pour contacter le runner
	Labels     map[string]string `json:"labels"`
	Status     string            `json:"status"` // ONLINE, OFFLINE, DRAINING
	LastSeenAt time.Time         `json:"last_seen_at"`
	CreatedAt  time.Time         `json:"created_at"`
}

// CreateRunnersTable crée la table runners si elle n'existe pas
func CreateRunnersTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS runners (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL DEFAULT '',
		labels TEXT NOT NULL DEFAULT '{}',
		status TEXT NOT NULL DEFAULT 'OFFLINE' CHECK(status IN ('ONLINE', 'OFFLINE', 'DRAINING')),
		last_seen_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_runners_status ON runners(status);
	CREATE INDEX IF NOT EXISTS idx_runners_name ON runners(name);
	`

	if _, err := db.Exec(query); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la création de la table runners")
		return err
	}

	log.Info().Msg("Table 'runners' créée ou déjà existante")
	return nil
}

// CreateRunner crée un nouvel runner
func CreateRunner(db *sql.DB, name string, labels map[string]string) (*Runner, error) {
	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO runners (name, url, labels, status, last_seen_at)
		VALUES (?, '', ?, 'OFFLINE', CURRENT_TIMESTAMP)
	`

	result, err := db.Exec(query, name, string(labelsJSON))
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetRunnerByID(db, int(id))
}

// GetRunnerByID récupère un runner par son ID
func GetRunnerByID(db *sql.DB, id int) (*Runner, error) {
	query := `
		SELECT id, name, url, labels, status, last_seen_at, created_at
		FROM runners
		WHERE id = ?
	`

	var runner Runner
	var labelsJSON string
	var lastSeenAt sql.NullTime

	err := db.QueryRow(query, id).Scan(
		&runner.ID,
		&runner.Name,
		&runner.URL,
		&labelsJSON,
		&runner.Status,
		&lastSeenAt,
		&runner.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if lastSeenAt.Valid {
		runner.LastSeenAt = lastSeenAt.Time
	}

	if err := json.Unmarshal([]byte(labelsJSON), &runner.Labels); err != nil {
		runner.Labels = make(map[string]string)
	}

	return &runner, nil
}

// GetRunnerByName récupère un runner par son nom (hostname)
func GetRunnerByName(db *sql.DB, name string) (*Runner, error) {
	query := `
		SELECT id, name, url, labels, status, last_seen_at, created_at
		FROM runners
		WHERE name = ?
	`

	var runner Runner
	var labelsJSON string
	var lastSeenAt sql.NullTime

	err := db.QueryRow(query, name).Scan(
		&runner.ID,
		&runner.Name,
		&runner.URL,
		&labelsJSON,
		&runner.Status,
		&lastSeenAt,
		&runner.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if lastSeenAt.Valid {
		runner.LastSeenAt = lastSeenAt.Time
	}

	if err := json.Unmarshal([]byte(labelsJSON), &runner.Labels); err != nil {
		runner.Labels = make(map[string]string)
	}

	return &runner, nil
}

func GetAllRunners(db *sql.DB) ([]Runner, error) {
	query := `
		SELECT id, name, url, labels, status, last_seen_at, created_at
		FROM runners
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runners []Runner
	for rows.Next() {
		var runner Runner
		var labelsJSON string
		var lastSeenAt sql.NullTime

		err := rows.Scan(
			&runner.ID,
			&runner.Name,
			&runner.URL,
			&labelsJSON,
			&runner.Status,
			&lastSeenAt,
			&runner.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		if lastSeenAt.Valid {
			runner.LastSeenAt = lastSeenAt.Time
		}

		if err := json.Unmarshal([]byte(labelsJSON), &runner.Labels); err != nil {
			runner.Labels = make(map[string]string)
		}

		runners = append(runners, runner)
	}

	return runners, nil
}

// GetRunnersByStatus récupère les runners par statut
func GetRunnersByStatus(db *sql.DB, status string) ([]Runner, error) {
	query := `
		SELECT id, name, url, labels, status, last_seen_at, created_at
		FROM runners
		WHERE status = ?
		ORDER BY last_seen_at DESC
	`

	rows, err := db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runners []Runner
	for rows.Next() {
		var runner Runner
		var labelsJSON string
		var lastSeenAt sql.NullTime

		err := rows.Scan(
			&runner.ID,
			&runner.Name,
			&runner.URL,
			&labelsJSON,
			&runner.Status,
			&lastSeenAt,
			&runner.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		if lastSeenAt.Valid {
			runner.LastSeenAt = lastSeenAt.Time
		}

		if err := json.Unmarshal([]byte(labelsJSON), &runner.Labels); err != nil {
			runner.Labels = make(map[string]string)
		}

		runners = append(runners, runner)
	}

	return runners, nil
}

// UpdateRunnerStatus met à jour le statut d'un runner
func UpdateRunnerStatus(db *sql.DB, id int, status string) error {
	query := `
		UPDATE runners
		SET status = ?, last_seen_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.Exec(query, status, id)
	return err
}

// UpdateRunnerLabels met à jour les labels d'un runner
func UpdateRunnerLabels(db *sql.DB, id int, labels map[string]string) error {
	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return err
	}

	query := `
		UPDATE runners
		SET labels = ?
		WHERE id = ?
	`

	_, err = db.Exec(query, string(labelsJSON), id)
	return err
}

// UpdateRunnerURL met à jour l'URL d'un runner
func UpdateRunnerURL(db *sql.DB, id int, url string) error {
	query := `
		UPDATE runners
		SET url = ?
		WHERE id = ?
	`

	_, err := db.Exec(query, url, id)
	return err
}

// UpdateRunnerHeartbeat met à jour le last_seen_at d'un runner (heartbeat)
func UpdateRunnerHeartbeat(db *sql.DB, id int) error {
	query := `
		UPDATE runners
		SET last_seen_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.Exec(query, id)
	return err
}

// DeleteRunner supprime un runner
func DeleteRunner(db *sql.DB, id int) error {
	query := `DELETE FROM runners WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

// MarkStaleRunnersOffline marque comme OFFLINE les runners qui n'ont pas envoyé de heartbeat
// depuis plus de timeoutDuration
func MarkStaleRunnersOffline(db *sql.DB, timeoutDuration time.Duration) (int, error) {
	timeoutThreshold := time.Now().Add(-timeoutDuration)

	query := `
		UPDATE runners
		SET status = 'OFFLINE'
		WHERE status = 'ONLINE'
		AND last_seen_at < ?
	`

	result, err := db.Exec(query, timeoutThreshold)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}
