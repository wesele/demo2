package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wesele/demo2/internal/config"
)

func mockServer(t *testing.T, status int, body any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(body); err != nil {
			t.Fatalf("mock server encode: %v", err)
		}
	}))
}

func TestTranslateSuccess(t *testing.T) {
	srv := mockServer(t, http.StatusOK, map[string]any{
		"choices": []map[string]any{
			{"message": map[string]any{"content": "ls -la"}},
		},
	})
	defer srv.Close()

	cfg := &config.Config{Endpoint: srv.URL, APIKey: "test", Model: "gpt-4o"}
	cmd, err := Translate("list files", cfg, "linux", "bash", nil)
	if err != nil {
		t.Fatalf("Translate() error: %v", err)
	}
	if cmd != "ls -la" {
		t.Errorf("Translate() = %q, want %q", cmd, "ls -la")
	}
}

func TestTranslateStripsMarkdown(t *testing.T) {
	srv := mockServer(t, http.StatusOK, map[string]any{
		"choices": []map[string]any{
			{"message": map[string]any{"content": "```bash\nls -la\n```"}},
		},
	})
	defer srv.Close()

	cfg := &config.Config{Endpoint: srv.URL, APIKey: "test", Model: "gpt-4o"}
	cmd, err := Translate("list files", cfg, "linux", "bash", nil)
	if err != nil {
		t.Fatalf("Translate() error: %v", err)
	}
	if cmd != "ls -la" {
		t.Errorf("Translate() = %q, want %q", cmd, "ls -la")
	}
}

func TestTranslateUnauthorized(t *testing.T) {
	srv := mockServer(t, http.StatusUnauthorized, map[string]any{"error": map[string]any{"message": "invalid key"}})
	defer srv.Close()

	cfg := &config.Config{Endpoint: srv.URL, APIKey: "bad", Model: "gpt-4o"}
	_, err := Translate("list files", cfg, "linux", "bash", nil)
	if err == nil {
		t.Error("expected error for 401, got nil")
	}
}

func TestTranslateRateLimit(t *testing.T) {
	srv := mockServer(t, http.StatusTooManyRequests, map[string]any{})
	defer srv.Close()

	cfg := &config.Config{Endpoint: srv.URL, APIKey: "key", Model: "gpt-4o"}
	_, err := Translate("list files", cfg, "linux", "bash", nil)
	if err == nil {
		t.Error("expected error for 429, got nil")
	}
}

func TestTranslateAIError(t *testing.T) {
	srv := mockServer(t, http.StatusOK, map[string]any{
		"choices": []map[string]any{
			{"message": map[string]any{"content": "ERROR: query too vague"}},
		},
	})
	defer srv.Close()

	cfg := &config.Config{Endpoint: srv.URL, APIKey: "key", Model: "gpt-4o"}
	_, err := Translate("do the thing", cfg, "linux", "bash", nil)
	if err == nil {
		t.Error("expected error for AI ERROR: response, got nil")
	}
}

func TestCleanCommand(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"ls -la", "ls -la"},
		{"```\nls -la\n```", "ls -la"},
		{"```bash\nls -la\n```", "ls -la"},
		{"`ls -la`", "ls -la"},
		{"  ls -la  ", "ls -la"},
	}
	for _, c := range cases {
		got := cleanCommand(c.input)
		if got != c.want {
			t.Errorf("cleanCommand(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestDebugInfoPopulated(t *testing.T) {
	srv := mockServer(t, http.StatusOK, map[string]any{
		"choices": []map[string]any{
			{"message": map[string]any{"content": "ls -la"}},
		},
	})
	defer srv.Close()

	cfg := &config.Config{Endpoint: srv.URL, APIKey: "test", Model: "gpt-4o"}
	var dbg DebugInfo
	_, err := Translate("list files", cfg, "linux", "bash", &dbg)
	if err != nil {
		t.Fatalf("Translate() error: %v", err)
	}
	if dbg.SystemPrompt == "" {
		t.Error("DebugInfo.SystemPrompt should be populated")
	}
	if dbg.UserPrompt == "" {
		t.Error("DebugInfo.UserPrompt should be populated")
	}
	if dbg.RawResponse == "" {
		t.Error("DebugInfo.RawResponse should be populated")
	}
}
