package signals

import (
	"time"

	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
)

// DefaultSignalThreshold is the number of active signals that triggers a review.
const DefaultSignalThreshold = 3

// ReviewStatus describes whether a vision spec needs review and why.
type ReviewStatus struct {
	NeedsReview bool   `json:"needs_review"`
	Reason      string `json:"reason"` // "cadence_exceeded", "signal_threshold", "both", ""
	DaysSince   int    `json:"days_since"`
	SignalCount int    `json:"signal_count"`
}

// CheckReviewNeeded evaluates both time-based and signal-based review triggers.
// Called on vision spec load (check-on-load pattern, no daemon).
func CheckReviewNeeded(spec *specs.Spec, store *Store) ReviewStatus {
	status := ReviewStatus{}

	if !spec.IsVision() {
		return status
	}

	// Time check
	cadence := spec.EffectiveReviewCadenceDays()
	if spec.LastReviewedAt != "" {
		if lastReview, err := time.Parse(time.RFC3339, spec.LastReviewedAt); err == nil {
			status.DaysSince = int(time.Since(lastReview).Hours() / 24)
			if status.DaysSince >= cadence {
				status.NeedsReview = true
				status.Reason = "cadence_exceeded"
			}
		}
	} else {
		// Never reviewed — check against creation date
		if created, err := time.Parse(time.RFC3339, spec.CreatedAt); err == nil {
			status.DaysSince = int(time.Since(created).Hours() / 24)
			if status.DaysSince >= cadence {
				status.NeedsReview = true
				status.Reason = "cadence_exceeded"
			}
		}
	}

	// Signal check
	if store != nil {
		status.SignalCount = store.Count(spec.ID)
		if status.SignalCount >= DefaultSignalThreshold {
			if status.NeedsReview {
				status.Reason = "both"
			} else {
				status.NeedsReview = true
				status.Reason = "signal_threshold"
			}
		}
	}

	return status
}
