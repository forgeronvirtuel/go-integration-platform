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

	// Créer les tables
	err = createTables(db)
	require.NoError(t, err, "createTables ne devrait pas retourner d'erreur")

	// Vérifier que la table users existe
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableName)
	assert.NoError(t, err, "La table users devrait exister")
	assert.Equal(t, "users", tableName)

	// Vérifier que des données de test ont été insérées
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 3, count, "Il devrait y avoir 3 utilisateurs de test")
}

func TestUsersTableStructure(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTables(db)
	require.NoError(t, err)

	// Vérifier la structure de la table
	rows, err := db.Query("PRAGMA table_info(users)")
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
	assert.True(t, columns["email"], "La colonne email devrait exister")
	assert.True(t, columns["created_at"], "La colonne created_at devrait exister")
}

func TestExampleQuery(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTables(db)
	require.NoError(t, err)

	// Tester la fonction ExampleQuery
	err = ExampleQuery(db)
	assert.NoError(t, err, "ExampleQuery ne devrait pas retourner d'erreur")
}

func TestQueryUsers(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTables(db)
	require.NoError(t, err)

	// Récupérer tous les utilisateurs
	rows, err := db.Query("SELECT id, name, email FROM users")
	require.NoError(t, err)
	defer rows.Close()

	users := []struct {
		id    int
		name  string
		email string
	}{}

	for rows.Next() {
		var u struct {
			id    int
			name  string
			email string
		}
		err := rows.Scan(&u.id, &u.name, &u.email)
		require.NoError(t, err)
		users = append(users, u)
	}

	assert.Len(t, users, 3, "Il devrait y avoir 3 utilisateurs")
	assert.Equal(t, "Alice Dupont", users[0].name)
	assert.Equal(t, "alice@example.com", users[0].email)
}

func TestDatabasePersistence(t *testing.T) {
	tmpFile := "./test_persistence.db"
	defer os.Remove(tmpFile)

	// Première connexion: créer la DB
	db1, err := InitDB(tmpFile)
	require.NoError(t, err)

	var count1 int
	err = db1.QueryRow("SELECT COUNT(*) FROM users").Scan(&count1)
	require.NoError(t, err)
	db1.Close()

	// Deuxième connexion: vérifier que les données persistent
	db2, err := InitDB(tmpFile)
	require.NoError(t, err)
	defer db2.Close()

	var count2 int
	err = db2.QueryRow("SELECT COUNT(*) FROM users").Scan(&count2)
	require.NoError(t, err)

	assert.Equal(t, count1, count2, "Les données devraient persister entre les connexions")
}
