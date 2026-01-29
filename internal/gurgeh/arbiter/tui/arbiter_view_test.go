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

func TestArbiterSidebarUsesInterviewSteps(t *testing.T) {
	view := NewArbiterView("/tmp/test", nil)
	items := view.SidebarItems()

	if len(items) != 8 {
		t.Fatalf("expected 8 sidebar items, got %d", len(items))
	}
	if items[0].Label != "Vision" ||
		items[1].Label != "Problem" ||
		items[2].Label != "Users" ||
		items[3].Label != "Features + Goals" ||
		items[4].Label != "Requirements" ||
		items[5].Label != "Scope + Assumptions" ||
		items[6].Label != "Critical User Journeys" ||
		items[7].Label != "Acceptance Criteria" {
		t.Fatalf("unexpected sidebar labels: %#v", items)
	}
}
