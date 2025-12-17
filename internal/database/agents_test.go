package database

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB crée une base de données en mémoire pour les tests
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Activer les foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// Créer la table runners
	err = CreateRunnersTable(db)
	require.NoError(t, err)

	return db
}

func TestCreateRunner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	labels := map[string]string{
		"os":   "linux",
		"arch": "amd64",
	}

	runner, err := CreateRunner(db, "test-runner-1", labels)
	assert.NoError(t, err)
	assert.NotNil(t, runner)
	assert.Equal(t, "test-runner-1", runner.Name)
	assert.Equal(t, "OFFLINE", runner.Status)
	assert.Equal(t, "linux", runner.Labels["os"])
	assert.Equal(t, "amd64", runner.Labels["arch"])
}

func TestGetRunnerByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	labels := map[string]string{"env": "test"}
	created, _ := CreateRunner(db, "test-runner-2", labels)

	runner, err := GetRunnerByID(db, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, runner.ID)
	assert.Equal(t, "test-runner-2", runner.Name)
	assert.Equal(t, "test", runner.Labels["env"])
}

func TestGetRunnerByName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	labels := map[string]string{"region": "eu-west"}
	CreateRunner(db, "test-runner-3", labels)

	runner, err := GetRunnerByName(db, "test-runner-3")
	assert.NoError(t, err)
	assert.Equal(t, "test-runner-3", runner.Name)
	assert.Equal(t, "eu-west", runner.Labels["region"])
}

func TestGetAllRunners(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	CreateRunner(db, "runner-1", map[string]string{})
	CreateRunner(db, "runner-2", map[string]string{})
	CreateRunner(db, "runner-3", map[string]string{})

	runners, err := GetAllRunners(db)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(runners), 3)
}

func TestGetRunnersByStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	runner1, _ := CreateRunner(db, "runner-online-1", map[string]string{})
	runner2, _ := CreateRunner(db, "runner-online-2", map[string]string{})
	CreateRunner(db, "runner-offline", map[string]string{})

	// Passer 2 runners à ONLINE
	UpdateRunnerStatus(db, runner1.ID, "ONLINE")
	UpdateRunnerStatus(db, runner2.ID, "ONLINE")

	onlineRunners, err := GetRunnersByStatus(db, "ONLINE")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(onlineRunners), 2)

	offlineRunners, err := GetRunnersByStatus(db, "OFFLINE")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(offlineRunners), 1)
}

func TestUpdateRunnerStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	runner, _ := CreateRunner(db, "test-runner-status", map[string]string{})
	assert.Equal(t, "OFFLINE", runner.Status)

	err := UpdateRunnerStatus(db, runner.ID, "ONLINE")
	assert.NoError(t, err)

	updated, _ := GetRunnerByID(db, runner.ID)
	assert.Equal(t, "ONLINE", updated.Status)

	err = UpdateRunnerStatus(db, runner.ID, "DRAINING")
	assert.NoError(t, err)

	updated, _ = GetRunnerByID(db, runner.ID)
	assert.Equal(t, "DRAINING", updated.Status)
}

func TestUpdateRunnerLabels(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	initialLabels := map[string]string{"env": "dev"}
	runner, _ := CreateRunner(db, "test-runner-labels", initialLabels)

	newLabels := map[string]string{
		"env":    "prod",
		"region": "us-east",
	}
	err := UpdateRunnerLabels(db, runner.ID, newLabels)
	assert.NoError(t, err)

	updated, _ := GetRunnerByID(db, runner.ID)
	assert.Equal(t, "prod", updated.Labels["env"])
	assert.Equal(t, "us-east", updated.Labels["region"])
}

func TestUpdateRunnerHeartbeat(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	runner, _ := CreateRunner(db, "test-runner-heartbeat", map[string]string{})

	// Attendre un peu et mettre à jour le heartbeat
	time.Sleep(1 * time.Second)

	err := UpdateRunnerHeartbeat(db, runner.ID)
	assert.NoError(t, err)

	updated, _ := GetRunnerByID(db, runner.ID)
	// Vérifier que le heartbeat a été mis à jour (doit être plus récent)
	// On compare juste que le temps est proche de maintenant
	assert.WithinDuration(t, time.Now(), updated.LastSeenAt, 2*time.Second)
}

func TestDeleteRunner(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	runner, _ := CreateRunner(db, "test-runner-delete", map[string]string{})

	err := DeleteRunner(db, runner.ID)
	assert.NoError(t, err)

	_, err = GetRunnerByID(db, runner.ID)
	assert.Error(t, err)
}

func TestMarkStaleRunnersOffline(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Créer un runner et le mettre ONLINE
	runner, _ := CreateRunner(db, "test-runner-stale", map[string]string{})
	UpdateRunnerStatus(db, runner.ID, "ONLINE")

	// Simuler un runner ancien en modifiant manuellement last_seen_at
	_, err := db.Exec("UPDATE runners SET last_seen_at = datetime('now', '-10 minutes') WHERE id = ?", runner.ID)
	assert.NoError(t, err)

	// Marquer les runners sans heartbeat depuis 5 minutes comme OFFLINE
	count, err := MarkStaleRunnersOffline(db, 5*time.Minute)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1)

	// Vérifier que l'runner est maintenant OFFLINE
	updated, _ := GetRunnerByID(db, runner.ID)
	assert.Equal(t, "OFFLINE", updated.Status)
}

func TestRunnerUniqueNameConstraint(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	CreateRunner(db, "unique-runner", map[string]string{})

	// Tenter de créer un autre runner avec le même nom
	_, err := CreateRunner(db, "unique-runner", map[string]string{})
	assert.Error(t, err)
}
