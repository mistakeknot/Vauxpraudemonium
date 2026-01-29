# Interview Breadcrumb & Scan Step Navigation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** [none] (Task reference)

**Goal:** Collapse scan steps into a single Interview breadcrumb step and prevent scan substep navigation from hijacking the breadcrumb.

**Architecture:** Keep scan Vision/Problem/Users as internal substeps inside the kickoff view while the global onboarding breadcrumb reflects only high-level phases. Ensure scan completion transitions to Interview and scan substep navigation never emits breadcrumb navigation messages.

**Tech Stack:** Go, Bubble Tea, lipgloss

### Task 1: Add/adjust failing tests for breadcrumb + scan navigation

**Files:**
- Modify: `internal/tui/breadcrumb_test.go`
- Modify: `internal/tui/unified_app_test.go`
- Add: `internal/tui/views/kickoff_chat_test.go`

**Step 1: Write/confirm failing tests**

- Breadcrumb excludes Vision/Problem/Users and includes Interview.
- Codebase scan result sets onboarding state to Interview.
- Kickoff ctrl+right does not emit NavigateToStepMsg during scan review.

```go
func TestKickoffAcceptDoesNotNavigateBreadcrumb(t *testing.T) {
	v := NewKickoffView()
	v.scanResult = &tui.CodebaseScanResultMsg{Vision: "Vision text"}
	v.SetScanStepForTest(tui.OnboardingScanVision)
	v.SetScanCodebaseCallback(nil)

	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyCtrlRight})
	if cmd == nil {
		return
	}
	if msg := cmd(); msg != nil {
		if _, ok := msg.(tui.NavigateToStepMsg); ok {
			t.Fatalf("did not expect NavigateToStepMsg during scan review")
		}
	}
}
```

**Step 2: Run tests to confirm failure**

Run:
```bash
GOCACHE=/tmp/gocache go test ./internal/tui -run "TestBreadcrumbDoesNotIncludeScanSteps|TestScanResultSetsInterviewBreadcrumb|TestKickoffAcceptDoesNotNavigateBreadcrumb" -v
```
Expected: FAIL for breadcrumb labels, scan result state, or NavigateToStepMsg emission.

### Task 2: Update onboarding states and scan result handling

**Files:**
- Modify: `internal/tui/onboarding.go`
- Modify: `internal/tui/unified_app.go`

**Step 1: Remove scan states from breadcrumb list**

```go
func AllOnboardingStates() []OnboardingState {
	return []OnboardingState{
		OnboardingKickoff,
		OnboardingInterview,
		OnboardingSpecSummary,
		OnboardingEpicReview,
		OnboardingTaskReview,
		OnboardingComplete,
	}
}
```

**Step 2: Set interview as the active breadcrumb state after scan**

```go
case CodebaseScanResultMsg:
	if a.mode == ModeOnboarding {
		a.onboardingState = OnboardingInterview
		a.breadcrumb.SetCurrent(OnboardingInterview)
	}
```

**Step 3: Normalize onboarding header to reflect Interview**

```go
case OnboardingScanVision, OnboardingScanProblem, OnboardingScanUsers, OnboardingInterview:
	return "Project Setup"
```

**Step 4: Run tests**

```bash
GOCACHE=/tmp/gocache go test ./internal/tui -run "TestBreadcrumbDoesNotIncludeScanSteps|TestScanResultSetsInterviewBreadcrumb" -v
```
Expected: PASS

### Task 3: Stop scan substep navigation from touching breadcrumb

**Files:**
- Modify: `internal/tui/views/kickoff.go`

**Step 1: Remove NavigateToStepMsg emission during scan review**

- In `acceptScanStep`, drop the `NavigateToStepMsg` command.
- In `moveScanStepBack`, do not emit `NavigateToStepMsg`.

**Step 2: Run tests**

```bash
GOCACHE=/tmp/gocache go test ./internal/tui -run "TestKickoffAcceptDoesNotNavigateBreadcrumb" -v
```
Expected: PASS

### Task 4: Full test pass

**Files:**
- Test: `internal/tui`, `internal/tui/views`

**Step 1: Run full target tests**

```bash
GOCACHE=/tmp/gocache go test ./internal/tui -v
GOCACHE=/tmp/gocache go test ./internal/tui/views -v
```
Expected: PASS

### Task 5: Commit and push

**Step 1: Commit**

```bash
git add internal/tui/onboarding.go internal/tui/unified_app.go internal/tui/breadcrumb_test.go internal/tui/unified_app_test.go internal/tui/views/kickoff.go internal/tui/views/kickoff_chat_test.go docs/plans/2026-01-29-interview-breadcrumb-scan-nav.md

git commit -m "feat(tui): collapse scan steps into interview"
```

**Step 2: Push**

```bash
git push
```
