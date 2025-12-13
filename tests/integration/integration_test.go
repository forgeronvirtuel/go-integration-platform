package integration

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"forgeronvirtuel/gip/internal/database"
	"forgeronvirtuel/gip/internal/server"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPort   = "9999"
	testDBPath = "./test_integration.db"
	baseURL    = "http://localhost:9999/v1/"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Setup: Nettoyer les anciennes bases de test
	os.Remove(testDBPath)

	// Initialiser la base de données
	var err error
	testDB, err = database.InitDB(testDBPath)
	if err != nil {
		fmt.Printf("Erreur lors de l'initialisation de la DB: %v\n", err)
		os.Exit(1)
	}

	// Démarrer le serveur dans une goroutine
	go func() {
		server.Start(testPort, testDB, "./test-workspace")
	}()

	// Attendre que le serveur démarre
	time.Sleep(2 * time.Second)

	// Vérifier que le serveur est bien démarré
	for i := 0; i < 10; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			break
		}
		if i == 9 {
			fmt.Println("Le serveur n'a pas démarré à temps")
			os.Exit(1)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Exécuter les tests
	code := m.Run()

	// Cleanup
	testDB.Close()
	os.Remove(testDBPath)

	os.Exit(code)
}

func TestIntegrationHelloWorld(t *testing.T) {
	resp, err := http.Get(baseURL + "/")
	require.NoError(t, err, "La requête HTTP ne devrait pas échouer")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Le status code devrait être 200")

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Le décodage JSON ne devrait pas échouer")

	assert.Equal(t, "Hello World!", result["message"], "Le message devrait être 'Hello World!'")
	assert.Equal(t, "ok", result["status"], "Le status devrait être 'ok'")
}

func TestIntegrationHealthEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	require.NoError(t, err, "La requête HTTP ne devrait pas échouer")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Le status code devrait être 200")

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Le décodage JSON ne devrait pas échouer")

	assert.Equal(t, "ok", result["status"], "Le status devrait être 'ok'")
	assert.Equal(t, "connected", result["database"], "La database devrait être connectée")
}

func TestIntegrationProjectsEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/projects")
	require.NoError(t, err, "La requête HTTP ne devrait pas échouer")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Le status code devrait être 200")

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Le décodage JSON ne devrait pas échouer")

	projects, ok := result["projects"].([]interface{})
	require.True(t, ok, "Les projects devraient être un tableau")

	// Le tableau peut être vide initialement, c'est normal
	assert.NotNil(t, projects, "Le tableau projects ne devrait pas être nil")
}

func TestIntegration404NotFound(t *testing.T) {
	resp, err := http.Get(baseURL + "/nonexistent")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Le status code devrait être 404")
}

func TestIntegrationHeaders(t *testing.T) {
	resp, err := http.Get(baseURL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "application/json", "Le Content-Type devrait être application/json")
}

func TestIntegrationConcurrentRequests(t *testing.T) {
	// Tester que le serveur gère bien les requêtes concurrentes
	numRequests := 10
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(baseURL + "/health")
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("status code %d au lieu de 200", resp.StatusCode)
				return
			}

			done <- true
		}()
	}

	// Attendre que toutes les requêtes soient terminées
	successCount := 0
	for i := 0; i < numRequests; i++ {
		select {
		case <-done:
			successCount++
		case err := <-errors:
			t.Errorf("Erreur lors d'une requête concurrente: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout en attendant les requêtes concurrentes")
		}
	}

	assert.Equal(t, numRequests, successCount, "Toutes les requêtes devraient réussir")
}
