package logger

import (
	"fmt"
	"time"
)

// ANSI color codes for terminal output
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Bold    = "\033[1m"
	BgRed   = "\033[41m"
	BgGreen = "\033[42m"
	BgBlue  = "\033[44m"
)

// Logger provides colored, prefixed logging for demo output
type Logger struct {
	Prefix string
	Color  string
}

// NewMasterLogger creates a logger for a Cluster Manager node
func NewMasterLogger(nodeID int) *Logger {
	colors := []string{Blue, Magenta, Cyan}
	color := colors[nodeID%len(colors)]
	return &Logger{
		Prefix: fmt.Sprintf("MASTER-%d", nodeID),
		Color:  color,
	}
}

// NewExecutorLogger creates a logger for an Executor node
func NewExecutorLogger(executorID int) *Logger {
	colors := []string{Green, Yellow, Cyan}
	color := colors[executorID%len(colors)]
	return &Logger{
		Prefix: fmt.Sprintf("EXECUTOR-%d", executorID),
		Color:  color,
	}
}

// NewTokenLogger creates a logger for the Token Manager
func NewTokenLogger() *Logger {
	return &Logger{
		Prefix: "TOKEN-MGR",
		Color:  Yellow,
	}
}

// NewDeadlockLogger creates a logger for the Deadlock Detector
func NewDeadlockLogger() *Logger {
	return &Logger{
		Prefix: "DEADLOCK",
		Color:  Red,
	}
}

// NewSystemLogger creates a logger for system-level messages
func NewSystemLogger() *Logger {
	return &Logger{
		Prefix: "SYSTEM",
		Color:  White,
	}
}

func (l *Logger) timestamp() string {
	return time.Now().Format("15:04:05")
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[%s] [%s%s%s] %s%s\n", l.Color, l.timestamp(), Bold, l.Prefix, Reset+l.Color, msg, Reset)
}

// Success logs a success message
func (l *Logger) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[%s] [%s%s%s] ✓ %s%s\n", Green, l.timestamp(), Bold, l.Prefix, Reset+Green, msg, Reset)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[%s] [%s%s%s] ⚠ %s%s\n", Yellow, l.timestamp(), Bold, l.Prefix, Reset+Yellow, msg, Reset)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[%s] [%s%s%s] ✗ %s%s\n", Red, l.timestamp(), Bold, l.Prefix, Reset+Red, msg, Reset)
}

// Critical logs a critical event with background highlighting
func (l *Logger) Critical(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s[%s] [%s] ★ %s%s\n", BgRed, White, l.timestamp(), l.Prefix, msg, Reset)
}

// Election logs an election-related event
func (l *Logger) Election(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s[%s] [%s%s%s] ⚡ %s%s\n", Magenta, l.timestamp(), Bold, l.Prefix, Reset+Magenta, msg, Reset)
}

// Banner prints a large banner for demo section headers
func Banner(title string) {
	line := "═══════════════════════════════════════════════════════════════"
	fmt.Printf("\n%s%s%s\n", Bold+Cyan, line, Reset)
	fmt.Printf("%s%s  %s%s\n", Bold+Cyan, "║", title, Reset)
	fmt.Printf("%s%s%s\n\n", Bold+Cyan, line, Reset)
}
