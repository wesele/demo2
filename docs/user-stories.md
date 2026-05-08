# User Stories
## ai-cmd — AI-Powered Command Line Helper

**Version:** 1.0  
**Date:** 2026-05-08  
**Status:** Confirmed  
**References:** `technical-spec.md` v1.0, `high-level-design.md` v1.0

---

## 1. Story Map Overview

Stories are grouped into **Epics** aligned to product capabilities. Each story follows the standard format:

> **As a** [persona], **I want to** [action], **so that** [benefit].

Acceptance criteria use the **Given / When / Then** (GWT) pattern.

---

## 2. Personas

| ID | Persona | Description |
|----|---------|-------------|
| P1 | **CLI User** | Developer or sysadmin who regularly uses a terminal but sometimes forgets exact command syntax |
| P2 | **New User** | Someone setting up `ai-cmd` for the first time with no prior knowledge of the tool |
| P3 | **Power User** | DevOps engineer running `ai-cmd` in scripts, CI pipelines, or multi-server environments |
| P4 | **Debugging User** | Developer troubleshooting why the tool is producing incorrect or unexpected output |

---

## 3. Epics

| Epic ID | Name | Description |
|---------|------|-------------|
| E-1 | Command Translation | Convert natural language to OS-aware shell commands via AI |
| E-2 | Safety & Confirmation | Classify danger level and require explicit confirmation before execution |
| E-3 | Configuration Management | Setup wizard, config file persistence, and env-var override |
| E-4 | Developer Experience | Help text, debug mode, and error messaging |
| E-5 | Build & Distribution | Cross-platform binary, CI/CD pipeline, and release automation |

---

## 4. User Stories

---

### EPIC E-1: Command Translation

---

#### US-101 — Natural Language Query

**As a** CLI User,  
**I want to** type `ai <natural language description>` in my terminal,  
**so that** I get the correct shell command without having to search online.

**Priority:** Must Have  
**Story Points:** 5

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-101-1 | The tool is configured | I run `ai list all running processes` | The tool displays a single shell command appropriate for my OS |
| AC-101-2 | I am on Linux/macOS | I run any query | The returned command uses Bash/Zsh syntax |
| AC-101-3 | I am on Windows | I run any query | The returned command uses PowerShell syntax |
| AC-101-4 | The AI returns a markdown code block | I run any query | The tool strips the fences and displays only the raw command |

**Definition of Done:**
- [ ] `internal/ai` package implemented and unit-tested
- [ ] `internal/shell/detector.go` detects OS and shell correctly on all 3 platforms
- [ ] Response cleaning strips ``` fences and leading/trailing whitespace
- [ ] Integration test passes with mock HTTP server

---

#### US-102 — OS and Shell Context Awareness

**As a** CLI User,  
**I want** the generated command to match my current OS and shell,  
**so that** I don't have to mentally translate between Bash and PowerShell syntax.

**Priority:** Must Have  
**Story Points:** 3

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-102-1 | I am on Windows (PowerShell) | I ask to find large files | The command uses `Get-ChildItem` / `Where-Object`, not `find` |
| AC-102-2 | I am on Linux (Bash) | I ask to kill a process | The command uses `kill` / `lsof`, not `Stop-Process` |
| AC-102-3 | `$SHELL` is `/bin/zsh` | I run a query | The system prompt includes `zsh` as the shell context |

**Definition of Done:**
- [ ] `ShellInfo` struct carries `OS`, `Shell` fields
- [ ] System prompt template substitutes `{OS}` and `{SHELL}` correctly
- [ ] Unit tests cover Windows, Linux, macOS detection paths

---

#### US-103 — Single-Shot Execution Model

**As a** CLI User,  
**I want** each `ai` invocation to be independent,  
**so that** previous queries do not pollute the context of new ones.

**Priority:** Must Have  
**Story Points:** 1

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-103-1 | I ran `ai list files` previously | I run `ai delete temp folder` | No history from the previous query is sent to the AI |

**Definition of Done:**
- [ ] No session state is stored between invocations
- [ ] Confirmed in code review that `messages` array contains only system + current user message

---

### EPIC E-2: Safety & Confirmation

---

#### US-201 — Danger Level Color Coding

**As a** CLI User,  
**I want** the suggested command to be displayed with a color-coded danger label,  
**so that** I am visually warned before running anything destructive.

**Priority:** Must Have  
**Story Points:** 3

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-201-1 | stdout is a TTY | The AI returns `ls -la` | The command is displayed with a green `[SAFE]` label |
| AC-201-2 | stdout is a TTY | The AI returns `rm -rf /tmp/logs` | The command is displayed with a red `[DANGER]` label |
| AC-201-3 | stdout is a TTY | The AI returns `mkdir /tmp/test` | The command is displayed with a yellow `[CAUTION]` label |
| AC-201-4 | stdout is piped (not a TTY) | Any command is returned | No ANSI color codes appear in the output |
| AC-201-5 | `NO_COLOR=1` is set | Any command is returned | No ANSI color codes appear in the output |

**Definition of Done:**
- [ ] `internal/classify` implements all DANGER/CAUTION regex patterns from spec §8
- [ ] `internal/ui/color.go` implements TTY detection and `NO_COLOR` check
- [ ] Unit tests cover all three classification levels plus edge cases

---

#### US-202 — Explicit Confirmation Before Execution

**As a** CLI User,  
**I want** to be required to press Enter before any command is executed,  
**so that** I never accidentally destroy my system.

**Priority:** Must Have  
**Story Points:** 2

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-202-1 | A command is displayed | I press Enter | The command executes |
| AC-202-2 | A command is displayed | I press Ctrl+C | The tool exits without executing anything |
| AC-202-3 | A command is displayed | I press Ctrl+D (EOF) | The tool exits without executing anything |
| AC-202-4 | A DANGER command is displayed | I press Enter | The command still executes (user confirmed) |

**Definition of Done:**
- [ ] `internal/ui/prompt.go` implements `Confirm() bool`
- [ ] Ctrl+C and EOF both return `false` from `Confirm()`
- [ ] No command is ever spawned when `Confirm()` returns `false`

---

#### US-203 — Command Execution with Output Streaming

**As a** CLI User,  
**I want** the executed command's output to stream directly to my terminal,  
**so that** I see results in real time just as if I had typed the command myself.

**Priority:** Must Have  
**Story Points:** 3

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-203-1 | I confirm a long-running command | The command runs | stdout and stderr are streamed live (not buffered) |
| AC-203-2 | The command exits with code 0 | Execution completes | `ai-cmd` exits with code 0 |
| AC-203-3 | The command exits with code 2 | Execution completes | `ai-cmd` exits with code 2 |
| AC-203-4 | I am on Linux | I confirm any command | The command is spawned via `$SHELL -c` |
| AC-203-5 | I am on Windows | I confirm any command | The command is spawned via `powershell.exe -Command` |

**Definition of Done:**
- [ ] `internal/shell/executor.go` uses `os/exec` with `Stdout`/`Stderr` connected to `os.Stdout`/`os.Stderr`
- [ ] Exit code propagation verified in integration test

---

### EPIC E-3: Configuration Management

---

#### US-301 — Interactive Configuration Wizard

**As a** New User,  
**I want to** run `ai -c` to be guided through setup interactively,  
**so that** I don't have to manually create or edit config files.

**Priority:** Must Have  
**Story Points:** 3

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-301-1 | No config file exists | I run `ai -c` | I am prompted for endpoint, API key, and model name in sequence |
| AC-301-2 | I provide a valid endpoint URL | The wizard validates it | It is accepted and I am moved to the next prompt |
| AC-301-3 | I provide an invalid URL | The wizard validates it | I see an error and am re-prompted |
| AC-301-4 | I complete all prompts | The wizard finishes | `~/.ai-cmd/config.json` is created with `0600` permissions |
| AC-301-5 | A config file already exists | I run `ai -c` | The wizard shows current values and lets me keep or overwrite each |

**Definition of Done:**
- [ ] `internal/config/wizard.go` implemented
- [ ] File saved with `0600` permissions on Unix; directory created if absent
- [ ] Input validation tested for URL format and non-empty strings

---

#### US-302 — Environment Variable Configuration Override

**As a** Power User,  
**I want to** set `AI_CMD_ENDPOINT`, `AI_CMD_API_KEY`, and `AI_CMD_MODEL` environment variables,  
**so that** I can configure the tool in CI pipelines or scripts without touching disk.

**Priority:** Must Have  
**Story Points:** 2

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-302-1 | `AI_CMD_API_KEY` is set | I run any query | The env var key is used, not the config file key |
| AC-302-2 | Both env var and config file are present | I run any query | The env var takes precedence |
| AC-302-3 | No config file and no env vars | I run any query | The tool prints an error prompting `ai -c` and exits 1 |
| AC-302-4 | Only `AI_CMD_API_KEY` env var is set | I run any query | Endpoint and model fall back to config file or defaults |

**Definition of Done:**
- [ ] `internal/config/config.go` implements three-tier resolution (env > file > default)
- [ ] Unit tests cover all permutations of present/absent env vars and config file

---

#### US-303 — Config File Security

**As a** CLI User,  
**I want** my config file to be stored securely,  
**so that** my API key is not readable by other users on a shared system.

**Priority:** Must Have  
**Story Points:** 1

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-303-1 | I run `ai -c` on Linux/macOS | The wizard saves the file | File permissions are exactly `0600` |
| AC-303-2 | The `~/.ai-cmd` directory did not exist | The wizard runs | The directory is created with `0700` permissions |

**Definition of Done:**
- [ ] `os.WriteFile` called with `0600` mode
- [ ] `os.MkdirAll` called with `0700` mode
- [ ] Verified on Linux with `stat` in integration test

---

### EPIC E-4: Developer Experience

---

#### US-401 — Help Text

**As a** CLI User,  
**I want to** run `ai -h` or `ai --help`,  
**so that** I can quickly see how to use the tool without reading documentation.

**Priority:** Must Have  
**Story Points:** 1

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-401-1 | Any state | I run `ai -h` | Usage, all flags, and at least 2 examples are printed to stdout |
| AC-401-2 | Any state | I run `ai --help` | Same output as `-h` |
| AC-401-3 | Any state | I run `ai -h` | The tool exits with code 0 |

**Definition of Done:**
- [ ] Help text includes synopsis, all flags (`-c`, `-d`, `-h`), and 2+ examples
- [ ] Implemented via `flag.Usage` override in `main.go`

---

#### US-402 — Debug Mode

**As a** Debugging User,  
**I want to** run `ai -d <query>`,  
**so that** I can inspect the exact prompt sent to the AI and the raw response received.

**Priority:** Should Have  
**Story Points:** 2

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-402-1 | `-d` flag is passed | I run a query | The full system + user prompt is printed to stderr before the API call |
| AC-402-2 | `-d` flag is passed | The API responds | The raw HTTP response body is printed to stderr |
| AC-402-3 | `-d` flag is passed | Debug output appears | The API key is masked as `sk-****` |
| AC-402-4 | `-d` flag is passed | Debug output appears | Normal classify → confirm → execute flow continues afterward |
| AC-402-5 | `-d` flag is passed | An API error occurs | The raw error response body is shown, then the error message |

**Definition of Done:**
- [ ] Debug sections delimited by `--- DEBUG: REQUEST ---` and `--- DEBUG: RESPONSE ---`
- [ ] API key masking applies to both request headers and any echoed config values
- [ ] Unit test verifies key masking logic

---

#### US-403 — Human-Readable Error Messages

**As a** CLI User,  
**I want** clear, actionable error messages when something goes wrong,  
**so that** I know exactly what to do to fix the issue.

**Priority:** Must Have  
**Story Points:** 2

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-403-1 | I run `ai` with no query | - | I see a usage hint and exit code 1 |
| AC-403-2 | The API key is wrong | I run a query | I see `Error: Invalid API key. Run 'ai -c' to reconfigure.` |
| AC-403-3 | The network is unavailable | I run a query | I see `Error: Request timed out. Check your connection.` |
| AC-403-4 | The AI cannot generate a command | I run a query | I see `AI could not generate a command: <reason>` |
| AC-403-5 | The API returns 429 | I run a query | I see `Error: API rate limit exceeded. Please try again later.` |

**Definition of Done:**
- [ ] All error scenarios from spec §9 are mapped to messages in an error handler
- [ ] All errors go to stderr; exit code is always 1 for application errors

---

### EPIC E-5: Build & Distribution

---

#### US-501 — Project Scaffold and CI

**As a** Developer,  
**I want** a properly scaffolded Go project with linting and testing in CI,  
**so that** every pull request is automatically validated.

**Priority:** Must Have  
**Story Points:** 3

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-501-1 | A PR is opened | CI runs | `golangci-lint` and `go test ./...` execute and must pass |
| AC-501-2 | Any commit | CI runs | A `go build` check confirms the binary compiles |
| AC-501-3 | The repo is cloned fresh | `make build` is run | A working binary is produced in `./bin/` |

**Definition of Done:**
- [ ] `go.mod` initialised with module name `github.com/wesele/demo2`
- [ ] `Makefile` has `build`, `test`, `lint`, `clean` targets
- [ ] `.github/workflows/ci.yml` implemented and passing

---

#### US-502 — Cross-Platform Release Binary

**As a** CLI User,  
**I want to** download a single pre-built binary for my OS and architecture,  
**so that** I don't need a Go toolchain to install the tool.

**Priority:** Must Have  
**Story Points:** 3

**Acceptance Criteria:**

| # | Given | When | Then |
|---|-------|------|------|
| AC-502-1 | A git tag `vX.Y.Z` is pushed | Release CI runs | Binaries are built for all 5 targets and attached to a GitHub Release |
| AC-502-2 | I download the Linux binary | I run it on Linux AMD64 | It executes without any shared library errors |
| AC-502-3 | I download the Windows binary | I run it on Windows AMD64 | It executes as a native `.exe` |
| AC-502-4 | Release CI runs | - | Release notes are auto-generated from git log |

**Supported Targets:**
- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`

**Definition of Done:**
- [ ] `.github/workflows/release.yml` implemented
- [ ] Zero CGO confirmed (`CGO_ENABLED=0`)
- [ ] All 5 binaries attached to GitHub Release

---

## 5. Story Summary

| Epic | Story ID | Title | Priority | Points |
|------|----------|-------|----------|--------|
| E-1 | US-101 | Natural Language Query | Must Have | 5 |
| E-1 | US-102 | OS and Shell Context Awareness | Must Have | 3 |
| E-1 | US-103 | Single-Shot Execution Model | Must Have | 1 |
| E-2 | US-201 | Danger Level Color Coding | Must Have | 3 |
| E-2 | US-202 | Explicit Confirmation Before Execution | Must Have | 2 |
| E-2 | US-203 | Command Execution with Output Streaming | Must Have | 3 |
| E-3 | US-301 | Interactive Configuration Wizard | Must Have | 3 |
| E-3 | US-302 | Environment Variable Configuration Override | Must Have | 2 |
| E-3 | US-303 | Config File Security | Must Have | 1 |
| E-4 | US-401 | Help Text | Must Have | 1 |
| E-4 | US-402 | Debug Mode | Should Have | 2 |
| E-4 | US-403 | Human-Readable Error Messages | Must Have | 2 |
| E-5 | US-501 | Project Scaffold and CI | Must Have | 3 |
| E-5 | US-502 | Cross-Platform Release Binary | Must Have | 3 |
| | | **Total** | | **34** |

---

## 6. Prioritisation (MoSCoW)

| Priority | Stories |
|----------|---------|
| **Must Have** | US-101, US-102, US-103, US-201, US-202, US-203, US-301, US-302, US-303, US-401, US-403, US-501, US-502 |
| **Should Have** | US-402 |
| **Could Have** | Shell auto-completion (future) |
| **Won't Have (v1.0)** | Session memory, GUI, undo/rollback, telemetry |

---

## 7. Suggested Sprint Plan

| Sprint | Stories | Goal |
|--------|---------|------|
| Sprint 1 | US-501, US-301, US-302, US-303 | Repo live, config system working |
| Sprint 2 | US-101, US-102, US-103 | AI translation end-to-end |
| Sprint 3 | US-201, US-202, US-203 | Safe execution with UX guardrails |
| Sprint 4 | US-401, US-402, US-403, US-502 | Polish, debug mode, release pipeline |
