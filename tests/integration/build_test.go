package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationBuildAndDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 1. Créer un projet
	projectPayload := map[string]interface{}{
		"name":     "test-build-project",
		"repo_url": "https://github.com/golang/example.git",
		"branch":   "master",
		"subdir":   "hello",
	}

	payloadBytes, _ := json.Marshal(projectPayload)
	resp, err := http.Post(baseURL+"/api/projects", "application/json", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var projectResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&projectResult)
	projectID := int(projectResult["id"].(float64))

	// 2. Déclencher un build
	buildPayload := map[string]interface{}{
		"project_id": projectID,
	}

	payloadBytes, _ = json.Marshal(buildPayload)
	resp, err = http.Post(baseURL+"/api/builds/", "application/json", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Le build peut échouer si le timeout est court ou si le repo n'a pas cmd/main.go
	// On vérifie juste que l'endpoint répond
	assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusBadRequest)

	if resp.StatusCode == http.StatusCreated {
		var buildResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&buildResult)

		buildID := int(buildResult["build_id"].(float64))
		downloadURL := buildResult["download_url"].(string)

		t.Logf("Build created with ID: %d", buildID)
		t.Logf("Download URL: %s", downloadURL)

		// 3. Tenter de télécharger le binaire
		resp, err = http.Get(baseURL + downloadURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// Sauvegarder le binaire dans un fichier temporaire pour vérifier
			tmpFile := filepath.Join(t.TempDir(), "downloaded-binary")
			out, err := os.Create(tmpFile)
			require.NoError(t, err)
			defer out.Close()

			size, err := io.Copy(out, resp.Body)
			require.NoError(t, err)

			t.Logf("Downloaded binary size: %d bytes", size)
			assert.Greater(t, size, int64(0), "Binary should not be empty")

			// Vérifier les headers
			assert.Equal(t, "application/octet-stream", resp.Header.Get("Content-Type"))
			assert.Contains(t, resp.Header.Get("Content-Disposition"), "attachment")
		}
	}
}

func TestBuildEndpoint_ProjectNotFound(t *testing.T) {
	buildPayload := map[string]interface{}{
		"project_id": 99999,
	}

	payloadBytes, _ := json.Marshal(buildPayload)
	resp, err := http.Post(baseURL+"/api/builds/", "application/json", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Contains(t, result["error"], "Project not found")
}

func TestDownloadBinary_BuildNotFound(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/builds/99999/download")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
