package chroot

import (
	"fmt"

	"github.com/grantios/instagrant/pkg/phase"
)

// SnapperPhase configures Btrfs snapshots with Snapper
type SnapperPhase struct {
	phase.BasePhase
}

func NewSnapperPhase() *SnapperPhase {
	return &SnapperPhase{
		BasePhase: phase.NewBasePhase(
			"snapper",
			"Configure Btrfs snapshots",
			true, // isChroot
		),
	}
}

func (s *SnapperPhase) Execute(ctx *phase.Context) error {
	h := ctx.Helper
	exec := ctx.Exec
	
	ctx.Logger.Stage("Installing snapper")
	
	// Install snapper packages
	if err := h.Pacman("snapper", "snap-pac"); err != nil {
		return fmt.Errorf("failed to install snapper: %w", err)
	}
	
	// Install snapper-rollback from AUR
	ctx.Logger.Info("Installing snapper-rollback from AUR...")
	if err := h.InstallAURPackages(ctx.Config.User.Username, "snapper-rollback"); err != nil {
		return fmt.Errorf("failed to install snapper-rollback: %w", err)
	}
	
	ctx.Logger.Stage("Configuring snapper")
	
	// Setup snapper with proper subvolume mounting
	setupCmd := `umount /.snapshots 2>/dev/null; rm -rf /.snapshots && 
		snapper --no-dbus -c root create-config / && 
		btrfs subvolume delete /.snapshots 2>/dev/null; mkdir /.snapshots && 
		mount -a && chmod 750 /.snapshots`
	
	if err := exec.Chroot(ctx.TargetDir, setupCmd); err != nil {
		return fmt.Errorf("failed to configure snapper: %w", err)
	}
	
	// Configure snapper-rollback to use /.btrfsroot
	rollbackCmd := `sed -i 's|mountpoint = /btrfsroot|mountpoint = /.btrfsroot|' /etc/snapper-rollback.conf`
	if err := exec.Chroot(ctx.TargetDir, rollbackCmd); err != nil {
		return fmt.Errorf("failed to configure snapper-rollback: %w", err)
	}
	
	// Enable snapper timers
	if err := h.EnableServices("snapper-timeline.timer", "snapper-cleanup.timer"); err != nil {
		return fmt.Errorf("failed to enable snapper timers: %w", err)
	}
	
	ctx.Logger.Success("Snapper configured")
	return nil
}
