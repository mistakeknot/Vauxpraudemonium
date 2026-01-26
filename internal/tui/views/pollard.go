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

// PollardView displays research insights
type PollardView struct {
	client   *autarch.Client
	insights []autarch.Insight
	selected int
	width    int
	height   int
	loading  bool
	err      error
}

// NewPollardView creates a new Pollard view
func NewPollardView(client *autarch.Client) *PollardView {
	return &PollardView{
		client: client,
	}
}

type insightsLoadedMsg struct {
	insights []autarch.Insight
	err      error
}

// Init implements View
func (v *PollardView) Init() tea.Cmd {
	return v.loadInsights()
}

func (v *PollardView) loadInsights() tea.Cmd {
	return func() tea.Msg {
		insights, err := v.client.ListInsights("", "")
		return insightsLoadedMsg{insights: insights, err: err}
	}
}

// Update implements View
func (v *PollardView) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height - 4
		return v, nil

	case insightsLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
		} else {
			v.insights = msg.insights
		}
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if v.selected < len(v.insights)-1 {
				v.selected++
			}
		case "k", "up":
			if v.selected > 0 {
				v.selected--
			}
		case "r":
			v.loading = true
			return v, v.loadInsights()
		}
	}

	return v, nil
}

// View implements View
func (v *PollardView) View() string {
	if v.loading {
		return pkgtui.LabelStyle.Render("Loading insights...")
	}

	if v.err != nil {
		return tui.ErrorView(v.err)
	}

	if len(v.insights) == 0 {
		return v.renderEmptyState()
	}

	return v.renderSplitView()
}

func (v *PollardView) renderEmptyState() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		pkgtui.TitleStyle.Render("Research Insights"),
		"",
		pkgtui.LabelStyle.Render("No insights found"),
		"",
		pkgtui.LabelStyle.Render("Run Pollard hunters to gather research insights."),
	)
}

func (v *PollardView) renderSplitView() string {
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

func (v *PollardView) renderList(width int) string {
	var lines []string

	// Group by category
	grouped := map[string][]autarch.Insight{}
	for _, i := range v.insights {
		grouped[i.Category] = append(grouped[i.Category], i)
	}

	lines = append(lines, pkgtui.TitleStyle.Render("Insights"))
	lines = append(lines, "")

	idx := 0
	for category, insights := range grouped {
		// Category header
		header := fmt.Sprintf("▾ %s (%d)", category, len(insights))
		lines = append(lines, pkgtui.SubtitleStyle.Render(header))

		for _, i := range insights {
			title := i.Title
			if len(title) > 30 {
				title = title[:27] + "..."
			}

			line := "  " + title
			if idx == v.selected {
				line = pkgtui.SelectedStyle.Render(line)
			} else {
				line = pkgtui.UnselectedStyle.Render(line)
			}
			lines = append(lines, line)
			idx++
		}
	}

	return strings.Join(lines, "\n")
}

func (v *PollardView) renderDetail(width int) string {
	var lines []string

	lines = append(lines, pkgtui.TitleStyle.Render("Details"))
	lines = append(lines, "")

	if len(v.insights) == 0 || v.selected >= len(v.insights) {
		lines = append(lines, pkgtui.LabelStyle.Render("No insight selected"))
		return strings.Join(lines, "\n")
	}

	i := v.insights[v.selected]

	lines = append(lines, fmt.Sprintf("Title: %s", i.Title))
	lines = append(lines, fmt.Sprintf("Category: %s", i.Category))
	lines = append(lines, fmt.Sprintf("Source: %s", i.Source))
	lines = append(lines, fmt.Sprintf("Score: %.2f", i.Score))
	lines = append(lines, "")

	if i.Body != "" {
		lines = append(lines, pkgtui.SubtitleStyle.Render("Summary"))
		// Wrap body text
		wrapped := wordWrap(i.Body, width-4)
		lines = append(lines, wrapped...)
		lines = append(lines, "")
	}

	if i.URL != "" {
		lines = append(lines, fmt.Sprintf("URL: %s", i.URL))
	}

	if i.SpecID != "" {
		lines = append(lines, fmt.Sprintf("Linked Spec: %s", i.SpecID))
	}

	return strings.Join(lines, "\n")
}

func wordWrap(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	var lines []string
	var current strings.Builder

	for _, word := range words {
		if current.Len()+len(word)+1 > width {
			if current.Len() > 0 {
				lines = append(lines, current.String())
				current.Reset()
			}
		}
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(word)
	}

	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return lines
}

// Focus implements View
func (v *PollardView) Focus() tea.Cmd {
	return v.loadInsights()
}

// Blur implements View
func (v *PollardView) Blur() {}

// Name implements View
func (v *PollardView) Name() string {
	return "Pollard"
}

// ShortHelp implements View
func (v *PollardView) ShortHelp() string {
	return "j/k navigate  r refresh"
}

// Commands implements CommandProvider
func (v *PollardView) Commands() []tui.Command {
	return []tui.Command{
		{
			Name:        "Run Research",
			Description: "Execute Pollard hunters",
			Action: func() tea.Cmd {
				return nil
			},
		},
		{
			Name:        "Link Insight",
			Description: "Link insight to a spec",
			Action: func() tea.Cmd {
				return nil
			},
		},
	}
}
