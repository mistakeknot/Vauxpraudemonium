# Tandemonium RejectTask Ready-State Fix Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** `none (no bead in use)`

**Goal:** Remove the redundant "rejected" status update so RejectTask only returns tasks to "ready" and clears the review queue.

**Architecture:** Keep the existing transaction in `RejectTask`, but eliminate the intermediate status update. Add a regression test that fails if `RejectTask` attempts to set status to "rejected" by installing a trigger in a temp SQLite DB.

**Tech Stack:** Go, SQLite, Go testing.

---

### Task 1: Add failing regression test

**Files:**
- Modify: `internal/tandemonium/storage/review_test.go`

**Step 1: Write the failing test (trigger blocks rejected status)**

```go
func TestRejectTaskDoesNotSetRejected(t *testing.T) {
    db, err := OpenTemp()
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()
    if err := Migrate(db); err != nil {
        t.Fatal(err)
    }
    if err := InsertTask(db, Task{ID: "TAND-004", Title: "Test", Status: "review"}); err != nil {
        t.Fatal(err)
    }
    if err := AddToReviewQueue(db, "TAND-004"); err != nil {
        t.Fatal(err)
    }
    if _, err := db.Exec(`
CREATE TRIGGER reject_status_block
BEFORE UPDATE ON tasks
WHEN NEW.status = 'rejected'
BEGIN
    SELECT RAISE(FAIL, 'rejected status not allowed');
END;
`); err != nil {
        t.Fatal(err)
    }
    if err := RejectTask(db, "TAND-004"); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tandemonium/storage -run TestRejectTaskDoesNotSetRejected`

Expected: FAIL with `unexpected error: rejected status not allowed` (or similar).

---

### Task 2: Remove redundant status update

**Files:**
- Modify: `internal/tandemonium/storage/review.go`

**Step 1: Update RejectTask to only set "ready"**

```go
func RejectTask(db *sql.DB, id string) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    if _, err := tx.Exec(`UPDATE tasks SET status = ? WHERE id = ?`, "ready", id); err != nil {
        _ = tx.Rollback()
        return err
    }
    if _, err := tx.Exec(`DELETE FROM review_queue WHERE task_id = ?`, id); err != nil {
        _ = tx.Rollback()
        return err
    }
    return tx.Commit()
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./internal/tandemonium/storage -run TestRejectTaskDoesNotSetRejected`

Expected: PASS.

---

### Task 3: Verify, update todo, commit

**Files:**
- Modify: `docs/tandemonium/todos/008-pending-p2-reject-task-bug.md`

**Step 1: Run storage tests**

Run: `go test ./internal/tandemonium/storage`

Expected: PASS.

**Step 2: (Optional) Run full test suite**

Run: `go test ./...`

Expected: PASS.

**Step 3: Update todo**

- Set `status: done`
- Check off acceptance criteria
- Add a work log entry dated 2026-01-22 noting the redundant update removal + regression test

**Step 4: Commit**

```bash
git add internal/tandemonium/storage/review.go internal/tandemonium/storage/review_test.go docs/tandemonium/todos/008-pending-p2-reject-task-bug.md docs/plans/2026-01-22-tandemonium-reject-task-fix-plan.md
git commit -m "fix(tandemonium): drop redundant reject status update"
```

---

Plan complete and saved to `docs/plans/2026-01-22-tandemonium-reject-task-fix-plan.md`.

Two execution options:

1. Subagent-Driven (this session) — I dispatch a fresh subagent per task, review between tasks
2. Parallel Session (separate) — Open a new session with executing-plans and batch execution

Which approach?
