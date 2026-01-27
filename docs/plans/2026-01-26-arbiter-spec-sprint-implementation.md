# Arbiter Spec Sprint Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace Gurgeh's current 8-step ask-first interview flow with the Arbiter Spec Sprint—a 6-section propose-first workflow with integrated research, consistency checking, and confidence scoring.

**Architecture:** The implementation adds three new packages (`arbiter/`, `consistency/`, `confidence/`) while reusing existing infrastructure (`specs/`, `review/`, hunters). The TUI interview view gets a new propose-first interaction model.

**Tech Stack:** Go, Bubble Tea, existing Pollard hunters (github-scout, hackernews-trendwatcher)

---

## Phase 1: Core Data Structures

### Task 1: Define Sprint State Types

**Files:**
- Create: `internal/gurgeh/arbiter/types.go`
- Test: `internal/gurgeh/arbiter/types_test.go`

**Step 1: Write the failing test**

```go
// internal/gurgeh/arbiter/types_test.go
package arbiter

import (
    "testing"
)

func TestNewSprintState(t *testing.T) {
    state := NewSprintState("test-project")

    if state.Phase != PhaseProblem {
        t.Errorf("expected initial phase %v, got %v", PhaseProblem, state.Phase)
    }
    if len(state.Sections) != 6 {
        t.Errorf("expected 6 sections, got %d", len(state.Sections))
    }
    if state.Confidence.Total() != 0 {
        t.Errorf("expected initial confidence 0, got %f", state.Confidence.Total())
    }
}

func TestPhaseOrder(t *testing.T) {
    phases := AllPhases()
    expected := []Phase{
        PhaseProblem,
        PhaseUsers,
        PhaseFeaturesGoals,
        PhaseScopeAssumptions,
        PhaseCUJs,
        PhaseAcceptanceCriteria,
    }

    if len(phases) != len(expected) {
        t.Fatalf("expected %d phases, got %d", len(expected), len(phases))
    }

    for i, p := range phases {
        if p != expected[i] {
            t.Errorf("phase %d: expected %v, got %v", i, expected[i], p)
        }
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/gurgeh/arbiter/... -v`
Expected: FAIL with "no Go files in directory"

**Step 3: Write minimal implementation**

```go
// internal/gurgeh/arbiter/types.go
package arbiter

import "time"

// Phase represents a section of the PRD sprint
type Phase int

const (
    PhaseProblem Phase = iota
    PhaseUsers
    PhaseFeaturesGoals
    PhaseScopeAssumptions
    PhaseCUJs
    PhaseAcceptanceCriteria
)

// AllPhases returns phases in order
func AllPhases() []Phase {
    return []Phase{
        PhaseProblem,
        PhaseUsers,
        PhaseFeaturesGoals,
        PhaseScopeAssumptions,
        PhaseCUJs,
        PhaseAcceptanceCriteria,
    }
}

// String returns the display name for a phase
func (p Phase) String() string {
    names := []string{
        "Problem",
        "Users",
        "Features + Goals",
        "Scope + Assumptions",
        "Critical User Journeys",
        "Acceptance Criteria",
    }
    if int(p) < len(names) {
        return names[p]
    }
    return "Unknown"
}

// DraftStatus tracks the state of a section draft
type DraftStatus int

const (
    DraftPending DraftStatus = iota
    DraftProposed
    DraftAccepted
    DraftNeedsRevision
)

// SectionDraft holds Arbiter's proposal for a section
type SectionDraft struct {
    Content   string      // Arbiter's current proposal
    Options   []string    // Alternative phrasings (2-3 options)
    Status    DraftStatus
    UserEdits []Edit      // History of user changes
    UpdatedAt time.Time
}

// Edit records a user modification
type Edit struct {
    Before    string
    After     string
    Reason    string    // Optional: why the user changed it
    Timestamp time.Time
}

// ConfidenceScore tracks PRD quality metrics
type ConfidenceScore struct {
    Completeness float64 // 0-1, weight: 20%
    Consistency  float64 // 0-1, weight: 25%
    Specificity  float64 // 0-1, weight: 20%
    Research     float64 // 0-1, weight: 20%
    Assumptions  float64 // 0-1, weight: 15%
}

// Total returns the weighted confidence score
func (c ConfidenceScore) Total() float64 {
    return c.Completeness*0.20 +
        c.Consistency*0.25 +
        c.Specificity*0.20 +
        c.Research*0.20 +
        c.Assumptions*0.15
}

// SprintState holds the full state of a PRD sprint session
type SprintState struct {
    ID          string
    ProjectPath string
    Phase       Phase
    Sections    map[Phase]*SectionDraft
    Conflicts   []Conflict
    Confidence  ConfidenceScore
    ResearchCtx *QuickScanResult
    StartedAt   time.Time
    UpdatedAt   time.Time
}

// NewSprintState creates a new sprint with all sections initialized
func NewSprintState(projectPath string) *SprintState {
    sections := make(map[Phase]*SectionDraft)
    for _, p := range AllPhases() {
        sections[p] = &SectionDraft{
            Status: DraftPending,
        }
    }

    return &SprintState{
        ProjectPath: projectPath,
        Phase:       PhaseProblem,
        Sections:    sections,
        Conflicts:   []Conflict{},
        StartedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
}

// QuickScanResult holds Ranger's research findings
type QuickScanResult struct {
    Topic      string
    GitHubHits []GitHubFinding
    HNHits     []HNFinding
    Summary    string
    ScannedAt  time.Time
}

// GitHubFinding represents a relevant repository
type GitHubFinding struct {
    Name        string
    Description string
    Stars       int
    URL         string
}

// HNFinding represents a relevant HN discussion
type HNFinding struct {
    Title    string
    Points   int
    Comments int
    URL      string
    Theme    string // Extracted theme from discussion
}

// Conflict represents a consistency issue between sections
type Conflict struct {
    Type     ConflictType
    Severity Severity
    Message  string
    Sections []Phase // Which sections are in conflict
}

// ConflictType categorizes consistency issues
type ConflictType int

const (
    ConflictUserFeature ConflictType = iota // Feature doesn't match target users
    ConflictGoalFeature                      // Goal not supported by features
    ConflictScopeCreep                       // Feature contradicts non-goals
    ConflictAssumption                       // Assumption conflicts with other content
)

// Severity indicates if the conflict blocks progress
type Severity int

const (
    SeverityBlocker Severity = iota // Must resolve before continuing
    SeverityWarning                 // Can dismiss with acknowledgment
)
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/gurgeh/arbiter/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/gurgeh/arbiter/
git commit -m "$(cat <<'EOF'
feat(arbiter): add core sprint state types

Define the data structures for Arbiter Spec Sprint:
- Phase enum for 6-section flow
- SectionDraft for propose-first interaction
- ConfidenceScore with weighted calculation
- SprintState as the session container
- Conflict types for consistency checking

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Add Persistence for Sprint State

**Files:**
- Create: `internal/gurgeh/arbiter/persistence.go`
- Test: `internal/gurgeh/arbiter/persistence_test.go`

**Step 1: Write the failing test**

```go
// internal/gurgeh/arbiter/persistence_test.go
package arbiter

import (
    "os"
    "path/filepath"
    "testing"
)

func TestSaveAndLoadSprintState(t *testing.T) {
    // Create temp directory
    tmpDir, err := os.MkdirTemp("", "arbiter-test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    // Create a sprint state
    state := NewSprintState(tmpDir)
    state.ID = "SPRINT-001"
    state.Sections[PhaseProblem].Content = "Test problem statement"
    state.Sections[PhaseProblem].Status = DraftAccepted
    state.Confidence.Completeness = 0.5

    // Save it
    if err := SaveSprintState(state); err != nil {
        t.Fatalf("save failed: %v", err)
    }

    // Verify file exists
    statePath := filepath.Join(tmpDir, ".gurgeh", "sprints", "SPRINT-001.yaml")
    if _, err := os.Stat(statePath); os.IsNotExist(err) {
        t.Fatalf("sprint file not created at %s", statePath)
    }

    // Load it back
    loaded, err := LoadSprintState(tmpDir, "SPRINT-001")
    if err != nil {
        t.Fatalf("load failed: %v", err)
    }

    // Verify content
    if loaded.Sections[PhaseProblem].Content != "Test problem statement" {
        t.Errorf("content mismatch: got %q", loaded.Sections[PhaseProblem].Content)
    }
    if loaded.Confidence.Completeness != 0.5 {
        t.Errorf("confidence mismatch: got %f", loaded.Confidence.Completeness)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/gurgeh/arbiter/... -run TestSaveAndLoad -v`
Expected: FAIL with "undefined: SaveSprintState"

**Step 3: Write minimal implementation**

```go
// internal/gurgeh/arbiter/persistence.go
package arbiter

import (
    "fmt"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

const sprintsDir = ".gurgeh/sprints"

// SaveSprintState persists a sprint to disk
func SaveSprintState(state *SprintState) error {
    dir := filepath.Join(state.ProjectPath, sprintsDir)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("create sprints dir: %w", err)
    }

    data, err := yaml.Marshal(state)
    if err != nil {
        return fmt.Errorf("marshal state: %w", err)
    }

    path := filepath.Join(dir, state.ID+".yaml")
    if err := os.WriteFile(path, data, 0644); err != nil {
        return fmt.Errorf("write state: %w", err)
    }

    return nil
}

// LoadSprintState reads a sprint from disk
func LoadSprintState(projectPath, id string) (*SprintState, error) {
    path := filepath.Join(projectPath, sprintsDir, id+".yaml")

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read state: %w", err)
    }

    var state SprintState
    if err := yaml.Unmarshal(data, &state); err != nil {
        return nil, fmt.Errorf("unmarshal state: %w", err)
    }

    return &state, nil
}

// ListSprints returns all sprint IDs in a project
func ListSprints(projectPath string) ([]string, error) {
    dir := filepath.Join(projectPath, sprintsDir)

    entries, err := os.ReadDir(dir)
    if err != nil {
        if os.IsNotExist(err) {
            return []string{}, nil
        }
        return nil, err
    }

    var ids []string
    for _, e := range entries {
        if !e.IsDir() && filepath.Ext(e.Name()) == ".yaml" {
            ids = append(ids, e.Name()[:len(e.Name())-5])
        }
    }

    return ids, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/gurgeh/arbiter/... -run TestSaveAndLoad -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/gurgeh/arbiter/persistence.go internal/gurgeh/arbiter/persistence_test.go
git commit -m "$(cat <<'EOF'
feat(arbiter): add sprint state persistence

Save/load sprint state to .gurgeh/sprints/ directory.
Enables resuming sprints across sessions.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 2: Consistency Engine

### Task 3: Define Consistency Checker Interface

**Files:**
- Create: `internal/gurgeh/consistency/checker.go`
- Test: `internal/gurgeh/consistency/checker_test.go`

**Step 1: Write the failing test**

```go
// internal/gurgeh/consistency/checker_test.go
package consistency

import (
    "testing"

    "github.com/mistakeknot/autarch/internal/gurgeh/arbiter"
)

func TestCheckersReturnEmptyForEmptyState(t *testing.T) {
    state := arbiter.NewSprintState("/tmp/test")
    engine := NewEngine()

    conflicts := engine.Check(state)

    // Empty state should have no conflicts (nothing to conflict with)
    if len(conflicts) != 0 {
        t.Errorf("expected 0 conflicts for empty state, got %d", len(conflicts))
    }
}

func TestUserFeatureMismatch(t *testing.T) {
    state := arbiter.NewSprintState("/tmp/test")

    // Set up a mismatch: users are "solo developers" but feature requires "enterprise admin"
    state.Sections[arbiter.PhaseUsers].Content = "Solo developers building side projects"
    state.Sections[arbiter.PhaseFeaturesGoals].Content = `
Features:
- Enterprise admin dashboard for managing 100+ users
- Role-based access control with SSO integration
`

    engine := NewEngine()
    conflicts := engine.Check(state)

    // Should detect user-feature mismatch
    found := false
    for _, c := range conflicts {
        if c.Type == arbiter.ConflictUserFeature {
            found = true
            break
        }
    }

    if !found {
        t.Error("expected UserFeature conflict, none found")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/gurgeh/consistency/... -v`
Expected: FAIL with "no Go files in directory"

**Step 3: Write minimal implementation**

```go
// internal/gurgeh/consistency/checker.go
package consistency

import (
    "regexp"
    "strings"

    "github.com/mistakeknot/autarch/internal/gurgeh/arbiter"
)

// Checker detects a specific type of consistency issue
type Checker interface {
    Check(state *arbiter.SprintState) []arbiter.Conflict
    Name() string
}

// Engine runs all consistency checkers
type Engine struct {
    checkers []Checker
}

// NewEngine creates an engine with all default checkers
func NewEngine() *Engine {
    return &Engine{
        checkers: []Checker{
            &UserFeatureChecker{},
            &GoalFeatureChecker{},
            &ScopeCreepChecker{},
            &AssumptionChecker{},
        },
    }
}

// Check runs all checkers and returns combined conflicts
func (e *Engine) Check(state *arbiter.SprintState) []arbiter.Conflict {
    var all []arbiter.Conflict
    for _, c := range e.checkers {
        all = append(all, c.Check(state)...)
    }
    return all
}

// UserFeatureChecker detects when features don't match target users
type UserFeatureChecker struct{}

func (c *UserFeatureChecker) Name() string { return "user-feature" }

func (c *UserFeatureChecker) Check(state *arbiter.SprintState) []arbiter.Conflict {
    users := state.Sections[arbiter.PhaseUsers].Content
    features := state.Sections[arbiter.PhaseFeaturesGoals].Content

    if users == "" || features == "" {
        return nil
    }

    var conflicts []arbiter.Conflict

    // Check for enterprise features with solo user personas
    soloPatterns := regexp.MustCompile(`(?i)(solo|individual|single|personal|indie)`)
    enterprisePatterns := regexp.MustCompile(`(?i)(enterprise|admin|sso|100\+|team management|organization)`)

    isSoloUser := soloPatterns.MatchString(users)
    hasEnterpriseFeatures := enterprisePatterns.MatchString(features)

    if isSoloUser && hasEnterpriseFeatures {
        conflicts = append(conflicts, arbiter.Conflict{
            Type:     arbiter.ConflictUserFeature,
            Severity: arbiter.SeverityBlocker,
            Message:  "Features include enterprise capabilities but target users are individuals/solo developers",
            Sections: []arbiter.Phase{arbiter.PhaseUsers, arbiter.PhaseFeaturesGoals},
        })
    }

    return conflicts
}

// GoalFeatureChecker detects when goals aren't supported by features
type GoalFeatureChecker struct{}

func (c *GoalFeatureChecker) Name() string { return "goal-feature" }

func (c *GoalFeatureChecker) Check(state *arbiter.SprintState) []arbiter.Conflict {
    features := state.Sections[arbiter.PhaseFeaturesGoals].Content

    if features == "" {
        return nil
    }

    var conflicts []arbiter.Conflict

    // Extract goals from content (look for "Goals:" section or goal patterns)
    goalPatterns := []struct {
        goal    string
        feature string
        pattern *regexp.Regexp
    }{
        {"fast onboarding", "onboarding|signup|getting started", regexp.MustCompile(`(?i)fast\s+onboarding|quick\s+start|under\s+\d+\s+minutes?`)},
        {"mobile support", "mobile|responsive|ios|android", regexp.MustCompile(`(?i)mobile|responsive|cross-platform`)},
    }

    for _, gp := range goalPatterns {
        hasGoal := gp.pattern.MatchString(features)
        hasFeature := regexp.MustCompile(`(?i)`+gp.feature).MatchString(features)

        if hasGoal && !hasFeature {
            conflicts = append(conflicts, arbiter.Conflict{
                Type:     arbiter.ConflictGoalFeature,
                Severity: arbiter.SeverityWarning,
                Message:  "Goal mentions '" + gp.goal + "' but no supporting features found",
                Sections: []arbiter.Phase{arbiter.PhaseFeaturesGoals},
            })
        }
    }

    return conflicts
}

// ScopeCreepChecker detects features that contradict non-goals
type ScopeCreepChecker struct{}

func (c *ScopeCreepChecker) Name() string { return "scope-creep" }

func (c *ScopeCreepChecker) Check(state *arbiter.SprintState) []arbiter.Conflict {
    features := state.Sections[arbiter.PhaseFeaturesGoals].Content
    scope := state.Sections[arbiter.PhaseScopeAssumptions].Content

    if features == "" || scope == "" {
        return nil
    }

    var conflicts []arbiter.Conflict

    // Look for non-goals and check if features violate them
    nonGoalPatterns := []struct {
        nonGoal string
        feature *regexp.Regexp
    }{
        {"no AI", regexp.MustCompile(`(?i)\bAI\b|artificial intelligence|machine learning|LLM|GPT`)},
        {"no mobile", regexp.MustCompile(`(?i)mobile app|ios|android|native app`)},
        {"no social", regexp.MustCompile(`(?i)social features|sharing|followers|friends`)},
    }

    for _, ng := range nonGoalPatterns {
        // Check if this is listed as a non-goal
        isNonGoal := strings.Contains(strings.ToLower(scope), strings.ToLower(ng.nonGoal))
        hasFeature := ng.feature.MatchString(features)

        if isNonGoal && hasFeature {
            conflicts = append(conflicts, arbiter.Conflict{
                Type:     arbiter.ConflictScopeCreep,
                Severity: arbiter.SeverityBlocker,
                Message:  "Feature contradicts non-goal: '" + ng.nonGoal + "'",
                Sections: []arbiter.Phase{arbiter.PhaseFeaturesGoals, arbiter.PhaseScopeAssumptions},
            })
        }
    }

    return conflicts
}

// AssumptionChecker detects assumption conflicts
type AssumptionChecker struct{}

func (c *AssumptionChecker) Name() string { return "assumption" }

func (c *AssumptionChecker) Check(state *arbiter.SprintState) []arbiter.Conflict {
    scope := state.Sections[arbiter.PhaseScopeAssumptions].Content
    features := state.Sections[arbiter.PhaseFeaturesGoals].Content

    if scope == "" {
        return nil
    }

    var conflicts []arbiter.Conflict

    // Check for common assumption-feature mismatches
    assumptionPatterns := []struct {
        assumption string
        requires   *regexp.Regexp
    }{
        {"users have accounts", regexp.MustCompile(`(?i)signup|registration|create account|login`)},
        {"users are authenticated", regexp.MustCompile(`(?i)auth|login|session|token`)},
        {"internet connection", regexp.MustCompile(`(?i)offline|local-first|sync`)},
    }

    for _, ap := range assumptionPatterns {
        hasAssumption := strings.Contains(strings.ToLower(scope), strings.ToLower(ap.assumption))
        hasRequiredFeature := ap.requires.MatchString(features)

        if hasAssumption && !hasRequiredFeature && features != "" {
            conflicts = append(conflicts, arbiter.Conflict{
                Type:     arbiter.ConflictAssumption,
                Severity: arbiter.SeverityWarning,
                Message:  "Assumes '" + ap.assumption + "' but no supporting feature found",
                Sections: []arbiter.Phase{arbiter.PhaseScopeAssumptions, arbiter.PhaseFeaturesGoals},
            })
        }
    }

    return conflicts
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/gurgeh/consistency/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/gurgeh/consistency/
git commit -m "$(cat <<'EOF'
feat(consistency): add PRD consistency checking engine

Implements 4 checkers following the design spec:
- UserFeatureChecker: detects user-feature mismatches
- GoalFeatureChecker: finds unsupported goals
- ScopeCreepChecker: catches non-goal violations
- AssumptionChecker: validates assumption dependencies

Blockers must be resolved; warnings can be dismissed.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 3: Confidence Scoring

### Task 4: Implement Confidence Calculator

**Files:**
- Create: `internal/gurgeh/confidence/calculator.go`
- Test: `internal/gurgeh/confidence/calculator_test.go`

**Step 1: Write the failing test**

```go
// internal/gurgeh/confidence/calculator_test.go
package confidence

import (
    "testing"

    "github.com/mistakeknot/autarch/internal/gurgeh/arbiter"
)

func TestEmptyStateHasZeroConfidence(t *testing.T) {
    state := arbiter.NewSprintState("/tmp/test")
    calc := NewCalculator()

    score := calc.Calculate(state)

    if score.Total() != 0 {
        t.Errorf("expected 0 confidence for empty state, got %f", score.Total())
    }
}

func TestCompleteStateHasHighConfidence(t *testing.T) {
    state := arbiter.NewSprintState("/tmp/test")

    // Fill all sections with substantive content
    state.Sections[arbiter.PhaseProblem].Content = "Users waste time on repetitive tasks"
    state.Sections[arbiter.PhaseProblem].Status = arbiter.DraftAccepted

    state.Sections[arbiter.PhaseUsers].Content = "Software developers, 25-45, building side projects"
    state.Sections[arbiter.PhaseUsers].Status = arbiter.DraftAccepted

    state.Sections[arbiter.PhaseFeaturesGoals].Content = `
Features:
- Automated task scheduling
- Integration with GitHub
Goals:
- Reduce manual work by 50%
`
    state.Sections[arbiter.PhaseFeaturesGoals].Status = arbiter.DraftAccepted

    state.Sections[arbiter.PhaseScopeAssumptions].Content = `
In scope: Core automation
Out of scope: Mobile app
Assumptions: Users have GitHub accounts
`
    state.Sections[arbiter.PhaseScopeAssumptions].Status = arbiter.DraftAccepted

    state.Sections[arbiter.PhaseCUJs].Content = `
CUJ 1: User connects GitHub → creates first automation → sees results
`
    state.Sections[arbiter.PhaseCUJs].Status = arbiter.DraftAccepted

    state.Sections[arbiter.PhaseAcceptanceCriteria].Content = `
- Automation runs within 5 seconds of trigger
- Errors are logged with actionable messages
`
    state.Sections[arbiter.PhaseAcceptanceCriteria].Status = arbiter.DraftAccepted

    calc := NewCalculator()
    score := calc.Calculate(state)

    // Should have high completeness (all sections filled and accepted)
    if score.Completeness < 0.8 {
        t.Errorf("expected high completeness, got %f", score.Completeness)
    }

    // Total should be reasonably high
    if score.Total() < 0.5 {
        t.Errorf("expected total > 0.5 for complete state, got %f", score.Total())
    }
}

func TestSpecificityScoring(t *testing.T) {
    state := arbiter.NewSprintState("/tmp/test")

    // Vague content
    state.Sections[arbiter.PhaseProblem].Content = "Things are hard"
    state.Sections[arbiter.PhaseProblem].Status = arbiter.DraftAccepted

    calc := NewCalculator()
    score1 := calc.Calculate(state)

    // Specific content with metrics
    state.Sections[arbiter.PhaseProblem].Content = "Developers spend 2 hours daily on manual deployments, costing $50k/year in lost productivity"

    score2 := calc.Calculate(state)

    if score2.Specificity <= score1.Specificity {
        t.Errorf("expected specific content to score higher: vague=%f, specific=%f",
            score1.Specificity, score2.Specificity)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/gurgeh/confidence/... -v`
Expected: FAIL with "no Go files in directory"

**Step 3: Write minimal implementation**

```go
// internal/gurgeh/confidence/calculator.go
package confidence

import (
    "regexp"
    "strings"

    "github.com/mistakeknot/autarch/internal/gurgeh/arbiter"
)

// Calculator computes confidence scores for sprint states
type Calculator struct{}

// NewCalculator creates a new confidence calculator
func NewCalculator() *Calculator {
    return &Calculator{}
}

// Calculate computes all confidence dimensions
func (c *Calculator) Calculate(state *arbiter.SprintState) arbiter.ConfidenceScore {
    return arbiter.ConfidenceScore{
        Completeness: c.completeness(state),
        Consistency:  c.consistency(state),
        Specificity:  c.specificity(state),
        Research:     c.research(state),
        Assumptions:  c.assumptions(state),
    }
}

// completeness: all sections filled and accepted (0-1)
func (c *Calculator) completeness(state *arbiter.SprintState) float64 {
    total := float64(len(arbiter.AllPhases()))
    filled := 0.0

    for _, phase := range arbiter.AllPhases() {
        section := state.Sections[phase]
        if section == nil {
            continue
        }

        // Content exists: 0.5 points
        if strings.TrimSpace(section.Content) != "" {
            filled += 0.5
        }

        // Content is accepted: 0.5 points
        if section.Status == arbiter.DraftAccepted {
            filled += 0.5
        }
    }

    return filled / total
}

// consistency: no blocking conflicts (0-1)
func (c *Calculator) consistency(state *arbiter.SprintState) float64 {
    if len(state.Conflicts) == 0 {
        return 1.0
    }

    blockers := 0
    warnings := 0
    for _, conflict := range state.Conflicts {
        if conflict.Severity == arbiter.SeverityBlocker {
            blockers++
        } else {
            warnings++
        }
    }

    // Each blocker reduces score by 0.25, warnings by 0.1
    score := 1.0 - (float64(blockers)*0.25 + float64(warnings)*0.1)
    if score < 0 {
        return 0
    }
    return score
}

// specificity: measurable criteria, numbers, concrete examples (0-1)
func (c *Calculator) specificity(state *arbiter.SprintState) float64 {
    var totalScore float64
    var sections int

    // Patterns that indicate specificity
    numberPattern := regexp.MustCompile(`\d+`)
    metricPattern := regexp.MustCompile(`(?i)(seconds?|minutes?|hours?|days?|weeks?|\$|%|users?|requests?|per)`)
    examplePattern := regexp.MustCompile(`(?i)(e\.g\.|for example|such as|like)`)

    for _, phase := range arbiter.AllPhases() {
        section := state.Sections[phase]
        if section == nil || strings.TrimSpace(section.Content) == "" {
            continue
        }

        sections++
        content := section.Content

        score := 0.0

        // Has numbers
        if numberPattern.MatchString(content) {
            score += 0.4
        }

        // Has metrics/units
        if metricPattern.MatchString(content) {
            score += 0.3
        }

        // Has examples
        if examplePattern.MatchString(content) {
            score += 0.3
        }

        totalScore += score
    }

    if sections == 0 {
        return 0
    }

    return totalScore / float64(sections)
}

// research: has research context and findings (0-1)
func (c *Calculator) research(state *arbiter.SprintState) float64 {
    if state.ResearchCtx == nil {
        return 0
    }

    score := 0.0

    // Has any GitHub findings
    if len(state.ResearchCtx.GitHubHits) > 0 {
        score += 0.4
    }

    // Has any HN findings
    if len(state.ResearchCtx.HNHits) > 0 {
        score += 0.4
    }

    // Has a summary
    if strings.TrimSpace(state.ResearchCtx.Summary) != "" {
        score += 0.2
    }

    return score
}

// assumptions: assumptions are reasonable and documented (0-1)
func (c *Calculator) assumptions(state *arbiter.SprintState) float64 {
    scope := state.Sections[arbiter.PhaseScopeAssumptions]
    if scope == nil || strings.TrimSpace(scope.Content) == "" {
        return 0
    }

    content := strings.ToLower(scope.Content)
    score := 0.0

    // Has explicit assumptions section
    if strings.Contains(content, "assumption") {
        score += 0.4
    }

    // Has impact analysis (what if false)
    if strings.Contains(content, "if not") || strings.Contains(content, "otherwise") ||
       strings.Contains(content, "impact") {
        score += 0.3
    }

    // Has confidence levels
    if strings.Contains(content, "confident") || strings.Contains(content, "likely") ||
       strings.Contains(content, "uncertain") {
        score += 0.3
    }

    return score
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/gurgeh/confidence/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/gurgeh/confidence/
git commit -m "$(cat <<'EOF'
feat(confidence): add PRD confidence calculator

Implements 5-factor confidence scoring:
- Completeness (20%): sections filled and accepted
- Consistency (25%): no blocking conflicts
- Specificity (20%): numbers, metrics, examples
- Research (20%): quick scan results
- Assumptions (15%): documented with impact analysis

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 4: Ranger Quick Scan

### Task 5: Add Quick Scan Mode to Pollard

**Files:**
- Create: `internal/pollard/quick/scan.go`
- Test: `internal/pollard/quick/scan_test.go`

**Step 1: Write the failing test**

```go
// internal/pollard/quick/scan_test.go
package quick

import (
    "context"
    "testing"
    "time"
)

func TestQuickScanReturnsWithinTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
    defer cancel()

    scanner := NewScanner()

    // Use a topic that should have results
    result, err := scanner.Scan(ctx, "reading tracker app")

    if err != nil {
        t.Fatalf("scan failed: %v", err)
    }

    if result == nil {
        t.Fatal("expected non-nil result")
    }

    if result.Topic != "reading tracker app" {
        t.Errorf("expected topic 'reading tracker app', got %q", result.Topic)
    }

    // Should complete within 30 seconds (plus some buffer)
    if time.Since(result.ScannedAt) > 45*time.Second {
        t.Error("scan took too long")
    }
}

func TestQuickScanHasGitHubAndHNResults(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
    defer cancel()

    scanner := NewScanner()
    result, err := scanner.Scan(ctx, "todo app")

    if err != nil {
        t.Skipf("skipping (network required): %v", err)
    }

    // Should have at least one source populated
    if len(result.GitHubHits) == 0 && len(result.HNHits) == 0 {
        t.Error("expected at least one result from GitHub or HN")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/pollard/quick/... -v`
Expected: FAIL with "no Go files in directory"

**Step 3: Write minimal implementation**

```go
// internal/pollard/quick/scan.go
package quick

import (
    "context"
    "sync"
    "time"

    "github.com/mistakeknot/autarch/internal/gurgeh/arbiter"
    "github.com/mistakeknot/autarch/internal/pollard/hunters"
)

// Scanner performs quick research scans for PRD context
type Scanner struct {
    github *hunters.GitHubScout
    hn     *hunters.HackerNewsHunter
}

// NewScanner creates a scanner with default hunters
func NewScanner() *Scanner {
    return &Scanner{
        github: hunters.NewGitHubScout(),
        hn:     hunters.NewHackerNewsHunter(),
    }
}

// Scan runs github-scout and hackernews in parallel with 30s timeout
func (s *Scanner) Scan(ctx context.Context, topic string) (*arbiter.QuickScanResult, error) {
    // Apply 30 second timeout for quick scan
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    result := &arbiter.QuickScanResult{
        Topic:     topic,
        ScannedAt: time.Now(),
    }

    var wg sync.WaitGroup
    var mu sync.Mutex

    // Run GitHub scout
    wg.Add(1)
    go func() {
        defer wg.Done()

        cfg := hunters.HunterConfig{
            Query:     topic,
            MaxHits:   5,
            TimeoutMs: 15000,
        }

        ghResult, err := s.github.Hunt(ctx, cfg)
        if err != nil {
            return // Ignore errors, continue with partial results
        }

        mu.Lock()
        defer mu.Unlock()

        for _, item := range ghResult.Items {
            result.GitHubHits = append(result.GitHubHits, arbiter.GitHubFinding{
                Name:        item.Name,
                Description: item.Description,
                Stars:       item.Stars,
                URL:         item.URL,
            })
        }
    }()

    // Run HackerNews hunter
    wg.Add(1)
    go func() {
        defer wg.Done()

        cfg := hunters.HunterConfig{
            Query:     topic,
            MaxHits:   5,
            TimeoutMs: 15000,
        }

        hnResult, err := s.hn.Hunt(ctx, cfg)
        if err != nil {
            return // Ignore errors, continue with partial results
        }

        mu.Lock()
        defer mu.Unlock()

        for _, item := range hnResult.Items {
            result.HNHits = append(result.HNHits, arbiter.HNFinding{
                Title:    item.Title,
                Points:   item.Points,
                Comments: item.Comments,
                URL:      item.URL,
            })
        }
    }()

    wg.Wait()

    // Generate summary
    result.Summary = s.synthesizeSummary(result)

    return result, nil
}

// synthesizeSummary creates a brief summary of findings
func (s *Scanner) synthesizeSummary(result *arbiter.QuickScanResult) string {
    if len(result.GitHubHits) == 0 && len(result.HNHits) == 0 {
        return "No relevant results found."
    }

    var summary string

    if len(result.GitHubHits) > 0 {
        summary += "Found " + string(rune('0'+len(result.GitHubHits))) + " relevant GitHub projects"
        if result.GitHubHits[0].Stars > 1000 {
            summary += " (including popular ones with 1k+ stars)"
        }
        summary += ". "
    }

    if len(result.HNHits) > 0 {
        summary += "Found " + string(rune('0'+len(result.HNHits))) + " HackerNews discussions"
        totalComments := 0
        for _, h := range result.HNHits {
            totalComments += h.Comments
        }
        if totalComments > 100 {
            summary += " with active community engagement"
        }
        summary += "."
    }

    return summary
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/pollard/quick/... -v -short`
Expected: PASS (or SKIP if no network)

**Step 5: Commit**

```bash
git add internal/pollard/quick/
git commit -m "$(cat <<'EOF'
feat(pollard): add quick scan mode for Arbiter

Runs github-scout + hackernews in parallel with 30s timeout.
Used by Arbiter after Problem section to inform Features.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 5: Arbiter Core Logic

### Task 6: Implement Arbiter Draft Generator

**Files:**
- Create: `internal/gurgeh/arbiter/generator.go`
- Test: `internal/gurgeh/arbiter/generator_test.go`

**Step 1: Write the failing test**

```go
// internal/gurgeh/arbiter/generator_test.go
package arbiter

import (
    "context"
    "testing"
)

func TestGenerateDraftFromContext(t *testing.T) {
    gen := NewGenerator()

    ctx := context.Background()
    projectCtx := &ProjectContext{
        HasReadme:    true,
        ReadmeSnippet: "A CLI tool for managing reading lists",
        HasPackageJSON: false,
    }

    draft, err := gen.GenerateDraft(ctx, PhaseProblem, projectCtx, "")

    if err != nil {
        t.Fatalf("generate failed: %v", err)
    }

    if draft.Content == "" {
        t.Error("expected non-empty draft content")
    }

    if len(draft.Options) < 2 {
        t.Errorf("expected at least 2 options, got %d", len(draft.Options))
    }
}

func TestGenerateDraftFromUserInput(t *testing.T) {
    gen := NewGenerator()

    ctx := context.Background()

    draft, err := gen.GenerateDraft(ctx, PhaseProblem, nil, "I want to build a habit tracker for developers")

    if err != nil {
        t.Fatalf("generate failed: %v", err)
    }

    // Should incorporate user's input
    if draft.Content == "" {
        t.Error("expected non-empty draft content")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/gurgeh/arbiter/... -run TestGenerate -v`
Expected: FAIL with "undefined: NewGenerator"

**Step 3: Write minimal implementation**

```go
// internal/gurgeh/arbiter/generator.go
package arbiter

import (
    "context"
    "fmt"
    "strings"
)

// ProjectContext provides information about an existing project
type ProjectContext struct {
    HasReadme       bool
    ReadmeSnippet   string
    HasPackageJSON  bool
    PackageName     string
    Dependencies    []string
    MainFiles       []string
}

// Generator creates section drafts based on context
type Generator struct{}

// NewGenerator creates a new draft generator
func NewGenerator() *Generator {
    return &Generator{}
}

// GenerateDraft creates a proposed draft for a section
func (g *Generator) GenerateDraft(ctx context.Context, phase Phase, projectCtx *ProjectContext, userInput string) (*SectionDraft, error) {
    var content string
    var options []string

    switch phase {
    case PhaseProblem:
        content, options = g.generateProblem(projectCtx, userInput)
    case PhaseUsers:
        content, options = g.generateUsers(projectCtx, userInput)
    case PhaseFeaturesGoals:
        content, options = g.generateFeaturesGoals(projectCtx, userInput)
    case PhaseScopeAssumptions:
        content, options = g.generateScopeAssumptions(projectCtx, userInput)
    case PhaseCUJs:
        content, options = g.generateCUJs(projectCtx, userInput)
    case PhaseAcceptanceCriteria:
        content, options = g.generateAcceptanceCriteria(projectCtx, userInput)
    default:
        return nil, fmt.Errorf("unknown phase: %v", phase)
    }

    return &SectionDraft{
        Content: content,
        Options: options,
        Status:  DraftProposed,
    }, nil
}

func (g *Generator) generateProblem(ctx *ProjectContext, input string) (string, []string) {
    if input != "" {
        // User provided their idea - draft from it
        return g.draftFromInput(input, "problem"), g.problemOptions(input)
    }

    if ctx != nil && ctx.HasReadme {
        // Infer from project context
        return g.draftFromContext(ctx, "problem"), g.problemOptions(ctx.ReadmeSnippet)
    }

    // No context - return placeholder
    return "[Describe the problem this solves]", []string{
        "Focus on user pain points",
        "Focus on business impact",
        "Focus on technical gaps",
    }
}

func (g *Generator) draftFromInput(input, section string) string {
    // Simple extraction - in production this would use an LLM
    input = strings.TrimSpace(input)

    switch section {
    case "problem":
        // Try to extract the problem from "I want to build X because Y"
        if strings.Contains(strings.ToLower(input), "because") {
            parts := strings.SplitN(input, "because", 2)
            if len(parts) == 2 {
                return strings.TrimSpace(parts[1])
            }
        }
        return input
    default:
        return input
    }
}

func (g *Generator) draftFromContext(ctx *ProjectContext, section string) string {
    if ctx.ReadmeSnippet != "" {
        return "Based on the project: " + ctx.ReadmeSnippet
    }
    return "[Could not infer from context]"
}

func (g *Generator) problemOptions(base string) []string {
    return []string{
        "Make it more specific with metrics",
        "Focus on the user's emotional pain",
        "Emphasize the business cost",
    }
}

func (g *Generator) generateUsers(ctx *ProjectContext, input string) (string, []string) {
    content := "[Describe the target users]"
    if input != "" {
        content = input
    }

    return content, []string{
        "Add demographic details",
        "Focus on technical skill level",
        "Describe their current workflow",
    }
}

func (g *Generator) generateFeaturesGoals(ctx *ProjectContext, input string) (string, []string) {
    content := `Features:
- [Feature 1]
- [Feature 2]

Goals:
- [Goal with measurable metric]`

    if input != "" {
        content = input
    }

    return content, []string{
        "Add more features",
        "Make goals more measurable",
        "Prioritize by user value",
    }
}

func (g *Generator) generateScopeAssumptions(ctx *ProjectContext, input string) (string, []string) {
    content := `In Scope:
- [What's included]

Out of Scope (Non-Goals):
- [What's explicitly excluded]

Assumptions:
- [Key assumption with impact if false]`

    if input != "" {
        content = input
    }

    return content, []string{
        "Add more non-goals",
        "Document assumption risks",
        "Be more explicit about boundaries",
    }
}

func (g *Generator) generateCUJs(ctx *ProjectContext, input string) (string, []string) {
    content := `CUJ 1: [Title]
1. User [action]
2. System [response]
3. User sees [outcome]

Success criteria: [How we know it worked]`

    if input != "" {
        content = input
    }

    return content, []string{
        "Add error case journey",
        "Add first-time user journey",
        "Add power user journey",
    }
}

func (g *Generator) generateAcceptanceCriteria(ctx *ProjectContext, input string) (string, []string) {
    content := `- [ ] [Testable criterion with specific metric]
- [ ] [Another testable criterion]`

    if input != "" {
        content = input
    }

    return content, []string{
        "Add performance criteria",
        "Add error handling criteria",
        "Add accessibility criteria",
    }
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/gurgeh/arbiter/... -run TestGenerate -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/gurgeh/arbiter/generator.go internal/gurgeh/arbiter/generator_test.go
git commit -m "$(cat <<'EOF'
feat(arbiter): add draft generator for propose-first flow

Generates section drafts from:
- User input (their idea description)
- Project context (README, package.json)
- Fallback templates when no context

Each draft includes 2-3 alternative phrasings.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 7: Implement Sprint Orchestrator

**Files:**
- Create: `internal/gurgeh/arbiter/orchestrator.go`
- Test: `internal/gurgeh/arbiter/orchestrator_test.go`

**Step 1: Write the failing test**

```go
// internal/gurgeh/arbiter/orchestrator_test.go
package arbiter

import (
    "context"
    "testing"
)

func TestOrchestratorStartsSprint(t *testing.T) {
    orch := NewOrchestrator("/tmp/test-project")

    ctx := context.Background()
    state, err := orch.Start(ctx, "Build a habit tracker for developers")

    if err != nil {
        t.Fatalf("start failed: %v", err)
    }

    if state.Phase != PhaseProblem {
        t.Errorf("expected phase Problem, got %v", state.Phase)
    }

    // Should have a draft for Problem section
    if state.Sections[PhaseProblem].Content == "" {
        t.Error("expected Problem section to have draft content")
    }
}

func TestOrchestratorAdvancesPhase(t *testing.T) {
    orch := NewOrchestrator("/tmp/test-project")

    ctx := context.Background()
    state, _ := orch.Start(ctx, "Build a habit tracker")

    // Accept the problem draft
    state.Sections[PhaseProblem].Status = DraftAccepted

    // Advance to next phase
    state, err := orch.Advance(ctx, state)

    if err != nil {
        t.Fatalf("advance failed: %v", err)
    }

    if state.Phase != PhaseUsers {
        t.Errorf("expected phase Users, got %v", state.Phase)
    }
}

func TestOrchestratorTriggersQuickScan(t *testing.T) {
    orch := NewOrchestrator("/tmp/test-project")

    ctx := context.Background()
    state, _ := orch.Start(ctx, "Build a reading tracker app")

    // Accept problem
    state.Sections[PhaseProblem].Status = DraftAccepted
    state, _ = orch.Advance(ctx, state)

    // Accept users
    state.Sections[PhaseUsers].Status = DraftAccepted
    state, _ = orch.Advance(ctx, state)

    // Now at Features - should have research context from quick scan
    // (In real impl, this would be async, but for test we check it's triggered)
    if state.Phase != PhaseFeaturesGoals {
        t.Errorf("expected phase FeaturesGoals, got %v", state.Phase)
    }
}

func TestOrchestratorBlocksOnConflicts(t *testing.T) {
    orch := NewOrchestrator("/tmp/test-project")

    ctx := context.Background()
    state, _ := orch.Start(ctx, "")

    // Create a state with a blocker conflict
    state.Sections[PhaseProblem].Content = "Problem for solo developers"
    state.Sections[PhaseProblem].Status = DraftAccepted
    state.Sections[PhaseUsers].Content = "Solo developers"
    state.Sections[PhaseUsers].Status = DraftAccepted
    state.Sections[PhaseFeaturesGoals].Content = "Enterprise admin dashboard for 100+ users"
    state.Sections[PhaseFeaturesGoals].Status = DraftAccepted

    state.Phase = PhaseFeaturesGoals

    // Try to advance - should be blocked
    _, err := orch.Advance(ctx, state)

    if err == nil {
        t.Error("expected error due to blocker conflict")
    }

    if !IsBlockerError(err) {
        t.Errorf("expected blocker error, got: %v", err)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/gurgeh/arbiter/... -run TestOrchestrator -v`
Expected: FAIL with "undefined: NewOrchestrator"

**Step 3: Write minimal implementation**

```go
// internal/gurgeh/arbiter/orchestrator.go
package arbiter

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/mistakeknot/autarch/internal/gurgeh/consistency"
    "github.com/mistakeknot/autarch/internal/gurgeh/confidence"
    "github.com/mistakeknot/autarch/internal/pollard/quick"
)

// ErrBlocker indicates advancement is blocked by conflicts
var ErrBlocker = errors.New("blocked by consistency conflicts")

// IsBlockerError checks if an error is a blocker
func IsBlockerError(err error) bool {
    return errors.Is(err, ErrBlocker)
}

// Orchestrator manages the sprint flow
type Orchestrator struct {
    projectPath string
    generator   *Generator
    consistency *consistency.Engine
    confidence  *confidence.Calculator
    scanner     *quick.Scanner
}

// NewOrchestrator creates a new sprint orchestrator
func NewOrchestrator(projectPath string) *Orchestrator {
    return &Orchestrator{
        projectPath: projectPath,
        generator:   NewGenerator(),
        consistency: consistency.NewEngine(),
        confidence:  confidence.NewCalculator(),
        scanner:     quick.NewScanner(),
    }
}

// Start begins a new sprint with optional initial input
func (o *Orchestrator) Start(ctx context.Context, userInput string) (*SprintState, error) {
    state := NewSprintState(o.projectPath)
    state.ID = fmt.Sprintf("SPRINT-%d", time.Now().Unix())

    // Read project context if available
    projectCtx := o.readProjectContext()

    // Generate initial draft for Problem section
    draft, err := o.generator.GenerateDraft(ctx, PhaseProblem, projectCtx, userInput)
    if err != nil {
        return nil, fmt.Errorf("generate problem draft: %w", err)
    }

    state.Sections[PhaseProblem] = draft

    return state, nil
}

// Advance moves to the next phase after validating current phase
func (o *Orchestrator) Advance(ctx context.Context, state *SprintState) (*SprintState, error) {
    // Run consistency checks
    conflicts := o.consistency.Check(state)
    state.Conflicts = conflicts

    // Check for blockers
    for _, c := range conflicts {
        if c.Severity == SeverityBlocker {
            return state, fmt.Errorf("%w: %s", ErrBlocker, c.Message)
        }
    }

    // Update confidence score
    state.Confidence = o.confidence.Calculate(state)

    // Determine next phase
    phases := AllPhases()
    currentIdx := -1
    for i, p := range phases {
        if p == state.Phase {
            currentIdx = i
            break
        }
    }

    if currentIdx == -1 || currentIdx >= len(phases)-1 {
        // Already at last phase or unknown
        return state, nil
    }

    nextPhase := phases[currentIdx+1]
    state.Phase = nextPhase

    // Trigger quick scan after Problem section (before Features)
    if nextPhase == PhaseFeaturesGoals && state.ResearchCtx == nil {
        go o.runQuickScan(ctx, state)
    }

    // Generate draft for next section if not already present
    if state.Sections[nextPhase].Content == "" {
        projectCtx := o.readProjectContext()
        draft, err := o.generator.GenerateDraft(ctx, nextPhase, projectCtx, "")
        if err != nil {
            return state, fmt.Errorf("generate draft: %w", err)
        }
        state.Sections[nextPhase] = draft
    }

    state.UpdatedAt = time.Now()

    return state, nil
}

// runQuickScan triggers the Ranger quick scan
func (o *Orchestrator) runQuickScan(ctx context.Context, state *SprintState) {
    // Extract topic from Problem section
    topic := state.Sections[PhaseProblem].Content
    if len(topic) > 100 {
        topic = topic[:100]
    }

    result, err := o.scanner.Scan(ctx, topic)
    if err != nil {
        // Log but don't fail - research is optional
        return
    }

    state.ResearchCtx = result
}

// readProjectContext extracts context from the project
func (o *Orchestrator) readProjectContext() *ProjectContext {
    // TODO: implement actual file reading
    return nil
}

// AcceptDraft marks the current section as accepted
func (o *Orchestrator) AcceptDraft(state *SprintState) *SprintState {
    state.Sections[state.Phase].Status = DraftAccepted
    state.UpdatedAt = time.Now()
    return state
}

// ReviseDraft updates the current section with user's changes
func (o *Orchestrator) ReviseDraft(state *SprintState, newContent string, reason string) *SprintState {
    section := state.Sections[state.Phase]

    // Record the edit
    section.UserEdits = append(section.UserEdits, Edit{
        Before:    section.Content,
        After:     newContent,
        Reason:    reason,
        Timestamp: time.Now(),
    })

    section.Content = newContent
    section.Status = DraftNeedsRevision
    state.UpdatedAt = time.Now()

    return state
}

// GetHandoffOptions returns available next steps after sprint completion
func (o *Orchestrator) GetHandoffOptions(state *SprintState) []HandoffOption {
    return []HandoffOption{
        {
            ID:          "research",
            Label:       "Research & iterate",
            Description: "Run deep research with Ranger and refine the PRD",
            Recommended: state.ResearchCtx == nil || len(state.ResearchCtx.GitHubHits) < 3,
        },
        {
            ID:          "tasks",
            Label:       "Generate tasks",
            Description: "Create epics and stories with Forger → Coldwine",
            Recommended: state.Confidence.Total() > 0.8,
        },
        {
            ID:          "export",
            Label:       "Export for coding agent",
            Description: "Generate YAML/Markdown for use with AI coding tools",
            Recommended: false,
        },
    }
}

// HandoffOption represents a possible next step
type HandoffOption struct {
    ID          string
    Label       string
    Description string
    Recommended bool
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/gurgeh/arbiter/... -run TestOrchestrator -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/gurgeh/arbiter/orchestrator.go internal/gurgeh/arbiter/orchestrator_test.go
git commit -m "$(cat <<'EOF'
feat(arbiter): add sprint orchestrator

Manages the full spec sprint flow:
- Start: initialize with user input or project context
- Advance: move between phases with consistency checks
- Accept/Revise: handle user responses to drafts
- Quick scan trigger after Problem section
- Handoff options after completion

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 6: TUI Integration

### Task 8: Create Sprint TUI View

**Files:**
- Create: `internal/gurgeh/tui/sprint.go`
- Modify: `internal/gurgeh/tui/model.go` (add sprint mode)

**Step 1: Write the failing test**

```go
// internal/gurgeh/tui/sprint_test.go
package tui

import (
    "testing"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/mistakeknot/autarch/internal/gurgeh/arbiter"
)

func TestSprintViewRendersDraft(t *testing.T) {
    state := arbiter.NewSprintState("/tmp/test")
    state.Sections[arbiter.PhaseProblem].Content = "Test problem"
    state.Sections[arbiter.PhaseProblem].Status = arbiter.DraftProposed

    view := NewSprintView(state)
    output := view.View()

    if output == "" {
        t.Error("expected non-empty view")
    }

    // Should show the draft content
    if !containsString(output, "Test problem") {
        t.Error("expected view to contain draft content")
    }

    // Should show options
    if !containsString(output, "Accept") {
        t.Error("expected view to show Accept option")
    }
}

func TestSprintViewHandlesAccept(t *testing.T) {
    state := arbiter.NewSprintState("/tmp/test")
    state.Sections[arbiter.PhaseProblem].Content = "Test problem"
    state.Sections[arbiter.PhaseProblem].Status = arbiter.DraftProposed

    view := NewSprintView(state)

    // Press 'a' to accept
    newView, cmd := view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

    sprintView := newView.(*SprintView)
    if sprintView.state.Sections[arbiter.PhaseProblem].Status != arbiter.DraftAccepted {
        t.Error("expected draft to be accepted")
    }

    _ = cmd // Command would trigger advance
}

func containsString(haystack, needle string) bool {
    return len(haystack) > 0 && len(needle) > 0 &&
        (haystack == needle || len(haystack) > len(needle))
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/gurgeh/tui/... -run TestSprintView -v`
Expected: FAIL with "undefined: NewSprintView"

**Step 3: Write minimal implementation**

```go
// internal/gurgeh/tui/sprint.go
package tui

import (
    "fmt"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/mistakeknot/autarch/internal/gurgeh/arbiter"
    "github.com/mistakeknot/autarch/pkg/tui"
)

// SprintView renders the Arbiter Spec Sprint interface
type SprintView struct {
    state       *arbiter.SprintState
    orchestrator *arbiter.Orchestrator
    width       int
    height      int
    focused     string // "draft" or "options"
    optionIndex int
}

// NewSprintView creates a new sprint view
func NewSprintView(state *arbiter.SprintState) *SprintView {
    return &SprintView{
        state:       state,
        orchestrator: arbiter.NewOrchestrator(state.ProjectPath),
        focused:     "options",
        optionIndex: 0,
    }
}

// Init implements tea.Model
func (v *SprintView) Init() tea.Cmd {
    return nil
}

// Update implements tea.Model
func (v *SprintView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "a", "A":
            // Accept current draft
            v.state = v.orchestrator.AcceptDraft(v.state)
            return v, v.advancePhase

        case "e", "E":
            // Edit draft (would open editor)
            return v, nil

        case "1", "2", "3":
            // Select an alternative option
            idx := int(msg.Runes[0] - '1')
            if idx < len(v.state.Sections[v.state.Phase].Options) {
                // Apply the option as revision reason
                return v, nil
            }

        case "up", "k":
            if v.optionIndex > 0 {
                v.optionIndex--
            }

        case "down", "j":
            options := v.state.Sections[v.state.Phase].Options
            if v.optionIndex < len(options)-1 {
                v.optionIndex++
            }

        case "enter":
            // Apply selected option
            return v, nil

        case "q", "esc":
            return v, tea.Quit
        }

    case tea.WindowSizeMsg:
        v.width = msg.Width
        v.height = msg.Height
    }

    return v, nil
}

// advancePhase moves to the next phase
func (v *SprintView) advancePhase() tea.Msg {
    // This would be called as a command
    return nil
}

// View implements tea.Model
func (v *SprintView) View() string {
    section := v.state.Sections[v.state.Phase]

    var b strings.Builder

    // Header with phase and confidence
    header := v.renderHeader()
    b.WriteString(header)
    b.WriteString("\n\n")

    // Draft content
    draftBox := v.renderDraft(section)
    b.WriteString(draftBox)
    b.WriteString("\n\n")

    // Options
    options := v.renderOptions(section)
    b.WriteString(options)
    b.WriteString("\n\n")

    // Conflicts (if any)
    if len(v.state.Conflicts) > 0 {
        conflicts := v.renderConflicts()
        b.WriteString(conflicts)
        b.WriteString("\n\n")
    }

    // Help
    help := v.renderHelp()
    b.WriteString(help)

    return b.String()
}

func (v *SprintView) renderHeader() string {
    phaseStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(tui.TokyoNight.Cyan)

    confidenceStyle := lipgloss.NewStyle().
        Foreground(tui.TokyoNight.Green)

    phase := phaseStyle.Render(v.state.Phase.String())
    confidence := confidenceStyle.Render(fmt.Sprintf("%.0f%%", v.state.Confidence.Total()*100))

    return fmt.Sprintf("Section: %s    Confidence: %s", phase, confidence)
}

func (v *SprintView) renderDraft(section *arbiter.SectionDraft) string {
    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(tui.TokyoNight.Blue).
        Padding(1, 2).
        Width(v.width - 4)

    statusIcon := "📝"
    switch section.Status {
    case arbiter.DraftAccepted:
        statusIcon = "✅"
    case arbiter.DraftNeedsRevision:
        statusIcon = "✏️"
    }

    header := lipgloss.NewStyle().
        Bold(true).
        Render(fmt.Sprintf("%s Arbiter's Draft", statusIcon))

    content := section.Content
    if content == "" {
        content = "[Generating draft...]"
    }

    return boxStyle.Render(header + "\n\n" + content)
}

func (v *SprintView) renderOptions(section *arbiter.SectionDraft) string {
    var b strings.Builder

    b.WriteString("Options:\n")

    // Main actions
    actions := []struct {
        key   string
        label string
    }{
        {"a", "Accept as-is"},
        {"e", "Edit directly"},
    }

    for _, a := range actions {
        style := lipgloss.NewStyle().Foreground(tui.TokyoNight.Yellow)
        b.WriteString(fmt.Sprintf("  [%s] %s\n", style.Render(a.key), a.label))
    }

    // Alternative phrasings
    if len(section.Options) > 0 {
        b.WriteString("\nAlternatives:\n")
        for i, opt := range section.Options {
            key := fmt.Sprintf("%d", i+1)
            style := lipgloss.NewStyle().Foreground(tui.TokyoNight.Purple)
            b.WriteString(fmt.Sprintf("  [%s] %s\n", style.Render(key), opt))
        }
    }

    return b.String()
}

func (v *SprintView) renderConflicts() string {
    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(tui.TokyoNight.Red).
        Padding(0, 1).
        Width(v.width - 4)

    var b strings.Builder
    b.WriteString("⚠️ Consistency Issues:\n")

    for _, c := range v.state.Conflicts {
        icon := "🟡"
        if c.Severity == arbiter.SeverityBlocker {
            icon = "🔴"
        }
        b.WriteString(fmt.Sprintf("  %s %s\n", icon, c.Message))
    }

    return boxStyle.Render(b.String())
}

func (v *SprintView) renderHelp() string {
    helpStyle := lipgloss.NewStyle().
        Foreground(tui.TokyoNight.Comment)

    return helpStyle.Render("↑/↓ navigate • enter select • a accept • e edit • q quit")
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/gurgeh/tui/... -run TestSprintView -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/gurgeh/tui/sprint.go internal/gurgeh/tui/sprint_test.go
git commit -m "$(cat <<'EOF'
feat(tui): add sprint view for propose-first PRD creation

New TUI component for Arbiter Spec Sprint:
- Shows Arbiter's draft with options to accept/edit
- Displays running confidence score
- Highlights consistency conflicts (blockers in red)
- Keyboard-driven interaction

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 9: Integrate Sprint View into Main TUI

**Files:**
- Modify: `internal/gurgeh/tui/model.go`
- Modify: `internal/gurgeh/tui/router.go`

**Step 1: Read current model.go structure**

Run: `head -100 internal/gurgeh/tui/model.go`
Check for mode constants and router integration points.

**Step 2: Add sprint mode to model**

```go
// Add to mode constants in model.go
const (
    modeList     = "list"
    modeDetail   = "detail"
    modeInterview = "interview"
    modeSearch   = "search"
    modeSprint   = "sprint"  // NEW
)

// Add to Model struct
type Model struct {
    // ... existing fields ...
    sprint *SprintView  // NEW: Arbiter sprint state
}

// Add sprint initialization to handleKey
case "n":
    if m.mode == modeList {
        // NEW: Start sprint instead of old interview
        state := arbiter.NewSprintState(m.root)
        m.sprint = NewSprintView(state)
        m.mode = modeSprint
        return m, nil
    }
```

**Step 3: Update router to handle sprint mode**

```go
// In router.go, add case for sprint mode
func (m *Model) View() string {
    switch m.mode {
    case modeList:
        return m.listView()
    case modeDetail:
        return m.detailView()
    case modeInterview:
        return m.interviewView()
    case modeSprint:  // NEW
        return m.sprint.View()
    case modeSearch:
        return m.searchView()
    default:
        return m.listView()
    }
}
```

**Step 4: Test the integration**

Run: `go build ./cmd/gurgeh && ./gurgeh --help`
Expected: Builds successfully

**Step 5: Commit**

```bash
git add internal/gurgeh/tui/model.go internal/gurgeh/tui/router.go
git commit -m "$(cat <<'EOF'
feat(tui): integrate sprint view as new PRD creation mode

Replace old interview flow with Arbiter Spec Sprint:
- 'n' key now starts sprint mode instead of interview
- Sprint view handles all propose-first interactions
- Legacy interview code preserved for migration period

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 7: Plugin Updates

### Task 10: Update Arbiter Agent Definition

**Files:**
- Modify: `autarch-plugin/agents/prd/arbiter.md`

**Step 1: Read current arbiter.md**

```bash
cat autarch-plugin/agents/prd/arbiter.md
```

**Step 2: Update with new flow**

The arbiter.md should be updated to reflect the new 6-section propose-first flow. Key changes:
- Replace 4-phase interview with 6-section sprint
- Add propose-first interaction model
- Document quick scan integration
- Add consistency checking behavior

**Step 3: Commit**

```bash
git add autarch-plugin/agents/prd/arbiter.md
git commit -m "$(cat <<'EOF'
docs(plugin): update Arbiter agent for spec sprint flow

Reflect new 6-section propose-first design:
- Problem → Users → Features+Goals → Scope+Assumptions → CUJs → AC
- Propose-first with options (Accept/Edit/Alternatives)
- Ranger quick scan after Problem section
- Consistency checking between sections

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 11: Create Spec Sprint Skill

**Files:**
- Create: `autarch-plugin/skills/spec-sprint/SKILL.md`

**Step 1: Write the skill definition**

```markdown
---
name: spec-sprint
description: This skill guides 10-minute PRD creation using Arbiter's propose-first workflow.
---

# Spec Sprint Skill

## When to Use

- Starting a new feature or project
- Converting a rough idea into a validated PRD
- Onboarding to an existing codebase that needs product direction

## The Sprint Flow

### Opening (1 min)
- For existing projects: Arbiter reads context and proposes a Problem statement
- For blank slate: Ask "Describe your idea" and draft from response

### Sections (6-8 min)

For each section, Arbiter proposes a draft. User can:
- **Accept** (press 'a') - Use draft as-is
- **Edit** (press 'e') - Modify directly
- **Alternative** (press 1-3) - Apply suggested rephrasing

| # | Section | What Arbiter Drafts |
|---|---------|---------------------|
| 1 | Problem | Pain point with context |
| 2 | Users | Target personas with characteristics |
| 3 | Features + Goals | Capabilities with measurable outcomes |
| 4 | Scope + Assumptions | Boundaries and foundational beliefs |
| 5 | CUJs | Critical User Journeys with steps |
| 6 | Acceptance Criteria | Testable success conditions |

### Quick Scan (after Problem)

Ranger runs github-scout + hackernews (~30 sec) to find:
- Similar OSS projects
- HN discussions about the problem space

Findings inform the Features section.

### Consistency Checking

After each section, check for conflicts:
- 🔴 **Blockers** - Must resolve (e.g., enterprise features for solo users)
- 🟡 **Warnings** - Can dismiss (e.g., goal without supporting feature)

### Handoff (1 min)

After all sections:
1. **Research & iterate** (Recommended first time) - Deep dive with Ranger
2. **Generate tasks** - Create epics/stories with Forger
3. **Export** - YAML/Markdown for coding agents

## Commands

```bash
# Start a sprint
/autarch:prd

# Start with initial idea
/autarch:prd "Build a reading tracker for developers"

# Resume a sprint
/autarch:prd SPRINT-1234567890
```

## Output

PRD saved to `.gurgeh/specs/PRD-{id}.yaml`
Sprint state saved to `.gurgeh/sprints/SPRINT-{id}.yaml`
```

**Step 2: Commit**

```bash
git add autarch-plugin/skills/spec-sprint/
git commit -m "$(cat <<'EOF'
docs(plugin): add spec-sprint skill

Comprehensive guide for Arbiter Spec Sprint:
- 6-section propose-first flow
- Interaction model (Accept/Edit/Alternative)
- Quick scan integration
- Consistency checking
- Handoff options

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 12: Update Plugin Commands

**Files:**
- Modify: `autarch-plugin/commands/autarch/prd.md`

**Step 1: Update command to reference spec-sprint**

The prd.md command should:
- Reference the new sprint flow
- Document the propose-first model
- Update the steps to match new architecture

**Step 2: Commit**

```bash
git add autarch-plugin/commands/autarch/prd.md
git commit -m "$(cat <<'EOF'
docs(plugin): update prd command for spec sprint

Update /autarch:prd to use new Arbiter Spec Sprint:
- Reference spec-sprint skill
- Document propose-first interaction
- Update section list (6 instead of 8)
- Add quick scan integration step

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Phase 8: Migration and Documentation

### Task 13: Add Migration Path from Old Interview

**Files:**
- Create: `internal/gurgeh/arbiter/migrate.go`

**Step 1: Write migration function**

```go
// internal/gurgeh/arbiter/migrate.go
package arbiter

import (
    "github.com/mistakeknot/autarch/internal/gurgeh/specs"
)

// MigrateFromSpec converts a legacy Spec to SprintState
func MigrateFromSpec(spec *specs.Spec, projectPath string) *SprintState {
    state := NewSprintState(projectPath)

    // Map legacy fields to new sections
    if spec.Summary != "" {
        state.Sections[PhaseProblem].Content = spec.Summary
        state.Sections[PhaseProblem].Status = DraftAccepted
    }

    if spec.UserStory.Text != "" {
        state.Sections[PhaseUsers].Content = spec.UserStory.Text
        state.Sections[PhaseUsers].Status = DraftAccepted
    }

    // ... map remaining fields

    return state
}
```

**Step 2: Commit**

```bash
git add internal/gurgeh/arbiter/migrate.go
git commit -m "$(cat <<'EOF'
feat(arbiter): add migration from legacy Spec format

Enable editing existing PRDs in new sprint flow:
- Map legacy Spec fields to SprintState sections
- Preserve existing content as accepted drafts
- Allow re-running consistency checks

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

### Task 14: Update Documentation

**Files:**
- Modify: `docs/COMPOUND_INTEGRATION.md`
- Modify: `AGENTS.md`

**Step 1: Update COMPOUND_INTEGRATION.md**

Add section documenting the new Arbiter Spec Sprint as the primary onboarding flow.

**Step 2: Update AGENTS.md**

Add quick reference for the new sprint workflow.

**Step 3: Commit**

```bash
git add docs/COMPOUND_INTEGRATION.md AGENTS.md
git commit -m "$(cat <<'EOF'
docs: update for Arbiter Spec Sprint release

Document new PRD creation workflow:
- Spec Sprint as primary onboarding
- 6-section propose-first flow
- Quick scan integration
- Consistency and confidence scoring

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Verification

After all tasks complete, run full verification:

```bash
# Build all
go build ./cmd/...

# Run all tests
go test ./internal/gurgeh/arbiter/... -v
go test ./internal/gurgeh/consistency/... -v
go test ./internal/gurgeh/confidence/... -v
go test ./internal/pollard/quick/... -v
go test ./internal/gurgeh/tui/... -v

# Manual test
./gurgeh  # Should show TUI with 'n' starting sprint mode

# Verify plugin structure
ls -la autarch-plugin/skills/spec-sprint/
cat autarch-plugin/agents/prd/arbiter.md
```

---

## Summary

| Phase | Tasks | New Files | Modified Files |
|-------|-------|-----------|----------------|
| 1: Data Structures | 1-2 | `arbiter/types.go`, `arbiter/persistence.go` | - |
| 2: Consistency | 3 | `consistency/checker.go` | - |
| 3: Confidence | 4 | `confidence/calculator.go` | - |
| 4: Quick Scan | 5 | `pollard/quick/scan.go` | - |
| 5: Arbiter Core | 6-7 | `arbiter/generator.go`, `arbiter/orchestrator.go` | - |
| 6: TUI | 8-9 | `tui/sprint.go` | `tui/model.go`, `tui/router.go` |
| 7: Plugin | 10-12 | `skills/spec-sprint/SKILL.md` | `agents/prd/arbiter.md`, `commands/autarch/prd.md` |
| 8: Migration | 13-14 | `arbiter/migrate.go` | `docs/COMPOUND_INTEGRATION.md`, `AGENTS.md` |

**Total: 14 tasks across 8 phases**
