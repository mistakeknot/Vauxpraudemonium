package tui

import (
	"strings"
	"testing"

	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
)

func TestListScreenRendersSelection(t *testing.T) {
	state := NewSharedState()
	state.Summaries = []specs.Summary{{ID: "PRD-001", Title: "Alpha"}}
	out := (&ListScreen{}).View(state)
	if !strings.Contains(out, "PRD-001") {
		t.Fatalf("expected list item")
	}
}
