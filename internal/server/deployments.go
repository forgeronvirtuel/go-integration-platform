package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"forgeronvirtuel/gip/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type DeploymentHandler struct {
	DB *sql.DB
}

type CreateDeploymentRequest struct {
	BuildID int  `json:"build_id" binding:"required"`
	AgentID *int `json:"agent_id,omitempty"`
}

type UpdateDeploymentStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending deploying deployed failed"`
}

type UpdateDeploymentAgentRequest struct {
	AgentID *int `json:"agent_id"`
}

type UpdateDeploymentLogRequest struct {
	LogOutput string `json:"log_output" binding:"required"`
}

// CreateDeployment creates a new deployment.
// If an agent ID is provided, it checks that the agent exists but ignore its status.
func (h *DeploymentHandler) CreateDeployment(c *gin.Context) {
	var req CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check that the build exists and is successful
	build, err := database.GetBuildByID(h.DB, strconv.Itoa(req.BuildID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Build not found"})
		return
	}

	if build.Status != "success" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Build must be successful to deploy"})
		return
	}

	// If an agent ID is provided, check that the agent exists
	if req.AgentID != nil {
		_, err := database.GetAgentByID(h.DB, *req.AgentID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
	}

	// Push the deployment to the database
	deployment, err := database.CreateDeployment(h.DB, req.BuildID, req.AgentID)
	if err != nil {
		log.Error().Err(err).Msg("Error creating deployment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deployment"})
		return
	}

	log.Info().
		Int("deployment_id", deployment.ID).
		Int("build_id", req.BuildID).
		Msg("DDeployment created successfully")

	c.JSON(http.StatusCreated, deployment)
}

// GetDeployment get a deployment by ID
func (h *DeploymentHandler) GetDeployment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	deployment, err := database.GetDeploymentByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}

	c.JSON(http.StatusOK, deployment)
}

type GetAllDeploymentsResponse struct {
	Deployments []database.Deployment `json:"deployments"`
	Count       int                   `json:"count"`
}

func (h *DeploymentHandler) GetAllDeployments(c *gin.Context) {
	deployments, err := database.GetAllDeployments(h.DB)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching deployments")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deployments"})
		return
	}

	c.JSON(http.StatusOK, GetAllDeploymentsResponse{
		Deployments: deployments,
		Count:       len(deployments),
	})
}

// GetDeploymentsByBuildID gets all deployments for a build
func (h *DeploymentHandler) GetDeploymentsByBuildID(c *gin.Context) {
	buildIDStr := c.Query("build_id")
	if buildIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "build_id query parameter required"})
		return
	}

	buildID, err := strconv.Atoi(buildIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid build_id"})
		return
	}

	deployments, err := database.GetDeploymentsByBuildID(h.DB, buildID)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching deployments")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deployments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deployments": deployments,
		"count":       len(deployments),
	})
}

// GetDeploymentsByAgentID gets all deployments for an agent
func (h *DeploymentHandler) GetDeploymentsByAgentID(c *gin.Context) {
	agentIDStr := c.Query("agent_id")
	if agentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id query parameter required"})
		return
	}

	agentID, err := strconv.Atoi(agentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent_id"})
		return
	}

	deployments, err := database.GetDeploymentsByAgentID(h.DB, agentID)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching deployments")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deployments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deployments": deployments,
		"count":       len(deployments),
	})
}

// UpdateDeploymentStatus updates the status of a deployment
func (h *DeploymentHandler) UpdateDeploymentStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	var req UpdateDeploymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	deployment, err := database.UpdateDeploymentStatus(h.DB, id, req.Status)
	if err != nil {
		log.Error().Err(err).Msg("Error updating deployment status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deployment status"})
		return
	}

	log.Info().
		Int("deployment_id", id).
		Str("status", req.Status).
		Msg("Deployment status updated")

	c.JSON(http.StatusOK, deployment)
}

// UpdateDeploymentAgent updates the agent of a deployment
// If an agent ID is provided, it checks that the agent exists but ignore its status.
func (h *DeploymentHandler) UpdateDeploymentAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	var req UpdateDeploymentAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check that the agent exists if provided
	if req.AgentID != nil {
		_, err := database.GetAgentByID(h.DB, *req.AgentID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
	}

	deployment, err := database.UpdateDeploymentAgent(h.DB, id, req.AgentID)
	if err != nil {
		log.Error().Err(err).Msg("Error updating deployment agent")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deployment agent"})
		return
	}

	log.Info().
		Int("deployment_id", id).
		Msg("Deployment agent updated")

	c.JSON(http.StatusOK, deployment)
}

// UpdateDeploymentLog updates the log output of a deployment
func (h *DeploymentHandler) UpdateDeploymentLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	var req UpdateDeploymentLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	err = database.UpdateDeploymentLog(h.DB, id, req.LogOutput)
	if err != nil {
		log.Error().Err(err).Msg("Error updating deployment log")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deployment log"})
		return
	}

	deployment, err := database.GetDeploymentByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated deployment"})
		return
	}

	c.JSON(http.StatusOK, deployment)
}

// DeleteDeployment deletes a deployment
func (h *DeploymentHandler) DeleteDeployment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	err = database.DeleteDeployment(h.DB, id)
	if err != nil {
		log.Error().Err(err).Msg("Error deleting deployment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete deployment"})
		return
	}

	log.Info().Int("deployment_id", id).Msg("Deployment deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Deployment deleted successfully"})
}

// ExecuteDeployment triggers the execution of a deployment by sending a request to the runner
func (h *DeploymentHandler) ExecuteDeployment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	// Get the deployment
	deployment, err := database.GetDeploymentByID(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}

	// Check that the deployment has an agent assigned
	if deployment.AgentID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Deployment has no agent assigned"})
		return
	}

	// Get the agent to retrieve its URL
	agent, err := database.GetAgentByID(h.DB, *deployment.AgentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// Check that the agent has a URL configured
	if agent.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Agent has no URL configured"})
		return
	}

	// Check that the agent is online
	if agent.Status != "ONLINE" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Agent is not online"})
		return
	}

	// Update deployment status to deploying
	_, err = database.UpdateDeploymentStatus(h.DB, id, "deploying")
	if err != nil {
		log.Error().Err(err).Int("deployment_id", id).Msg("Failed to update deployment status to deploying")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deployment status"})
		return
	}

	// Prepare the request to send to the runner
	// Get the control plane URL from the request (we'll use the Host header)
	controlPlaneURL := fmt.Sprintf("http://%s", c.Request.Host)

	requestPayload := map[string]interface{}{
		"deployment_id":     id,
		"build_id":          deployment.BuildID,
		"control_plane_url": controlPlaneURL,
	}

	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal request payload")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare request"})
		return
	}

	// Send the request to the runner
	runnerURL := fmt.Sprintf("%s/deploy", agent.URL)
	resp, err := http.Post(runnerURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().Err(err).Str("runner_url", runnerURL).Msg("Failed to send request to runner")
		// Revert deployment status to pending
		database.UpdateDeploymentStatus(h.DB, id, "pending")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to contact runner"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		log.Error().Int("status_code", resp.StatusCode).Str("runner_url", runnerURL).Msg("Runner returned error")
		// Revert deployment status to pending
		database.UpdateDeploymentStatus(h.DB, id, "pending")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Runner rejected the request"})
		return
	}

	log.Info().Int("deployment_id", id).Int("agent_id", *deployment.AgentID).Msg("Deployment execution started")
	c.JSON(http.StatusOK, gin.H{"message": "Deployment execution started"})
}

// setupDeploymentRoutes  configure the routes for deployments
func setupDeploymentRoutes(v1 gin.IRouter, db *sql.DB) {
	handler := &DeploymentHandler{DB: db}

	deployments := v1.Group("/deployments")
	{
		deployments.POST("", handler.CreateDeployment)
		deployments.GET("", handler.GetAllDeployments)
		deployments.GET("/by-build", handler.GetDeploymentsByBuildID)
		deployments.GET("/by-agent", handler.GetDeploymentsByAgentID)
		deployments.GET("/:id", handler.GetDeployment)
		deployments.PUT("/:id/status", handler.UpdateDeploymentStatus)
		deployments.PUT("/:id/agent", handler.UpdateDeploymentAgent)
		deployments.PUT("/:id/log", handler.UpdateDeploymentLog)
		deployments.POST("/:id/execute", handler.ExecuteDeployment)
		deployments.DELETE("/:id", handler.DeleteDeployment)
	}
}
