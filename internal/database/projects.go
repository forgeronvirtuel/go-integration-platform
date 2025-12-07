package database

import (
	"database/sql"
	"time"
)

// Project représente un projet Go déployable
type Project struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	RepoURL   string    `json:"repo_url"`
	Branch    string    `json:"branch"`
	Subdir    string    `json:"subdir"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateProjectsTable crée la table projects dans la base de données
func CreateProjectsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		repo_url TEXT NOT NULL,
		branch TEXT NOT NULL DEFAULT 'main',
		subdir TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(query); err != nil {
		return err
	}

	return nil
}

// CreateProject insère un nouveau projet dans la base de données
func CreateProject(db *sql.DB, name, repoURL, branch, subdir string) (*Project, error) {
	if branch == "" {
		branch = "main"
	}

	query := `
	INSERT INTO projects (name, repo_url, branch, subdir)
	VALUES (?, ?, ?, ?)
	`

	result, err := db.Exec(query, name, repoURL, branch, subdir)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetProjectByID(db, int(id))
}

// GetProjectByID récupère un projet par son ID
func GetProjectByID(db *sql.DB, id int) (*Project, error) {
	query := `
	SELECT id, name, repo_url, branch, subdir, created_at, updated_at
	FROM projects
	WHERE id = ?
	`

	project := &Project{}
	err := db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.RepoURL,
		&project.Branch,
		&project.Subdir,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return project, nil
}

// GetProjectByName récupère un projet par son nom
func GetProjectByName(db *sql.DB, name string) (*Project, error) {
	query := `
	SELECT id, name, repo_url, branch, subdir, created_at, updated_at
	FROM projects
	WHERE name = ?
	`

	project := &Project{}
	err := db.QueryRow(query, name).Scan(
		&project.ID,
		&project.Name,
		&project.RepoURL,
		&project.Branch,
		&project.Subdir,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return project, nil
}

// GetAllProjects récupère tous les projets
func GetAllProjects(db *sql.DB) ([]*Project, error) {
	query := `
	SELECT id, name, repo_url, branch, subdir, created_at, updated_at
	FROM projects
	ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []*Project{}
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.RepoURL,
			&project.Branch,
			&project.Subdir,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

// UpdateProject met à jour un projet existant
func UpdateProject(db *sql.DB, id int, name, repoURL, branch, subdir string) (*Project, error) {
	query := `
	UPDATE projects
	SET name = ?, repo_url = ?, branch = ?, subdir = ?, updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`

	_, err := db.Exec(query, name, repoURL, branch, subdir, id)
	if err != nil {
		return nil, err
	}

	return GetProjectByID(db, id)
}

// DeleteProject supprime un projet
func DeleteProject(db *sql.DB, id int) error {
	query := `DELETE FROM projects WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}
