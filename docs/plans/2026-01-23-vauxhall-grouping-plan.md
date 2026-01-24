# Vauxhall Project Grouping (TUI + Web) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** `Vauxpraudemonium-3yl` (Task reference)

**Goal:** Group sessions and agents by project in the Vauxhall TUI and Web UI, with collapsible groups in the TUI.

**Architecture:** Build grouped views derived from `ProjectPath`. For TUI, flatten grouped items into list items that include group headers and support expand/collapse per tab. For Web, group server-side and render group headers in templates.

**Tech Stack:** Go, Bubble Tea, lipgloss, html/template, htmx

---

### Task 1: TUI group header item + grouping helpers

**Files:**
- Modify: `internal/vauxhall/tui/model.go`
- Test: `internal/vauxhall/tui/model_layout_test.go`

**Step 1: Write the failing tests**

Add to `internal/vauxhall/tui/model_layout_test.go`:
```go
func TestSessionGroupingBuildsHeaders(t *testing.T) {
	agg := &fakeAggLayout{state: aggregator.State{
		Projects: []discovery.Project{{Path: "/p/one"}, {Path: "/p/two"}},
		Sessions: []aggregator.TmuxSession{{Name: "a", ProjectPath: "/p/one"}, {Name: "b", ProjectPath: "/p/two"}},
	}}
	m := New(agg, "")
	m.activeTab = TabSessions
	m.updateLists()
	items := m.sessionList.Items()
	if len(items) != 4 {
		t.Fatalf("expected 4 items (2 headers + 2 sessions), got %d", len(items))
	}
	if _, ok := items[0].(GroupHeaderItem); !ok {
		t.Fatalf("expected header as first item")
	}
}

func TestAgentGroupingBuildsHeaders(t *testing.T) {
	agg := &fakeAggLayout{state: aggregator.State{
		Projects: []discovery.Project{{Path: "/p/one"}},
		Agents: []aggregator.Agent{{Name: "Alpha", ProjectPath: "/p/one"}},
	}}
	m := New(agg, "")
	m.activeTab = TabAgents
	m.updateLists()
	items := m.agentList.Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 items (header + agent), got %d", len(items))
	}
	if _, ok := items[0].(GroupHeaderItem); !ok {
		t.Fatalf("expected header as first item")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/vauxhall/tui -run TestSessionGroupingBuildsHeaders -v`
Expected: FAIL (missing GroupHeaderItem / no grouping)

**Step 3: Write minimal implementation**

In `internal/vauxhall/tui/model.go`:
- Add `GroupHeaderItem` type implementing `list.Item` (Title/Description/FilterValue).
- Add grouping helpers:
  - `groupSessionsByProject([]SessionItem) []list.Item`
  - `groupAgentsByProject([]AgentItem) []list.Item`
- Insert headers before items, skip empty groups.
- Unassigned `ProjectPath` -> header title `Unassigned`.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/vauxhall/tui -run TestSessionGroupingBuildsHeaders -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/vauxhall/tui/model.go internal/vauxhall/tui/model_layout_test.go
git commit -m "feat(vauxhall): add tui grouping headers"
```

---

### Task 2: TUI expand/collapse per group

**Files:**
- Modify: `internal/vauxhall/tui/model.go`
- Test: `internal/vauxhall/tui/model_layout_test.go`

**Step 1: Write the failing tests**

Add to `internal/vauxhall/tui/model_layout_test.go`:
```go
func TestGroupCollapseHidesItems(t *testing.T) {
	agg := &fakeAggLayout{state: aggregator.State{
		Sessions: []aggregator.TmuxSession{{Name: "a", ProjectPath: "/p/one"}},
	}}
	m := New(agg, "")
	m.activeTab = TabSessions
	m.groupExpanded = map[string]bool{"sessions:/p/one": false}
	m.updateLists()
	items := m.sessionList.Items()
	if len(items) != 1 {
		t.Fatalf("expected only header when collapsed")
	}
	if _, ok := items[0].(GroupHeaderItem); !ok {
		t.Fatalf("expected header item")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/vauxhall/tui -run TestGroupCollapseHidesItems -v`
Expected: FAIL

**Step 3: Write minimal implementation**

In `internal/vauxhall/tui/model.go`:
- Add `groupExpanded map[string]bool` to `Model`.
- Add helper `groupKey(tab Tab, projectPath string) string`.
- Default to expanded when not present.
- Apply expand/collapse in grouping helpers.
- Add key handler for `g`: if selected item is `GroupHeaderItem`, toggle expand and call `updateLists()`.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/vauxhall/tui -run TestGroupCollapseHidesItems -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/vauxhall/tui/model.go internal/vauxhall/tui/model_layout_test.go
git commit -m "feat(vauxhall): add group expand/collapse"
```

---

### Task 3: Filters apply before grouping

**Files:**
- Modify: `internal/vauxhall/tui/model_layout_test.go`
- Modify: `internal/vauxhall/tui/model.go`

**Step 1: Write failing test**

Add to `internal/vauxhall/tui/model_layout_test.go`:
```go
func TestGroupingAppliesAfterFilter(t *testing.T) {
	agg := &fakeAggLayout{state: aggregator.State{
		Sessions: []aggregator.TmuxSession{{Name: "codex", ProjectPath: "/p/one"}, {Name: "claude", ProjectPath: "/p/two"}},
	}}
	m := New(agg, "")
	m.activeTab = TabSessions
	m.filterStates[TabSessions] = parseFilter("codex")
	m.updateLists()
	items := m.sessionList.Items()
	if len(items) != 2 {
		t.Fatalf("expected 1 header + 1 item after filter")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/vauxhall/tui -run TestGroupingAppliesAfterFilter -v`
Expected: FAIL

**Step 3: Write minimal implementation**

Ensure `updateLists()` applies `filterSessionItems` / `filterAgentItems` before grouping helpers.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/vauxhall/tui -run TestGroupingAppliesAfterFilter -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/vauxhall/tui/model.go internal/vauxhall/tui/model_layout_test.go
git commit -m "feat(vauxhall): filter before grouping"
```

---

### Task 4: Web grouping rendering

**Files:**
- Modify: `internal/vauxhall/web/server.go`
- Modify: `internal/vauxhall/web/templates/sessions.html`
- Modify: `internal/vauxhall/web/templates/agents.html`
- Test: `internal/vauxhall/web/server_filter_test.go`

**Step 1: Write failing tests**

Add to `internal/vauxhall/web/server_filter_test.go`:
```go
func TestSessionsGroupedByProject(t *testing.T) {
	agg := &fakeAgg{state: aggregator.State{Sessions: []aggregator.TmuxSession{{Name: "a", ProjectPath: "/p/one"}, {Name: "b", ProjectPath: "/p/two"}}}}
	srv := NewServer(config.ServerConfig{}, agg)
	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	w := httptest.NewRecorder()
	srv.handleSessions(w, req)
	body := w.Body.String()
	if !strings.Contains(body, "Project: one") || !strings.Contains(body, "Project: two") {
		t.Fatalf("expected project group headers in response")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/vauxhall/web -run TestSessionsGroupedByProject -v`
Expected: FAIL

**Step 3: Write minimal implementation**

In `internal/vauxhall/web/server.go`:
- Add a small `type WebGroup[T any] struct { Name string; Path string; Items []T }` in file.
- In `handleSessions` and `handleAgents`, group after filtering and pass `Groups` to template, also include flat list as fallback if needed.
- Group order: sort by `basename` (Unassigned last).
- For missing `ProjectPath`, use `Name: "Unassigned"`.

Update templates to render groups:
- Sessions/Agents: loop over `.Groups` and render header row + items within.
- For sessions header include text `Project: <name>` to satisfy test (can be styled).

**Step 4: Run test to verify it passes**

Run: `go test ./internal/vauxhall/web -run TestSessionsGroupedByProject -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/vauxhall/web/server.go internal/vauxhall/web/templates/sessions.html internal/vauxhall/web/templates/agents.html internal/vauxhall/web/server_filter_test.go
git commit -m "feat(vauxhall): group web lists by project"
```

---

### Task 5: Full test run

**Files:**
- Test: `internal/vauxhall/tui/...`
- Test: `internal/vauxhall/web/...`

**Step 1: Run full tests**

Run: `go test ./internal/vauxhall/tui -v`
Expected: PASS

Run: `go test ./internal/vauxhall/web -v`
Expected: PASS

**Step 2: Optional full suite**

Run: `go test ./...`
Expected: PASS

**Step 3: Commit plan completion note (optional)**

```bash
git status -sb
```
