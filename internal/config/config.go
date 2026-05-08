// Package config handles loading, saving, and resolving ai-cmd configuration.
// Resolution order: environment variables > ~/.ai-cmd/config.json > built-in defaults.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	defaultEndpoint = "https://api.openai.com/v1"
	defaultModel    = "gpt-4o"
	configDirName   = ".ai-cmd"
	configFileName  = "config.json"
)

// Config holds all runtime configuration for ai-cmd.
type Config struct {
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
}

// ConfigDir returns the path to the ~/.ai-cmd directory.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, configDirName), nil
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// Load returns a Config resolved from env vars, the config file, and built-in defaults.
// Returns an error only when the config is incomplete (no API key available anywhere).
func Load() (*Config, error) {
	cfg := &Config{
		Endpoint: defaultEndpoint,
		Model:    defaultModel,
	}

	// Layer 1: config file
	if err := loadFile(cfg); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Layer 2: environment variable overrides
	if v := os.Getenv("AI_CMD_ENDPOINT"); v != "" {
		cfg.Endpoint = v
	}
	if v := os.Getenv("AI_CMD_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("AI_CMD_MODEL"); v != "" {
		cfg.Model = v
	}

	if cfg.APIKey == "" {
		return nil, errors.New("no API key configured — run 'ai -c' to get started, or set AI_CMD_API_KEY")
	}

	return cfg, nil
}

// loadFile reads the config file into cfg, leaving fields unchanged if the file
// does not exist or a field is absent in the JSON.
func loadFile(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cfg)
}

// Save persists cfg to ~/.ai-cmd/config.json with 0600 permissions.
func Save(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	// Create directory with restricted permissions.
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}

	path := filepath.Join(dir, configFileName)
	perm := os.FileMode(0600)
	// On Windows 0600 has no enforced effect, but we set it for correctness.
	_ = runtime.GOOS
	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// MaskKey returns a redacted version of the API key safe for display.
func MaskKey(key string) string {
	if len(key) <= 7 {
		return "****"
	}
	return key[:7] + "****"
}
