package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

// BreadcrumbStep represents a step in the breadcrumb
type BreadcrumbStep struct {
	ID       string
	Label    string
	State    OnboardingState
	Unlocked bool // Whether this step has been reached
}

// Breadcrumb displays navigable steps in the onboarding flow
type Breadcrumb struct {
	steps    []BreadcrumbStep
	current  int
	selected int // For keyboard navigation (-1 means not navigating)
	width    int
}

// NewBreadcrumb creates a new breadcrumb with the onboarding steps
func NewBreadcrumb() *Breadcrumb {
	// Derive steps from OnboardingState enum - single source of truth
	states := AllOnboardingStates()
	steps := make([]BreadcrumbStep, len(states))
	for i, state := range states {
		steps[i] = BreadcrumbStep{
			ID:       state.ID(),
			Label:    state.Label(),
			State:    state,
			Unlocked: i == 0, // Only first step unlocked initially
		}
	}

	return &Breadcrumb{
		steps:    steps,
		current:  0,
		selected: -1,
	}
}

// SetWidth sets the available width
func (b *Breadcrumb) SetWidth(w int) {
	b.width = w
}

// SetCurrent sets the current step and unlocks all steps up to it
func (b *Breadcrumb) SetCurrent(state OnboardingState) {
	for i, step := range b.steps {
		if step.State == state {
			b.current = i
			// Unlock all steps up to and including current
			for j := 0; j <= i; j++ {
				b.steps[j].Unlocked = true
			}
			break
		}
	}
	b.selected = -1 // Reset selection when changing current
}

// StartNavigation enables keyboard navigation mode
func (b *Breadcrumb) StartNavigation() {
	b.selected = b.current
}

// StopNavigation disables keyboard navigation mode
func (b *Breadcrumb) StopNavigation() {
	b.selected = -1
}

// IsNavigating returns true if in navigation mode
func (b *Breadcrumb) IsNavigating() bool {
	return b.selected >= 0
}

// Update handles keyboard navigation
func (b *Breadcrumb) Update(msg tea.Msg) (*Breadcrumb, tea.Cmd) {
	if !b.IsNavigating() {
		return b, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			// Find previous unlocked step
			for i := b.selected - 1; i >= 0; i-- {
				if b.steps[i].Unlocked {
					b.selected = i
					break
				}
			}
		case "right", "l":
			// Find next unlocked step
			for i := b.selected + 1; i < len(b.steps); i++ {
				if b.steps[i].Unlocked {
					b.selected = i
					break
				}
			}
		case "enter":
			if b.selected >= 0 && b.selected < len(b.steps) && b.steps[b.selected].Unlocked {
				targetState := b.steps[b.selected].State
				b.selected = -1
				return b, func() tea.Msg {
					return NavigateToStepMsg{State: targetState}
				}
			}
		case "esc":
			b.selected = -1
		}
	}

	return b, nil
}

// View renders the breadcrumb
func (b *Breadcrumb) View() string {
	var parts []string

	// Use a nice arrow separator
	separatorStyle := lipgloss.NewStyle().
		Foreground(pkgtui.ColorMuted).
		Padding(0, 1)
	separator := separatorStyle.Render("›")

	for i, step := range b.steps {
		var style lipgloss.Style

		if i == b.current {
			// Current step - highlighted with background
			style = lipgloss.NewStyle().
				Background(pkgtui.ColorPrimary).
				Foreground(pkgtui.ColorBg).
				Bold(true).
				Padding(0, 1)
		} else if step.Unlocked {
			// Unlocked but not current - subtle but clickable
			style = lipgloss.NewStyle().
				Foreground(pkgtui.ColorFgDim).
				Padding(0, 1)
		} else {
			// Locked - greyed out
			style = lipgloss.NewStyle().
				Foreground(pkgtui.ColorMuted).
				Padding(0, 1)
		}

		// Add selection indicator if navigating
		label := step.Label
		if b.selected == i {
			style = style.Underline(true)
			if step.Unlocked && i != b.current {
				style = style.Background(pkgtui.ColorBgLighter)
			}
		}

		parts = append(parts, style.Render(label))

		// Add separator except after last item
		if i < len(b.steps)-1 {
			parts = append(parts, separator)
		}
	}

	return strings.Join(parts, "")
}

// NavigateToStepMsg is sent when user navigates to a step via breadcrumb
type NavigateToStepMsg struct {
	State OnboardingState
}
