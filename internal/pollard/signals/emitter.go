// Package signals provides Pollard-specific signal emission.
// Pollard emits competitor_shipped and research_invalidation signals
// by comparing scan results against linked spec assumptions.
package signals

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/mistakeknot/autarch/pkg/signals"
)

// Emitter creates signals from Pollard scan results.
type Emitter struct{}

// NewEmitter creates a Pollard signal emitter.
func NewEmitter() *Emitter {
	return &Emitter{}
}

// CompetitorShipped creates a signal when a competitor releases something relevant.
func (e *Emitter) CompetitorShipped(specID, competitor, detail string) signals.Signal {
	return signals.Signal{
		ID:        generateID(),
		Type:      signals.SignalCompetitorShipped,
		Source:    "pollard",
		SpecID:    specID,
		Severity:  signals.SeverityWarning,
		Title:     "Competitor shipped: " + competitor,
		Detail:    detail,
		CreatedAt: time.Now(),
	}
}

// ResearchInvalidation creates a signal when new research contradicts a spec assumption.
func (e *Emitter) ResearchInvalidation(specID, assumptionID, detail string) signals.Signal {
	return signals.Signal{
		ID:        generateID(),
		Type:      signals.SignalResearchInvalidation,
		Source:    "pollard",
		SpecID:    specID,
		Severity:  signals.SeverityCritical,
		Title:     "Research invalidates assumption " + assumptionID,
		Detail:    detail,
		CreatedAt: time.Now(),
	}
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "sig-" + hex.EncodeToString(b)
}
