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
	if p >= 0 && int(p) < len(names) {
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
	ConflictGoalFeature                     // Goal not supported by features
	ConflictScopeCreep                      // Feature contradicts non-goals
	ConflictAssumption                      // Assumption conflicts with other content
)

// Severity indicates if the conflict blocks progress
type Severity int

const (
	SeverityBlocker Severity = iota // Must resolve before continuing
	SeverityWarning                 // Can dismiss with acknowledgment
)
