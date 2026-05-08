// ai-cmd: AI-powered command line helper.
// Translates natural language to shell commands with danger classification and confirmation.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/wesele/demo2/internal/ai"
	"github.com/wesele/demo2/internal/classify"
	"github.com/wesele/demo2/internal/config"
	"github.com/wesele/demo2/internal/shell"
	"github.com/wesele/demo2/internal/ui"
)

const version = "1.0.0"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("ai", flag.ContinueOnError)
	flagConfigure := fs.Bool("c", false, "Interactive configuration wizard")
	flagDebug := fs.Bool("d", false, "Debug mode: print raw AI request/response details")
	flagHelp := fs.Bool("h", false, "Show this help message")
	flagVersion := fs.Bool("v", false, "Print version and exit")

	fs.Usage = printHelp

	if err := fs.Parse(args); err != nil {
		printHelp()
		return 1
	}

	if *flagHelp {
		printHelp()
		return 0
	}

	if *flagVersion {
		fmt.Printf("ai-cmd version %s\n", version)
		return 0
	}

	if *flagConfigure {
		if err := config.RunWizard(); err != nil {
			ui.PrintError(err.Error())
			return 1
		}
		return 0
	}

	// Remaining args form the natural-language query.
	query := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if query == "" {
		ui.PrintError("no query provided")
		printHelp()
		return 1
	}

	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		ui.PrintError(err.Error())
		return 1
	}

	// Detect OS and shell.
	shellInfo := shell.Detect()

	// Optionally collect debug info.
	var debugInfo *ai.DebugInfo
	if *flagDebug {
		debugInfo = &ai.DebugInfo{}
	}

	// Call the AI.
	cmd, err := ai.Translate(query, cfg, shellInfo.OS, shellInfo.Shell, debugInfo)

	// Print debug output before any errors so the user can see what happened.
	if *flagDebug && debugInfo != nil {
		ui.PrintDebugSection("REQUEST")
		ui.PrintDebug("System prompt:\n%s", debugInfo.SystemPrompt)
		ui.PrintDebug("\nUser query: %s", debugInfo.UserPrompt)
		ui.PrintDebug("Endpoint:   %s/chat/completions", cfg.Endpoint)
		ui.PrintDebug("API Key:    %s", config.MaskKey(cfg.APIKey))
		ui.PrintDebug("Model:      %s", cfg.Model)
		ui.PrintDebugSection("RESPONSE")
		ui.PrintDebug("%s", debugInfo.RawResponse)
		ui.PrintDebugSection("END DEBUG")
	}

	if err != nil {
		ui.PrintError(err.Error())
		return 1
	}

	// Classify and display the command.
	level := classify.Classify(cmd)
	ui.PrintCommand(cmd, level)

	// Ask for confirmation.
	if !ui.Confirm() {
		fmt.Println("Aborted.")
		return 0
	}

	// Execute the command.
	if err := shell.Execute(cmd); err != nil {
		return shell.ExitCode(err)
	}
	return 0
}

func printHelp() {
	fmt.Print(`Usage:
  ai [flags] <natural language query>

Flags:
  -c    Interactive configuration wizard
  -d    Debug mode: print raw AI request/response details
  -h    Show this help message
  -v    Print version and exit

Configuration (in order of precedence):
  Environment variables: AI_CMD_ENDPOINT, AI_CMD_API_KEY, AI_CMD_MODEL
  Config file:           ~/.ai-cmd/config.json
  Run 'ai -c' to configure interactively.

Danger levels:
  [SAFE]    Read-only commands (shown in green)
  [CAUTION] Commands that create or modify (shown in yellow)
  [DANGER]  Destructive or irreversible commands (shown in red)

Examples:
  ai find all log files larger than 100MB
  ai kill the process on port 8080
  ai show disk usage sorted by size
  ai -d list all running docker containers
  ai -c

`)
}
