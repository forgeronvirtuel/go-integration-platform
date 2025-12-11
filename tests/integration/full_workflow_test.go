package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullBuildWorkflow teste l'intégralité du workflow :
// 1. Créer un faux projet Go local avec un main.go qui affiche "Hello World"
// 2. Initialiser un repo git local
// 3. Créer le projet dans la plateforme
// 4. Lancer le build du binaire
// 5. Télécharger le binaire
// 6. Exécuter le binaire et vérifier la sortie
func TestFullBuildWorkflow(t *testing.T) {
	// === ÉTAPE 1: Créer un projet Go local ===
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "hello-world-project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err, "Création du répertoire projet")

	// Créer le répertoire cmd
	cmdDir := filepath.Join(projectDir, "cmd")
	err = os.MkdirAll(cmdDir, 0755)
	require.NoError(t, err, "Création du répertoire cmd")

	// Créer le fichier main.go
	mainGoContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello World from GIP Build!")
}
`
	mainGoPath := filepath.Join(cmdDir, "main.go")
	err = os.WriteFile(mainGoPath, []byte(mainGoContent), 0644)
	require.NoError(t, err, "Création de main.go")

	// Créer go.mod
	goModContent := `module example.com/hello-world

go 1.21
`
	goModPath := filepath.Join(projectDir, "go.mod")
	err = os.WriteFile(goModPath, []byte(goModContent), 0644)
	require.NoError(t, err, "Création de go.mod")

	// === ÉTAPE 2: Initialiser un repo git local ===
	// Initialiser git
	cmd := exec.Command("git", "init")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git init devrait réussir: %s", output)

	// Configurer git user (nécessaire pour commit)
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = projectDir
	err = cmd.Run()
	require.NoError(t, err, "git config user.email")

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = projectDir
	err = cmd.Run()
	require.NoError(t, err, "git config user.name")

	// Ajouter tous les fichiers
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = projectDir
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "git add devrait réussir: %s", output)

	// Commit
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = projectDir
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "git commit devrait réussir: %s", output)

	t.Logf("Projet Git créé dans: %s", projectDir)

	// === ÉTAPE 3: Créer le projet dans la plateforme ===
	projectPayload := map[string]interface{}{
		"name":     "hello-world-test",
		"repo_url": projectDir, // Utiliser le chemin local comme URL (git supporte les chemins locaux)
		"branch":   "master",
	}

	payloadBytes, err := json.Marshal(projectPayload)
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/api/projects", "application/json", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err, "Création du projet via API")
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "Le projet devrait être créé")

	var projectResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&projectResult)
	require.NoError(t, err)

	projectID := int(projectResult["id"].(float64))
	t.Logf("Projet créé avec ID: %d", projectID)

	// === ÉTAPE 4: Lancer le build ===
	buildPayload := map[string]interface{}{
		"project_id": projectID,
	}

	payloadBytes, err = json.Marshal(buildPayload)
	require.NoError(t, err)

	t.Log("Lancement du build (peut prendre quelques secondes)...")
	resp, err = http.Post(baseURL+"/api/builds/", "application/json", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err, "Lancement du build via API")
	defer resp.Body.Close()

	// Lire la réponse
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	if resp.StatusCode != http.StatusCreated {
		t.Logf("Réponse du build: %s", string(bodyBytes))
		require.Equal(t, http.StatusCreated, resp.StatusCode, "Le build devrait réussir")
	}

	var buildResult map[string]interface{}
	err = json.Unmarshal(bodyBytes, &buildResult)
	require.NoError(t, err)

	buildID := int(buildResult["build_id"].(float64))
	buildStatus := buildResult["status"].(string)
	downloadURL := buildResult["download_url"].(string)

	t.Logf("Build créé avec ID: %d, Status: %s", buildID, buildStatus)
	t.Logf("URL de téléchargement: %s", downloadURL)

	require.Equal(t, "success", buildStatus, "Le build devrait être en succès")
	require.NotEmpty(t, downloadURL, "L'URL de téléchargement devrait être fournie")

	// === ÉTAPE 5: Télécharger le binaire ===
	t.Log("Téléchargement du binaire...")
	resp, err = http.Get(baseURL + downloadURL)
	require.NoError(t, err, "Téléchargement du binaire")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Le téléchargement devrait réussir")

	// Vérifier les headers
	contentType := resp.Header.Get("Content-Type")
	assert.Equal(t, "application/octet-stream", contentType, "Content-Type devrait être octet-stream")

	contentDisposition := resp.Header.Get("Content-Disposition")
	assert.Contains(t, contentDisposition, "attachment", "Content-Disposition devrait contenir attachment")

	// Sauvegarder le binaire
	binaryPath := filepath.Join(tempDir, "downloaded-binary")
	outFile, err := os.Create(binaryPath)
	require.NoError(t, err, "Création du fichier pour le binaire")

	size, err := io.Copy(outFile, resp.Body)
	require.NoError(t, err, "Écriture du binaire")
	require.Greater(t, size, int64(0), "Le binaire ne devrait pas être vide")

	// Fermer le fichier AVANT de l'utiliser
	outFile.Close()

	t.Logf("Binaire téléchargé: %d bytes", size)

	// Rendre le binaire exécutable
	err = os.Chmod(binaryPath, 0755)
	require.NoError(t, err, "Chmod du binaire")

	// === ÉTAPE 6: Exécuter le binaire et vérifier la sortie ===
	t.Log("Exécution du binaire...")

	// Attendre un peu pour être sûr que le fichier est bien fermé
	time.Sleep(200 * time.Millisecond)

	cmd = exec.Command(binaryPath)
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "L'exécution du binaire devrait réussir: %s", output)

	outputStr := string(output)
	t.Logf("Sortie du binaire: %s", outputStr)

	// Vérifier que la sortie contient "Hello World"
	assert.Contains(t, outputStr, "Hello World from GIP Build!", "La sortie devrait contenir le message attendu")

	t.Log("✅ Test complet réussi !")
}

// TestFullBuildWorkflow_WithSubdir teste le même workflow mais avec un sous-répertoire
func TestFullBuildWorkflow_WithSubdir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long test in short mode")
	}

	// === ÉTAPE 1: Créer un projet Go local avec sous-répertoire ===
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "multi-module-project")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Créer un sous-répertoire pour le service
	serviceDir := filepath.Join(projectDir, "services", "api")
	err = os.MkdirAll(serviceDir, 0755)
	require.NoError(t, err)

	// Créer cmd dans le sous-répertoire
	cmdDir := filepath.Join(serviceDir, "cmd")
	err = os.MkdirAll(cmdDir, 0755)
	require.NoError(t, err)

	// Créer main.go
	mainGoContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello from subdir build!")
}
`
	err = os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(mainGoContent), 0644)
	require.NoError(t, err)

	// Créer go.mod dans le sous-répertoire
	goModContent := `module example.com/services/api

go 1.21
`
	err = os.WriteFile(filepath.Join(serviceDir, "go.mod"), []byte(goModContent), 0644)
	require.NoError(t, err)

	// Initialiser git à la racine
	cmd := exec.Command("git", "init")
	cmd.Dir = projectDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = projectDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = projectDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = projectDir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = projectDir
	require.NoError(t, cmd.Run())

	// === ÉTAPE 2: Créer le projet avec subdir ===
	projectPayload := map[string]interface{}{
		"name":     "subdir-test",
		"repo_url": projectDir,
		"branch":   "master",
		"subdir":   "services/api",
	}

	payloadBytes, err := json.Marshal(projectPayload)
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/api/projects", "application/json", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var projectResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&projectResult)
	require.NoError(t, err)

	projectID := int(projectResult["id"].(float64))

	// === ÉTAPE 3: Build et test ===
	buildPayload := map[string]interface{}{
		"project_id": projectID,
	}

	payloadBytes, err = json.Marshal(buildPayload)
	require.NoError(t, err)

	resp, err = http.Post(baseURL+"/api/builds/", "application/json", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	if resp.StatusCode != http.StatusCreated {
		t.Logf("Build response: %s", string(bodyBytes))
		t.FailNow()
	}

	var buildResult map[string]interface{}
	err = json.Unmarshal(bodyBytes, &buildResult)
	require.NoError(t, err)

	downloadURL := buildResult["download_url"].(string)

	// Télécharger et exécuter
	resp, err = http.Get(baseURL + downloadURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	binaryPath := filepath.Join(tempDir, "subdir-binary")
	outFile, err := os.Create(binaryPath)
	require.NoError(t, err)

	_, err = io.Copy(outFile, resp.Body)
	require.NoError(t, err)

	// Fermer le fichier AVANT de l'utiliser
	outFile.Close()

	err = os.Chmod(binaryPath, 0755)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	cmd = exec.Command(binaryPath)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Output: %s", output)

	assert.Contains(t, string(output), "Hello from subdir build!")

	t.Log("✅ Test avec sous-répertoire réussi !")
}

// TestBuildWorkflow_FailureCases teste les cas d'échec
func TestBuildWorkflow_FailureCases(t *testing.T) {
	// Test 1: Projet sans cmd/main.go
	t.Run("NoMainGo", func(t *testing.T) {
		tempDir := t.TempDir()
		projectDir := filepath.Join(tempDir, "no-main-project")
		err := os.MkdirAll(projectDir, 0755)
		require.NoError(t, err)

		// Créer juste go.mod, pas de cmd/main.go
		goModContent := `module example.com/nomain

go 1.21
`
		err = os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(goModContent), 0644)
		require.NoError(t, err)

		// Init git
		cmd := exec.Command("git", "init")
		cmd.Dir = projectDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = projectDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = projectDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = projectDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = projectDir
		require.NoError(t, cmd.Run())

		// Créer le projet
		projectPayload := map[string]interface{}{
			"name":     "no-main-test",
			"repo_url": projectDir,
			"branch":   "master",
		}

		payloadBytes, err := json.Marshal(projectPayload)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/api/projects", "application/json", bytes.NewBuffer(payloadBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		var projectResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&projectResult)
		projectID := int(projectResult["id"].(float64))

		// Tenter le build
		buildPayload := map[string]interface{}{
			"project_id": projectID,
		}

		payloadBytes, _ = json.Marshal(buildPayload)
		resp, err = http.Post(baseURL+"/api/builds/", "application/json", bytes.NewBuffer(payloadBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Le build devrait échouer
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Le build devrait échouer sans cmd/main.go")

		var buildResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&buildResult)
		assert.Contains(t, buildResult["error"], "cmd/main.go not found")
	})

	// Test 2: Code avec erreur de compilation
	t.Run("CompilationError", func(t *testing.T) {
		tempDir := t.TempDir()
		projectDir := filepath.Join(tempDir, "compile-error-project")
		cmdDir := filepath.Join(projectDir, "cmd")
		err := os.MkdirAll(cmdDir, 0755)
		require.NoError(t, err)

		// Créer un main.go avec une erreur de syntaxe
		mainGoContent := `package main

import "fmt"

func main() {
	fmt.Println("Missing quote)
}
`
		err = os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(mainGoContent), 0644)
		require.NoError(t, err)

		goModContent := `module example.com/error

go 1.21
`
		err = os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(goModContent), 0644)
		require.NoError(t, err)

		// Init git
		cmd := exec.Command("git", "init")
		cmd.Dir = projectDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = projectDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = projectDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = projectDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = projectDir
		require.NoError(t, cmd.Run())

		// Créer projet et build
		projectPayload := map[string]interface{}{
			"name":     "compile-error-test",
			"repo_url": projectDir,
			"branch":   "master",
		}

		payloadBytes, _ := json.Marshal(projectPayload)
		resp, err := http.Post(baseURL+"/api/projects", "application/json", bytes.NewBuffer(payloadBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		var projectResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&projectResult)
		projectID := int(projectResult["id"].(float64))

		buildPayload := map[string]interface{}{
			"project_id": projectID,
		}

		payloadBytes, _ = json.Marshal(buildPayload)
		resp, err = http.Post(baseURL+"/api/builds/", "application/json", bytes.NewBuffer(payloadBytes))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Le build devrait échouer
		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Le build devrait échouer avec erreur de compilation")

		var buildResult map[string]interface{}
		json.Unmarshal(bodyBytes, &buildResult)

		// Les logs devraient contenir des infos sur l'erreur
		if logs, ok := buildResult["logs"].(string); ok {
			t.Logf("Logs d'erreur: %s", logs)
		}
	})
}
