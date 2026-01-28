package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mistakeknot/autarch/internal/coldwine/tasks"
)

func TestTaskDetailBackUsesCommonBack(t *testing.T) {
	view := NewTaskDetailView(tasks.TaskProposal{}, nil)
	called := false
	view.SetCallbacks(nil, func() tea.Cmd {
		called = true
		return nil
	})

	_, _ = view.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !called {
		t.Fatalf("expected back handler on common back key")
	}
}
