date: 2026-01-28
topic: agent-panel-streaming-diff

# Agent Panel Streaming + Diff UX

## What We're Building
We’re aligning Autarch’s agent panel UX with Cursor/Copilot patterns: the chat panel streams agent output, while the main (left) view shows live unified diffs during agent runs and then reverts to the normal document view on completion. After completion, the right panel shows a concise edit-log summary. Users can continue chatting to iterate, and a one-click revert rolls back the last run (PRD flow is a single file/sections).

## Why This Approach
Streaming in the chat panel keeps the user anchored in a conversational workflow, while a live diff in the main view makes edits visible and reviewable without cluttering the chat. Reverting on completion restores the normal document view, aligning with the "live diff → review summary" model used by Cursor.

## Key Decisions
- Live stream goes to the chat panel; the main view shows live unified diff only while the agent runs.
- After completion, the main view reverts to the normal document view and the right panel shows an edit-log summary.
- Revert is one-click and rolls back the last run (no per-file revert for PRD flow).
- Chat settings toggles (auto-scroll, show history on new chat, message grouping) persist across sessions and apply to all views.
- Combine model + agent picker into a single "current model" control in the composer row.

## Open Questions
- Exact diff source for PRD flow (content snapshot vs file-based diff) and where the last-run snapshot should be stored.

## Next Steps
→ /workflows:plan for implementation details.
