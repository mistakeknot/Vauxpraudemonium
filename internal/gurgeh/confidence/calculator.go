package confidence

import (
	"regexp"
	"strings"

	"github.com/mistakeknot/autarch/internal/gurgeh/arbiter"
)

var (
	numberPattern  = regexp.MustCompile(`\d+`)
	metricPattern  = regexp.MustCompile(`(?i)(seconds?|minutes?|hours?|days?|weeks?|\$|%|users?|requests?|per)`)
	examplePattern = regexp.MustCompile(`(?i)(e\.g\.|for example|such as|like)`)
)

// Calculator computes a running confidence score for a PRD sprint.
type Calculator struct{}

// NewCalculator creates a new Calculator.
func NewCalculator() *Calculator {
	return &Calculator{}
}

// Calculate returns a ConfidenceScore for the current sprint state.
func (c *Calculator) Calculate(state *arbiter.SprintState) arbiter.ConfidenceScore {
	if state == nil {
		return arbiter.ConfidenceScore{}
	}
	if state.Sections == nil {
		return arbiter.ConfidenceScore{}
	}
	return arbiter.ConfidenceScore{
		Completeness: c.completeness(state),
		Consistency:  c.consistency(state),
		Specificity:  c.specificity(state),
		Research:     c.research(state),
		Assumptions:  c.assumptions(state),
	}
}

func (c *Calculator) completeness(state *arbiter.SprintState) float64 {
	total := float64(len(arbiter.AllPhases()))
	if total == 0 {
		return 0
	}
	filled := 0.0
	for _, phase := range arbiter.AllPhases() {
		section := state.Sections[phase]
		if section == nil {
			continue
		}
		if strings.TrimSpace(section.Content) != "" {
			filled += 0.5
		}
		if section.Status == arbiter.DraftAccepted {
			filled += 0.5
		}
	}
	return filled / total
}

func (c *Calculator) consistency(state *arbiter.SprintState) float64 {
	// No content means no consistency to evaluate
	hasContent := false
	for _, phase := range arbiter.AllPhases() {
		if s := state.Sections[phase]; s != nil && strings.TrimSpace(s.Content) != "" {
			hasContent = true
			break
		}
	}
	if !hasContent {
		return 0
	}
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
	score := 1.0 - (float64(blockers)*0.25 + float64(warnings)*0.1)
	if score < 0 {
		return 0
	}
	return score
}

func (c *Calculator) specificity(state *arbiter.SprintState) float64 {
	var totalScore float64
	var sections int
	for _, phase := range arbiter.AllPhases() {
		section := state.Sections[phase]
		if section == nil || strings.TrimSpace(section.Content) == "" {
			continue
		}
		sections++
		content := section.Content
		score := 0.0
		if numberPattern.MatchString(content) {
			score += 0.4
		}
		if metricPattern.MatchString(content) {
			score += 0.3
		}
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

func (c *Calculator) research(state *arbiter.SprintState) float64 {
	if state.ResearchCtx == nil {
		return 0
	}
	score := 0.0
	if len(state.ResearchCtx.GitHubHits) > 0 {
		score += 0.4
	}
	if len(state.ResearchCtx.HNHits) > 0 {
		score += 0.4
	}
	if strings.TrimSpace(state.ResearchCtx.Summary) != "" {
		score += 0.2
	}
	return score
}

func (c *Calculator) assumptions(state *arbiter.SprintState) float64 {
	scope := state.Sections[arbiter.PhaseScopeAssumptions]
	if scope == nil || strings.TrimSpace(scope.Content) == "" {
		return 0
	}
	content := strings.ToLower(scope.Content)
	score := 0.0
	if strings.Contains(content, "assumption") {
		score += 0.4
	}
	if strings.Contains(content, "if not") || strings.Contains(content, "otherwise") || strings.Contains(content, "impact") {
		score += 0.3
	}
	if strings.Contains(content, "confident") || strings.Contains(content, "likely") || strings.Contains(content, "uncertain") {
		score += 0.3
	}
	return score
}
