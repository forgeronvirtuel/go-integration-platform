package database

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitDB(t *testing.T) {
	// Créer un fichier temporaire pour la base de données
	tmpFile := "./test_data.db"
	defer os.Remove(tmpFile)

	db, err := InitDB(tmpFile)
	require.NoError(t, err, "InitDB ne devrait pas retourner d'erreur")
	require.NotNil(t, db, "La connexion DB ne devrait pas être nulle")
	defer db.Close()

	// Vérifier que la connexion fonctionne
	err = db.Ping()
	assert.NoError(t, err, "La base de données devrait être accessible")
}

func TestCreateTables(t *testing.T) {
	// Utiliser une base de données en mémoire pour les tests
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Activer les foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// Créer les tables
	err = createTables(db)
	require.NoError(t, err, "createTables ne devrait pas retourner d'erreur")

	// Vérifier que la table projects existe
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='projects'").Scan(&tableName)
	assert.NoError(t, err, "La table projects devrait exister")
	assert.Equal(t, "projects", tableName)
}

func TestProjectsTableStructure(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTables(db)
	require.NoError(t, err)

	// Vérifier la structure de la table projects
	rows, err := db.Query("PRAGMA table_info(projects)")
	require.NoError(t, err)
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString

		err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk)
		require.NoError(t, err)
		columns[name] = true
	}

	// Vérifier que toutes les colonnes attendues existent
	assert.True(t, columns["id"], "La colonne id devrait exister")
	assert.True(t, columns["name"], "La colonne name devrait exister")
	assert.True(t, columns["repo_url"], "La colonne repo_url devrait exister")
	assert.True(t, columns["branch"], "La colonne branch devrait exister")
	assert.True(t, columns["subdir"], "La colonne subdir devrait exister")
	assert.True(t, columns["created_at"], "La colonne created_at devrait exister")
	assert.True(t, columns["updated_at"], "La colonne updated_at devrait exister")
}

func TestDatabasePersistence(t *testing.T) {
	tmpFile := "./test_persistence.db"
	defer os.Remove(tmpFile)

	// Première connexion: créer la DB et ajouter un projet
	db1, err := InitDB(tmpFile)
	require.NoError(t, err)

	_, err = CreateProject(db1, "test-project", "https://github.com/test/test.git", "main", "")
	require.NoError(t, err)

	var count1 int
	err = db1.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count1)
	require.NoError(t, err)
	assert.Equal(t, 1, count1, "Il devrait y avoir 1 projet")
	db1.Close()

	// Deuxième connexion: vérifier que les données persistent
	db2, err := InitDB(tmpFile)
	require.NoError(t, err)
	defer db2.Close()

	var count2 int
	err = db2.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count2)
	require.NoError(t, err)

	assert.Equal(t, count1, count2, "Les données devraient persister entre les connexions")
}
