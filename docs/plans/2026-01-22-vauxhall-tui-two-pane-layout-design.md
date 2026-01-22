# Vauxhall TUI Two-Pane Layout - Design

Date: 2026-01-22
Owner: Vauxhall
Status: Draft (approved in chat)

## Summary

Refactor the Vauxhall TUI into a two-pane layout with Projects pinned left and the current view (Dashboard, Sessions, Agents) on the right. Sessions and agents are filtered by the selected project, with an "All Projects" item for global scope. This improves project linkage and aligns sessions/agents relative to projects.

## Goals

- Projects always visible and selectable on the left.
- Sessions and agents filtered by selected project.
- Dashboard available in right pane (project-scoped or global).
- Maintain keyboard-first workflow without regressions.

## Non-Goals

- Web UI changes in this phase.
- Deep session-agent linking beyond project path matching.
- TUI visual redesign beyond layout and filtering.

## Layout & Navigation

- Two columns: **Projects (left)**, **Main view (right)**.
- Right view tabs: **Dashboard / Sessions / Agents**.
- Projects tab removed; projects are always visible on the left.
- Add synthetic "All Projects" entry at top (disables filtering).

### Focus + Keying

- Default focus: right pane list.
- `[` / `]` (or Ctrl+Left/Right) toggles focus between panes.
- `/` filters the active pane list.
- Sessions actions (n/r/k/f/a) apply in Sessions view on right.

## Data Flow & Filtering

- Selected project path from left list drives filtering:
  - Sessions: `ProjectPath == selectedPath`
  - Agents: `ProjectPath == selectedPath`
  - Dashboard: project-scoped summary when selectedPath set, otherwise global
- If "All Projects" selected, show unfiltered lists.

## Error Handling & Edge Cases

- No projects: show empty left list and global right view.
- Project selected with no sessions/agents: show empty state with project name.
- Sessions without project path appear only under "All Projects".
- Guard against narrow terminal widths; clamp column widths and padding.
- If terminal too narrow, fallback to single column with a compact project selector line.

## Testing

- Unit tests for:
  - Filtering logic (project selection affects sessions/agents)
  - Focus switching and filter routing
  - Layout width clamping (no panics)

## Open Questions

- Should Agents view show "unlinked" agents (no project path) when a project is selected?
- Should Dashboard view include quick project header for clarity?

