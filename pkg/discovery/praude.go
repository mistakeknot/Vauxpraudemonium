package discovery

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PraudeSpec represents a PRD spec from Praude.
type PraudeSpec struct {
	ID           string   `yaml:"id"`
	Title        string   `yaml:"title"`
	Status       string   `yaml:"status"`
	Summary      string   `yaml:"summary"`
	Requirements []string `yaml:"requirements"`
	Acceptance   []struct {
		ID          string `yaml:"id"`
		Description string `yaml:"description"`
	} `yaml:"acceptance_criteria"`
	CriticalUserJourneys []struct {
		ID       string   `yaml:"id"`
		Title    string   `yaml:"title"`
		Priority string   `yaml:"priority"`
		Steps    []string `yaml:"steps"`
	} `yaml:"critical_user_journeys"`
}

// PraudePRD represents a PRD document (with features).
type PraudePRD struct {
	ID       string `yaml:"id"`
	Title    string `yaml:"title"`
	Version  string `yaml:"version"`
	Status   string `yaml:"status"`
	Features []struct {
		ID                 string   `yaml:"id"`
		Title              string   `yaml:"title"`
		Status             string   `yaml:"status"`
		Summary            string   `yaml:"summary"`
		Requirements       []string `yaml:"requirements"`
		AcceptanceCriteria []struct {
			ID          string `yaml:"id"`
			Description string `yaml:"description"`
		} `yaml:"acceptance_criteria"`
	} `yaml:"features"`
}

// PraudeSpecs loads all specs from a project's .praude/specs directory.
func PraudeSpecs(root string) ([]PraudeSpec, error) {
	specsDir := filepath.Join(root, ".praude", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []PraudeSpec{}, nil
		}
		return nil, err
	}

	var specs []PraudeSpec
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(specsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var spec PraudeSpec
		if err := yaml.Unmarshal(data, &spec); err != nil {
			continue
		}
		// Skip if it looks like a PRD (has version field)
		var prd PraudePRD
		if yaml.Unmarshal(data, &prd) == nil && prd.Version != "" {
			continue
		}
		if spec.ID != "" {
			specs = append(specs, spec)
		}
	}
	return specs, nil
}

// PraudePRDs loads all PRDs from a project's .praude/specs directory.
func PraudePRDs(root string) ([]PraudePRD, error) {
	specsDir := filepath.Join(root, ".praude", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []PraudePRD{}, nil
		}
		return nil, err
	}

	var prds []PraudePRD
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(specsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var prd PraudePRD
		if err := yaml.Unmarshal(data, &prd); err != nil {
			continue
		}
		// Only include files with the PRD structure (has Version field)
		if prd.Version != "" {
			prds = append(prds, prd)
		}
	}
	return prds, nil
}

// FindPRD finds a PRD by ID or version name.
func FindPRD(root, idOrVersion string) (*PraudePRD, error) {
	prds, err := PraudePRDs(root)
	if err != nil {
		return nil, err
	}
	for _, prd := range prds {
		if prd.ID == idOrVersion || prd.Version == idOrVersion {
			return &prd, nil
		}
	}
	return nil, nil
}

// FindSpec finds a spec by ID.
func FindSpec(root, id string) (*PraudeSpec, error) {
	specs, err := PraudeSpecs(root)
	if err != nil {
		return nil, err
	}
	for _, spec := range specs {
		if spec.ID == id {
			return &spec, nil
		}
	}
	return nil, nil
}

// CountPraudeSpecs returns the total number of specs.
func CountPraudeSpecs(root string) int {
	specs, _ := PraudeSpecs(root)
	return len(specs)
}

// CountPraudePRDs returns the total number of PRDs.
func CountPraudePRDs(root string) int {
	prds, _ := PraudePRDs(root)
	return len(prds)
}

// CountPRDFeatures returns the total number of features across all PRDs.
func CountPRDFeatures(root string) int {
	prds, _ := PraudePRDs(root)
	count := 0
	for _, prd := range prds {
		count += len(prd.Features)
	}
	return count
}

// PraudeHasData returns true if Praude has any specs or PRDs.
func PraudeHasData(root string) bool {
	return CountPraudeSpecs(root) > 0 || CountPraudePRDs(root) > 0
}

// GetApprovedFeatures returns all features with approved or in_progress status.
func GetApprovedFeatures(root string) ([]struct {
	PRDID   string
	Feature struct {
		ID      string
		Title   string
		Status  string
		Summary string
	}
}, error) {
	prds, err := PraudePRDs(root)
	if err != nil {
		return nil, err
	}

	var features []struct {
		PRDID   string
		Feature struct {
			ID      string
			Title   string
			Status  string
			Summary string
		}
	}

	for _, prd := range prds {
		for _, f := range prd.Features {
			if f.Status == "approved" || f.Status == "in_progress" {
				features = append(features, struct {
					PRDID   string
					Feature struct {
						ID      string
						Title   string
						Status  string
						Summary string
					}
				}{
					PRDID: prd.ID,
					Feature: struct {
						ID      string
						Title   string
						Status  string
						Summary string
					}{
						ID:      f.ID,
						Title:   f.Title,
						Status:  f.Status,
						Summary: f.Summary,
					},
				})
			}
		}
	}
	return features, nil
}
