package cmd

import (
	"forgeronvirtuel/gip/internal/runner"
	"forgeronvirtuel/gip/internal/server"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "Start the runner runner",
	Long:  `Start an runner runner that registers with the control plane and execute deployment ordered by control plane.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().
			Str("control_plane", controlPlaneURL).
			Str("name", runnerName).
			Str("address", address).
			Str("port", port).
			Interface("labels", runnerLabels).
			Msg("Starting runner")

		// Check if a runner with this name already exists
		runnerID, err := runner.GetRunnerId(controlPlaneURL, runnerName)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get runner by name")
		}
		log.Info().Int("runner_id", runnerID).Msg("Runner name verification completed")

		// Register the runner
		if runnerID == 0 {
			log.Info().Msg("runner not registered, processing...")
			if runnerID, err = runner.RegisterRunner(controlPlaneURL, runnerName, runnerLabels, runnerURL); err != nil {
				log.Fatal().Err(err).Msg("failed to register runner")
			}
		} else {
			// If the runner already exists, update its URL
			if err := runner.UpdateRunnerURL(controlPlaneURL, runnerID, runnerURL); err != nil {
				log.Warn().Err(err).Msg("failed to update runner URL")
			}
		}

		log.Info().Int("runner_id", runnerID).Msg("runner sucessfully registered")

		// Set the runner to ONLINE
		if err := runner.UpdateRunnerStatus(controlPlaneURL, runnerID, "ONLINE"); err != nil {
			log.Warn().Err(err).Msg("failed to set runner status to ONLINE")
		} else {
			log.Info().Msg("runner set to ONLINE")
		}

		// Start the heartbeat goroutine
		stopChan := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)
		go runner.StartHeartbeat(controlPlaneURL, runnerID, stopChan, &wg)

		// Start the HTTP server to receive deployment requests
		wg.Add(1)
		go server.StartRunnerServer(runnerID, runnerURL, controlPlaneURL, stopChan, &wg)

		// Wait for a stop signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		<-sigChan
		log.Info().Msg("stop signal received, stopping runner...")

		// Stop the heartbeat goroutine
		close(stopChan)

		// Set the runner to OFFLINE before exiting
		if err := runner.UpdateRunnerStatus(controlPlaneURL, runnerID, "OFFLINE"); err != nil {
			log.Warn().Err(err).Msg("failed to set runner status to OFFLINE")
		} else {
			log.Info().Msg("runner set to OFFLINE")
		}

		wg.Wait()

		log.Info().Msg("Runner stopped gracefully")
	},
}

func init() {
	rootCmd.AddCommand(runnerCmd)

	// Determine default name (hostname)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-runner"
	}

	runnerCmd.Flags().StringVarP(&controlPlaneURL, "control-plane", "c", "", "URL of the control plane (e.g., http://localhost:3000)")
	runnerCmd.MarkFlagRequired("control-plane")

	runnerCmd.Flags().StringVarP(&runnerName, "name", "n", hostname, "Name of the runner (default is hostname)")
	runnerCmd.Flags().StringVarP(&runnerURL, "url", "u", "http://localhost:3000", "URL of the runner (e.g., http://my-runner:3000)")
	runnerCmd.Flags().StringToStringVarP(&runnerLabels, "labels", "l", nil, "runner labels (format: key1=value1,key2=value2)")
	runnerCmd.Flags().StringVarP(&workspaceDir, "workspace", "w", "./workspace", "Workspace directory for projects")
}
