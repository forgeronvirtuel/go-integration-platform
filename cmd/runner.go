package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type runnerRegistrationRequest struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	URL    string            `json:"url"`
}

type runnerResponse struct {
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
	Short: "Start the runner runner",
	Long:  `Start an runner runner that registers with the control plane and execute deployment ordered by control plane.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().
			Str("control_plane", controlPlaneURL).
			Str("name", runnerName).
			Str("address", address).
			Str("port", port).
			Interface("labels", runnerLabels).
			Msg("Starting runner")

		// Check if a runner with this name already exists
		runnerID, err := getRunnerId(controlPlaneURL, runnerName)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get runner by name")
		}
		log.Info().Int("runner_id", runnerID).Msg("Runner name verification completed")

		// Register the runner
		if runnerID == 0 {
			log.Info().Msg("runner not registered, processing...")
			if runnerID, err = registerRunner(controlPlaneURL, runnerName, runnerLabels, runnerURL); err != nil {
				log.Fatal().Err(err).Msg("failed to register runner")
			}
		} else {
			// If the runner already exists, update its URL
			if err := updateRunnerURL(controlPlaneURL, runnerID, runnerURL); err != nil {
				log.Warn().Err(err).Msg("failed to update runner URL")
			}
		}

		log.Info().Int("runner_id", runnerID).Msg("runner sucessfully registered")

		// Set the runner to ONLINE
		if err := updateRunnerStatus(controlPlaneURL, runnerID, "ONLINE"); err != nil {
			log.Warn().Err(err).Msg("failed to set runner status to ONLINE")
		} else {
			log.Info().Msg("runner set to ONLINE")
		}

		// Start the heartbeat goroutine
		stopChan := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)
		go startHeartbeat(controlPlaneURL, runnerID, stopChan, &wg)

		// Start the HTTP server to receive deployment requests
		wg.Add(1)
		go startHTTPServer(runnerID, controlPlaneURL, stopChan, &wg)

		// Wait for a stop signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		<-sigChan
		log.Info().Msg("stop signal received, stopping runner...")

		// Stop the heartbeat goroutine
		close(stopChan)

		// Set the runner to OFFLINE before exiting
		if err := updateRunnerStatus(controlPlaneURL, runnerID, "OFFLINE"); err != nil {
			log.Warn().Err(err).Msg("failed to set runner status to OFFLINE")
		} else {
			log.Info().Msg("runner set to OFFLINE")
		}

		wg.Wait()

		log.Info().Msg("Runner stopped gracefully")
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

	runnerCmd.Flags().StringVarP(&runnerName, "name", "n", hostname, "Name of the runner (default is hostname)")
	runnerCmd.Flags().StringVarP(&runnerURL, "url", "u", "http://localhost:3000", "URL of the runner (e.g., http://my-runner:3000)")
	runnerCmd.Flags().StringToStringVarP(&runnerLabels, "labels", "l", nil, "runner labels (format: key1=value1,key2=value2)")
	runnerCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", "./workspace", "Workspace directory for projects")
}

func getRunnerId(controlPlaneURL, name string) (int, error) {
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

	var runnerResp runnerResponse
	if err := json.NewDecoder(resp.Body).Decode(&runnerResp); err != nil {
		return 0, fmt.Errorf("error during response deserialization: %w", err)
	}

	return runnerResp.ID, nil
}

// registerRunner registers the runner with the control plane
func registerRunner(controlPlaneURL, name string, labels map[string]string, url string) (int, error) {
	urlEndpoint := fmt.Sprintf("%s/v1/api/runners/register", controlPlaneURL)

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

	var runnerResp runnerResponse
	if err := json.NewDecoder(resp.Body).Decode(&runnerResp); err != nil {
		return 0, fmt.Errorf("error during response deserialization: %w", err)
	}

	return runnerResp.ID, nil
}

// updateRunnerStatus updates the status of the runner
func updateRunnerStatus(controlPlaneURL string, runnerID int, status string) error {
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

// startHeartbeat envoie des heartbeats réguliers au control plane
func startHeartbeat(controlPlaneURL string, runnerID int, stopChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Envoyer un premier heartbeat immédiatement
	sendHeartbeat(controlPlaneURL, runnerID)

	for {
		select {
		case <-ticker.C:
			sendHeartbeat(controlPlaneURL, runnerID)
		case <-stopChan:
			log.Info().Msg("Arrêt de la goroutine de heartbeat")
			return
		}
	}
}

// sendHeartbeat envoie un heartbeat au control plane
func sendHeartbeat(controlPlaneURL string, runnerID int) {
	url := fmt.Sprintf("%s/v1/api/runners/%d/heartbeat", controlPlaneURL, runnerID)

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

// updateRunnerURL updates the URL of the runner
func updateRunnerURL(controlPlaneURL string, runnerID int, url string) error {
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

// startHTTPServer démarre le serveur HTTP pour recevoir les demandes de déploiement
func startHTTPServer(runnerID int, controlPlaneURL string, stopChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// Extract port from runnerURL
	port := "8080"
	if len(runnerURL) > 7 && runnerURL[:7] == "http://" {
		parts := bytes.Split([]byte(runnerURL[7:]), []byte(":"))
		if len(parts) > 1 {
			port = string(parts[1])
		}
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
		go executeDeployment(runnerID, deployRequest.DeploymentID, deployRequest.BuildID, deployRequest.ControlPlaneURL)

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

	timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(timeoutContext); err != nil {
		log.Error().Err(err).Msg("Error during HTTP server shutdown")
	} else {
		log.Info().Msg("HTTP server stopped gracefully")
	}
}

// executeDeployment télécharge et exécute un build
func executeDeployment(runnerID, deploymentID, buildID int, controlPlaneURL string) {
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
