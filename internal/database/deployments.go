package database

import (
	"database/sql"
	"time"

	"github.com/rs/zerolog/log"
)

type Deployment struct {
	ID        int        `json:"id"`
	BuildID   int        `json:"build_id"`
	Status    string     `json:"status"` // pending, deploying, deployed, failed
	AgentID   *int       `json:"agent_id,omitempty"`
	LogOutput string     `json:"log_output"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// CreateDeploymentsTable crée la table deployments si elle n'existe pas
func CreateDeploymentsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS deployments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		build_id INTEGER NOT NULL,
		status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'deploying', 'deployed', 'failed')),
		agent_id INTEGER,
		log_output TEXT,
		started_at DATETIME,
		ended_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE,
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE SET NULL
	);
	CREATE INDEX IF NOT EXISTS idx_deployments_build_id ON deployments(build_id);
	CREATE INDEX IF NOT EXISTS idx_deployments_agent_id ON deployments(agent_id);
	CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	log.Info().Msg("Table 'deployments' créée ou déjà existante")
	return nil
}

// CreateDeployment crée un nouveau déploiement
func CreateDeployment(db *sql.DB, buildID int, agentID *int) (*Deployment, error) {
	startedAt := time.Now()
	result, err := db.Exec(
		"INSERT INTO deployments (build_id, agent_id, status, started_at) VALUES (?, ?, ?, ?)",
		buildID, agentID, "pending", startedAt,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	deployment := &Deployment{
		ID:        int(id),
		BuildID:   buildID,
		AgentID:   agentID,
		Status:    "pending",
		StartedAt: startedAt,
		CreatedAt: startedAt,
	}

	log.Info().
		Int("deployment_id", deployment.ID).
		Int("build_id", buildID).
		Msg("Déploiement créé")

	return deployment, nil
}

// GetDeploymentByID récupère un déploiement par son ID
func GetDeploymentByID(db *sql.DB, id int) (*Deployment, error) {
	deployment := &Deployment{}
	var endedAt sql.NullTime
	var agentID sql.NullInt64
	var logOutput sql.NullString

	err := db.QueryRow(
		"SELECT id, build_id, status, agent_id, log_output, started_at, ended_at, created_at FROM deployments WHERE id = ?",
		id,
	).Scan(&deployment.ID, &deployment.BuildID, &deployment.Status, &agentID, &logOutput, &deployment.StartedAt, &endedAt, &deployment.CreatedAt)

	if err != nil {
		return nil, err
	}

	if endedAt.Valid {
		deployment.EndedAt = &endedAt.Time
	}

	if agentID.Valid {
		agentIDInt := int(agentID.Int64)
		deployment.AgentID = &agentIDInt
	}

	if logOutput.Valid {
		deployment.LogOutput = logOutput.String
	}

	return deployment, nil
}

// GetAllDeployments récupère tous les déploiements
func GetAllDeployments(db *sql.DB) ([]Deployment, error) {
	rows, err := db.Query(
		"SELECT id, build_id, status, agent_id, log_output, started_at, ended_at, created_at FROM deployments ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []Deployment
	for rows.Next() {
		var deployment Deployment
		var endedAt sql.NullTime
		var agentID sql.NullInt64
		var logOutput sql.NullString

		err := rows.Scan(&deployment.ID, &deployment.BuildID, &deployment.Status, &agentID, &logOutput, &deployment.StartedAt, &endedAt, &deployment.CreatedAt)
		if err != nil {
			return nil, err
		}

		if endedAt.Valid {
			deployment.EndedAt = &endedAt.Time
		}

		if agentID.Valid {
			agentIDInt := int(agentID.Int64)
			deployment.AgentID = &agentIDInt
		}

		if logOutput.Valid {
			deployment.LogOutput = logOutput.String
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// GetDeploymentsByBuildID récupère tous les déploiements pour un build donné
func GetDeploymentsByBuildID(db *sql.DB, buildID int) ([]Deployment, error) {
	rows, err := db.Query(
		"SELECT id, build_id, status, agent_id, log_output, started_at, ended_at, created_at FROM deployments WHERE build_id = ? ORDER BY created_at DESC",
		buildID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []Deployment
	for rows.Next() {
		var deployment Deployment
		var endedAt sql.NullTime
		var agentID sql.NullInt64
		var logOutput sql.NullString

		err := rows.Scan(&deployment.ID, &deployment.BuildID, &deployment.Status, &agentID, &logOutput, &deployment.StartedAt, &endedAt, &deployment.CreatedAt)
		if err != nil {
			return nil, err
		}

		if endedAt.Valid {
			deployment.EndedAt = &endedAt.Time
		}

		if agentID.Valid {
			agentIDInt := int(agentID.Int64)
			deployment.AgentID = &agentIDInt
		}

		if logOutput.Valid {
			deployment.LogOutput = logOutput.String
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// GetDeploymentsByAgentID récupère tous les déploiements pour un agent donné
func GetDeploymentsByAgentID(db *sql.DB, agentID int) ([]Deployment, error) {
	rows, err := db.Query(
		"SELECT id, build_id, status, agent_id, log_output, started_at, ended_at, created_at FROM deployments WHERE agent_id = ? ORDER BY created_at DESC",
		agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []Deployment
	for rows.Next() {
		var deployment Deployment
		var endedAt sql.NullTime
		var agentID sql.NullInt64
		var logOutput sql.NullString

		err := rows.Scan(&deployment.ID, &deployment.BuildID, &deployment.Status, &agentID, &logOutput, &deployment.StartedAt, &endedAt, &deployment.CreatedAt)
		if err != nil {
			return nil, err
		}

		if endedAt.Valid {
			deployment.EndedAt = &endedAt.Time
		}

		if agentID.Valid {
			agentIDInt := int(agentID.Int64)
			deployment.AgentID = &agentIDInt
		}

		if logOutput.Valid {
			deployment.LogOutput = logOutput.String
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// UpdateDeploymentStatus met à jour le statut d'un déploiement
func UpdateDeploymentStatus(db *sql.DB, id int, status string) (*Deployment, error) {
	var endedAt *time.Time
	if status == "deployed" || status == "failed" {
		now := time.Now()
		endedAt = &now
	}

	_, err := db.Exec(
		"UPDATE deployments SET status = ?, ended_at = ? WHERE id = ?",
		status, endedAt, id,
	)
	if err != nil {
		return nil, err
	}

	log.Info().
		Int("deployment_id", id).
		Str("status", status).
		Msg("Statut du déploiement mis à jour")

	return GetDeploymentByID(db, id)
}

// UpdateDeploymentLog met à jour les logs d'un déploiement
func UpdateDeploymentLog(db *sql.DB, id int, logOutput string) error {
	_, err := db.Exec(
		"UPDATE deployments SET log_output = ? WHERE id = ?",
		logOutput, id,
	)
	if err != nil {
		return err
	}

	log.Debug().
		Int("deployment_id", id).
		Msg("Logs du déploiement mis à jour")

	return nil
}

// UpdateDeploymentAgent met à jour l'agent assigné à un déploiement
func UpdateDeploymentAgent(db *sql.DB, id int, agentID *int) (*Deployment, error) {
	_, err := db.Exec(
		"UPDATE deployments SET agent_id = ? WHERE id = ?",
		agentID, id,
	)
	if err != nil {
		return nil, err
	}

	log.Info().
		Int("deployment_id", id).
		Msg("Agent du déploiement mis à jour")

	return GetDeploymentByID(db, id)
}

// DeleteDeployment supprime un déploiement
func DeleteDeployment(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM deployments WHERE id = ?", id)
	if err != nil {
		return err
	}

	log.Info().
		Int("deployment_id", id).
		Msg("Déploiement supprimé")

	return nil
}
