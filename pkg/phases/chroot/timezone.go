package chroot

import (
	"fmt"

	"github.com/grantios/instagrant/pkg/phase"
)

// TimezonePhase configures timezone and locale
type TimezonePhase struct {
	phase.BasePhase
}

func NewTimezonePhase() *TimezonePhase {
	return &TimezonePhase{
		BasePhase: phase.NewBasePhase(
			"timezone",
			"Configure timezone and locale",
			true, // isChroot
		),
	}
}

func (t *TimezonePhase) Execute(ctx *phase.Context) error {
	cfg := ctx.Config
	exec := ctx.Exec
	h := ctx.Helper
	
	ctx.Logger.Stage("Setting timezone to %s", cfg.System.Timezone)
	
	// Set timezone and hardware clock
	if err := exec.Chroot(ctx.TargetDir, fmt.Sprintf("ln -sf /usr/share/zoneinfo/%s /etc/localtime && hwclock --systohc", cfg.System.Timezone)); err != nil {
		return fmt.Errorf("failed to set timezone: %w", err)
	}
	
	ctx.Logger.Stage("Setting locale to %s", cfg.System.Locale)
	
	// Uncomment locale and generate
	localeCmd := fmt.Sprintf("sed -i 's/^#%s/%s/' /etc/locale.gen && locale-gen", cfg.System.Locale, cfg.System.Locale)
	if err := exec.Chroot(ctx.TargetDir, localeCmd); err != nil {
		return fmt.Errorf("failed to configure locale: %w", err)
	}
	
	// Write config files
	h.WriteFile("/etc/locale.conf", fmt.Sprintf("LANG=%s\n", cfg.System.Locale))
	h.WriteFile("/etc/vconsole.conf", fmt.Sprintf("KEYMAP=%s\n", cfg.System.Keymap))
	
	ctx.Logger.Success("Timezone and locale configured")
	return nil
}
