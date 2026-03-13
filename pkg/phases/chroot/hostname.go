package chroot

import (
	"fmt"

	"github.com/grantios/instagrant/pkg/phase"
)

// HostnamePhase sets the hostname
type HostnamePhase struct {
	phase.BasePhase
}

func NewHostnamePhase() *HostnamePhase {
	return &HostnamePhase{
		BasePhase: phase.NewBasePhase(
			"hostname",
			"Set hostname",
			true, // isChroot
		),
	}
}

func (h *HostnamePhase) Execute(ctx *phase.Context) error {
	hostname := ctx.Config.System.Hostname
	helper := ctx.Helper
	
	ctx.Logger.Stage("Setting hostname to %s", hostname)
	
	// Write hostname and hosts file
	hostsContent := fmt.Sprintf(`127.0.0.1   localhost
::1         localhost
127.0.1.1   %s.localdomain %s
`, hostname, hostname)
	
	if err := helper.WriteFile("/etc/hostname", hostname+"\n"); err != nil {
		return fmt.Errorf("failed to write hostname: %w", err)
	}
	if err := helper.WriteFile("/etc/hosts", hostsContent); err != nil {
		return err
	}
	
	ctx.Logger.Success("Hostname configured")
	return nil
}
