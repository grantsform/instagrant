package chroot

import (
	"fmt"
	"strings"
	"time"

	"github.com/grantios/instagrant/pkg/phase"
)

// BootloaderPhase installs and configures systemd-boot
type BootloaderPhase struct {
	phase.BasePhase
}

// getQuarter returns the current quarter (Q1, Q2, Q3, Q4)
func getQuarter() string {
	month := time.Now().Month()
	switch {
	case month >= 1 && month <= 3:
		return "Q1"
	case month >= 4 && month <= 6:
		return "Q2"
	case month >= 7 && month <= 9:
		return "Q3"
	case month >= 10 && month <= 12:
		return "Q4"
	}
	return "Q1" // fallback
}

// getBootTitle generates the boot entry title
func getBootTitle() string {
	year := time.Now().Year()
	quarter := getQuarter()
	return fmt.Sprintf("GrantiOS :: ArchLinux %s %d", quarter, year)
}

func NewBootloaderPhase() *BootloaderPhase {
	return &BootloaderPhase{
		BasePhase: phase.NewBasePhase(
			"bootloader",
			"Install and configure bootloader",
			true, // isChroot
		),
	}
}

func (b *BootloaderPhase) Execute(ctx *phase.Context) error {
	h := ctx.Helper
	exec := ctx.Exec
	
	ctx.Logger.Stage("Installing systemd-boot")
	if err := exec.Chroot(ctx.TargetDir, "bootctl install --esp-path=/boot"); err != nil {
		return fmt.Errorf("failed to install bootloader: %w", err)
	}
	
	ctx.Logger.Stage("Configuring bootloader")
	
	// Get root partition UUID using findmnt (more reliable than assuming partition number)
	rootDeviceCmd := "findmnt -n -o SOURCE /"
	rootDevice, err := exec.ChrootOutput(ctx.TargetDir, rootDeviceCmd)
	if err != nil {
		return fmt.Errorf("failed to get root device: %w", err)
	}
	rootDevice = strings.TrimSpace(rootDevice)
	
	// Handle btrfs subvolume notation like /dev/sda3[/@]
	if strings.Contains(rootDevice, "[") {
		rootDevice = strings.Split(rootDevice, "[")[0]
	}
	
	rootUUIDCmd := fmt.Sprintf("blkid -s UUID -o value %s", rootDevice)
	rootUUID, err := exec.ChrootOutput(ctx.TargetDir, rootUUIDCmd)
	if err != nil {
		return fmt.Errorf("failed to get root partition UUID: %w", err)
	}
	rootUUID = strings.TrimSpace(rootUUID)
	
	kernelParams := fmt.Sprintf("root=UUID=%s rootflags=subvol=@ rw %s", rootUUID, ctx.Config.System.BootOptions)
	title := getBootTitle()
	
	// Write loader config
	h.WriteFile("/boot/loader/loader.conf", "default arch-current.conf\ntimeout 3\nconsole-mode max\neditor no\n")
	
	// Create entries for both kernels
	kernels := []struct {
		name string
		label string
	}{
		{"linux", "Current"},
		{"linux-lts", "Stable"},
	}
	
	for i, k := range kernels {
		entryName := fmt.Sprintf("arch-%s.conf", k.name)
		if i == 0 {
			entryName = "arch-current.conf" // Default entry
		}
		
		h.WriteFile(fmt.Sprintf("/boot/loader/entries/%s", entryName), 
			fmt.Sprintf("title   %s (%s)\nlinux   /vmlinuz-%s\ninitrd  /initramfs-%s.img\noptions %s\n", title, k.label, k.name, k.name, kernelParams))
	}
	
	// Update EFI boot variables
	if err := exec.Chroot(ctx.TargetDir, "bootctl update"); err != nil {
		return fmt.Errorf("failed to update bootloader: %w", err)
	}
	
	ctx.Logger.Success("Bootloader configured")
	return nil
}
