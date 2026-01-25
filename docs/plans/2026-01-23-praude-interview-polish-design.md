# Praude Interview UI Polish Design

**Bead:** `Autarch-rte`

## Goal
Make the interview UI feel like a modern agent TUI by polishing the composer, chat transcript, and header nav without changing core behavior.

## Scope
- Composer panel redesign (title bar, bordered input, compact status line)
- Chat transcript styling (role badges, spacing, light dividers)
- Header nav styling (pill-like steps, active emphasis, responsive collapse)

Out of scope: new workflows, new steps, data model changes.

---

## Composer Redesign
The composer is the primary interaction surface and should read like a dedicated control panel:

- **Title bar:** `Compose · <Step>` (e.g., `Compose · Vision`) with a subtle style distinct from the transcript.
- **Input box:** fixed 6-line input area with ASCII-safe border; add a single inner margin so text doesn’t touch borders.
- **Status line:** compact hints and cursor location on a single line:
  - `Enter: iterate · [ / ]: prev/next · Ctrl+O: open · \: swap · (line X, col Y)`
- **Layout:** composer is bottom-anchored in the chat panel; transcript scrolls above it.

**Rationale:** The composer should feel intentional and “full width,” with predictable height and clear affordances.

---

## Chat Transcript Styling
The transcript should feel conversational and scannable without heavy chrome:

- **Role badges:** prefix each message with `[User]` or `[Agent]` on its own line.
- **Message body:** one-line indent for the message text to create visual rhythm.
- **Spacing:** one blank line between messages; optional light divider like `· · ·` between turns.
- **Header:** show `PM-focused agent: Codex CLI / Claude Code` at the top of transcript area.
- **Empty state:** muted “No messages yet.”
- **Wrap:** message text should wrap to panel width to avoid overflow.

**Rationale:** Light structure improves scanability without overwhelming the screen.

---

## Header Nav Styling
The header nav should read as a step navigator:

- **Pills:** ASCII-safe pill labels, e.g., `[Scan]  [Confirm]  [Bootstrap]  [Vision] ...`.
- **Active step:** emphasized with stronger style or doubled brackets `[[Vision]]`.
- **Spacing:** consistent double-space between pills.
- **Responsive collapse:** when width is tight, show the active step plus neighbors with an ellipsis, e.g., `... [Users] [[Problem]] [Requirements] ...`.
- **Divider:** optional subtle line below header to separate nav from panels.

**Rationale:** Always-visible orientation without wrapping noise.

---

## Implementation Notes
- Centralize styles in `internal/praude/tui/styles.go` to avoid inline color usage.
- Keep all output ASCII-safe; no ANSI escape sequences inside markdown outputs.
- Ensure transcript rendering remains fast; avoid per-character styling.

## Success Criteria
- Composer looks deliberate and stable across terminal sizes.
- Transcript is readable with clear turn structure.
- Header nav communicates current step at a glance.
