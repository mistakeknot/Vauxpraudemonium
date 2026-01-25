# Praude Interview Iteration + New PRD Flow Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** `Autarch-interview-iter` (Task reference)

**Goal:** Make `n` create a new PRD and immediately start an agent-driven interview; make `g` re-interview the selected PRD with per-step iteration (enter) and step navigation (`[`/`]`), preserving non-interview fields.

**Architecture:** Add interview state for target spec, per-step answers/drafts, and agent iteration. Implement a synchronous agent runner to return draft text per step. Create a blank PRD on `n`, load existing PRD on `g`, and merge updated fields into the existing spec on finalize.

**Tech Stack:** Go 1.24+, Bubble Tea, YAML, exec.Command

Note: user explicitly requested no worktrees; implementation will be in the current working tree.

---

### Task 1: Add failing TUI tests for `n`, re-interview, and iteration

**Files:**
- Modify: `internal/praude/tui/interview_test.go`
- Create: `internal/praude/tui/interview_iteration_test.go`

**Step 1: Write the failing tests**

Create `internal/praude/tui/interview_iteration_test.go`:

```go
package tui

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/mistakeknot/autarch/internal/praude/agents"
)

func TestNewKeyStartsInterviewForNewSpec(t *testing.T) {
    root := t.TempDir()
    _ = os.MkdirAll(filepath.Join(root, ".praude", "specs"), 0o755)
    cwd, _ := os.Getwd()
    defer func() { _ = os.Chdir(cwd) }()
    _ = os.Chdir(root)

    m := NewModel()
    m = pressKey(m, "n")
    if m.mode != "interview" {
        t.Fatalf("expected interview mode")
    }
    entries, _ := os.ReadDir(filepath.Join(root, ".praude", "specs"))
    if len(entries) != 1 {
        t.Fatalf("expected new spec file")
    }
}

func TestInterviewEnterIteratesDraft(t *testing.T) {
    root := t.TempDir()
    _ = os.MkdirAll(filepath.Join(root, ".praude", "specs"), 0o755)
    cwd, _ := os.Getwd()
    defer func() { _ = os.Chdir(cwd) }()
    _ = os.Chdir(root)

    oldRun := runAgent
    runAgent = func(p agents.Profile, briefPath string) ([]byte, error) {
        return []byte("Drafted vision"), nil
    }
    defer func() { runAgent = oldRun }()

    m := NewModel()
    m = pressKey(m, "n")
    m = typeAndEnter(m, "Initial vision")
    out := m.View()
    if !strings.Contains(out, "Drafted vision") {
        t.Fatalf("expected draft in view")
    }
}
```

Update `internal/praude/tui/interview_test.go` to use `n` for new interview instead of `g`, and to navigate with `]` for next steps (no Enter-advance for text steps).

**Step 2: Run tests to verify failure**

Run:
```
go test ./internal/praude/tui -run TestNewKeyStartsInterviewForNewSpec
```

Expected: FAIL (no `n` handling, no runAgent, no iteration behavior).

**Step 3: Implement minimal code to pass**
(see tasks 2-3 for code changes)

**Step 4: Run tests to verify pass**
```
go test ./internal/praude/tui -run TestNewKeyStartsInterviewForNewSpec
```

Expected: PASS

**Step 5: Commit**
```
git add internal/praude/tui/interview_test.go internal/praude/tui/interview_iteration_test.go
git commit -m "test(praude): cover interview new/iterate flow"
```

---

### Task 2: Add synchronous agent runner for interview iteration

**Files:**
- Modify: `internal/praude/agents/agents.go`
- Create: `internal/praude/agents/agents_run_test.go`
- Modify: `internal/praude/tui/agent_launch.go`

**Step 1: Write the failing test**

Create `internal/praude/agents/agents_run_test.go`:

```go
package agents

import (
    "os"
    "path/filepath"
    "testing"
)

func TestRunReturnsStdout(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "input.txt")
    if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
        t.Fatal(err)
    }
    out, err := Run(Profile{Command: "cat", Args: []string{}}, path)
    if err != nil {
        t.Fatal(err)
    }
    if string(out) != "hello" {
        t.Fatalf("expected stdout")
    }
}
```

**Step 2: Run test to verify it fails**
```
go test ./internal/praude/agents -run TestRunReturnsStdout
```

Expected: FAIL (Run missing).

**Step 3: Implement minimal code**

In `internal/praude/agents/agents.go`, add:

```go
func Run(p Profile, briefPath string) ([]byte, error) {
    return runWithEnv(p, briefPath, nil)
}

func RunSubagent(p Profile, briefPath string) ([]byte, error) {
    return runWithEnv(p, briefPath, []string{"PRAUDE_SUBAGENT=1"})
}

func runWithEnv(p Profile, briefPath string, extraEnv []string) ([]byte, error) {
    if _, err := lookPath(p.Command); err != nil {
        return nil, err
    }
    cmd := buildCommand(p, briefPath, extraEnv)
    return cmd.CombinedOutput()
}
```

In `internal/praude/tui/agent_launch.go`, add:

```go
var runAgent = agents.Run
var runSubagent = agents.RunSubagent
```

**Step 4: Run test to verify it passes**
```
go test ./internal/praude/agents -run TestRunReturnsStdout
```

Expected: PASS

**Step 5: Commit**
```
git add internal/praude/agents/agents.go internal/praude/agents/agents_run_test.go internal/praude/tui/agent_launch.go
git commit -m "feat(praude): add synchronous agent runner"
```

---

### Task 3: Implement `n` create+interview and `g` re-interview with iteration

**Files:**
- Modify: `internal/praude/tui/model.go`
- Modify: `internal/praude/tui/interview.go`
- Modify: `internal/praude/tui/overlay.go`
- Modify: `internal/praude/specs/create.go`
- Create: `internal/praude/tui/interview_agent.go`

**Step 1: Write the failing tests**
(covered by Task 1)

**Step 2: Implement minimal code**

- Add `specs.CreateBlank(dir, now)` that writes a minimal PRD with only id/status/created_at/title/summary placeholders.
- In `model.go`, handle `n` in list mode: call `CreateBlank`, reload summaries, select the new PRD, and enter interview mode.
- For `g`, if a PRD is selected, load it and start interview in re-interview mode; otherwise set status error.
- Extend `interviewState` with:
  - `targetID`, `targetPath`
  - `baseSpec specs.Spec`
  - `answers map[interviewStep]string`
  - `drafts map[interviewStep]string`
- Implement per-step iteration:
  - `enter` calls `iterateStep()` for text steps, which writes a brief and runs `runAgent`/`runSubagent` synchronously.
  - `[`/`]` move between steps; `]` on last step finalizes.
- Implement `interview_agent.go`:
  - `writeInterviewBrief(step, answer, draft, spec)` to `.praude/briefs/`.
  - `parseAgentDraft(output []byte) string` (trim output; treat full output as draft).
- Merge on finalize: use `effectiveAnswer = draft if present else answer` to update Title, Summary, Requirements, UserStory; preserve other fields from `baseSpec`.
- Update interview UI to show step hints and current draft text.

**Step 3: Run tests to verify pass**
```
go test ./internal/praude/tui -run TestNewKeyStartsInterviewForNewSpec
```

Expected: PASS

**Step 4: Commit**
```
git add internal/praude/tui/model.go internal/praude/tui/interview.go internal/praude/tui/overlay.go internal/praude/specs/create.go internal/praude/tui/interview_agent.go
git commit -m "feat(praude): add new PRD interview iteration"
```

---

### Task 4: Update remaining interview tests and run full suite

**Files:**
- Modify: `internal/praude/tui/interview_test.go`

**Step 1: Update tests**
- Replace `g` with `n` where a new interview is expected.
- Use `]` to advance between text steps.
- Stub `runAgent` where needed to produce deterministic drafts.

**Step 2: Run tests**
```
go test ./internal/praude/tui
```

Expected: PASS

**Step 3: Run full suite**
```
go test ./...
```

Expected: PASS

**Step 4: Commit**
```
git add internal/praude/tui/interview_test.go
git commit -m "test(praude): update interview flow expectations"
```
