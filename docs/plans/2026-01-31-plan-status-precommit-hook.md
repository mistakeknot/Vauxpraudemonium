# Plan Status Pre-Commit Hook Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** `[n/a] (Task reference)`

**Goal:** Generate a single canonical `docs/plans/STATUS.md` on every commit via a pre-commit hook, using a deterministic plan-status generator.

**Architecture:** Add a small internal `planstatus` package that collects plan evidence (todos, mapped paths, git history, derived evidence) and emits a stable markdown report. Expose a new `autarch plan-status` CLI to generate the report. Wire a repo-local pre-commit hook to invoke the command and stage the status file.

**Tech Stack:** Go (cobra CLI, stdlib), git hooks (pre-commit), shell script for hook wrapper.

---

## Task 1: Add plan-status generator package

**Files:**
- Create: `internal/planstatus/report.go`
- Create: `internal/planstatus/report_test.go`

**Step 1: Write the failing test**

```go
func TestReportIncludesDerivedEvidence(t *testing.T) {
    // Build a minimal temp repo with a plans dir + dummy files.
    // Use a stub git-info provider so we can assert deterministic output.
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/planstatus -run TestReportIncludesDerivedEvidence -v`
Expected: FAIL (package not found or missing symbols).

**Step 3: Write minimal implementation**

- Implement a `Generator` that:
  - Scans `docs/plans/` for `*.md` excluding `INDEX.md`, `*-design.md`, `*-audit.md`.
  - Extracts backticked paths + `Create:` paths.
  - Applies legacy path mappings:
    - `internal/vauxhall/` → `internal/bigend/`
    - `internal/praude/` → `internal/gurgeh/`
    - `internal/tandemonium/` → `internal/coldwine/`
    - `cmd/vauxhall/` → `cmd/bigend/`
    - `.praude/` → `.gurgeh/`
    - `.tandemonium/` → `.coldwine/`
  - Loads todos from `todos/*.md` and detects plan references.
  - Uses a `GitInfo` interface to get last commit dates (stubbed in tests).
  - Applies derived evidence for `2026-01-28-feat-coordination-api-foundation-plan.md`.
  - Emits a stable markdown report with counts and a table.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/planstatus -run TestReportIncludesDerivedEvidence -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/planstatus/report.go internal/planstatus/report_test.go
git commit -m "feat(planstatus): add report generator"
```

---

## Task 2: Add `autarch plan-status` command

**Files:**
- Modify: `cmd/autarch/main.go`
- Create: `internal/planstatus/cli.go`
- Create: `internal/planstatus/cli_test.go`

**Step 1: Write the failing test**

```go
func TestPlanStatusCommandWritesFile(t *testing.T) {
    // Run the command with temp output path and assert file exists.
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/planstatus -run TestPlanStatusCommandWritesFile -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

- Add `plan-status` cobra command to `cmd/autarch`.
- Support flags:
  - `--repo` (default: `.`)
  - `--intermute` (default: `/root/projects/Intermute` if exists)
  - `--output` (default: `docs/plans/STATUS.md`)
- Use generator to write the report only when content changed.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/planstatus -run TestPlanStatusCommandWritesFile -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add cmd/autarch/main.go internal/planstatus/cli.go internal/planstatus/cli_test.go
git commit -m "feat(planstatus): add autarch plan-status command"
```

---

## Task 3: Add pre-commit hook to update STATUS.md

**Files:**
- Create: `.githooks/pre-commit`
- Create: `scripts/hooks/install-git-hooks.sh`
- Modify: `docs/QUICK_REFERENCE.md`

**Step 1: Write the failing test**

```bash
# Manual test plan (hook behavior):
# - Modify a plan file or todo.
# - Run pre-commit script directly.
# - Verify docs/plans/STATUS.md updates and is staged.
```

**Step 2: Run test to verify it fails**

Run: `./.githooks/pre-commit` (before file exists)
Expected: FAIL (no hook).

**Step 3: Write minimal implementation**

- `.githooks/pre-commit` runs:
  - `./dev autarch plan-status --output docs/plans/STATUS.md`
  - `git add docs/plans/STATUS.md`
- `scripts/hooks/install-git-hooks.sh` sets:
  - `git config core.hooksPath .githooks`
- Update `docs/QUICK_REFERENCE.md` with setup step.

**Step 4: Run test to verify it passes**

Run: `./.githooks/pre-commit`
Expected: STATUS file updated and staged.

**Step 5: Commit**

```bash
git add .githooks/pre-commit scripts/hooks/install-git-hooks.sh docs/QUICK_REFERENCE.md
git commit -m "feat(planstatus): add pre-commit hook"
```

---

## Task 4: Regenerate canonical status file

**Files:**
- Modify: `docs/plans/STATUS.md`

**Step 1: Generate**

Run: `./dev autarch plan-status --output docs/plans/STATUS.md`
Expected: `docs/plans/STATUS.md` updated.

**Step 2: Commit**

```bash
git add docs/plans/STATUS.md
git commit -m "docs(plans): update status report"
```

---

## Notes

- User requested **no worktrees**; execute in the main working tree.
- Keep report file path stable (`docs/plans/STATUS.md`) as the single source of truth.
- Prefer ASCII text only.

---

Plan complete and saved to `docs/plans/2026-01-31-plan-status-precommit-hook.md`. Two execution options:

1. Subagent-Driven (this session)
2. Parallel Session (separate)

Which approach?
