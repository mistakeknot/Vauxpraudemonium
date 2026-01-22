package tui

import (
	"testing"

	"github.com/mistakeknot/vauxpraudemonium/internal/praude/specs"
)

func TestGroupTreeFlatten(t *testing.T) {
	summaries := []specs.Summary{
		{ID: "PRD-1", Title: "A", Status: "draft"},
		{ID: "PRD-2", Title: "B", Status: "research"},
	}
	tree := NewGroupTree(summaries, map[string]bool{"draft": true, "research": true})
	items := tree.Flatten()
	if len(items) < 4 {
		t.Fatalf("expected headers and items")
	}
	if items[0].Type != ItemTypeGroup || items[1].Type != ItemTypePRD {
		t.Fatalf("expected group then item")
	}
}
