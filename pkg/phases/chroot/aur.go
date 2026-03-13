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

		// Determine which packages are already installed so we only attempt missing ones
		missing := []string{}
		for _, pkg := range packages {
			installed, instErr := h.IsPacmanPackageInstalled(pkg)
			if instErr != nil {
				ctx.Logger.Warn("failed to check if package %s is installed: %v", pkg, instErr)
				missing = append(missing, pkg)
				continue
			}
			if !installed {
				missing = append(missing, pkg)
			}
		}

		if len(missing) == 0 {
			ctx.Logger.Success("All AUR packages already installed")
			return nil
		}

		ctx.Logger.Stage("Installing %d missing AUR packages", len(missing))
		ctx.Logger.Debug("AUR missing packages: %s", strings.Join(missing, " "))

		if err := h.InstallAURPackages(username, missing...); err != nil {
			ctx.Logger.Warn("Initial AUR install attempt failed: %v", err)

			// Retry only packages still missing after first pass
			retry := []string{}
			for _, pkg := range missing {
				installed, instErr := h.IsPacmanPackageInstalled(pkg)
				if instErr != nil {
					ctx.Logger.Warn("failed to check if package %s is installed: %v", pkg, instErr)
					retry = append(retry, pkg)
					continue
				}
				if !installed {
					retry = append(retry, pkg)
				}
			}

			if len(retry) > 0 {
				ctx.Logger.Stage("Retrying AUR install for %d remaining packages", len(retry))
				ctx.Logger.Debug("AUR retry packages: %s", strings.Join(retry, " "))
				if err2 := h.InstallAURPackages(username, retry...); err2 != nil {
					ctx.Logger.Warn("Retry failed: %v", err2)
				}
			}

			// Determine final failures and log them
			failed := []string{}
			for _, pkg := range missing {
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
				logContent := fmt.Sprintf("%s AUR build failures:\n%s\n", time.Now().Format(time.RFC3339), strings.Join(failed, "\n"))
				if err := h.AppendFile("/tios-log.txt", logContent); err != nil {
					ctx.Logger.Warn("failed to write AUR failure log: %v", err)
				} else {
					ctx.Logger.Warn("AUR build failures logged to /tios-log.txt")
				}
			}
			return nil
		}

		ctx.Logger.Success("AUR packages installed")
	}
	
	return nil
}
