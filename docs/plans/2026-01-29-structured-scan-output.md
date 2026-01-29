# Structured Scan Output Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** [none] (Task reference)

**Goal:** Update the code scan prompt and parsing to produce strictly structured phase artifacts with evidence and quality scores.

**Architecture:** Expand the scan agent prompt to output a JSON envelope containing per‑phase artifacts (vision/problem/users) with evidence + quality. Parse into new types, validate via existing validators, and keep legacy fields for backward compatibility.

**Tech Stack:** Go 1.24

---

### Task 1: Add structured scan response types

**Files:**
- Modify: `internal/autarch/agent/scan.go`
- Create: `internal/autarch/agent/scan_structured.go`
- Test: `internal/autarch/agent/scan_structured_test.go`

**Step 1: Write failing test**
Create a test that parses a structured scan JSON envelope and populates `ScanResult` artifacts.

```go
func TestParseStructuredScanResponse(t *testing.T) {
	content := `{
  "project_name": "Autarch",
  "description": "...",
  "artifacts": {
    "vision": { ... },
    "problem": { ... },
    "users": { ... }
  }
}`
	res, err := parseScanResponse(content)
	if err != nil { t.Fatal(err) }
	if res.PhaseArtifacts == nil { t.Fatal("expected artifacts") }
}
```

**Step 2: Run test to confirm failure**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -run TestParseStructuredScanResponse -v
```
Expected: FAIL

**Step 3: Implement minimal types + parse path**
- Add `PhaseArtifacts` field to `ScanResult` (map or typed struct).
- Implement `structuredScanResponse` type and parse in `parseScanResponse`.
- Preserve existing fields for backward compatibility.

**Step 4: Run test to confirm pass**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -run TestParseStructuredScanResponse -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add internal/autarch/agent/scan.go internal/autarch/agent/scan_structured.go internal/autarch/agent/scan_structured_test.go

git commit -m "feat(agent): parse structured scan artifacts"
```

---

### Task 2: Update scan prompt to request structured artifacts

**Files:**
- Modify: `internal/autarch/agent/scan.go`
- Test: `internal/autarch/agent/scan_structured_test.go`

**Step 1: Write failing test**
Add a test asserting the prompt contains the new JSON envelope and evidence requirements.

**Step 2: Run test to confirm failure**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -run TestScanPromptRequestsStructuredArtifacts -v
```
Expected: FAIL

**Step 3: Implement prompt update**
- Output format: `{ project_name, description, artifacts: { vision, problem, users } }`
- Require each artifact to include evidence + quality.

**Step 4: Run test to confirm pass**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -run TestScanPromptRequestsStructuredArtifacts -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add internal/autarch/agent/scan.go internal/autarch/agent/scan_structured_test.go

git commit -m "feat(agent): request structured scan artifacts"
```

---

### Task 3: Validate structured artifacts and map to UI fields

**Files:**
- Modify: `internal/autarch/agent/scan.go`
- Modify: `internal/autarch/agent/scan_validate_legacy.go`
- Test: `internal/autarch/agent/scan_structured_test.go`

**Step 1: Write failing test**
Validate that structured artifacts are passed through `ValidatePhaseArtifact` and errors surfaced.

**Step 2: Run test to confirm failure**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -run TestStructuredScanValidation -v
```
Expected: FAIL

**Step 3: Implement validation wiring**
- If structured artifacts present, validate each using `ValidatePhaseArtifact`.
- Map vision/problem/users summaries to existing `ScanResult` fields for UI compatibility.
- Fall back to legacy validation if artifacts missing.

**Step 4: Run test to confirm pass**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -run TestStructuredScanValidation -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add internal/autarch/agent/scan.go internal/autarch/agent/scan_validate_legacy.go internal/autarch/agent/scan_structured_test.go

git commit -m "feat(agent): validate structured scan artifacts"
```

---

### Task 4: Full test pass

**Files:**
- Test: `internal/autarch/agent`

**Step 1: Run tests**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -v
```
Expected: PASS

---

### Task 5: Commit and push

**Step 1: Final commit (if needed)**
```bash
git status --short
```
If any remaining changes:
```bash
git add internal/autarch/agent/scan.go internal/autarch/agent/scan_structured.go internal/autarch/agent/scan_structured_test.go

git commit -m "feat(agent): structured scan artifacts"
```

**Step 2: Push**
```bash
git push
```
