package views

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/mistakeknot/autarch/internal/tui"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

// KickoffView is the initial view for starting new projects or resuming drafts.
// It provides a text input for project description and a list of recent projects.
type KickoffView struct {
	input      textinput.Model
	recents    []RecentProject
	selected   int
	focusInput bool // true = input focused, false = recents focused
	width      int
	height     int
	loading    bool
	err        error

	// Callbacks for navigation
	onProjectStart func(project *Project) tea.Cmd
}

// RecentProject represents a project that can be resumed or continued.
type RecentProject struct {
	ID       string
	Name     string
	Status   string    // "draft", "complete"
	LastOpen time.Time
	Path     string
}

// Project represents a new or existing project.
type Project struct {
	ID          string
	Name        string
	Description string
	Path        string
	CreatedAt   time.Time
}

// NewKickoffView creates a new kickoff view.
func NewKickoffView() *KickoffView {
	ti := textinput.New()
	ti.Placeholder = "Describe what you want to build..."
	ti.CharLimit = 500
	ti.Width = 60

	return &KickoffView{
		input:      ti,
		focusInput: true,
	}
}

// SetProjectStartCallback sets the callback for when a project is started.
func (v *KickoffView) SetProjectStartCallback(cb func(*Project) tea.Cmd) {
	v.onProjectStart = cb
}

// Init implements View
func (v *KickoffView) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		v.loadRecentProjects(),
	)
}

// recentsLoadedMsg is sent when recent projects are loaded.
type recentsLoadedMsg struct {
	recents []RecentProject
	err     error
}

// projectCreatedMsg is sent when a new project is created.
type projectCreatedMsg struct {
	project *Project
	err     error
}

func (v *KickoffView) loadRecentProjects() tea.Cmd {
	return func() tea.Msg {
		recents, err := loadRecentProjectsFromDisk()
		return recentsLoadedMsg{recents: recents, err: err}
	}
}

// loadRecentProjectsFromDisk reads recent projects from ~/.autarch/projects/
func loadRecentProjectsFromDisk() ([]RecentProject, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	projectsDir := filepath.Join(home, ".autarch", "projects")
	entries, err := os.ReadDir(projectsDir)
	if os.IsNotExist(err) {
		return nil, nil // No projects yet
	}
	if err != nil {
		return nil, err
	}

	var recents []RecentProject
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(projectsDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Try to read project metadata
		status := "complete"
		metaPath := filepath.Join(projectPath, "meta.json")
		if _, err := os.Stat(filepath.Join(projectPath, "draft.json")); err == nil {
			status = "draft"
		}

		// Use directory name as project name
		name := entry.Name()
		if metaData, err := os.ReadFile(metaPath); err == nil {
			// Could parse JSON for better name, but keep it simple
			_ = metaData
		}

		recents = append(recents, RecentProject{
			ID:       entry.Name(),
			Name:     name,
			Status:   status,
			LastOpen: info.ModTime(),
			Path:     projectPath,
		})
	}

	// Sort by last open time, most recent first
	sort.Slice(recents, func(i, j int) bool {
		return recents[i].LastOpen.After(recents[j].LastOpen)
	})

	// Limit to 10 most recent
	if len(recents) > 10 {
		recents = recents[:10]
	}

	return recents, nil
}

// Update implements View
func (v *KickoffView) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height - 4
		v.input.Width = min(60, v.width-10)
		return v, nil

	case recentsLoadedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
		} else {
			v.recents = msg.recents
		}
		return v, nil

	case projectCreatedMsg:
		v.loading = false
		if msg.err != nil {
			v.err = msg.err
			return v, nil
		}
		if v.onProjectStart != nil {
			return v, v.onProjectStart(msg.project)
		}
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Toggle focus between input and recents
			if len(v.recents) > 0 {
				v.focusInput = !v.focusInput
				if v.focusInput {
					v.input.Focus()
				} else {
					v.input.Blur()
				}
			}
			return v, nil

		case "up", "k":
			if !v.focusInput && v.selected > 0 {
				v.selected--
			}
			return v, nil

		case "down", "j":
			if !v.focusInput && v.selected < len(v.recents)-1 {
				v.selected++
			}
			return v, nil

		case "enter":
			if v.focusInput && strings.TrimSpace(v.input.Value()) != "" {
				// Create new project
				v.loading = true
				return v, v.createProject(v.input.Value())
			} else if !v.focusInput && len(v.recents) > 0 {
				// Resume/open existing project
				recent := v.recents[v.selected]
				project := &Project{
					ID:        recent.ID,
					Name:      recent.Name,
					Path:      recent.Path,
					CreatedAt: recent.LastOpen,
				}
				if v.onProjectStart != nil {
					return v, v.onProjectStart(project)
				}
			}
			return v, nil
		}

		// Pass key to input if focused
		if v.focusInput {
			var cmd tea.Cmd
			v.input, cmd = v.input.Update(msg)
			return v, cmd
		}
	}

	return v, nil
}

func (v *KickoffView) createProject(description string) tea.Cmd {
	return func() tea.Msg {
		home, err := os.UserHomeDir()
		if err != nil {
			return projectCreatedMsg{err: err}
		}

		// Generate project ID and slug
		projectID := uuid.New().String()
		slug := slugify(description)
		if len(slug) > 30 {
			slug = slug[:30]
		}
		slug = fmt.Sprintf("%s-%s", slug, projectID[:8])

		projectPath := filepath.Join(home, ".autarch", "projects", slug)
		if err := os.MkdirAll(projectPath, 0755); err != nil {
			return projectCreatedMsg{err: err}
		}

		project := &Project{
			ID:          projectID,
			Name:        slug,
			Description: description,
			Path:        projectPath,
			CreatedAt:   time.Now(),
		}

		return projectCreatedMsg{project: project}
	}
}

// slugify converts a description to a URL-friendly slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, s)

	// Collapse multiple dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")

	return s
}

// View implements View
func (v *KickoffView) View() string {
	if v.loading {
		return pkgtui.LabelStyle.Render("Creating project...")
	}

	if v.err != nil {
		return tui.ErrorView(v.err)
	}

	var sections []string

	// Header
	header := lipgloss.NewStyle().
		Foreground(pkgtui.ColorPrimary).
		Bold(true).
		MarginBottom(1).
		Render("What do you want to build?")
	sections = append(sections, header)

	// Input field
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1)

	if v.focusInput {
		inputStyle = inputStyle.BorderForeground(pkgtui.ColorPrimary)
	} else {
		inputStyle = inputStyle.BorderForeground(pkgtui.ColorMuted)
	}

	inputBox := inputStyle.Render(v.input.View())
	sections = append(sections, inputBox)

	// Hint
	hint := pkgtui.LabelStyle.Render("Describe your project idea in a sentence or two")
	sections = append(sections, hint)

	// Recent projects
	if len(v.recents) > 0 {
		sections = append(sections, "")
		recentHeader := pkgtui.SubtitleStyle.Render("Recent Projects")
		sections = append(sections, recentHeader)

		for i, r := range v.recents {
			line := v.renderRecentProject(r, i == v.selected && !v.focusInput)
			sections = append(sections, line)
		}
	}

	// Help
	sections = append(sections, "")
	helpStyle := pkgtui.LabelStyle
	if len(v.recents) > 0 {
		sections = append(sections, helpStyle.Render("Tab to switch focus  Enter to select"))
	} else {
		sections = append(sections, helpStyle.Render("Enter to create project"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (v *KickoffView) renderRecentProject(r RecentProject, selected bool) string {
	// Status icon
	var icon string
	var iconStyle lipgloss.Style
	if r.Status == "draft" {
		icon = "◐"
		iconStyle = lipgloss.NewStyle().Foreground(pkgtui.ColorWarning)
	} else {
		icon = "●"
		iconStyle = lipgloss.NewStyle().Foreground(pkgtui.ColorSuccess)
	}

	// Time ago
	timeAgo := timeAgoString(r.LastOpen)

	// Line content
	content := fmt.Sprintf("%s %s  %s",
		iconStyle.Render(icon),
		r.Name,
		pkgtui.LabelStyle.Render(timeAgo),
	)

	if selected {
		return pkgtui.SelectedStyle.Render("> " + content)
	}
	return "  " + content
}

func timeAgoString(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("Jan 2")
	}
}

// Focus implements View
func (v *KickoffView) Focus() tea.Cmd {
	v.focusInput = true
	v.input.Focus()
	return textinput.Blink
}

// Blur implements View
func (v *KickoffView) Blur() {
	v.input.Blur()
}

// Name implements View
func (v *KickoffView) Name() string {
	return "Kickoff"
}

// ShortHelp implements View
func (v *KickoffView) ShortHelp() string {
	return "enter create  tab switch"
}

// Commands implements CommandProvider
func (v *KickoffView) Commands() []tui.Command {
	return []tui.Command{
		{
			Name:        "New Project",
			Description: "Start a new project",
			Action: func() tea.Cmd {
				v.focusInput = true
				v.input.Focus()
				return textinput.Blink
			},
		},
	}
}
