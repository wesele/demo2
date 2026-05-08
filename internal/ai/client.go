package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wesele/demo2/internal/config"
)

const (
	httpTimeout = 30 * time.Second
	maxTokens   = 256
	temperature = 0.0
)

// chatRequest is the OpenAI-compatible chat completion request body.
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse is a minimal subset of the OpenAI chat completion response.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// DebugInfo carries raw request/response data for debug mode.
type DebugInfo struct {
	SystemPrompt string
	UserPrompt   string
	RawResponse  string
}

// Translate sends query to the AI API and returns the generated shell command.
// If debug is non-nil, it is populated with the raw request and response data.
func Translate(query string, cfg *config.Config, shellOS, shellName string, debug *DebugInfo) (string, error) {
	sysPrompt := buildSystemPrompt(shellOS, shellName)

	if debug != nil {
		debug.SystemPrompt = sysPrompt
		debug.UserPrompt = query
	}

	reqBody := chatRequest{
		Model: cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: query},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("encoding request: %w", err)
	}

	endpoint := strings.TrimRight(cfg.Endpoint, "/") + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		// Detect timeout specifically.
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			return "", fmt.Errorf("request timed out — check your connection")
		}
		return "", fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if debug != nil {
		debug.RawResponse = string(rawBody)
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return "", fmt.Errorf("invalid API key — run 'ai -c' to reconfigure")
	case http.StatusTooManyRequests:
		return "", fmt.Errorf("API rate limit exceeded — please try again later")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(rawBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(rawBody, &chatResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("AI returned no choices")
	}

	cmd := cleanCommand(chatResp.Choices[0].Message.Content)

	if strings.HasPrefix(strings.TrimSpace(cmd), "ERROR:") {
		reason := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(cmd), "ERROR:"))
		return "", fmt.Errorf("AI could not generate a command: %s", reason)
	}

	if cmd == "" {
		return "", fmt.Errorf("AI returned an empty command")
	}

	return cmd, nil
}

// cleanCommand strips markdown code fences and extraneous whitespace.
func cleanCommand(s string) string {
	s = strings.TrimSpace(s)
	// Strip triple-backtick fences (```bash\n...\n``` or ```\n...\n```).
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		var inner []string
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if strings.HasPrefix(trimmed, "```") {
				continue
			}
			inner = append(inner, l)
		}
		s = strings.TrimSpace(strings.Join(inner, "\n"))
	}
	// Strip single backticks.
	s = strings.Trim(s, "`")
	return strings.TrimSpace(s)
}
