package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

// UICallback is called when UI is active to send logs to it
var UICallback func(string)

// Logger provides structured logging with colors
type Logger struct {
	verbose bool
	logFile *os.File
}

// New creates a new logger instance
func New(verbose bool) *Logger {
	// Create or append to debug.log
	logFile, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to open debug.log: %v\n", err)
	}

	return &Logger{
		verbose: verbose,
		logFile: logFile,
	}
}

// Close closes the log file
func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}

// log writes to both console and debug.log
func (l *Logger) log(level, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	
	// Write to debug.log
	if l.logFile != nil {
		fmt.Fprintf(l.logFile, "[%s] [%s] %s\n", timestamp, level, message)
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log("INFO", format, args...)
	message := fmt.Sprintf(format, args...)
	
	if UICallback != nil {
		UICallback(message)
	} else {
		color.Blue("[INFO] ")
		fmt.Println(message)
	}
}

// Success logs a success message
func (l *Logger) Success(format string, args ...interface{}) {
	l.log("SUCCESS", format, args...)
	message := fmt.Sprintf(format, args...)
	
	if UICallback != nil {
		UICallback("✓ " + message)
	} else {
		color.Green("[SUCCESS] ")
		fmt.Println(message)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log("WARN", format, args...)
	message := fmt.Sprintf(format, args...)
	
	if UICallback != nil {
		UICallback("⚠ " + message)
	} else {
		color.Yellow("[WARN] ")
		fmt.Println(message)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
	message := fmt.Sprintf(format, args...)
	
	if UICallback != nil {
		UICallback("✗ " + message)
	} else {
		color.Red("[ERROR] ")
		fmt.Println(message)
	}
}

// Step logs a step message
func (l *Logger) Step(format string, args ...interface{}) {
	l.log("STEP", format, args...)
	message := fmt.Sprintf(format, args...)
	
	if UICallback != nil {
		UICallback("▶ " + message)
	} else {
		color.Magenta("[STEP] ")
		fmt.Println(message)
	}
}

// Stage logs a stage message (sub-step)
func (l *Logger) Stage(format string, args ...interface{}) {
	l.log("STAGE", format, args...)
	message := fmt.Sprintf(format, args...)
	
	if UICallback != nil {
		UICallback("  ▸ " + message)
	} else {
		color.Cyan("[STAGE] ")
		fmt.Println(message)
	}
}

// Debug logs a debug message (only if verbose)
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.verbose {
		l.log("DEBUG", format, args...)
		color.White("[DEBUG] ")
		fmt.Printf(format+"\n", args...)
	}
}

// Command logs a command that will be executed
func (l *Logger) Command(cmd string) {
	l.log("CMD", cmd)
	if UICallback != nil {
		UICallback("$ " + cmd)
	} else if l.verbose {
		color.Cyan("$ ")
		fmt.Println(cmd)
	}
}
