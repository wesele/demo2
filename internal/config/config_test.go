package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Set only API key via env; everything else uses defaults.
	t.Setenv("AI_CMD_API_KEY", "test-key")
	t.Setenv("AI_CMD_ENDPOINT", "")
	t.Setenv("AI_CMD_MODEL", "")

	// Use a temp dir so we don't pick up a real config file.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.APIKey != "test-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "test-key")
	}
	if cfg.Endpoint != defaultEndpoint {
		t.Errorf("Endpoint = %q, want %q", cfg.Endpoint, defaultEndpoint)
	}
	if cfg.Model != defaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model, defaultModel)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)

	// Write a config file with specific values.
	cfgDir := filepath.Join(tmp, ".ai-cmd")
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		t.Fatal(err)
	}
	fileContent := `{"endpoint":"https://file-endpoint","api_key":"file-key","model":"file-model"}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(fileContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Env var should override file.
	t.Setenv("AI_CMD_API_KEY", "env-key")
	t.Setenv("AI_CMD_ENDPOINT", "https://env-endpoint")
	t.Setenv("AI_CMD_MODEL", "env-model")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.APIKey != "env-key" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "env-key")
	}
	if cfg.Endpoint != "https://env-endpoint" {
		t.Errorf("Endpoint = %q, want %q", cfg.Endpoint, "https://env-endpoint")
	}
	if cfg.Model != "env-model" {
		t.Errorf("Model = %q, want %q", cfg.Model, "env-model")
	}
}

func TestLoadNoAPIKeyErrors(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("AI_CMD_API_KEY", "")
	t.Setenv("AI_CMD_ENDPOINT", "")
	t.Setenv("AI_CMD_MODEL", "")

	_, err := Load()
	if err == nil {
		t.Error("expected error when no API key is set, got nil")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("AI_CMD_API_KEY", "")
	t.Setenv("AI_CMD_ENDPOINT", "")
	t.Setenv("AI_CMD_MODEL", "")

	want := &Config{
		Endpoint: "https://custom.example.com/v1",
		APIKey:   "sk-abc123",
		Model:    "gpt-4-turbo",
	}
	if err := Save(want); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify file permissions on non-Windows only (Windows does not enforce Unix mode bits).
	path, _ := ConfigPath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat config file: %v", err)
	}
	if runtime.GOOS != "windows" {
		if info.Mode().Perm()&0077 != 0 {
			t.Errorf("config file permissions %o are too permissive", info.Mode().Perm())
		}
	}

	t.Setenv("AI_CMD_API_KEY", want.APIKey)
	got, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}
	if got.Endpoint != want.Endpoint {
		t.Errorf("Endpoint = %q, want %q", got.Endpoint, want.Endpoint)
	}
	if got.Model != want.Model {
		t.Errorf("Model = %q, want %q", got.Model, want.Model)
	}
}

func TestMaskKey(t *testing.T) {
	cases := []struct {
		key  string
		want string
	}{
		{"sk-abc123456", "sk-abc1****"},
		{"short", "****"},
		{"sk-1234567", "sk-1234****"},
	}
	for _, c := range cases {
		got := MaskKey(c.key)
		if got != c.want {
			t.Errorf("MaskKey(%q) = %q, want %q", c.key, got, c.want)
		}
	}
}
