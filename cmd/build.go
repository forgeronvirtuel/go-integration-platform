package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"forgeronvirtuel/gip/internal/database"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	projectID   int
	projectName string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a project",
	Long:  `Build a project by cloning its repository and compiling the Go binary.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().
			Int("project_id", projectID).
			Str("project_name", projectName).
			Str("db", dbPath).
			Str("workspace", workspaceDir).
			Msg("Starting build command")

		// Validate required flags
		if projectID == 0 && projectName == "" {
			log.Fatal().Msg("Either --project-id or --project-name must be specified")
		}

		// Initialize the database
		db, err := database.InitDB(dbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize the database")
		}
		defer db.Close()

		// Get the project
		var project *database.Project
		if projectID != 0 {
			project, err = database.GetProjectByID(db, projectID)
			if err != nil {
				log.Fatal().Err(err).Int("project_id", projectID).Msg("Project not found")
			}
		} else {
			project, err = database.GetProjectByName(db, projectName)
			if err != nil {
				log.Fatal().Err(err).Str("project_name", projectName).Msg("Project not found")
			}
		}

		log.Info().
			Int("id", project.ID).
			Str("name", project.Name).
			Str("repo_url", project.RepoURL).
			Str("branch", project.Branch).
			Msg("Project found")

		// Create build record in database with pending status
		build, err := database.CreateBuild(db, project.ID, project.Branch)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create build record")
		}

		log.Info().Int("build_id", build.ID).Msg("Build record created")

		// Update status to building
		err = database.UpdateBuildStatus(db, build.ID, "building", "")
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to update build status")
		}

		// Execute the build
		if err := executeBuild(db, project, build, workspaceDir); err != nil {
			log.Error().Err(err).Msg("Build failed")
			os.Exit(1)
		}

		log.Info().
			Int("build_id", build.ID).
			Str("status", "success").
			Msg("Build completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().IntVarP(&projectID, "project-id", "i", 0, "ID of the project to build")
	buildCmd.Flags().StringVarP(&projectName, "project-name", "n", "", "Name of the project to build")
	buildCmd.Flags().StringVarP(&dbPath, "database", "d", "./data.db", "Path to the SQLite database file")
	buildCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", "./workspace", "Workspace directory for projects")
}

func executeBuild(db *sql.DB, project *database.Project, build *database.Build, workspace string) error {
	logBuf := &bytes.Buffer{}
	logWriter := io.MultiWriter(logBuf, os.Stdout)

	// Global timeout for the whole pipeline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Clone the repository into the workspace directory
	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		database.UpdateBuildStatus(db, build.ID, "failed", "Failed to get absolute workspace path")
		return fmt.Errorf("failed to get absolute workspace path: %w", err)
	}

	repoPath := filepath.Join(absWorkspace, fmt.Sprintf("project-%d", project.ID))

	log.Info().Str("repo_path", repoPath).Msg("Cleaning workspace")
	err = os.RemoveAll(repoPath)
	if err != nil {
		database.UpdateBuildStatus(db, build.ID, "failed", "Failed to clean workspace")
		return fmt.Errorf("failed to clean workspace: %w", err)
	}

	log.Info().
		Str("repo_url", project.RepoURL).
		Str("repo_path", repoPath).
		Msg("Cloning repository")

	_, err = git.PlainClone(repoPath, &git.CloneOptions{
		URL: project.RepoURL,
	})
	if err != nil {
		database.UpdateBuildStatus(db, build.ID, "failed", "Failed to clone repository")
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	sourceDir := repoPath
	if project.Subdir != "" {
		sourceDir = filepath.Join(repoPath, project.Subdir)
	}

	// Check if a file in "cmd/main.go" exists
	mainGoPath := filepath.Join(sourceDir, "cmd", "main.go")
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		database.UpdateBuildStatus(db, build.ID, "failed", "cmd/main.go not found in the repository")
		return fmt.Errorf("cmd/main.go not found in the repository")
	}

	log.Info().Msg("Downloading Go modules")
	// Download Go modules
	if err := runCmd(ctx, sourceDir, logWriter, "go", "mod", "download"); err != nil {
		database.UpdateBuildStatus(db, build.ID, "failed", logBuf.String())
		return fmt.Errorf("failed to download modules: %w", err)
	}

	// Build the binary
	outDir := filepath.Join(repoPath, "out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		database.UpdateBuildStatus(db, build.ID, "failed", logBuf.String())
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate binary name from project name
	binaryName := fmt.Sprintf("%s-%d", project.Name, build.ID)
	binaryPath := filepath.Join(outDir, binaryName)

	log.Info().
		Str("binary_path", binaryPath).
		Msg("Building binary")

	buildArgs := []string{
		"build",
		"-o", binaryPath,
		"./cmd/main.go",
	}

	if err := runCmd(ctx, sourceDir, logWriter, "go", buildArgs...); err != nil {
		database.UpdateBuildStatus(db, build.ID, "failed", logBuf.String())
		return fmt.Errorf("failed to build: %w", err)
	}

	// Update build status to success with binary path stored in log_output
	err = database.UpdateBuildStatus(db, build.ID, "success", fmt.Sprintf("Binary: %s\n\n%s", binaryPath, logBuf.String()))
	if err != nil {
		return fmt.Errorf("build succeeded but failed to update status: %w", err)
	}

	log.Info().
		Str("binary_path", binaryPath).
		Msg("Binary built successfully")

	return nil
}

func runCmd(ctx context.Context, dir string, logWriter io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	fmt.Fprintf(logWriter, "Running: %s %v\n", name, args)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s %v: %w", name, args, err)
	}

	return nil
}
