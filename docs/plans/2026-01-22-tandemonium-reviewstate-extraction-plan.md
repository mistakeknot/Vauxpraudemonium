# Tandemonium ReviewState Extraction Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** `none (no bead in use)`

**Goal:** Extract review-related fields and behavior from the TUI Model into a dedicated ReviewState to reduce model bloat.

**Architecture:** Introduce a `ReviewState` struct in `internal/tandemonium/tui/model.go` (or a new file) holding all review-specific state. Update Model to embed or reference ReviewState and adjust methods to access review fields through this struct. Keep behavior and UI output unchanged.

**Tech Stack:** Go, Bubble Tea TUI.

---

### Task 1: Add failing tests for ReviewState plumbing

**Files:**
- Modify: `internal/tandemonium/tui/approve_key_test.go`
- Modify: `internal/tandemonium/tui/review_actions_test.go`

**Step 1: Write failing test (access via ReviewState)**

```go
func TestReviewStatePlumbsPendingApprove(t *testing.T) {
    m := NewModel()
    m.Review.PendingApproveTask = "TAND-001"
    if m.Review.PendingApproveTask != "TAND-001" {
        t.Fatalf("expected pending approve task")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tandemonium/tui -run TestReviewStatePlumbsPendingApprove`

Expected: FAIL with `Review` undefined on Model.

**Step 3: Add a second failing test (review detail/diff)**

```go
func TestReviewStateHoldsDetailAndDiff(t *testing.T) {
    m := NewModel()
    m.Review.Detail.TaskID = "TAND-002"
    m.Review.Diff.Files = []string{"a.txt"}
    if m.Review.Detail.TaskID != "TAND-002" || len(m.Review.Diff.Files) != 1 {
        t.Fatalf("expected review detail/diff to be stored")
    }
}
```

**Step 4: Run test to verify it fails**

Run: `go test ./internal/tandemonium/tui -run TestReviewStateHoldsDetailAndDiff`

Expected: FAIL with `Review` undefined on Model.

---

### Task 2: Introduce ReviewState struct

**Files:**
- Modify: `internal/tandemonium/tui/model.go`

**Step 1: Add ReviewState struct**

```go
type ReviewState struct {
    Queue           []string
    Branches        map[string]string
    Selected        int
    PendingApprove  string
    ShowDiffs       bool
    InputMode       ReviewInputMode
    Input           string
    PendingReject   bool
    PendingExplain  bool
    Detail          ReviewDetail
    Diff            ReviewDiffState
    Loader          ReviewLoader
    DetailLoader    func(taskID string) (ReviewDetail, error)
    DiffLoader      func(taskID string) (ReviewDiffState, error)
    ActionWriter    func(taskID, text string) error
    StoryUpdater    func(taskID, text string) error
    Rejecter        func(taskID string) error
    ExplainWriter   func(taskID, text string) error
    Acceptor        func(taskID string) error
    Reverter        func(taskID, path string) error
}
```

**Step 2: Add Review field to Model and initialize defaults in NewModel**

```go
type Model struct {
    // ...
    Review ReviewState
    // ...
}
```

Initialize `Review` with default `Branches: map[string]string{}`, `Queue: []string{}`, `Selected: 0`.

**Step 3: Run tests to verify they pass**

Run: `go test ./internal/tandemonium/tui -run TestReviewStatePlumbsPendingApprove`

Expected: PASS.

---

### Task 3: Migrate review-related fields and method usages

**Files:**
- Modify: `internal/tandemonium/tui/model.go`

**Step 1: Replace direct fields with ReviewState fields**

Examples:
- `m.ReviewQueue` → `m.Review.Queue`
- `m.ReviewBranches` → `m.Review.Branches`
- `m.SelectedReview` → `m.Review.Selected`
- `m.PendingApproveTask` → `m.Review.PendingApprove`
- `m.ReviewShowDiffs` → `m.Review.ShowDiffs`
- `m.ReviewInputMode` → `m.Review.InputMode`
- `m.ReviewInput` → `m.Review.Input`
- `m.ReviewPendingReject` → `m.Review.PendingReject`
- `m.MVPExplainPending` → `m.Review.PendingExplain`
- `m.ReviewDetail` → `m.Review.Detail`
- `m.ReviewDiff` → `m.Review.Diff`
- loader/callback fields under `m.Review.*`

**Step 2: Run targeted tests**

Run: `go test ./internal/tandemonium/tui -run TestReviewStateHoldsDetailAndDiff`

Expected: PASS.

---

### Task 4: Run suite, update todo, commit

**Files:**
- Modify: `docs/tandemonium/todos/007-pending-p2-tui-model-bloat.md`

**Step 1: Run relevant tests**

Run: `go test ./internal/tandemonium/tui`

Expected: PASS.

**Step 2: Update todo**

Mark todo as done with note that ReviewState extracted.

**Step 3: Commit**

```bash
git add internal/tandemonium/tui/model.go internal/tandemonium/tui/approve_key_test.go internal/tandemonium/tui/review_actions_test.go docs/tandemonium/todos/007-pending-p2-tui-model-bloat.md docs/plans/2026-01-22-tandemonium-reviewstate-extraction-plan.md
git commit -m "refactor(tandemonium): extract ReviewState"
```

---

Plan complete and saved to `docs/plans/2026-01-22-tandemonium-reviewstate-extraction-plan.md`.

Two execution options:

1. Subagent-Driven (this session) — I dispatch a fresh subagent per task, review between tasks
2. Parallel Session (separate) — Open a new session with executing-plans and batch execution

Which approach?
