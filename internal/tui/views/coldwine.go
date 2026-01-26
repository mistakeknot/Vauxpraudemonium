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

// ColdwineView displays epics, stories, and tasks
type ColdwineView struct {
	client   *autarch.Client
	epics    []autarch.Epic
	stories  []autarch.Story
	tasks    []autarch.Task
	selected int
	width    int
	height   int
	loading  bool
	err      error
}

// NewColdwineView creates a new Coldwine view
func NewColdwineView(client *autarch.Client) *ColdwineView {
	return &ColdwineView{
		client: client,
	}
}

type epicsLoadedMsg struct {
	epics   []autarch.Epic
	stories []autarch.Story
	tasks   []autarch.Task
	err     error
}

// Init implements View
func (v *ColdwineView) Init() tea.Cmd {
	return v.loadData()
}

func (v *ColdwineView) loadData() tea.Cmd {
	return func() tea.Msg {
		epics, err := v.client.ListEpics("")
		if err != nil {
			return epicsLoadedMsg{err: err}
		}
		stories, err := v.client.ListStories("")
		if err != nil {
			return epicsLoadedMsg{err: err}
		}
		tasks, err := v.client.ListTasks("", "")
		if err != nil {
			return epicsLoadedMsg{err: err}
		}
		return epicsLoadedMsg{epics: epics, stories: stories, tasks: tasks}
	}
}

// Update implements View
func (v *ColdwineView) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height - 4
		return v, nil

	case epicsLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
		} else {
			v.epics = msg.epics
			v.stories = msg.stories
			v.tasks = msg.tasks
		}
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if v.selected < len(v.epics)-1 {
				v.selected++
			}
		case "k", "up":
			if v.selected > 0 {
				v.selected--
			}
		case "r":
			v.loading = true
			return v, v.loadData()
		}
	}

	return v, nil
}

// View implements View
func (v *ColdwineView) View() string {
	if v.loading {
		return pkgtui.LabelStyle.Render("Loading epics...")
	}

	if v.err != nil {
		return tui.ErrorView(v.err)
	}

	if len(v.epics) == 0 {
		return v.renderEmptyState()
	}

	return v.renderSplitView()
}

func (v *ColdwineView) renderEmptyState() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		pkgtui.TitleStyle.Render("Epics & Tasks"),
		"",
		pkgtui.LabelStyle.Render("No epics found"),
		"",
		pkgtui.LabelStyle.Render("Create an epic to break down a spec into implementable work."),
	)
}

func (v *ColdwineView) renderSplitView() string {
	listWidth := v.width / 3
	detailWidth := v.width - listWidth - 3

	list := v.renderList(listWidth)
	detail := v.renderDetail(detailWidth)

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

func (v *ColdwineView) renderList(width int) string {
	var lines []string

	lines = append(lines, pkgtui.TitleStyle.Render("Epics"))
	lines = append(lines, "")

	for i, e := range v.epics {
		// Count stories for this epic
		storyCount := 0
		doneCount := 0
		for _, s := range v.stories {
			if s.EpicID == e.ID {
				storyCount++
				if s.Status == autarch.StoryStatusDone {
					doneCount++
				}
			}
		}

		icon := v.statusIcon(e.Status)
		title := e.Title
		if title == "" {
			title = e.ID[:8]
		}

		line := fmt.Sprintf("%s %s (%d/%d)", icon, title, doneCount, storyCount)
		if i == v.selected {
			line = pkgtui.SelectedStyle.Render(line)
		} else {
			line = pkgtui.UnselectedStyle.Render(line)
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (v *ColdwineView) renderDetail(width int) string {
	var lines []string

	lines = append(lines, pkgtui.TitleStyle.Render("Stories"))
	lines = append(lines, "")

	if len(v.epics) == 0 || v.selected >= len(v.epics) {
		lines = append(lines, pkgtui.LabelStyle.Render("No epic selected"))
		return strings.Join(lines, "\n")
	}

	e := v.epics[v.selected]

	lines = append(lines, fmt.Sprintf("Epic: %s", e.Title))
	lines = append(lines, fmt.Sprintf("Status: %s", e.Status))
	lines = append(lines, "")

	// Show stories for this epic
	lines = append(lines, pkgtui.SubtitleStyle.Render("Stories"))

	foundStories := false
	for _, s := range v.stories {
		if s.EpicID == e.ID {
			foundStories = true
			icon := v.storyStatusIcon(s.Status)
			lines = append(lines, fmt.Sprintf("  %s %s", icon, s.Title))
		}
	}

	if !foundStories {
		lines = append(lines, pkgtui.LabelStyle.Render("  No stories"))
	}

	return strings.Join(lines, "\n")
}

func (v *ColdwineView) statusIcon(status autarch.EpicStatus) string {
	switch status {
	case autarch.EpicStatusOpen:
		return pkgtui.StatusIdle.Render("○")
	case autarch.EpicStatusInProgress:
		return pkgtui.StatusRunning.Render("●")
	case autarch.EpicStatusDone:
		return pkgtui.StatusRunning.Render("✓")
	default:
		return pkgtui.StatusIdle.Render("?")
	}
}

func (v *ColdwineView) storyStatusIcon(status autarch.StoryStatus) string {
	switch status {
	case autarch.StoryStatusTodo:
		return pkgtui.StatusIdle.Render("○")
	case autarch.StoryStatusInProgress:
		return pkgtui.StatusRunning.Render("●")
	case autarch.StoryStatusReview:
		return pkgtui.StatusWaiting.Render("◐")
	case autarch.StoryStatusDone:
		return pkgtui.StatusRunning.Render("✓")
	default:
		return pkgtui.StatusIdle.Render("?")
	}
}

// Focus implements View
func (v *ColdwineView) Focus() tea.Cmd {
	return v.loadData()
}

// Blur implements View
func (v *ColdwineView) Blur() {}

// Name implements View
func (v *ColdwineView) Name() string {
	return "Coldwine"
}

// ShortHelp implements View
func (v *ColdwineView) ShortHelp() string {
	return "j/k navigate  r refresh"
}

// Commands implements CommandProvider
func (v *ColdwineView) Commands() []tui.Command {
	return []tui.Command{
		{
			Name:        "New Epic",
			Description: "Create a new epic",
			Action: func() tea.Cmd {
				return nil
			},
		},
		{
			Name:        "New Story",
			Description: "Create a new story",
			Action: func() tea.Cmd {
				return nil
			},
		},
		{
			Name:        "New Task",
			Description: "Create a new task",
			Action: func() tea.Cmd {
				return nil
			},
		},
	}
}
