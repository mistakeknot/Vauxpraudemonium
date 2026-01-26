package sync

import (
	"testing"

	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
	"github.com/mistakeknot/autarch/pkg/intermute"
)

func TestToIntermuteSpec(t *testing.T) {
	spec := specs.Spec{
		ID:        "PRD-001",
		Title:     "Test PRD",
		CreatedAt: "2026-01-20T10:00:00Z",
		Status:    "draft",
		Summary:   "A test vision",
		UserStory: specs.UserStory{Text: "User problem statement"},
	}

	iSpec := toIntermuteSpec(spec, "test-project")

	if iSpec.ID != "PRD-001" {
		t.Errorf("ID = %v, want %v", iSpec.ID, "PRD-001")
	}
	if iSpec.Project != "test-project" {
		t.Errorf("Project = %v, want %v", iSpec.Project, "test-project")
	}
	if iSpec.Title != "Test PRD" {
		t.Errorf("Title = %v, want %v", iSpec.Title, "Test PRD")
	}
	if iSpec.Vision != "A test vision" {
		t.Errorf("Vision = %v, want %v", iSpec.Vision, "A test vision")
	}
	if iSpec.Problem != "User problem statement" {
		t.Errorf("Problem = %v, want %v", iSpec.Problem, "User problem statement")
	}
	if iSpec.Status != intermute.SpecStatusDraft {
		t.Errorf("Status = %v, want %v", iSpec.Status, intermute.SpecStatusDraft)
	}
}

func TestFromIntermuteSpec(t *testing.T) {
	iSpec := intermute.Spec{
		ID:      "PRD-002",
		Project: "test-project",
		Title:   "Imported PRD",
		Vision:  "Vision text",
		Problem: "Problem statement",
		Status:  intermute.SpecStatusValidated,
	}

	spec := fromIntermuteSpec(iSpec)

	if spec.ID != "PRD-002" {
		t.Errorf("ID = %v, want %v", spec.ID, "PRD-002")
	}
	if spec.Title != "Imported PRD" {
		t.Errorf("Title = %v, want %v", spec.Title, "Imported PRD")
	}
	if spec.Summary != "Vision text" {
		t.Errorf("Summary = %v, want %v", spec.Summary, "Vision text")
	}
	if spec.UserStory.Text != "Problem statement" {
		t.Errorf("UserStory.Text = %v, want %v", spec.UserStory.Text, "Problem statement")
	}
	if spec.Status != "validated" {
		t.Errorf("Status = %v, want %v", spec.Status, "validated")
	}
}

func TestMapSpecStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected intermute.SpecStatus
	}{
		{"draft", intermute.SpecStatusDraft},
		{"Draft", intermute.SpecStatusDraft},
		{"DRAFT", intermute.SpecStatusDraft},
		{"research", intermute.SpecStatusResearch},
		{"validated", intermute.SpecStatusValidated},
		{"approved", intermute.SpecStatusValidated},
		{"archived", intermute.SpecStatusArchived},
		{"done", intermute.SpecStatusArchived},
		{"unknown", intermute.SpecStatusDraft},
		{"", intermute.SpecStatusDraft},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapSpecStatus(tt.input)
			if result != tt.expected {
				t.Errorf("mapSpecStatus(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewSyncer(t *testing.T) {
	syncer := NewSyncer(nil, "test-project")
	if syncer == nil {
		t.Error("NewSyncer returned nil")
	}
	if syncer.project != "test-project" {
		t.Errorf("project = %v, want %v", syncer.project, "test-project")
	}
}
