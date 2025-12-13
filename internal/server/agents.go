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

type AgentHandler struct {
	DB *sql.DB
}

type CreateAgentRequest struct {
	Name   string            `json:"name" binding:"required"`
	Labels map[string]string `json:"labels"`
}

type UpdateAgentStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=ONLINE OFFLINE DRAINING"`
}

type UpdateAgentLabelsRequest struct {
	Labels map[string]string `json:"labels" binding:"required"`
}

// CreateAgent crée un nouvel agent
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Vérifier si un agent avec ce nom existe déjà
	existingAgent, err := database.GetAgentByName(h.DB, req.Name)
	if err == nil && existingAgent != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Agent with this name already exists"})
		return
	}

	// Créer l'agent
	agent, err := database.CreateAgent(h.DB, req.Name, req.Labels)
	if err != nil {
		log.Error().Err(err).Msg("Erreur lors de la création de l'agent")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent"})
		return
	}

	log.Info().Int("agent_id", agent.ID).Str("name", agent.Name).Msg("Agent créé avec succès")
	c.JSON(http.StatusCreated, agent)
}

// GetAgent récupère un agent par ID
func (h *AgentHandler) GetAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	agent, err := database.GetAgentByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// GetAllAgents récupère tous les agents
func (h *AgentHandler) GetAllAgents(c *gin.Context) {
	// Optionnel: filtrer par statut
	status := c.Query("status")

	var agents []database.Agent
	var err error

	if status != "" {
		agents, err = database.GetAgentsByStatus(h.DB, status)
	} else {
		agents, err = database.GetAllAgents(h.DB)
	}

	if err != nil {
		log.Error().Err(err).Msg("Erreur lors de la récupération des agents")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"count":  len(agents),
	})
}

// // UpdateAgentStatus met à jour le statut d'un agent
// func (h *AgentHandler) UpdateAgentStatus(c *gin.Context) {
// 	idStr := c.Param("id")
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
// 		return
// 	}

// 	var req UpdateAgentStatusRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
// 		return
// 	}

// 	// Vérifier que l'agent existe
// 	agent, err := database.GetAgentByID(h.DB, id)
// 	if err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
// 		return
// 	}

// 	// Mettre à jour le statut
// 	if err := database.UpdateAgentStatus(h.DB, id, req.Status); err != nil {
// 		log.Error().Err(err).Msg("Erreur lors de la mise à jour du statut de l'agent")
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent status"})
// 		return
// 	}

// 	log.Info().Int("agent_id", id).Str("old_status", agent.Status).Str("new_status", req.Status).Msg("Statut de l'agent mis à jour")

// 	// Récupérer l'agent mis à jour
// 	updatedAgent, _ := database.GetAgentByID(h.DB, id)
// 	c.JSON(http.StatusOK, updatedAgent)
// }

// UpdateAgentLabels met à jour les labels d'un agent
func (h *AgentHandler) UpdateAgentLabels(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	var req UpdateAgentLabelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Vérifier que l'agent existe
	_, err = database.GetAgentByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Mettre à jour les labels
	if err := database.UpdateAgentLabels(h.DB, id, req.Labels); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la mise à jour des labels de l'agent")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent labels"})
		return
	}

	log.Info().Int("agent_id", id).Msg("Labels de l'agent mis à jour")

	// Récupérer l'agent mis à jour
	updatedAgent, _ := database.GetAgentByID(h.DB, id)
	c.JSON(http.StatusOK, updatedAgent)
}

// Heartbeat enregistre un heartbeat pour un agent
func (h *AgentHandler) Heartbeat(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	// Vérifier que l'agent existe
	agent, err := database.GetAgentByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Mettre à jour le heartbeat
	if err := database.UpdateAgentHeartbeat(h.DB, id); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la mise à jour du heartbeat de l'agent")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update heartbeat"})
		return
	}

	// Si l'agent était OFFLINE, le passer à ONLINE
	if agent.Status == "OFFLINE" {
		if err := database.UpdateAgentStatus(h.DB, id, "ONLINE"); err != nil {
			log.Error().Err(err).Msg("Erreur lors de la mise à jour du statut de l'agent")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Heartbeat registered",
		"last_seen_at": time.Now(),
	})
}

// DeleteAgent supprime un agent
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	// Vérifier que l'agent existe
	_, err = database.GetAgentByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Supprimer l'agent
	if err := database.DeleteAgent(h.DB, id); err != nil {
		log.Error().Err(err).Msg("Erreur lors de la suppression de l'agent")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agent"})
		return
	}

	log.Info().Int("agent_id", id).Msg("Agent supprimé avec succès")
	c.JSON(http.StatusOK, gin.H{"message": "Agent deleted successfully"})
}

// setupAgentRoutes configure les routes pour les agents
func setupAgentRoutes(router gin.IRouter, db *sql.DB) {
	handler := &AgentHandler{DB: db}
	agents := router.Group("/api/agents")
	{
		agents.POST("/register", handler.CreateAgent) // Créer un agent
		agents.GET("", handler.GetAllAgents)          // Lister tous les agents (avec filtre status optionnel)
		agents.GET("/:id", handler.GetAgent)          // Récupérer un agent par ID
		// agents.PUT("/:id/status", handler.UpdateAgentStatus) // Mettre à jour le statut
		agents.PUT("/:id/labels", handler.UpdateAgentLabels) // Mettre à jour les labels
		agents.POST("/:id/heartbeat", handler.Heartbeat)     // Heartbeat
		agents.DELETE("/:id", handler.DeleteAgent)           // Supprimer un agent
	}
}
