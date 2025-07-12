// Package logger contains loggic for logging.
package logger

import (
	"fmt"
	"os"
)

// Basic golds the config for a basic logger.
type Basic struct {
	verbose bool
}

// NewBasic creates a new basic logger.
func NewBasic(verbose bool) *Basic {
	return &Basic{
		verbose: verbose,
	}
}

// Debug print a debug message if verbose mode is enabled.
func (b Basic) Debug(msg string, args ...any) {
	if b.verbose {
		b.logLine("DEBUG", fmt.Sprintf(msg, args...))
	}
}

// Error prints an info level message.
func (b Basic) Error(msg string, args ...any) {
	b.logLine(" ERROR", fmt.Sprintf(msg, args...))
}

// Fatal prints a fatal level message then exits.
func (b Basic) Fatal(msg string, args ...any) {
	b.logLine("FATAL", fmt.Sprintf(msg, args...))
	os.Exit(1)
}

// Info prints an info level message.
func (b Basic) Info(msg string, args ...any) {
	b.logLine(" INFO", fmt.Sprintf(msg, args...))
}

func (b Basic) logLine(level string, msg string) {
	fmt.Printf("[plex2m3u] %s: %s\n", level, msg)
}
