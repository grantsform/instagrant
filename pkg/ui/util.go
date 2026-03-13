package ui

import (
	"fmt"
	"strings"

	"github.com/grantios/instagrant/pkg/config"
)

// renderField renders a selectable field with cursor, label, and value
func renderField(cursor string, label string, value string, isSelected bool) string {
	if isSelected {
		cursor = "▶ "
	} else {
		cursor = "  "
	}
	return fmt.Sprintf("%s%s: %s\n",
		cursor,
		FieldStyle.Render(label),
		ValueStyle.Render(value))
}

// renderToggleField renders a boolean toggle field
func renderToggleField(label string, enabled bool, isSelected bool) string {
	cursor := "  "
	if isSelected {
		cursor = "▶ "
	}
	value := "No"
	if enabled {
		value = "Yes"
	}
	return fmt.Sprintf("%s%s: %s\n",
		cursor,
		FieldStyle.Render(label),
		ValueStyle.Render(value))
}

// formatPackageList formats a package list for display
func formatPackageList(packages []string) string {
	if len(packages) == 0 {
		return "0 packages"
	}
	if len(packages) <= 5 {
		return strings.Join(packages, ", ")
	}
	return fmt.Sprintf("%d packages", len(packages))
}

// parsePreserveDrives parses preserve drives from text input
func parsePreserveDrives(value string) []config.PreserveDrive {
	if value == "" {
		return []config.PreserveDrive{}
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
	return drives
}

// formatPreserveDrives formats preserve drives for display in text input
func formatPreserveDrives(drives []config.PreserveDrive) string {
	if len(drives) == 0 {
		return ""
	}
	parts := []string{}
	for _, d := range drives {
		parts = append(parts, fmt.Sprintf("%s:%s", d.Device, d.MountPoint))
	}
	return strings.Join(parts, " ")
}

// parseExternalDrives parses external drives from text input
func parseExternalDrives(value string) []config.ExternalDrive {
	if value == "" {
		return []config.ExternalDrive{}
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
	return drives
}

// formatExternalDrives formats external drives for display in text input
func formatExternalDrives(drives []config.ExternalDrive) string {
	if len(drives) == 0 {
		return ""
	}
	parts := []string{}
	for _, d := range drives {
		parts = append(parts, fmt.Sprintf("%s:%s:%s:%s", d.Device, d.MountPoint, d.Label, d.Filesystem))
	}
	return strings.Join(parts, " ")
}