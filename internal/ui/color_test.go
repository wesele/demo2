package ui

import (
	"testing"
)

func TestColorEnabledNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if colorEnabled() {
		t.Error("colorEnabled() should return false when NO_COLOR is set")
	}
}

func TestColorEnabledNoColorEmpty(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	// In test environment stdout is not a TTY, so colorEnabled should be false.
	got := colorEnabled()
	// We just verify it doesn't panic; the actual value depends on the test runner TTY.
	_ = got
}
