package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"forgeronvirtuel/gip/internal/deployment"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type HealthHandler struct {
	DB *sql.DB
}

func (h *HealthHandler) Health(c *gin.Context) {
	if err := h.DB.Ping(); err != nil {
		log.Error().Err(err).Msg("Erreur lors du ping de la base de données")
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"error":  "database unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"database": "connected",
	})
}

// SetupControlPlaneRouter crée et configure le router Gin avec toutes les routes
func SetupControlPlaneRouter(db *sql.DB, workspace string) *gin.Engine {
	if workspace == "" {
		workspace = "./workspace"
	}
	router := gin.Default()

	router.Static("/static", "./web/static")
	router.StaticFile("/", "./web/index.html")

	api := router.Group("/api/v1/")

	// Endpoint: Health with DB check
	healthHandler := &HealthHandler{DB: db}
	api.GET("/health", healthHandler.Health)

	setupProjectRoutes(api, db)
	setupBuildRoutes(api, db, workspace)
	setupRunnerRoutes(api, db)
	setupDeploymentRoutes(api, db)
	return router
}

func StartControlPlaneServer(address, port string, db *sql.DB, workspace string) {
	gin.SetMode(gin.ReleaseMode)

	router := SetupControlPlaneRouter(db, workspace)

	log.Info().Str("address", address).Str("port", port).Msg("Control Plane HTTP server starting...")
	if err := router.Run(address + ":" + port); err != nil {
		log.Fatal().Err(err).Msg("Failed to start Control Plane HTTP server")
	}
}

func StartRunnerServer(runnerID int, runnerURL, controlPlaneURL string, stopChan chan struct{}, wg *sync.WaitGroup) {
	gin.SetMode(gin.ReleaseMode)

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
		go deployment.ExecuteDeployment(runnerID, deployRequest.DeploymentID, deployRequest.BuildID, deployRequest.ControlPlaneURL)

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
