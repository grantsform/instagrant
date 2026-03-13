package chroot

import (
	"fmt"

	"github.com/grantios/instagrant/pkg/phase"
)

// TiosDirPhase creates /tios directory and clones tios repo
type TiosDirPhase struct {
	phase.BasePhase
}

func NewTiosDirPhase() *TiosDirPhase {
	return &TiosDirPhase{
		BasePhase: phase.NewBasePhase(
			"tiosdir",
			"Create /tios directory and clone tios repository",
			true, // isChroot
		),
	}
}

func (t *TiosDirPhase) Execute(ctx *phase.Context) error {
	exec := ctx.Exec

	ctx.Logger.Stage("Setting up /tios directory")

	// Create /tios directory
	if err := exec.Chroot(ctx.TargetDir, "mkdir -p /tios"); err != nil {
		return fmt.Errorf("failed to create /tios directory: %w", err)
	}

	// Create /tios/repo directory
	if err := exec.Chroot(ctx.TargetDir, "mkdir -p /tios/repo"); err != nil {
		return fmt.Errorf("failed to create /tios/repo directory: %w", err)
	}

	// Clone tios repository
	ctx.Logger.Info("Cloning tios repository...")
	cloneCmd := fmt.Sprintf("git clone https://github.com/grantsform/grantios /tios/repo")
	if err := exec.Chroot(ctx.TargetDir, cloneCmd); err != nil {
		ctx.Logger.Warn("Failed to clone tios repository: %v", err)
	}

	// Create /tios/bin directory if it doesn't exist
	if err := exec.Chroot(ctx.TargetDir, "mkdir -p /tios/bin"); err != nil {
		return fmt.Errorf("failed to create /tios/bin directory: %w", err)
	}

	// Add /tios/bin to PATH for both root and user via /etc/profile.d
	ctx.Logger.Info("Adding /tios/bin to system PATH...")
	profileContent := `#!/bin/bash
export PATH="/tios/bin:$PATH"
`
	profileCmd := fmt.Sprintf("echo '%s' > /etc/profile.d/tios.sh && chmod +x /etc/profile.d/tios.sh", profileContent)
	if err := exec.Chroot(ctx.TargetDir, profileCmd); err != nil {
		return fmt.Errorf("failed to create /etc/profile.d/tios.sh: %w", err)
	}

	ctx.Logger.Success("/tios directory setup completed")
	return nil
}