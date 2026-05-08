# High Level Design (HLD)
## ai-cmd — AI-Powered Command Line Helper

**Version:** 1.0  
**Date:** 2026-05-08  
**Status:** Draft  
**References:** `technical-spec.md` v1.0, `user-stories.md` v1.0

---

## 1. Purpose

This document describes the high-level design of `ai-cmd`. It bridges the confirmed technical specification and the implementation, giving the engineering team a shared architectural blueprint before detailed (low-level) design and coding begins.

---

## 2. Design Principles

| Principle | Application |
|-----------|-------------|
| **Single Responsibility** | Each internal package owns exactly one concern (config, AI, classify, shell, UI). |
| **Fail Safe** | Any error — network, config, AI — produces a human-readable message and a non-zero exit code; the command is never executed on ambiguous output. |
| **Minimal Surface** | No external runtime dependencies; pure Go stdlib + one HTTP client. Zero CGO. |
| **Graceful Degradation** | Color/ANSI output is opt-in based on TTY detection; the tool remains fully functional in plain-text mode. |
| **Security by Default** | API keys are never echoed; config files are created with `0600` permissions. |

---

## 3. System Context Diagram

```
 ┌─────────────────────────────────────────────────────────────────┐
 │                          User's Terminal                        │
 │                                                                 │
 │   $ ai kill the process on port 8080                           │
 │                          │                                      │
 │              ┌───────────▼────────────┐                        │
 │              │       ai-cmd binary    │                        │
 │              │   (single Go binary)   │                        │
 │              └───────────┬────────────┘                        │
 │                          │ HTTPS / OpenAI-compatible REST API  │
 │              ┌───────────▼────────────┐                        │
 │              │    AI API Endpoint     │                        │
 │              │  (OpenAI / compatible) │                        │
 │              └────────────────────────┘                        │
 │                                                                 │
 │   Config: ~/.ai-cmd/config.json  |  Env vars: AI_CMD_*        │
 └─────────────────────────────────────────────────────────────────┘
```

---

## 4. Component Architecture

### 4.1 Component Overview

```
cmd/ai-cmd/
  main.go                  ← Entry point: flag parsing, orchestration

internal/
  config/
    config.go              ← Config struct, load/save, env override
    wizard.go              ← Interactive -c setup wizard
  ai/
    client.go              ← HTTP client, request builder, response parser
    prompt.go              ← System/user prompt templates
  classify/
    classifier.go          ← Danger-level regex engine
    patterns.go            ← DANGER / CAUTION pattern definitions
  shell/
    detector.go            ← OS + active shell detection
    executor.go            ← Spawn child process, stream I/O
  ui/
    color.go               ← ANSI color helpers, TTY detection
    prompt.go              ← Confirmation prompt, user input
```

### 4.2 Component Responsibilities

| Component | Responsibility | Key Interfaces |
|-----------|---------------|----------------|
| `main` | Parse flags; wire components; orchestrate flow | `os.Args`, exit codes |
| `config` | Load JSON config; apply env var overrides; persist via wizard | `Config` struct, `Load()`, `Save()` |
| `ai` | Build prompt with OS/shell context; POST to API; parse JSON response; strip prose | `Translate(query, config) (string, error)` |
| `classify` | Match command string against regex patterns; return level | `Classify(cmd string) Level` |
| `shell` | Detect `GOOS`, `$SHELL`/`$COMSPEC`; spawn child with inherited env | `Detect() ShellInfo`, `Execute(cmd string) error` |
| `ui` | Render colored output; render confirmation prompt; mask API key | `Print(cmd, level)`, `Confirm() bool` |

---

## 5. Data Flow

### 5.1 Normal Query Flow

```
User input
    │
    ▼
[main] Parse flags & query text
    │
    ▼
[config] Load config (file → env override)
    │  config not found? → error + hint to run `ai -c`
    ▼
[shell.Detect] Identify OS + shell
    │
    ▼
[ai.Translate] Build prompt → POST to API → parse response
    │  HTTP/network error? → human-readable error, exit 1
    │  AI returns ERROR:?  → display reason, exit 1
    ▼
[classify.Classify] Match command against DANGER/CAUTION patterns
    │
    ▼
[ui.Print] Render command with color label (SAFE/CAUTION/DANGER)
    │
    ▼
[ui.Confirm] Display confirmation prompt; wait for Enter or Ctrl+C
    │  Ctrl+C / Ctrl+D? → exit 0 (no execution)
    ▼
[shell.Execute] Spawn child process; stream stdout+stderr
    │
    ▼
Exit with child process exit code
```

### 5.2 Configuration Wizard Flow (`ai -c`)

```
[main] -c flag detected
    │
    ▼
[config.Load] Attempt to load existing config (show current values if found)
    │
    ▼
[ui] Prompt for endpoint (validate URL format)
[ui] Prompt for API key (non-empty, masked input)
[ui] Prompt for model name (non-empty)
    │
    ▼
[config.Save] Write ~/.ai-cmd/config.json with 0600 permissions
    │
    ▼
Print success message; exit 0
```

### 5.3 Debug Flow (`ai -d <query>`)

```
[main] -d flag detected; normal flow begins
    │
    ▼
[ai] Before HTTP call: print "--- DEBUG: REQUEST ---" + full prompt to stderr
    │
    ▼
[ai] After HTTP response: print "--- DEBUG: RESPONSE ---" + raw body to stderr
    │  Mask api_key as sk-****
    ▼
Continue with normal classify → display → confirm → execute flow
```

---

## 6. Module Interaction Diagram

```
┌──────────┐     flags/query     ┌──────────────────────────────────────────┐
│  main    │────────────────────▶│              Orchestration               │
└──────────┘                     └────┬──────┬──────┬──────┬──────┬────────┘
                                      │      │      │      │      │
                               config │  ai  │shell │class │  ui  │
                                      ▼      ▼      ▼      ▼      ▼
                                   Load  Translate Detect Classify Print
                                   Save            Execute         Confirm
```

---

## 7. Key Design Decisions

### 7.1 Single Binary, No CGO

`ai-cmd` compiles to a standalone binary with no shared library dependencies. Cross-compilation is achieved via `GOOS`/`GOARCH`. This satisfies NFR-2 (portability) and simplifies distribution.

### 7.2 Client-Side Danger Classification

Classification uses pre-compiled regex patterns in `internal/classify/patterns.go` (not a second AI round-trip). This keeps latency near zero and makes the pattern list auditable and easy to extend without model changes.

### 7.3 Prompt Engineering for Determinism

`temperature: 0` and a strict system prompt ("Output ONLY the raw shell command") minimise hallucination and prose leakage. Response cleaning (strip markdown fences) acts as a safety net.

### 7.4 Child Process Execution Strategy

The tool spawns the command via the user's current shell (`$SHELL -c` / `powershell.exe -Command`) rather than `exec` directly. This preserves aliases, functions, and the working directory, at the cost of a shell startup overhead that is acceptable for interactive use.

### 7.5 Config Resolution Order

`Env vars > config file > defaults` is the industry-standard twelve-factor app pattern. It allows CI/CD pipelines to inject credentials without touching disk, while interactive users benefit from the wizard.

### 7.6 TTY Detection for Color

ANSI color codes are only emitted when stdout is a TTY **and** the `NO_COLOR` environment variable is not set. This follows the [no-color.org](https://no-color.org) convention, ensuring the tool composes cleanly in scripts.

---

## 8. Directory & Repository Layout

```
ai-cmd/
├── cmd/
│   └── ai-cmd/
│       └── main.go
├── internal/
│   ├── ai/
│   │   ├── client.go
│   │   └── prompt.go
│   ├── classify/
│   │   ├── classifier.go
│   │   └── patterns.go
│   ├── config/
│   │   ├── config.go
│   │   └── wizard.go
│   ├── shell/
│   │   ├── detector.go
│   │   └── executor.go
│   └── ui/
│       ├── color.go
│       └── prompt.go
├── docs/
│   ├── requirement.md
│   ├── technical-spec.md
│   ├── high-level-design.md
│   └── user-stories.md
├── .github/
│   └── workflows/
│       ├── ci.yml          ← lint + test on PR
│       └── release.yml     ← cross-compile + publish on tag
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

## 9. CI/CD Pipeline

```
  ┌──────────────────────────────────────────────────────────┐
  │  Pull Request                                            │
  │  ci.yml:  golangci-lint → go test ./... → build check   │
  └──────────────────────────────────────────────────────────┘
                          │ merge to main
                          ▼
  ┌──────────────────────────────────────────────────────────┐
  │  Git Tag (vX.Y.Z)                                        │
  │  release.yml:                                            │
  │    Cross-compile (linux/amd64, linux/arm64,              │
  │                   darwin/amd64, darwin/arm64,            │
  │                   windows/amd64)                         │
  │    → GitHub Release with binary attachments              │
  └──────────────────────────────────────────────────────────┘
```

---

## 10. Security Considerations

| Risk | Mitigation |
|------|-----------|
| API key exposure in logs | Key masked in debug mode (`sk-****`); never printed in normal mode |
| API key exposure on disk | Config file created with `0600` permissions; directory `~/.ai-cmd` with `0700` |
| Prompt injection via query | System prompt strictly instructs model to output only a command; any non-command output is caught by the `ERROR:` check and rejected |
| Accidental destructive execution | Two safeguards: danger classification warns user visually; explicit Enter confirmation required |
| Supply chain | Zero external Go dependencies beyond stdlib; no CGO; reproducible builds |

---

## 11. Error Handling Strategy

All errors propagate to `main` as Go `error` values. `main` maps them to:
- A human-readable message printed to **stderr**.
- An appropriate non-zero exit code (`1` for all application errors; child exit code for execution failures).

No panics are allowed to reach the user; all `panic`-prone operations (type assertions, slice access) are guarded.

---

## 12. Testing Strategy

| Layer | What is tested | Tool |
|-------|---------------|------|
| Unit | `classify`: all DANGER/CAUTION/SAFE patterns; edge cases | `go test` |
| Unit | `config`: load, save, env override, missing file | `go test` |
| Unit | `ai`: response parsing, markdown stripping, error detection | `go test` with HTTP mock |
| Unit | `ui`: TTY detection, color suppression, key masking | `go test` |
| Unit | `shell`: OS detection logic | `go test` |
| Integration | Full flow with a mock HTTP server returning known responses | `go test -tags integration` |
| Manual / E2E | Smoke test against real API endpoint before release | Checklist in `docs/release-checklist.md` |

---

## 13. Milestones & Delivery Plan

| Milestone | Description | Key Deliverables |
|-----------|-------------|-----------------|
| **M0 — Project Setup** | Repo scaffold, CI skeleton, Go module init | `go.mod`, `Makefile`, `ci.yml` |
| **M1 — Core Engine** | Config, AI client, classifier, shell detector | `internal/config`, `internal/ai`, `internal/classify`, `internal/shell` |
| **M2 — CLI & UX** | Flag parsing, color UI, confirmation prompt | `cmd/ai-cmd/main.go`, `internal/ui` |
| **M3 — Hardening** | Error handling, debug mode, security review | All error paths, `-d` flag, key masking |
| **M4 — Release** | Cross-compile, CI release pipeline, README | `release.yml`, `README.md`, GitHub Release v1.0.0 |

---

## 14. Open Questions

| # | Question | Owner | Target |
|---|----------|-------|--------|
| OQ-1 | Should the `-d` flag mask the full API key or show last 4 chars? | Tech Lead | M3 |
| OQ-2 | Timeout value for AI API HTTP request — 30 s default? | Tech Lead | M1 |
| OQ-3 | Should `ai -c` support non-interactive / `--yes` flag for scripted setup? | Product | M2 |
