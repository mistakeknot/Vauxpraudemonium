# Tandemonium TUI Vauxhall Style Port Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** `Autarch-78w (Task reference)`

**Goal:** Port Tandemonium’s TUI rendering to use Vauxhall’s lipgloss-based styling while keeping the current two-pane layout and behaviors intact.

**Architecture:** Introduce Tandemonium-specific style helpers that wrap the shared Tokyo Night palette, then refactor the render path in `internal/tandemonium/tui/model.go` to use lipgloss panels, tabs, and footer (matching Vauxhall). Replace ANSI `colorize` usage with lipgloss styles and update tests to assert the new header/tab/footer structure.

**Tech Stack:** Go, Bubble Tea, Lipgloss (shared styles in `pkg/tui`).

**Worktree:** User requested no worktree; execute on current branch.

---

### Task 1: Shared style adapter for Tandemonium

**Files:**
- Create: `internal/tandemonium/tui/styles.go`
- Modify: `internal/tandemonium/tui/model_test.go`

**Step 1: Write failing test (header uses TitleStyle)**

```go
func TestHeaderUsesTitleStyle(t *testing.T) {
	m := NewModel()
	m.Width = 120
	out := m.View()
	if !strings.Contains(stripANSI(out), "Tandemonium") {
		t.Fatalf("expected title in header")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tandemonium/tui -run TestHeaderUsesTitleStyle`

Expected: FAIL (no header rendering yet).

**Step 3: Implement minimal styles adapter**

```go
// styles.go
package tui

import shared "github.com/mistakeknot/autarch/pkg/tui"

var (
	BaseStyle = shared.BaseStyle
	PanelStyle = shared.PanelStyle
	PaneFocusedStyle = shared.PaneFocusedStyle
	PaneUnfocusedStyle = shared.PaneUnfocusedStyle
	TitleStyle = shared.TitleStyle
	SubtitleStyle = shared.SubtitleStyle
	LabelStyle = shared.LabelStyle
	SelectedStyle = shared.SelectedStyle
	UnselectedStyle = shared.UnselectedStyle
	HelpKeyStyle = shared.HelpKeyStyle
	HelpDescStyle = shared.HelpDescStyle
	TabStyle = shared.TabStyle
	ActiveTabStyle = shared.ActiveTabStyle
)
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tandemonium/tui -run TestHeaderUsesTitleStyle`

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tandemonium/tui/styles.go internal/tandemonium/tui/model_test.go
git commit -m "feat(tandemonium): add vauxhall style adapter"
```

---

### Task 2: Header + tab bar + footer lipgloss render

**Files:**
- Modify: `internal/tandemonium/tui/model.go`
- Test: `internal/tandemonium/tui/view_test.go`

**Step 1: Write failing test (tab bar shows Fleet/Review)**

```go
func TestTabBarRenders(t *testing.T) {
	m := NewModel()
	m.Width = 120
	out := m.View()
	clean := stripANSI(out)
	if !strings.Contains(clean, "Fleet") || !strings.Contains(clean, "Review") {
		t.Fatalf("expected tab bar")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tandemonium/tui -run TestTabBarRenders`

Expected: FAIL.

**Step 3: Implement lipgloss header/tab/footer**

```go
// model.go (inside View/render)
header := TitleStyle.Render("Tandemonium")
if m.FilterMode != "" {
	header += " " + LabelStyle.Render("("+m.FilterMode+")")
}

tabs := []string{renderTab("Fleet", m.ViewMode == ViewFleet), renderTab("Review", m.ViewMode == ViewReview)}
if m.RightTab == RightTabCoord {
	tabs = append(tabs, renderTab("Coord", true))
}

tabBar := lipgloss.JoinHorizontal(lipgloss.Center, tabs...)
footer := HelpKeyStyle.Render("n") + HelpDescStyle.Render(" new • ") +
	HelpKeyStyle.Render("/") + HelpDescStyle.Render(" search • ") +
	HelpKeyStyle.Render("?") + HelpDescStyle.Render(" help")
```

Ensure tabs use `TabStyle` / `ActiveTabStyle`, and footer uses shared help styles.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tandemonium/tui -run TestTabBarRenders`

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tandemonium/tui/model.go internal/tandemonium/tui/view_test.go
git commit -m "feat(tandemonium): add vauxhall header tabs footer"
```

---

### Task 3: Two-pane rendering with lipgloss panels

**Files:**
- Modify: `internal/tandemonium/tui/model.go`
- Test: `internal/tandemonium/tui/view_test.go`

**Step 1: Write failing test (panes render with borders)**

```go
func TestTwoPaneLayoutRenders(t *testing.T) {
	m := NewModel()
	m.Width = 120
	m.Height = 40
	out := m.View()
	clean := stripANSI(out)
	if !strings.Contains(clean, "│") && !strings.Contains(clean, "┐") {
		t.Fatalf("expected pane borders")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tandemonium/tui -run TestTwoPaneLayoutRenders`

Expected: FAIL.

**Step 3: Implement lipgloss panels for left/right panes**

```go
leftStyle := PaneUnfocusedStyle
rightStyle := PaneUnfocusedStyle
if m.FocusPane == FocusTasks { leftStyle = PaneFocusedStyle } else { rightStyle = PaneFocusedStyle }

leftView := leftStyle.Width(leftW).Render(left)
rightView := rightStyle.Width(rightW).Render(right)

body := lipgloss.JoinHorizontal(lipgloss.Top, leftView, "  ", rightView)
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tandemonium/tui -run TestTwoPaneLayoutRenders`

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tandemonium/tui/model.go internal/tandemonium/tui/view_test.go
git commit -m "feat(tandemonium): render panes with lipgloss styles"
```

---

### Task 4: Replace ANSI color helpers with lipgloss styles

**Files:**
- Modify: `internal/tandemonium/tui/model.go`
- Modify: `internal/tandemonium/tui/status_test.go`

**Step 1: Write failing test (status badges use styled labels)**

```go
func TestStatusBadgeUsesStyledLabel(t *testing.T) {
	if !strings.Contains(stripANSI(statusBadge("running")), "RUN") {
		t.Fatalf("expected running badge")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tandemonium/tui -run TestStatusBadgeUsesStyledLabel`

Expected: FAIL if statusBadge still uses raw ANSI.

**Step 3: Replace colorize + ANSI helpers**

```go
func statusBadge(status string) string {
	switch status {
	case "running":
		return StatusRunning.Render("[RUN]")
	case "review":
		return StatusWaiting.Render("[REV]")
	case "blocked":
		return StatusError.Render("[BLK]")
	case "done":
		return StatusRunning.Render("[DONE]")
	default:
		return StatusIdle.Render("[TODO]")
	}
}
```

Replace `colorize()` usage with `SelectedStyle`/`UnselectedStyle` when highlighting list rows.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tandemonium/tui -run TestStatusBadgeUsesStyledLabel`

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tandemonium/tui/model.go internal/tandemonium/tui/status_test.go
git commit -m "refactor(tandemonium): replace ansi color helpers"
```

---

### Task 5: Full verification

**Files:**
- None

**Step 1: Run full test suite**

Run: `go test ./...`

Expected: PASS.

---

Plan complete and saved to `docs/plans/2026-01-23-tandemonium-tui-vauxhall-port-plan.md`.

Two execution options:

1. Subagent-Driven (this session) — I dispatch a fresh subagent per task, review between tasks
2. Parallel Session (separate) — Open a new session with executing-plans and batch execution

Which approach?
