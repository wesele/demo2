package shell

import (
	"testing"
)

func TestUnixShellName(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"/bin/bash", "bash"},
		{"/usr/bin/zsh", "zsh"},
		{"/usr/local/bin/fish", "fish"},
		{"/bin/sh", "sh"},
		{"/usr/bin/pwsh", "powershell"},
		{"/usr/bin/powershell", "powershell"},
		{"/bin/unknown", "sh"},
	}
	for _, c := range cases {
		got := unixShellName(c.path)
		if got != c.want {
			t.Errorf("unixShellName(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

func TestDetectReturnsSomething(t *testing.T) {
	info := Detect()
	if info.OS == "" {
		t.Error("Detect().OS is empty")
	}
	if info.Shell == "" {
		t.Error("Detect().Shell is empty")
	}
}
