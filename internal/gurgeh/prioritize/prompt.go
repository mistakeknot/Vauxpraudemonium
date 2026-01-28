package prioritize

import (
	"fmt"
	"strings"
)

// BuildRankingPrompt constructs the agent prompt for feature ranking.
func BuildRankingPrompt(input RankingInput) string {
	var sb strings.Builder

	sb.WriteString("You are a product strategy agent. Given the following context, rank features by what to build next.\n\n")

	// Spec context
	if input.Spec != nil {
		sb.WriteString("## Spec: " + input.Spec.Title + "\n\n")
		if input.Spec.Summary != "" {
			sb.WriteString("Summary: " + input.Spec.Summary + "\n\n")
		}

		if len(input.Spec.Goals) > 0 {
			sb.WriteString("### Goals\n")
			for _, g := range input.Spec.Goals {
				sb.WriteString(fmt.Sprintf("- %s: %s (metric: %s, target: %s)\n", g.ID, g.Description, g.Metric, g.Target))
			}
			sb.WriteString("\n")
		}

		if len(input.Spec.Hypotheses) > 0 {
			sb.WriteString("### Hypotheses\n")
			for _, h := range input.Spec.Hypotheses {
				sb.WriteString(fmt.Sprintf("- %s [%s]: %s (timebox: %d days)\n", h.ID, h.Status, h.Statement, h.TimeboxDays))
			}
			sb.WriteString("\n")
		}
	}

	// Active signals
	if len(input.Signals) > 0 {
		sb.WriteString("### Active Signals\n")
		for _, s := range input.Signals {
			sb.WriteString(fmt.Sprintf("- [%s] %s: %s (%s)\n", s.Severity, s.Type, s.Title, s.Detail))
		}
		sb.WriteString("\n")
	}

	// Research
	if len(input.Research) > 0 {
		sb.WriteString("### Research Findings\n")
		for _, r := range input.Research {
			sb.WriteString(fmt.Sprintf("- %s (%s, relevance: %.1f)\n", r.Title, r.SourceType, r.Relevance))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`## Instructions

Rank features by what to build next. For each, explain why in 2-3 sentences.
Consider: signal urgency, research backing, execution readiness, hypothesis value.

Output as JSON array of objects with fields: feature_id, title, rank, reasoning, confidence (0-1).
`)

	return sb.String()
}
