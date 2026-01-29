# Scan Artifact Validation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Bead:** [none] (Task reference)

**Goal:** Add strict Go validation for scan artifacts (phase + synthesis) with schema files, evidence checks, and quality gates.

**Architecture:** Store JSON schema files in-repo for contract clarity and embed them for reference; implement manual strict decoding + validation gates without new dependencies. Provide validators that return structured errors suitable for UI blocking and pipeline enforcement.

**Tech Stack:** Go 1.24, standard library (encoding/json, io/fs), go:embed

---

### Task 1: Add schema files for scan artifacts

**Files:**
- Create: `internal/autarch/agent/schemas/scan/base.json`
- Create: `internal/autarch/agent/schemas/scan/evidence.json`
- Create: `internal/autarch/agent/schemas/scan/vision.json`
- Create: `internal/autarch/agent/schemas/scan/problem.json`
- Create: `internal/autarch/agent/schemas/scan/users.json`
- Create: `internal/autarch/agent/schemas/scan/features.json`
- Create: `internal/autarch/agent/schemas/scan/requirements.json`
- Create: `internal/autarch/agent/schemas/scan/scope.json`
- Create: `internal/autarch/agent/schemas/scan/cujs.json`
- Create: `internal/autarch/agent/schemas/scan/acceptance.json`
- Create: `internal/autarch/agent/schemas/scan/synthesis.json`

**Step 1: Write schema files**
Use the agreed schemas (base + per‑phase + synthesis) as JSON files.

**Step 2: Commit**
```bash
git add internal/autarch/agent/schemas/scan/*.json

git commit -m "feat(agent): add scan artifact schemas"
```

---

### Task 2: Add validator types and strict decoding helpers

**Files:**
- Create: `internal/autarch/agent/scan_validate.go`
- Create: `internal/autarch/agent/scan_validate_types.go`
- Create: `internal/autarch/agent/scan_validate_embed.go`
- Test: `internal/autarch/agent/scan_validate_test.go`

**Step 1: Write failing tests**
Create tests for:
- schema/shape validation via strict decoding (unknown field → error)
- evidence length < 2 → error
- evidence quote missing in file → error
- confidence below threshold → error
- quality below thresholds requires open_questions

Example test scaffold:
```go
func TestValidatePhaseArtifact_RejectsUnknownField(t *testing.T) {
	input := []byte(`{"phase":"vision","version":"v1","summary":"ok","goals":["g"],"non_goals":[],"evidence":[],"open_questions":[],"quality":{"clarity":1,"completeness":1,"grounding":1,"consistency":1},"extra":"nope"}`)
	res := ValidatePhaseArtifact("vision", input, fakeLookup{})
	if res.OK {
		t.Fatal("expected validation to fail")
	}
}
```

**Step 2: Run tests to confirm failure**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -run TestValidatePhaseArtifact -v
```
Expected: FAIL with validation errors.

**Step 3: Implement minimal validator types**
Define:
- `ValidationError` + `ValidationResult`
- `EvidenceItem` struct
- `QualityScores` struct
- per‑phase structs (VisionArtifact, ProblemArtifact, …) that embed a base struct
- `decodeStrict[T any](raw []byte) (T, error)` using `json.Decoder` + `DisallowUnknownFields`
- `EvidenceLookup` interface with `Exists(path)` and `ContainsQuote(path, quote)`

**Step 4: Run tests to confirm pass**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -run TestValidatePhaseArtifact -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add internal/autarch/agent/scan_validate*.go internal/autarch/agent/scan_validate_test.go

git commit -m "feat(agent): add scan artifact validator"
```

---

### Task 3: Implement phase and synthesis validation gates

**Files:**
- Modify: `internal/autarch/agent/scan_validate.go`
- Test: `internal/autarch/agent/scan_validate_test.go`

**Step 1: Add failing tests for thresholds**
Cover:
- clarity/completeness/grounding/consistency thresholds
- missing open_questions when quality below threshold
- synthesis cross_phase_alignment threshold

**Step 2: Run tests to confirm failure**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -run TestValidatePhaseArtifact_Quality -v
```
Expected: FAIL

**Step 3: Implement gates**
- Evidence checks: min 2 items, confidence >= 0.35, path exists, quote present
- Quality thresholds: clarity >= 0.55, completeness >= 0.55, grounding >= 0.60, consistency >= 0.50
- If any quality fails → require open_questions
- Synthesis: cross_phase_alignment >= 0.60

**Step 4: Run tests to confirm pass**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -run TestValidatePhaseArtifact_Quality -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add internal/autarch/agent/scan_validate.go internal/autarch/agent/scan_validate_test.go

git commit -m "feat(agent): enforce scan artifact quality gates"
```

---

### Task 4: Add embedded schema registry (for reference & future tooling)

**Files:**
- Modify: `internal/autarch/agent/scan_validate_embed.go`
- Test: `internal/autarch/agent/scan_validate_test.go`

**Step 1: Add failing test for schema registry**
Ensure registry can return raw schema bytes by name.

**Step 2: Run test to confirm failure**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -run TestSchemaRegistry -v
```
Expected: FAIL

**Step 3: Implement embed + registry**
Use `//go:embed schemas/scan/*.json` and expose:
- `SchemaFor(phase string) ([]byte, bool)`
- `SynthesisSchema() []byte`

**Step 4: Run tests to confirm pass**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -run TestSchemaRegistry -v
```
Expected: PASS

**Step 5: Commit**
```bash
git add internal/autarch/agent/scan_validate_embed.go internal/autarch/agent/scan_validate_test.go

git commit -m "feat(agent): embed scan schemas"
```

---

### Task 5: Full test pass

**Files:**
- Test: `internal/autarch/agent`

**Step 1: Run tests**
```bash
GOCACHE=/tmp/gocache go test ./internal/autarch/agent -v
```
Expected: PASS

---

### Task 6: Commit and push

**Step 1: Final commit (if needed)**
```bash
git status --short
```
If any remaining changes:
```bash
git add internal/autarch/agent/scan_validate*.go internal/autarch/agent/schemas/scan/*.json

git commit -m "feat(agent): scan artifact validation"
```

**Step 2: Push**
```bash
git push
```
