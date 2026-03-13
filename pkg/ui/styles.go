package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Shared styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			PaddingLeft(2).
			PaddingBottom(1)

	AsciiStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 2)

	GreenBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#10B981")).
			Padding(1, 2)

	// Status colors
	PendingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	RunningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true)
	CompleteStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	FailedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true)
	DescStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB"))
	LogStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB"))

	// Config UI styles
	SelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	NormalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB"))
	FieldStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	ValueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
	HintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Italic(true)
	CategoryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true).Underline(true)
	WarningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true)

	// Additional config styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true).
			Padding(1, 2)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 2)
)

// Phase categories for grouping
var (
	SetupPhases  = []string{"presetup", "partition", "mount", "pacstrap", "fstab"}
	ChrootPhases = []string{"timezone", "hostname", "user", "pkgs", "aur", "bootloader", "services", "skeleton", "snapper"}
	AfterPhases  = []string{"cleanup"}
)

// Helper to check if slice contains string
func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// GetCategory returns the category name for a phase
func GetCategory(phaseName string) string {
	if Contains(SetupPhases, phaseName) {
		return "SETUP"
	}
	if Contains(ChrootPhases, phaseName) {
		return "CHROOT"
	}
	if Contains(AfterPhases, phaseName) {
		return "AFTER"
	}
	return ""
}

// TruncateString truncates a string to maxLen
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// WrapString wraps text to fit within maxWidth
func WrapString(s string, maxWidth int) []string {
	if len(s) == 0 {
		return []string{""}
	}
	
	var lines []string
	for i := 0; i < len(s); i += maxWidth {
		end := i + maxWidth
		if end > len(s) {
			end = len(s)
		}
		lines = append(lines, s[i:end])
	}
	return lines
}
