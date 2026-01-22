package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFilterClearsOnEscape(t *testing.T) {
	m := New(&fakeAggLayout{}, "")
	m = m.withFilterActive("codex")
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm := updated.(Model)
	if mm.filterInput.Value() != "" {
		t.Fatalf("expected empty filter")
	}
	if mm.filterActive {
		t.Fatalf("expected filter inactive")
	}
}
