package cmd

import (
	"forgeronvirtuel/gip/internal/database"
	"forgeronvirtuel/gip/internal/server"
	"forgeronvirtuel/gip/internal/workspacemanager"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	port         string
	dbPath       string
	workspaceDir string
)

var serveCmd = &cobra.Command{
	Use:   "control-plane",
	Short: "Start the control plane server",
	Long:  `The control plane server manages projects and orchestrates operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Str("port", port).Str("db", dbPath).Str("workspace", workspaceDir).Msg("Starting control plane server")

		// Validate and check the workspace directory
		if err := workspacemanager.ValidateWorkspaceDir(workspaceDir); err != nil {
			log.Fatal().Err(err).Str("workspace", workspaceDir).Msg("Workspace directory is invalid or inaccessible")
		}

		log.Info().Str("workspace", workspaceDir).Msg("Workspace directory validated")

		// Initialize the database
		db, err := database.InitDB(dbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize the database	")
		}
		defer db.Close()

		log.Info().Msg("Database initialized successfully")

		// Start the server
		server.Start(port, db, workspaceDir)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&port, "port", "p", "3000", "Server listening port")
	serveCmd.Flags().StringVarP(&dbPath, "database", "d", "./data.db", "Path to the SQLite database file")
	serveCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", "./workspace", "Workspace directory for projects")
}
