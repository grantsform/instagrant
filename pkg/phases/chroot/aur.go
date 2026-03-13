package chroot

import (
	"fmt"
	"strings"
	"time"

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
			ctx.Logger.Warn("Initial AUR install attempt failed: %v", err)

			// Retry only packages that are not already installed
			remaining := []string{}
			for _, pkg := range packages {
				installed, instErr := h.IsPacmanPackageInstalled(pkg)
				if instErr != nil {
					ctx.Logger.Warn("failed to check if package %s is installed: %v", pkg, instErr)
					remaining = append(remaining, pkg)
					continue
				}
				if !installed {
					remaining = append(remaining, pkg)
				}
			}

			if len(remaining) > 0 {
				ctx.Logger.Stage("Retrying AUR install for %d remaining packages", len(remaining))
				ctx.Logger.Debug("AUR retry packages: %s", strings.Join(remaining, " "))
				if err2 := h.InstallAURPackages(username, remaining...); err2 != nil {
					ctx.Logger.Warn("Retry failed: %v", err2)
				}
			}

			// Report final failures but don't exit installation
			failed := []string{}
			for _, pkg := range packages {
				installed, instErr := h.IsPacmanPackageInstalled(pkg)
				if instErr != nil {
					ctx.Logger.Warn("failed to check if package %s is installed: %v", pkg, instErr)
					failed = append(failed, pkg)
					continue
				}
				if !installed {
					failed = append(failed, pkg)
				}
			}

			if len(failed) > 0 {
				// Persist failures so users can inspect after installation
				logContent := fmt.Sprintf("%s AUR build failures:\n%s\n", time.Now().Format(time.RFC3339), strings.Join(failed, "\n"))
				if err := h.WriteFile("/tios-log.txt", logContent); err != nil {
					ctx.Logger.Warn("failed to write AUR failure log: %v", err)
				} else {
					ctx.Logger.Warn("AUR build failures logged to /tios-log.txt")
				}
			}
		} else {
			ctx.Logger.Success("AUR packages installed")
		}
	}
	
	return nil
}
