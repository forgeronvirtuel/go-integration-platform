package cmd

import (
	"forgeronvirtuel/gip/internal/database"
	"forgeronvirtuel/gip/internal/server"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	port   string
	dbPath string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Démarre le serveur HTTP",
	Long:  `Démarre le serveur HTTP avec Gin et initialise la base de données SQLite.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Str("port", port).Str("db", dbPath).Msg("Démarrage du serveur")

		// Initialiser la base de données
		db, err := database.InitDB(dbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Impossible d'initialiser la base de données")
		}
		defer db.Close()

		log.Info().Msg("Base de données initialisée avec succès")

		// Démarrer le serveur
		server.Start(port, db)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&port, "port", "p", "3000", "Port d'écoute du serveur")
	serveCmd.Flags().StringVarP(&dbPath, "database", "d", "./data.db", "Chemin vers le fichier de base de données SQLite")
}
