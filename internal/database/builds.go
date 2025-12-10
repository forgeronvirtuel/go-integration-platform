package database

import (
	"database/sql"
	"time"

	"github.com/rs/zerolog/log"
)

type Build struct {
	ID        int
	ProjectID int
	Branch    string
	Status    string
	LogOutput string
	StartedAt time.Time
	EndedAt   sql.NullTime
	CreatedAt time.Time
}

// CreateBuildsTable crée la table builds si elle n'existe pas
func CreateBuildsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS builds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		branch TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		log_output TEXT,
		started_at DATETIME,
		ended_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_builds_project_id ON builds(project_id);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	log.Info().Msg("Table 'builds' créée ou déjà existante")
	return nil
}

// CreateBuild crée un nouveau build
func CreateBuild(db *sql.DB, projectID int, branch string) (*Build, error) {
	startedAt := time.Now()
	result, err := db.Exec(
		"INSERT INTO builds (project_id, branch, status, started_at) VALUES (?, ?, ?, ?)",
		projectID, branch, "pending", startedAt,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Build{
		ID:        int(id),
		ProjectID: projectID,
		Branch:    branch,
		Status:    "pending",
		StartedAt: startedAt,
	}, nil
}

// GetBuildByID récupère un build par son ID
func GetBuildByID(db *sql.DB, id string) (*Build, error) {
	build := &Build{}
	err := db.QueryRow(
		"SELECT id, project_id, branch, status, log_output, started_at, ended_at, created_at FROM builds WHERE id = ?",
		id,
	).Scan(&build.ID, &build.ProjectID, &build.Branch, &build.Status, &build.LogOutput, &build.StartedAt, &build.EndedAt, &build.CreatedAt)

	if err != nil {
		return nil, err
	}

	return build, nil
}

// UpdateBuildStatus met à jour le statut d'un build
func UpdateBuildStatus(db *sql.DB, id int, status string, logOutput string) error {
	var err error
	if status == "success" || status == "failed" {
		// Si le build est terminé, on met à jour ended_at
		_, err = db.Exec(
			"UPDATE builds SET status = ?, log_output = ?, ended_at = CURRENT_TIMESTAMP WHERE id = ?",
			status, logOutput, id,
		)
	} else {
		// Sinon on met juste à jour le statut et les logs
		_, err = db.Exec(
			"UPDATE builds SET status = ?, log_output = ? WHERE id = ?",
			status, logOutput, id,
		)
	}

	if err != nil {
		return err
	}

	log.Info().Int("id", id).Str("status", status).Msg("Statut du build mis à jour")
	return nil
}

// GetAllBuilds récupère tous les builds
func GetAllBuilds(db *sql.DB) ([]Build, error) {
	rows, err := db.Query(
		"SELECT id, project_id, branch, status, log_output, started_at, ended_at, created_at FROM builds ORDER BY id DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var builds []Build
	for rows.Next() {
		var build Build
		err := rows.Scan(&build.ID, &build.ProjectID, &build.Branch, &build.Status, &build.LogOutput, &build.StartedAt, &build.EndedAt, &build.CreatedAt)
		if err != nil {
			return nil, err
		}
		builds = append(builds, build)
	}

	return builds, nil
}

// GetBuildsByProjectID récupère tous les builds d'un projet
func GetBuildsByProjectID(db *sql.DB, projectID int) ([]Build, error) {
	rows, err := db.Query(
		"SELECT id, project_id, branch, status, log_output, started_at, ended_at, created_at FROM builds WHERE project_id = ? ORDER BY id DESC",
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var builds []Build
	for rows.Next() {
		var build Build
		err := rows.Scan(&build.ID, &build.ProjectID, &build.Branch, &build.Status, &build.LogOutput, &build.StartedAt, &build.EndedAt, &build.CreatedAt)
		if err != nil {
			return nil, err
		}
		builds = append(builds, build)
	}

	return builds, nil
}

// DeleteBuild supprime un build
func DeleteBuild(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM builds WHERE id = ?", id)
	if err != nil {
		return err
	}

	log.Info().Int("id", id).Msg("Build supprimé avec succès")
	return nil
}
