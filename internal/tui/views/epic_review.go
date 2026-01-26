package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakeknot/autarch/internal/coldwine/epics"
	"github.com/mistakeknot/autarch/internal/tui"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

// EpicReviewView displays proposed epics for review and editing.
type EpicReviewView struct {
	proposals []epics.EpicProposal
	selected  int
	expanded  map[int]bool
	width     int
	height    int
	editing   bool // In inline edit mode
	editField string

	// Callbacks
	onAccept     func(proposals []epics.EpicProposal) tea.Cmd
	onRegenerate func() tea.Cmd
}

// NewEpicReviewView creates a new epic review view.
func NewEpicReviewView(proposals []epics.EpicProposal) *EpicReviewView {
	return &EpicReviewView{
		proposals: proposals,
		expanded:  make(map[int]bool),
	}
}

// SetCallbacks sets the action callbacks.
func (v *EpicReviewView) SetCallbacks(
	onAccept func([]epics.EpicProposal) tea.Cmd,
	onRegenerate func() tea.Cmd,
) {
	v.onAccept = onAccept
	v.onRegenerate = onRegenerate
}

// Init implements View
func (v *EpicReviewView) Init() tea.Cmd {
	return nil
}

// Update implements View
func (v *EpicReviewView) Update(msg tea.Msg) (tui.View, tea.Cmd) {
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
				return v, v.onAccept(v.proposals)
			}
			return v, nil

		case "r":
			// Regenerate (with warning if edited)
			if v.hasEdits() {
				// Show warning - for now just regenerate
				// TODO: Add confirmation dialog
			}
			if v.onRegenerate != nil {
				return v, v.onRegenerate()
			}
			return v, nil

		case "e":
			// Edit selected
			v.editing = true
			return v, nil

		case "d":
			// Delete selected
			if v.selected >= 0 && v.selected < len(v.proposals) {
				v.proposals = append(v.proposals[:v.selected], v.proposals[v.selected+1:]...)
				if v.selected >= len(v.proposals) {
					v.selected = len(v.proposals) - 1
				}
			}
			return v, nil

		case "j", "down":
			if v.selected < len(v.proposals)-1 {
				v.selected++
			}
			return v, nil

		case "k", "up":
			if v.selected > 0 {
				v.selected--
			}
			return v, nil

		case "space":
			// Toggle expand
			v.expanded[v.selected] = !v.expanded[v.selected]
			return v, nil

		case "esc":
			if v.editing {
				v.editing = false
			}
			return v, nil

		case "+":
			// Increase priority
			if v.selected >= 0 && v.selected < len(v.proposals) {
				v.proposals[v.selected].Edited = true
				switch v.proposals[v.selected].Priority {
				case epics.PriorityP3:
					v.proposals[v.selected].Priority = epics.PriorityP2
				case epics.PriorityP2:
					v.proposals[v.selected].Priority = epics.PriorityP1
				case epics.PriorityP1:
					v.proposals[v.selected].Priority = epics.PriorityP0
				}
			}
			return v, nil

		case "-":
			// Decrease priority
			if v.selected >= 0 && v.selected < len(v.proposals) {
				v.proposals[v.selected].Edited = true
				switch v.proposals[v.selected].Priority {
				case epics.PriorityP0:
					v.proposals[v.selected].Priority = epics.PriorityP1
				case epics.PriorityP1:
					v.proposals[v.selected].Priority = epics.PriorityP2
				case epics.PriorityP2:
					v.proposals[v.selected].Priority = epics.PriorityP3
				}
			}
			return v, nil
		}
	}

	return v, nil
}

func (v *EpicReviewView) hasEdits() bool {
	for _, p := range v.proposals {
		if p.Edited {
			return true
		}
	}
	return false
}

// View implements View
func (v *EpicReviewView) View() string {
	if len(v.proposals) == 0 {
		return pkgtui.LabelStyle.Render("No epics proposed")
	}

	var sections []string

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(pkgtui.ColorPrimary).
		Bold(true).
		MarginBottom(1)
	totalTasks := epics.EstimateTotalTasks(v.proposals)
	header := fmt.Sprintf("Epic Review (%d epics, ~%d tasks)", len(v.proposals), totalTasks)
	sections = append(sections, headerStyle.Render(header))

	// Epic list
	for i, p := range v.proposals {
		isSelected := i == v.selected
		isExpanded := v.expanded[i]
		epicView := v.renderEpic(p, isSelected, isExpanded)
		sections = append(sections, epicView)
	}

	sections = append(sections, "")

	// Actions
	sections = append(sections, v.renderActions())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (v *EpicReviewView) renderEpic(p epics.EpicProposal, selected, expanded bool) string {
	var lines []string

	// Epic header
	var selector string
	if selected {
		selector = ">"
	} else {
		selector = " "
	}

	// Size badge
	sizeStyle := lipgloss.NewStyle().
		Background(pkgtui.ColorBgLight).
		Foreground(pkgtui.ColorFgDim).
		Padding(0, 1)

	sizeBadge := sizeStyle.Render(string(p.Size))

	// Priority badge
	var priorityStyle lipgloss.Style
	switch p.Priority {
	case epics.PriorityP0:
		priorityStyle = lipgloss.NewStyle().Foreground(pkgtui.ColorError).Bold(true)
	case epics.PriorityP1:
		priorityStyle = lipgloss.NewStyle().Foreground(pkgtui.ColorWarning).Bold(true)
	case epics.PriorityP2:
		priorityStyle = lipgloss.NewStyle().Foreground(pkgtui.ColorPrimary)
	default:
		priorityStyle = lipgloss.NewStyle().Foreground(pkgtui.ColorMuted)
	}
	priorityBadge := priorityStyle.Render(string(p.Priority))

	// Title
	titleStyle := lipgloss.NewStyle().Foreground(pkgtui.ColorFg)
	if selected {
		titleStyle = titleStyle.Bold(true).Foreground(pkgtui.ColorPrimary)
	}

	// Edited indicator
	editedMark := ""
	if p.Edited {
		editedMark = lipgloss.NewStyle().Foreground(pkgtui.ColorWarning).Render(" (edited)")
	}

	header := fmt.Sprintf("%s %s %s  %s  %s%s",
		selector,
		p.ID,
		sizeBadge,
		priorityBadge,
		titleStyle.Render(p.Title),
		editedMark,
	)
	lines = append(lines, header)

	// Dependencies
	if len(p.Dependencies) > 0 {
		depStyle := pkgtui.LabelStyle.MarginLeft(4)
		deps := strings.Join(p.Dependencies, ", ")
		lines = append(lines, depStyle.Render(fmt.Sprintf("depends on: %s", deps)))
	}

	// Task estimate
	taskStyle := pkgtui.LabelStyle.MarginLeft(4)
	lines = append(lines, taskStyle.Render(fmt.Sprintf("~%d tasks", p.TaskCount)))

	// Expanded view
	if expanded {
		// Description
		if p.Description != "" {
			descStyle := pkgtui.LabelStyle.MarginLeft(4).Width(v.width - 10)
			lines = append(lines, "")
			lines = append(lines, descStyle.Render(p.Description))
		}

		// Stories
		if len(p.Stories) > 0 {
			lines = append(lines, "")
			storyHeaderStyle := pkgtui.SubtitleStyle.MarginLeft(4)
			lines = append(lines, storyHeaderStyle.Render("Stories:"))
			for _, s := range p.Stories {
				storySizeStyle := lipgloss.NewStyle().
					Background(pkgtui.ColorBgLight).
					Foreground(pkgtui.ColorFgDim).
					Padding(0, 1)
				storyLine := fmt.Sprintf("      %s %s - %s",
					storySizeStyle.Render(string(s.Size)),
					s.ID,
					s.Title,
				)
				lines = append(lines, storyLine)
			}
		}
	}

	lines = append(lines, "") // Spacer

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (v *EpicReviewView) renderActions() string {
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
		keyStyle.Render("+/-"),
		descStyle.Render("priority")))

	actions = append(actions, fmt.Sprintf("%s %s",
		keyStyle.Render("r"),
		descStyle.Render("regenerate")))

	actions = append(actions, fmt.Sprintf("%s %s",
		keyStyle.Render("space"),
		descStyle.Render("expand")))

	return strings.Join(actions, "  ")
}

// Focus implements View
func (v *EpicReviewView) Focus() tea.Cmd {
	return nil
}

// Blur implements View
func (v *EpicReviewView) Blur() {}

// Name implements View
func (v *EpicReviewView) Name() string {
	return "Epics"
}

// ShortHelp implements View
func (v *EpicReviewView) ShortHelp() string {
	return "enter accept  e edit  d delete  space expand"
}

// GetProposals returns the current proposals (potentially edited).
func (v *EpicReviewView) GetProposals() []epics.EpicProposal {
	return v.proposals
}
