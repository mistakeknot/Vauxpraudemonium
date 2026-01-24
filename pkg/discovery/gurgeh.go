package discovery

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// GurgDir is the config directory for gurgeh
const GurgDir = ".gurgeh"

// LegacyPraudeDir is the legacy config directory for backward compatibility
const LegacyPraudeDir = ".praude"

// gurgRootDir returns the specs directory, checking .gurgeh first then .praude
func gurgRootDir(root string) string {
	gurgPath := filepath.Join(root, GurgDir)
	if _, err := os.Stat(gurgPath); err == nil {
		return gurgPath
	}
	praudePath := filepath.Join(root, LegacyPraudeDir)
	if _, err := os.Stat(praudePath); err == nil {
		return praudePath
	}
	return gurgPath // default to new path
}

// GurgSpec represents a PRD spec from Gurgeh.
type GurgSpec struct {
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

// GurgPRD represents a PRD document (with features).
type GurgPRD struct {
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

// GurgSpecs loads all specs from a project's .gurgeh/specs directory (or .praude for legacy).
func GurgSpecs(root string) ([]GurgSpec, error) {
	specsDir := filepath.Join(gurgRootDir(root), "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []GurgSpec{}, nil
		}
		return nil, err
	}

	var specs []GurgSpec
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(specsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var spec GurgSpec
		if err := yaml.Unmarshal(data, &spec); err != nil {
			continue
		}
		// Skip if it looks like a PRD (has version field)
		var prd GurgPRD
		if yaml.Unmarshal(data, &prd) == nil && prd.Version != "" {
			continue
		}
		if spec.ID != "" {
			specs = append(specs, spec)
		}
	}
	return specs, nil
}

// GurgPRDs loads all PRDs from a project's .gurgeh/specs directory (or .praude for legacy).
func GurgPRDs(root string) ([]GurgPRD, error) {
	specsDir := filepath.Join(gurgRootDir(root), "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []GurgPRD{}, nil
		}
		return nil, err
	}

	var prds []GurgPRD
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(specsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var prd GurgPRD
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
func FindPRD(root, idOrVersion string) (*GurgPRD, error) {
	prds, err := GurgPRDs(root)
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
func FindSpec(root, id string) (*GurgSpec, error) {
	specs, err := GurgSpecs(root)
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

// CountGurgSpecs returns the total number of specs.
func CountGurgSpecs(root string) int {
	specs, _ := GurgSpecs(root)
	return len(specs)
}

// CountGurgPRDs returns the total number of PRDs.
func CountGurgPRDs(root string) int {
	prds, _ := GurgPRDs(root)
	return len(prds)
}

// CountPRDFeatures returns the total number of features across all PRDs.
func CountPRDFeatures(root string) int {
	prds, _ := GurgPRDs(root)
	count := 0
	for _, prd := range prds {
		count += len(prd.Features)
	}
	return count
}

// GurgHasData returns true if Gurgeh has any specs or PRDs.
func GurgHasData(root string) bool {
	return CountGurgSpecs(root) > 0 || CountGurgPRDs(root) > 0
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
	prds, err := GurgPRDs(root)
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
