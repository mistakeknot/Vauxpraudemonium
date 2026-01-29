date: 2026-01-29
topic: unified-tui-shortcuts

# Unified TUI Shortcuts

## What We're Building
A unified shortcut policy for the global three‑pane UI that eliminates conflicts with text input. The goal is to make keyboard behavior predictable across all tools and views, especially when the chat input is focused. This focuses on standardizing shortcut types (non‑printable/modifier keys only) and clarifying pane focus behavior.

## Why This Approach
We explored how current shortcuts collide with text input and how Tab focus is intercepted by dashboard tabs. Printable keys (letters, digits, symbols) create accidental triggers while typing, and multiple panes compete for the same keys. The simplest way to eliminate ambiguity is to forbid printable shortcuts entirely in the unified UI, including view‑specific actions, and use only non‑printable or modifier combos.

## Key Decisions
- **No printable shortcuts anywhere in the unified UI.** All actions must use non‑printable keys (function keys, arrows) or modifier combos (Ctrl/Alt). This removes typing collisions.
- **Pane focus stays on Tab/Shift+Tab.** Tab navigation is reserved for pane focus; tab switching moves to modifier keys.
- **Dashboard tab switching uses Ctrl+Left/Ctrl+Right.** Avoids conflict with typing and preserves Tab focus behavior.

## Open Questions
- Which non‑printable or modifier shortcuts should replace common actions (help, refresh, search, back, quit) in each view?
- Should we provide secondary fallbacks for terminals that do not emit Ctrl+Left/Ctrl+Right?
- How should help overlays present the new shortcut scheme for discoverability?

## Next Steps
→ `/workflows:plan` to define the concrete keymap changes and affected files.
