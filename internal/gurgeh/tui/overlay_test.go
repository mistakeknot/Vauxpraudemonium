package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHelpOverlayToggle(t *testing.T) {
	withTempRoot(t, func(root string) {
		m := NewModel()
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyF1})
		m = updated.(Model)
		if !strings.Contains(stripANSI(m.View()), "Keyboard Shortcuts") {
			t.Fatalf("expected help overlay")
		}
	})
}

func TestTutorialOverlayToggle(t *testing.T) {
	withTempRoot(t, func(root string) {
		m := NewModel()
		m = pressKey(m, "`")
		if !strings.Contains(stripANSI(m.View()), "Tutorial") {
			t.Fatalf("expected tutorial overlay")
		}
	})
}
