# Hide System Labels in Chat Panel Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** n/a (no bead provided)

**Goal:** Hide repeated "System:" headings in the chat panel while keeping message content intact.

**Architecture:** Update chat history rendering to omit role headers for system messages, and add a focused unit test that asserts system messages render without the role label.

**Tech Stack:** Go, Bubble Tea, lipgloss

### Task 1: Suppress system role headers and add a test

**Files:**
- Modify: `pkg/tui/chatpanel.go`
- Create: `pkg/tui/chatpanel_test.go`

**Step 1: Write the failing test**

Create `pkg/tui/chatpanel_test.go`:

```go
package tui

import (
    "strings"
    "testing"
)

func TestChatPanelHidesSystemRoleLabel(t *testing.T) {
    panel := NewChatPanel()
    panel.SetSize(60, 20)
    panel.AddMessage("system", "Welcome")

    view := panel.View()
    if strings.Contains(view, "System:") {
        t.Fatalf("expected System label to be hidden, got %q", view)
    }
    if !strings.Contains(view, "Welcome") {
        t.Fatalf("expected system content to be rendered")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-cache go test ./pkg/tui -run TestChatPanelHidesSystemRoleLabel -v`

Expected: FAIL (System label still present).

**Step 3: Write minimal implementation**

In `pkg/tui/chatpanel.go`, adjust `renderHistory` to skip role headers when `msg.Role` is `"system"` (case-insensitive), but still render content.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-cache go test ./pkg/tui -run TestChatPanelHidesSystemRoleLabel -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/tui/chatpanel.go pkg/tui/chatpanel_test.go
git commit -m "feat(tui): hide system role labels in chat panel"
```
