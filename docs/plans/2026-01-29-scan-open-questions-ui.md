# Scan Open Questions UI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** [none] (Task reference)

**Goal:** Show Open Questions for the active scan substep in the review pane.

**Architecture:** Extend the scan review doc panel to render a new “Open Questions” section when the current phase artifact includes open questions. This uses the existing phase artifacts already wired to the UI.

**Tech Stack:** Go 1.24, Bubble Tea

---

### Task 1: Add Open Questions rendering

**Files:**
- Modify: `internal/tui/views/kickoff.go`
- Test: `internal/tui/views/kickoff_doc_test.go`

**Step 1: Write failing test**
```go
func TestKickoffDocPanelShowsOpenQuestions(t *testing.T) {
	v := NewKickoffView()
	v.docPanel.SetSize(80, 30)
	_, _ = v.Update(tui.CodebaseScanResultMsg{
		PhaseArtifacts: &tui.PhaseArtifacts{
			Vision: &tui.VisionArtifact{
				Summary: "Vision text",
				OpenQuestions: []string{"What is the primary goal?"},
			},
		},
	})
	view := v.docPanel.View()
	if !strings.Contains(view, "Open Questions") {
		t.Fatalf("expected open questions section")
	}
}
```

**Step 2: Run test to confirm failure**
```bash
GOCACHE=/tmp/gocache go test ./internal/tui/views -run TestKickoffDocPanelShowsOpenQuestions -v
```
Expected: FAIL

**Step 3: Implement minimal rendering**
- In `addScanEvidenceSections`, after Evidence/Quality, check `artifact.OpenQuestions`.
- Render a `DocSection` titled “Open Questions” with bullet list.

**Step 4: Run test to confirm pass**
```bash
GOCACHE=/tmp/gocache go test ./internal/tui/views -run TestKickoffDocPanelShowsOpenQuestions -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add internal/tui/views/kickoff.go internal/tui/views/kickoff_doc_test.go

git commit -m "feat(tui): show scan open questions"
```

---

### Task 2: Full test pass

**Files:**
- Test: `internal/tui/views`, `internal/tui`

**Step 1: Run tests**
```bash
GOCACHE=/tmp/gocache go test ./internal/tui/views -v
GOCACHE=/tmp/gocache go test ./internal/tui -v
```
Expected: PASS

---

### Task 3: Commit and push

**Step 1: Final commit (if needed)**
```bash
git status --short
```
If any remaining changes:
```bash
git add internal/tui/views/kickoff.go internal/tui/views/kickoff_doc_test.go

git commit -m "feat(tui): show scan open questions"
```

**Step 2: Push**
```bash
git push
```
