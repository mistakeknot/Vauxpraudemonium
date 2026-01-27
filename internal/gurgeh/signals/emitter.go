// Package signals provides Gurgeh-specific signal emission.
// Gurgeh emits assumption_decayed, hypothesis_stale, and spec_health_low signals
// checked on spec load — no background process.
package signals

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
	"github.com/mistakeknot/autarch/pkg/signals"
)

// Emitter creates signals from Gurgeh spec state.
type Emitter struct{}

// NewEmitter creates a Gurgeh signal emitter.
func NewEmitter() *Emitter {
	return &Emitter{}
}

// CheckSpec evaluates a spec for all Gurgeh-emittable signals.
func (e *Emitter) CheckSpec(spec *specs.Spec) []signals.Signal {
	var result []signals.Signal
	result = append(result, e.checkAssumptionDecay(spec)...)
	result = append(result, e.checkHypothesisStale(spec)...)
	result = append(result, e.checkSpecHealth(spec)...)
	return result
}

func (e *Emitter) checkAssumptionDecay(spec *specs.Spec) []signals.Signal {
	decayed := specs.CheckAssumptionDecay(spec)
	var result []signals.Signal
	for _, a := range decayed {
		result = append(result, signals.Signal{
			ID:            generateID(),
			Type:          signals.SignalAssumptionDecayed,
			Source:        "gurgeh",
			SpecID:        spec.ID,
			AffectedField: "assumptions",
			Severity:      signals.SeverityWarning,
			Title:         fmt.Sprintf("Assumption %s decayed", a.ID),
			Detail:        fmt.Sprintf("%s — confidence dropped to %s", a.Description, a.Confidence),
			CreatedAt:     time.Now(),
		})
	}
	return result
}

func (e *Emitter) checkHypothesisStale(spec *specs.Spec) []signals.Signal {
	now := time.Now()
	specCreated, err := time.Parse(time.RFC3339, spec.CreatedAt)
	if err != nil {
		return nil
	}

	var result []signals.Signal
	for _, h := range spec.Hypotheses {
		if h.Status != "untested" || h.TimeboxDays <= 0 {
			continue
		}
		deadline := specCreated.Add(time.Duration(h.TimeboxDays) * 24 * time.Hour)
		if now.After(deadline) {
			result = append(result, signals.Signal{
				ID:            generateID(),
				Type:          signals.SignalHypothesisStale,
				Source:        "gurgeh",
				SpecID:        spec.ID,
				AffectedField: "hypotheses",
				Severity:      signals.SeverityWarning,
				Title:         fmt.Sprintf("Hypothesis %s is stale", h.ID),
				Detail:        fmt.Sprintf("%s — timebox of %d days expired", h.Statement, h.TimeboxDays),
				CreatedAt:     time.Now(),
			})
		}
	}
	return result
}

func (e *Emitter) checkSpecHealth(spec *specs.Spec) []signals.Signal {
	// Simple health check: specs with no goals, no requirements, or >50% low-confidence assumptions
	issues := 0
	if len(spec.Goals) == 0 {
		issues++
	}
	if len(spec.Requirements) == 0 && len(spec.StructuredRequirements) == 0 {
		issues++
	}
	lowConf := 0
	for _, a := range spec.Assumptions {
		if a.Confidence == "low" {
			lowConf++
		}
	}
	if len(spec.Assumptions) > 0 && float64(lowConf)/float64(len(spec.Assumptions)) > 0.5 {
		issues++
	}

	if issues >= 2 {
		return []signals.Signal{{
			ID:            generateID(),
			Type:          signals.SignalSpecHealthLow,
			Source:        "gurgeh",
			SpecID:        spec.ID,
			AffectedField: "health",
			Severity:      signals.SeverityCritical,
			Title:         "Spec health is low",
			Detail:        fmt.Sprintf("%d health issues detected: missing goals/requirements or majority low-confidence assumptions", issues),
			CreatedAt:     time.Now(),
		}}
	}
	return nil
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "sig-" + hex.EncodeToString(b)
}
