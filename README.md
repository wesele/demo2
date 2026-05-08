# ai-cmd — AI-Powered Command Line Helper

> Translate natural language into the right shell command for your OS — safely.

## Overview

`ai-cmd` is a cross-platform CLI tool written in Go. Type what you want to do in plain English; `ai-cmd` calls an OpenAI-compatible AI API, returns the correct shell command, warns you with a colour-coded danger level, and only executes after you press Enter.

```
$ ai kill the process on port 8080
[DANGER] kill -9 $(lsof -t -i:8080)

Run this command? [Enter to confirm / Ctrl+C to cancel]:
Process 1234 killed.
```

## Documentation

| Document | Description |
|----------|-------------|
| [`docs/requirement.md`](docs/requirement.md) | Original user requirements |
| [`docs/technical-spec.md`](docs/technical-spec.md) | Technical Requirements & Specification |
| [`docs/high-level-design.md`](docs/high-level-design.md) | High Level Design |
| [`docs/user-stories.md`](docs/user-stories.md) | User Stories & Acceptance Criteria |

## Quick Start

```bash
# Configure (first time)
ai -c

# Run a query
ai find all log files larger than 100MB

# Debug mode
ai -d list all running docker containers

# Help
ai -h
```

## Configuration

Resolution order: **environment variables > `~/.ai-cmd/config.json` > built-in defaults**

| Environment Variable | Purpose |
|---------------------|---------|
| `AI_CMD_ENDPOINT` | API endpoint (default: `https://api.openai.com/v1`) |
| `AI_CMD_API_KEY` | API key |
| `AI_CMD_MODEL` | Model name (default: `gpt-4o`) |

## Supported Platforms

| OS | Architecture |
|----|-------------|
| Linux | amd64, arm64 |
| macOS | amd64, arm64 |
| Windows | amd64 |

## Development

```bash
make build    # compile binary to ./bin/
make test     # run unit tests
make lint     # run golangci-lint
make release  # cross-compile all targets
```

Requires Go 1.22+.

## License

MIT
