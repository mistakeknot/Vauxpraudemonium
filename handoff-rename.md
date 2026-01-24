# Handoff: Rename Plan (Bigend / Gurgeh / Coldwine) - Delete After Review

IMPORTANT: Please delete this file after you have reviewed and transferred the notes.

## Decision Summary
User wants a full renaming to unify names across the product suite using character names:
- Vauxhall -> Bigend
- Praude -> Gurgeh
- Tandemonium -> Coldwine
- Pollard remains as a submodule/aux component

Scope directive from user: "rename everything" (code, docs, CLI, config, dirs, package names,
commands, and branding). This implies a wide refactor and consistent compatibility strategy.

## Recommended Approach (Full Rename)
1) Define the rename map and compatibility rules.
2) Use git mv for dirs/files, then update imports and package names.
3) Update all user-facing strings, docs, READMEs, and config paths.
4) Add backward-compat shims for old config paths and CLI aliases (recommended).
5) Run gofmt + go test ./... and any JS/TS builds if touched.

## Rename Map (High-level)
- Binaries/commands: vauxhall -> bigend, praude -> gurgeh, tandemonium -> coldwine
- Directories:
  - cmd/vauxhall -> cmd/bigend
  - internal/vauxhall -> internal/bigend
  - docs/vauxhall -> docs/bigend
  - cmd/praude -> cmd/gurgeh
  - internal/praude -> internal/gurgeh
  - docs/praude -> docs/gurgeh
  - cmd/tandemonium -> cmd/coldwine
  - internal/tandemonium -> internal/coldwine
  - docs/tandemonium -> docs/coldwine
- Project docs: README.md, docs/plans/*, AGENTS.md, and any tool-specific docs
- Config folders and files:
  - ~/.config/vauxhall -> ~/.config/bigend
  - ~/.config/praude -> ~/.config/gurgeh
  - ~/.config/tandemonium -> ~/.config/coldwine
  - .praude -> .gurgeh
  - .tandemonium -> .coldwine
- Environment variables:
  - VAUXHALL_* -> BIGEND_*
  - PRAUDE_* -> GURGEH_*
  - TANDEMONIUM_* -> COLDWINE_*

## Code-level Renaming (Go)
- Package names: package vauxhall -> package bigend (etc.)
- Import paths change accordingly (search/replace after git mv)
- Update module-level references in Go code and tests
- Update any hard-coded names in:
  - CLI help, usage strings
  - log/slog logger names
  - config section keys
  - file path templates
  - dev script entries (./dev vauxhall -> ./dev bigend)

## Pollard Handling
- Pollard remains and should not be renamed.
- Identify Pollard-related packages, docs, or config and ensure they keep their name.
- If Pollard is referenced as a submodule of a specific product, update surrounding
  names (e.g., "Bigend Pollard"), but do not rename Pollard itself.

## Compatibility / Migration Strategy (Strongly Recommended)
- Backward compatibility for config locations:
  - If new config path not present, check old path and log a warning.
- CLI aliases:
  - Consider wrapper binaries (or a message) for old command names that print
    a "renamed to" notice.
- Any file naming changes in .praude/.tandemonium equivalents should be migrated
  or read with fallback logic to avoid breaking existing repos.

## Risk Areas
- Massive import churn in Go: use gofmt and go test frequently.
- config path changes can silently break user setups; add fallback logic.
- docs/plans and historic references: keep consistency but avoid breaking
  historical filenames if they are used programmatically.
- dev script targets, CI, or packaging might rely on old names.

## Suggested Tactical Plan
1) Rename config folders (.praude/.tandemonium -> .gurgeh/.coldwine) and decide
   whether to keep legacy fallbacks.
2) Rename directories via git mv (cmd/internal/docs) first.
3) Update Go package names and imports (gofmt).
4) Update dev script, README, docs, CLI help strings.
5) Update config names, env vars, and path logic with fallbacks.
6) Run tests and fix build errors.
7) Update any release scripts or external references if present.

## Open Questions for Owner
- Should repo name Vauxpraudemonium remain, or should it also be renamed?
- Do we want to keep legacy config directories forever or deprecate on a timeline?
- Should we preserve old CLI names as aliases or remove outright?

## Suggested Search Commands
- Find all occurrences:
  - rg -n "Vauxhall|Praude|Tandemonium|vauxhall|praude|tandemonium" .
- Go packages:
  - rg -n "package (vauxhall|praude|tandemonium)" internal cmd
- Config paths:
  - rg -n "\.praude|\.tandemonium|~/.config" .
- Env vars:
  - rg -n "VAUXHALL_|PRAUDE_|TANDEMONIUM_" .

Delete this file after review.
