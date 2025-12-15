package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type AgentRegistrationRequest struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	URL    string            `json:"url"`
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
	Short: "Start the agent runner",
	Long:  `Start an agent runner that registers with the control plane and execute deployment ordered by control plane.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().
			Str("control_plane", controlPlaneURL).
			Str("name", runnerName).
			Str("address", address).
			Str("port", port).
			Interface("labels", runnerLabels).
			Msg("Démarrage du runner")

		// Vérifier si un agent avec ce nom existe déjà
		agentID, err := getAgentId(controlPlaneURL, runnerName)
		if err != nil {
			log.Fatal().Err(err).Msg("Erreur lors de la vérification de l'existence du nom de l'agent")
		}
		log.Info().Int("agent_id", agentID).Msg("Vérification du nom de l'agent terminée")

		// Enregistrer l'agent
		if agentID == 0 {
			log.Info().Msg("Agent not registered, processing...")
			if agentID, err = registerAgent(controlPlaneURL, runnerName, runnerLabels, runnerURL); err != nil {
				log.Fatal().Err(err).Msg("Impossible d'enregistrer l'agent")
			}
		} else {
			// Si l'agent existe déjà, mettre à jour son URL
			if err := updateAgentURL(controlPlaneURL, agentID, runnerURL); err != nil {
				log.Warn().Err(err).Msg("Impossible de mettre à jour l'URL de l'agent")
			}
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

		// Démarrer le serveur HTTP pour recevoir les demandes de déploiement
		go startHTTPServer(agentID, controlPlaneURL, stopChan)

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

	// Determine default name (hostname)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-runner"
	}

	runnerCmd.Flags().StringVarP(&controlPlaneURL, "control-plane", "c", "", "URL of the control plane (e.g., http://localhost:3000)")
	runnerCmd.MarkFlagRequired("control-plane")

	runnerCmd.Flags().StringVarP(&runnerName, "name", "n", hostname, "Name of the agent (default is hostname)")
	runnerCmd.Flags().StringToStringVarP(&runnerLabels, "labels", "l", nil, "Agent labels (format: key1=value1,key2=value2)")
	runnerCmd.Flags().StringVarP(&port, "port", "p", "3000", "Server listening port")
	runnerCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", "./workspace", "Workspace directory for projects")
}

func getAgentId(controlPlaneURL, name string) (int, error) {
	url := fmt.Sprintf("%s/v1/api/agents/by-name/%s", controlPlaneURL, name)

	log.Debug().Str("url", url).Msg("Vérification de l'existence du nom de l'agent")
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la requête HTTP: %w", err)
	}
	defer resp.Body.Close()

	log.Debug().Str("status", resp.Status).Msg("get agent response received")
	if resp.StatusCode != http.StatusOK {
		return 0, nil
	}

	var agentResp AgentResponse
	if err := json.NewDecoder(resp.Body).Decode(&agentResp); err != nil {
		return 0, fmt.Errorf("erreur lors de la désérialisation de la réponse: %w", err)
	}

	return agentResp.ID, nil
}

// registerAgent enregistre l'agent auprès du control plane
func registerAgent(controlPlaneURL, name string, labels map[string]string, url string) (int, error) {
	urlEndpoint := fmt.Sprintf("%s/v1/api/agents/register", controlPlaneURL)

	request := AgentRegistrationRequest{
		Name:   name,
		Labels: labels,
		URL:    url,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la sérialisation JSON: %w", err)
	}

	log.Debug().Str("url", urlEndpoint).Str("body", string(jsonData)).Msg("Envoi de la requête d'enregistrement")

	resp, err := http.Post(urlEndpoint, "application/json", bytes.NewBuffer(jsonData))
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

// updateAgentURL met à jour l'URL de l'agent
func updateAgentURL(controlPlaneURL string, agentID int, url string) error {
	urlEndpoint := fmt.Sprintf("%s/v1/api/agents/%d/url", controlPlaneURL, agentID)

	urlData := map[string]string{"url": url}
	jsonData, err := json.Marshal(urlData)
	if err != nil {
		return fmt.Errorf("erreur lors de la sérialisation JSON: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, urlEndpoint, bytes.NewBuffer(jsonData))
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

// startHTTPServer démarre le serveur HTTP pour recevoir les demandes de déploiement
func startHTTPServer(agentID int, controlPlaneURL string, stopChan chan struct{}) {
	// Extract port from runnerURL
	var port string
	if len(runnerURL) > 7 && runnerURL[:7] == "http://" {
		parts := bytes.Split([]byte(runnerURL[7:]), []byte(":"))
		if len(parts) > 1 {
			port = string(parts[1])
		} else {
			port = "8080"
		}
	} else {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Endpoint pour recevoir les demandes de déploiement
	mux.HandleFunc("/deploy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var deployRequest struct {
			DeploymentID    int    `json:"deployment_id"`
			BuildID         int    `json:"build_id"`
			ControlPlaneURL string `json:"control_plane_url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&deployRequest); err != nil {
			log.Error().Err(err).Msg("Failed to decode deployment request")
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		log.Info().
			Int("deployment_id", deployRequest.DeploymentID).
			Int("build_id", deployRequest.BuildID).
			Msg("Received deployment request")

		// Lancer le déploiement dans une goroutine pour ne pas bloquer la réponse
		go executeDeployment(agentID, deployRequest.DeploymentID, deployRequest.BuildID, deployRequest.ControlPlaneURL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"message": "Deployment started"})
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Info().Str("port", port).Msg("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("HTTP server error")
		}
	}()

	<-stopChan
	log.Info().Msg("Stopping HTTP server")
	server.Close()
}

// executeDeployment télécharge et exécute un build
func executeDeployment(agentID, deploymentID, buildID int, controlPlaneURL string) {
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
	updateDeploymentStatus(controlPlaneURL, deploymentID, "deployed", "Deployment completed successfully")
	log.Info().Int("deployment_id", deploymentID).Msg("Deployment completed")
}

// updateDeploymentStatus met à jour le statut d'un déploiement
func updateDeploymentStatus(controlPlaneURL string, deploymentID int, status, logOutput string) error {
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
