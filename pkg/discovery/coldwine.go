package discovery

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const coldwineDir = ".coldwine"
const legacyTandemoniumDir = ".tandemonium"

func coldwineRootDir(root string) string {
	path := filepath.Join(root, coldwineDir)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	legacyPath := filepath.Join(root, legacyTandemoniumDir)
	if _, err := os.Stat(legacyPath); err == nil {
		return legacyPath
	}
	return path
}

// ColdwineEpic represents an epic from Coldwine.
type ColdwineEpic struct {
	ID                 string `yaml:"id"`
	Title              string `yaml:"title"`
	Summary            string `yaml:"summary"`
	Status             string `yaml:"status"`
	Priority           string `yaml:"priority"`
	AcceptanceCriteria []string `yaml:"acceptance_criteria"`
	Stories            []struct {
		ID                 string   `yaml:"id"`
		Title              string   `yaml:"title"`
		Status             string   `yaml:"status"`
		Priority           string   `yaml:"priority"`
		AcceptanceCriteria []string `yaml:"acceptance_criteria"`
	} `yaml:"stories"`
}

// ColdwineEpics loads all epics from a project's .coldwine/specs directory (or .tandemonium for legacy).
func ColdwineEpics(root string) ([]ColdwineEpic, error) {
	specsDir := filepath.Join(coldwineRootDir(root), "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ColdwineEpic{}, nil
		}
		return nil, err
	}

	var epics []ColdwineEpic
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(specsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var epic ColdwineEpic
		if err := yaml.Unmarshal(data, &epic); err != nil {
			continue
		}
		if epic.ID != "" {
			epics = append(epics, epic)
		}
	}
	return epics, nil
}

// FindEpic finds an epic by ID.
func FindEpic(root, id string) (*ColdwineEpic, error) {
	epics, err := ColdwineEpics(root)
	if err != nil {
		return nil, err
	}
	for _, epic := range epics {
		if epic.ID == id {
			return &epic, nil
		}
	}
	return nil, nil
}

// CountColdwineEpics returns the total number of epics.
func CountColdwineEpics(root string) int {
	epics, _ := ColdwineEpics(root)
	return len(epics)
}

// CountColdwineStories returns the total number of stories across all epics.
func CountColdwineStories(root string) int {
	epics, _ := ColdwineEpics(root)
	count := 0
	for _, e := range epics {
		count += len(e.Stories)
	}
	return count
}

// ColdwineHasData returns true if Coldwine has any epics.
func ColdwineHasData(root string) bool {
	return CountColdwineEpics(root) > 0
}

// EpicsByStatus returns epics grouped by status.
func EpicsByStatus(root string) (map[string][]ColdwineEpic, error) {
	epics, err := ColdwineEpics(root)
	if err != nil {
		return nil, err
	}

	byStatus := make(map[string][]ColdwineEpic)
	for _, e := range epics {
		status := e.Status
		if status == "" {
			status = "unknown"
		}
		byStatus[status] = append(byStatus[status], e)
	}
	return byStatus, nil
}

// BlockedEpics returns epics with blocked status.
func BlockedEpics(root string) ([]ColdwineEpic, error) {
	epics, err := ColdwineEpics(root)
	if err != nil {
		return nil, err
	}

	var blocked []ColdwineEpic
	for _, e := range epics {
		if e.Status == "blocked" {
			blocked = append(blocked, e)
		}
	}
	return blocked, nil
}

// EpicsWithoutAcceptanceCriteria returns epics missing acceptance criteria.
func EpicsWithoutAcceptanceCriteria(root string) ([]ColdwineEpic, error) {
	epics, err := ColdwineEpics(root)
	if err != nil {
		return nil, err
	}

	var missing []ColdwineEpic
	for _, e := range epics {
		if len(e.AcceptanceCriteria) == 0 {
			missing = append(missing, e)
		}
	}
	return missing, nil
}
