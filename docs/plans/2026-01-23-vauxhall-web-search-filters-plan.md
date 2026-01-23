# Vauxhall Web Search + Status Filters Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** `Vauxpraudemonium-fuo` (Task reference)

**Goal:** Add a web UI search bar with status-token filtering for Sessions and Agents, mirroring TUI behavior.

**Architecture:** Add filter parsing + matching helpers in `internal/vauxhall/web`, wire query param filtering into `/sessions` and `/agents`, and update templates to include a search input that updates the list via htmx (with fallback to full-page GET).

**Tech Stack:** Go, net/http, html/template, htmx, Tailwind

---

### Task 1: Add filter parsing + matching helpers

**Files:**
- Create: `internal/vauxhall/web/filter.go`
- Create: `internal/vauxhall/web/filter_test.go`

**Step 1: Write the failing tests**

Create `internal/vauxhall/web/filter_test.go`:
```go
package web

import (
	"testing"

	"github.com/mistakeknot/vauxpraudemonium/internal/vauxhall/aggregator"
	"github.com/mistakeknot/vauxpraudemonium/internal/vauxhall/tmux"
)

func TestParseFilterStatusTokens(t *testing.T) {
	state := parseFilter("!waiting codex")
	if !state.Statuses[tmux.StatusWaiting] {
		t.Fatalf("expected waiting status")
	}
	if len(state.Terms) != 1 || state.Terms[0] != "codex" {
		t.Fatalf("expected codex term")
	}
}

func TestFilterSessionsAppliesStatusAndText(t *testing.T) {
	sessions := []aggregator.TmuxSession{
		{Name: "codex-a", AgentType: "codex"},
		{Name: "codex-b", AgentType: "codex"},
		{Name: "claude"},
	}
	statusBySession := map[string]tmux.Status{
		"codex-a": tmux.StatusWaiting,
		"codex-b": tmux.StatusRunning,
		"claude":  tmux.StatusWaiting,
	}
	state := parseFilter("!waiting codex")
	filtered := filterSessions(sessions, state, statusBySession)
	if len(filtered) != 1 || filtered[0].Name != "codex-a" {
		t.Fatalf("unexpected filter result")
	}
}

func TestFilterAgentsMatchesUnknownStatus(t *testing.T) {
	agents := []aggregator.Agent{
		{Name: "Copper"},
		{Name: "Rose", SessionName: "rose"},
	}
	statusBySession := map[string]tmux.Status{
		"rose": tmux.StatusRunning,
	}
	state := parseFilter("!unknown")
	filtered := filterAgents(agents, state, statusBySession)
	if len(filtered) != 1 || filtered[0].Name != "Copper" {
		t.Fatalf("expected unknown agent")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/vauxhall/web -run TestParseFilterStatusTokens -v`
Expected: FAIL with “undefined: parseFilter”

**Step 3: Write minimal implementation**

Create `internal/vauxhall/web/filter.go` with:
- `type FilterState struct { Raw string; Terms []string; Statuses map[tmux.Status]bool }`
- `func parseFilter(input string) FilterState` supporting `!running`, `!waiting`, `!idle`, `!error`, `!unknown`
- `func filterSessions(sessions []aggregator.TmuxSession, state FilterState, statusBySession map[string]tmux.Status) []aggregator.TmuxSession`
  - Text terms match against `Name`, `AgentName`, `AgentType`, `ProjectPath`
  - If status tokens present, match against `statusBySession[name]` (missing → `unknown`)
- `func filterAgents(agents []aggregator.Agent, state FilterState, statusBySession map[string]tmux.Status) []aggregator.Agent`
  - Text terms match against `Name`, `Program`, `Model`, `ProjectPath`
  - Status comes from `agent.SessionName` via `statusBySession` (missing → `unknown`)

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/vauxhall/web -run TestParseFilterStatusTokens -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/vauxhall/web/filter.go internal/vauxhall/web/filter_test.go
git commit -m "feat(vauxhall): add web filter helpers"
```

---

### Task 2: Wire filtering into handlers (query param)

**Files:**
- Modify: `internal/vauxhall/web/server.go`
- Create: `internal/vauxhall/web/server_filter_test.go`

**Step 1: Write the failing tests**

Create `internal/vauxhall/web/server_filter_test.go`:
```go
package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mistakeknot/vauxpraudemonium/internal/vauxhall/aggregator"
	"github.com/mistakeknot/vauxpraudemonium/internal/vauxhall/config"
	"github.com/mistakeknot/vauxpraudemonium/internal/vauxhall/tmux"
)

type fakeStatusClient struct {
	bySession map[string]tmux.Status
}

func (f fakeStatusClient) DetectStatus(name string) tmux.Status {
	if status, ok := f.bySession[name]; ok {
		return status
	}
	return tmux.StatusUnknown
}

func TestHandleSessionsFiltersByQuery(t *testing.T) {
	agg := &fakeAgg{state: aggregator.State{
		Sessions: []aggregator.TmuxSession{{Name: "codex"}, {Name: "claude"}},
	}}
	srv := NewServer(config.ServerConfig{}, agg)
	srv.statusClient = fakeStatusClient{bySession: map[string]tmux.Status{"codex": tmux.StatusWaiting}}

	req := httptest.NewRequest(http.MethodGet, "/sessions?q=!waiting", nil)
	w := httptest.NewRecorder()
	srv.handleSessions(w, req)
	body := w.Body.String()
	if !strings.Contains(body, "codex") || strings.Contains(body, "claude") {
		t.Fatalf("expected filtered sessions in response")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/vauxhall/web -run TestHandleSessionsFiltersByQuery -v`
Expected: FAIL (query not applied)

**Step 3: Write minimal implementation**

In `internal/vauxhall/web/server.go`:
- Add `statusClient` to `Server` (interface with `DetectStatus`), default to `tmux.NewClient()` in `NewServer`.
- In `handleSessions`:
  - Read `q := r.URL.Query().Get("q")`
  - `state := parseFilter(q)`
  - If `state.Statuses` not empty, build `statusBySession` using `statusClient.DetectStatus`.
  - `sessions := filterSessions(state.Sessions, state, statusBySession)`
  - Render template with `Sessions` + `Query`.
- In `handleAgents`:
  - Read `q` and compute `statusBySession` if needed.
  - `agents := filterAgents(state.Agents, state, statusBySession)`
  - Render template with `Agents` + `Query`.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/vauxhall/web -run TestHandleSessionsFiltersByQuery -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/vauxhall/web/server.go internal/vauxhall/web/server_filter_test.go
git commit -m "feat(vauxhall): apply web query filters"
```

---

### Task 3: Add search input to Sessions/Agents templates

**Files:**
- Modify: `internal/vauxhall/web/templates/sessions.html`
- Modify: `internal/vauxhall/web/templates/agents.html`

**Step 1: Write the failing test**

Add to `internal/vauxhall/web/server_filter_test.go`:
```go
func TestSessionsTemplateShowsQueryValue(t *testing.T) {
	agg := &fakeAgg{state: aggregator.State{Sessions: []aggregator.TmuxSession{{Name: "codex"}}}}
	srv := NewServer(config.ServerConfig{}, agg)
	req := httptest.NewRequest(http.MethodGet, "/sessions?q=codex", nil)
	w := httptest.NewRecorder()
	srv.handleSessions(w, req)
	if !strings.Contains(w.Body.String(), "value=\"codex\"") {
		t.Fatalf("expected query value in input")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/vauxhall/web -run TestSessionsTemplateShowsQueryValue -v`
Expected: FAIL (no input)

**Step 3: Write minimal implementation**

Update `sessions.html` and `agents.html`:
- Add a search input near the header with:
  - `name="q"`, `value="{{.Query}}"`
  - `hx-get="/sessions"` or `/agents`
  - `hx-trigger="input changed delay:300ms"`
  - `hx-target="#sessions-list"` / `#agents-list`
  - `hx-select="#sessions-list"` / `#agents-list`
- Add a small hint text for status tokens: `!running !waiting !idle !error !unknown`
- Update the Refresh button to include `hx-select` so htmx swaps only the list.
- Wrap the list in an element with matching `id` (`sessions-list`, `agents-list`).

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/vauxhall/web -run TestSessionsTemplateShowsQueryValue -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/vauxhall/web/templates/sessions.html internal/vauxhall/web/templates/agents.html internal/vauxhall/web/server_filter_test.go
git commit -m "feat(vauxhall): add web search inputs"
```

---

### Task 4: Full web test run

**Files:**
- Test: `internal/vauxhall/web/...`

**Step 1: Run full tests**

Run: `go test ./internal/vauxhall/web -v`
Expected: PASS

**Step 2: (Optional) full suite**

Run: `go test ./...`
Expected: PASS

**Step 3: Commit plan completion note (optional)**

```bash
git status -sb
```
