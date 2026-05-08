# Technical Requirements & Specification
## ai-cmd — AI-Powered Command Line Helper

**Version:** 1.0  
**Date:** 2026-05-08  
**Status:** Draft

---

## 1. Overview

`ai-cmd` is a cross-platform command-line tool written in **Go** that translates natural-language descriptions into shell commands using an AI backend. It targets developers and system administrators who work across Windows (PowerShell) and Linux/macOS (Bash/Zsh) environments.

---

## 2. Goals & Non-Goals

### 2.1 Goals
- Accept a natural-language query and return the correct shell command for the detected OS/shell.
- Classify command danger level and present it visually before execution.
- Require explicit user confirmation before any command is run.
- Support any OpenAI-compatible API endpoint.
- Provide an interactive setup wizard and environment-variable fallback for configuration.
- Provide help (`-h`) and debug (`-d`) modes.

### 2.2 Non-Goals
- Multi-turn conversation / session memory (out of scope for v1.0).
- GUI or web interface.
- Plugin or extension system.
- Built-in rollback / undo of executed commands.

---

## 3. User Stories

| ID | As a… | I want to… | So that… |
|----|--------|-----------|----------|
| US-1 | CLI user | Type `ai <natural language>` | I get the right shell command without Googling |
| US-2 | CLI user | See the command color-coded by danger level | I am warned before something destructive runs |
| US-3 | CLI user | Press Enter to confirm before execution | I never accidentally destroy my system |
| US-4 | New user | Run `ai -c` to configure the tool interactively | I don't have to edit config files manually |
| US-5 | Power user | Set config via environment variables | I can configure the tool in CI or scripted environments |
| US-6 | Any user | Run `ai -h` | I can see usage instructions quickly |
| US-7 | Debugging user | Run `ai -d <query>` | I can inspect the raw AI response and any errors |

---

## 4. Functional Requirements

### 4.1 Command Translation (FR-1)

- **FR-1.1** The tool shall accept one or more free-text words after the binary name as the natural-language query (e.g., `ai find all log files larger than 100MB`).
- **FR-1.2** The tool shall detect the host operating system (Windows, Linux, macOS) and active shell (PowerShell, CMD, Bash, Zsh) and include this context in the AI prompt.
- **FR-1.3** The AI shall return exactly one shell command. The tool shall strip any markdown fences or extra prose from the response before displaying it.
- **FR-1.4** Each invocation is independent (single-shot); no history is sent to the AI.

### 4.2 Danger Classification & Display (FR-2)

The tool shall classify the returned command into one of three danger levels and render it accordingly in the terminal:

| Level | Color | Label | Example commands |
|-------|-------|-------|-----------------|
| SAFE | Green | `[SAFE]` | `ls`, `cat`, `echo`, `pwd`, `Get-ChildItem` |
| CAUTION | Yellow | `[CAUTION]` | `mkdir`, `cp`, `mv`, `touch`, `Set-Content` |
| DANGER | Red | `[DANGER]` | `rm -rf`, `del /f /s /q`, `kill`, `format`, `dd` |

- **FR-2.1** Classification shall be performed client-side using a keyword/pattern list (not a second AI call) to keep latency low.
- **FR-2.2** Color output shall be suppressed when stdout is not a TTY (e.g., when piped).

### 4.3 Confirmation Flow (FR-3)

- **FR-3.1** After displaying the classified command, the tool shall print a confirmation prompt:
  ```
  Run this command? [Enter to confirm / Ctrl+C to cancel]:
  ```
- **FR-3.2** The command shall **only** execute after the user presses **Enter**.
- **FR-3.3** Pressing **Ctrl+C** or **Ctrl+D** shall abort without executing anything.
- **FR-3.4** The tool shall execute the command in the **current shell's environment** so that environment variables, aliases, and working directory are preserved.
  - On Linux/macOS: spawn via `$SHELL -c "<command>"`.
  - On Windows: spawn via `powershell.exe -Command "<command>"`.
- **FR-3.5** The tool shall stream stdout and stderr from the executed command directly to the terminal.
- **FR-3.6** The exit code of the child process shall be propagated as the exit code of `ai-cmd`.

### 4.4 Configuration (FR-4)

#### 4.4.1 Interactive Setup (`ai -c`)

- **FR-4.1** Running `ai -c` shall start an interactive wizard that prompts for:
  1. **API Endpoint** — default: `https://api.openai.com/v1`
  2. **API Key**
  3. **Model name** — default: `gpt-4o`
- **FR-4.2** Input shall be validated:
  - Endpoint must be a valid URL.
  - API Key must be a non-empty string.
  - Model must be a non-empty string.
- **FR-4.3** The wizard shall save the values to `~/.ai-cmd/config.json` with `0600` file permissions (owner read/write only).
- **FR-4.4** If the file already exists, the wizard shall display the current values and allow the user to overwrite or keep each one.

#### 4.4.2 Config File Schema

```json
{
  "endpoint": "https://api.openai.com/v1",
  "api_key":  "sk-...",
  "model":    "gpt-4o"
}
```

#### 4.4.3 Environment Variable Override

- **FR-4.5** The following environment variables shall override config file values at runtime:

| Variable | Overrides |
|----------|-----------|
| `AI_CMD_ENDPOINT` | `endpoint` |
| `AI_CMD_API_KEY` | `api_key` |
| `AI_CMD_MODEL` | `model` |

- **FR-4.6** Environment variables take precedence over the config file. The config file takes precedence over built-in defaults.

#### 4.4.4 Config Resolution Order

```
Environment Variables  >  ~/.ai-cmd/config.json  >  Built-in Defaults
```

### 4.5 Help (`ai -h`) (FR-5)

- **FR-5.1** `ai -h` (or `ai --help`) shall print a usage summary to stdout including:
  - Synopsis / usage line.
  - Description of all flags.
  - At least two usage examples.

Example output:
```
Usage:
  ai [flags] <natural language query>

Flags:
  -c    Interactive configuration wizard
  -d    Debug mode: print raw AI response and request details
  -h    Show this help message

Examples:
  ai find all log files larger than 100MB
  ai kill the process on port 8080
  ai -d list all running docker containers
```

### 4.6 Debug Mode (`ai -d`) (FR-6)

- **FR-6.1** When `-d` is passed, the tool shall print the following to stderr before the normal flow:
  1. The full prompt sent to the AI (system + user message).
  2. The raw HTTP response body from the AI API.
  3. Any parsing errors encountered.
- **FR-6.2** Debug output shall be clearly delimited (e.g., `--- DEBUG: REQUEST ---` / `--- DEBUG: RESPONSE ---`).
- **FR-6.3** The normal confirmation and execution flow continues after debug output.

---

## 5. Non-Functional Requirements

| ID | Category | Requirement |
|----|----------|-------------|
| NFR-1 | Performance | Time from user pressing Enter on the query to command display shall be ≤ 5 s on a typical broadband connection (excluding AI API latency). |
| NFR-2 | Portability | The binary shall be distributed as a single self-contained executable. Supported targets: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`. |
| NFR-3 | Security | The API key shall never be printed to stdout/stderr in normal mode. It shall be masked in debug mode (e.g., `sk-****`). |
| NFR-4 | Security | Config file permissions shall be set to `0600` on Unix systems. |
| NFR-5 | Usability | Color output shall fall back gracefully when the terminal does not support ANSI codes (detect via `NO_COLOR` env var and TTY check). |
| NFR-6 | Reliability | Network errors and non-200 HTTP responses from the AI API shall produce a human-readable error message and exit with code `1`. |
| NFR-7 | Maintainability | Danger-level keyword/pattern list shall be stored in a dedicated, easily editable source file. |

---

## 6. System Architecture

```
┌───────────────────────────────────────────────────────┐
│                     ai-cmd binary                     │
│                                                       │
│  ┌──────────┐   ┌──────────────┐   ┌───────────────┐ │
│  │  CLI     │──▶│  Config      │   │  AI Client    │ │
│  │  Parser  │   │  Resolver    │──▶│  (HTTP/REST)  │ │
│  └──────────┘   └──────────────┘   └───────┬───────┘ │
│       │                                    │         │
│       ▼                                    ▼         │
│  ┌──────────┐   ┌──────────────┐   ┌───────────────┐ │
│  │  OS/Shell│   │  Danger      │◀──│  Response     │ │
│  │  Detector│   │  Classifier  │   │  Parser       │ │
│  └──────────┘   └──────────────┘   └───────────────┘ │
│                        │                             │
│                        ▼                             │
│                ┌───────────────┐                     │
│                │  Confirmation │                     │
│                │  & Executor   │                     │
│                └───────────────┘                     │
└───────────────────────────────────────────────────────┘
```

### 6.1 Modules

| Module | Responsibility |
|--------|----------------|
| `cmd/ai-cmd/main.go` | Entry point; CLI flag parsing |
| `internal/config` | Load, validate, and save configuration |
| `internal/shell` | OS/shell detection; command execution |
| `internal/ai` | Build prompt; call OpenAI-compatible API; parse response |
| `internal/classify` | Danger-level classification via pattern matching |
| `internal/ui` | ANSI color output; confirmation prompt; TTY detection |

---

## 7. AI Prompt Design

### 7.1 System Prompt

```
You are a shell command expert. The user is on {OS} using {SHELL}.
Your task: translate the user's natural-language request into a single, correct shell command.
Rules:
- Output ONLY the raw shell command. No explanations, no markdown, no code fences.
- If you cannot produce a safe and valid command, output: ERROR: <reason>
```

### 7.2 User Message

```
{natural-language query}
```

### 7.3 API Call Parameters

| Parameter | Value |
|-----------|-------|
| `model` | From config |
| `messages` | `[{role: system, ...}, {role: user, ...}]` |
| `temperature` | `0` (deterministic output preferred) |
| `max_tokens` | `256` |

---

## 8. Danger Classification Rules

Classification is applied client-side using regex pattern matching on the command string.

### 8.1 DANGER Patterns (Red)

```
rm\s+-[^\s]*r   # recursive remove (Unix)
del\s+/[sf]     # force/recursive delete (Windows)
mkfs            # format filesystem
dd\s+if=        # disk write
kill\s+-9       # force kill process
pkill|killall   # kill by name
:(){ :|:& };:   # fork bomb
> /dev/sd       # write to block device
shutdown|reboot|halt|poweroff
format\s+[A-Za-z]:  # Windows format drive
```

### 8.2 CAUTION Patterns (Yellow)

```
mkdir|touch|cp|mv|ln   # create/modify files
chmod|chown            # permission changes
Set-Content|New-Item|Copy-Item|Move-Item  # PowerShell equivalents
curl.*-o|wget          # file downloads
pip install|npm install|apt|yum|brew      # package installs
systemctl\s+(start|stop|restart|enable|disable)
```

### 8.3 SAFE (Green)

Any command not matching DANGER or CAUTION patterns.

---

## 9. Error Handling

| Scenario | Behavior |
|----------|----------|
| No query provided | Print usage hint and exit `1` |
| Config not found and no env vars | Prompt user to run `ai -c` and exit `1` |
| API key invalid / 401 response | Print `Error: Invalid API key. Run 'ai -c' to reconfigure.` and exit `1` |
| API rate limit / 429 | Print `Error: API rate limit exceeded. Please try again later.` and exit `1` |
| Network timeout | Print `Error: Request timed out. Check your connection.` and exit `1` |
| AI returns `ERROR: <reason>` | Display `AI could not generate a command: <reason>` and exit `1` |
| Command execution fails | Print stderr from child process; exit with child's exit code |

---

## 10. CLI Interface Summary

```
ai <query>          Translate query and run with confirmation
ai -c               Interactive configuration wizard
ai -h / --help      Show help
ai -d <query>       Debug mode: show raw AI request/response, then normal flow
```

Flag parsing shall use Go's standard `flag` package. Flags must appear **before** the query.

---

## 11. Configuration File Location

| OS | Path |
|----|------|
| Linux / macOS | `~/.ai-cmd/config.json` |
| Windows | `%USERPROFILE%\.ai-cmd\config.json` |

---

## 12. Build & Distribution

- **Language:** Go 1.22+
- **Build tool:** `Makefile` with targets: `build`, `test`, `lint`, `release`
- **Release artifacts:** Single static binary per target platform, named `ai-cmd` (Unix) / `ai-cmd.exe` (Windows)
- **Cross-compilation:** Via `GOOS`/`GOARCH` env vars; no CGO dependencies
- **Linter:** `golangci-lint`
- **Testing:** `go test ./...`; unit tests for classifier, config resolver, response parser, and UI helpers

---

## 13. Out of Scope (v1.0)

- Persistent command history / session context
- Plugin system for custom AI providers
- GUI / TUI interface
- Undo / rollback of executed commands
- Shell auto-completion scripts
- Telemetry / usage analytics
