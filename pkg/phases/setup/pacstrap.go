package setup

import (
	"fmt"
	"strings"

	"github.com/grantios/instagrant/pkg/phase"
)

// PacstrapPhase installs the base system
type PacstrapPhase struct {
	phase.BasePhase
}

func NewPacstrapPhase() *PacstrapPhase {
	return &PacstrapPhase{
		BasePhase: phase.NewBasePhase(
			"pacstrap",
			"Install base system with pacstrap",
			false, // isChroot
		),
	}
}

func (p *PacstrapPhase) Execute(ctx *phase.Context) error {
	exec := ctx.Exec
	
	ctx.Logger.Stage("Updating package database")
	if err := exec.Run("pacman", "-Sy"); err != nil {
		return fmt.Errorf("failed to update package database: %w", err)
	}
	
	ctx.Logger.Stage("Installing base packages")
	
	// Combine base and extra packages
	packages := append(ctx.Config.Packages.Base, ctx.Config.Packages.Extra...)
	args := append([]string{"-K", ctx.TargetDir}, packages...)
	
	ctx.Logger.Info("Installing %d packages", len(packages))
	ctx.Logger.Debug("Packages: %s", strings.Join(packages, " "))
	
	if err := exec.Run("pacstrap", args...); err != nil {
		return fmt.Errorf("pacstrap failed: %w", err)
	}
	
	ctx.Logger.Success("Base system installed")
	return nil
}
