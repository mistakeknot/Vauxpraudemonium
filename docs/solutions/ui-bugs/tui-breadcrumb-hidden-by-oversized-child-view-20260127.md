---
title: "TUI Breadcrumb Header Hidden by Oversized Child View"
category: ui-bugs
tags: [bubbletea, lipgloss, tui, layout, window-size, breadcrumb]
module: internal/tui
symptom: "Breadcrumb navigation not visible on kickoff screen"
root_cause: "Child views received full terminal height via WindowSizeMsg, rendering content that pushed header off-screen"
date: 2026-01-27
---

# TUI Breadcrumb Header Hidden by Oversized Child View

## Problem

When running `./dev autarch tui`, the breadcrumb navigation bar (`Project › Interview › Spec › Epics › Tasks › Dashboard`) was invisible on the kickoff (first) screen. The footer rendered correctly, but the header area appeared as blank dark lines.

## Symptoms

- Header area rendered with correct background color (`ColorBgDark`) but no text content
- Footer showed `ctrl+b jump` (breadcrumb shortcut) proving onboarding mode was active
- `tmux capture-pane` showed two blank lines at top, then content starting immediately
- Unit tests proved `Breadcrumb.View()` returned correct text — the component itself was fine

## Investigation

1. Verified `a.mode == ModeOnboarding` at startup (mode=0, correct)
2. Verified `Breadcrumb.View()` renders text via unit test (76 chars, all labels present)
3. Verified `HeaderStyle.Render(breadcrumb.View())` produces 3-line output with text on line 2
4. Added debug file write to `View()` — confirmed it was called with correct dimensions
5. **Key discovery**: `tmux capture-pane` showed only 37 lines for a 40-line terminal — the output was being clipped

## Root Cause

In `unified_app.go`, the `WindowSizeMsg` handler passed the **full terminal dimensions** to child views:

```go
// BEFORE (broken)
case tea.WindowSizeMsg:
    a.width = msg.Width
    a.height = msg.Height
    // ...
    if a.currentView != nil {
        a.currentView, cmd = a.currentView.Update(msg) // Full height!
        return a, cmd
    }
```

The `View()` method then allocated space for header (3 lines) + content + footer (3 lines):

```go
contentHeight := a.height - headerHeight - footerHeight
```

But the child view (KickoffView) had already sized itself to the **full** terminal height via `splitLayout.SetSize(v.width, v.height)`. When lipgloss joined header + oversized content + footer vertically, the total exceeded terminal height. Bubbletea's alt-screen clipped from the **top**, hiding the header/breadcrumb.

## Solution

Pass a reduced `WindowSizeMsg` to child views, subtracting header and footer height:

```go
// AFTER (fixed) — unified_app.go:148-159
case tea.WindowSizeMsg:
    a.width = msg.Width
    a.height = msg.Height
    // ...
    if a.currentView != nil {
        headerHeight := 3
        footerHeight := 3
        contentMsg := tea.WindowSizeMsg{
            Width:  msg.Width,
            Height: msg.Height - headerHeight - footerHeight,
        }
        a.currentView, cmd = a.currentView.Update(contentMsg)
        return a, cmd
    }
```

## Prevention

**Pattern**: In any Bubble Tea app with chrome (header, footer, sidebar), always subtract chrome dimensions from `WindowSizeMsg` before passing to child models. The parent owns the layout; children should only know about their allocated space.

**Checklist for new layout containers**:
- [ ] Does the container add chrome (header, footer, borders)?
- [ ] Does it subtract chrome dimensions from `WindowSizeMsg` before passing to children?
- [ ] Does `View()` use the same height arithmetic as the `WindowSizeMsg` reduction?

## Related

- `docs/solutions/ui-bugs/tui-dimension-mismatch-splitlayout-20260126.md` — similar layout sizing issue
- `internal/tui/unified_app.go:148-159` — the fix location
- `internal/tui/views/kickoff.go:255` — child view that consumed full height
