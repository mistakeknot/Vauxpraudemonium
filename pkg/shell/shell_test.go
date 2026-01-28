package shell

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestShellCtrlCQuits(t *testing.T) {
	m := New(nil, "")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(Model)
	if cmd == nil {
		t.Fatalf("expected quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg")
	}
	if !m.quitting {
		t.Fatalf("expected quitting state")
	}
}

func TestShellQDoesNotQuit(t *testing.T) {
	m := New(nil, "")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(Model)
	if cmd != nil {
		if _, ok := cmd().(tea.QuitMsg); ok {
			t.Fatalf("did not expect quit command")
		}
	}
	if m.quitting {
		t.Fatalf("did not expect quitting state")
	}
}
