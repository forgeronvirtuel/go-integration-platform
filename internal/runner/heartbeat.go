package runner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// StartHeartbeat envoie des heartbeats réguliers au control plane
func StartHeartbeat(controlPlaneURL string, runnerID int, stopChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Envoyer un premier heartbeat immédiatement
	SendHeartbeat(controlPlaneURL, runnerID)

	for {
		select {
		case <-ticker.C:
			SendHeartbeat(controlPlaneURL, runnerID)
		case _, ok := <-stopChan:
			if !ok {
				log.Info().Msg("Arrêt de la goroutine de heartbeat")
				return
			}
		}
	}
}

type HeartbeatResponse struct {
	Message    string `json:"message"`
	LastSeenAt string `json:"last_seen_at"`
}

// SendHeartbeat envoie un heartbeat au control plane
func SendHeartbeat(controlPlaneURL string, runnerID int) {
	url := fmt.Sprintf("%s/api/v1/runners/%d/heartbeat", controlPlaneURL, runnerID)

	log.Debug().Str("url", url).Msg("Envoi du heartbeat")

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		log.Error().Err(err).Msg("Erreur lors de l'envoi du heartbeat")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn().Int("status_code", resp.StatusCode).Msg("Code de statut inattendu pour le heartbeat")
		return
	}

	var heartbeatResp HeartbeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&heartbeatResp); err != nil {
		log.Warn().Err(err).Msg("Erreur lors de la désérialisation de la réponse heartbeat")
		return
	}

	log.Info().
		Str("message", heartbeatResp.Message).
		Str("last_seen_at", heartbeatResp.LastSeenAt).
		Msg("Heartbeat envoyé avec succès")
}
