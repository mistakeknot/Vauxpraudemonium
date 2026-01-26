package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

// TabBar renders a horizontal tab bar
type TabBar struct {
	tabs   []string
	active int
	width  int
}

// NewTabBar creates a new tab bar
func NewTabBar(tabs []string) *TabBar {
	return &TabBar{
		tabs: tabs,
	}
}

// SetActive sets the active tab
func (t *TabBar) SetActive(index int) {
	if index >= 0 && index < len(t.tabs) {
		t.active = index
	}
}

// Active returns the active tab index
func (t *TabBar) Active() int {
	return t.active
}

// SetWidth sets the tab bar width
func (t *TabBar) SetWidth(width int) {
	t.width = width
}

// Next moves to the next tab
func (t *TabBar) Next() {
	t.active = (t.active + 1) % len(t.tabs)
}

// Prev moves to the previous tab
func (t *TabBar) Prev() {
	t.active = (t.active - 1 + len(t.tabs)) % len(t.tabs)
}

// View renders the tab bar
func (t *TabBar) View() string {
	var tabs []string

	for i, name := range t.tabs {
		// Add number prefix
		numStyle := lipgloss.NewStyle().
			Foreground(pkgtui.ColorMuted)
		num := numStyle.Render(string('1' + rune(i)))

		if i == t.active {
			tabStyle := lipgloss.NewStyle().
				Background(pkgtui.ColorPrimary).
				Foreground(pkgtui.ColorBg).
				Bold(true).
				Padding(0, 2)
			tabs = append(tabs, tabStyle.Render(num+" "+name))
		} else {
			tabStyle := lipgloss.NewStyle().
				Foreground(pkgtui.ColorFgDim).
				Padding(0, 2)
			tabs = append(tabs, tabStyle.Render(num+" "+name))
		}
	}

	row := strings.Join(tabs, "")
	return row
}

// TabNames returns the list of tab names
func (t *TabBar) TabNames() []string {
	return t.tabs
}
