package chroot

import (
	"fmt"

	"github.com/grantios/instagrant/pkg/phase"
)

// MkinitcpioPhase configures and regenerates initramfs
type MkinitcpioPhase struct {
	phase.BasePhase
}

func NewMkinitcpioPhase() *MkinitcpioPhase {
	return &MkinitcpioPhase{
		BasePhase: phase.NewBasePhase(
			"mkinitcpio",
			"Configure and regenerate initramfs",
			true, // isChroot
		),
	}
}

func (m *MkinitcpioPhase) Execute(ctx *phase.Context) error {
	exec := ctx.Exec
	
	ctx.Logger.Stage("Configuring mkinitcpio")
	
	// Check if kernel is installed
	kernelCheckCmd := fmt.Sprintf("pacman -Q %s", ctx.Config.System.Kernel)
	if err := exec.Chroot(ctx.TargetDir, kernelCheckCmd); err != nil {
		ctx.Logger.Warn("Kernel %s not found, skipping mkinitcpio configuration", ctx.Config.System.Kernel)
		return nil
	}
	
	ctx.Logger.Stage("Regenerating initramfs")
	
	// Regenerate initramfs for all kernels
	if err := exec.Chroot(ctx.TargetDir, "mkinitcpio -P"); err != nil {
		ctx.Logger.Warn("Failed to regenerate initramfs: %v", err)
		ctx.Logger.Info("This may be due to missing kernel modules or incomplete installation")
		ctx.Logger.Info("You can manually regenerate later with: mkinitcpio -P")
		// Don't fail the installation for this
		return nil
	}
	
	ctx.Logger.Success("Initramfs configured and regenerated")
	return nil
}