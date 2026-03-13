package cmd

import (
	"fmt"
	"os"

	"github.com/grantios/instagrant/pkg/config"
	"github.com/grantios/instagrant/pkg/installer"
	"github.com/grantios/instagrant/pkg/logger"
	"github.com/grantios/instagrant/pkg/ui"
	"github.com/spf13/cobra"
)

func runInstall(cmd *cobra.Command, args []string) error {
	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("this command must be run as root")
	}

	// Initialize logger
	log := logger.New(verbose)

	var cfg *config.Config
	var err error

	// If no config file specified, show interactive UI
	if cfgFile == "" {
		log.Info("Starting interactive configuration...")
		
		// List available configs (from embedded and current directory)
		configs := config.ListConfigs()
		
		cfg, err = ui.StartConfigUI(configs)
		if err != nil {
			return fmt.Errorf("configuration cancelled or failed: %w", err)
		}
	} else {
		// Load configuration from file
		log.Info("Loading configuration from: %s", cfgFile)
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	log.Success("Configuration validated successfully")

	// Configuration is already confirmed in UI or loaded from file
	// Start installation directly

	// Create installer
	inst := installer.New(cfg, log, dryRun)

	// Resume from specific phase if requested
	if resumeFrom != "" {
		log.Info("Resuming from phase: %s", resumeFrom)
		if err := inst.ResumeFrom(resumeFrom); err != nil {
			return fmt.Errorf("failed to resume: %w", err)
		}
	} else {
		// Run full installation (blocks until user completes post-install menu)
		if err := inst.Run(); err != nil {
			return fmt.Errorf("installation failed: %w", err)
		}
	}
	
	return nil
}
