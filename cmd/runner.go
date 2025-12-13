package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	controlPlaneURL string
	runnerName      string
	runnerLabels    map[string]string
)

type AgentRegistrationRequest struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type AgentResponse struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Labels     map[string]string `json:"labels"`
	Status     string            `json:"status"`
	LastSeenAt string            `json:"last_seen_at"`
	CreatedAt  string            `json:"created_at"`
}

type HeartbeatResponse struct {
	Message    string `json:"message"`
	LastSeenAt string `json:"last_seen_at"`
}

var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "Démarre un agent runner",
	Long:  `Démarre un agent runner qui s'enregistre auprès du control plane et envoie des heartbeats réguliers.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().
			Str("control_plane", controlPlaneURL).
			Str("name", runnerName).
			Interface("labels", runnerLabels).
			Msg("Démarrage du runner")

		// Enregistrer l'agent
		agentID, err := registerAgent(controlPlaneURL, runnerName, runnerLabels)
		if err != nil {
			log.Fatal().Err(err).Msg("Impossible d'enregistrer l'agent")
		}

		log.Info().Int("agent_id", agentID).Msg("Agent enregistré avec succès")

		// Mettre l'agent en ONLINE
		if err := updateAgentStatus(controlPlaneURL, agentID, "ONLINE"); err != nil {
			log.Warn().Err(err).Msg("Impossible de mettre l'agent en ONLINE")
		} else {
			log.Info().Msg("Agent mis en ONLINE")
		}

		// Démarrer la goroutine de heartbeat
		stopChan := make(chan struct{})
		go startHeartbeat(controlPlaneURL, agentID, stopChan)

		// Attendre un signal d'arrêt
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		<-sigChan
		log.Info().Msg("Signal d'arrêt reçu, arrêt du runner...")

		// Arrêter la goroutine de heartbeat
		close(stopChan)

		// Mettre l'agent en OFFLINE avant de quitter
		if err := updateAgentStatus(controlPlaneURL, agentID, "OFFLINE"); err != nil {
			log.Warn().Err(err).Msg("Impossible de mettre l'agent en OFFLINE")
		} else {
			log.Info().Msg("Agent mis en OFFLINE")
		}

		log.Info().Msg("Runner arrêté proprement")
	},
}

func init() {
	rootCmd.AddCommand(runnerCmd)

	// Déterminer le nom par défaut (hostname)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-runner"
	}

	// Déterminer les labels par défaut
	defaultLabels := map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	}

	runnerCmd.Flags().StringVarP(&controlPlaneURL, "control-plane", "c", "", "URL du control plane (ex: http://localhost:3000)")
	runnerCmd.MarkFlagRequired("control-plane")

	runnerCmd.Flags().StringVarP(&runnerName, "name", "n", hostname, "Nom de l'agent (hostname par défaut)")
	runnerCmd.Flags().StringToStringVarP(&runnerLabels, "labels", "l", defaultLabels, "Labels de l'agent (format: key1=value1,key2=value2)")
}

// registerAgent enregistre l'agent auprès du control plane
func registerAgent(controlPlaneURL, name string, labels map[string]string) (int, error) {
	url := fmt.Sprintf("%s/v1/api/agents", controlPlaneURL)

	request := AgentRegistrationRequest{
		Name:   name,
		Labels: labels,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la sérialisation JSON: %w", err)
	}

	log.Debug().Str("url", url).Str("body", string(jsonData)).Msg("Envoi de la requête d'enregistrement")

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la requête HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("code de statut inattendu: %d", resp.StatusCode)
	}

	var agentResp AgentResponse
	if err := json.NewDecoder(resp.Body).Decode(&agentResp); err != nil {
		return 0, fmt.Errorf("erreur lors de la désérialisation de la réponse: %w", err)
	}

	return agentResp.ID, nil
}

// updateAgentStatus met à jour le statut de l'agent
func updateAgentStatus(controlPlaneURL string, agentID int, status string) error {
	url := fmt.Sprintf("%s/v1/api/agents/%d/status", controlPlaneURL, agentID)

	statusData := map[string]string{"status": status}
	jsonData, err := json.Marshal(statusData)
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation JSON: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la requête: %w", err)
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

// startHeartbeat envoie des heartbeats réguliers au control plane
func startHeartbeat(controlPlaneURL string, agentID int, stopChan chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Envoyer un premier heartbeat immédiatement
	sendHeartbeat(controlPlaneURL, agentID)

	for {
		select {
		case <-ticker.C:
			sendHeartbeat(controlPlaneURL, agentID)
		case <-stopChan:
			log.Info().Msg("Arrêt de la goroutine de heartbeat")
			return
		}
	}
}

// sendHeartbeat envoie un heartbeat au control plane
func sendHeartbeat(controlPlaneURL string, agentID int) {
	url := fmt.Sprintf("%s/v1/api/agents/%d/heartbeat", controlPlaneURL, agentID)

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
