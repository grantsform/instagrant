package ui

import (
	"fmt"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PhaseStatus represents the current status of a phase
type PhaseStatus int

const (
	StatusPending PhaseStatus = iota
	StatusRunning
	StatusComplete
	StatusFailed
)

// PhaseInfo contains information about an installation phase
type PhaseInfo struct {
	Name        string
	Description string
	Status      PhaseStatus
}

// Model represents the TUI state
type Model struct {
	phases          []PhaseInfo
	currentPhase    int
	logs            []string
	logOffset       int // 0 = bottom, positive = scrolled up
	width           int
	height          int
	mu              sync.Mutex
	postInstall     bool
	selectedOption  int
	exitAction      int // 0 = none, 1 = reboot, 2 = chroot, 3 = exit
	confirmingFirst bool // first confirmation (lowercase y/n)
	confirmingFinal bool // final confirmation (uppercase Y/N)
}

// Messages for updating the UI
type (
	PhaseUpdateMsg struct {
		Index  int
		Status PhaseStatus
	}
	LogMsg struct {
		Text string
	}
)

func NewModel(phases []PhaseInfo) *Model {
	return &Model{
		phases:       phases,
		currentPhase: -1,
		logs:         []string{},
		width:        80,
		height:       24,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.postInstall {
			// Handle final confirmation (uppercase Y/N)
			if m.confirmingFinal {
				switch msg.String() {
				case "Y":
					return m, m.executePostInstallOption()
				case "N", "n", "y", "escape":
					m.confirmingFinal = false
					m.confirmingFirst = false
				}
				return m, nil
			}
			// Handle first confirmation (lowercase y/n)
			if m.confirmingFirst {
				switch msg.String() {
				case "y":
					m.confirmingFinal = true
				case "n", "Y", "N", "escape":
					m.confirmingFirst = false
				}
				return m, nil
			}
			// Normal menu navigation
			switch msg.String() {
			case "up", "k":
				if m.selectedOption > 0 {
					m.selectedOption--
				}
			case "down", "j":
				if m.selectedOption < 2 {
					m.selectedOption++
				}
			case "enter":
				m.confirmingFirst = true
				return m, nil
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		} else {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up":
				m.logOffset++
				if m.logOffset > len(m.logs) {
					m.logOffset = len(m.logs)
				}
			case "down":
				m.logOffset--
				if m.logOffset < 0 {
					m.logOffset = 0
				}
			case "pgup":
				m.logOffset += m.height - 10
				if m.logOffset > len(m.logs) {
					m.logOffset = len(m.logs)
				}
			case "pgdown":
				m.logOffset -= m.height - 10
				if m.logOffset < 0 {
					m.logOffset = 0
				}
			case "home":
				m.logOffset = 0
			case "end":
				m.logOffset = len(m.logs)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case PhaseUpdateMsg:
		if msg.Index >= 0 && msg.Index < len(m.phases) {
			m.phases[msg.Index].Status = msg.Status
			if msg.Status == StatusRunning {
				m.currentPhase = msg.Index
			}
		}
		// Check if all phases are complete
		allComplete := true
		for _, phase := range m.phases {
			if phase.Status != StatusComplete && phase.Status != StatusFailed {
				allComplete = false
				break
			}
		}
		if allComplete && !m.postInstall {
			m.postInstall = true
			m.selectedOption = 0
		}

	case LogMsg:
		m.logs = append(m.logs, msg.Text)
		if len(m.logs) > 1000 {
			m.logs = m.logs[len(m.logs)-1000:]
		}
	}

	return m, nil
}

func (m *Model) View() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.postInstall {
		return m.buildPostInstallView()
	}

	if m.width < 40 || m.height < 10 {
		return "Terminal too small"
	}

	leftWidth := 40
	rightWidth := m.width - leftWidth - 6
	contentHeight := m.height - 2

	// Build phase list
	phasePanel := m.buildPhasePanel(leftWidth, contentHeight)
	logPanel := m.buildLogPanel(rightWidth, contentHeight)

	// Join panels and place in fixed container to prevent overflow
	combined := lipgloss.JoinHorizontal(lipgloss.Top, phasePanel, logPanel)
	return lipgloss.NewStyle().MaxHeight(m.height).MaxWidth(m.width).Render(combined)
}

func (m *Model) buildPhasePanel(width, height int) string {
	var lines []string
	var lastCat string
	phaseLineIndices := make([]int, len(m.phases))

	// Build phase lines with categories
	for i, phase := range m.phases {
		cat := GetCategory(phase.Name)
		if cat != lastCat && cat != "" {
			if lastCat != "" {
				lines = append(lines, "")
			}
			lines = append(lines, CategoryStyle.Render(cat))
			lastCat = cat
		}

		phaseLineIndices[i] = len(lines)

		// Icon and style based on status
		icon, style := m.getPhaseStyle(phase.Status)
		line := fmt.Sprintf("%s %02d. %s", icon, i+1, phase.Name)
		lines = append(lines, style.Render(line))
		
		// Only show description for active (running) phase
		if phase.Status == StatusRunning {
			lines = append(lines, DescStyle.Render("   "+phase.Description))
		}
	}

	// Scroll to keep current phase visible
	maxLines := height - 6
	startIdx := 0
	if len(lines) > maxLines && m.currentPhase >= 0 && m.currentPhase < len(phaseLineIndices) {
		currentPhaseStart := phaseLineIndices[m.currentPhase]
		startIdx = currentPhaseStart - maxLines/2
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx+maxLines > len(lines) {
			startIdx = len(lines) - maxLines
		}
	}

	endIdx := startIdx + maxLines
	if endIdx > len(lines) {
		endIdx = len(lines)
	}

	content := TitleStyle.Render("Installation Phases") + "\n"
	content += strings.Join(lines[startIdx:endIdx], "\n")

	return BorderStyle.Width(width).Height(height).Render(content)
}

func (m *Model) buildLogPanel(width, height int) string {
	logWidth := width - 6
	logHeight := height - 5 // Reserve space for progress bar

	// Wrap and prepare log lines
	var wrappedLogs []string
	for _, log := range m.logs {
		wrappedLogs = append(wrappedLogs, WrapString(log, logWidth)...)
	}

	// Apply scroll offset
	totalLines := len(wrappedLogs)
	var visibleLines []string

	if totalLines > logHeight {
		endIdx := totalLines - m.logOffset
		startIdx := endIdx - logHeight
		if startIdx < 0 {
			startIdx = 0
		}
		if endIdx > totalLines {
			endIdx = totalLines
		}
		if endIdx > startIdx {
			visibleLines = wrappedLogs[startIdx:endIdx]
		}
	} else {
		visibleLines = wrappedLogs
	}

	// Build content
	scrollInfo := ""
	if m.logOffset > 0 {
		scrollInfo = fmt.Sprintf(" (↑%d)", m.logOffset)
	}

	content := TitleStyle.Render("Output"+scrollInfo) + "\n"
	for i := 0; i < len(visibleLines) && i < logHeight; i++ {
		content += LogStyle.Render(TruncateString(visibleLines[i], logWidth)) + "\n"
	}

	// Add progress bar at the bottom
	progress := fmt.Sprintf("Phase: %02d/%02d", m.currentPhase+1, len(m.phases))
	content += lipgloss.PlaceHorizontal(logWidth, lipgloss.Right, progress) + "\n"

	return GreenBorder.Width(width).Height(height).Render(content)
}

func (m *Model) buildPostInstallView() string {
	// Show confirmation dialog if confirming
	if m.confirmingFirst || m.confirmingFinal {
		return m.buildConfirmationView()
	}

	var b strings.Builder

	b.WriteString(HeaderStyle.Render("Installation Complete!"))
	b.WriteString("\n\n")

	b.WriteString("What would you like to do?\n\n")

	options := []string{
		"Unmount and reboot",
		"Enter chroot environment",
		"Exit installer",
	}

	for i, option := range options {
		cursor := "  "
		if i == m.selectedOption {
			cursor = "▶ "
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, option))
	}

	b.WriteString("\n")
	b.WriteString(HintStyle.Render("↑/↓: Navigate | Enter: Select | Ctrl+C: Exit"))

	content := BoxStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *Model) buildConfirmationView() string {
	var b strings.Builder

	// Get the action text based on selected option
	var actionTitle, actionDesc string
	switch m.selectedOption {
	case 0:
		actionTitle = "Unmount and Reboot"
		actionDesc = "This will unmount all filesystems and reboot the system."
	case 1:
		actionTitle = "Enter Chroot"
		actionDesc = "This will drop you into a shell in the new system.\nType 'exit' to leave when done."
	case 2:
		actionTitle = "Exit Installer"
		actionDesc = "This will exit the installer.\nFilesystems will remain mounted."
	}

	b.WriteString(WarningStyle.Render("⚠ " + actionTitle))
	b.WriteString("\n\n")
	b.WriteString(actionDesc)
	b.WriteString("\n\n")

	if m.confirmingFinal {
		b.WriteString(WarningStyle.Render("Are you SURE?"))
		b.WriteString("\n\n")
		b.WriteString(HintStyle.Render("Press Y to confirm, N to cancel"))
	} else {
		b.WriteString("Proceed?\n\n")
		b.WriteString(HintStyle.Render("Press y to continue, n to cancel"))
	}

	content := BoxStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *Model) getPhaseStyle(status PhaseStatus) (string, lipgloss.Style) {
	switch status {
	case StatusPending:
		return "○", PendingStyle
	case StatusRunning:
		return "●", RunningStyle
	case StatusComplete:
		return "✓", CompleteStyle
	case StatusFailed:
		return "✗", FailedStyle
	default:
		return "○", PendingStyle
	}
}

func (m *Model) executePostInstallOption() tea.Cmd {
	switch m.selectedOption {
	case 0: // Unmount and reboot
		m.exitAction = 1
		return tea.Quit
	case 1: // Enter chroot
		m.exitAction = 2
		return tea.Quit
	case 2: // Exit installer
		m.exitAction = 3
		return tea.Quit
	default:
		return nil
	}
}

// GetExitAction returns the action to take after the UI exits
func (m *Model) GetExitAction() int {
	return m.exitAction
}
