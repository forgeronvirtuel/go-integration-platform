package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gip",
	Short: "Go Integration Platform - Un serveur HTTP minimaliste",
	Long:  `GIP est un serveur HTTP minimaliste construit avec Gin, utilisant SQLite comme base de données.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Erreur lors de l'exécution de la commande")
		os.Exit(1)
	}
}

func init() {
	// Configuration de zerolog
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}
