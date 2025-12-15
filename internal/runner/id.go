package runner

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

type RunnerResponse struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`
	Status     string            `json:"status"`
	LastSeenAt string            `json:"last_seen_at"`
	CreatedAt  string            `json:"created_at"`
}

func GetRunnerId(controlPlaneURL, name string) (int, error) {
	url := fmt.Sprintf("%s/v1/api/runners/by-name/%s", controlPlaneURL, name)

	log.Debug().Str("url", url).Msg("Checking if runner name exists")
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error during HTTP request: %w", err)
	}
	defer resp.Body.Close()

	log.Debug().Str("status", resp.Status).Msg("get runner response received")
	if resp.StatusCode != http.StatusOK {
		return 0, nil
	}

	var runnerResp RunnerResponse
	if err := json.NewDecoder(resp.Body).Decode(&runnerResp); err != nil {
		return 0, fmt.Errorf("error during response deserialization: %w", err)
	}

	return runnerResp.ID, nil
}
