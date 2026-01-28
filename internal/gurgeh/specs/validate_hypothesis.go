package specs

// ValidateHypothesis marks a hypothesis as validated or invalidated.
func ValidateHypothesis(spec *Spec, hypothesisID, status, result string) bool {
	for i := range spec.Hypotheses {
		if spec.Hypotheses[i].ID == hypothesisID {
			spec.Hypotheses[i].Status = status
			spec.Hypotheses[i].Result = result
			return true
		}
	}
	return false
}

// StaleHypotheses returns hypotheses that are past their timebox and still untested.
func StaleHypotheses(spec *Spec) []Hypothesis {
	var stale []Hypothesis
	for _, h := range spec.Hypotheses {
		if h.Status == "untested" && h.TimeboxDays > 0 {
			// Staleness is checked by signal emitter using spec.CreatedAt
			stale = append(stale, h)
		}
	}
	return stale
}
