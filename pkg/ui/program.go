package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	program  *tea.Program
	model    *Model
	active   bool
	mu       sync.Mutex
	doneChan   chan struct{}
	exitAction int
)

// Start initializes and starts the TUI
func Start(phases []PhaseInfo) error {
	mu.Lock()
	model = NewModel(phases)
	active = true
	doneChan = make(chan struct{})
	exitAction = 0
	mu.Unlock()
	
	program = tea.NewProgram(model, tea.WithAltScreen())
	
	// Run in background
	go func() {
		if _, err := program.Run(); err != nil {
			// Handle error silently or log it
		}
		mu.Lock()
		active = false
		exitAction = model.GetExitAction()
		mu.Unlock()
		close(doneChan)
	}()
	
	return nil
}

// Stop stops the TUI
func Stop() {
	if program != nil {
		program.Quit()
	}
	mu.Lock()
	active = false
	mu.Unlock()
}

// Wait blocks until the TUI exits (user selects an option from post-install menu)
func Wait() {
	if doneChan != nil {
		<-doneChan
	}
}

// UpdatePhase updates a phase status in the TUI
func UpdatePhase(index int, status PhaseStatus) {
	mu.Lock()
	defer mu.Unlock()
	if program != nil && model != nil && active {
		program.Send(PhaseUpdateMsg{Index: index, Status: status})
	}
}

// clearTerminal clears the terminal screen
func clearTerminal() {
	// Try using the clear command first
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		// Fallback to ANSI escape sequences
		fmt.Print("\033[H\033[2J")
	}
}

// AddLog adds a log line to the TUI
func AddLog(text string) {
	mu.Lock()
	defer mu.Unlock()
	if program != nil && model != nil && active {
		// Clean up the text
		text = strings.TrimSpace(text)
		if text != "" {
			program.Send(LogMsg{Text: text})
		}
	}
}

// TUIWriter is an io.Writer that sends output only to the TUI
type TUIWriter struct{}

func (w *TUIWriter) Write(p []byte) (n int, err error) {
	text := string(p)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			AddLog(line)
		}
	}
	// Return success but don't write to actual stdout
	return len(p), nil
}

// IsActive returns whether the TUI is currently active
func IsActive() bool {
	mu.Lock()
	defer mu.Unlock()
	return active
}

// GetExitAction returns the action selected by user in post-install menu
// 0 = none, 1 = reboot, 2 = chroot, 3 = exit
func GetExitAction() int {
	mu.Lock()
	defer mu.Unlock()
	return exitAction
}
