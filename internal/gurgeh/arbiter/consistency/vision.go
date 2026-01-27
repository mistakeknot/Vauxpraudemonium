package consistency

import "strings"

// VisionInfo holds vision spec sections for vertical alignment checks.
type VisionInfo struct {
	Goals       []string // vision principles
	Assumptions []string // strategic bets
}

// ConflictTypeVisionAlignment is the TypeCode for vision alignment conflicts.
const ConflictTypeVisionAlignment = 4

// CheckVisionAlignment checks PRD sections against vision spec fields.
// Returns warnings (severity=1) only — vision conflicts never block.
//
// Two checks in v1:
//   - PhaseProblem (1) ↔ Goals: problem should reference at least one vision principle
//   - PhaseFeaturesGoals (3) ↔ Assumptions: features should not contradict strategic bets
func CheckVisionAlignment(vision *VisionInfo, sections map[int]*SectionInfo) []Conflict {
	if vision == nil {
		return nil
	}

	var conflicts []Conflict

	// Check 1: Problem ↔ Goals — problem should relate to at least one principle
	if problem, ok := sections[1]; ok && problem.Accepted && len(vision.Goals) > 0 {
		if !anyKeywordOverlap(problem.Content, vision.Goals) {
			conflicts = append(conflicts, Conflict{
				TypeCode: ConflictTypeVisionAlignment,
				Severity: 1, // warning, not blocker
				Message:  "Problem statement doesn't reference any vision principles",
				Sections: []int{1},
			})
		}
	}

	// Check 2: Features ↔ Assumptions — features should not contradict strategic bets
	if features, ok := sections[3]; ok && features.Accepted && len(vision.Assumptions) > 0 {
		for _, bet := range vision.Assumptions {
			if detectContradiction(features.Content, bet) {
				conflicts = append(conflicts, Conflict{
					TypeCode: ConflictTypeVisionAlignment,
					Severity: 1,
					Message:  "Feature may contradict strategic bet: " + truncate(bet, 80),
					Sections: []int{3},
				})
			}
		}
	}

	return conflicts
}

// anyKeywordOverlap returns true if the content shares meaningful words
// with at least one of the reference strings.
func anyKeywordOverlap(content string, references []string) bool {
	contentWords := significantWords(content)
	for _, ref := range references {
		refWords := significantWords(ref)
		for w := range refWords {
			if contentWords[w] {
				return true
			}
		}
	}
	return false
}

// detectContradiction returns true if content contains negation patterns
// combined with keywords from the bet.
func detectContradiction(content, bet string) bool {
	contentLower := strings.ToLower(content)
	betWords := significantWords(bet)

	negations := []string{"not ", "never ", "won't ", "don't ", "no ", "without "}
	for _, neg := range negations {
		idx := strings.Index(contentLower, neg)
		if idx < 0 {
			continue
		}
		// Check if any bet keyword appears near the negation (within 100 chars)
		vicinity := contentLower[idx:]
		if len(vicinity) > 120 {
			vicinity = vicinity[:120]
		}
		for w := range betWords {
			if strings.Contains(vicinity, w) {
				return true
			}
		}
	}
	return false
}

// significantWords extracts lowercase words with 4+ characters.
var stopWords = map[string]bool{
	"that": true, "this": true, "with": true, "from": true, "have": true,
	"will": true, "been": true, "they": true, "were": true, "their": true,
	"what": true, "when": true, "where": true, "which": true, "about": true,
	"would": true, "there": true, "could": true, "should": true, "other": true,
	"each": true, "than": true, "them": true, "into": true, "more": true,
	"also": true, "some": true, "such": true, "only": true, "then": true,
}

func significantWords(s string) map[string]bool {
	words := make(map[string]bool)
	for _, w := range strings.Fields(strings.ToLower(s)) {
		w = strings.Trim(w, ".,;:!?\"'()[]{}—-")
		if len(w) >= 4 && !stopWords[w] {
			words[w] = true
		}
	}
	return words
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
