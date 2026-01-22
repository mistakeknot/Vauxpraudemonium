# Vauxhall M1B Agent Deck Parity - Design

Date: 2026-01-22
Owner: Vauxhall
Status: Draft (approved in chat)

## Summary

Deliver "Agent Deck parity" for Vauxhall in two phases. Phase 1 adds tmux session control (new/rename/restart/fork/attach), web+TUI parity for search/group/actions, and a repo-only MCP manager (start/stop/status + logs). Phase 2 adds tmux status-bar notifier (opt-in with backups) and MCP socket pooling. Control actions are enabled by default.

## Goals

- Achieve functional parity with Agent Deck for session control and MCP management.
- Keep the system deterministic and safe with clear error messaging.
- Share one control path for web and TUI via the aggregator.
- Preserve Vauxhall's mission-control focus while enabling action commands.

## Non-Goals

- Managing external MCPs used by agents in Phase 1.
- Automatic tmux config edits without explicit user action.
- Deep tmux session cloning (duplicate all windows/panes).

## Phase Scope

### Phase 1 (Parity Core)
- tmux control actions: new, rename, restart (kill+recreate), fork (new agent session in same project), attach (command output).
- Agent command resolution: config (Vauxhall + Praude) then fallback defaults (`claude`, `codex`).
- MCP manager: repo MCPs only (`mcp-server/`, `mcp-client/`) with start/stop/status/log tail.
- Web + TUI parity: search, group, status, action menu, MCP panel.

### Phase 2 (Parity Extended)
- tmux status-bar notifier (opt-in apply; config backup).
- MCP socket pooling + auto-restart hooks.
- Optional duplicate tmux session action (if still desired).

## Architecture

- **tmux Client** (`internal/vauxhall/tmux`): extend to support session lifecycle actions.
- **MCP Manager** (`internal/vauxhall/mcp`): lightweight supervisor for repo MCP processes.
- **Aggregator** (`internal/vauxhall/aggregator`): single action entrypoint for web + TUI.
- **Web + TUI**: thin UI layers calling aggregator actions; consistent state model.

### Control Plane Separation
- Observation remains read-only and cached.
- Actions execute through explicit methods with structured errors and logging.

## Data Model Additions

Extend tmux session model with action metadata:

- `Command` (resolved agent command used for new/restart/fork)
- `ProjectPath` (matched to discovery)
- `AgentType` (claude/codex/unknown)

MCP manager structures:

- `Component` (`server` | `client`)
- `Status` (`running` | `stopped` | `error`)
- `Pid`, `StartTime`, `LastError`, `LogTail`

## tmux Actions

- **New**: create a new tmux session with cwd and agent command.
- **Rename**: rename session; error on collision.
- **Restart**: kill session, recreate with same project + command.
- **Fork**: create new agent session in same project using resolved command.
- **Attach**: return attach command in web; execute attach in TUI with warning.

Notes:
- If agent type is unknown, prompt for selection in UI.
- If tmux isn't running, return user-friendly error.

## MCP Manager (Phase 1)

- Detect repo components by presence of `mcp-server/` and `mcp-client/`.
- Resolve command priority:
  1) `~/.config/vauxhall/config.toml`
  2) project config (e.g., `.praude/config.toml` if used)
  3) defaults (`npm run dev`, or explicit script if defined)
- Idempotent actions (start returns running if already running).
- Capture last 50 lines of stdout/stderr per component.

## UI Parity

### Web
- Sessions table with status pills, search, group by project/agent.
- Row actions: new, rename, restart, fork, attach.
- MCP control panel per project with toggles and log drawer.

### TUI
- Keyboard-first actions:
  - `/` search
  - `g` group toggle
  - `a` attach
  - `n` new
  - `r` rename
  - `k` restart
  - `f` fork
  - `m` MCP panel
  - `space` toggle component

## Error Handling & Safety

- Action errors surfaced as toasts/alerts (web) and status lines (TUI).
- Restart confirms if session is attached.
- Rename collisions and missing tmux show clear messages.
- MCP failures include command + error output.

## tmux Status-Bar Notifier (Phase 2)

- Opt-in apply from Vauxhall UI.
- Backup original config to `~/.config/vauxhall/backups/tmux.conf.<timestamp>` (or `.tmux.conf.<timestamp>`).
- Generate a status-right snippet indicating waiting sessions.

## Testing

- Unit tests for tmux parsing and action command composition (injectable exec runner).
- MCP manager tests for start/stop idempotency, PID tracking, log tailing.
- Web handler tests for action endpoints and error mapping.
- TUI model tests for key bindings and state transitions.
- Manual checklist for end-to-end verification.

## Open Questions

- Do we want a dedicated "duplicate tmux session" action in Phase 2?
- Should attach in TUI be disabled when running in headless mode?
- Should MCP manager support pnpm/yarn auto-detect in Phase 1?

