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

// GurgehView displays specs (PRDs)
type GurgehView struct {
	client   *autarch.Client
	specs    []autarch.Spec
	selected int
	width    int
	height   int
	loading  bool
	err      error
}

// NewGurgehView creates a new Gurgeh view
func NewGurgehView(client *autarch.Client) *GurgehView {
	return &GurgehView{
		client: client,
	}
}

type specsLoadedMsg struct {
	specs []autarch.Spec
	err   error
}

// Init implements View
func (v *GurgehView) Init() tea.Cmd {
	return v.loadSpecs()
}

func (v *GurgehView) loadSpecs() tea.Cmd {
	return func() tea.Msg {
		specs, err := v.client.ListSpecs("")
		return specsLoadedMsg{specs: specs, err: err}
	}
}

// Update implements View
func (v *GurgehView) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height - 4
		return v, nil

	case specsLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
		} else {
			v.specs = msg.specs
		}
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if v.selected < len(v.specs)-1 {
				v.selected++
			}
		case "k", "up":
			if v.selected > 0 {
				v.selected--
			}
		case "r":
			v.loading = true
			return v, v.loadSpecs()
		}
	}

	return v, nil
}

// View implements View
func (v *GurgehView) View() string {
	if v.loading {
		return pkgtui.LabelStyle.Render("Loading specs...")
	}

	if v.err != nil {
		return tui.ErrorView(v.err)
	}

	if len(v.specs) == 0 {
		return v.renderEmptyState()
	}

	return v.renderSplitView()
}

func (v *GurgehView) renderEmptyState() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		pkgtui.TitleStyle.Render("Specs"),
		"",
		pkgtui.LabelStyle.Render("No specs found"),
		"",
		pkgtui.LabelStyle.Render("Create a new spec with 'n' or use the command palette."),
	)
}

func (v *GurgehView) renderSplitView() string {
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

func (v *GurgehView) renderList(width int) string {
	var lines []string

	// Group by status
	grouped := map[autarch.SpecStatus][]autarch.Spec{}
	statusOrder := []autarch.SpecStatus{
		autarch.SpecStatusDraft,
		autarch.SpecStatusResearch,
		autarch.SpecStatusValidated,
		autarch.SpecStatusArchived,
	}

	for _, s := range v.specs {
		grouped[s.Status] = append(grouped[s.Status], s)
	}

	lines = append(lines, pkgtui.TitleStyle.Render("Specs"))
	lines = append(lines, "")

	idx := 0
	for _, status := range statusOrder {
		specs := grouped[status]
		if len(specs) == 0 {
			continue
		}

		// Status header
		header := fmt.Sprintf("▾ %s (%d)", status, len(specs))
		lines = append(lines, pkgtui.SubtitleStyle.Render(header))

		for _, s := range specs {
			title := s.Title
			if title == "" {
				title = s.ID[:8]
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

func (v *GurgehView) renderDetail(width int) string {
	var lines []string

	lines = append(lines, pkgtui.TitleStyle.Render("Details"))
	lines = append(lines, "")

	if len(v.specs) == 0 || v.selected >= len(v.specs) {
		lines = append(lines, pkgtui.LabelStyle.Render("No spec selected"))
		return strings.Join(lines, "\n")
	}

	s := v.specs[v.selected]

	lines = append(lines, fmt.Sprintf("Title: %s", s.Title))
	lines = append(lines, fmt.Sprintf("Status: %s", s.Status))
	lines = append(lines, fmt.Sprintf("Project: %s", s.Project))
	lines = append(lines, "")

	if s.Vision != "" {
		lines = append(lines, pkgtui.SubtitleStyle.Render("Vision"))
		lines = append(lines, s.Vision)
		lines = append(lines, "")
	}

	if s.Problem != "" {
		lines = append(lines, pkgtui.SubtitleStyle.Render("Problem"))
		lines = append(lines, s.Problem)
		lines = append(lines, "")
	}

	if s.Users != "" {
		lines = append(lines, pkgtui.SubtitleStyle.Render("Users"))
		lines = append(lines, s.Users)
	}

	return strings.Join(lines, "\n")
}

// Focus implements View
func (v *GurgehView) Focus() tea.Cmd {
	return v.loadSpecs()
}

// Blur implements View
func (v *GurgehView) Blur() {}

// Name implements View
func (v *GurgehView) Name() string {
	return "Gurgeh"
}

// ShortHelp implements View
func (v *GurgehView) ShortHelp() string {
	return "j/k navigate  n new  r refresh"
}

// Commands implements CommandProvider
func (v *GurgehView) Commands() []tui.Command {
	return []tui.Command{
		{
			Name:        "New Spec",
			Description: "Create a new specification",
			Action: func() tea.Cmd {
				// TODO: implement
				return nil
			},
		},
		{
			Name:        "Refresh Specs",
			Description: "Reload spec list",
			Action: func() tea.Cmd {
				return v.loadSpecs()
			},
		},
	}
}
