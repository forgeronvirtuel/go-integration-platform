package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// UpdateRunnerStatus updates the status of the runner
func UpdateRunnerStatus(controlPlaneURL string, runnerID int, status string) error {
	url := fmt.Sprintf("%s/v1/api/runners/%d/status", controlPlaneURL, runnerID)

	statusData := map[string]string{"status": status}
	jsonData, err := json.Marshal(statusData)
	if err != nil {
		return fmt.Errorf("error during JSON serialization : %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error during request creation: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erreur lors de la requête HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("code de statut inattendu: %d", resp.StatusCode)
	}

	return nil
}
