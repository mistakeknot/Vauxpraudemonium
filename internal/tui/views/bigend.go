package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakeknot/autarch/internal/tui"
	"github.com/mistakeknot/autarch/pkg/autarch"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

// BigendView displays sessions and agent overview
type BigendView struct {
	client   *autarch.Client
	sessions []autarch.Session
	selected int
	width    int
	height   int
	loading  bool
	err      error
}

// NewBigendView creates a new Bigend view
func NewBigendView(client *autarch.Client) *BigendView {
	return &BigendView{
		client: client,
	}
}

// sessionsLoadedMsg is sent when sessions are loaded
type sessionsLoadedMsg struct {
	sessions []autarch.Session
	err      error
}

// Init implements View
func (v *BigendView) Init() tea.Cmd {
	return v.loadSessions()
}

func (v *BigendView) loadSessions() tea.Cmd {
	return func() tea.Msg {
		sessions, err := v.client.ListSessions("")
		return sessionsLoadedMsg{sessions: sessions, err: err}
	}
}

// Update implements View
func (v *BigendView) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height - 4 // Account for tabs and footer
		return v, nil

	case sessionsLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
		} else {
			v.sessions = msg.sessions
		}
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if v.selected < len(v.sessions)-1 {
				v.selected++
			}
		case "k", "up":
			if v.selected > 0 {
				v.selected--
			}
		case "r":
			v.loading = true
			return v, v.loadSessions()
		}
	}

	return v, nil
}

// View implements View
func (v *BigendView) View() string {
	if v.loading {
		return pkgtui.LabelStyle.Render("Loading sessions...")
	}

	if v.err != nil {
		return tui.ErrorView(v.err)
	}

	if len(v.sessions) == 0 {
		return v.renderEmptyState()
	}

	return v.renderSplitView()
}

func (v *BigendView) renderEmptyState() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		pkgtui.TitleStyle.Render("Sessions"),
		"",
		pkgtui.LabelStyle.Render("No sessions running"),
		"",
		pkgtui.LabelStyle.Render("Sessions will appear here when agents are started."),
	)
}

func (v *BigendView) renderSplitView() string {
	// Calculate widths
	listWidth := v.width / 3
	detailWidth := v.width - listWidth - 3

	// Render list
	list := v.renderList(listWidth)

	// Render detail
	detail := v.renderDetail(detailWidth)

	// Join horizontally
	listStyle := lipgloss.NewStyle().
		Width(listWidth).
		Height(v.height)

	detailStyle := lipgloss.NewStyle().
		Width(detailWidth).
		Height(v.height).
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(pkgtui.ColorMuted).
		PaddingLeft(1)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		listStyle.Render(list),
		detailStyle.Render(detail),
	)
}

func (v *BigendView) renderList(width int) string {
	var lines []string

	lines = append(lines, pkgtui.TitleStyle.Render("Sessions"))
	lines = append(lines, "")

	for i, s := range v.sessions {
		icon := v.statusIcon(s.Status)
		name := s.Name
		if name == "" {
			name = s.ID[:8]
		}

		line := fmt.Sprintf("%s %s", icon, name)
		if i == v.selected {
			line = pkgtui.SelectedStyle.Render(line)
		} else {
			line = pkgtui.UnselectedStyle.Render(line)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (v *BigendView) renderDetail(width int) string {
	var lines []string

	lines = append(lines, pkgtui.TitleStyle.Render("Details"))
	lines = append(lines, "")

	if len(v.sessions) == 0 || v.selected >= len(v.sessions) {
		lines = append(lines, pkgtui.LabelStyle.Render("No session selected"))
		return strings.Join(lines, "\n")
	}

	s := v.sessions[v.selected]

	lines = append(lines, fmt.Sprintf("Name: %s", s.Name))
	lines = append(lines, fmt.Sprintf("Agent: %s", s.Agent))
	lines = append(lines, fmt.Sprintf("Status: %s", s.Status))
	lines = append(lines, fmt.Sprintf("Project: %s", s.Project))

	if s.TaskID != "" {
		lines = append(lines, fmt.Sprintf("Task: %s", s.TaskID))
	}

	lines = append(lines, fmt.Sprintf("Started: %s", s.StartedAt.Format("2006-01-02 15:04")))

	return strings.Join(lines, "\n")
}

func (v *BigendView) statusIcon(status autarch.SessionStatus) string {
	switch status {
	case autarch.SessionStatusRunning:
		return pkgtui.StatusRunning.Render("●")
	case autarch.SessionStatusIdle:
		return pkgtui.StatusIdle.Render("○")
	case autarch.SessionStatusError:
		return pkgtui.StatusError.Render("✕")
	default:
		return pkgtui.StatusIdle.Render("?")
	}
}

// Focus implements View
func (v *BigendView) Focus() tea.Cmd {
	return v.loadSessions()
}

// Blur implements View
func (v *BigendView) Blur() {}

// Name implements View
func (v *BigendView) Name() string {
	return "Bigend"
}

// ShortHelp implements View
func (v *BigendView) ShortHelp() string {
	return "j/k navigate  r refresh"
}

// Commands implements CommandProvider
func (v *BigendView) Commands() []tui.Command {
	return []tui.Command{
		{
			Name:        "New Session",
			Description: "Start a new agent session",
			Action: func() tea.Cmd {
				// TODO: implement
				return nil
			},
		},
		{
			Name:        "Refresh Sessions",
			Description: "Reload session list",
			Action: func() tea.Cmd {
				return v.loadSessions()
			},
		},
	}
}
