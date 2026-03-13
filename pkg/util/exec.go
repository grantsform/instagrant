package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Executor handles command execution with logging and UI integration
type Executor struct {
	dryRun    bool
	logFunc   func(string)
	cmdFunc   func(string)
	uiWriter  *UIWriter
}

// UIWriter writes command output to UI when active
type UIWriter struct {
	logFunc     func(string)
	mu          sync.Mutex
	lastUpdate  time.Time
	buffer      []string
	minInterval time.Duration
}

func (w *UIWriter) Write(p []byte) (n int, err error) {
	if w.logFunc != nil {
		w.mu.Lock()
		defer w.mu.Unlock()
		
		for _, line := range strings.Split(string(p), "\n") {
			if line = strings.TrimSpace(line); line != "" {
				w.buffer = append(w.buffer, line)
			}
		}
		
		// Throttle updates to prevent UI overflow
		if time.Since(w.lastUpdate) >= w.minInterval && len(w.buffer) > 0 {
			// Send only the last few lines if buffer is large
			start := 0
			if len(w.buffer) > 5 {
				start = len(w.buffer) - 5
			}
			for _, line := range w.buffer[start:] {
				w.logFunc(line)
			}
			w.buffer = nil
			w.lastUpdate = time.Now()
		}
	}
	return len(p), nil
}

// Flush sends any remaining buffered lines
func (w *UIWriter) Flush() {
	if w.logFunc != nil {
		w.mu.Lock()
		defer w.mu.Unlock()
		
		for _, line := range w.buffer {
			w.logFunc(line)
		}
		w.buffer = nil
	}
}

// NewExecutor creates a new command executor
func NewExecutor(dryRun bool, logFunc, cmdFunc func(string)) *Executor {
	var writer *UIWriter
	if logFunc != nil {
		writer = &UIWriter{
			logFunc:     logFunc,
			minInterval: 100 * time.Millisecond,
		}
	}
	return &Executor{
		dryRun:   dryRun,
		logFunc:  logFunc,
		cmdFunc:  cmdFunc,
		uiWriter: writer,
	}
}

// Run executes a command on the host system
func (e *Executor) Run(name string, args ...string) error {
	cmdStr := name + " " + strings.Join(args, " ")
	if e.cmdFunc != nil {
		e.cmdFunc(cmdStr)
	}
	
	if e.dryRun {
		return nil
	}
	
	cmd := exec.Command(name, args...)
	if e.uiWriter != nil {
		cmd.Stdout = e.uiWriter
		cmd.Stderr = e.uiWriter
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	err := cmd.Run()
	if e.uiWriter != nil {
		e.uiWriter.Flush()
	}
	return err
}

// RunSh executes a shell command
func (e *Executor) RunSh(command string) error {
	if e.cmdFunc != nil {
		e.cmdFunc(command)
	}
	
	if e.dryRun {
		return nil
	}
	
	cmd := exec.Command("bash", "-c", command)
	if e.uiWriter != nil {
		cmd.Stdout = e.uiWriter
		cmd.Stderr = e.uiWriter
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	err := cmd.Run()
	if e.uiWriter != nil {
		e.uiWriter.Flush()
	}
	return err
}

// Chroot executes a command inside a chroot environment
func (e *Executor) Chroot(targetDir, command string) error {
	cmdStr := fmt.Sprintf("(chroot) %s", command)
	if e.cmdFunc != nil {
		e.cmdFunc(cmdStr)
	}
	
	if e.dryRun {
		return nil
	}
	
	cmd := exec.Command("arch-chroot", targetDir, "bash", "-c", command)
	if e.uiWriter != nil {
		cmd.Stdout = e.uiWriter
		cmd.Stderr = e.uiWriter
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	err := cmd.Run()
	if e.uiWriter != nil {
		e.uiWriter.Flush()
	}
	return err
}

// ChrootOutput executes a command in chroot and returns its output
func (e *Executor) ChrootOutput(targetDir, command string) (string, error) {
	if e.cmdFunc != nil {
		e.cmdFunc(fmt.Sprintf("(chroot) %s", command))
	}
	
	if e.dryRun {
		return "", nil
	}
	
	cmd := exec.Command("arch-chroot", targetDir, "bash", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ChrootSilent executes a command in chroot without logging it (for sensitive operations)
func (e *Executor) ChrootSilent(targetDir, command string) error {
	// Don't log sensitive commands - skip cmdFunc call
	
	if e.dryRun {
		return nil
	}
	
	cmd := exec.Command("arch-chroot", targetDir, "bash", "-c", command)
	if e.uiWriter != nil {
		cmd.Stdout = e.uiWriter
		cmd.Stderr = e.uiWriter
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	err := cmd.Run()
	if e.uiWriter != nil {
		e.uiWriter.Flush()
	}
	return err
}
