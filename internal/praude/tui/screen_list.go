package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type ListScreen struct{}

func (s *ListScreen) Update(msg tea.Msg, state *SharedState) (Screen, Intent) {
	return s, Intent{}
}

func (s *ListScreen) View(state *SharedState) string {
	return joinLines(renderList(state))
}

func (s *ListScreen) Title() string {
	return "LIST"
}

func renderList(state *SharedState) []string {
	lines := []string{"PRDs"}
	if state == nil {
		return lines
	}
	items := filterSummaries(state.Summaries, state.Filter)
	if len(items) == 0 {
		return append(lines, "No PRDs yet.")
	}
	for i, s := range items {
		prefix := "  "
		if i == state.Selected {
			prefix = "> "
		}
		lines = append(lines, prefix+s.ID+" "+s.Title)
	}
	return lines
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

func renderGroupList(items []Item, cursor, viewOffset, height int) string {
	if len(items) == 0 {
		return "No PRDs yet."
	}
	if height < 1 {
		height = 1
	}
	var b strings.Builder
	maxVisible := height
	if viewOffset > 0 {
		above := viewOffset
		b.WriteString(fmt.Sprintf("  ⋮ +%d above", above))
		b.WriteString("\n")
		maxVisible--
	}
	visibleCount := 0
	for i := viewOffset; i < len(items) && visibleCount < maxVisible; i++ {
		item := items[i]
		selected := i == cursor
		b.WriteString(renderGroupListItem(item, selected))
		visibleCount++
		if visibleCount < maxVisible && i < len(items)-1 {
			b.WriteString("\n")
		}
	}
	remaining := len(items) - (viewOffset + visibleCount)
	if remaining > 0 {
		if visibleCount > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("  ⋮ +%d below", remaining))
	}
	return b.String()
}

func renderGroupListItem(item Item, selected bool) string {
	prefix := " "
	if selected {
		prefix = "▶"
	}
	if item.Type == ItemTypeGroup && item.Group != nil {
		expand := "▸"
		if item.Group.Expanded {
			expand = "▾"
		}
		name := strings.ToUpper(item.Group.Name)
		count := len(item.Group.Items)
		return fmt.Sprintf("%s %s %s (%d)", prefix, expand, name, count)
	}
	if item.Summary == nil {
		return prefix + " ..."
	}
	connector := "├─"
	if item.IsLastInGroup {
		connector = "└─"
	}
	return fmt.Sprintf("%s %s %s %s", prefix, connector, item.Summary.ID, item.Summary.Title)
}
