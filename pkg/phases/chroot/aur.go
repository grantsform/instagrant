package chroot

import (
	"fmt"
	"strings"

	"github.com/grantios/instagrant/pkg/phase"
)

// AURPhase installs AUR packages using yay
type AURPhase struct {
	phase.BasePhase
}

func NewAURPhase() *AURPhase {
	return &AURPhase{
		BasePhase: phase.NewBasePhase(
			"aur",
			"Install AUR packages",
			true, // isChroot
		),
	}
}

func (a *AURPhase) Execute(ctx *phase.Context) error {
	username := ctx.Config.User.Username
	homeDir := ctx.Config.User.HomeDir
	h := ctx.Helper
	
	ctx.Logger.Stage("Installing yay AUR helper")
	
	if err := h.InstallYay(username, homeDir); err != nil {
		return fmt.Errorf("failed to install yay: %w", err)
	}
	
	ctx.Logger.Success("yay AUR helper installed")
	
	// Install AUR packages if specified
	if packages := ctx.Config.Packages.AUR; len(packages) > 0 {
		ctx.Logger.Stage("Installing %d AUR packages", len(packages))
		ctx.Logger.Debug("AUR packages: %s", strings.Join(packages, " "))
		
		if err := h.InstallAURPackages(username, packages...); err != nil {
			return fmt.Errorf("failed to install AUR packages: %w", err)
		}
		
		ctx.Logger.Success("AUR packages installed")
	}
	
	return nil
}
