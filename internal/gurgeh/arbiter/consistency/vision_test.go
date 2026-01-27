package consistency

import "testing"

func TestCheckVisionAlignment_NilVision(t *testing.T) {
	sections := map[int]*SectionInfo{
		1: {Content: "Some problem", Accepted: true},
	}
	conflicts := CheckVisionAlignment(nil, sections)
	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts with nil vision, got %d", len(conflicts))
	}
}

func TestCheckVisionAlignment_ProblemMatchesGoals(t *testing.T) {
	vision := &VisionInfo{
		Goals: []string{"Enable AI-assisted product development from research through execution"},
	}
	sections := map[int]*SectionInfo{
		1: {Content: "Product development is fragmented — research and execution are disconnected", Accepted: true},
	}

	conflicts := CheckVisionAlignment(vision, sections)
	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts when problem matches goals, got %d: %v", len(conflicts), conflicts)
	}
}

func TestCheckVisionAlignment_ProblemNoOverlap(t *testing.T) {
	vision := &VisionInfo{
		Goals: []string{"Optimize supply chain logistics for perishable goods"},
	}
	sections := map[int]*SectionInfo{
		1: {Content: "Users need better color themes for their terminal emulator", Accepted: true},
	}

	conflicts := CheckVisionAlignment(vision, sections)
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict for non-overlapping problem, got %d", len(conflicts))
	}
	if conflicts[0].Severity != 1 {
		t.Error("vision alignment conflicts should be warnings, not blockers")
	}
	if conflicts[0].TypeCode != ConflictTypeVisionAlignment {
		t.Errorf("expected TypeCode %d, got %d", ConflictTypeVisionAlignment, conflicts[0].TypeCode)
	}
}

func TestCheckVisionAlignment_FeatureContradictsAssumption(t *testing.T) {
	vision := &VisionInfo{
		Assumptions: []string{"Bigend is a read-only observer that never writes data"},
	}
	sections := map[int]*SectionInfo{
		3: {Content: "Bigend will not remain read-only — it needs write capabilities for dashboard configuration", Accepted: true},
	}

	conflicts := CheckVisionAlignment(vision, sections)
	if len(conflicts) == 0 {
		t.Error("expected conflict when feature contradicts a strategic bet")
	}
}

func TestCheckVisionAlignment_FeatureAlignedWithAssumption(t *testing.T) {
	vision := &VisionInfo{
		Assumptions: []string{"Each tool works standalone with optional Intermute coordination"},
	}
	sections := map[int]*SectionInfo{
		3: {Content: "The feature degrades gracefully when Intermute coordination is unavailable", Accepted: true},
	}

	conflicts := CheckVisionAlignment(vision, sections)
	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts when features align, got %d", len(conflicts))
	}
}

func TestCheckVisionAlignment_UnacceptedSectionsSkipped(t *testing.T) {
	vision := &VisionInfo{
		Goals: []string{"Something completely unrelated to anything"},
	}
	sections := map[int]*SectionInfo{
		1: {Content: "Terminal color themes", Accepted: false}, // not accepted = skip
	}

	conflicts := CheckVisionAlignment(vision, sections)
	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts for unaccepted sections, got %d", len(conflicts))
	}
}
