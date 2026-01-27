package arbiter

import (
	"context"
	"strings"
	"testing"
)

func TestGenerateDraftFromContext(t *testing.T) {
	gen := NewGenerator()
	ctx := context.Background()
	projectCtx := &ProjectContext{
		HasReadme:      true,
		ReadmeSnippet:  "A CLI tool for managing reading lists",
		HasPackageJSON: false,
	}
	draft, err := gen.GenerateDraft(ctx, PhaseProblem, projectCtx, "")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if draft.Content == "" {
		t.Error("expected non-empty draft content")
	}
	if len(draft.Options) < 2 {
		t.Errorf("expected at least 2 options, got %d", len(draft.Options))
	}
	if draft.Status != DraftProposed {
		t.Errorf("expected DraftProposed status, got %d", draft.Status)
	}
	if !strings.Contains(draft.Content, "reading lists") {
		t.Error("expected draft to reference project context")
	}
}

func TestGenerateDraftFromUserInput(t *testing.T) {
	gen := NewGenerator()
	ctx := context.Background()
	draft, err := gen.GenerateDraft(ctx, PhaseProblem, nil, "I want to build a habit tracker for developers")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if draft.Content == "" {
		t.Error("expected non-empty draft content")
	}
	if draft.Status != DraftProposed {
		t.Errorf("expected DraftProposed status, got %d", draft.Status)
	}
}

func TestGenerateDraftFromInputWithBecause(t *testing.T) {
	gen := NewGenerator()
	ctx := context.Background()
	draft, err := gen.GenerateDraft(ctx, PhaseProblem, nil, "I want to build a task manager because existing tools are too complex")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if !strings.Contains(draft.Content, "existing tools are too complex") {
		t.Error("expected draft to extract reason after 'because'")
	}
}

func TestGenerateDraftFallback(t *testing.T) {
	gen := NewGenerator()
	ctx := context.Background()
	draft, err := gen.GenerateDraft(ctx, PhaseProblem, nil, "")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if draft.Content == "" {
		t.Error("expected non-empty fallback draft")
	}
}

func TestGenerateAllPhases(t *testing.T) {
	gen := NewGenerator()
	ctx := context.Background()
	for _, phase := range AllPhases() {
		draft, err := gen.GenerateDraft(ctx, phase, nil, "test input")
		if err != nil {
			t.Fatalf("phase %s: generate failed: %v", phase, err)
		}
		if draft.Content == "" {
			t.Errorf("phase %s: expected non-empty content", phase)
		}
		if len(draft.Options) < 2 {
			t.Errorf("phase %s: expected at least 2 options, got %d", phase, len(draft.Options))
		}
		if draft.Status != DraftProposed {
			t.Errorf("phase %s: expected DraftProposed", phase)
		}
	}
}

func TestGenerateUnknownPhase(t *testing.T) {
	gen := NewGenerator()
	ctx := context.Background()
	_, err := gen.GenerateDraft(ctx, Phase(99), nil, "test")
	if err == nil {
		t.Error("expected error for unknown phase")
	}
}

func TestGenerateDraftFallbackNoContext(t *testing.T) {
	gen := NewGenerator()
	ctx := context.Background()
	draft, err := gen.GenerateDraft(ctx, PhaseProblem, nil, "")
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if draft.Content == "" {
		t.Error("expected fallback content")
	}
	if len(draft.Options) < 2 {
		t.Errorf("expected options, got %d", len(draft.Options))
	}
}
