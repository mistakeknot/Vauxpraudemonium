package specs

// NeedsVisionSpec returns true when at least one PRD exists but no vision spec
// has been created yet. This gates users into creating a vision spec before
// their second PRD.
func NeedsVisionSpec(summaries []Summary) bool {
	hasPRD := false
	hasVision := false
	for _, s := range summaries {
		switch s.Type {
		case SpecTypeVision:
			hasVision = true
		default:
			hasPRD = true
		}
	}
	return hasPRD && !hasVision
}
