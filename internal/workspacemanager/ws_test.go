package workspacemanager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateWorkspaceDir(t *testing.T) {
	// Test avec un répertoire temporaire valide
	tmpDir := t.TempDir()
	err := ValidateWorkspaceDir(tmpDir)
	assert.NoError(t, err, "Un répertoire temporaire valide devrait être accepté")
}

func TestValidateWorkspaceDir_NonExistent(t *testing.T) {
	// Test avec un répertoire qui n'existe pas (devrait être créé)
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "new-workspace")

	err := ValidateWorkspaceDir(newDir)
	assert.NoError(t, err, "Le répertoire devrait être créé automatiquement")

	// Vérifier que le répertoire existe maintenant
	info, err := os.Stat(newDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir(), "Le chemin devrait être un répertoire")
}

func TestValidateWorkspaceDir_File(t *testing.T) {
	// Test avec un fichier au lieu d'un répertoire
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file.txt")

	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	require.NoError(t, err)

	err = ValidateWorkspaceDir(tmpFile)
	assert.Error(t, err, "Un fichier ne devrait pas être accepté comme workspace")
}

func TestValidateWorkspaceDir_ReadOnly(t *testing.T) {
	// Test avec un répertoire en lecture seule
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")

	err := os.Mkdir(readOnlyDir, 0755)
	require.NoError(t, err)

	// Rendre le répertoire en lecture seule
	err = os.Chmod(readOnlyDir, 0444)
	require.NoError(t, err)

	// Restaurer les permissions après le test
	defer os.Chmod(readOnlyDir, 0755)

	err = ValidateWorkspaceDir(readOnlyDir)
	assert.Error(t, err, "Un répertoire en lecture seule ne devrait pas être accepté")
}

func TestValidateWorkspaceDir_Writable(t *testing.T) {
	// Test avec un répertoire accessible en lecture/écriture
	tmpDir := t.TempDir()

	err := ValidateWorkspaceDir(tmpDir)
	assert.NoError(t, err, "Un répertoire accessible en lecture/écriture devrait être accepté")

	// Vérifier qu'on peut créer un fichier dans le répertoire
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	assert.NoError(t, err, "On devrait pouvoir créer un fichier dans le workspace")
}
