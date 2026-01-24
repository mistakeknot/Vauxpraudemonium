package plan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPlan(t *testing.T) {
	p := NewPlan("praude", "interview")

	if p.Tool != "praude" {
		t.Errorf("expected tool=praude, got %s", p.Tool)
	}
	if p.Action != "interview" {
		t.Errorf("expected action=interview, got %s", p.Action)
	}
	if p.Version != Version {
		t.Errorf("expected version=%s, got %s", Version, p.Version)
	}
	if !p.Ready {
		t.Error("expected Ready=true by default")
	}
}

func TestPlanRecommendations(t *testing.T) {
	p := NewPlan("praude", "interview")

	// Add info recommendation - should stay ready
	p.AddRecommendation(Recommendation{
		Type:     TypeIntegration,
		Severity: SeverityInfo,
		Message:  "Test info",
	})
	if !p.Ready {
		t.Error("info recommendation should not affect Ready")
	}

	// Add warning recommendation - should stay ready
	p.AddRecommendation(Recommendation{
		Type:     TypeValidation,
		Severity: SeverityWarning,
		Message:  "Test warning",
	})
	if !p.Ready {
		t.Error("warning recommendation should not affect Ready")
	}
	if !p.HasWarnings() {
		t.Error("expected HasWarnings=true")
	}

	// Add error recommendation - should become not ready
	p.AddRecommendation(Recommendation{
		Type:     TypePrereq,
		Severity: SeverityError,
		Message:  "Test error",
	})
	if p.Ready {
		t.Error("error recommendation should set Ready=false")
	}
	if !p.HasErrors() {
		t.Error("expected HasErrors=true")
	}
}

func TestPlanSetGetItems(t *testing.T) {
	p := NewPlan("pollard", "scan")

	type TestItems struct {
		Hunters []string `json:"hunters"`
		Count   int      `json:"count"`
	}

	items := TestItems{
		Hunters: []string{"github", "arxiv"},
		Count:   2,
	}

	if err := p.SetItems(items); err != nil {
		t.Fatalf("SetItems failed: %v", err)
	}

	var result TestItems
	if err := p.GetItems(&result); err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}

	if len(result.Hunters) != 2 {
		t.Errorf("expected 2 hunters, got %d", len(result.Hunters))
	}
	if result.Count != 2 {
		t.Errorf("expected count=2, got %d", result.Count)
	}
}

func TestPlanSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()

	p := NewPlan("praude", "interview")
	p.Summary = "Test plan"
	p.AddRecommendation(Recommendation{
		Type:     TypeValidation,
		Severity: SeverityWarning,
		Message:  "Test warning",
	})

	path, err := p.Save(tmpDir)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, ".praude", "pending", "interview-plan.json")
	if path != expectedPath {
		t.Errorf("expected path=%s, got %s", expectedPath, path)
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("plan file not created: %v", err)
	}

	// Load it back
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Tool != "praude" {
		t.Errorf("loaded tool mismatch: %s", loaded.Tool)
	}
	if loaded.Summary != "Test plan" {
		t.Errorf("loaded summary mismatch: %s", loaded.Summary)
	}
	if len(loaded.Recommendations) != 1 {
		t.Errorf("expected 1 recommendation, got %d", len(loaded.Recommendations))
	}
}

func TestLoadPending(t *testing.T) {
	tmpDir := t.TempDir()

	p := NewPlan("pollard", "scan")
	p.Summary = "Pending scan"

	_, err := p.Save(tmpDir)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadPending(tmpDir, "pollard", "scan")
	if err != nil {
		t.Fatalf("LoadPending failed: %v", err)
	}

	if loaded.Summary != "Pending scan" {
		t.Errorf("loaded summary mismatch: %s", loaded.Summary)
	}
}

func TestClearPending(t *testing.T) {
	tmpDir := t.TempDir()

	p := NewPlan("praude", "interview")
	path, err := p.Save(tmpDir)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("plan file not created: %v", err)
	}

	// Clear it
	if err := ClearPending(tmpDir, "praude", "interview"); err != nil {
		t.Fatalf("ClearPending failed: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("plan file should be deleted")
	}

	// Clear again should not error
	if err := ClearPending(tmpDir, "praude", "interview"); err != nil {
		t.Fatalf("ClearPending on missing file should not error: %v", err)
	}
}

func TestFilterBySeverity(t *testing.T) {
	p := NewPlan("test", "test")
	p.AddRecommendation(Recommendation{Severity: SeverityInfo, Message: "info1"})
	p.AddRecommendation(Recommendation{Severity: SeverityWarning, Message: "warn1"})
	p.AddRecommendation(Recommendation{Severity: SeverityWarning, Message: "warn2"})
	p.AddRecommendation(Recommendation{Severity: SeverityError, Message: "err1"})

	infos := p.FilterBySeverity(SeverityInfo)
	if len(infos) != 1 {
		t.Errorf("expected 1 info, got %d", len(infos))
	}

	warnings := p.FilterBySeverity(SeverityWarning)
	if len(warnings) != 2 {
		t.Errorf("expected 2 warnings, got %d", len(warnings))
	}

	errors := p.FilterBySeverity(SeverityError)
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
}

func TestFilterByType(t *testing.T) {
	p := NewPlan("test", "test")
	p.AddRecommendation(Recommendation{Type: TypeIntegration, Message: "int1"})
	p.AddRecommendation(Recommendation{Type: TypeValidation, Message: "val1"})
	p.AddRecommendation(Recommendation{Type: TypeIntegration, Message: "int2"})

	integrations := p.FilterByType(TypeIntegration)
	if len(integrations) != 2 {
		t.Errorf("expected 2 integrations, got %d", len(integrations))
	}

	validations := p.FilterByType(TypeValidation)
	if len(validations) != 1 {
		t.Errorf("expected 1 validation, got %d", len(validations))
	}
}
