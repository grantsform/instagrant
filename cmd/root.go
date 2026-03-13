package cmd

import (
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	dryRun     bool
	verbose    bool
	resumeFrom string
)

var rootCmd = &cobra.Command{
	Use:   "instagrant",
	Short: "Instagrant - Go-based Arch Linux installer",
	Long: `Instagrant is a modular, API-driven Arch Linux installer built in Go.
It provides a config-driven approach to installing Arch Linux with support for
Btrfs snapshots, multiple desktop environments, and automated package management.`,
	Version: "0.1.0",
	RunE:    runInstall,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (optional, will show interactive UI if not provided)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done without executing")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&resumeFrom, "resume-from", "", "resume installation from specific phase")
}
