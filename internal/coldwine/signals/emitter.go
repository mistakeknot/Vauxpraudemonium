// Package signals provides Coldwine-specific signal emission.
// Coldwine emits execution_drift signals when tasks significantly exceed
// estimates or when agents repeatedly fail on the same story.
package signals

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/mistakeknot/autarch/pkg/signals"
)

// Emitter creates signals from Coldwine execution state.
type Emitter struct{}

// NewEmitter creates a Coldwine signal emitter.
func NewEmitter() *Emitter {
	return &Emitter{}
}

// TaskDurationDrift detects when a task takes >3x its estimate.
func (e *Emitter) TaskDurationDrift(specID, taskID, taskTitle string, estimatedMinutes, actualMinutes int) *signals.Signal {
	if estimatedMinutes <= 0 || actualMinutes <= 3*estimatedMinutes {
		return nil
	}
	s := signals.Signal{
		ID:        generateID(),
		Type:      signals.SignalExecutionDrift,
		Source:    "coldwine",
		SpecID:    specID,
		Severity:  signals.SeverityWarning,
		Title:     fmt.Sprintf("Task %s exceeded estimate by %.0fx", taskID, float64(actualMinutes)/float64(estimatedMinutes)),
		Detail:    fmt.Sprintf("%s: estimated %dm, actual %dm", taskTitle, estimatedMinutes, actualMinutes),
		CreatedAt: time.Now(),
	}
	return &s
}

// AgentFailureDrift detects when >2 agent failures occur on the same story.
func (e *Emitter) AgentFailureDrift(specID, storyID, storyTitle string, failureCount int) *signals.Signal {
	if failureCount <= 2 {
		return nil
	}
	s := signals.Signal{
		ID:        generateID(),
		Type:      signals.SignalExecutionDrift,
		Source:    "coldwine",
		SpecID:    specID,
		Severity:  signals.SeverityCritical,
		Title:     fmt.Sprintf("Story %s has %d agent failures", storyID, failureCount),
		Detail:    fmt.Sprintf("%s: repeated agent failures suggest the task may need redesign", storyTitle),
		CreatedAt: time.Now(),
	}
	return &s
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "sig-" + hex.EncodeToString(b)
}
