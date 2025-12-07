package server

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Erreur lors de l'ouverture de la base de données: %v", err)
	}

	// Créer la table users pour les tests
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("Erreur lors de la création de la table: %v", err)
	}

	// Insérer des données de test
	insertQuery := `
	INSERT INTO users (name, email) VALUES
		('Test User', 'test@example.com'),
		('John Doe', 'john@example.com');
	`
	if _, err := db.Exec(insertQuery); err != nil {
		t.Fatalf("Erreur lors de l'insertion des données de test: %v", err)
	}

	return db
}

func setupTestRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return SetupRouter(db)
}

func TestHelloWorldRoute(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := setupTestRouter(db)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Hello World!")
	assert.Contains(t, w.Body.String(), "ok")
}

func TestHealthRoute(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := setupTestRouter(db)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
	assert.Contains(t, w.Body.String(), "connected")
}

func TestUsersRoute(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	router := setupTestRouter(db)

	req, _ := http.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test User")
	assert.Contains(t, w.Body.String(), "test@example.com")
	assert.Contains(t, w.Body.String(), "John Doe")
}

func TestHealthRouteWithClosedDB(t *testing.T) {
	db := setupTestDB(t)
	db.Close() // Fermer la DB pour simuler une erreur

	router := setupTestRouter(db)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "database unavailable")
}
