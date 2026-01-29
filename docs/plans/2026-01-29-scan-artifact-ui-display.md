# Scan Artifact UI Display Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** [none] (Task reference)

**Goal:** Show scan artifact evidence + quality scores directly in the review panes during scan signoff.

**Architecture:** Extend scan result message to carry phase artifacts, and render an evidence + quality section in the doc panel for the active scan substep. Keep legacy fields as fallback.

**Tech Stack:** Go 1.24, Bubble Tea, lipgloss

---

### Task 1: Plumb structured artifacts to the UI

**Files:**
- Modify: `internal/tui/messages.go`
- Modify: `internal/tui/unified_app.go`
- Modify: `internal/tui/views/kickoff.go`
- Test: `internal/tui/views/kickoff_doc_test.go`

**Step 1: Write failing test**
```go
func TestKickoffDocPanelShowsEvidenceForVision(t *testing.T) {
	v := NewKickoffView()
	_, _ = v.Update(tui.CodebaseScanResultMsg{
		PhaseArtifacts: &tui.PhaseArtifacts{
			Vision: &tui.VisionArtifact{Summary: "Vision", Evidence: []tui.EvidenceItem{{Path:"README.md", Quote:"Autarch"}}},
		},
	})
	view := v.docPanel.Render()
	if !strings.Contains(view, "Evidence") {
		t.Fatalf("expected evidence section")
	}
}
```

**Step 2: Run test to confirm failure**
```bash
GOCACHE=/tmp/gocache go test ./internal/tui/views -run TestKickoffDocPanelShowsEvidenceForVision -v
```
Expected: FAIL

**Step 3: Implement minimal plumbing**
- Add `PhaseArtifacts` + `EvidenceItem` + `VisionArtifact`/`ProblemArtifact`/`UsersArtifact` types to `internal/tui/messages.go` (mirroring agent types).
- Map `agent.ScanResult.PhaseArtifacts` into `CodebaseScanResultMsg` in `internal/tui/unified_app.go`.
- Store `PhaseArtifacts` in `KickoffView` on scan result.

**Step 4: Run test to confirm pass**
```bash
GOCACHE=/tmp/gocache go test ./internal/tui/views -run TestKickoffDocPanelShowsEvidenceForVision -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add internal/tui/messages.go internal/tui/unified_app.go internal/tui/views/kickoff.go internal/tui/views/kickoff_doc_test.go

git commit -m "feat(tui): plumb scan artifacts to kickoff"
```

---

### Task 2: Render evidence + quality for current scan substep

**Files:**
- Modify: `internal/tui/views/kickoff.go`
- Test: `internal/tui/views/kickoff_doc_test.go`

**Step 1: Write failing tests**
- Evidence section visible for Vision/Problem/Users when artifacts present.
- Quality scores show for current step.

**Step 2: Run tests to confirm failure**
```bash
GOCACHE=/tmp/gocache go test ./internal/tui/views -run TestKickoffDocPanelShowsEvidence -v
```
Expected: FAIL

**Step 3: Implement render**
- When `scanReview` and `PhaseArtifacts` for current step exist:
  - Add a `DocSection` titled “Evidence” with bullet list of `path: quote`.
  - Add a `DocSection` titled “Quality” with formatted scores (clarity/completeness/grounding/consistency).
- Fallback to legacy summary text if artifacts absent.

**Step 4: Run tests to confirm pass**
```bash
GOCACHE=/tmp/gocache go test ./internal/tui/views -run TestKickoffDocPanelShowsEvidence -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add internal/tui/views/kickoff.go internal/tui/views/kickoff_doc_test.go

git commit -m "feat(tui): render scan evidence and quality"
```

---

### Task 3: Full test pass

**Files:**
- Test: `internal/tui/views`, `internal/tui`

**Step 1: Run tests**
```bash
GOCACHE=/tmp/gocache go test ./internal/tui/views -v
GOCACHE=/tmp/gocache go test ./internal/tui -v
```
Expected: PASS

---

### Task 4: Commit and push

**Step 1: Final commit (if needed)**
```bash
git status --short
```
If any remaining changes:
```bash
git add internal/tui/messages.go internal/tui/unified_app.go internal/tui/views/kickoff.go internal/tui/views/kickoff_doc_test.go

git commit -m "feat(tui): show scan evidence and quality"
```

**Step 2: Push**
```bash
git push
```
