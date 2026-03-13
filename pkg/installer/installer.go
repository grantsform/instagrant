package installer

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/grantios/instagrant/pkg/config"
	pkglogger "github.com/grantios/instagrant/pkg/logger"
	"github.com/grantios/instagrant/pkg/phase"
	"github.com/grantios/instagrant/pkg/phases/after"
	"github.com/grantios/instagrant/pkg/phases/chroot"
	"github.com/grantios/instagrant/pkg/phases/setup"
	"github.com/grantios/instagrant/pkg/ui"
	"github.com/grantios/instagrant/pkg/util"
)

// Installer manages the installation process
type Installer struct {
	config   *config.Config
	logger   *pkglogger.Logger
	dryRun   bool
	registry *phase.Registry
	context  *phase.Context
}

// New creates a new installer instance
func New(cfg *config.Config, log *pkglogger.Logger, dryRun bool) *Installer {
	registry := phase.NewRegistry()
	
	// Register all phases
	registry.Register(
		setup.NewPreSetupPhase(),
		setup.NewPartitionPhase(),
		setup.NewMountPhase(),
		setup.NewPacstrapPhase(),
		setup.NewFstabPhase(),
		chroot.NewTimezonePhase(),
		chroot.NewHostnamePhase(),
		chroot.NewUserPhase(),
		chroot.NewPackagesPhase(),
		chroot.NewAURPhase(),
		chroot.NewMkinitcpioPhase(),
		chroot.NewBootloaderPhase(),
		chroot.NewServicesPhase(),
		chroot.NewSkeletonPhase(),
		chroot.NewTiosDirPhase(),
		chroot.NewSpecificsPhase(),
		chroot.NewSnapperPhase(),
		after.NewCleanupPhase(),
	)
	
	// Create executor with UI integration
	exec := util.NewExecutor(dryRun, ui.AddLog, log.Command)
	helper := util.NewPhaseHelper(exec, cfg.Target)
	
	context := &phase.Context{
		Config:    cfg,
		Logger:    log,
		DryRun:    dryRun,
		TargetDir: cfg.Target,
		State:     &phase.State{},
		Exec:      exec,
		Helper:    helper,
	}
	
	return &Installer{
		config:   cfg,
		logger:   log,
		dryRun:   dryRun,
		registry: registry,
		context:  context,
	}
}

// Run executes the full installation
func (i *Installer) Run() error {
	phases := i.registry.GetPhases()
	
	// Build phase list for UI
	phaseList := make([]ui.PhaseInfo, len(phases))
	for idx, p := range phases {
		phaseList[idx] = ui.PhaseInfo{
			Name:        p.Name(),
			Description: p.Description(),
			Status:      ui.StatusPending,
		}
	}
	
	// Start TUI
	if !i.dryRun {
		if err := ui.Start(phaseList); err != nil {
			return fmt.Errorf("failed to start UI: %w", err)
		}
		
		// Hook logger to UI
		pkglogger.UICallback = ui.AddLog
		defer func() {
			pkglogger.UICallback = nil
		}()
	}
	
	i.logger.Info("Starting installation with %d phases", len(phases))
	ui.AddLog(fmt.Sprintf("Starting installation with %d phases", len(phases)))
	
	for idx, p := range phases {
		ui.UpdatePhase(idx, ui.StatusRunning)
		i.logger.Step("Phase %d/%d: %s", idx+1, len(phases), p.Description())
		ui.AddLog(fmt.Sprintf("▶ Phase %d/%d: %s", idx+1, len(phases), p.Description()))
		
		// Update state
		i.context.State.CurrentPhase = p.Name()
		
		// Validate phase
		if err := p.Validate(i.context); err != nil {
			ui.UpdatePhase(idx, ui.StatusFailed)
			ui.AddLog(fmt.Sprintf("✗ Validation failed: %v", err))
			return fmt.Errorf("phase %s validation failed: %w", p.Name(), err)
		}
		
		// Execute phase
		if err := i.executePhase(p); err != nil {
			i.context.State.Failed = true
			i.context.State.ErrorMessage = err.Error()
			ui.UpdatePhase(idx, ui.StatusFailed)
			ui.AddLog(fmt.Sprintf("✗ Phase failed: %v", err))
			return fmt.Errorf("phase %s failed: %w", p.Name(), err)
		}
		
		// Mark as completed
		i.context.State.Completed = append(i.context.State.Completed, p.Name())
		ui.UpdatePhase(idx, ui.StatusComplete)
		i.logger.Success("Phase %s completed", p.Name())
		ui.AddLog(fmt.Sprintf("✓ Phase %s completed", p.Name()))
	}
	
	// Wait for user to interact with post-install menu
	if !i.dryRun {
		ui.Wait()
		
		// Reset terminal to sane state after TUI exits
		exec.Command("stty", "sane").Run()
		fmt.Print("\033[?25h") // Show cursor
		
		targetDir := i.config.Target
		
		// Handle the exit action
		switch ui.GetExitAction() {
		case 1: // Reboot
			exec.Command("sh", "-c", fmt.Sprintf("umount -R %s && reboot", targetDir)).Run()
		case 2: // Chroot
			fmt.Printf("\nEntering chroot environment at %s. Type 'exit' to leave.\n", targetDir)
			cmd := exec.Command("arch-chroot", targetDir)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			fmt.Println("\nExited chroot. You can now manually unmount and reboot.")
		case 3: // Exit
			// Just exit normally
		}
	}
	
	return nil
}

// ResumeFrom resumes installation from a specific phase
func (i *Installer) ResumeFrom(phaseName string) error {
	phases := i.registry.GetPhasesFrom(phaseName)
	if phases == nil {
		return fmt.Errorf("phase %s not found", phaseName)
	}
	
	i.logger.Info("Resuming from phase: %s (%d phases remaining)", phaseName, len(phases))
	
	for idx, p := range phases {
		i.logger.Step("Phase %d/%d: %s", idx+1, len(phases), p.Description())
		
		i.context.State.CurrentPhase = p.Name()
		
		if err := p.Validate(i.context); err != nil {
			return fmt.Errorf("phase %s validation failed: %w", p.Name(), err)
		}
		
		if err := i.executePhase(p); err != nil {
			i.context.State.Failed = true
			i.context.State.ErrorMessage = err.Error()
			return fmt.Errorf("phase %s failed: %w", p.Name(), err)
		}
		
		i.context.State.Completed = append(i.context.State.Completed, p.Name())
		i.logger.Success("Phase %s completed", p.Name())
	}
	
	return nil
}

// executePhase runs a single phase, handling chroot if needed
func (i *Installer) executePhase(p phase.Phase) error {
	if i.dryRun {
		i.logger.Info("[DRY RUN] Would execute phase: %s", p.Name())
		return nil
	}
	
	if p.IsChroot() {
		return i.executeInChroot(p)
	}
	
	return p.Execute(i.context)
}

// executeInChroot runs a phase inside the chroot environment
func (i *Installer) executeInChroot(p phase.Phase) error {
	// For chroot phases, we need to execute them in the target system
	// This is a simplified approach - in production you'd want more sophisticated chroot handling
	
	i.logger.Debug("Executing phase %s in chroot", p.Name())
	
	// Copy installer to target
	if err := i.copyInstallerToTarget(); err != nil {
		return fmt.Errorf("failed to copy installer to target: %w", err)
	}
	
	// Execute the phase
	return p.Execute(i.context)
}

// copyInstallerToTarget copies the installer binary to the target system
func (i *Installer) copyInstallerToTarget() error {
	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	
	targetPath := fmt.Sprintf("%s/usr/local/bin/instagrant", i.context.TargetDir)
	
	// Create directory
	if err := os.MkdirAll(fmt.Sprintf("%s/usr/local/bin", i.context.TargetDir), 0755); err != nil {
		return err
	}
	
	// Copy file
	cmd := exec.Command("cp", exePath, targetPath)
	if err := cmd.Run(); err != nil {
		return err
	}
	
	// Make executable
	return os.Chmod(targetPath, 0755)
}
