package signals

import (
	"testing"
	"time"

	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
	"github.com/mistakeknot/autarch/pkg/signals"
)

func TestCheckReviewNeeded_NonVision(t *testing.T) {
	spec := &specs.Spec{ID: "PRD-1", Type: "prd"}
	status := CheckReviewNeeded(spec, nil)
	if status.NeedsReview {
		t.Error("non-vision specs should never need review")
	}
}

func TestCheckReviewNeeded_CadenceExceeded(t *testing.T) {
	spec := &specs.Spec{
		ID:                "VIS-1",
		Type:              "vision",
		LastReviewedAt:    time.Now().Add(-40 * 24 * time.Hour).Format(time.RFC3339),
		ReviewCadenceDays: 30,
	}
	status := CheckReviewNeeded(spec, nil)
	if !status.NeedsReview {
		t.Error("expected review needed when cadence exceeded")
	}
	if status.Reason != "cadence_exceeded" {
		t.Errorf("expected reason cadence_exceeded, got %s", status.Reason)
	}
	if status.DaysSince < 39 {
		t.Errorf("expected ~40 days since, got %d", status.DaysSince)
	}
}

func TestCheckReviewNeeded_CadenceNotExceeded(t *testing.T) {
	spec := &specs.Spec{
		ID:                "VIS-1",
		Type:              "vision",
		LastReviewedAt:    time.Now().Add(-10 * 24 * time.Hour).Format(time.RFC3339),
		ReviewCadenceDays: 30,
	}
	status := CheckReviewNeeded(spec, nil)
	if status.NeedsReview {
		t.Error("review should not be needed when within cadence")
	}
}

func TestCheckReviewNeeded_SignalThreshold(t *testing.T) {
	store := testStore(t)

	// Add 3 signals (threshold)
	for i, field := range []string{"goals", "assumptions", "scope"} {
		store.Emit(signals.Signal{
			ID:            generateID(),
			Type:          signals.SignalAssumptionDecayed,
			Source:        "gurgeh",
			SpecID:        "VIS-1",
			AffectedField: field,
			Severity:      signals.SeverityWarning,
			Detail:        "test",
			CreatedAt:     time.Now(),
		})
		_ = i
	}

	spec := &specs.Spec{
		ID:             "VIS-1",
		Type:           "vision",
		LastReviewedAt: time.Now().Add(-5 * 24 * time.Hour).Format(time.RFC3339),
	}

	status := CheckReviewNeeded(spec, store)
	if !status.NeedsReview {
		t.Error("expected review needed when signal threshold reached")
	}
	if status.Reason != "signal_threshold" {
		t.Errorf("expected reason signal_threshold, got %s", status.Reason)
	}
	if status.SignalCount != 3 {
		t.Errorf("expected signal count 3, got %d", status.SignalCount)
	}
}

func TestCheckReviewNeeded_Both(t *testing.T) {
	store := testStore(t)

	for _, field := range []string{"goals", "assumptions", "scope"} {
		store.Emit(signals.Signal{
			ID:            generateID(),
			Type:          signals.SignalAssumptionDecayed,
			Source:        "gurgeh",
			SpecID:        "VIS-1",
			AffectedField: field,
			Severity:      signals.SeverityWarning,
			Detail:        "test",
			CreatedAt:     time.Now(),
		})
	}

	spec := &specs.Spec{
		ID:                "VIS-1",
		Type:              "vision",
		LastReviewedAt:    time.Now().Add(-40 * 24 * time.Hour).Format(time.RFC3339),
		ReviewCadenceDays: 30,
	}

	status := CheckReviewNeeded(spec, store)
	if !status.NeedsReview {
		t.Error("expected review needed")
	}
	if status.Reason != "both" {
		t.Errorf("expected reason both, got %s", status.Reason)
	}
}

func TestCheckReviewNeeded_NeverReviewed(t *testing.T) {
	spec := &specs.Spec{
		ID:        "VIS-1",
		Type:      "vision",
		CreatedAt: time.Now().Add(-45 * 24 * time.Hour).Format(time.RFC3339),
	}
	status := CheckReviewNeeded(spec, nil)
	if !status.NeedsReview {
		t.Error("expected review needed for spec never reviewed and past cadence")
	}
}
