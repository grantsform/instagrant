package setup

import (
	"fmt"

	"github.com/grantios/instagrant/pkg/phase"
	"github.com/grantios/instagrant/pkg/util"
)

// PreSetupPhase unmounts and wipes filesystems before installation
type PreSetupPhase struct {
	phase.BasePhase
}

// NewPreSetupPhase creates a new pre-setup phase
func NewPreSetupPhase() phase.Phase {
	return &PreSetupPhase{
		BasePhase: phase.NewBasePhase(
			"presetup",
			"Unmount and wipe existing filesystems",
			false,
		),
	}
}

// Validate checks if the disk exists
func (p *PreSetupPhase) Validate(ctx *phase.Context) error {
	disk := ctx.Config.Disk.Device
	ctx.Logger.Info("Checking disk: %s", disk)
	
	if !util.IsBlockDevice(disk) {
		return fmt.Errorf("disk %s does not exist or is not a block device", disk)
	}
	
	return nil
}

// Execute unmounts and wipes the target disk
func (p *PreSetupPhase) Execute(ctx *phase.Context) error {
	disk := ctx.Config.Disk.Device
	exec := ctx.Exec
	
	ctx.Logger.Info("Preparing disk %s for installation...", disk)

	// Ensure target directory exists
	ctx.Logger.Info("Ensuring target directory %s exists...", ctx.TargetDir)
	if err := exec.Run("mkdir", "-p", ctx.TargetDir); err != nil {
		ctx.Logger.Warn("Failed to create target directory %s: %v", ctx.TargetDir, err)
	}

	// Unmount target directory recursively and all partitions
	ctx.Logger.Info("Unmounting existing mounts...")
	exec.Run("umount", "-R", ctx.TargetDir)
	
	// Find and unmount all partitions on target disk (simplified)
	exec.RunSh(fmt.Sprintf("lsblk -ln -o NAME %s | tail -n +2 | xargs -I{} umount /dev/{} 2>/dev/null || true", disk))

	// Disable swap, wipe, and prepare disk
	ctx.Logger.Info("Preparing disk...")
	exec.Run("swapoff", "-a")
	exec.Run("wipefs", "-af", disk)
	exec.Run("dd", "if=/dev/zero", "of="+disk, "bs=1M", "count=100", "status=none")
	exec.Run("sync")

	// Enable multilib repository and update
	ctx.Logger.Info("Enabling multilib repository...")
	exec.Run("sed", "-i", "/\\[multilib\\]/,/Include/s/^#//", "/etc/pacman.conf")
	exec.Run("pacman", "-Sy")

	ctx.Logger.Success("Disk prepared successfully")
	return nil
}
