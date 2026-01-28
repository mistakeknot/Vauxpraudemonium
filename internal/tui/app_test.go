package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

type noopView struct{}

func (v *noopView) Init() tea.Cmd                             { return nil }
func (v *noopView) Update(msg tea.Msg) (pkgtui.View, tea.Cmd) { return v, nil }
func (v *noopView) View() string                              { return "content" }
func (v *noopView) Focus() tea.Cmd                            { return nil }
func (v *noopView) Blur()                                     {}
func (v *noopView) Name() string                              { return "Test" }
func (v *noopView) ShortHelp() string                         { return "Tab focus" }

func TestAppHelpOverlayToggles(t *testing.T) {
	app := NewApp(nil, &noopView{})
	_, _ = app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	updated, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	app = updated.(*App)
	if cmd == nil {
		t.Fatalf("expected help command")
	}
	updated, _ = app.Update(cmd())
	app = updated.(*App)

	if !strings.Contains(app.View(), "Keyboard Shortcuts") {
		t.Fatalf("expected help overlay")
	}
}

func TestAppCtrlCQuitsWithPaletteVisible(t *testing.T) {
	app := NewApp(nil, &noopView{})
	app.palette.Show()

	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatalf("expected quit command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg")
	}
}
