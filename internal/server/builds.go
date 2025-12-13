package server

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"forgeronvirtuel/gip/internal/database"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v6"
)

type BuildHandler struct {
	DB        *sql.DB
	workspace string
}

type CreateBuildRequest struct {
	ProjectID int `json:"project_id" binding:"required"`
}

func (h *BuildHandler) CreateBuild(c *gin.Context) {

	logBuf := &bytes.Buffer{}
	logWriter := io.MultiWriter(logBuf)

	// Global timeout for the whole pipeline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var req CreateBuildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request payload"})
		return
	}

	project, err := database.GetProjectByID(h.DB, req.ProjectID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Project not found"})
		return
	}

	// Create build record in database with pending status
	build, err := database.CreateBuild(h.DB, req.ProjectID, project.Branch)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create build record"})
		return
	}

	// Update status to building
	err = database.UpdateBuildStatus(h.DB, build.ID, "building", "")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update build status"})
		return
	}

	// Clone the repository into the workspace directory
	// Always use absolute path
	absWorkspace, err := filepath.Abs(h.workspace)
	if err != nil {
		database.UpdateBuildStatus(h.DB, build.ID, "failed", "Failed to get absolute workspace path")
		c.JSON(500, gin.H{"error": "Failed to get absolute workspace path"})
		return
	}

	repoPath := filepath.Join(absWorkspace, fmt.Sprintf("project-%d", project.ID))
	err = os.RemoveAll(repoPath)
	if err != nil {
		database.UpdateBuildStatus(h.DB, build.ID, "failed", "Failed to clean workspace")
		c.JSON(500, gin.H{"error": "Failed to clean workspace"})
		return
	}

	_, err = git.PlainClone(repoPath, &git.CloneOptions{
		URL: project.RepoURL,
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to clone repository"})
		return
	}

	sourceDir := repoPath
	if project.Subdir != "" {
		sourceDir = filepath.Join(repoPath, project.Subdir)
	}

	// Check if a file in "cmd/main.go" exists
	mainGoPath := filepath.Join(sourceDir, "cmd", "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		c.JSON(400, gin.H{"error": "cmd/main.go not found in the repository"})
		return
	}

	// Step 2 (optional but recommended): download Go modules
	if err := runCmd(ctx, sourceDir, logWriter, "go", "mod", "download"); err != nil {
		// Not fatal in all cases, but usually indicates a real problem
		c.JSON(400, gin.H{"error": err.Error(), "logs": logBuf.String()})
		return
	}

	// Step 3: build the binary
	outDir := filepath.Join(repoPath, "out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		database.UpdateBuildStatus(h.DB, build.ID, "failed", logBuf.String())
		c.JSON(400, gin.H{"error": err.Error(), "logs": logBuf.String()})
		return
	}

	// Generate binary name from project name
	binaryName := fmt.Sprintf("%s-%d", project.Name, build.ID)
	binaryPath := filepath.Join(outDir, binaryName)
	// binaryPath is already absolute since repoPath is absolute

	buildArgs := []string{
		"build",
		"-o", binaryPath,
		"./cmd/main.go",
	}

	if err := runCmd(ctx, sourceDir, logWriter, "go", buildArgs...); err != nil {
		database.UpdateBuildStatus(h.DB, build.ID, "failed", logBuf.String())
		c.JSON(400, gin.H{"error": err.Error(), "logs": logBuf.String()})
		return
	}

	// Update build status to success with binary path stored in log_output
	err = database.UpdateBuildStatus(h.DB, build.ID, "success", fmt.Sprintf("Binary: %s\n\n%s", binaryPath, logBuf.String()))
	if err != nil {
		c.JSON(500, gin.H{"error": "Build succeeded but failed to update status"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"build_id":     build.ID,
		"status":       "success",
		"binary_path":  binaryPath,
		"download_url": fmt.Sprintf("/api/builds/%d/download", build.ID),
	})
}

// DownloadBinary permet de télécharger le binaire généré par un build
func (h *BuildHandler) DownloadBinary(c *gin.Context) {
	buildID := c.Param("id")

	// Récupérer le build depuis la DB
	build, err := database.GetBuildByID(h.DB, buildID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Build not found"})
		return
	}

	// Vérifier que le build est en succès
	if build.Status != "success" {
		c.JSON(400, gin.H{"error": "Build is not successful, cannot download binary"})
		return
	}

	// Extraire le chemin du binaire depuis log_output (première ligne)
	var binaryPath string

	if len(build.LogOutput) > 8 && build.LogOutput[:8] == "Binary: " {
		// Extraire le chemin (jusqu'au premier \n)
		for i := 8; i < len(build.LogOutput); i++ {
			if build.LogOutput[i] == '\n' {
				binaryPath = build.LogOutput[8:i]
				break
			}
		}
	}

	if binaryPath == "" {
		c.JSON(500, gin.H{"error": "Binary path not found in build logs"})
		return
	}

	// Vérifier que le fichier existe
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": "Binary file not found on disk"})
		return
	}

	// Récupérer le projet pour obtenir le nom
	project, err := database.GetProjectByID(h.DB, build.ProjectID)
	if err != nil {
		// Fallback si on ne peut pas récupérer le projet
		c.File(binaryPath)
		return
	}

	// Définir le nom du fichier pour le téléchargement
	filename := fmt.Sprintf("%s-%d", project.Name, build.ID)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.File(binaryPath)
}

// GetBuild récupère les détails d'un build spécifique
func (h *BuildHandler) GetBuild(c *gin.Context) {
	buildID := c.Param("id")

	build, err := database.GetBuildByID(h.DB, buildID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Build not found"})
		return
	}

	c.JSON(200, gin.H{
		"id":         build.ID,
		"project_id": build.ProjectID,
		"branch":     build.Branch,
		"status":     build.Status,
		"log_output": build.LogOutput,
		"started_at": build.StartedAt,
		"ended_at":   build.EndedAt,
		"created_at": build.CreatedAt,
	})
}

// GetBuildsByProject récupère tous les builds d'un projet
func (h *BuildHandler) GetBuildsByProject(c *gin.Context) {
	projectIDStr := c.Param("project_id")

	// Convertir le project_id en int
	var projectID int
	if _, err := fmt.Sscanf(projectIDStr, "%d", &projectID); err != nil {
		c.JSON(400, gin.H{"error": "Invalid project ID"})
		return
	}

	builds, err := database.GetBuildsByProjectID(h.DB, projectID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch builds"})
		return
	}

	// Convertir les builds en réponse JSON
	var response []gin.H
	for _, build := range builds {
		response = append(response, gin.H{
			"id":         build.ID,
			"project_id": build.ProjectID,
			"branch":     build.Branch,
			"status":     build.Status,
			"started_at": build.StartedAt,
			"ended_at":   build.EndedAt,
			"created_at": build.CreatedAt,
		})
	}

	c.JSON(200, response)
}

func setupBuildRoutes(router *gin.RouterGroup, db *sql.DB, workspace string) {
	handler := BuildHandler{DB: db, workspace: workspace}
	builds := router.Group("/api/builds")
	{
		builds.POST("/", handler.CreateBuild)
		builds.GET("/:id", handler.GetBuild)
		builds.GET("/:id/download", handler.DownloadBinary)
		builds.GET("/project/:project_id", handler.GetBuildsByProject)
	}
}

// runCmd runs a command with a controlled environment and logs its output.
// It avoids using any shell ("bash -c") to prevent injection issues.
func runCmd(ctx context.Context, workDir string, log io.Writer, name string, args ...string) error {
	fmt.Fprintf(log, "==> Running: %s %v (in %s)\n", name, args, workDir)

	// Convert to absolute path to ensure GOCACHE and GOMODCACHE are absolute
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = absWorkDir

	// Minimal, controlled environment:
	// - PATH is kept from parent (to find git/go)
	// - HOME is set to workDir (avoid polluting real home)
	// - No proxy, no extra env from the parent process.
	env := []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + absWorkDir,
		"GOMODCACHE=" + filepath.Join(absWorkDir, ".gomodcache"),
		"GOCACHE=" + filepath.Join(absWorkDir, ".gocache"),
	}
	cmd.Env = env

	cmd.Stdout = log
	cmd.Stderr = log

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
