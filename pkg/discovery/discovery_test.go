package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestToolDir(t *testing.T) {
	dir := ToolDir("/project", "praude")
	expected := "/project/.praude"
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestToolExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Tool doesn't exist
	if ToolExists(tmpDir, "praude") {
		t.Error("expected praude to not exist")
	}

	// Create the tool directory
	if err := os.MkdirAll(filepath.Join(tmpDir, ".praude"), 0755); err != nil {
		t.Fatal(err)
	}

	// Now it should exist
	if !ToolExists(tmpDir, "praude") {
		t.Error("expected praude to exist")
	}
}

func TestFindProjectRoot(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a nested directory structure
	subDir := filepath.Join(tmpDir, "sub", "dir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// No tool directories - should return start path
	root, err := FindProjectRoot(subDir)
	if err != nil {
		t.Fatal(err)
	}
	if root != subDir {
		t.Errorf("expected %s (fallback), got %s", subDir, root)
	}

	// Create .pollard in tmpDir
	if err := os.MkdirAll(filepath.Join(tmpDir, ".pollard"), 0755); err != nil {
		t.Fatal(err)
	}

	// Now should find tmpDir
	root, err = FindProjectRoot(subDir)
	if err != nil {
		t.Fatal(err)
	}
	if root != tmpDir {
		t.Errorf("expected %s, got %s", tmpDir, root)
	}
}

func TestPollardInsights(t *testing.T) {
	tmpDir := t.TempDir()

	// Empty - should return empty slice
	insights, err := PollardInsights(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(insights) != 0 {
		t.Errorf("expected 0 insights, got %d", len(insights))
	}

	// Create insights directory with a file
	insightsDir := filepath.Join(tmpDir, ".pollard", "insights")
	if err := os.MkdirAll(insightsDir, 0755); err != nil {
		t.Fatal(err)
	}

	insightYAML := `id: INS-001
title: Test Insight
category: competitive
findings:
  - title: Finding 1
    relevance: high
    description: Test finding
`
	if err := os.WriteFile(filepath.Join(insightsDir, "ins-001.yaml"), []byte(insightYAML), 0644); err != nil {
		t.Fatal(err)
	}

	insights, err = PollardInsights(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(insights) != 1 {
		t.Errorf("expected 1 insight, got %d", len(insights))
	}
	if insights[0].ID != "INS-001" {
		t.Errorf("expected ID=INS-001, got %s", insights[0].ID)
	}
}

func TestPraudeSpecs(t *testing.T) {
	tmpDir := t.TempDir()

	// Empty - should return empty slice
	specs, err := PraudeSpecs(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 0 {
		t.Errorf("expected 0 specs, got %d", len(specs))
	}

	// Create specs directory with a file
	specsDir := filepath.Join(tmpDir, ".praude", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}

	specYAML := `id: PRD-001
title: Test PRD
status: draft
summary: A test PRD
requirements:
  - "REQ-001: First requirement"
`
	if err := os.WriteFile(filepath.Join(specsDir, "prd-001.yaml"), []byte(specYAML), 0644); err != nil {
		t.Fatal(err)
	}

	specs, err = PraudeSpecs(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 1 {
		t.Errorf("expected 1 spec, got %d", len(specs))
	}
	if specs[0].ID != "PRD-001" {
		t.Errorf("expected ID=PRD-001, got %s", specs[0].ID)
	}
}

func TestTandemoniumEpics(t *testing.T) {
	tmpDir := t.TempDir()

	// Empty - should return empty slice
	epics, err := TandemoniumEpics(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(epics) != 0 {
		t.Errorf("expected 0 epics, got %d", len(epics))
	}

	// Create specs directory with a file
	specsDir := filepath.Join(tmpDir, ".tandemonium", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}

	epicYAML := `id: EPIC-001
title: Test Epic
status: todo
priority: p1
stories:
  - id: STORY-001
    title: Test Story
    status: todo
`
	if err := os.WriteFile(filepath.Join(specsDir, "epic-001.yaml"), []byte(epicYAML), 0644); err != nil {
		t.Fatal(err)
	}

	epics, err = TandemoniumEpics(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(epics) != 1 {
		t.Errorf("expected 1 epic, got %d", len(epics))
	}
	if epics[0].ID != "EPIC-001" {
		t.Errorf("expected ID=EPIC-001, got %s", epics[0].ID)
	}
	if len(epics[0].Stories) != 1 {
		t.Errorf("expected 1 story, got %d", len(epics[0].Stories))
	}
}

func TestCountFunctions(t *testing.T) {
	tmpDir := t.TempDir()

	// All counts should be 0 for empty directory
	if count := CountPollardInsights(tmpDir); count != 0 {
		t.Errorf("expected 0 insights, got %d", count)
	}
	if count := CountPraudeSpecs(tmpDir); count != 0 {
		t.Errorf("expected 0 specs, got %d", count)
	}
	if count := CountTandemoniumEpics(tmpDir); count != 0 {
		t.Errorf("expected 0 epics, got %d", count)
	}

	// HasData should be false
	if PollardHasData(tmpDir) {
		t.Error("expected PollardHasData=false")
	}
	if PraudeHasData(tmpDir) {
		t.Error("expected PraudeHasData=false")
	}
	if TandemoniumHasData(tmpDir) {
		t.Error("expected TandemoniumHasData=false")
	}
}
