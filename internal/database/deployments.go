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
	RunnerID  *int       `json:"runner_id,omitempty"`
	LogOutput string     `json:"log_output"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// CreateDeploymentsTable creates the deployments table if it does not exist
func CreateDeploymentsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS deployments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		build_id INTEGER NOT NULL,
		status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'deploying', 'deployed', 'failed')),
		runner_id INTEGER,
		log_output TEXT,
		started_at DATETIME,
		ended_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE,
		FOREIGN KEY (runner_id) REFERENCES runners(id) ON DELETE SET NULL
	);
	CREATE INDEX IF NOT EXISTS idx_deployments_build_id ON deployments(build_id);
	CREATE INDEX IF NOT EXISTS idx_deployments_runner_id ON deployments(runner_id);
	CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	log.Info().Msg("Table 'deployments' créée ou déjà existante")
	return nil
}

// CreateDeployment creates a new deployment
func CreateDeployment(db *sql.DB, buildID int, RunnerID *int) (*Deployment, error) {
	startedAt := time.Now()
	result, err := db.Exec(
		"INSERT INTO deployments (build_id, runner_id, status, started_at) VALUES (?, ?, ?, ?)",
		buildID, RunnerID, "pending", startedAt,
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
		RunnerID:  RunnerID,
		Status:    "pending",
		StartedAt: startedAt,
		CreatedAt: startedAt,
	}

	log.Info().
		Int("deployment_id", deployment.ID).
		Int("build_id", buildID).
		Int("runner_id", *RunnerID).
		Msg("Deployment created successfully")

	return deployment, nil
}

// GetDeploymentByID retrieves a deployment by its ID
func GetDeploymentByID(db *sql.DB, id int) (*Deployment, error) {
	deployment := &Deployment{}
	var endedAt sql.NullTime
	var runnerID sql.NullInt64
	var logOutput sql.NullString

	err := db.QueryRow(
		"SELECT id, build_id, status, runner_id, log_output, started_at, ended_at, created_at FROM deployments WHERE id = ?",
		id,
	).Scan(&deployment.ID, &deployment.BuildID, &deployment.Status, &runnerID, &logOutput, &deployment.StartedAt, &endedAt, &deployment.CreatedAt)

	if err != nil {
		return nil, err
	}

	if endedAt.Valid {
		deployment.EndedAt = &endedAt.Time
	}

	if runnerID.Valid {
		runnerIDInt := int(runnerID.Int64)
		deployment.RunnerID = &runnerIDInt
	}

	if logOutput.Valid {
		deployment.LogOutput = logOutput.String
	}

	return deployment, nil
}

// GetAllDeployments retrieves all deployments
func GetAllDeployments(db *sql.DB) ([]Deployment, error) {
	rows, err := db.Query(
		"SELECT id, build_id, status, runner_id, log_output, started_at, ended_at, created_at FROM deployments ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []Deployment
	for rows.Next() {
		var deployment Deployment
		var endedAt sql.NullTime
		var runnerID sql.NullInt64
		var logOutput sql.NullString

		err := rows.Scan(&deployment.ID, &deployment.BuildID, &deployment.Status, &runnerID, &logOutput, &deployment.StartedAt, &endedAt, &deployment.CreatedAt)
		if err != nil {
			return nil, err
		}

		if endedAt.Valid {
			deployment.EndedAt = &endedAt.Time
		}

		if runnerID.Valid {
			runnerIDInt := int(runnerID.Int64)
			deployment.RunnerID = &runnerIDInt
		}

		if logOutput.Valid {
			deployment.LogOutput = logOutput.String
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// GetDeploymentsByBuildID get all deployments for a build
func GetDeploymentsByBuildID(db *sql.DB, buildID int) ([]Deployment, error) {
	rows, err := db.Query(
		"SELECT id, build_id, status, runner_id, log_output, started_at, ended_at, created_at FROM deployments WHERE build_id = ? ORDER BY created_at DESC",
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
		var runnerID sql.NullInt64
		var logOutput sql.NullString

		err := rows.Scan(&deployment.ID, &deployment.BuildID, &deployment.Status, &runnerID, &logOutput, &deployment.StartedAt, &endedAt, &deployment.CreatedAt)
		if err != nil {
			return nil, err
		}

		if endedAt.Valid {
			deployment.EndedAt = &endedAt.Time
		}

		if runnerID.Valid {
			runnerIDInt := int(runnerID.Int64)
			deployment.RunnerID = &runnerIDInt
		}

		if logOutput.Valid {
			deployment.LogOutput = logOutput.String
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// GetDeploymentsByRunnerID gets all deployments for an runner
func GetDeploymentsByRunnerID(db *sql.DB, runnerID int) ([]Deployment, error) {
	rows, err := db.Query(
		"SELECT id, build_id, status, runner_id, log_output, started_at, ended_at, created_at FROM deployments WHERE runner_id = ? ORDER BY created_at DESC",
		runnerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []Deployment
	for rows.Next() {
		var deployment Deployment
		var endedAt sql.NullTime
		var runnerID sql.NullInt64
		var logOutput sql.NullString

		err := rows.Scan(&deployment.ID, &deployment.BuildID, &deployment.Status, &runnerID, &logOutput, &deployment.StartedAt, &endedAt, &deployment.CreatedAt)
		if err != nil {
			return nil, err
		}

		if endedAt.Valid {
			deployment.EndedAt = &endedAt.Time
		}

		if runnerID.Valid {
			runnerIDInt := int(runnerID.Int64)
			deployment.RunnerID = &runnerIDInt
		}

		if logOutput.Valid {
			deployment.LogOutput = logOutput.String
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

// UpdateDeploymentStatus updates the status of a deployment
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
		Msg("Deployment status updated")

	return GetDeploymentByID(db, id)
}

// UpdateDeploymentLog updates the log output of a deployment
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
		Msg("Deployment logs updated")

	return nil
}

// UpdateDeploymentRunner updates the runner assigned to a deployment
func UpdateDeploymentRunner(db *sql.DB, id int, runnerID *int) (*Deployment, error) {
	_, err := db.Exec(
		"UPDATE deployments SET runner_id = ? WHERE id = ?",
		runnerID, id,
	)
	if err != nil {
		return nil, err
	}

	log.Info().
		Int("deployment_id", id).
		Msg("Deployment runner updated")

	return GetDeploymentByID(db, id)
}

// DeleteDeployment deletes a deployment
func DeleteDeployment(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM deployments WHERE id = ?", id)
	if err != nil {
		return err
	}

	log.Info().
		Int("deployment_id", id).
		Msg("Deployment deleted successfully")

	return nil
}
