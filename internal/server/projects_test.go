package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"forgeronvirtuel/gip/internal/database"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProjectTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	err = database.CreateProjectsTable(db)
	require.NoError(t, err)

	return db
}

func TestCreateProjectEndpoint(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := SetupRouter(db, "")

	reqBody := CreateProjectRequest{
		Name:    "api-users",
		RepoURL: "https://github.com/user/api-users.git",
		Branch:  "main",
		Subdir:  "",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response database.Project
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "api-users", response.Name)
	assert.Equal(t, "https://github.com/user/api-users.git", response.RepoURL)
	assert.Equal(t, "main", response.Branch)
	assert.NotZero(t, response.ID)
}

func TestCreateProjectWithDefaultBranch(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := SetupRouter(db, "")

	reqBody := CreateProjectRequest{
		Name:    "test-project",
		RepoURL: "https://github.com/user/test.git",
		// Branch non spécifié
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response database.Project
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "main", response.Branch, "La branche par défaut devrait être 'main'")
}

func TestGetAllProjectsEndpoint(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	// Créer quelques projets de test
	database.CreateProject(db, "project1", "https://github.com/user/p1.git", "main", "")
	database.CreateProject(db, "project2", "https://github.com/user/p2.git", "develop", "api")

	gin.SetMode(gin.TestMode)
	router := SetupRouter(db, "")

	req, _ := http.NewRequest("GET", "/api/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	projects := response["projects"].([]interface{})
	assert.Len(t, projects, 2)
	assert.Equal(t, float64(2), response["count"])
}

func TestGetProjectByIDEndpoint(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	created, _ := database.CreateProject(db, "api-users", "https://github.com/user/api-users.git", "main", "")

	gin.SetMode(gin.TestMode)
	router := SetupRouter(db, "")

	req, _ := http.NewRequest("GET", "/api/projects/"+string(rune(created.ID+'0')), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response database.Project
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, created.ID, response.ID)
	assert.Equal(t, "api-users", response.Name)
}

func TestGetProjectByNameEndpoint(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	database.CreateProject(db, "api-users", "https://github.com/user/api-users.git", "main", "")

	gin.SetMode(gin.TestMode)
	router := SetupRouter(db, "")

	req, _ := http.NewRequest("GET", "/api/projects/by-name/api-users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response database.Project
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "api-users", response.Name)
}

func TestUpdateProjectEndpoint(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	created, _ := database.CreateProject(db, "api-users", "https://github.com/user/api-users.git", "main", "")

	gin.SetMode(gin.TestMode)
	router := SetupRouter(db, "")

	reqBody := UpdateProjectRequest{
		Name:    "api-users-v2",
		RepoURL: "https://github.com/user/api-users-v2.git",
		Branch:  "develop",
		Subdir:  "services/api",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/projects/"+string(rune(created.ID+'0')), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response database.Project
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "api-users-v2", response.Name)
	assert.Equal(t, "develop", response.Branch)
	assert.Equal(t, "services/api", response.Subdir)
}

func TestDeleteProjectEndpoint(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	created, _ := database.CreateProject(db, "api-users", "https://github.com/user/api-users.git", "main", "")

	gin.SetMode(gin.TestMode)
	router := SetupRouter(db, "")

	req, _ := http.NewRequest("DELETE", "/api/projects/"+string(rune(created.ID+'0')), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Vérifier que le projet a bien été supprimé
	_, err := database.GetProjectByID(db, created.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestGetProjectNotFound(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := SetupRouter(db, "")

	req, _ := http.NewRequest("GET", "/api/projects/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "project not found", response["error"])
}

func TestCreateProjectInvalidRequest(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := SetupRouter(db, "")

	// Request sans nom (champ requis)
	reqBody := map[string]string{
		"repo_url": "https://github.com/user/test.git",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProjectNotFound(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := SetupRouter(db, "")

	reqBody := UpdateProjectRequest{
		Name:    "test",
		RepoURL: "https://github.com/user/test.git",
		Branch:  "main",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/api/projects/999", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteProjectNotFound(t *testing.T) {
	db := setupProjectTestDB(t)
	defer db.Close()

	gin.SetMode(gin.TestMode)
	router := SetupRouter(db, "")

	req, _ := http.NewRequest("DELETE", "/api/projects/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
