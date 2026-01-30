package thinking

import (
	"strings"
	"testing"
)

func TestRenderPreamble_AllPhases(t *testing.T) {
	for phase, shape := range PhaseDefault {
		result := RenderPreamble(shape, phase)
		if result == "" {
			t.Errorf("RenderPreamble(%s, %q) returned empty", shape, phase)
		}
	}
}

func TestRenderPreamble_ShapeNone(t *testing.T) {
	for phase := range PhaseDefault {
		result := RenderPreamble(ShapeNone, phase)
		if result != "" {
			t.Errorf("ShapeNone for %q should return empty, got %q", phase, result)
		}
	}
}

func TestRenderPreamble_UnknownPhase(t *testing.T) {
	result := RenderPreamble(ShapeDeductive, "Nonexistent")
	if result != "" {
		t.Errorf("unknown phase should return empty, got %q", result)
	}
}

func TestRenderForPhase(t *testing.T) {
	tests := []struct {
		phase    string
		wantNon  bool
		contains string
	}{
		{"Vision", true, "criteria"},
		{"Problem", true, "fail"},
		{"Users", true, "persona"},
		{"Features + Goals", true, "schema"},
		{"Requirements", true, "given"},
		{"Scope + Assumptions", true, "scope creep"},
		{"Critical User Journeys", true, "principles"},
		{"Acceptance Criteria", true, "testable"},
		{"Nonexistent", false, ""},
	}

	for _, tt := range tests {
		result := RenderForPhase(tt.phase)
		if tt.wantNon && result == "" {
			t.Errorf("RenderForPhase(%q) returned empty", tt.phase)
		}
		if !tt.wantNon && result != "" {
			t.Errorf("RenderForPhase(%q) should be empty, got %q", tt.phase, result)
		}
		if tt.contains != "" && !strings.Contains(strings.ToLower(result), tt.contains) {
			t.Errorf("RenderForPhase(%q) should contain %q", tt.phase, tt.contains)
		}
	}
}

func TestFormatPreamble(t *testing.T) {
	// Non-empty
	result := FormatPreamble("hello")
	if !strings.Contains(result, "Thinking Preamble") {
		t.Error("FormatPreamble should include delimiter")
	}
	if !strings.Contains(result, "hello") {
		t.Error("FormatPreamble should include content")
	}

	// Empty
	if FormatPreamble("") != "" {
		t.Error("FormatPreamble of empty should be empty")
	}
	if FormatPreamble("   ") != "" {
		t.Error("FormatPreamble of whitespace should be empty")
	}
}

func TestShapeString(t *testing.T) {
	if ShapeDeductive.String() != "Deductive" {
		t.Errorf("got %q", ShapeDeductive.String())
	}
	if Shape(99).String() != "Unknown" {
		t.Errorf("out of range should be Unknown")
	}
}
