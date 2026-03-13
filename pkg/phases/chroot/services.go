package chroot

import (
	"fmt"
	"strings"

	"github.com/grantios/instagrant/pkg/phase"
)

// ServicesPhase enables system services
type ServicesPhase struct {
	phase.BasePhase
}

func NewServicesPhase() *ServicesPhase {
	return &ServicesPhase{
		BasePhase: phase.NewBasePhase(
			"services",
			"Enable system services",
			true, // isChroot
		),
	}
}

func (s *ServicesPhase) Execute(ctx *phase.Context) error {
	services := ctx.Config.System.Services
	h := ctx.Helper
	exec := ctx.Exec
	
	// Add display manager based on desktop
	dmMap := map[string]string{
		"plasma":   "sddm.service",
		"hyprland": "sddm.service",
		"gnome":    "gdm.service",
	}
	
	if dm := dmMap[ctx.Config.System.Desktop]; dm != "" {
		services = append(services, dm)
		
		// Configure SDDM
		if strings.Contains(dm, "sddm") {
			ctx.Logger.Stage("Configuring SDDM")
			h.MkdirAll("/etc/sddm.conf.d", 0755)
			h.WriteFile("/etc/sddm.conf.d/00-keymap.conf", `[General]
Numlock=on

[X11]
ServerArguments=-nolisten tcp -ardelay 200 -arinterval 30

[Theme]
Current=breeze

[Users]
MaximumUid=65000
MinimumUid=1000
`)
			exec.Chroot(ctx.TargetDir, fmt.Sprintf("localectl set-x11-keymap %s", ctx.Config.System.Keymap))
		}
	}
	
	// Replace username placeholder
	for i, svc := range services {
		services[i] = strings.ReplaceAll(svc, "${USERNAME}", ctx.Config.User.Username)
	}
	
	ctx.Logger.Stage("Enabling %d services", len(services))
	ctx.Logger.Debug("Services: %s", strings.Join(services, " "))
	
	if err := h.EnableServices(services...); err != nil {
		ctx.Logger.Warn("Some services failed to enable: %v", err)
	}
	
	ctx.Logger.Success("Services enabled")
	return nil
}
