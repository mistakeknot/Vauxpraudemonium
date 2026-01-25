# Autarch Project Overview

Unified monorepo for AI agent development tools:
- **Vauxhall**: Multi-project agent mission control (web + TUI)
- **Praude**: TUI-first PRD generation and validation
- **Tandemonium**: Task orchestration for human-AI collaboration
- **Pollard**: Continuous research intelligence (in development)

## Tech Stack
- Language: Go 1.24+
- Module: `github.com/mistakeknot/autarch`
- TUI: Bubble Tea + lipgloss (Tokyo Night colors)
- Web: net/http + htmx + Tailwind
- Database: SQLite (WAL mode, modernc.org/sqlite for CGO-free)

## Project Structure
```
cmd/           - Entry points (vauxhall/, praude/, tandemonium/, pollard/)
internal/      - Tool-specific code (vauxhall/, praude/, tandemonium/, pollard/)
pkg/           - Shared code (tui/, agenttargets/)
docs/          - Documentation per tool
```
