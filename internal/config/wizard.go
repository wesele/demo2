package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// RunWizard runs the interactive configuration wizard (-c flag).
// It loads any existing config, prompts the user to keep or overwrite each field,
// then saves the result.
func RunWizard() error {
	existing := &Config{
		Endpoint: defaultEndpoint,
		Model:    defaultModel,
	}
	// Best-effort load of existing config; ignore errors.
	_ = loadFile(existing)

	fmt.Println("ai-cmd configuration wizard")
	fmt.Println("Press Enter to keep the current value shown in brackets.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Prompt endpoint.
	endpoint, err := promptValidated(reader,
		fmt.Sprintf("API Endpoint [%s]: ", existing.Endpoint),
		existing.Endpoint,
		validateURL,
	)
	if err != nil {
		return err
	}

	// Prompt API key (masked display).
	maskedCurrent := ""
	if existing.APIKey != "" {
		maskedCurrent = MaskKey(existing.APIKey)
	}
	apiKey, err := promptValidated(reader,
		fmt.Sprintf("API Key [%s]: ", maskedCurrent),
		existing.APIKey,
		validateNonEmpty,
	)
	if err != nil {
		return err
	}

	// Prompt model.
	model, err := promptValidated(reader,
		fmt.Sprintf("Model [%s]: ", existing.Model),
		existing.Model,
		validateNonEmpty,
	)
	if err != nil {
		return err
	}

	cfg := &Config{
		Endpoint: endpoint,
		APIKey:   apiKey,
		Model:    model,
	}

	if err := Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	path, _ := ConfigPath()
	fmt.Printf("\nConfiguration saved to %s\n", path)
	return nil
}

// promptValidated prints prompt, reads a line from r, returns defaultVal if the
// line is empty, and re-prompts on validation failure.
func promptValidated(r *bufio.Reader, prompt, defaultVal string, validate func(string) error) (string, error) {
	for {
		fmt.Print(prompt)
		line, err := r.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("reading input: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			line = defaultVal
		}
		if err := validate(line); err != nil {
			fmt.Fprintf(os.Stderr, "  Invalid input: %v\n", err)
			continue
		}
		return line, nil
	}
}

func validateURL(s string) error {
	u, err := url.ParseRequestURI(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("%q is not a valid HTTP/HTTPS URL", s)
	}
	return nil
}

func validateNonEmpty(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}
