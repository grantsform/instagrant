package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/grantios/instagrant/pkg/config"
	"github.com/grantios/instagrant/pkg/util"
)

func (m *ConfigModel) View() string {
	if m.width < 40 || m.height < 15 {
		return "Terminal too small"
	}

	// Show disclaimer first if not agreed
	if m.showingDisclaimer {
		return m.buildDisclaimerView()
	}
	
	// Show requirements after disclaimer
	if m.showingRequirements {
		return m.buildRequirementsView()
	}

	var b strings.Builder

	// GRANTIOS header
	grantiosLines := []string{
		" ██████╗ ██████╗  █████╗ ███╗   ██╗████████╗██╗ ██████╗ ███████╗",
		"██╔════╝ ██╔══██╗██╔══██╗████╗  ██║╚══██╔══╝██║██╔═══██╗██╔════╝",
		"██║  ███╗██████╔╝███████║██╔██╗ ██║   ██║   ██║██║   ██║███████╗",
		"██║   ██║██╔══██╗██╔══██║██║╚██╗██║   ██║   ██║██║   ██║╚════██║",
		"╚██████╔╝██║  ██║██║  ██║██║ ╚████║   ██║   ██║╚██████╔╝███████║",
		" ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝   ╚═╝ ╚═════╝ ╚══════╝",
	}
	
	// Render the art left-aligned
	for _, line := range grantiosLines {
		b.WriteString(AsciiStyle.Render(line) + "\n")
	}
	b.WriteString("\n")

	// Show disk overwrite warning if active
	if m.showingDiskWarn {
		b.WriteString(FailedStyle.Render("⚠️  WARNING: Disk Overwrite Confirmation"))
		b.WriteString("\n\n")
		
		b.WriteString("You are about to start the installation process.\n")
		b.WriteString("This will COMPLETELY ERASE all data on the selected disks.\n\n")
		
		b.WriteString(fmt.Sprintf("Target Disk: %s (%s)\n", m.config.Disk.Device, m.diskSize))
		
		// List external drives that will be formatted
		if len(m.config.External) > 0 {
			b.WriteString("\nExternal Drives (will be formatted):\n")
			for _, ext := range m.config.External {
				b.WriteString(fmt.Sprintf("  • %s → %s (%s)\n", ext.Device, ext.MountPoint, ext.Filesystem))
			}
		}
		
		b.WriteString("\n")
		b.WriteString(FailedStyle.Render("All partitions and data will be permanently lost."))
		b.WriteString("\n\n")
		
		b.WriteString("Are you sure you want to continue?\n\n")
		
		b.WriteString(HintStyle.Render("Press [y] to continue or [n] to go back"))
		b.WriteString("\n")
		b.WriteString(HintStyle.Render("(Confirm with lowercase)"))
		
		content := BoxStyle.Render(b.String())
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	// Show final confirmation if active
	if m.showingLastChance {
		b.WriteString(FailedStyle.Render("🚨 LAST CHANCE - NO GOING BACK AFTER THIS!"))
		b.WriteString("\n\n")
		
		b.WriteString("This is your FINAL confirmation.\n\n")
		
		b.WriteString(fmt.Sprintf("Target Disk: %s (%s)\n", m.config.Disk.Device, m.diskSize))
		
		// List external drives that will be formatted
		if len(m.config.External) > 0 {
			b.WriteString("\nExternal Drives (will be formatted):\n")
			for _, ext := range m.config.External {
				b.WriteString(fmt.Sprintf("  • %s → %s (%s)\n", ext.Device, ext.MountPoint, ext.Filesystem))
			}
			b.WriteString("\n")
		}
		
		b.WriteString(fmt.Sprintf("Hostname: %s\n", m.config.System.Hostname))
		b.WriteString(fmt.Sprintf("Username: %s\n\n", m.config.User.Username))
		
		b.WriteString(FailedStyle.Render("Once you press Y, ALL disks above will be wiped immediately."))
		b.WriteString("\n\n")
		
		b.WriteString("Proceed with installation?\n\n")
		
		b.WriteString(HintStyle.Render("Press [Y] to START INSTALLATION or [N] to go back"))
		b.WriteString("\n")
		b.WriteString(HintStyle.Render("(Confirm with UPPERCASE)"))
		
		content := BoxStyle.Render(b.String())
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	// Show preserve drive selector if active
	if m.showingPreserveSelector {
		b.WriteString(HeaderStyle.Render("Select Preserve Drive"))
		b.WriteString("\n\n")
		
		switch m.inputStep {
		case 0: // Select disk
			b.WriteString("Available disks:\n\n")
			for i, disk := range m.availableDisks {
				cursor := "  "
				if i == m.selectedDiskIndex {
					cursor = "▶ "
				}
				size := "unknown size"
				if s, err := util.GetDiskSize(disk); err == nil {
					size = util.FormatBytes(s)
				}
				b.WriteString(fmt.Sprintf("%s%s (%s)\n", cursor, disk, size))
			}
			b.WriteString("\n")
			b.WriteString(HintStyle.Render("↑/↓: Navigate | Enter: Select | Esc: Cancel"))
		case 1: // Input mount point
			b.WriteString(fmt.Sprintf("Selected disk: %s\n\n", m.currentPreserve.Device))
			b.WriteString("Mount point:\n")
			b.WriteString(m.textInput.View() + "\n")
			b.WriteString(HintStyle.Render("Enter: Add drive | Esc: Back"))
		}
		
		content := BoxStyle.Render(b.String())
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	// Show external drive selector if active
	if m.showingExternalSelector {
		b.WriteString(HeaderStyle.Render("Select External Drive"))
		b.WriteString("\n\n")
		
		switch m.inputStep {
		case 0: // Select disk
			b.WriteString("Available disks:\n\n")
			for i, disk := range m.availableDisks {
				cursor := "  "
				if i == m.selectedDiskIndex {
					cursor = "▶ "
				}
				size := "unknown size"
				if s, err := util.GetDiskSize(disk); err == nil {
					size = util.FormatBytes(s)
				}
				b.WriteString(fmt.Sprintf("%s%s (%s)\n", cursor, disk, size))
			}
			b.WriteString("\n")
			b.WriteString(HintStyle.Render("↑/↓: Navigate | Enter: Select | Esc: Cancel"))
		case 1: // Input name
			b.WriteString(fmt.Sprintf("Selected disk: %s\n\n", m.currentDrive.Device))
			b.WriteString("Mount name:\n")
			b.WriteString(m.textInput.View() + "\n")
			b.WriteString(HintStyle.Render("Enter: Confirm | Esc: Back"))
		case 2: // Input mount point
			b.WriteString(fmt.Sprintf("Selected disk: %s\n", m.currentDrive.Device))
			b.WriteString(fmt.Sprintf("Mount name: %s\n\n", m.currentDrive.Label))
			b.WriteString("Mount point:\n")
			b.WriteString(m.textInput.View() + "\n")
			b.WriteString(HintStyle.Render("Enter: Next | Esc: Back"))
		case 3: // Select filesystem
			filesystems := []string{"xfs", "ext4", "btrfs", "ntfs", "exfat"}
			b.WriteString(fmt.Sprintf("Selected disk: %s\n", m.currentDrive.Device))
			b.WriteString(fmt.Sprintf("Mount name: %s\n", m.currentDrive.Label))
			b.WriteString(fmt.Sprintf("Mount point: %s\n\n", m.currentDrive.MountPoint))
			b.WriteString("Filesystem:\n")
			for i, fs := range filesystems {
				cursor := "  "
				if i == m.selectedFsIndex {
					cursor = "▶ "
				}
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, fs))
			}
			b.WriteString("\n")
			b.WriteString(HintStyle.Render("↑/↓: Navigate | Enter: Add drive | Esc: Back"))
		}
		
		content := BoxStyle.Render(b.String())
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	if m.config == nil {
		// Show config selection
		b.WriteString(HeaderStyle.Render("Select Installation Configuration"))
		b.WriteString("\n\n")

		for i, cfg := range m.configs {
			if cfg == "---" {
				// Render separator
				b.WriteString("  " + HintStyle.Render("────────────────────") + "\n")
				continue
			}
			cursor := "  "
			style := NormalStyle
			if i == m.selectedIndex {
				cursor = "▶ "
				style = SelectedStyle
			}
			// Add indicator for local configs that can be deleted
			configDisplay := cfg
			if config.IsLocalConfig(cfg) {
				configDisplay = cfg + " " + HintStyle.Render("[local]")
			}
			b.WriteString(cursor + style.Render(configDisplay) + "\n")
		}

		b.WriteString("\n")

		if m.errorMsg != "" {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("Error: " + m.errorMsg) + "\n\n")
		}

		// Show delete hint if on a local config
		if m.selectedIndex < len(m.configs) && m.configs[m.selectedIndex] != "---" && config.IsLocalConfig(m.configs[m.selectedIndex]) {
			b.WriteString(HintStyle.Render("↑/↓: Navigate | Enter: Select | d: Delete | Ctrl+C: Quit"))
		} else {
			b.WriteString(HintStyle.Render("↑/↓: Navigate | Enter: Select | Ctrl+C: Quit"))
		}
	} else {
		// Show config details and editing
		b.WriteString(HeaderStyle.Render(fmt.Sprintf("Configuration: %s", m.config.Profile)))
		b.WriteString("\n\n")

		fields := []struct {
			label string
			value string
			field string
		}{
			{"Target", m.config.Target, "target"},
			{"Device", m.config.Disk.Device, "device"},
			{"Hostname", m.config.System.Hostname, "hostname"},
			{"Username", m.config.User.Username, "username"},
			{"Password", strings.Repeat("*", len(m.config.User.Password)), "password"},
			{"Root Password", strings.Repeat("*", len(m.config.User.RootPassword)), "rootpassword"},
			{"Home Dir", m.config.User.HomeDir, "homedir"},
			{"Timezone", m.config.System.Timezone, "timezone"},
			{"Locale", m.config.System.Locale, "locale"},
			{"Keymap", m.config.System.Keymap, "keymap"},
		}

		for i, f := range fields {
			cursor := "  "
			if i+len(m.configs) == m.selectedIndex {
				cursor = "▶ "
			}

			line := fmt.Sprintf("%s%s: %s",
				cursor,
				FieldStyle.Render(f.label),
				ValueStyle.Render(f.value))
			b.WriteString(line + "\n")
		}

		// Legacy boot
		var cursor string
		cursor = "  "
		if len(m.configs)+10 == m.selectedIndex {
			cursor = "▶ "
		}
		legacyBootValue := "No"
		if m.config.Disk.LegacyBoot {
			legacyBootValue = "Yes"
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n",
			cursor,
			FieldStyle.Render("Legacy Boot"),
			ValueStyle.Render(legacyBootValue)))

		// Extra packages
		cursor = "  "
		if len(m.configs)+11 == m.selectedIndex {
			cursor = "▶ "
		}
		extraPkgs := fmt.Sprintf("%d packages", len(m.config.Packages.Extra))
		if len(m.config.Packages.Extra) > 0 && len(m.config.Packages.Extra) <= 5 {
			extraPkgs = strings.Join(m.config.Packages.Extra, ", ")
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n",
			cursor,
			FieldStyle.Render("Extra Packages"),
			ValueStyle.Render(extraPkgs)))

		// AUR packages
		cursor = "  "
		if len(m.configs)+12 == m.selectedIndex {
			cursor = "▶ "
		}
		aurPkgs := fmt.Sprintf("%d packages", len(m.config.Packages.AUR))
		if len(m.config.Packages.AUR) > 0 && len(m.config.Packages.AUR) <= 5 {
			aurPkgs = strings.Join(m.config.Packages.AUR, ", ")
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n",
			cursor,
			FieldStyle.Render("AUR Packages"),
			ValueStyle.Render(aurPkgs)))

		// Preserve room
		cursor = "  "
		if len(m.configs)+13 == m.selectedIndex {
			cursor = "▶ "
		}
		preserveRoomValue := "No"
		if m.config.Disk.PreserveRoom {
			preserveRoomValue = "Yes"
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n",
			cursor,
			FieldStyle.Render("Preserve Room"),
			ValueStyle.Render(preserveRoomValue)))

		// Preserve drives
		cursor = "  "
		if len(m.configs)+14 == m.selectedIndex {
			cursor = "▶ "
		}
		var preserveDrivesDisplay string
		if len(m.config.Preserve) == 0 {
			preserveDrivesDisplay = HintStyle.Render("none (e.g., /dev/sdb1:/mnt/data)")
		} else if len(m.config.Preserve) <= 2 {
			drives := []string{}
			for _, d := range m.config.Preserve {
				drives = append(drives, fmt.Sprintf("%s→%s", d.Device, d.MountPoint))
			}
			preserveDrivesDisplay = ValueStyle.Render(strings.Join(drives, ", "))
		} else {
			preserveDrivesDisplay = ValueStyle.Render(fmt.Sprintf("%d drives", len(m.config.Preserve)))
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n",
			cursor,
			FieldStyle.Render("Preserve Drives"),
			preserveDrivesDisplay))

		// External drives
		cursor = "  "
		if len(m.configs)+15 == m.selectedIndex {
			cursor = "▶ "
		}
		var externalDrivesDisplay string
		if len(m.config.External) == 0 {
			externalDrivesDisplay = HintStyle.Render("none (e.g., /dev/sdb:/mnt/media:media:xfs)")
		} else if len(m.config.External) <= 2 {
			drives := []string{}
			for _, d := range m.config.External {
				drives = append(drives, fmt.Sprintf("%s→%s", d.Device, d.MountPoint))
			}
			externalDrivesDisplay = ValueStyle.Render(strings.Join(drives, ", "))
		} else {
			externalDrivesDisplay = ValueStyle.Render(fmt.Sprintf("%d drives", len(m.config.External)))
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n",
			cursor,
			FieldStyle.Render("External Drives"),
			externalDrivesDisplay))

		// Disable Skeleton toggle
		cursor = "  "
		if len(m.configs)+16 == m.selectedIndex {
			cursor = "▶ "
		}
		disableSkeleton := "No"
		if m.config.Skeleton.Default == "" && m.config.Skeleton.Profile == "" {
			disableSkeleton = "Yes"
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n",
			cursor,
			FieldStyle.Render("Disable Skeleton"),
			ValueStyle.Render(disableSkeleton)))

		b.WriteString("\n")

		// Separator before action button
		b.WriteString(HintStyle.Render("  ────────────────────") + "\n\n")

		// Confirm button - changes based on config type
		var confirmText string
		if m.configName == "default" {
			confirmText = "[ Confirm and Use as Default ]"
		} else if m.configName == "template" {
			confirmText = "[ Confirm and Make New Template ]"
		} else {
			confirmText = "[ Confirm and Start Installation ]"
		}

		if len(m.configs)+17 == m.selectedIndex {
			confirmText = SelectedStyle.Render("▶ " + confirmText)
		} else {
			confirmText = NormalStyle.Render("  " + confirmText)
		}
		b.WriteString(confirmText + "\n\n")

		if m.askingForName {
			b.WriteString(HintStyle.Render("Enter template name:\n"))
			b.WriteString(m.textInput.View() + "\n")
			b.WriteString(HintStyle.Render("Enter: Save template | Esc: Cancel"))
		} else if m.editing {
			b.WriteString(HintStyle.Render("Editing: ") + m.editField + "\n")
			b.WriteString(m.textInput.View() + "\n")
			if m.editField == "password" || m.editField == "rootpassword" {
				visMode := "hidden"
				if m.showPassword {
					visMode = "visible"
				}
				b.WriteString(HintStyle.Render(fmt.Sprintf("Tab: Toggle visibility (%s) | Enter: Save | Esc: Cancel", visMode)))
			} else {
				b.WriteString(HintStyle.Render("Enter: Save | Esc: Cancel"))
			}
		} else {
			b.WriteString(HintStyle.Render("↑/↓: Navigate | Enter/e: Edit | Ctrl+C: Quit"))
		}
	}

	content := BoxStyle.Render(b.String())
	
	// Center the content in the terminal
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *ConfigModel) buildDisclaimerView() string {
	var b strings.Builder

	// GRANTIOS header
	grantiosLines := []string{
		" ██████╗ ██████╗  █████╗ ███╗   ██╗████████╗██╗ ██████╗ ███████╗",
		"██╔════╝ ██╔══██╗██╔══██╗████╗  ██║╚══██╔══╝██║██╔═══██╗██╔════╝",
		"██║  ███╗██████╔╝███████║██╔██╗ ██║   ██║   ██║██║   ██║███████╗",
		"██║   ██║██╔══██╗██╔══██║██║╚██╗██║   ██║   ██║██║   ██║╚════██║",
		"╚██████╔╝██║  ██║██║  ██║██║ ╚████║   ██║   ██║╚██████╔╝███████║",
		" ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝   ╚═╝ ╚═════╝ ╚══════╝",
	}

	for _, line := range grantiosLines {
		b.WriteString(AsciiStyle.Render(line) + "\n")
	}
	b.WriteString("\n")

	b.WriteString(HeaderStyle.Render("Welcome to GRANTIOS Installer"))
	b.WriteString("\n\n")

	b.WriteString(WarningStyle.Render("⚠ EXPERIMENTAL SOFTWARE - USE AT YOUR OWN RISK"))
	b.WriteString("\n\n")

	disclaimerText := `This software is provided "AS IS" without warranty of any kind,
express or implied, including but not limited to the warranties
of merchantability, fitness for a particular purpose, and
noninfringement.

@GrantsForm x Joshua Steven Grant (aka Jost Grant) holds ZERO
liability for any damages, data loss, system failures, or any
other issues that may occur from the use of this software.

By proceeding, you acknowledge that:
  • This is experimental software
  • You use it entirely at your own risk
  • The author(s) accept NO civil or legal liability
  • You are solely responsible for any consequences`

	b.WriteString(disclaimerText)
	b.WriteString("\n\n")

	b.WriteString(FailedStyle.Render("─────────────────────────────────────────────────"))
	b.WriteString("\n\n")

	b.WriteString("Press ")
	b.WriteString(WarningStyle.Render("A"))
	b.WriteString(" to AGREE and continue, or ")
	b.WriteString(HintStyle.Render("q"))
	b.WriteString(" to quit.\n")

	content := BoxStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *ConfigModel) buildRequirementsView() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render("⚠️  REQUIREMENTS CHECK"))
	b.WriteString("\n\n")

	b.WriteString(WarningStyle.Render("STORAGE REQUIREMENTS:"))
	b.WriteString("\n")
	b.WriteString("• Minimum 175GB root partition (install may fail with less)\n")
	b.WriteString("• Ideally 320GB+ for Btrfs snapshots and rollbacks\n")
	b.WriteString("• Btrfs provides excellent snapshot capabilities, but needs space for rollbacks\n\n")

	b.WriteString(WarningStyle.Render("NETWORK REQUIREMENTS:"))
	b.WriteString("\n")
	b.WriteString("• Stable internet connection required\n")
	b.WriteString("• Ethernet preferred over WiFi\n")
	b.WriteString("• WiFi may cause failures when installing many extra/AUR packages\n\n")

	b.WriteString(WarningStyle.Render("POWER REQUIREMENTS:"))
	b.WriteString("\n")
	b.WriteString("• Device should be plugged in (not on battery)\n")
	b.WriteString("• Installation process can take time and should not be interrupted\n\n")

	b.WriteString(FailedStyle.Render("─────────────────────────────────────────────────"))
	b.WriteString("\n\n")

	b.WriteString("Press ")
	b.WriteString(WarningStyle.Render("ENTER"))
	b.WriteString(" or ")
	b.WriteString(WarningStyle.Render("SPACE"))
	b.WriteString(" to continue, or ")
	b.WriteString(HintStyle.Render("q"))
	b.WriteString(" to quit.\n")

	content := BoxStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}