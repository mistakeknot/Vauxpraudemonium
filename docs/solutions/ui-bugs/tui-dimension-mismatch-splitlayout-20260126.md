---
module: autarch/pkg/tui
date: 2026-01-26
problem_type: ui_bug
component: tui_layout
symptoms:
  - Stray vertical bars (│) appearing in split layouts
  - Misaligned borders and content overflow
  - Child views sized to full terminal width despite parent padding
  - Visual artifacts when rendering styled text in padded containers
root_cause: Views receiving raw terminal dimensions via tea.WindowSizeMsg but parent unified_app.View() wraps content with Padding(1, 3). This 6-char horizontal and 2-line vertical padding was not accounted for in child view sizing, causing content overflow.
severity: high
tags:
  - tui
  - lipgloss
  - bubble-tea
  - layout
  - visual-bug
  - dimension-mismatch
  - ansi-handling
---

# TUI Dimension Mismatch: Parent Padding vs Child Sizing

## Problem

Visual artifacts in TUI split layouts: stray vertical bars (`│`), misaligned borders, and content overflow.

## Root Cause

The `unified_app.View()` method wraps all child view content with `Padding(1, 3)`, which adds 2 vertical and 6 horizontal padding characters to the content area. However, child views (kickoff and interview) were receiving raw terminal dimensions via `tea.WindowSizeMsg`, causing them to render content assuming the full terminal width and height. This size mismatch resulted in visual artifacts: stray vertical bars, misaligned borders, and horizontal overflow.

## Solution

Modified the window size handling in both affected views to account for the unified app's content padding:

**BEFORE (incorrect):**
```go
case tea.WindowSizeMsg:
    v.width = msg.Width
    v.height = msg.Height - 4
```

**AFTER (correct):**
```go
case tea.WindowSizeMsg:
    // Account for unified_app's content padding (Padding(1, 3) = 6 horizontal, 2 vertical)
    v.width = msg.Width - 6
    v.height = msg.Height - 4 - 2
```

The fix subtracts 6 characters from width (left padding: 3, right padding: 3) and 2 characters from height (top padding: 1, bottom padding: 1), ensuring child views render content within the actual available space.

## Files Changed

- `internal/tui/views/kickoff.go` (lines 251-258)
- `internal/tui/views/interview.go` (lines 163-166)
- `pkg/tui/splitlayout.go` (uses ANSI-aware width calculation via `ansi.StringWidth()`)

## Prevention

### Detection - Catch Early
- Monitor for warning signs: horizontal overflow, truncated text, right pane extending beyond terminal edge, or background color bleeding
- Add dimension validation: `leftWidth + separator(1) + rightWidth ≤ availableWidth`
- Test at multiple terminal sizes (40w, 80w, 120w, 160w)

### Best Practices
- Always subtract parent padding BEFORE passing dimensions to children
- Document padding assumptions in SetSize() comments
- Never add `lipgloss.Style.Width()` constraints in child View() methods that differ from SetSize()

### Testing
- Create test matrix covering tiny (40w), small (80w), medium (120w), large (160w) terminals
- Visual regression: assert every line matches exact width using `ansi.StringWidth(line) == expected`

## Key Insight

**The One Rule:** Pass exact available space (after parent padding) to components.

**The Math:**
```
Available Width = Terminal Width - Parent Horizontal Padding
Left + Separator(1) + Right = Available Width  (MUST equal, never exceed)
```

## Related

- [docs/plans/2026-01-25-unified-autarch-tui-design.md](../../plans/2026-01-25-unified-autarch-tui-design.md) - Unified TUI design
- [docs/plans/2026-01-22-vauxhall-tui-two-pane-layout-design.md](../../plans/2026-01-22-vauxhall-tui-two-pane-layout-design.md) - Two-pane layout design

## Additional Note: ANSI Width Calculation

When padding/truncating styled text, use ANSI-aware width functions:

**Wrong** (counts escape codes as visible):
```go
import "github.com/mattn/go-runewidth"
width := runewidth.StringWidth(styledLine)  // Overcounts!
```

**Correct** (ignores escape codes):
```go
import "github.com/charmbracelet/x/ansi"
width := ansi.StringWidth(styledLine)  // Accurate
```
