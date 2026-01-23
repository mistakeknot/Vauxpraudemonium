package epics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteEpicsCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	epics := []Epic{
		{
			ID:       "EPIC-001",
			Title:    "Auth",
			Summary:  "User auth and sessions",
			Status:   StatusTodo,
			Priority: PriorityP1,
			AcceptanceCriteria: []string{
				"Login works",
			},
			Risks: []string{
				"OAuth latency",
			},
			Estimate: "M",
			Stories: []Story{
				{
					ID:       "EPIC-001-S01",
					Title:    "Login form",
					Summary:  "Email/password flow",
					Status:   StatusTodo,
					Priority: PriorityP1,
					Estimate: "S",
				},
			},
		},
	}

	if err := WriteEpics(dir, epics, WriteOptions{Existing: ExistingOverwrite}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "EPIC-001.yaml")); err != nil {
		t.Fatalf("expected epic file: %v", err)
	}
}
