# Autarch

> **AI Agent Development Tools Suite**

Autarch is a unified monorepo for AI agent development tools, following the Culture novel ship naming convention.

## Tools

| Tool | Purpose | Command |
|------|---------|---------|
| **Bigend** | Multi-project agent mission control (web + TUI) | `./dev bigend` |
| **Gurgeh** | TUI-first PRD generation and validation | `./dev gurgeh` |
| **Coldwine** | Task orchestration for human-AI collaboration | `./dev coldwine` |
| **Pollard** | Continuous research intelligence (hunters + reports) | `go run ./cmd/pollard` |

## Tech Stack

- **Language:** Go 1.24+
- **Module:** `github.com/mistakeknot/autarch`
- **TUI:** Bubble Tea + Lip Gloss (Tokyo Night palette)
- **Web:** net/http + htmx + Tailwind
- **Database:** SQLite (WAL mode)

## Prerequisites

- macOS or Linux
- Go 1.24+
- tmux (for Coldwine)
- Git 2.23+ (for worktrees)

## Quick Start

```bash
# Build all tools
go build ./cmd/...

# Run individual tools
./dev bigend      # Mission control (web mode)
./dev bigend --tui # Mission control (TUI mode)
./dev gurgeh      # PRD generation
./dev coldwine    # Task orchestration

# Pollard (research intelligence)
go run ./cmd/pollard init
go run ./cmd/pollard scan
go run ./cmd/pollard report
```

## Project Structure

```
autarch/
├── cmd/
│   ├── bigend/       # Mission control entry point
│   ├── gurgeh/       # PRD generation entry point
│   ├── coldwine/     # Task orchestration entry point
│   └── pollard/      # Research intelligence entry point
├── internal/
│   ├── bigend/       # Bigend-specific code
│   ├── gurgeh/       # Gurgeh-specific code
│   ├── coldwine/     # Coldwine-specific code
│   └── pollard/      # Pollard-specific code
├── pkg/
│   ├── contract/     # Cross-tool entity types
│   ├── discovery/    # Project discovery
│   ├── events/       # Event spine (SQLite)
│   ├── shell/        # Shell integration
│   └── tui/          # Shared TUI components
└── docs/             # Documentation
```

## Configuration

Each tool has its own config directory:
- Bigend: `.bigend/`
- Gurgeh: `.gurgeh/` (legacy: `.praude/`)
- Coldwine: `.coldwine/` (legacy: `.tandemonium/`)
- Pollard: `.pollard/`

Global agent targets: `~/.config/autarch/agents.toml`

## Intermute Integration

Autarch modules will auto-register with Intermute when `INTERMUTE_URL` is set:

```bash
export INTERMUTE_URL="http://localhost:7338"
export INTERMUTE_AGENT_NAME="my-agent"   # optional
export INTERMUTE_PROJECT="my-project"    # optional
```

Bigend handles session I/O; Intermute provides coordination and messaging.

## Documentation

See `AGENTS.md` for comprehensive development guide.
