package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mistakeknot/autarch/internal/gurgeh/arbiter"
)

func TestArbiterViewCtrlCQuits(t *testing.T) {
	view := NewArbiterView("/tmp/test", nil)
	view.state = arbiter.NewSprintState("/tmp/test")

	_, cmd := view.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatalf("expected quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg")
	}
}

func TestArbiterViewQDoesNotQuit(t *testing.T) {
	view := NewArbiterView("/tmp/test", nil)
	view.state = arbiter.NewSprintState("/tmp/test")

	_, cmd := view.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatalf("did not expect quit command")
		}
	}
}
