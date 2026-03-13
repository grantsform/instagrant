package after

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/grantios/instagrant/pkg/phase"
	"github.com/grantios/instagrant/pkg/ui"
)

// tuiWriter writes only to UI when active
type tuiWriter struct{}

func (w *tuiWriter) Write(p []byte) (n int, err error) {
	text := string(p)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			ui.AddLog(line)
		}
	}
	return len(p), nil
}

// runCommand executes a command and logs it
func runCommand(ctx *phase.Context, name string, args ...string) error {
	ctx.Logger.Command(fmt.Sprintf("%s %s", name, strings.Join(args, " ")))
	
	if ctx.DryRun {
		return nil
	}
	
	cmd := exec.Command(name, args...)
	if ui.IsActive() {
		cmd.Stdout = &tuiWriter{}
		cmd.Stderr = &tuiWriter{}
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		ctx.Logger.Error("Command failed: %v", err)
		return err
	}
	
	return nil
}

// runInChroot executes a command inside the chroot environment
func runInChroot(ctx *phase.Context, command string) error {
	ctx.Logger.Command(fmt.Sprintf("(chroot) %s", command))
	
	if ctx.DryRun {
		return nil
	}
	
	cmd := exec.Command("arch-chroot", ctx.TargetDir, "bash", "-c", command)
	if ui.IsActive() {
		cmd.Stdout = &tuiWriter{}
		cmd.Stderr = &tuiWriter{}
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		ctx.Logger.Error("Chroot command failed: %v", err)
		return err
	}
	
	return nil
}

// CleanupPhase performs final cleanup after chroot
type CleanupPhase struct {
	phase.BasePhase
}

// NewCleanupPhase creates a new cleanup phase
func NewCleanupPhase() phase.Phase {
	return &CleanupPhase{
		BasePhase: phase.NewBasePhase(
			"cleanup",
			"Perform final cleanup and unmounting",
			false, // Not a chroot phase
		),
	}
}

// Execute performs cleanup operations
func (p *CleanupPhase) Execute(ctx *phase.Context) error {
	ctx.Logger.Info("Running cleanup operations...")

	if ctx.DryRun {
		ctx.Logger.Info("[DRY RUN] Would clean pacman cache")
		ctx.Logger.Info("[DRY RUN] Would remove temporary files")
		return nil
	}

	// Clean pacman cache in chroot
	ctx.Logger.Info("Cleaning pacman cache...")
	if err := runInChroot(ctx, "pacman -Scc --noconfirm"); err != nil {
		ctx.Logger.Warn("Failed to clean pacman cache: %v", err)
	}

	// Remove any temporary installation files
	ctx.Logger.Info("Removing temporary files...")
	tempFiles := []string{
		ctx.TargetDir + "/var/cache/pacman/pkg/*",
		ctx.TargetDir + "/root/.bash_history",
		ctx.TargetDir + "/tmp/*",
	}
	
	for _, file := range tempFiles {
		if err := runCommand(ctx, "rm", "-rf", file); err != nil {
			ctx.Logger.Warn("Failed to remove %s: %v", file, err)
		}
	}

	// Sync filesystem
	ctx.Logger.Info("Syncing filesystems...")
	if err := runCommand(ctx, "sync"); err != nil {
		ctx.Logger.Warn("Failed to sync: %v", err)
	}

	// NOTE: We intentionally do NOT unmount here.
	// The target directory is left mounted so the user can:
	// 1. Enter chroot environment for manual configuration
	// 2. Choose "Unmount and reboot" which handles unmounting

	ctx.Logger.Success("Cleanup completed")
	return nil
}
