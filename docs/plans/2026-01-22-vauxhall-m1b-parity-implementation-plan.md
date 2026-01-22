# Vauxhall M1B Parity Core Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** `Vauxpraudemonium-8gv` (Task reference)

**Goal:** Add Agent Deck-style session control + repo MCP manager with web and TUI parity.

**Architecture:** Extend tmux client with action methods and injectable runner; add MCP manager package; wire actions through the aggregator so both web and TUI call the same control APIs; update web/TUI to expose actions and MCP panel.

**Tech Stack:** Go 1.24+, Bubble Tea + lipgloss, net/http + htmx, tmux CLI, SQLite (read-only), BurntSushi/toml.

---

## Task 1: tmux Action Runner + Session Control Methods

**Files:**
- Create: `internal/vauxhall/tmux/runner.go`
- Modify: `internal/vauxhall/tmux/client.go`
- Test: `internal/vauxhall/tmux/client_actions_test.go`

**Step 1: Write failing test for new session command composition**

```go
package tmux

import "testing"

type fakeRunner struct {
	calls [][]string
}

func (f *fakeRunner) Run(name string, args ...string) (string, string, error) {
	f.calls = append(f.calls, append([]string{name}, args...))
	return "", "", nil
}

func TestClientNewSessionCommand(t *testing.T) {
	fr := &fakeRunner{}
	c := NewClientWithRunner(fr)
	err := c.NewSession("claude-demo", "/root/projects/demo", []string{"claude"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fr.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fr.calls))
	}
	want := []string{"tmux", "new-session", "-d", "-s", "claude-demo", "-c", "/root/projects/demo", "claude"}
	got := fr.calls[0]
	if len(got) != len(want) {
		t.Fatalf("unexpected arg count: %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d: got %q want %q", i, got[i], want[i])
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/vauxhall/tmux -run TestClientNewSessionCommand -v`
Expected: FAIL with “NewClientWithRunner undefined” or “NewSession undefined”.

**Step 3: Implement runner interface + NewSession**

```go
// runner.go
package tmux

type Runner interface {
	Run(name string, args ...string) (stdout, stderr string, err error)
}

// client.go (excerpt)
type Client struct {
	tmuxPath string
	runner   Runner
	cache    *sessionCache
}

func NewClientWithRunner(r Runner) *Client {
	c := NewClient()
	c.runner = r
	return c
}

func (c *Client) run(args ...string) (string, string, error) {
	if c.runner == nil {
		c.runner = &execRunner{tmuxPath: c.tmuxPath}
	}
	return c.runner.Run(c.tmuxPath, args...)
}

type execRunner struct{ tmuxPath string }

func (r *execRunner) Run(name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func (c *Client) NewSession(name, path string, command []string) error {
	args := []string{"new-session", "-d", "-s", name}
	if path != "" {
		args = append(args, "-c", path)
	}
	args = append(args, command...)
	_, stderr, err := c.run(args...)
	if err != nil {
		return fmt.Errorf("failed to create session: %w: %s", err, stderr)
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/vauxhall/tmux -run TestClientNewSessionCommand -v`
Expected: PASS.

**Step 5: Add rename/kill/attach command tests**

```go
func TestClientRenameSessionCommand(t *testing.T) { /* assert tmux rename-session */ }
func TestClientKillSessionCommand(t *testing.T) { /* assert tmux kill-session */ }
```

**Step 6: Run tests to verify they fail**

Run: `go test ./internal/vauxhall/tmux -run TestClientRenameSessionCommand -v`
Expected: FAIL (methods missing).

**Step 7: Implement RenameSession + KillSession helpers**

```go
func (c *Client) RenameSession(oldName, newName string) error {
	_, stderr, err := c.run("rename-session", "-t", oldName, newName)
	if err != nil {
		return fmt.Errorf("failed to rename session: %w: %s", err, stderr)
	}
	return nil
}

func (c *Client) KillSession(name string) error {
	_, stderr, err := c.run("kill-session", "-t", name)
	if err != nil {
		return fmt.Errorf("failed to kill session: %w: %s", err, stderr)
	}
	return nil
}
```

**Step 8: Run full tmux package tests**

Run: `go test ./internal/vauxhall/tmux -v`
Expected: PASS.

**Step 9: Commit**

```bash
git add internal/vauxhall/tmux/runner.go internal/vauxhall/tmux/client.go internal/vauxhall/tmux/client_actions_test.go
git commit -m "feat(vauxhall): add tmux action runner and session controls"
```

---

## Task 2: Agent Command Resolution (Config + Praude + Defaults)

**Files:**
- Modify: `internal/vauxhall/config/config.go`
- Create: `internal/vauxhall/agentcmd/resolver.go`
- Test: `internal/vauxhall/agentcmd/resolver_test.go`

**Step 1: Write failing resolver test**

```go
package agentcmd

import (
	"testing"
	vconfig "github.com/mistakeknot/vauxpraudemonium/internal/vauxhall/config"
)

func TestResolveCommandFallback(t *testing.T) {
	cfg := &vconfig.Config{}
	r := NewResolver(cfg)
	cmd, args := r.Resolve("claude", "/root/projects/demo")
	if cmd != "claude" {
		t.Fatalf("expected claude fallback, got %q", cmd)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/vauxhall/agentcmd -run TestResolveCommandFallback -v`
Expected: FAIL (package not found).

**Step 3: Implement config structures + resolver**

```go
// config.go (excerpt)

type AgentCommand struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
}

type MCPComponentConfig struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
	Workdir string   `toml:"workdir"`
}

type MCPConfig struct {
	Server MCPComponentConfig `toml:"server"`
	Client MCPComponentConfig `toml:"client"`
}

type Config struct {
	Server    ServerConfig    `toml:"server"`
	Discovery DiscoveryConfig `toml:"discovery"`
	Tmux      TmuxConfig      `toml:"tmux"`
	Agents    map[string]AgentCommand `toml:"agents"`
	MCP       MCPConfig       `toml:"mcp"`
}
```

```go
// resolver.go
package agentcmd

import (
	strings

	pconfig "github.com/mistakeknot/vauxpraudemonium/internal/praude/config"
	vconfig "github.com/mistakeknot/vauxpraudemonium/internal/vauxhall/config"
)

type Resolver struct { cfg *vconfig.Config }

func NewResolver(cfg *vconfig.Config) *Resolver { return &Resolver{cfg: cfg} }

func (r *Resolver) Resolve(agentType, projectPath string) (string, []string) {
	key := strings.ToLower(agentType)
	if r.cfg != nil && r.cfg.Agents != nil {
		if cmd, ok := r.cfg.Agents[key]; ok && cmd.Command != "" {
			return cmd.Command, cmd.Args
		}
	}
	if projectPath != "" {
		if pcfg, err := pconfig.LoadFromRoot(projectPath); err == nil {
			if ap, ok := pcfg.Agents[key]; ok && ap.Command != "" {
				return ap.Command, ap.Args
			}
		}
	}
	if key == "codex" {
		return "codex", nil
	}
	return "claude", nil
}
```

**Step 4: Run tests**

Run: `go test ./internal/vauxhall/agentcmd -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/vauxhall/config/config.go internal/vauxhall/agentcmd/resolver.go internal/vauxhall/agentcmd/resolver_test.go
git commit -m "feat(vauxhall): add agent command resolver"
```

---

## Task 3: MCP Manager (Repo MCPs Only)

**Files:**
- Create: `internal/vauxhall/mcp/manager.go`
- Create: `internal/vauxhall/mcp/manager_test.go`

**Step 1: Write failing test for idempotent start/stop**

```go
package mcp

import "testing"

func TestManagerStartStopIdempotent(t *testing.T) {
	m := NewManager()
	if err := m.Stop("/root/projects/demo", "server"); err != nil {
		t.Fatalf("stop should be idempotent: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/vauxhall/mcp -run TestManagerStartStopIdempotent -v`
Expected: FAIL (package missing).

**Step 3: Implement manager skeleton + status tracking**

```go
package mcp

import (
	"context"
	"sync"
	"time"
)

type Status string

const (
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusError   Status = "error"
)

type ComponentStatus struct {
	ProjectPath string
	Component   string // server | client
	Status      Status
	Pid         int
	StartedAt   time.Time
	LastError   string
	LogTail     []string
}

type Manager struct {
	mu     sync.RWMutex
	items  map[string]*ComponentStatus
}

func NewManager() *Manager { return &Manager{items: make(map[string]*ComponentStatus)} }

func key(project, component string) string { return project + "::" + component }

func (m *Manager) Stop(project, component string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := key(project, component)
	if _, ok := m.items[k]; !ok {
		m.items[k] = &ComponentStatus{ProjectPath: project, Component: component, Status: StatusStopped}
		return nil
	}
	m.items[k].Status = StatusStopped
	return nil
}

func (m *Manager) Start(ctx context.Context, project, component string, cmd []string, workdir string) error {
	// implementation stub for now
	return nil
}

func (m *Manager) Status(project, component string) *ComponentStatus { /* ... */ return nil }
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/vauxhall/mcp -run TestManagerStartStopIdempotent -v`
Expected: PASS.

**Step 5: Add start/stop process runner + log tail tests**

```go
func TestManagerLogTailUpdates(t *testing.T) {
	// use fake process runner to push log lines and verify tail cap
}
```

**Step 6: Implement process runner injection + log tail ring buffer**

- Add `Runner` interface to allow tests to inject fake processes.
- Cap `LogTail` at 50 lines.
- Set `StatusError` with `LastError` on spawn failure.

**Step 7: Run MCP manager tests**

Run: `go test ./internal/vauxhall/mcp -v`
Expected: PASS.

**Step 8: Commit**

```bash
git add internal/vauxhall/mcp/manager.go internal/vauxhall/mcp/manager_test.go
git commit -m "feat(vauxhall): add repo MCP manager"
```

---

## Task 4: Aggregator Actions + MCP Status Wiring

**Files:**
- Modify: `internal/vauxhall/aggregator/aggregator.go`
- Modify: `cmd/vauxhall/main.go`
- Test: `internal/vauxhall/aggregator/aggregator_actions_test.go`

**Step 1: Write failing test for restart action**

```go
package aggregator

import "testing"

type fakeTmux struct { killed, created bool }

func (f *fakeTmux) KillSession(name string) error { f.killed = true; return nil }
func (f *fakeTmux) NewSession(name, path string, cmd []string) error { f.created = true; return nil }

func TestRestartSession(t *testing.T) {
	agg := New(nil)
	agg.tmuxClient = &fakeTmux{}
	if err := agg.RestartSession("demo", "/root/projects/demo", []string{"claude"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/vauxhall/aggregator -run TestRestartSession -v`
Expected: FAIL (method missing / type mismatch).

**Step 3: Add tmux client interface + action methods**

```go
// aggregator.go (excerpt)
type tmuxAPI interface {
	ListSessions() ([]tmux.Session, error)
	IsAvailable() bool
	DetectStatus(string) tmux.Status
	NewSession(name, path string, cmd []string) error
	RenameSession(oldName, newName string) error
	KillSession(name string) error
}
```

Add action methods:

```go
func (a *Aggregator) RestartSession(name, path string, cmd []string) error {
	if err := a.tmuxClient.KillSession(name); err != nil { return err }
	return a.tmuxClient.NewSession(name, path, cmd)
}
```

**Step 4: Add MCP manager to aggregator state**

- Store `mcp.Manager` on aggregator.
- Add `MCPStatuses` to `State` (map of project -> []ComponentStatus).
- Populate in `Refresh`.

**Step 5: Run tests**

Run: `go test ./internal/vauxhall/aggregator -v`
Expected: PASS.

**Step 6: Commit**

```bash
git add internal/vauxhall/aggregator/aggregator.go cmd/vauxhall/main.go internal/vauxhall/aggregator/aggregator_actions_test.go
git commit -m "feat(vauxhall): wire aggregator session actions and MCP status"
```

---

## Task 5: Web Actions + MCP Panel

**Files:**
- Modify: `internal/vauxhall/web/server.go`
- Modify: `internal/vauxhall/web/templates/sessions.html`
- Modify: `internal/vauxhall/web/templates/projects.html`
- Test: `internal/vauxhall/web/server_actions_test.go`

**Step 1: Write failing handler test for restart endpoint**

```go
package web

import (
	"net/http/httptest"
	"testing"
)

func TestRestartSessionEndpoint(t *testing.T) {
	srv := NewServer(config.ServerConfig{}, fakeAgg{})
	req := httptest.NewRequest("POST", "/api/sessions/demo/restart", nil)
	w := httptest.NewRecorder()
	srv.handleSessionRestart(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/vauxhall/web -run TestRestartSessionEndpoint -v`
Expected: FAIL (handler missing).

**Step 3: Implement session action endpoints**

- `POST /api/sessions/{name}/restart`
- `POST /api/sessions/{name}/rename` (body JSON: `{"name":"new"}`)
- `POST /api/sessions/new` (JSON: name, project, agent)
- `POST /api/sessions/{name}/fork`

**Step 4: Add MCP action endpoints**

- `POST /api/projects/{path}/mcp/{component}/start`
- `POST /api/projects/{path}/mcp/{component}/stop`

**Step 5: Update templates to add action buttons + MCP panel**

- Sessions page row actions with htmx
- Project page MCP panel with start/stop toggles and log tail display

**Step 6: Run web tests**

Run: `go test ./internal/vauxhall/web -v`
Expected: PASS.

**Step 7: Commit**

```bash
git add internal/vauxhall/web/server.go internal/vauxhall/web/templates/sessions.html internal/vauxhall/web/templates/projects.html internal/vauxhall/web/server_actions_test.go
git commit -m "feat(vauxhall): add web session actions and MCP panel"
```

---

## Task 6: TUI Actions + MCP Panel

**Files:**
- Modify: `internal/vauxhall/tui/model.go`
- Modify: `internal/vauxhall/tui/styles.go`
- Test: `internal/vauxhall/tui/model_actions_test.go`

**Step 1: Write failing TUI keybinding test**

```go
package tui

import (
	"testing"
	tea "github.com/charmbracelet/bubbletea"
)

func TestKeybindingsSessionActions(t *testing.T) {
	m := New(fakeAgg{})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd == nil {
		t.Fatalf("expected cmd for new session")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/vauxhall/tui -run TestKeybindingsSessionActions -v`
Expected: FAIL (missing logic).

**Step 3: Implement session action prompts**

- Add simple prompt state for new/rename/fork using `bubbles/textinput`.
- Wire keys: `/` search, `g` group toggle, `a` attach, `n` new, `r` rename, `k` restart, `f` fork, `m` MCP panel.
- Send action requests through aggregator.

**Step 4: Add MCP panel view + toggles**

- Show per-project MCP components with status.
- `space` toggles start/stop.

**Step 5: Run TUI tests**

Run: `go test ./internal/vauxhall/tui -v`
Expected: PASS.

**Step 6: Commit**

```bash
git add internal/vauxhall/tui/model.go internal/vauxhall/tui/styles.go internal/vauxhall/tui/model_actions_test.go
git commit -m "feat(vauxhall): add TUI session actions and MCP panel"
```

---

## Task 7: Manual Verification Checklist

- Run: `./dev vauxhall` and verify session list, search, grouping, actions.
- Run: `./dev vauxhall --tui` and verify keybindings and MCP toggles.
- Confirm restart = kill + recreate.
- Confirm new/fork uses config → Praude → defaults.
- Confirm MCP manager starts/stops repo `mcp-server` + `mcp-client`.

---

Plan complete and saved to `docs/plans/2026-01-22-vauxhall-m1b-parity-implementation-plan.md`.

Two execution options:
1. Subagent-Driven (this session) - use @superpowers:subagent-driven-development
2. Parallel Session (separate) - new session uses @superpowers:executing-plans

Which approach?
