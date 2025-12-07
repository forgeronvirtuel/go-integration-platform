package database

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProjectsTable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = CreateProjectsTable(db)
	require.NoError(t, err, "CreateProjectsTable ne devrait pas retourner d'erreur")

	// Vérifier que la table existe
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='projects'").Scan(&tableName)
	assert.NoError(t, err, "La table projects devrait exister")
	assert.Equal(t, "projects", tableName)
}

func TestCreateProject(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = CreateProjectsTable(db)
	require.NoError(t, err)

	project, err := CreateProject(db, "api-users", "https://github.com/user/api-users.git", "main", "")
	require.NoError(t, err, "CreateProject ne devrait pas retourner d'erreur")
	require.NotNil(t, project)

	assert.Equal(t, "api-users", project.Name)
	assert.Equal(t, "https://github.com/user/api-users.git", project.RepoURL)
	assert.Equal(t, "main", project.Branch)
	assert.Equal(t, "", project.Subdir)
	assert.NotZero(t, project.ID)
	assert.NotZero(t, project.CreatedAt)
	assert.NotZero(t, project.UpdatedAt)
}

func TestCreateProjectWithSubdir(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = CreateProjectsTable(db)
	require.NoError(t, err)

	project, err := CreateProject(db, "api-orders", "https://github.com/user/monorepo.git", "develop", "services/api")
	require.NoError(t, err)

	assert.Equal(t, "api-orders", project.Name)
	assert.Equal(t, "develop", project.Branch)
	assert.Equal(t, "services/api", project.Subdir)
}

func TestCreateProjectDefaultBranch(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = CreateProjectsTable(db)
	require.NoError(t, err)

	// Créer un projet sans spécifier de branche
	project, err := CreateProject(db, "test-project", "https://github.com/user/test.git", "", "")
	require.NoError(t, err)

	assert.Equal(t, "main", project.Branch, "La branche par défaut devrait être 'main'")
}

func TestGetProjectByID(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = CreateProjectsTable(db)
	require.NoError(t, err)

	created, err := CreateProject(db, "test-project", "https://github.com/user/test.git", "main", "")
	require.NoError(t, err)

	retrieved, err := GetProjectByID(db, created.ID)
	require.NoError(t, err)

	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Name, retrieved.Name)
	assert.Equal(t, created.RepoURL, retrieved.RepoURL)
}

func TestGetProjectByName(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = CreateProjectsTable(db)
	require.NoError(t, err)

	created, err := CreateProject(db, "api-users", "https://github.com/user/api-users.git", "main", "")
	require.NoError(t, err)

	retrieved, err := GetProjectByName(db, "api-users")
	require.NoError(t, err)

	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, "api-users", retrieved.Name)
}

func TestGetAllProjects(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = CreateProjectsTable(db)
	require.NoError(t, err)

	// Créer plusieurs projets
	_, err = CreateProject(db, "project1", "https://github.com/user/p1.git", "main", "")
	require.NoError(t, err)
	_, err = CreateProject(db, "project2", "https://github.com/user/p2.git", "develop", "services/api")
	require.NoError(t, err)
	_, err = CreateProject(db, "project3", "https://github.com/user/p3.git", "main", "")
	require.NoError(t, err)

	projects, err := GetAllProjects(db)
	require.NoError(t, err)

	assert.Len(t, projects, 3, "Il devrait y avoir 3 projets")

	// Vérifier que tous les projets sont présents
	projectNames := make(map[string]bool)
	for _, p := range projects {
		projectNames[p.Name] = true
	}
	assert.True(t, projectNames["project1"])
	assert.True(t, projectNames["project2"])
	assert.True(t, projectNames["project3"])
}

func TestUpdateProject(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = CreateProjectsTable(db)
	require.NoError(t, err)

	created, err := CreateProject(db, "api-users", "https://github.com/user/api-users.git", "main", "")
	require.NoError(t, err)

	// Mettre à jour le projet
	updated, err := UpdateProject(db, created.ID, "api-users-v2", "https://github.com/user/api-users-v2.git", "develop", "api")
	require.NoError(t, err)

	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "api-users-v2", updated.Name)
	assert.Equal(t, "https://github.com/user/api-users-v2.git", updated.RepoURL)
	assert.Equal(t, "develop", updated.Branch)
	assert.Equal(t, "api", updated.Subdir)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt) || updated.UpdatedAt.Equal(created.UpdatedAt))
}

func TestDeleteProject(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = CreateProjectsTable(db)
	require.NoError(t, err)

	created, err := CreateProject(db, "api-users", "https://github.com/user/api-users.git", "main", "")
	require.NoError(t, err)

	// Supprimer le projet
	err = DeleteProject(db, created.ID)
	require.NoError(t, err)

	// Vérifier que le projet n'existe plus
	_, err = GetProjectByID(db, created.ID)
	assert.Error(t, err, "Le projet devrait avoir été supprimé")
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestUniqueProjectName(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = CreateProjectsTable(db)
	require.NoError(t, err)

	// Créer le premier projet
	_, err = CreateProject(db, "api-users", "https://github.com/user/api-users.git", "main", "")
	require.NoError(t, err)

	// Tenter de créer un projet avec le même nom
	_, err = CreateProject(db, "api-users", "https://github.com/user/another.git", "main", "")
	assert.Error(t, err, "La création d'un projet avec un nom déjà existant devrait échouer")
}
