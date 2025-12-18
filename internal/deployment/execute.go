package deployment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// executeDeployment télécharge et exécute un build
func ExecuteDeployment(runnerID, deploymentID, buildID int, controlPlaneURL string) {
	log.Info().
		Int("deployment_id", deploymentID).
		Int("build_id", buildID).
		Msg("Starting deployment execution")

	// Download the binary
	var downloadUrl = fmt.Sprintf("%s/api/v1/builds/%d/download", controlPlaneURL, buildID)
	log.Info().Str("download_url", downloadUrl).Msg("Downloading binary")
	binaryResp, err := http.Get(downloadUrl)
	if err != nil {
		log.Error().Err(err).Str("download_url", downloadUrl).Msg("Failed to download binary")
		UpdateDeploymentStatus(controlPlaneURL, deploymentID, "failed", fmt.Sprintf("Failed to download binary: %v", err))
		return
	}
	defer binaryResp.Body.Close()

	if binaryResp.StatusCode != http.StatusOK {
		log.Error().Int("status_code", binaryResp.StatusCode).Str("download_url", downloadUrl).Msg("Failed to download binary")
		UpdateDeploymentStatus(controlPlaneURL, deploymentID, "failed", fmt.Sprintf("Failed to download binary: HTTP %d", binaryResp.StatusCode))
		return
	}

	// Create a temporary directory for the binary
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("deployment_%d_", deploymentID))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create temporary directory")
		UpdateDeploymentStatus(controlPlaneURL, deploymentID, "failed", fmt.Sprintf("Failed to create temporary directory: %v", err))
		return
	}
	defer os.RemoveAll(tempDir)

	// Create the binary file path
	binaryPath := filepath.Join(tempDir, fmt.Sprintf("build_%d", buildID))
	binaryFile, err := os.Create(binaryPath)
	if err != nil {
		log.Error().Err(err).Str("binary_path", binaryPath).Msg("Failed to create binary file")
		UpdateDeploymentStatus(controlPlaneURL, deploymentID, "failed", fmt.Sprintf("Failed to create binary file: %v", err))
		return
	}
	defer binaryFile.Close()

	// Copy the downloaded binary to the file
	_, err = io.Copy(binaryFile, binaryResp.Body)
	if err != nil {
		log.Error().Err(err).Str("binary_path", binaryPath).Msg("Failed to save binary")
		UpdateDeploymentStatus(controlPlaneURL, deploymentID, "failed", fmt.Sprintf("Failed to save binary: %v", err))
		return
	}

	// Flush and close the file
	if err := binaryFile.Sync(); err != nil {
		log.Error().Err(err).Str("binary_path", binaryPath).Msg("Failed to flush binary file")
		UpdateDeploymentStatus(controlPlaneURL, deploymentID, "failed", fmt.Sprintf("Failed to flush binary file: %v", err))
		return
	}
	if err := binaryFile.Close(); err != nil {
		log.Error().Err(err).Str("binary_path", binaryPath).Msg("Failed to close binary file")
		UpdateDeploymentStatus(controlPlaneURL, deploymentID, "failed", fmt.Sprintf("Failed to close binary file: %v", err))
		return
	}

	// Make the binary executable
	if err := os.Chmod(binaryPath, 0755); err != nil {
		log.Error().Err(err).Str("binary_path", binaryPath).Msg("Failed to make binary executable")
		UpdateDeploymentStatus(controlPlaneURL, deploymentID, "failed", fmt.Sprintf("Failed to make binary executable: %v", err))
		return
	}

	log.Info().Str("binary_path", binaryPath).Msg("Binary saved successfully")

	// Execute the binary
	cmd := exec.Command(binaryPath)
	var outputBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	log.Info().Int("deployment_id", deploymentID).Msg("Executing binary")
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Int("deployment_id", deploymentID).Msg("Binary execution failed")
		UpdateDeploymentStatus(controlPlaneURL, deploymentID, "failed", fmt.Sprintf("Binary execution failed: %v\nOutput:\n%s", err, outputBuf.String()))
		return
	}
	UpdateDeploymentStatus(controlPlaneURL, deploymentID, "deployed", outputBuf.String())
	log.Info().Int("deployment_id", deploymentID).Msg("Deployment completed")
}

// UpdateDeploymentStatus met à jour le statut d'un déploiement
func UpdateDeploymentStatus(controlPlaneURL string, deploymentID int, status, logOutput string) error {
	// Update status
	statusURL := fmt.Sprintf("%s/api/v1/deployments/%d/status", controlPlaneURL, deploymentID)
	statusData := map[string]string{"status": status}
	jsonData, err := json.Marshal(statusData)
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation JSON du status: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, statusURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la requête de status: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erreur lors de la requête HTTP de status: %w", err)
	}
	defer resp.Body.Close()

	// Update logs
	logURL := fmt.Sprintf("%s/api/v1/deployments/%d/log", controlPlaneURL, deploymentID)
	logData := map[string]string{"log_output": logOutput}
	jsonData, err = json.Marshal(logData)
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation JSON du log: %w", err)
	}

	req, err = http.NewRequest(http.MethodPut, logURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la requête de log: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erreur lors de la requête HTTP de log: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
