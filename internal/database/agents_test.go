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

	// Créer la table agents
	err = CreateAgentsTable(db)
	require.NoError(t, err)

	return db
}

func TestCreateAgent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	labels := map[string]string{
		"os":   "linux",
		"arch": "amd64",
	}

	agent, err := CreateAgent(db, "test-agent-1", labels)
	assert.NoError(t, err)
	assert.NotNil(t, agent)
	assert.Equal(t, "test-agent-1", agent.Name)
	assert.Equal(t, "OFFLINE", agent.Status)
	assert.Equal(t, "linux", agent.Labels["os"])
	assert.Equal(t, "amd64", agent.Labels["arch"])
}

func TestGetAgentByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	labels := map[string]string{"env": "test"}
	created, _ := CreateAgent(db, "test-agent-2", labels)

	agent, err := GetAgentByID(db, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, agent.ID)
	assert.Equal(t, "test-agent-2", agent.Name)
	assert.Equal(t, "test", agent.Labels["env"])
}

func TestGetAgentByName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	labels := map[string]string{"region": "eu-west"}
	CreateAgent(db, "test-agent-3", labels)

	agent, err := GetAgentByName(db, "test-agent-3")
	assert.NoError(t, err)
	assert.Equal(t, "test-agent-3", agent.Name)
	assert.Equal(t, "eu-west", agent.Labels["region"])
}

func TestGetAllAgents(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	CreateAgent(db, "agent-1", map[string]string{})
	CreateAgent(db, "agent-2", map[string]string{})
	CreateAgent(db, "agent-3", map[string]string{})

	agents, err := GetAllAgents(db)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(agents), 3)
}

func TestGetAgentsByStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	agent1, _ := CreateAgent(db, "agent-online-1", map[string]string{})
	agent2, _ := CreateAgent(db, "agent-online-2", map[string]string{})
	CreateAgent(db, "agent-offline", map[string]string{})

	// Passer 2 agents à ONLINE
	UpdateAgentStatus(db, agent1.ID, "ONLINE")
	UpdateAgentStatus(db, agent2.ID, "ONLINE")

	onlineAgents, err := GetAgentsByStatus(db, "ONLINE")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(onlineAgents), 2)

	offlineAgents, err := GetAgentsByStatus(db, "OFFLINE")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(offlineAgents), 1)
}

func TestUpdateAgentStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	agent, _ := CreateAgent(db, "test-agent-status", map[string]string{})
	assert.Equal(t, "OFFLINE", agent.Status)

	err := UpdateAgentStatus(db, agent.ID, "ONLINE")
	assert.NoError(t, err)

	updated, _ := GetAgentByID(db, agent.ID)
	assert.Equal(t, "ONLINE", updated.Status)

	err = UpdateAgentStatus(db, agent.ID, "DRAINING")
	assert.NoError(t, err)

	updated, _ = GetAgentByID(db, agent.ID)
	assert.Equal(t, "DRAINING", updated.Status)
}

func TestUpdateAgentLabels(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	initialLabels := map[string]string{"env": "dev"}
	agent, _ := CreateAgent(db, "test-agent-labels", initialLabels)

	newLabels := map[string]string{
		"env":    "prod",
		"region": "us-east",
	}
	err := UpdateAgentLabels(db, agent.ID, newLabels)
	assert.NoError(t, err)

	updated, _ := GetAgentByID(db, agent.ID)
	assert.Equal(t, "prod", updated.Labels["env"])
	assert.Equal(t, "us-east", updated.Labels["region"])
}

func TestUpdateAgentHeartbeat(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	agent, _ := CreateAgent(db, "test-agent-heartbeat", map[string]string{})

	// Attendre un peu et mettre à jour le heartbeat
	time.Sleep(1 * time.Second)

	err := UpdateAgentHeartbeat(db, agent.ID)
	assert.NoError(t, err)

	updated, _ := GetAgentByID(db, agent.ID)
	// Vérifier que le heartbeat a été mis à jour (doit être plus récent)
	// On compare juste que le temps est proche de maintenant
	assert.WithinDuration(t, time.Now(), updated.LastSeenAt, 2*time.Second)
}

func TestDeleteAgent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	agent, _ := CreateAgent(db, "test-agent-delete", map[string]string{})

	err := DeleteAgent(db, agent.ID)
	assert.NoError(t, err)

	_, err = GetAgentByID(db, agent.ID)
	assert.Error(t, err)
}

func TestMarkStaleAgentsOffline(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Créer un agent et le mettre ONLINE
	agent, _ := CreateAgent(db, "test-agent-stale", map[string]string{})
	UpdateAgentStatus(db, agent.ID, "ONLINE")

	// Simuler un agent ancien en modifiant manuellement last_seen_at
	_, err := db.Exec("UPDATE agents SET last_seen_at = datetime('now', '-10 minutes') WHERE id = ?", agent.ID)
	assert.NoError(t, err)

	// Marquer les agents sans heartbeat depuis 5 minutes comme OFFLINE
	count, err := MarkStaleAgentsOffline(db, 5*time.Minute)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1)

	// Vérifier que l'agent est maintenant OFFLINE
	updated, _ := GetAgentByID(db, agent.ID)
	assert.Equal(t, "OFFLINE", updated.Status)
}

func TestAgentUniqueNameConstraint(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	CreateAgent(db, "unique-agent", map[string]string{})

	// Tenter de créer un autre agent avec le même nom
	_, err := CreateAgent(db, "unique-agent", map[string]string{})
	assert.Error(t, err)
}
