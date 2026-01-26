package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakeknot/autarch/internal/coldwine/tasks"
	"github.com/mistakeknot/autarch/internal/tui"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

// TaskReviewView displays proposed tasks for review and editing.
type TaskReviewView struct {
	tasks       []tasks.TaskProposal
	groupedView bool // Group by epic
	selected    int
	expanded    map[int]bool
	width       int
	height      int
	scrollOffset int

	// Callbacks
	onAccept func(tasks []tasks.TaskProposal) tea.Cmd
}

// NewTaskReviewView creates a new task review view.
func NewTaskReviewView(taskList []tasks.TaskProposal) *TaskReviewView {
	// Resolve cross-epic dependencies
	tasks.ResolveCrossEpicDependencies(taskList)

	return &TaskReviewView{
		tasks:       taskList,
		groupedView: true,
		expanded:    make(map[int]bool),
	}
}

// SetAcceptCallback sets the callback for when tasks are accepted.
func (v *TaskReviewView) SetAcceptCallback(cb func([]tasks.TaskProposal) tea.Cmd) {
	v.onAccept = cb
}

// Init implements View
func (v *TaskReviewView) Init() tea.Cmd {
	return nil
}

// Update implements View
func (v *TaskReviewView) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height - 4
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Accept all
			if v.onAccept != nil {
				return v, v.onAccept(v.tasks)
			}
			return v, nil

		case "g":
			// Toggle grouped view
			v.groupedView = !v.groupedView
			return v, nil

		case "e":
			// Edit selected
			// TODO: Add inline editing
			return v, nil

		case "d":
			// Delete selected
			if v.selected >= 0 && v.selected < len(v.tasks) {
				v.tasks = append(v.tasks[:v.selected], v.tasks[v.selected+1:]...)
				if v.selected >= len(v.tasks) {
					v.selected = len(v.tasks) - 1
				}
				// Recompute readiness
				tasks.ResolveCrossEpicDependencies(v.tasks)
			}
			return v, nil

		case "j", "down":
			if v.selected < len(v.tasks)-1 {
				v.selected++
				v.ensureVisible()
			}
			return v, nil

		case "k", "up":
			if v.selected > 0 {
				v.selected--
				v.ensureVisible()
			}
			return v, nil

		case "space":
			v.expanded[v.selected] = !v.expanded[v.selected]
			return v, nil

		case "tab":
			// Cycle through types
			if v.selected >= 0 && v.selected < len(v.tasks) {
				v.tasks[v.selected].Edited = true
				switch v.tasks[v.selected].Type {
				case tasks.TaskTypeImplementation:
					v.tasks[v.selected].Type = tasks.TaskTypeTest
				case tasks.TaskTypeTest:
					v.tasks[v.selected].Type = tasks.TaskTypeDocumentation
				case tasks.TaskTypeDocumentation:
					v.tasks[v.selected].Type = tasks.TaskTypeReview
				case tasks.TaskTypeReview:
					v.tasks[v.selected].Type = tasks.TaskTypeResearch
				case tasks.TaskTypeResearch:
					v.tasks[v.selected].Type = tasks.TaskTypeSetup
				default:
					v.tasks[v.selected].Type = tasks.TaskTypeImplementation
				}
			}
			return v, nil
		}
	}

	return v, nil
}

func (v *TaskReviewView) ensureVisible() {
	visibleHeight := v.height - 8
	if v.selected < v.scrollOffset {
		v.scrollOffset = v.selected
	} else if v.selected >= v.scrollOffset+visibleHeight {
		v.scrollOffset = v.selected - visibleHeight + 1
	}
}

// View implements View
func (v *TaskReviewView) View() string {
	if len(v.tasks) == 0 {
		return pkgtui.LabelStyle.Render("No tasks proposed")
	}

	var sections []string

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(pkgtui.ColorPrimary).
		Bold(true).
		MarginBottom(1)

	readyCount := tasks.CountReady(v.tasks)
	header := fmt.Sprintf("Task Review (%d tasks, %d ready)", len(v.tasks), readyCount)
	sections = append(sections, headerStyle.Render(header))

	// View mode indicator
	modeStyle := pkgtui.LabelStyle
	if v.groupedView {
		sections = append(sections, modeStyle.Render("Grouped by Epic (g to toggle)"))
	} else {
		sections = append(sections, modeStyle.Render("Flat list (g to toggle)"))
	}
	sections = append(sections, "")

	// Task list
	if v.groupedView {
		sections = append(sections, v.renderGroupedTasks()...)
	} else {
		sections = append(sections, v.renderFlatTasks()...)
	}

	sections = append(sections, "")

	// Actions
	sections = append(sections, v.renderActions())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (v *TaskReviewView) renderGroupedTasks() []string {
	var lines []string

	groups := tasks.GroupByEpic(v.tasks)

	// Get sorted epic IDs
	epicIDs := make([]string, 0, len(groups))
	for epicID := range groups {
		epicIDs = append(epicIDs, epicID)
	}

	globalIdx := 0
	for _, epicID := range epicIDs {
		epicTasks := groups[epicID]

		// Epic header
		epicStyle := pkgtui.SubtitleStyle.MarginTop(1)
		lines = append(lines, epicStyle.Render(epicID))

		for _, t := range epicTasks {
			isSelected := globalIdx == v.selected
			isExpanded := v.expanded[globalIdx]
			taskLine := v.renderTask(t, isSelected, isExpanded, 2)
			lines = append(lines, taskLine)
			globalIdx++
		}
	}

	return lines
}

func (v *TaskReviewView) renderFlatTasks() []string {
	var lines []string

	for i, t := range v.tasks {
		isSelected := i == v.selected
		isExpanded := v.expanded[i]
		taskLine := v.renderTask(t, isSelected, isExpanded, 0)
		lines = append(lines, taskLine)
	}

	return lines
}

func (v *TaskReviewView) renderTask(t tasks.TaskProposal, selected, expanded bool, indent int) string {
	var lines []string

	indentStr := strings.Repeat(" ", indent)

	// Selector
	var selector string
	if selected {
		selector = ">"
	} else {
		selector = " "
	}

	// Ready indicator
	var readyIcon string
	var readyStyle lipgloss.Style
	if t.Ready {
		readyIcon = "●"
		readyStyle = lipgloss.NewStyle().Foreground(pkgtui.ColorSuccess)
	} else {
		readyIcon = "○"
		readyStyle = lipgloss.NewStyle().Foreground(pkgtui.ColorMuted)
	}

	// Type badge
	typeStyle := lipgloss.NewStyle().
		Background(pkgtui.ColorBgLight).
		Foreground(pkgtui.ColorFgDim).
		Padding(0, 1)

	var typeAbbrev string
	switch t.Type {
	case tasks.TaskTypeImplementation:
		typeAbbrev = "impl"
	case tasks.TaskTypeTest:
		typeAbbrev = "test"
	case tasks.TaskTypeDocumentation:
		typeAbbrev = "docs"
	case tasks.TaskTypeReview:
		typeAbbrev = "review"
	case tasks.TaskTypeSetup:
		typeAbbrev = "setup"
	case tasks.TaskTypeResearch:
		typeAbbrev = "research"
	default:
		typeAbbrev = string(t.Type)
	}
	typeBadge := typeStyle.Render(typeAbbrev)

	// Title
	titleStyle := lipgloss.NewStyle().Foreground(pkgtui.ColorFg)
	if selected {
		titleStyle = titleStyle.Bold(true).Foreground(pkgtui.ColorPrimary)
	}

	// Edited indicator
	editedMark := ""
	if t.Edited {
		editedMark = lipgloss.NewStyle().Foreground(pkgtui.ColorWarning).Render(" *")
	}

	header := fmt.Sprintf("%s%s %s %s  %s%s",
		indentStr,
		selector,
		readyStyle.Render(readyIcon),
		typeBadge,
		titleStyle.Render(t.Title),
		editedMark,
	)
	lines = append(lines, header)

	// Expanded view
	if expanded {
		expandIndent := indentStr + "    "

		// Task ID
		idStyle := pkgtui.LabelStyle
		lines = append(lines, expandIndent+idStyle.Render(t.ID))

		// Description
		if t.Description != "" {
			descStyle := pkgtui.LabelStyle.Width(v.width - len(expandIndent) - 4)
			lines = append(lines, expandIndent+descStyle.Render(t.Description))
		}

		// Dependencies
		if len(t.Dependencies) > 0 {
			depStyle := pkgtui.LabelStyle
			deps := strings.Join(t.Dependencies, ", ")
			lines = append(lines, expandIndent+depStyle.Render("Depends on: "+deps))
		}

		lines = append(lines, "") // Spacer
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (v *TaskReviewView) renderActions() string {
	var actions []string

	keyStyle := pkgtui.HelpKeyStyle
	descStyle := pkgtui.HelpDescStyle

	actions = append(actions, fmt.Sprintf("%s %s",
		keyStyle.Render("Enter"),
		descStyle.Render("accept all")))

	actions = append(actions, fmt.Sprintf("%s %s",
		keyStyle.Render("e"),
		descStyle.Render("edit")))

	actions = append(actions, fmt.Sprintf("%s %s",
		keyStyle.Render("d"),
		descStyle.Render("delete")))

	actions = append(actions, fmt.Sprintf("%s %s",
		keyStyle.Render("tab"),
		descStyle.Render("change type")))

	actions = append(actions, fmt.Sprintf("%s %s",
		keyStyle.Render("g"),
		descStyle.Render("toggle grouping")))

	actions = append(actions, fmt.Sprintf("%s %s",
		keyStyle.Render("space"),
		descStyle.Render("expand")))

	return strings.Join(actions, "  ")
}

// Focus implements View
func (v *TaskReviewView) Focus() tea.Cmd {
	return nil
}

// Blur implements View
func (v *TaskReviewView) Blur() {}

// Name implements View
func (v *TaskReviewView) Name() string {
	return "Tasks"
}

// ShortHelp implements View
func (v *TaskReviewView) ShortHelp() string {
	return "enter accept  e edit  d delete  g group"
}

// GetTasks returns the current tasks (potentially edited).
func (v *TaskReviewView) GetTasks() []tasks.TaskProposal {
	return v.tasks
}
