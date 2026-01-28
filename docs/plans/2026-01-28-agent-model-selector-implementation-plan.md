# Agent Model Selector TUI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** `[bead-id] (Task reference)` — mandatory line tying the plan to the active bead/Task Master item.

**Goal:** Add a global agent/model selector under chat panes, toggled by F2 and palette, that switches the active coding agent (Codex/Claude) and stays consistent across TUIs.

**Architecture:** Introduce a shared `AgentSelector` component in `pkg/tui` that renders under chat panes and handles F2/arrow/number selection. Build available agent options by merging configured agent targets with auto-detected binaries (dedupe by name, prefer config). Wire selection changes into the unified app to update `codingAgent` and propagate the selection to all chat-pane views via a shared selector instance.

**Tech Stack:** Go, Bubble Tea, lipgloss, `pkg/agenttargets` registry, `internal/autarch/agent`.

---

### Task 1: Agent options merge helper (dedupe + precedence)

**Files:**
- Create: `pkg/agenttargets/detect.go`
- Test: `pkg/agenttargets/detect_test.go`

**Step 1: Write the failing test**

```go
func TestMergeDetectedPrefersConfig(t *testing.T) {
	detected := Registry{Targets: map[string]Target{
		"codex": {Name: "codex", Type: TargetDetected, Command: "codex"},
		"claude": {Name: "claude", Type: TargetDetected, Command: "claude"},
	}}
	global := Registry{Targets: map[string]Target{
		"codex": {Name: "codex", Type: TargetCommand, Command: "/bin/codex"},
	}}
	project := Registry{Targets: map[string]Target{
		"claude": {Name: "claude", Type: TargetCommand, Command: "/bin/claude"},
	}}

	merged := MergeDetected(detected, global, project)

	if merged.Targets["codex"].Command != "/bin/codex" {
		t.Fatalf("expected codex from global, got %q", merged.Targets["codex"].Command)
	}
	if merged.Targets["claude"].Command != "/bin/claude" {
		t.Fatalf("expected claude from project, got %q", merged.Targets["claude"].Command)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/agenttargets -run TestMergeDetectedPrefersConfig -v`
Expected: FAIL (MergeDetected missing)

**Step 3: Write minimal implementation**

```go
// DetectAvailableTargets returns detected targets using lookPath.
func DetectAvailableTargets(lookPath func(string) (string, error)) Registry {
	reg := Registry{Targets: map[string]Target{}}
	for name, target := range DefaultDetectedRegistry().Targets {
		if _, err := lookPath(target.Command); err == nil {
			reg.Targets[name] = target
		}
	}
	return reg
}

// MergeDetected merges detected → global → project (later overrides earlier).
func MergeDetected(detected, global, project Registry) Registry {
	merged := Merge(detected, global)
	merged = Merge(merged, project)
	return merged
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/agenttargets -run TestMergeDetectedPrefersConfig -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/agenttargets/detect.go pkg/agenttargets/detect_test.go
git commit -m "feat(agenttargets): add detected merge helpers"
```

---

### Task 2: Agent selector component (TUI)

**Files:**
- Create: `pkg/tui/agent_selector.go`
- Test: `pkg/tui/agent_selector_test.go`

**Step 1: Write the failing tests**

```go
func TestAgentSelectorToggleAndSelect(t *testing.T) {
	s := NewAgentSelector([]AgentOption{{Name: "codex"}, {Name: "claude"}})

	// F2 opens
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyF2})
	if !s.Open {
		t.Fatal("expected selector open after F2")
	}

	// Down + enter selects second option
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyDown})
	msg, _ := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if sel, ok := msg.(AgentSelectedMsg); !ok || sel.Name != "claude" {
		t.Fatalf("expected selection of claude, got %#v", msg)
	}
	if s.Open {
		t.Fatal("expected selector closed after selection")
	}
}

func TestAgentSelectorQuickPick(t *testing.T) {
	s := NewAgentSelector([]AgentOption{{Name: "codex"}, {Name: "claude"}})
	_, _ = s.Update(tea.KeyMsg{Type: tea.KeyF2})

	msg, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if sel, ok := msg.(AgentSelectedMsg); !ok || sel.Name != "claude" {
		t.Fatalf("expected selection of claude, got %#v", msg)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./pkg/tui -run TestAgentSelector -v`
Expected: FAIL (AgentSelector missing)

**Step 3: Write minimal implementation**

```go
type AgentOption struct {
	Name   string
	Source string
}

type AgentSelectedMsg struct {
	Name string
}

type AgentSelector struct {
	Options []AgentOption
	Open    bool
	Index   int
}

func NewAgentSelector(opts []AgentOption) *AgentSelector {
	return &AgentSelector{Options: opts}
}

func (s *AgentSelector) Update(msg tea.KeyMsg) (tea.Msg, tea.Cmd) {
	switch msg.Type {
	case tea.KeyF2:
		s.Open = !s.Open
		return nil, nil
	}
	if !s.Open {
		return nil, nil
	}

	switch msg.String() {
	case "esc":
		s.Open = false
	case "up":
		if s.Index > 0 { s.Index-- }
	case "down":
		if s.Index < len(s.Options)-1 { s.Index++ }
	case "enter":
		if len(s.Options) > 0 {
			opt := s.Options[s.Index]
			s.Open = false
			return AgentSelectedMsg{Name: opt.Name}, nil
		}
	case "1", "2":
		idx := int(msg.String()[0]-'1')
		if idx >= 0 && idx < len(s.Options) {
			opt := s.Options[idx]
			s.Open = false
			return AgentSelectedMsg{Name: opt.Name}, nil
		}
	}
	return nil, nil
}

func (s *AgentSelector) View() string {
	if len(s.Options) == 0 {
		return ""
	}
	if !s.Open {
		return lipgloss.NewStyle().Foreground(ColorMuted).Render("F2: agent")
	}
	var parts []string
	for i, opt := range s.Options {
		label := fmt.Sprintf("[%d] %s", i+1, opt.Name)
		if i == s.Index {
			label = SelectedStyle.Render(label)
		}
		parts = append(parts, label)
	}
	return lipgloss.NewStyle().Foreground(ColorFgDim).Render(\"Agent: \") + strings.Join(parts, \"  \")
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./pkg/tui -run TestAgentSelector -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/tui/agent_selector.go pkg/tui/agent_selector_test.go
git commit -m "feat(tui): add agent selector component"
```

---

### Task 3: Build agent options from config + detected

**Files:**
- Create: `internal/tui/agent_options.go`
- Test: `internal/tui/agent_options_test.go`

**Step 1: Write the failing test**

```go
func TestBuildAgentOptionsDedupePrefersConfig(t *testing.T) {
	options := buildAgentOptionsFromRegistry(
		agenttargets.Registry{Targets: map[string]agenttargets.Target{"codex": {Name: "codex", Command: "/bin/codex"}}},
	)
	if options[0].Name != "codex" {
		t.Fatalf("expected codex option, got %v", options)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui -run TestBuildAgentOptionsDedupePrefersConfig -v`
Expected: FAIL (helper missing)

**Step 3: Write minimal implementation**

```go
func LoadAgentOptions(projectRoot string) ([]pkgtui.AgentOption, error) {
	configDir, _ := os.UserConfigDir()
	globalPath := filepath.Join(configDir, "autarch", "agents.toml")
	global, project, err := agenttargets.Load(globalPath, projectRoot)
	if err != nil {
		return nil, err
	}
	detected := agenttargets.DetectAvailableTargets(exec.LookPath)
	merged := agenttargets.MergeDetected(detected, global, project)
	return buildAgentOptionsFromRegistry(merged), nil
}
```

Add a small helper that sorts options (e.g., codex/claude first, then alpha).

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui -run TestBuildAgentOptionsDedupePrefersConfig -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/agent_options.go internal/tui/agent_options_test.go
git commit -m "feat(tui): add agent option loader"
```

---

### Task 4: Wire selector into chat panes

**Files:**
- Modify: `pkg/tui/chatpanel.go`
- Modify: `internal/tui/views/kickoff.go`
- Modify: `internal/gurgeh/arbiter/tui/arbiter_view.go`
- Modify: `internal/tui/views/spec_summary.go`
- Modify: `internal/tui/views/epic_review.go`
- Modify: `internal/tui/views/task_review.go`
- Modify: `internal/tui/views/task_detail.go`
- Modify: `internal/tui/views/bigend.go`
- Modify: `internal/tui/views/coldwine.go`
- Modify: `internal/tui/views/gurgeh.go`
- Modify: `internal/tui/views/pollard.go`

**Step 1: Write a failing chatpanel render test**

```go
func TestChatPanelRendersSelector(t *testing.T) {
	panel := NewChatPanel()
	panel.SetSize(40, 20)
	selector := NewAgentSelector([]AgentOption{{Name: "codex"}, {Name: "claude"}})
	selector.Open = true
	panel.SetAgentSelector(selector)
	view := panel.View()
	if !strings.Contains(view, "codex") {
		t.Fatalf("expected selector in view")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/tui -run TestChatPanelRendersSelector -v`
Expected: FAIL

**Step 3: Implement selector rendering + update handling**

- Add `selector *AgentSelector` to `ChatPanel` with setter.
- Adjust height calculations to reserve 1–2 lines for selector when open/closed.
- In `Update`, handle `selector.Update` before composer and return selection message.
- In `View`, render selector below composer.

**Step 4: Update views with static chat panes**

- Add `agentSelector *pkgtui.AgentSelector` field to each view that renders chat.
- Accept selector in constructors (or via `SetAgentSelector` method).
- In each `Update`, pass key events to selector and return any `AgentSelectedMsg`.
- In each `renderChat`, append `selector.View()` under the chat pane.

**Step 5: Run tests**

Run: `go test ./pkg/tui -run TestChatPanelRendersSelector -v`
Expected: PASS

**Step 6: Commit**

```bash
git add pkg/tui/chatpanel.go internal/tui/views/*.go internal/gurgeh/arbiter/tui/arbiter_view.go
git add pkg/tui/chatpanel_test.go
git commit -m "feat(tui): render agent selector under chat panes"
```

---

### Task 5: Apply selection to active coding agent

**Files:**
- Modify: `internal/autarch/agent/detect.go`
- Modify: `internal/tui/unified_app.go`
- Modify: `cmd/autarch/main.go`
- Modify: `cmd/testui/main.go`

**Step 1: Write a failing test for resolving by name**

```go
func TestDetectAgentByNamePrefersName(t *testing.T) {
	agent, err := DetectAgentByName("codex", func(name string) (string, error) { return "/bin/codex", nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent.Type != TypeCodex {
		t.Fatalf("expected codex type, got %v", agent.Type)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/autarch/agent -run TestDetectAgentByNamePrefersName -v`
Expected: FAIL

**Step 3: Implement DetectAgentByName**

```go
func DetectAgentByName(name string, lookPath func(string) (string, error)) (*Agent, error) {
	switch strings.ToLower(name) {
	case "claude":
		path, err := lookPath("claude")
		if err != nil { return nil, err }
		return &Agent{Type: TypeClaude, Path: path, Version: getVersion(path, "--version")}, nil
	case "codex":
		path, err := lookPath("codex")
		if err != nil { return nil, err }
		return &Agent{Type: TypeCodex, Path: path, Version: getVersion(path, "--version")}, nil
	default:
		return nil, fmt.Errorf("unsupported agent %q", name)
	}
}
```

**Step 4: Wire selection into UnifiedApp**

- Add `agentSelector *pkgtui.AgentSelector` and `selectedAgent string` to `UnifiedApp`.
- Handle `AgentSelectedMsg` in `Update` to call `DetectAgentByName` and set `codingAgent`.
- On init, set default `selectedAgent` to detected preferred agent (`agent.DetectAgent()`) or first option.
- Use the shared selector instance when creating views (via new `SetAgentSelector` or constructor params).

**Step 5: Run tests**

Run: `go test ./internal/autarch/agent -run TestDetectAgentByNamePrefersName -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/autarch/agent/detect.go internal/tui/unified_app.go cmd/autarch/main.go cmd/testui/main.go
git add internal/autarch/agent/detect_test.go
git commit -m "feat(tui): switch coding agent via selector"
```

---

### Task 6: Command palette + help/doc updates

**Files:**
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/unified_app.go`
- Modify: `docs/tui/SHORTCUTS.md`
- Modify: `AGENTS.md`
- Modify: `internal/tui/views/*` (ShortHelp/FullHelp strings)

**Step 1: Add palette command**

Add a global command named "Switch agent/model" that toggles the selector open and focuses it.

**Step 2: Update help/shortcuts**

- Add `F2` to global help overlay and footer strings.
- Add `F2: agent` to ShortHelp where chat panes exist.
- Document `F2` in `docs/tui/SHORTCUTS.md` and `AGENTS.md`.

**Step 3: Run tests**

Run: `go test ./internal/tui/views -run TestViewShortHelpIncludesTab -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/tui/app.go internal/tui/unified_app.go internal/tui/views/*.go docs/tui/SHORTCUTS.md AGENTS.md
git add internal/tui/views/views_test.go
git commit -m "docs(tui): document F2 agent selector"
```

---

## Final verification

Run:
- `go test ./pkg/agenttargets -run TestMergeDetectedPrefersConfig -v`
- `go test ./pkg/tui -run TestAgentSelector -v`
- `go test ./internal/tui -run TestBuildAgentOptionsDedupePrefersConfig -v`
- `go test ./internal/autarch/agent -run TestDetectAgentByNamePrefersName -v`

---

Plan complete and saved to `docs/plans/2026-01-28-agent-model-selector-implementation-plan.md`.

Two execution options:
1) Subagent-Driven (this session)
2) Parallel Session (separate)

Which approach?
