date: 2026-01-28
topic: agent-model-selector-shortcut

# Agent Model Selector Shortcut

## What We're Building
Add a global way to switch the coding agent (Codex/Claude) in all Autarch TUIs that show the chat pane. The interaction should mirror Cursor’s agent panel: a small model selector under the chat pane, with both a command‑palette action and a dedicated shortcut. The desired dedicated shortcut is F2.

## Why This Approach
We need a fast, discoverable control that works across tools. A palette action covers discoverability and avoids key collisions, while a dedicated shortcut supports speed for frequent switching. Using the shared TUI shell keeps the behavior consistent across tools.

## Key Decisions
- Scope: all TUIs that include the chat pane (unified shell layout).
- Access: both command palette action (“Switch agent/model”) and dedicated shortcut.
- Dedicated shortcut: F2 (better terminal reliability).
- F2 toggles the selector open/closed, even when the composer is focused.
- Selector supports arrows + Enter and quick numeric picks (1/2), with Esc to close.
- Agent list sources: show auto-detected agents by default, plus configured targets.
- If both exist, dedupe by name and prefer configured entries.
- UX placement: small selector under the chat pane, consistent with existing composer hint/footers.

## Open Questions
- Terminal support: confirm F2 works in common terminals/tmux setups (expected to be reliable).
- Palette action: confirm command name and placement in any existing palette registry.
- Fallback: should we add an alternative binding (e.g., F2) if Ctrl+; is not detectable?
- Input mode: should Ctrl+; be active even when the composer has focus, or only when no text input is focused?

## Next Steps
→ Run `/workflows:plan` to define implementation steps and tests.
