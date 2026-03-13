package setup

import (
	"fmt"

	"github.com/grantios/instagrant/pkg/phase"
)

// FstabPhase generates fstab
type FstabPhase struct {
	phase.BasePhase
}

func NewFstabPhase() *FstabPhase {
	return &FstabPhase{
		BasePhase: phase.NewBasePhase(
			"fstab",
			"Generate fstab",
			false, // isChroot
		),
	}
}

func (f *FstabPhase) Execute(ctx *phase.Context) error {
	ctx.Logger.Stage("Generating fstab")
	
	if err := ctx.Exec.RunSh(fmt.Sprintf("genfstab -U %s >> %s/etc/fstab", ctx.TargetDir, ctx.TargetDir)); err != nil {
		return fmt.Errorf("failed to generate fstab: %w", err)
	}
	
	ctx.Logger.Success("fstab generated")
	return nil
}
