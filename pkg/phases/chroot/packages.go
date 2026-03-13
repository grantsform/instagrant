package chroot

import (
	"strings"

	"github.com/grantios/instagrant/pkg/phase"
)

// PackagesPhase installs additional packages
type PackagesPhase struct {
	phase.BasePhase
}

func NewPackagesPhase() *PackagesPhase {
	return &PackagesPhase{
		BasePhase: phase.NewBasePhase(
			"pkgs",
			"Install additional packages",
			true, // isChroot
		),
	}
}

func (p *PackagesPhase) Execute(ctx *phase.Context) error {
	// Desktop package sets
	desktopPkgs := map[string][]string{
		// Full KDE Plasma (includes applications)
		"plasma":         {"plasma-meta", "kde-applications-meta", "sddm", "xorg"},
		// Minimal KDE Plasma (core desktop without apps)
		"plasma-minimal": {"plasma-meta", "sddm", "xorg"},
		"hyprland":       {"hyprland", "waybar", "wofi", "kitty", "sddm"},
		"gnome":          {"gnome", "gnome-extra", "gdm"},
	}
	
	// GPU driver packages
	driver := ctx.Config.System.GPUDriver
	if driver == "auto" {
		ctx.Logger.Info("Auto-detecting GPU...")
		driver = "mesa" // Simplified
	}
	
	gpuPkgs := map[string][]string{
		"nvidia": {"nvidia", "nvidia-utils", "nvidia-lts", "nvidia-utils"},
		"amd":    {"mesa", "vulkan-radeon", "libva-mesa-driver"},
		"intel":  {"mesa", "vulkan-intel", "intel-media-driver"},
		"mesa":   {"mesa"},
	}
	
	// Note: Since we install both linux and linux-lts kernels, we include both nvidia drivers
	
	// Combine packages
	packages := append(desktopPkgs[ctx.Config.System.Desktop], gpuPkgs[driver]...)
	
	if len(packages) == 0 {
		ctx.Logger.Info("No additional packages to install")
		return nil
	}
	
	ctx.Logger.Stage("Installing %d packages", len(packages))
	ctx.Logger.Debug("Packages: %s", strings.Join(packages, " "))
	
	return ctx.Helper.Pacman(packages...)
}
