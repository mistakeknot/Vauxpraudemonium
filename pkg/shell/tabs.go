package shell

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/mistakeknot/autarch/pkg/tui"
)

// ToolTab represents a top-level tool tab
type ToolTab int

const (
	TabBigend ToolTab = iota
	TabPollard
	TabGurgeh
	TabColdwine
)

func (t ToolTab) String() string {
	switch t {
	case TabBigend:
		return "Bigend"
	case TabPollard:
		return "Pollard"
	case TabGurgeh:
		return "Gurgeh"
	case TabColdwine:
		return "Coldwine"
	default:
		return "Unknown"
	}
}

// Key returns the keyboard shortcut for the tab
func (t ToolTab) Key() string {
	switch t {
	case TabBigend:
		return "1"
	case TabPollard:
		return "2"
	case TabGurgeh:
		return "3"
	case TabColdwine:
		return "4"
	default:
		return ""
	}
}

// TabCount returns the number of tool tabs
const TabCount = 4

// RenderTabBar renders the tool tab bar
func RenderTabBar(active ToolTab) string {
	tabs := make([]string, TabCount)
	for i := 0; i < TabCount; i++ {
		tab := ToolTab(i)
		style := tui.TabStyle
		if tab == active {
			style = tui.ActiveTabStyle
		}
		tabs[i] = style.Render(fmt.Sprintf("%s %s", tab.Key(), tab.String()))
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, tabs...)
}

// RenderSubTabBar renders a tool's internal sub-tabs
func RenderSubTabBar(tabs []string, active int) string {
	if len(tabs) == 0 {
		return ""
	}

	rendered := make([]string, len(tabs))
	for i, tab := range tabs {
		style := tui.TabStyle
		if i == active {
			style = tui.ActiveTabStyle
		}
		rendered[i] = style.Render(tab)
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, rendered...)
}
