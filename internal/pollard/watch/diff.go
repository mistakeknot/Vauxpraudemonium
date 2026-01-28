package watch

import "fmt"

// WatchDiff captures the differences between two watch snapshots.
type WatchDiff struct {
	NewSources  int
	NewInsights int
	NewFiles    []string
	Summary     string
}

// diffSnapshots compares a previous and current snapshot.
// Returns nil if previous is nil (first run).
func diffSnapshots(previous, current *WatchSnapshot) *WatchDiff {
	if previous == nil {
		return &WatchDiff{
			NewSources:  current.TotalSources,
			NewInsights: current.TotalInsights,
			NewFiles:    current.OutputFiles,
			Summary:     fmt.Sprintf("First scan: %d sources, %d insights", current.TotalSources, current.TotalInsights),
		}
	}

	prevFiles := make(map[string]bool)
	for _, f := range previous.OutputFiles {
		prevFiles[f] = true
	}
	var newFiles []string
	for _, f := range current.OutputFiles {
		if !prevFiles[f] {
			newFiles = append(newFiles, f)
		}
	}

	newSources := current.TotalSources - previous.TotalSources
	newInsights := current.TotalInsights - previous.TotalInsights
	if newSources < 0 {
		newSources = 0
	}
	if newInsights < 0 {
		newInsights = 0
	}

	return &WatchDiff{
		NewSources:  newSources,
		NewInsights: newInsights,
		NewFiles:    newFiles,
		Summary:     fmt.Sprintf("Delta: +%d sources, +%d insights, %d new files", newSources, newInsights, len(newFiles)),
	}
}

// HasChanges returns true if anything new was found.
func (d *WatchDiff) HasChanges() bool {
	return d.NewSources > 0 || d.NewInsights > 0 || len(d.NewFiles) > 0
}
