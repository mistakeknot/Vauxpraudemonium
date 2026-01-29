# Scan Signoff Breadcrumb Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** n/a (no bead provided)

**Goal:** Add a persistent left-nav sidebar across all views, keep scan activity and the scanned description in the left document pane, and add full Arbiter phase scan signoff steps with sequential approval plus resuggest after each approval.

**Architecture:** Use ShellLayout (sidebar + doc + chat) in all views. Introduce a shared interview substep list (full Arbiter phases) rendered in the sidebar for both kickoff scan review and interview flows. Extend onboarding states to include scan-review steps that appear in the top breadcrumb at app launch. The Kickoff view keeps scan progress and the scanned content in the left doc pane without a “Scan Results” header. When the user approves a step, trigger a resuggest scan call that refines remaining fields, then advance to the next breadcrumb step; enforce sequential approval (no skipping ahead).

**Tech Stack:** Go, Bubble Tea, lipgloss

## Enhancement Summary

**Deepened on:** 2026-01-29  
**Sections enhanced:** 4  
**Skills applied:** agent-native-architecture (product implications)  
**Project learnings reviewed:** `docs/solutions/ui-bugs/tui-breadcrumb-hidden-by-oversized-child-view-20260127.md`, `docs/solutions/ui-bugs/tui-dimension-mismatch-splitlayout-20260126.md`

### Key Improvements
1. Add a persistent left-nav sidebar to all views so substeps live outside the header breadcrumb.
2. Add explicit breadcrumb-state handling to prevent hidden header regressions when scan steps are inserted.
3. Keep scan activity in the left pane while preserving stable sizing to avoid layout overflow.
4. Define approval semantics (per-step signoff + resuggest) using agent-native approval patterns and clear reversibility across all Arbiter phases.
5. Increase sidebar width to avoid truncating Arbiter phase labels.

### New Considerations Discovered
- Breadcrumb visibility is sensitive to child view sizing; scan steps must not reintroduce oversized views.
- Split layout width/height must stay consistent with padded root view to avoid stray borders.
- Kickoff/scan and interview must share the same substep list in the sidebar (full Arbiter phase list).
- Sidebar width must be increased so labels like “Critical User Journeys” render without ellipsis.

### Task 1: Add shared Arbiter phase substeps + sidebar for kickoff and interview

**Files:**
- Modify: `internal/tui/onboarding.go`
- Modify: `internal/tui/views/kickoff.go`
- Modify: `internal/gurgeh/arbiter/tui/arbiter_view.go`
- Modify: `pkg/tui/sidebar.go`
- Test: `internal/tui/views/kickoff_doc_test.go`
- Test: `internal/gurgeh/arbiter/tui/arbiter_view_test.go`

### Research Insights

**UI/UX Considerations:**
- Keep the sidebar list identical for kickoff scan review and interview (full Arbiter phases) so the mental model does not change after scan.
- Use clear status icons: current step (●), accepted (✓), pending (○).

**Edge Cases:**
- If the Arbiter view is not wired in, the sidebar still renders in kickoff.
- Sidebar focus should not steal input from the chat composer unless the user tabs into it.

**Step 1: Write the failing test**

Add to `internal/tui/views/kickoff_doc_test.go`:

```go
func TestKickoffSidebarUsesInterviewSteps(t *testing.T) {
	v := NewKickoffView()
	items := v.SidebarItems()
	if len(items) != 8 {
		t.Fatalf("expected 8 sidebar items, got %d", len(items))
	}
	if items[0].Label != "Vision" ||
		items[1].Label != "Problem" ||
		items[2].Label != "Users" ||
		items[3].Label != "Features + Goals" ||
		items[4].Label != "Requirements" ||
		items[5].Label != "Scope + Assumptions" ||
		items[6].Label != "Critical User Journeys" ||
		items[7].Label != "Acceptance Criteria" {
		t.Fatalf("unexpected sidebar labels: %#v", items)
	}
}
```

Add to `internal/gurgeh/arbiter/tui/arbiter_view_test.go`:

```go
func TestArbiterSidebarUsesInterviewSteps(t *testing.T) {
	view := NewArbiterView("/tmp/test", nil)
	items := view.SidebarItems()
	if len(items) != 8 {
		t.Fatalf("expected 8 sidebar items, got %d", len(items))
	}
	if items[0].Label != "Vision" ||
		items[1].Label != "Problem" ||
		items[2].Label != "Users" ||
		items[3].Label != "Features + Goals" ||
		items[4].Label != "Requirements" ||
		items[5].Label != "Scope + Assumptions" ||
		items[6].Label != "Critical User Journeys" ||
		items[7].Label != "Acceptance Criteria" {
		t.Fatalf("unexpected sidebar labels: %#v", items)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui/views -run TestKickoffSidebarUsesInterviewSteps -v`
Expected: FAIL (SidebarItems not implemented).

Run: `GOCACHE=/tmp/go-cache go test ./internal/gurgeh/arbiter/tui -run TestArbiterSidebarUsesInterviewSteps -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

- Add a shared interview substep helper in `internal/tui/onboarding.go`, for example:
  ```go
  type InterviewStep struct {
  	ID    string
  	Label string
  }

  func InterviewSteps() []InterviewStep {
  	return []InterviewStep{
  		{ID: "vision", Label: "Vision"},
  		{ID: "problem", Label: "Problem"},
  		{ID: "users", Label: "Users"},
  		{ID: "features", Label: "Features + Goals"},
  		{ID: "requirements", Label: "Requirements"},
  		{ID: "scope", Label: "Scope + Assumptions"},
  		{ID: "cujs", Label: "Critical User Journeys"},
  		{ID: "acceptance", Label: "Acceptance Criteria"},
  	}
  }
  ```
- Convert `KickoffView` to use `pkgtui.ShellLayout` (replace `SplitLayout`) and implement `SidebarItems()` to build items from `InterviewSteps()`.
- Convert `ArbiterView` to use `pkgtui.ShellLayout` and implement `SidebarItems()` using the same `InterviewSteps()` list.
- Map current step to the active icon and accepted steps to ✓.
- Increase `SidebarWidth` to 28 and `MaxLabelWidth` to 25 to fit “Critical User Journeys” without truncation.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui/views -run TestKickoffSidebarUsesInterviewSteps -v`
Expected: PASS.

Run: `GOCACHE=/tmp/go-cache go test ./internal/gurgeh/arbiter/tui -run TestArbiterSidebarUsesInterviewSteps -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tui/onboarding.go internal/tui/views/kickoff.go internal/gurgeh/arbiter/tui/arbiter_view.go internal/tui/views/kickoff_doc_test.go internal/gurgeh/arbiter/tui/arbiter_view_test.go
git commit -m "feat(tui): add shared interview sidebar steps"
```

### Task 2: Enable sidebar for review + summary views

**Files:**
- Modify: `internal/tui/views/spec_summary.go`
- Modify: `internal/tui/views/epic_review.go`
- Modify: `internal/tui/views/task_review.go`
- Modify: `internal/tui/views/task_detail.go`
- Test: `internal/tui/views/views_test.go`

### Research Insights

**Best Practices:**
- Each review view should provide `SidebarItems()` so ShellLayout can render a stable left nav.
- Keep sidebar items short to avoid truncation at 20 chars.

**Step 1: Write the failing test**

Update `internal/tui/views/views_test.go`:

```go
func TestReviewViewsProvideSidebarItems(t *testing.T) {
	if len(NewSpecSummaryView(nil, nil).SidebarItems()) == 0 {
		t.Fatalf("expected spec summary sidebar items")
	}
	if len(NewEpicReviewView(nil).SidebarItems()) == 0 {
		t.Fatalf("expected epic review sidebar items")
	}
	if len(NewTaskReviewView(nil).SidebarItems()) == 0 {
		t.Fatalf("expected task review sidebar items")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui/views -run TestReviewViewsProvideSidebarItems -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

- Implement `SidebarItems()` in `SpecSummaryView` using section labels (Vision, Problem, Users, Platform, Language, Requirements, Research).
- Implement `SidebarItems()` in `EpicReviewView` using epic titles.
- Implement `SidebarItems()` in `TaskReviewView` using task titles.
- Implement `SidebarItems()` in `TaskDetailView` using fixed labels (Overview, Acceptance, Notes).
- Replace `RenderWithoutSidebar` calls with `shell.Render(v.SidebarItems(), document, chat)`.
- Remove or update `TestReviewViewsUseRenderWithoutSidebar` to reflect sidebar usage.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui/views -run TestReviewViewsProvideSidebarItems -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tui/views/spec_summary.go internal/tui/views/epic_review.go internal/tui/views/task_review.go internal/tui/views/task_detail.go internal/tui/views/views_test.go
git commit -m "feat(tui): add sidebar items to review views"
```

### Task 3: Add onboarding states for scan signoff steps

**Files:**
- Modify: `internal/tui/onboarding.go`
- Modify: `internal/tui/breadcrumb.go`
- Test: `internal/tui/breadcrumb_test.go`

### Research Insights

**Best Practices:**
- Keep breadcrumb steps derived from a single source of truth (`AllOnboardingStates`) to avoid mismatched labels or ordering.
- Add a small test-only helper (e.g., `LabelsForTest`) to avoid using rendering paths in tests.

**UI/UX Considerations:**
- Ensure new breadcrumb steps don’t increase header height or cause wrap; keep labels short (Vision/Users/Problem).

**Edge Cases:**
- Breadcrumb visibility can disappear if child views receive full terminal height and overflow. Confirm `WindowSizeMsg` handling is consistent with root padding.

**References:**
- `docs/solutions/ui-bugs/tui-breadcrumb-hidden-by-oversized-child-view-20260127.md`

**Step 1: Write the failing test**

Add to `internal/tui/breadcrumb_test.go`:

```go
func TestBreadcrumbIncludesScanSteps(t *testing.T) {
	b := NewBreadcrumb()
	labels := b.LabelsForTest()
	want := []string{"Project", "Vision", "Users", "Problem"}
	for _, w := range want {
		found := false
		for _, label := range labels {
			if label == w {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected breadcrumb to include %q", w)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui -run TestBreadcrumbIncludesScanSteps -v`
Expected: FAIL (labels missing).

**Step 3: Write minimal implementation**

- Add new `OnboardingState` values: `OnboardingScanVision`, `OnboardingScanUsers`, `OnboardingScanProblem`.
- Insert them after `OnboardingKickoff` in `AllOnboardingStates()`.
- Add `ID()`/`Label()` mappings: Vision/Users/Problem.
- Update breadcrumb to expose `LabelsForTest()` if missing (test helper only).

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui -run TestBreadcrumbIncludesScanSteps -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tui/onboarding.go internal/tui/breadcrumb.go internal/tui/breadcrumb_test.go
git commit -m "feat(tui): add scan signoff breadcrumb steps"
```

### Task 4: Keep scan activity in left pane and show description there

**Files:**
- Modify: `internal/tui/views/kickoff.go`
- Test: `internal/tui/views/kickoff_doc_test.go`

### Research Insights

**Best Practices:**
- Keep the left document pane stable during scan to preserve breadcrumb visibility and avoid layout shifts.
- Render scan progress (e.g., “Preparing…”, “Analyzing…”) in the left pane; reserve the chat panel for agent conversation.

**Performance Considerations:**
- Avoid re-rendering overly long scan output in the doc pane; keep scan activity concise to reduce layout churn.

**Edge Cases:**
- SplitLayout artifacts appear if child views render at full terminal size rather than padded content size—do not replace the left pane with scan progress during load.

**References:**
- `docs/solutions/ui-bugs/tui-dimension-mismatch-splitlayout-20260126.md`

**Step 1: Write the failing test**

Add to `internal/tui/views/kickoff_doc_test.go`:

```go
func TestKickoffScanDescriptionRendersWithoutHeader(t *testing.T) {
	v := NewKickoffView()
	v.docPanel.SetSize(80, 20)

	_, _ = v.Update(tui.CodebaseScanResultMsg{Description: "Scanned description"})
	view := v.docPanel.View()
	if !strings.Contains(view, "Scanned description") {
		t.Fatalf("expected description in doc panel")
	}
	if strings.Contains(view, "Scan Results") {
		t.Fatalf("did not expect scan results header")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui/views -run TestKickoffScanDescriptionRendersWithoutHeader -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

- In `KickoffView.View()`, keep the left pane as the scan progress pane during loading (scan activity stays left).
- In `KickoffView.updateDocPanel()`, when `scanResult != nil`, render the description (and optional tech info) without a “Scan Results” header.
- Ensure scan progress lines and step details remain in the left pane; do not inject them into the chat composer.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui/views -run TestKickoffScanDescriptionRendersWithoutHeader -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tui/views/kickoff.go internal/tui/views/kickoff_doc_test.go
git commit -m "fix(tui): keep scan activity in agent panel"
```

### Task 5: Add scan signoff flow with per-step approval and resuggest

**Files:**
- Modify: `internal/tui/messages.go`
- Modify: `internal/tui/unified_app.go`
- Modify: `internal/tui/views/kickoff.go`
- Modify: `internal/autarch/agent/scan.go`
- Test: `internal/tui/views/kickoff_chat_test.go`

### Research Insights

**Best Practices:**
- Use explicit per-step approval (Enter to accept) with resuggest after each step; treat this as “suggest + apply” since it’s easy to revise.
- Make the resuggest loop explicit in chat (“Updating remaining sections…”) so users understand the agent is refining subsequent steps.

**Execution Pattern:**
- After a step is accepted, re-run scan with accepted fields pinned and only update the remaining fields (avoid overwriting confirmed text).
- Keep a minimal in-memory “accepted” map to feed the resuggest prompt.
 - Enforce sequential approval in breadcrumb navigation (disable jumps to future scan steps).

**Edge Cases:**
- Agent missing or resuggest fails → keep user in the same step and surface error in chat.
- Empty resuggest fields → keep prior values and annotate as “unchanged.”

**References:**
- `/root/.codex/skills/agent-native-architecture/references/product-implications.md`

**Step 1: Write the failing test**

Add to `internal/tui/views/kickoff_chat_test.go`:

```go
func TestKickoffAcceptsVisionStepAndAdvances(t *testing.T) {
	v := NewKickoffView()
	v.scanResult = &tui.CodebaseScanResultMsg{Vision: "V"}
	v.SetScanStepForTest(tui.OnboardingScanVision)

	_, _ = v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if v.ScanStepForTest() != tui.OnboardingScanUsers {
		t.Fatalf("expected step advance to users")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui/views -run TestKickoffAcceptsVisionStepAndAdvances -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

- Add scan-review state to `KickoffView`: current step (Vision/Users/Problem), accepted values, and a boolean for “in scan review mode.”
- Add a message and callback to request resuggest: `ScanResuggestMsg` and `onResuggest` callback.
- On Enter in scan review mode, mark current step accepted, emit resuggest command via callback, and advance breadcrumb state (Vision → Users → Problem).
- In `UnifiedApp`, handle new states in navigation and set the kickoff view’s active scan step when state changes.
 - In breadcrumb navigation, disallow jumping to scan steps beyond the current approved step.
- In `agent/scan.go`, add `ResuggestScanSummaryWithOutput(ctx, agent, path, accepted, onOutput)` that uses the existing scan prompt plus accepted fields to refine remaining fields.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui/views -run TestKickoffAcceptsVisionStepAndAdvances -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tui/messages.go internal/tui/unified_app.go internal/tui/views/kickoff.go internal/autarch/agent/scan.go internal/tui/views/kickoff_chat_test.go
git commit -m "feat(tui): add scan signoff flow with resuggest"
```

### Task 6: Transition from scan signoff to spec flow

**Files:**
- Modify: `internal/tui/unified_app.go`
- Modify: `internal/tui/views/kickoff.go`
- Test: `internal/tui/unified_app_test.go`

### Research Insights

**Best Practices:**
- When Problem is accepted, advance to Spec Summary with the accumulated, accepted values only.
- Ensure breadcrumb moves to Spec Summary explicitly to avoid orphaning the new scan steps.

**Edge Cases:**
- If only some fields are accepted, fill the rest from the latest resuggested values before entering Spec Summary.
- Maintain parity: user can still jump back to scan steps via breadcrumb without losing accepted content.

**Step 1: Write the failing test**

Add to `internal/tui/unified_app_test.go`:

```go
func TestScanSignoffCompletesToSpecSummary(t *testing.T) {
	app := NewUnifiedApp(nil)
	app.onboardingState = OnboardingScanProblem
	app.generatedEpics = nil

	_, _ = app.Update(ScanSignoffCompleteMsg{})
	if app.onboardingState != OnboardingSpecSummary {
		t.Fatalf("expected to enter spec summary after scan signoff")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui -run TestScanSignoffCompletesToSpecSummary -v`
Expected: FAIL.

**Step 3: Write minimal implementation**

- Emit `ScanSignoffCompleteMsg` when Problem step is accepted.
- In `UnifiedApp`, map accepted scan values into `SpecAcceptedMsg` or `InterviewCompleteMsg` to create the spec summary view.
- Ensure breadcrumb moves to `OnboardingSpecSummary`.

**Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/go-cache go test ./internal/tui -run TestScanSignoffCompletesToSpecSummary -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/tui/unified_app.go internal/tui/views/kickoff.go internal/tui/unified_app_test.go
git commit -m "feat(tui): advance from scan signoff to spec summary"
```
