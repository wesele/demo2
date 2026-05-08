// Package shell handles OS/shell detection and command execution.
package shell

import (
	"os"
	"runtime"
	"strings"
)

// Info holds detected OS and shell information.
type Info struct {
	OS    string // "windows", "linux", "darwin"
	Shell string // "powershell", "cmd", "bash", "zsh", "sh"
}

// Detect returns the current OS and active shell.
func Detect() Info {
	goos := runtime.GOOS

	var shell string
	switch goos {
	case "windows":
		shell = detectWindowsShell()
	default:
		shell = detectUnixShell()
	}

	return Info{OS: goos, Shell: shell}
}

func detectWindowsShell() string {
	// Check if running inside a known Unix-like shell on Windows.
	if sh := os.Getenv("SHELL"); sh != "" {
		return unixShellName(sh)
	}
	// PSModulePath is set by PowerShell.
	if os.Getenv("PSModulePath") != "" {
		return "powershell"
	}
	// ComSpec points to cmd.exe by default.
	if cs := os.Getenv("ComSpec"); strings.Contains(strings.ToLower(cs), "cmd.exe") {
		return "cmd"
	}
	return "powershell" // safe default on Windows
}

func detectUnixShell() string {
	if sh := os.Getenv("SHELL"); sh != "" {
		return unixShellName(sh)
	}
	return "sh"
}

func unixShellName(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.Contains(lower, "zsh"):
		return "zsh"
	case strings.Contains(lower, "bash"):
		return "bash"
	case strings.Contains(lower, "fish"):
		return "fish"
	case strings.Contains(lower, "powershell") || strings.Contains(lower, "pwsh"):
		return "powershell"
	default:
		return "sh"
	}
}
