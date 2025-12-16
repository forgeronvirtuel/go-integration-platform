package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
)

// Runner représente un agent de build
type Runner struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	URL        string            `json:"url"` // URL pour contacter le runner
	Labels     map[string]string `json:"labels"`
	Status     string            `json:"status"` // ONLINE, OFFLINE, DRAINING
	LastSeenAt time.Time         `json:"last_seen_at"`
	CreatedAt  time.Time         `json:"created_at"`
}

// CreateAgentsTable crée la table agents si elle n'existe pas
func CreateAgentsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS agents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL DEFAULT '',
		labels TEXT NOT NULL DEFAULT '{}',
		status TEXT NOT NULL DEFAULT 'OFFLINE' CHECK(status IN ('ONLINE', 'OFFLINE', 'DRAINING')),
		last_seen_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
	CREATE INDEX IF NOT EXISTS idx_agents_name ON agents(name);
	`

	if _, err := db.Exec(query); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la création de la table agents")
		return err
	}

	log.Info().Msg("Table 'agents' créée ou déjà existante")
	return nil
}

// CreateAgent crée un nouvel agent
func CreateAgent(db *sql.DB, name string, labels map[string]string) (*Runner, error) {
	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO agents (name, url, labels, status, last_seen_at)
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

	return GetAgentByID(db, int(id))
}

// GetAgentByID récupère un agent par son ID
func GetAgentByID(db *sql.DB, id int) (*Runner, error) {
	query := `
		SELECT id, name, url, labels, status, last_seen_at, created_at
		FROM agents
		WHERE id = ?
	`

	var agent Runner
	var labelsJSON string
	var lastSeenAt sql.NullTime

	err := db.QueryRow(query, id).Scan(
		&agent.ID,
		&agent.Name,
		&agent.URL,
		&labelsJSON,
		&agent.Status,
		&lastSeenAt,
		&agent.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if lastSeenAt.Valid {
		agent.LastSeenAt = lastSeenAt.Time
	}

	if err := json.Unmarshal([]byte(labelsJSON), &agent.Labels); err != nil {
		agent.Labels = make(map[string]string)
	}

	return &agent, nil
}

// GetRunnerByName récupère un agent par son nom (hostname)
func GetRunnerByName(db *sql.DB, name string) (*Runner, error) {
	query := `
		SELECT id, name, url, labels, status, last_seen_at, created_at
		FROM agents
		WHERE name = ?
	`

	var agent Runner
	var labelsJSON string
	var lastSeenAt sql.NullTime

	err := db.QueryRow(query, name).Scan(
		&agent.ID,
		&agent.Name,
		&agent.URL,
		&labelsJSON,
		&agent.Status,
		&lastSeenAt,
		&agent.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if lastSeenAt.Valid {
		agent.LastSeenAt = lastSeenAt.Time
	}

	if err := json.Unmarshal([]byte(labelsJSON), &agent.Labels); err != nil {
		agent.Labels = make(map[string]string)
	}

	return &agent, nil
}

func GetAllRunners(db *sql.DB) ([]Runner, error) {
	query := `
		SELECT id, name, url, labels, status, last_seen_at, created_at
		FROM agents
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

// GetAgentsByStatus récupère les agents par statut
func GetAgentsByStatus(db *sql.DB, status string) ([]Runner, error) {
	query := `
		SELECT id, name, url, labels, status, last_seen_at, created_at
		FROM agents
		WHERE status = ?
		ORDER BY last_seen_at DESC
	`

	rows, err := db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []Runner
	for rows.Next() {
		var agent Runner
		var labelsJSON string
		var lastSeenAt sql.NullTime

		err := rows.Scan(
			&agent.ID,
			&agent.Name,
			&agent.URL,
			&labelsJSON,
			&agent.Status,
			&lastSeenAt,
			&agent.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		if lastSeenAt.Valid {
			agent.LastSeenAt = lastSeenAt.Time
		}

		if err := json.Unmarshal([]byte(labelsJSON), &agent.Labels); err != nil {
			agent.Labels = make(map[string]string)
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

// UpdateAgentStatus met à jour le statut d'un agent
func UpdateAgentStatus(db *sql.DB, id int, status string) error {
	query := `
		UPDATE agents
		SET status = ?, last_seen_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.Exec(query, status, id)
	return err
}

// UpdateAgentLabels met à jour les labels d'un agent
func UpdateAgentLabels(db *sql.DB, id int, labels map[string]string) error {
	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return err
	}

	query := `
		UPDATE agents
		SET labels = ?
		WHERE id = ?
	`

	_, err = db.Exec(query, string(labelsJSON), id)
	return err
}

// UpdateAgentURL met à jour l'URL d'un agent
func UpdateAgentURL(db *sql.DB, id int, url string) error {
	query := `
		UPDATE agents
		SET url = ?
		WHERE id = ?
	`

	_, err := db.Exec(query, url, id)
	return err
}

// UpdateAgentHeartbeat met à jour le last_seen_at d'un agent (heartbeat)
func UpdateAgentHeartbeat(db *sql.DB, id int) error {
	query := `
		UPDATE agents
		SET last_seen_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := db.Exec(query, id)
	return err
}

// DeleteAgent supprime un agent
func DeleteAgent(db *sql.DB, id int) error {
	query := `DELETE FROM agents WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

// MarkStaleAgentsOffline marque comme OFFLINE les agents qui n'ont pas envoyé de heartbeat
// depuis plus de timeoutDuration
func MarkStaleAgentsOffline(db *sql.DB, timeoutDuration time.Duration) (int, error) {
	timeoutThreshold := time.Now().Add(-timeoutDuration)

	query := `
		UPDATE agents
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
