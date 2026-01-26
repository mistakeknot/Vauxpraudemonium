package cuj

import (
	"fmt"
	"time"

	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
)

// FromLegacy converts a legacy CriticalUserJourney (from specs.Spec) to a first-class CUJ.
// This enables migration of existing specs to use the new CUJ entity model.
func FromLegacy(legacy specs.CriticalUserJourney, specID, project string) *CUJ {
	// Convert steps from simple strings to structured Step objects
	steps := make([]Step, len(legacy.Steps))
	for i, stepStr := range legacy.Steps {
		steps[i] = Step{
			Order:    i + 1,
			Action:   stepStr,
			Expected: "", // Legacy format doesn't have expected outcomes
		}
	}

	// Map legacy priority string to Priority type
	priority := PriorityMedium
	switch legacy.Priority {
	case "high":
		priority = PriorityHigh
	case "low":
		priority = PriorityLow
	case "medium":
		priority = PriorityMedium
	}

	return &CUJ{
		ID:              legacy.ID,
		SpecID:          specID,
		Project:         project,
		Title:           legacy.Title,
		Priority:        priority,
		Steps:           steps,
		SuccessCriteria: legacy.SuccessCriteria,
		Status:          StatusDraft,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// ToLegacy converts a first-class CUJ back to the legacy format for backward compatibility.
// This is useful when writing specs that need to maintain the old format.
func ToLegacy(cuj *CUJ) specs.CriticalUserJourney {
	// Convert structured steps back to simple strings
	steps := make([]string, len(cuj.Steps))
	for i, step := range cuj.Steps {
		steps[i] = step.Action
	}

	// Convert priority to string
	priority := string(cuj.Priority)

	// Compute linked requirements from feature links (if needed)
	// Note: This is a placeholder - actual implementation would need to
	// query feature links and extract requirement IDs
	var linkedReqs []string

	return specs.CriticalUserJourney{
		ID:                 cuj.ID,
		Title:              cuj.Title,
		Priority:           priority,
		Steps:              steps,
		SuccessCriteria:    cuj.SuccessCriteria,
		LinkedRequirements: linkedReqs,
	}
}

// MigrateSpecCUJs extracts CUJs from a spec and returns them as first-class entities.
// This is used during the migration from embedded CUJs to standalone entities.
func MigrateSpecCUJs(spec *specs.Spec, project string) []*CUJ {
	cujs := make([]*CUJ, len(spec.CriticalUserJourneys))
	for i, legacy := range spec.CriticalUserJourneys {
		cujs[i] = FromLegacy(legacy, spec.ID, project)
	}
	return cujs
}

// GenerateID generates a unique CUJ ID based on spec ID and sequence number.
func GenerateID(specID string, seq int) string {
	return fmt.Sprintf("CUJ-%s-%03d", specID, seq)
}

// EnrichWithFeatureContext adds feature context to CUJs based on a PRD's features.
// This helps establish the relationship between CUJs and the features they support.
func EnrichWithFeatureContext(cujs []*CUJ, prd *specs.PRD) map[string][]string {
	// Build a map of CUJ ID -> feature IDs that reference it
	cujFeatures := make(map[string][]string)

	// In the PRD model, features have a CriticalUserJourneys field
	// We need to match them up
	for _, feature := range prd.Features {
		for _, legacyCUJ := range feature.CriticalUserJourneys {
			cujFeatures[legacyCUJ.ID] = append(cujFeatures[legacyCUJ.ID], feature.ID)
		}
	}

	return cujFeatures
}

// MigratePRDCUJs extracts all CUJs from a PRD (across all features) and returns them as first-class entities.
func MigratePRDCUJs(prd *specs.PRD, project string) []*CUJ {
	var cujs []*CUJ
	seen := make(map[string]bool) // Avoid duplicates if same CUJ appears in multiple features

	for _, feature := range prd.Features {
		for _, legacy := range feature.CriticalUserJourneys {
			if seen[legacy.ID] {
				continue
			}
			seen[legacy.ID] = true
			cuj := FromLegacy(legacy, prd.ID, project)
			cujs = append(cujs, cuj)
		}
	}
	return cujs
}
