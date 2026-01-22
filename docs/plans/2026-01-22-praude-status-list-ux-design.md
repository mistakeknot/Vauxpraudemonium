# Praude Status Grouped List UX (Agent-Deck Port)

## Goal
Port the agent-deck list UX into Praude’s TUI with minimal behavioral divergence. Add explicit PRD status fields and group the list by status. Preserve Praude’s existing interview/research/suggestions flows and detail panel.

## Scope
- Full port of agent-deck list UX patterns: hierarchical groups, expand/collapse, scroll indicators, selection styling, and search overlay.
- Responsive layout parity: single (<50 cols), stacked (50-79), dual (80+).
- Persist list UI state (expanded groups, last selection) in `.praude/state.json`.
- Introduce explicit `status` on PRD YAML and load it in summaries.

Out of scope:
- Global search across non-Praude data.
- Shared `pkg/tui` extraction (may follow later).

## Architecture
Praude’s TUI stays in `internal/praude/tui`. We add a local group-tree model (inspired by agent-deck) that organizes PRD summaries into status buckets. The list renderer consumes the flattened items and renders status headers + PRD entries with selection, icons, and scroll indicators. The detail panel remains Praude’s existing markdown view. Layout uses the same width breakpoints and normalization as agent-deck to avoid panel bleed.

## Components
1) Grouped List Renderer
- GroupTree (status groups) with expand/collapse state.
- Flattened item list for cursor navigation.
- Render functions for group headers + PRD rows (icons, badges).
- Scroll indicators for long lists.

2) Search Overlay
- Port agent-deck’s `Search` overlay (bubbles/textinput).
- `/` opens; `esc` closes; `enter` selects; `up/down` navigates results.
- Filters PRD id/title.

3) Responsive Layout
- Single column list (<50 width).
- Stacked list + detail (50-79).
- Dual columns (80+).
- Use height/width normalization guards to prevent layout bleed.

4) UI State Persistence
- `.praude/state.json` stores:
  - expanded/collapsed status groups
  - last selected PRD id
- Load on startup, save on change. Failures are non-fatal.

5) Status Field
- Add `status` to PRD YAML schema and `specs.Spec`.
- `LoadSummaries` reads `status` for list grouping.
- Missing or invalid status defaults to `draft`.

## Data Flow
1) Load summaries from `.praude/specs/` including status.
2) Build group tree by status.
3) Flatten group tree into cursorable list.
4) Render list + detail per responsive layout.
5) Search overlay filters PRDs and updates selection.
6) Persist expansion + selection state to `.praude/state.json`.

## Error Handling
- Missing `.praude/` or invalid specs: show “Not initialized” and prompt to run `praude init`.
- Invalid or missing status: default to `draft`, add a validation warning.
- State file read/write errors: non-fatal; use defaults and show a status message.
- Search overlay handles empty results gracefully.

## Testing
- `internal/praude/specs/load_test.go`: parse `status`, default behavior.
- New `internal/praude/tui` tests:
  - GroupTree grouping and flatten order.
  - Expand/collapse behavior.
  - Search overlay filtering and selection.
  - Layout mode selection and width/height normalization.

## Decisions
- Full agent-deck list UX port (not partial).
- Status values: interview, draft, research, suggestions, validated, archived.
- Persist UI state to `.praude/state.json`.
- Layout breakpoints copied from agent-deck.

