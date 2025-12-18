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
	BuildID  int  `json:"build_id" binding:"required"`
	RunnerID *int `json:"runner_id,omitempty"`
}

type UpdateDeploymentStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending deploying deployed failed"`
}

type UpdateDeploymentRunnerRequest struct {
	RunnerID *int `json:"runner_id"`
}

type UpdateDeploymentLogRequest struct {
	LogOutput string `json:"log_output" binding:"required"`
}

// CreateDeployment creates a new deployment.
// If an runner ID is provided, it checks that the runner exists but ignore its status.
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

	// If an runner ID is provided, check that the runner exists
	if req.RunnerID != nil {
		_, err := database.GetRunnerByID(h.DB, *req.RunnerID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
			return
		}
	}

	// Push the deployment to the database
	deployment, err := database.CreateDeployment(h.DB, req.BuildID, req.RunnerID)
	if err != nil {
		log.Error().Err(err).Msg("Error creating deployment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deployment"})
		return
	}

	log.Info().
		Int("deployment_id", deployment.ID).
		Int("build_id", req.BuildID).
		Msg("Deployment created successfully")

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
	log.Info().Msg("Fetching all deployments")
	deployments, err := database.GetAllDeployments(h.DB)
	log.Info().Msgf("Fetched %d deployments", len(deployments))
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

// GetDeploymentsByRunnerID gets all deployments for an runner
func (h *DeploymentHandler) GetDeploymentsByRunnerID(c *gin.Context) {
	runnerIDStr := c.Query("runner_id")
	if runnerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "runner_id query parameter required"})
		return
	}

	runnerID, err := strconv.Atoi(runnerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner_id"})
		return
	}

	deployments, err := database.GetDeploymentsByRunnerID(h.DB, runnerID)
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

// UpdateDeploymentRunner updates the runner of a deployment
// If an runner ID is provided, it checks that the runner exists but ignore its status.
func (h *DeploymentHandler) UpdateDeploymentRunner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	var req UpdateDeploymentRunnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check that the runner exists if provided
	if req.RunnerID != nil {
		_, err := database.GetRunnerByID(h.DB, *req.RunnerID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
			return
		}
	}

	deployment, err := database.UpdateDeploymentRunner(h.DB, id, req.RunnerID)
	if err != nil {
		log.Error().Err(err).Msg("Error updating deployment runner")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deployment runner"})
		return
	}

	log.Info().
		Int("deployment_id", id).
		Msg("Deployment runner updated")

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
		log.Err(err).Str("id_str", c.Param("id")).Msg("Invalid deployment ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}
	log.Info().Int("deployment_id", id).Msg("Executing deployment")

	// Get the deployment
	deployment, err := database.GetDeploymentByID(h.DB, id)
	if err != nil {
		log.Err(err).Int("deployment_id", id).Msg("Deployment not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}

	// Check that the deployment has a runner assigned
	if deployment.RunnerID == nil {
		log.Warn().Int("deployment_id", id).Msg("Deployment has no runner assigned")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Deployment has no runner assigned"})
		return
	}

	// Get the runner to retrieve its URL
	runner, err := database.GetRunnerByID(h.DB, *deployment.RunnerID)
	if err != nil {
		log.Err(err).Int("runner_id", *deployment.RunnerID).Msg("Runner not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
		return
	}
	log.Info().Int("runner_id", runner.ID).Str("runner_name", runner.Name).Msg("Runner retrieved for deployment")

	// Check that the runner has a URL configured
	if runner.URL == "" {
		log.Warn().Int("runner_id", runner.ID).Msg("Runner has no URL configured")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Runner has no URL configured"})
		return
	}
	log.Info().Int("runner_id", runner.ID).Str("runner_url", runner.URL).Msg("Runner URL verified")

	// Check that the runner is online
	if runner.Status != "ONLINE" {
		log.Warn().Int("runner_id", runner.ID).Str("runner_status", runner.Status).Msg("Runner is not online")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Runner is not online"})
		return
	}
	log.Info().Int("runner_id", runner.ID).Msg("Runner is online")

	// Update deployment status to deploying
	_, err = database.UpdateDeploymentStatus(h.DB, id, "deploying")
	if err != nil {
		log.Error().Err(err).Int("deployment_id", id).Msg("Failed to update deployment status to deploying")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deployment status"})
		return
	}
	log.Info().Int("deployment_id", id).Msg("Deployment status updated to deploying")

	// Prepare the request to send to the runner
	// Get the control plane URL from the request (we'll use the Host header)
	controlPlaneURL := fmt.Sprintf("http://%s", c.Request.Host)

	requestPayload := map[string]any{
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
	log.Info().Int("deployment_id", id).Msg("Request payload prepared for runner")

	// Send the request to the runner
	runnerURL := fmt.Sprintf("%s/api/v1/deploy", runner.URL)
	resp, err := http.Post(runnerURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().Err(err).Str("runner_url", runnerURL).Msg("Failed to send request to runner")
		// Revert deployment status to pending
		database.UpdateDeploymentStatus(h.DB, id, "pending")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to contact runner"})
		return
	}
	log.Info().Int("deployment_id", id).Str("runner_url", runnerURL).Msg("Request sent to runner successfully")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		log.Error().Int("status_code", resp.StatusCode).Str("runner_url", runnerURL).Msg("Runner returned error")
		// Revert deployment status to pending
		database.UpdateDeploymentStatus(h.DB, id, "pending")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Runner rejected the request"})
		return
	}

	log.Info().Int("deployment_id", id).Int("runner_id", *deployment.RunnerID).Msg("Deployment execution started")
	c.JSON(http.StatusOK, gin.H{"message": "Deployment execution started"})
}

// setupDeploymentRoutes  configure endpoints for deployments
func setupDeploymentRoutes(api *gin.RouterGroup, db *sql.DB) {
	handler := &DeploymentHandler{DB: db}

	deployments := api.Group("/deployments")
	{
		deployments.POST("", handler.CreateDeployment)
		deployments.GET("", handler.GetAllDeployments)
		deployments.GET("/by-build", handler.GetDeploymentsByBuildID)
		deployments.GET("/by-runner", handler.GetDeploymentsByRunnerID)
		deployments.GET("/:id", handler.GetDeployment)
		deployments.PUT("/:id/status", handler.UpdateDeploymentStatus)
		deployments.PUT("/:id/runner", handler.UpdateDeploymentRunner)
		deployments.PUT("/:id/log", handler.UpdateDeploymentLog)
		deployments.POST("/:id/execute", handler.ExecuteDeployment)
		deployments.DELETE("/:id", handler.DeleteDeployment)
	}
}
