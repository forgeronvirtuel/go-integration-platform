package database

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupDeploymentTestDB crée une base de données en mémoire pour les tests de déploiement
func setupDeploymentTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Activer les foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// Créer toutes les tables nécessaires
	err = CreateProjectsTable(db)
	require.NoError(t, err)

	err = CreateBuildsTable(db)
	require.NoError(t, err)

	err = CreateAgentsTable(db)
	require.NoError(t, err)

	err = CreateDeploymentsTable(db)
	require.NoError(t, err)

	return db
}

func TestCreateDeployment(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer un projet et un build
	project, err := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	require.NoError(t, err)

	build, err := CreateBuild(db, project.ID, "main")
	require.NoError(t, err)

	// Créer un agent
	agent, err := CreateAgent(db, "test-agent", map[string]string{"env": "test"})
	require.NoError(t, err)

	// Créer un déploiement avec agent
	agentID := agent.ID
	deployment, err := CreateDeployment(db, build.ID, &agentID)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)
	assert.Equal(t, build.ID, deployment.BuildID)
	assert.Equal(t, &agentID, deployment.AgentID)
	assert.Equal(t, "pending", deployment.Status)
	assert.NotZero(t, deployment.ID)

	// Créer un déploiement sans agent
	deployment2, err := CreateDeployment(db, build.ID, nil)
	assert.NoError(t, err)
	assert.NotNil(t, deployment2)
	assert.Nil(t, deployment2.AgentID)
}

func TestGetDeploymentByID(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer les dépendances
	project, _ := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	build, _ := CreateBuild(db, project.ID, "main")
	created, _ := CreateDeployment(db, build.ID, nil)

	// Récupérer le déploiement
	deployment, err := GetDeploymentByID(db, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, deployment.ID)
	assert.Equal(t, build.ID, deployment.BuildID)
	assert.Equal(t, "pending", deployment.Status)
}

func TestGetAllDeployments(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer les dépendances
	project, _ := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	build1, _ := CreateBuild(db, project.ID, "main")
	build2, _ := CreateBuild(db, project.ID, "develop")

	// Créer plusieurs déploiements
	CreateDeployment(db, build1.ID, nil)
	CreateDeployment(db, build2.ID, nil)
	CreateDeployment(db, build1.ID, nil)

	deployments, err := GetAllDeployments(db)
	assert.NoError(t, err)
	assert.Len(t, deployments, 3)
}

func TestGetDeploymentsByBuildID(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer les dépendances
	project, _ := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	build1, _ := CreateBuild(db, project.ID, "main")
	build2, _ := CreateBuild(db, project.ID, "develop")

	// Créer des déploiements pour différents builds
	CreateDeployment(db, build1.ID, nil)
	CreateDeployment(db, build1.ID, nil)
	CreateDeployment(db, build2.ID, nil)

	deployments, err := GetDeploymentsByBuildID(db, build1.ID)
	assert.NoError(t, err)
	assert.Len(t, deployments, 2)

	for _, deployment := range deployments {
		assert.Equal(t, build1.ID, deployment.BuildID)
	}
}

func TestGetDeploymentsByAgentID(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer les dépendances
	project, _ := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	build, _ := CreateBuild(db, project.ID, "main")
	agent1, _ := CreateAgent(db, "agent-1", map[string]string{})
	agent2, _ := CreateAgent(db, "agent-2", map[string]string{})

	// Créer des déploiements pour différents agents
	agent1ID := agent1.ID
	agent2ID := agent2.ID
	CreateDeployment(db, build.ID, &agent1ID)
	CreateDeployment(db, build.ID, &agent1ID)
	CreateDeployment(db, build.ID, &agent2ID)
	CreateDeployment(db, build.ID, nil)

	deployments, err := GetDeploymentsByAgentID(db, agent1.ID)
	assert.NoError(t, err)
	assert.Len(t, deployments, 2)

	for _, deployment := range deployments {
		assert.Equal(t, &agent1ID, deployment.AgentID)
	}
}

func TestUpdateDeploymentStatus(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer les dépendances
	project, _ := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	build, _ := CreateBuild(db, project.ID, "main")
	created, _ := CreateDeployment(db, build.ID, nil)

	// Mettre à jour vers deploying
	deployment, err := UpdateDeploymentStatus(db, created.ID, "deploying")
	assert.NoError(t, err)
	assert.Equal(t, "deploying", deployment.Status)
	assert.Nil(t, deployment.EndedAt)

	// Mettre à jour vers deployed (doit définir ended_at)
	deployment, err = UpdateDeploymentStatus(db, created.ID, "deployed")
	assert.NoError(t, err)
	assert.Equal(t, "deployed", deployment.Status)
	assert.NotNil(t, deployment.EndedAt)

	// Vérifier que ended_at est bien défini
	retrieved, _ := GetDeploymentByID(db, created.ID)
	assert.NotNil(t, retrieved.EndedAt)
}

func TestUpdateDeploymentStatusFailed(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer les dépendances
	project, _ := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	build, _ := CreateBuild(db, project.ID, "main")
	created, _ := CreateDeployment(db, build.ID, nil)

	// Mettre à jour vers failed (doit définir ended_at)
	deployment, err := UpdateDeploymentStatus(db, created.ID, "failed")
	assert.NoError(t, err)
	assert.Equal(t, "failed", deployment.Status)
	assert.NotNil(t, deployment.EndedAt)
}

func TestUpdateDeploymentLog(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer les dépendances
	project, _ := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	build, _ := CreateBuild(db, project.ID, "main")
	created, _ := CreateDeployment(db, build.ID, nil)

	// Mettre à jour les logs
	logOutput := "Deployment started\nCopying files...\nDone!"
	err := UpdateDeploymentLog(db, created.ID, logOutput)
	assert.NoError(t, err)

	// Vérifier que les logs ont été mis à jour
	deployment, _ := GetDeploymentByID(db, created.ID)
	assert.Equal(t, logOutput, deployment.LogOutput)
}

func TestUpdateDeploymentAgent(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer les dépendances
	project, _ := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	build, _ := CreateBuild(db, project.ID, "main")
	agent1, _ := CreateAgent(db, "agent-1", map[string]string{})
	agent2, _ := CreateAgent(db, "agent-2", map[string]string{})
	created, _ := CreateDeployment(db, build.ID, nil)

	// Assigner un agent
	agent1ID := agent1.ID
	deployment, err := UpdateDeploymentAgent(db, created.ID, &agent1ID)
	assert.NoError(t, err)
	assert.Equal(t, &agent1ID, deployment.AgentID)

	// Changer d'agent
	agent2ID := agent2.ID
	deployment, err = UpdateDeploymentAgent(db, created.ID, &agent2ID)
	assert.NoError(t, err)
	assert.Equal(t, &agent2ID, deployment.AgentID)

	// Retirer l'agent
	deployment, err = UpdateDeploymentAgent(db, created.ID, nil)
	assert.NoError(t, err)
	assert.Nil(t, deployment.AgentID)
}

func TestDeleteDeployment(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer les dépendances
	project, _ := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	build, _ := CreateBuild(db, project.ID, "main")
	created, _ := CreateDeployment(db, build.ID, nil)

	// Supprimer le déploiement
	err := DeleteDeployment(db, created.ID)
	assert.NoError(t, err)

	// Vérifier qu'il n'existe plus
	_, err = GetDeploymentByID(db, created.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeploymentForeignKeyConstraint(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer les dépendances
	project, _ := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	build, _ := CreateBuild(db, project.ID, "main")
	deployment, _ := CreateDeployment(db, build.ID, nil)

	// Supprimer le build (doit supprimer le déploiement en CASCADE)
	_, err := db.Exec("DELETE FROM builds WHERE id = ?", build.ID)
	assert.NoError(t, err)

	// Vérifier que le déploiement a été supprimé
	_, err = GetDeploymentByID(db, deployment.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeploymentAgentNullOnDelete(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer les dépendances
	project, _ := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	build, _ := CreateBuild(db, project.ID, "main")
	agent, _ := CreateAgent(db, "agent-1", map[string]string{})
	agentID := agent.ID
	deployment, _ := CreateDeployment(db, build.ID, &agentID)

	// Supprimer l'agent (doit mettre agent_id à NULL dans le déploiement)
	err := DeleteAgent(db, agent.ID)
	assert.NoError(t, err)

	// Vérifier que le déploiement existe toujours mais sans agent
	retrieved, err := GetDeploymentByID(db, deployment.ID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved.AgentID)
}

func TestDeploymentTimestamps(t *testing.T) {
	db := setupDeploymentTestDB(t)
	defer db.Close()

	// Créer les dépendances
	project, _ := CreateProject(db, "test-project", "https://github.com/test/repo.git", "main", "")
	build, _ := CreateBuild(db, project.ID, "main")

	beforeCreate := time.Now()
	time.Sleep(10 * time.Millisecond)

	deployment, err := CreateDeployment(db, build.ID, nil)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)
	afterCreate := time.Now()

	// Vérifier que les timestamps sont dans l'intervalle attendu
	assert.True(t, deployment.CreatedAt.After(beforeCreate))
	assert.True(t, deployment.CreatedAt.Before(afterCreate))
	assert.True(t, deployment.StartedAt.After(beforeCreate))
	assert.True(t, deployment.StartedAt.Before(afterCreate))
}
