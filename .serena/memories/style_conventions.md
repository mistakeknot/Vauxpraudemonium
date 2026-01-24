# Code Style & Conventions

## Package Organization
- `internal/` for all tool-specific packages
- `pkg/` only for shared code across tools
- Each tool has: cli/, config/, tui/, and domain packages

## Error Handling
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Return errors up the stack, handle at CLI level

## Logging
- Use `log/slog` with structured fields
- No external logging dependencies

## Naming
- Use descriptive names, avoid abbreviations
- Struct methods: verb-noun (LoadConfig, SaveState)
- Interfaces: -er suffix where appropriate (Scanner, Hunter)

## Dependencies
- Minimize external dependencies
- Prefer standard library
- SQLite: use modernc.org/sqlite (CGO-free)

## TUI Components
- Use shared `pkg/tui` for consistent styling
- Tokyo Night color palette
- Bubble Tea patterns for all TUIs
