package discovery

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// TandemoniumEpic represents an epic from Tandemonium.
type TandemoniumEpic struct {
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

// TandemoniumEpics loads all epics from a project's .tandemonium/specs directory.
func TandemoniumEpics(root string) ([]TandemoniumEpic, error) {
	specsDir := filepath.Join(root, ".tandemonium", "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []TandemoniumEpic{}, nil
		}
		return nil, err
	}

	var epics []TandemoniumEpic
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(specsDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var epic TandemoniumEpic
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
func FindEpic(root, id string) (*TandemoniumEpic, error) {
	epics, err := TandemoniumEpics(root)
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

// CountTandemoniumEpics returns the total number of epics.
func CountTandemoniumEpics(root string) int {
	epics, _ := TandemoniumEpics(root)
	return len(epics)
}

// CountTandemoniumStories returns the total number of stories across all epics.
func CountTandemoniumStories(root string) int {
	epics, _ := TandemoniumEpics(root)
	count := 0
	for _, e := range epics {
		count += len(e.Stories)
	}
	return count
}

// TandemoniumHasData returns true if Tandemonium has any epics.
func TandemoniumHasData(root string) bool {
	return CountTandemoniumEpics(root) > 0
}

// EpicsByStatus returns epics grouped by status.
func EpicsByStatus(root string) (map[string][]TandemoniumEpic, error) {
	epics, err := TandemoniumEpics(root)
	if err != nil {
		return nil, err
	}

	byStatus := make(map[string][]TandemoniumEpic)
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
func BlockedEpics(root string) ([]TandemoniumEpic, error) {
	epics, err := TandemoniumEpics(root)
	if err != nil {
		return nil, err
	}

	var blocked []TandemoniumEpic
	for _, e := range epics {
		if e.Status == "blocked" {
			blocked = append(blocked, e)
		}
	}
	return blocked, nil
}

// EpicsWithoutAcceptanceCriteria returns epics missing acceptance criteria.
func EpicsWithoutAcceptanceCriteria(root string) ([]TandemoniumEpic, error) {
	epics, err := TandemoniumEpics(root)
	if err != nil {
		return nil, err
	}

	var missing []TandemoniumEpic
	for _, e := range epics {
		if len(e.AcceptanceCriteria) == 0 {
			missing = append(missing, e)
		}
	}
	return missing, nil
}
