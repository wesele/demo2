package shell

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Execute runs cmd in the user's current shell, streaming stdout and stderr
// to the terminal. Returns the child process exit code wrapped in an error
// if non-zero, or nil on success.
func Execute(cmd string) error {
	var command *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		command = exec.Command("powershell.exe", "-NoProfile", "-Command", cmd)
	default:
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		command = exec.Command(shell, "-c", cmd)
	}

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		// Propagate exit code; callers inspect *exec.ExitError for the code.
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

// ExitCode extracts the exit code from an error returned by Execute.
// Returns 1 if the error is not an *exec.ExitError.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if ok := isExitError(err, &exitErr); ok {
		return exitErr.ExitCode()
	}
	return 1
}

func isExitError(err error, target **exec.ExitError) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*exec.ExitError); ok {
		*target = e
		return true
	}
	// Unwrap one level for wrapped errors.
	type unwrapper interface{ Unwrap() error }
	if u, ok := err.(unwrapper); ok {
		return isExitError(u.Unwrap(), target)
	}
	return false
}
