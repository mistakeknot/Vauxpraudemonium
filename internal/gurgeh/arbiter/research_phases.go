package arbiter

// PhaseResearchConfig maps an Arbiter sprint phase to the Pollard hunters,
// scan depth, and query extraction strategy needed for that phase.
type PhaseResearchConfig struct {
	Phase   Phase
	Hunters []string
	Mode    string // "quick", "balanced", "deep"
	// QueryExtractor returns a research query from the current sprint state.
	QueryExtractor func(state *SprintState) string
}

// DefaultResearchPlan returns the phase-specific research configurations.
// Not every phase triggers research — only those where evidence strengthens the spec.
func DefaultResearchPlan() []PhaseResearchConfig {
	return []PhaseResearchConfig{
		{
			Phase:   PhaseVision,
			Hunters: []string{"github-scout", "hackernews-trendwatcher"},
			Mode:    "quick",
			QueryExtractor: func(state *SprintState) string {
				return extractSectionContent(state, PhaseVision)
			},
		},
		{
			Phase:   PhaseProblem,
			Hunters: []string{"arxiv-scout", "openalex"},
			Mode:    "balanced",
			QueryExtractor: func(state *SprintState) string {
				return extractSectionContent(state, PhaseProblem)
			},
		},
		{
			Phase:   PhaseFeaturesGoals,
			Hunters: []string{"competitor-tracker", "github-scout"},
			Mode:    "deep",
			QueryExtractor: func(state *SprintState) string {
				return extractSectionContent(state, PhaseFeaturesGoals)
			},
		},
		{
			Phase:   PhaseRequirements,
			Hunters: []string{"github-scout"},
			Mode:    "balanced",
			QueryExtractor: func(state *SprintState) string {
				return extractSectionContent(state, PhaseRequirements)
			},
		},
	}
}

// ResearchConfigForPhase returns the research config for a phase, or nil if none.
func ResearchConfigForPhase(phase Phase) *PhaseResearchConfig {
	for _, cfg := range DefaultResearchPlan() {
		if cfg.Phase == phase {
			return &cfg
		}
	}
	return nil
}

// extractSectionContent returns truncated content from a sprint section.
func extractSectionContent(state *SprintState, phase Phase) string {
	if section, ok := state.Sections[phase]; ok && section.Content != "" {
		content := section.Content
		if len(content) > 200 {
			content = content[:200]
		}
		return content
	}
	return ""
}
