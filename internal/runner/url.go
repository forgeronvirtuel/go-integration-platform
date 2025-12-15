package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// UpdateRunnerURL updates the URL of the runner
func UpdateRunnerURL(controlPlaneURL string, runnerID int, url string) error {
	urlEndpoint := fmt.Sprintf("%s/v1/api/runners/%d/url", controlPlaneURL, runnerID)

	urlData := map[string]string{"url": url}
	jsonData, err := json.Marshal(urlData)
	if err != nil {
		return fmt.Errorf("error during JSON serialization: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, urlEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error during request creation: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error during HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
