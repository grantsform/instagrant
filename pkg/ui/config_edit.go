package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/grantios/instagrant/pkg/config"
)

func (m *ConfigModel) startEdit() {
	if m.config == nil {
		return
	}

	fieldIndex := m.selectedIndex - len(m.configs)
	if fieldIndex < 0 || fieldIndex > 16 {
		return
	}

	// These use toggles or wizards, skip text input:
	// 10: Legacy boot (toggle)
	// 11: Extra packages (space-separated text)
	// 12: AUR packages (space-separated text)
	// 13: Preserve room (toggle)
	// 14: Preserve drives (wizard)
	// 15: External drives (wizard)
	// 16: Skeleton disable (toggle)
	if fieldIndex == 10 || fieldIndex == 13 || fieldIndex == 14 || fieldIndex == 15 || fieldIndex == 16 {
		return
	}

	fields := []struct {
		name  string
		value string
	}{
		{"target", m.config.Target},            // 0
		{"device", m.config.Disk.Device},       // 1
		{"hostname", m.config.System.Hostname}, // 2
		{"username", m.config.User.Username},   // 3
		{"password", m.config.User.Password},   // 4
		{"rootpassword", m.config.User.RootPassword}, // 5
		{"homedir", m.config.User.HomeDir},     // 6
		{"timezone", m.config.System.Timezone}, // 7
		{"locale", m.config.System.Locale},     // 8
		{"keymap", m.config.System.Keymap},     // 9
		{"", ""},                               // 10: legacy boot (toggle, skip)
		{"extra", strings.Join(m.config.Packages.Extra, " ")}, // 11
		{"aur", strings.Join(m.config.Packages.AUR, " ")},     // 12
	}

	if fieldIndex < len(fields) && fields[fieldIndex].name != "" {
		m.editField = fields[fieldIndex].name
		m.textInput.SetValue(fields[fieldIndex].value)

		// Set echo mode for password fields
		if m.editField == "password" || m.editField == "rootpassword" {
			m.textInput.EchoMode = textinput.EchoPassword
			m.showPassword = false
		} else {
			m.textInput.EchoMode = textinput.EchoNormal
		}

		m.editing = true
	}
}

func (m *ConfigModel) applyEdit() {
	if m.config == nil || m.editField == "" {
		return
	}

	value := m.textInput.Value()

	switch m.editField {
	case "target":
		m.config.Target = value
	case "device":
		m.config.Disk.Device = value
	case "hostname":
		m.config.System.Hostname = value
	case "username":
		m.config.User.Username = value
	case "password":
		m.config.User.Password = value
	case "rootpassword":
		m.config.User.RootPassword = value
	case "homedir":
		m.config.User.HomeDir = value
	case "timezone":
		m.config.System.Timezone = value
	case "locale":
		m.config.System.Locale = value
	case "keymap":
		m.config.System.Keymap = value
	case "legacy":
		// Parse boolean value for legacy boot
		m.config.Disk.LegacyBoot = (strings.ToLower(strings.TrimSpace(value)) == "true" || 
			strings.ToLower(strings.TrimSpace(value)) == "yes" || 
			strings.ToLower(strings.TrimSpace(value)) == "1")
	case "preserve-room":
		// Parse boolean value for preserve room
		m.config.Disk.PreserveRoom = (strings.ToLower(strings.TrimSpace(value)) == "true" || 
			strings.ToLower(strings.TrimSpace(value)) == "yes" || 
			strings.ToLower(strings.TrimSpace(value)) == "1")
	case "extra":
		// Parse space-separated package list
		if value == "" {
			m.config.Packages.Extra = []string{}
		} else {
			m.config.Packages.Extra = strings.Fields(value)
		}
	case "aur":
		// Parse space-separated package list
		if value == "" {
			m.config.Packages.AUR = []string{}
		} else {
			m.config.Packages.AUR = strings.Fields(value)
		}
	case "external":
		m.parseExternalDrives(value)
	}
}

func (m *ConfigModel) formatPreserveDrives() string {
	if len(m.config.Preserve) == 0 {
		return ""
	}
	parts := []string{}
	for _, d := range m.config.Preserve {
		parts = append(parts, fmt.Sprintf("%s:%s", d.Device, d.MountPoint))
	}
	return strings.Join(parts, " ")
}

func (m *ConfigModel) parsePreserveDrives(value string) {
	if value == "" {
		m.config.Preserve = []config.PreserveDrive{}
		return
	}

	drives := []config.PreserveDrive{}
	for _, part := range strings.Fields(value) {
		pieces := strings.SplitN(part, ":", 2)
		if len(pieces) == 2 {
			drives = append(drives, config.PreserveDrive{
				Device:     pieces[0],
				MountPoint: pieces[1],
			})
		}
	}
	m.config.Preserve = drives
}

func (m *ConfigModel) formatExternalDrives() string {
	if len(m.config.External) == 0 {
		return ""
	}
	parts := []string{}
	for _, d := range m.config.External {
		parts = append(parts, fmt.Sprintf("%s:%s:%s:%s", d.Device, d.MountPoint, d.Label, d.Filesystem))
	}
	return strings.Join(parts, " ")
}

func (m *ConfigModel) parseExternalDrives(value string) {
	if value == "" {
		m.config.External = []config.ExternalDrive{}
		return
	}

	drives := []config.ExternalDrive{}
	for _, part := range strings.Fields(value) {
		pieces := strings.SplitN(part, ":", 4)
		if len(pieces) == 4 {
			drives = append(drives, config.ExternalDrive{
				Device:     pieces[0],
				MountPoint: pieces[1],
				Label:      pieces[2],
				Filesystem: pieces[3],
			})
		}
	}
	m.config.External = drives
}