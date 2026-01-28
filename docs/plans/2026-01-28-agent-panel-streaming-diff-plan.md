# Agent Panel Streaming + Diff + Settings Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** n/a (no bead provided)

**Goal:** Unify model/agent selection, add persistent chat settings, stream agent output in the chat panel, show live diffs in the main view during agent runs, and show an edit-log summary with one-click revert after completion.

**Architecture:** Introduce persistent chat settings loaded at app start and applied to all chat panels, replace the separate agent selector with a single “Model” control in the composer row, and add an agent-run pipeline that streams output to the chat panel while updating a diff view in the main document pane. After completion, emit a summary message and allow a one-click revert of the last run.

**Tech Stack:** Go, Bubble Tea, lipgloss, BurntSushi/toml

### Task 1: Add persistent chat settings (auto-scroll, show history on new chat, message grouping)

**Files:**
- Create: `internal/tui/chat_settings.go`
- Create: `internal/tui/chat_settings_test.go`
- Modify: `pkg/tui/chatpanel.go`

**Step 1: Write the failing test**

Create `internal/tui/chat_settings_test.go`:

```go
package tui

import (
    "os"
    "path/filepath"
    "testing"
)

func TestChatSettingsLoadDefaultsAndPersist(t *testing.T) {
    dir := t.TempDir()
    t.Setenv("XDG_CONFIG_HOME", dir)

    cfg, err := LoadChatSettings()
    if err != nil {
        t.Fatalf("load settings: %v", err)
    }
    if !cfg.AutoScroll || !cfg.ShowHistoryOnNewChat || !cfg.GroupMessages {
        t.Fatalf("expected defaults on")
    }

    cfg.AutoScroll = false
    cfg.GroupMessages = false
    if err := SaveChatSettings(cfg); err != nil {
        t.Fatalf("save settings: %v", err)
    }

    loaded, err := LoadChatSettings()
    if err != nil {
        t.Fatalf("reload settings: %v", err)
    }
    if loaded.AutoScroll || loaded.GroupMessages {
        t.Fatalf("expected persisted values")
    }

    // Ensure file exists
    path := filepath.Join(dir, "autarch", "ui.toml")
    if _, err := os.Stat(path); err != nil {
        t.Fatalf("expected settings file: %v", err)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui -run TestChatSettingsLoadDefaultsAndPersist -v`

Expected: FAIL (missing settings loader).

**Step 3: Write minimal implementation**

Create `internal/tui/chat_settings.go` with:
- `type ChatSettings struct { AutoScroll, ShowHistoryOnNewChat, GroupMessages bool }`
- `DefaultChatSettings()` returning all true
- `LoadChatSettings()` reading `~/.config/autarch/ui.toml` (respect XDG_CONFIG_HOME)
- `SaveChatSettings(settings ChatSettings)`

Update `pkg/tui/chatpanel.go`:
- Add `settings ChatSettings` (mirrored in pkg via small struct or setter)
- `SetSettings(ChatSettings)`
- `AddMessage` honors `AutoScroll`
- `ClearMessages` honors `ShowHistoryOnNewChat` in callers (leave Clear as is; caller decides)
- Implement `GroupMessages` in `renderHistory` (skip repeated role headers when enabled)

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui -run TestChatSettingsLoadDefaultsAndPersist -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tui/chat_settings.go internal/tui/chat_settings_test.go pkg/tui/chatpanel.go
git commit -m "feat(tui): add persistent chat settings"
```

### Task 2: Add chat settings UI (all views) and apply to chat panels

**Files:**
- Create: `pkg/tui/chat_settings_panel.go`
- Modify: `internal/tui/unified_app.go`
- Modify: `internal/tui/views/*` (where chat panel is used)
- Test: `internal/tui/unified_app_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/unified_app_test.go`:

```go
func TestChatSettingsTogglePersistsAndApplies(t *testing.T) {
    dir := t.TempDir()
    t.Setenv("XDG_CONFIG_HOME", dir)

    app := NewUnifiedApp(nil)
    app.chatSettings = DefaultChatSettings()

    // Toggle auto-scroll off
    app.chatSettings.AutoScroll = false
    if err := SaveChatSettings(app.chatSettings); err != nil {
        t.Fatalf("save settings: %v", err)
    }

    loaded, err := LoadChatSettings()
    if err != nil {
        t.Fatalf("reload settings: %v", err)
    }
    if loaded.AutoScroll {
        t.Fatalf("expected autos-scroll off")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui -run TestChatSettingsTogglePersistsAndApplies -v`

Expected: FAIL (missing fields/plumbing).

**Step 3: Write minimal implementation**

- Add a `ChatSettingsPanel` in `pkg/tui` that renders toggles and handles key input (j/k to move, space to toggle, esc to close).
- In `internal/tui/unified_app.go`:
  - Load settings on Init.
  - Add `chatSettings ChatSettings`, `chatSettingsOpen bool`.
  - Add palette command “Chat settings”.
  - Add key binding `,` or `ctrl+,` to open settings in all modes.
  - When settings change, call `SaveChatSettings` and broadcast to current view if it implements `SetChatSettings`.
- Add interface `ChatSettingsSetter` in `internal/tui/types.go` and implement it in views that render chat panels.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui -run TestChatSettingsTogglePersistsAndApplies -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/tui/chat_settings_panel.go internal/tui/unified_app.go internal/tui/types.go internal/tui/views/* internal/tui/unified_app_test.go
git commit -m "feat(tui): add chat settings panel and persistence"
```

### Task 3: Combine model + agent picker into a single composer-row control

**Files:**
- Modify: `pkg/tui/agent_selector.go`
- Modify: `pkg/tui/chatpanel.go`
- Modify: `internal/tui/unified_app.go`
- Test: `pkg/tui/agent_selector_test.go`

**Step 1: Write the failing test**

Extend `pkg/tui/agent_selector_test.go`:

```go
func TestAgentSelectorRendersAsModelControl(t *testing.T) {
    sel := NewAgentSelector([]AgentOption{{Name: "codex"}})
    view := sel.View()
    if !strings.Contains(view, "Model") {
        t.Fatalf("expected model label")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-cache go test ./pkg/tui -run TestAgentSelectorRendersAsModelControl -v`

Expected: FAIL.

**Step 3: Write minimal implementation**

- Update selector label to “Model” and render the current model in the composer row (e.g., `Model: Codex (F2)`), while keeping the drop-down list when open.
- Update `ChatPanel.View()` to render the selector in the composer row (not as a separate line) by passing the selector’s compact label into the composer title or hint.
- Update help text in `internal/tui/unified_app.go` from “agent selector” to “model selector”.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-cache go test ./pkg/tui -run TestAgentSelectorRendersAsModelControl -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/tui/agent_selector.go pkg/tui/chatpanel.go internal/tui/unified_app.go pkg/tui/agent_selector_test.go
git commit -m "feat(tui): unify model and agent picker"
```

### Task 4: Stream agent output to chat panel

**Files:**
- Modify: `internal/autarch/agent/epics.go`
- Modify: `internal/autarch/agent/tasks.go`
- Modify: `internal/tui/unified_app.go`
- Modify: `internal/tui/messages.go`
- Modify: `internal/tui/views/kickoff.go`
- Test: `internal/tui/unified_app_test.go`

**Step 1: Write the failing test**

Add to `internal/tui/unified_app_test.go`:

```go
func TestAgentStreamMessagesRouteToChat(t *testing.T) {
    app := NewUnifiedApp(nil)
    msg := AgentStreamMsg{Line: "hello"}
    _, _ = app.Update(msg)
    // Expect no panic and chat append via view setter (verified via view test helpers)
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui -run TestAgentStreamMessagesRouteToChat -v`

Expected: FAIL (message type missing).

**Step 3: Write minimal implementation**

- Add `AgentStreamMsg`, `AgentRunStartedMsg`, `AgentRunFinishedMsg`, and `AgentEditSummaryMsg` in `internal/tui/messages.go`.
- Add a `ChatStreamSetter` interface (e.g., `AppendChatLine(string)`). Implement on views with chat panels by forwarding to `ChatPanel.AddMessage("agent", line)`.
- Wire streaming callbacks in `GenerateEpics`/`GenerateTasks` to use `GenerateWithOutput` and send `AgentStreamMsg` into the UI loop.
- In `KickoffView`, send scan `AgentLine` into chat panel while scan is running.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui -run TestAgentStreamMessagesRouteToChat -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/autarch/agent/epics.go internal/autarch/agent/tasks.go internal/tui/unified_app.go internal/tui/messages.go internal/tui/views/kickoff.go internal/tui/unified_app_test.go
git commit -m "feat(tui): stream agent output to chat panel"
```

### Task 5: Live diff in main view + edit summary + one-click revert

**Files:**
- Create: `pkg/tui/diff.go`
- Modify: `internal/tui/unified_app.go`
- Modify: `internal/tui/views/spec_summary.go`
- Modify: `internal/tui/messages.go`
- Test: `pkg/tui/diff_test.go`

**Step 1: Write the failing test**

Create `pkg/tui/diff_test.go`:

```go
package tui

import "testing"

func TestUnifiedDiffFromStrings(t *testing.T) {
    diff, err := UnifiedDiff("before\n", "after\n", "a.md")
    if err != nil || len(diff) == 0 {
        t.Fatalf("expected diff output")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-cache go test ./pkg/tui -run TestUnifiedDiffFromStrings -v`

Expected: FAIL (helper missing).

**Step 3: Write minimal implementation**

- Implement `UnifiedDiff(before, after, label string) ([]string, error)` in `pkg/tui/diff.go` using temp files and `git diff --no-index --unified=3` (no new deps).
- Add diff state to `SpecSummaryView` (e.g., `showDiff bool`, `diffLines []string`) and render diff instead of normal document while `showDiff` is true.
- In `UnifiedApp`, when an agent run starts, capture the pre-run PRD content snapshot; when the run finishes, compute the diff and send `AgentRunStartedMsg` + `AgentRunFinishedMsg` to toggle diff view and then revert to normal view.
- Add a one-click `RevertLastRun` command (palette + key) that restores the pre-run snapshot and clears diff state.
- Emit `AgentEditSummaryMsg` to the chat panel after completion with a short edit log.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-cache go test ./pkg/tui -run TestUnifiedDiffFromStrings -v`

Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/tui/diff.go pkg/tui/diff_test.go internal/tui/unified_app.go internal/tui/messages.go internal/tui/views/spec_summary.go
git commit -m "feat(tui): show live diff and add one-click revert"
```
