# Vauxpraudemonium

> See `AGENTS.md` for comprehensive development guide.

## Overview

Unified monorepo for AI agent development tools:
- **Vauxhall**: Multi-project agent mission control (web + TUI)
- **Praude**: TUI-first PRD generation and validation
- **Tandemonium**: Task orchestration for human-AI collaboration
- **Pollard**: General-purpose research intelligence (tech, medicine, law, economics, etc.)

## Quick Commands

```bash
# Build and run
./dev vauxhall --tui    # Vauxhall TUI mode
./dev vauxhall          # Vauxhall web mode
./dev praude            # Praude TUI
./dev tandemonium       # Tandemonium TUI

# Pollard CLI
go run ./cmd/pollard init           # Initialize .pollard/
go run ./cmd/pollard scan           # Run all hunters
go run ./cmd/pollard scan --hunter github-scout
go run ./cmd/pollard scan --hunter openalex   # Multi-domain academic
go run ./cmd/pollard scan --hunter pubmed     # Medical research
go run ./cmd/pollard report         # Generate landscape report
go run ./cmd/pollard report --type competitive

# Build all
go build ./cmd/...

# Test
go test ./...
```

## Key Paths

| Path | Purpose |
|------|---------|
| `cmd/` | Entry points for each tool |
| `internal/{tool}/` | Tool-specific code |
| `pkg/tui/` | Shared TUI styles (Tokyo Night) |
| `docs/{tool}/` | Tool-specific documentation |
| `.pollard/` | Pollard data directory (sources, insights, reports) |

## Design Decisions (Do Not Re-Ask)

- Module: `github.com/mistakeknot/vauxpraudemonium`
- Shared TUI package with Tokyo Night colors
- Bubble Tea for all TUIs
- htmx + Tailwind for Vauxhall web
- SQLite for local state (read-only to external DBs)
- tmux integration via CLI commands
- Pollard tech hunters use free API tiers (no auth required)
- Pollard general-purpose hunters: some require API keys (USDA, CourtListener)
- Intermute for cross-tool messaging (file-based until HTTP API built)
