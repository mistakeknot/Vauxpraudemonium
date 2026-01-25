package tui

import (
	"github.com/charmbracelet/lipgloss"
	shared "github.com/mistakeknot/autarch/pkg/tui"
)

var (
	BaseStyle     = shared.BaseStyle
	PanelStyle    = shared.PanelStyle
	TitleStyle    = shared.TitleStyle
	SubtitleStyle = shared.SubtitleStyle
	LabelStyle    = shared.LabelStyle

	SelectedStyle   = shared.SelectedStyle
	UnselectedStyle = shared.UnselectedStyle

	HelpKeyStyle  = shared.HelpKeyStyle
	HelpDescStyle = shared.HelpDescStyle

	TabStyle       = shared.TabStyle
	ActiveTabStyle = shared.ActiveTabStyle

	StatusRunningStyle = shared.StatusRunning
	StatusWaitingStyle = shared.StatusWaiting
	StatusIdleStyle    = shared.StatusIdle
	StatusErrorStyle   = shared.StatusError

	PaneFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(shared.ColorPrimary)
	PaneUnfocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(shared.ColorMuted)
)
