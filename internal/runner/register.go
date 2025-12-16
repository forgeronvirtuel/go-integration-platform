package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

type runnerRegistrationRequest struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	URL    string            `json:"url"`
}

// RegisterRunner registers the runner with the control plane
func RegisterRunner(controlPlaneURL, name string, labels map[string]string, url string) (int, error) {
	urlEndpoint := fmt.Sprintf("%s/api/v1/runners/register", controlPlaneURL)

	request := runnerRegistrationRequest{
		Name:   name,
		Labels: labels,
		URL:    url,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return 0, fmt.Errorf("error during JSON serialization: %w", err)
	}

	log.Debug().Str("url", urlEndpoint).Str("body", string(jsonData)).Msg("Sending registration request")

	resp, err := http.Post(urlEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("error during HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var runnerResp RunnerResponse
	if err := json.NewDecoder(resp.Body).Decode(&runnerResp); err != nil {
		return 0, fmt.Errorf("error during response deserialization: %w", err)
	}

	return runnerResp.ID, nil
}
