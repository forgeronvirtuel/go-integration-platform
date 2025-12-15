package deployment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// executeDeployment télécharge et exécute un build
func ExecuteDeployment(runnerID, deploymentID, buildID int, controlPlaneURL string) {
	log.Info().
		Int("deployment_id", deploymentID).
		Int("build_id", buildID).
		Msg("Starting deployment execution")

	// TODO: Implémenter le téléchargement et l'exécution du build
	// 1. Télécharger le binaire depuis le control plane
	// 2. Exécuter le binaire
	// 3. Capturer les logs
	// 4. Envoyer les mises à jour de statut au control plane

	// Pour l'instant, on simule un déploiement réussi
	time.Sleep(5 * time.Second)

	// Mettre à jour le statut du déploiement
	UpdateDeploymentStatus(controlPlaneURL, deploymentID, "deployed", "Deployment completed successfully")
	log.Info().Int("deployment_id", deploymentID).Msg("Deployment completed")
}

// UpdateDeploymentStatus met à jour le statut d'un déploiement
func UpdateDeploymentStatus(controlPlaneURL string, deploymentID int, status, logOutput string) error {
	// Update status
	statusURL := fmt.Sprintf("%s/v1/api/deployments/%d/status", controlPlaneURL, deploymentID)
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
	logURL := fmt.Sprintf("%s/v1/api/deployments/%d/log", controlPlaneURL, deploymentID)
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
