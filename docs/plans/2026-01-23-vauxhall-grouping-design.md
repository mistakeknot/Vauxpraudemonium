# Vauxhall Project-Centric Grouping (TUI + Web) — Design

**Date:** 2026-01-23

## Goal
Add project-centric grouping for Sessions and Agents in both the TUI and Web UI. Grouping is derived directly from `ProjectPath` with no new persistence or config.

## Scope
- TUI: Sessions + Agents tabs grouped by project
- Web: Sessions + Agents pages grouped by project
- Filters apply within groups
- Collapse/expand groups in TUI only

## UX Summary

### TUI
- Right pane (Sessions/Agents) shows group headers + items beneath each project.
- Headers show project name + count. Default expanded.
- `g` toggles expand/collapse when selected item is a group header.
- Filter applies first; empty groups are hidden.
- “Unassigned” group for empty `ProjectPath`.

### Web
- Sessions/Agents pages display project headers + cards beneath each group.
- Search input remains; filtering reduces items and hides empty groups.
- No collapse/expand yet.

## Architecture

### TUI
- Build grouped item lists in `updateLists()`:
  - Filter sessions/agents first.
  - Group by `ProjectPath` into `map[path][]Item`.
  - Flatten into `[]list.Item` with `GroupHeaderItem` + children.
- Track expand state per tab + project path:
  - `groupExpanded map[string]bool` keyed as `sessions:<path>` / `agents:<path>`.
- Default expanded unless explicitly collapsed.

### Web
- In `handleSessions` / `handleAgents`:
  - Filter then group into `[]Group` (Project + Items).
  - Sort groups by project basename (Unassigned last).
  - Render header + cards in templates.

## Data Flow

1. **Aggregator** provides `State.Sessions` and `State.Agents`.
2. **Filters** reduce those lists.
3. **Grouping**: map by `ProjectPath`.
4. **Render**:
   - TUI: flattened list w/ headers
   - Web: grouped list w/ header row per group

## Error Handling
- Missing `ProjectPath` → “Unassigned” group.
- Empty groups are skipped.

## Testing

### TUI
- Group headers + items present in correct order.
- Collapse hides children while keeping header.
- Filtered grouping hides empty groups.
- Unassigned group appears for empty `ProjectPath`.
- `g` toggles group state only when header selected.

### Web
- Response includes group headers and only matching items.
- Filtered results hide empty groups.
- Unassigned group rendered for empty `ProjectPath`.

## Non‑Goals
- Manual group creation or persistence.
- Web collapse/expand.
- CLI group management (future work).
