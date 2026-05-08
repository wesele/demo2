// Package ui handles terminal output, color rendering, and user prompts.
package ui

import (
	"fmt"
	"os"

	"github.com/wesele/demo2/internal/classify"
)

// ANSI color codes.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorBold   = "\033[1m"
)

// IsTTY returns true when stdout is an interactive terminal.
func IsTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// colorEnabled returns true when ANSI color should be used.
func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return IsTTY()
}

// PrintCommand renders the classified command with the appropriate color label.
func PrintCommand(cmd string, level classify.Level) {
	if colorEnabled() {
		var color string
		switch level {
		case classify.Danger:
			color = colorRed
		case classify.Caution:
			color = colorYellow
		default:
			color = colorGreen
		}
		fmt.Printf("%s%s[%s]%s %s%s%s\n",
			color, colorBold, level.String(), colorReset,
			colorBold, cmd, colorReset,
		)
	} else {
		fmt.Printf("[%s] %s\n", level.String(), cmd)
	}
}

// PrintError writes a formatted error message to stderr.
func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
}

// PrintDebugSection prints a labelled debug section header to stderr.
func PrintDebugSection(label string) {
	fmt.Fprintf(os.Stderr, "\n--- DEBUG: %s ---\n", label)
}

// PrintDebug writes a debug line to stderr.
func PrintDebug(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
