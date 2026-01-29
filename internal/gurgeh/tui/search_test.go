package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
)

func TestSearchFiltersList(t *testing.T) {
	state := NewSharedState()
	state.Summaries = []specs.Summary{
		{ID: "PRD-001", Title: "Alpha"},
		{ID: "PRD-002", Title: "Beta"},
	}
	state.Filter = "Alpha"
	items := filterSummaries(state.Summaries, state.Filter)
	if len(items) != 1 {
		t.Fatalf("expected filtered list")
	}
}

func TestSearchModalConsumesKeys(t *testing.T) {
	withTempRoot(t, func(root string) {
		m := NewModel()
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
		m = updated.(Model)
		if m.searchOverlay == nil || !m.searchOverlay.Visible() {
			t.Fatalf("expected search overlay visible")
		}
	})
}
