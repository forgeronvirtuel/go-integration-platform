package server

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"forgeronvirtuel/gip/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type RunnerHandler struct {
	DB *sql.DB
}

type CreateRunnerRequest struct {
	Name   string            `json:"name" binding:"required"`
	Labels map[string]string `json:"labels"`
	URL    string            `json:"url"`
}

type UpdateRunnerStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=ONLINE OFFLINE DRAINING"`
}

type UpdateRunnerLabelsRequest struct {
	Labels map[string]string `json:"labels" binding:"required"`
}

type UpdateRunnerURLRequest struct {
	URL string `json:"url" binding:"required"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CreateRunner crée un nouvel runner
func (h *RunnerHandler) CreateRunner(c *gin.Context) {
	var req CreateRunnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Vérifier si uwn runner avec ce nom existe déjà
	existingRunner, err := database.GetRunnerByName(h.DB, req.Name)
	if err == nil && existingRunner != nil {
		c.JSON(http.StatusConflict, APIError{Code: ERR_CDE_AGENT_NAME_EXISTS, Message: "Runner with this name already exists"})
		return
	}

	// Créer l'runner
	runner, err := database.CreateRunner(h.DB, req.Name, req.Labels)
	if err != nil {
		log.Error().Err(err).Msg("Erreur lors de la création de l'runner")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create runner"})
		return
	}

	// Mettre à jour l'URL si fournie
	if req.URL != "" {
		if err := database.UpdateRunnerURL(h.DB, runner.ID, req.URL); err != nil {
			log.Warn().Err(err).Int("runner_id", runner.ID).Msg("Failed to update runner URL")
		}
		runner.URL = req.URL
	}

	log.Info().Int("runner_id", runner.ID).Str("name", runner.Name).Msg("Runner créé avec succès")
	c.JSON(http.StatusCreated, runner)
}

// GetRunner récupère un runner par ID
func (h *RunnerHandler) GetRunner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner ID"})
		return
	}

	runner, err := database.GetRunnerByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
		return
	}

	c.JSON(http.StatusOK, runner)
}

// GetRunner récupère un runner par ID
func (h *RunnerHandler) GetRunnerByName(c *gin.Context) {
	name := c.Param("name")

	runner, err := database.GetRunnerByName(h.DB, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
		return
	}

	c.JSON(http.StatusOK, runner)
}

// GetAllRunners récupère tous les runners
func (h *RunnerHandler) GetAllRunners(c *gin.Context) {
	// Optionnel: filtrer par statut
	status := c.Query("status")

	var runners []database.Runner
	var err error

	if status != "" {
		runners, err = database.GetRunnersByStatus(h.DB, status)
	} else {
		runners, err = database.GetAllRunners(h.DB)
	}

	if err != nil {
		log.Error().Err(err).Msg("failed to fetch runners")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch runners"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"runners": runners,
		"count":   len(runners),
	})
}

// UpdateRunnerStatus met à jour le statut d'un runner
func (h *RunnerHandler) UpdateRunnerStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner ID"})
		return
	}

	var req UpdateRunnerStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Vérifier que l'runner existe
	runner, err := database.GetRunnerByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
		return
	}

	// Mettre à jour le statut
	if err := database.UpdateRunnerStatus(h.DB, id, req.Status); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la mise à jour du statut de l'runner")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update runner status"})
		return
	}

	log.Info().Int("runner_id", id).Str("old_status", runner.Status).Str("new_status", req.Status).Msg("Statut de l'runner mis à jour")

	// Récupérer l'runner mis à jour
	updatedRunner, _ := database.GetRunnerByID(h.DB, id)
	c.JSON(http.StatusOK, updatedRunner)
}

// UpdateRunnerLabels met à jour les labels d'un runner
func (h *RunnerHandler) UpdateRunnerLabels(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner ID"})
		return
	}

	var req UpdateRunnerLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Vérifier que l'runner existe
	_, err = database.GetRunnerByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
		return
	}

	// Mettre à jour les labels
	if err := database.UpdateRunnerLabels(h.DB, id, req.Labels); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la mise à jour des labels de l'runner")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update runner labels"})
		return
	}

	log.Info().Int("runner_id", id).Msg("Labels de l'runner mis à jour")

	// Récupérer l'runner mis à jour
	updatedRunner, _ := database.GetRunnerByID(h.DB, id)
	c.JSON(http.StatusOK, updatedRunner)
}

// UpdateRunnerURL met à jour l'URL d'un runner
func (h *RunnerHandler) UpdateRunnerURL(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner ID"})
		return
	}

	var req UpdateRunnerURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Vérifier que l'runner existe
	_, err = database.GetRunnerByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
		return
	}

	// Mettre à jour l'URL
	if err := database.UpdateRunnerURL(h.DB, id, req.URL); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la mise à jour de l'URL de l'runner")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update runner URL"})
		return
	}

	log.Info().Int("runner_id", id).Str("url", req.URL).Msg("URL de l'runner mis à jour")

	// Récupérer l'runner mis à jour
	updatedRunner, _ := database.GetRunnerByID(h.DB, id)
	c.JSON(http.StatusOK, updatedRunner)
}

// Heartbeat enregistre un heartbeat pour un runner
func (h *RunnerHandler) Heartbeat(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner ID"})
		return
	}

	// Vérifier que l'runner existe
	runner, err := database.GetRunnerByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
		return
	}

	// Mettre à jour le heartbeat
	if err := database.UpdateRunnerHeartbeat(h.DB, id); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la mise à jour du heartbeat de l'runner")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update heartbeat"})
		return
	}

	// Si l'runner était OFFLINE, le passer à ONLINE
	if runner.Status == "OFFLINE" {
		if err := database.UpdateRunnerStatus(h.DB, id, "ONLINE"); err != nil {
			log.Error().Err(err).Msg("Erreur lors de la mise à jour du statut de l'runner")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Heartbeat registered",
		"last_seen_at": time.Now(),
	})
}

// DeleteRunner supprime un runner
func (h *RunnerHandler) DeleteRunner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner ID"})
		return
	}

	// Vérifier que l'runner existe
	_, err = database.GetRunnerByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
		return
	}

	// Supprimer l'runner
	if err := database.DeleteRunner(h.DB, id); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la suppression de l'runner")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete runner"})
		return
	}

	log.Info().Int("runner_id", id).Msg("Runner supprimé avec succès")
	c.JSON(http.StatusOK, gin.H{"message": "Runner deleted successfully"})
}

// setupRunnerRoutes configure endpoints for managing runners (old name: runners)
func setupRunnerRoutes(api *gin.RouterGroup, db *sql.DB) {
	handler := &RunnerHandler{DB: db}
	runners := api.Group("/runners")
	{
		runners.POST("/register", handler.CreateRunner)        // Créer un runner
		runners.GET("", handler.GetAllRunners)                 // Lister tous les runners (avec filtre status optionnel)
		runners.GET("/:id", handler.GetRunner)                 // Récupérer un runner par ID
		runners.GET("/by-name/:name", handler.GetRunnerByName) // Récupérer un runner par nom
		runners.PUT("/:id/status", handler.UpdateRunnerStatus) // Mettre à jour le statut
		runners.PUT("/:id/labels", handler.UpdateRunnerLabels) // Mettre à jour les labels
		runners.PUT("/:id/url", handler.UpdateRunnerURL)       // Mettre à jour l'URL
		runners.POST("/:id/heartbeat", handler.Heartbeat)      // Heartbeat
		runners.DELETE("/:id", handler.DeleteRunner)           // Supprimer un runner
	}
}
