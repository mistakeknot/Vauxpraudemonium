package tui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakeknot/autarch/internal/coldwine/epics"
	"github.com/mistakeknot/autarch/internal/coldwine/tasks"
	"github.com/mistakeknot/autarch/internal/pollard/research"
	"github.com/mistakeknot/autarch/pkg/autarch"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

// AppMode represents the current mode of the app
type AppMode int

const (
	ModeOnboarding AppMode = iota
	ModeDashboard
)

// UnifiedApp is the main application that handles both onboarding and dashboard modes
type UnifiedApp struct {
	client *autarch.Client
	mode   AppMode

	// Onboarding state
	onboardingState   OnboardingState
	currentView       View
	projectID         string
	projectName       string
	projectDesc       string
	generatedEpics    []epics.EpicProposal
	generatedTasks    []tasks.TaskProposal
	researchCoord     *research.Coordinator
	ctx               context.Context
	cancel            context.CancelFunc

	// Dashboard state
	tabs      *TabBar
	dashViews []View
	palette   *Palette

	// UI state
	width  int
	height int
	err    error

	// View factories (injected from main.go)
	createKickoffView     func() View
	createEpicReviewView  func([]epics.EpicProposal) View
	createTaskReviewView  func([]tasks.TaskProposal) View
	createTaskDetailView  func(tasks.TaskProposal, *research.Coordinator) View
	createDashboardViews  func(*autarch.Client) []View
}

// NewUnifiedApp creates a new unified application
func NewUnifiedApp(client *autarch.Client) *UnifiedApp {
	ctx, cancel := context.WithCancel(context.Background())

	tabNames := []string{"Bigend", "Gurgeh", "Coldwine", "Pollard"}
	app := &UnifiedApp{
		client:          client,
		mode:            ModeOnboarding,
		onboardingState: OnboardingKickoff,
		tabs:            NewTabBar(tabNames),
		palette:         NewPalette(),
		researchCoord:   research.NewCoordinator(nil),
		ctx:             ctx,
		cancel:          cancel,
	}

	return app
}

// SetViewFactories sets the factory functions for creating views
func (a *UnifiedApp) SetViewFactories(
	kickoff func() View,
	epicReview func([]epics.EpicProposal) View,
	taskReview func([]tasks.TaskProposal) View,
	taskDetail func(tasks.TaskProposal, *research.Coordinator) View,
	dashViews func(*autarch.Client) []View,
) {
	a.createKickoffView = kickoff
	a.createEpicReviewView = epicReview
	a.createTaskReviewView = taskReview
	a.createTaskDetailView = taskDetail
	a.createDashboardViews = dashViews
}

// Init implements tea.Model
func (a *UnifiedApp) Init() tea.Cmd {
	// Start with kickoff view
	if a.createKickoffView != nil {
		a.currentView = a.createKickoffView()
		return a.currentView.Init()
	}
	return nil
}

// Update implements tea.Model
func (a *UnifiedApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.tabs.SetWidth(msg.Width)
		a.palette.SetSize(msg.Width, msg.Height)

		// Pass to current view
		if a.currentView != nil {
			var cmd tea.Cmd
			a.currentView, cmd = a.currentView.Update(msg)
			return a, cmd
		}
		return a, nil

	case tea.KeyMsg:
		// Handle palette if visible
		if a.palette.Visible() {
			var cmd tea.Cmd
			a.palette, cmd = a.palette.Update(msg)
			return a, cmd
		}

		switch msg.String() {
		case "ctrl+c":
			a.cancel()
			return a, tea.Quit

		case "ctrl+p":
			if a.mode == ModeDashboard {
				return a, a.palette.Show()
			}
		}

		// In dashboard mode, handle tab switching
		if a.mode == ModeDashboard {
			switch msg.String() {
			case "q":
				return a, tea.Quit
			case "1", "2", "3", "4":
				idx := int(msg.String()[0] - '1')
				return a, a.switchDashboardTab(idx)
			case "tab":
				return a, a.switchDashboardTab((a.tabs.Active() + 1) % len(a.dashViews))
			}
		}

	// Handle view transition messages
	case ProjectCreatedMsg:
		return a, a.handleProjectCreated(msg)

	case EpicsGeneratedMsg:
		return a, a.handleEpicsGenerated(msg)

	case EpicsAcceptedMsg:
		return a, a.handleEpicsAccepted(msg)

	case TasksGeneratedMsg:
		return a, a.handleTasksGenerated(msg)

	case TasksAcceptedMsg:
		return a, a.handleTasksAccepted(msg)

	case NavigateToTaskDetailMsg:
		return a, a.showTaskDetail(msg.Task)

	case NavigateBackMsg:
		return a, a.navigateBack()

	case OnboardingCompleteMsg:
		return a, a.enterDashboard()
	}

	// Pass to current view
	if a.currentView != nil {
		var cmd tea.Cmd
		a.currentView, cmd = a.currentView.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a *UnifiedApp) handleProjectCreated(msg ProjectCreatedMsg) tea.Cmd {
	a.projectID = msg.ProjectID
	a.projectName = msg.ProjectName
	a.projectDesc = msg.Description

	// Transition to epic generation (skip interview for now, generate from description)
	a.onboardingState = OnboardingEpicReview
	return a.generateEpicsFromDescription()
}

func (a *UnifiedApp) generateEpicsFromDescription() tea.Cmd {
	return func() tea.Msg {
		// Create a simple epic generator based on the description
		gen := epics.NewGenerator()
		proposals := gen.GenerateFromDescription(a.projectDesc)
		return EpicsGeneratedMsg{Epics: proposals}
	}
}

func (a *UnifiedApp) handleEpicsGenerated(msg EpicsGeneratedMsg) tea.Cmd {
	a.generatedEpics = msg.Epics

	// Show epic review view
	if a.createEpicReviewView != nil {
		a.currentView = a.createEpicReviewView(msg.Epics)
		return tea.Batch(
			a.currentView.Init(),
			a.sendWindowSize(),
		)
	}
	return nil
}

func (a *UnifiedApp) handleEpicsAccepted(msg EpicsAcceptedMsg) tea.Cmd {
	a.generatedEpics = msg.Epics
	a.onboardingState = OnboardingTaskReview

	// Generate tasks from epics
	return a.generateTasksFromEpics()
}

func (a *UnifiedApp) generateTasksFromEpics() tea.Cmd {
	return func() tea.Msg {
		gen := tasks.NewGenerator()
		taskList, err := gen.GenerateFromEpics(a.generatedEpics)
		if err != nil {
			// Return empty list on error for now
			return TasksGeneratedMsg{Tasks: nil}
		}
		return TasksGeneratedMsg{Tasks: taskList}
	}
}

func (a *UnifiedApp) handleTasksGenerated(msg TasksGeneratedMsg) tea.Cmd {
	a.generatedTasks = msg.Tasks

	// Show task review view
	if a.createTaskReviewView != nil {
		a.currentView = a.createTaskReviewView(msg.Tasks)
		return tea.Batch(
			a.currentView.Init(),
			a.sendWindowSize(),
		)
	}
	return nil
}

func (a *UnifiedApp) handleTasksAccepted(msg TasksAcceptedMsg) tea.Cmd {
	a.generatedTasks = msg.Tasks
	a.onboardingState = OnboardingComplete

	// Transition to dashboard
	return func() tea.Msg {
		return OnboardingCompleteMsg{
			ProjectID:   a.projectID,
			ProjectName: a.projectName,
		}
	}
}

func (a *UnifiedApp) showTaskDetail(task tasks.TaskProposal) tea.Cmd {
	if a.createTaskDetailView != nil {
		a.currentView = a.createTaskDetailView(task, a.researchCoord)
		return tea.Batch(
			a.currentView.Init(),
			a.sendWindowSize(),
		)
	}
	return nil
}

func (a *UnifiedApp) navigateBack() tea.Cmd {
	// Return to appropriate view based on state
	switch a.onboardingState {
	case OnboardingTaskReview:
		if a.createTaskReviewView != nil {
			a.currentView = a.createTaskReviewView(a.generatedTasks)
			return tea.Batch(a.currentView.Init(), a.sendWindowSize())
		}
	case OnboardingComplete:
		// In dashboard mode, go back to Bigend
		if len(a.dashViews) > 0 {
			a.currentView = a.dashViews[0]
			return a.currentView.Focus()
		}
	}
	return nil
}

func (a *UnifiedApp) enterDashboard() tea.Cmd {
	a.mode = ModeDashboard

	// Create dashboard views
	if a.createDashboardViews != nil {
		a.dashViews = a.createDashboardViews(a.client)
		if len(a.dashViews) > 0 {
			a.currentView = a.dashViews[0]

			// Initialize all views
			var cmds []tea.Cmd
			for _, v := range a.dashViews {
				cmds = append(cmds, v.Init())
			}
			cmds = append(cmds, a.currentView.Focus())
			cmds = append(cmds, a.sendWindowSize())
			return tea.Batch(cmds...)
		}
	}
	return nil
}

func (a *UnifiedApp) switchDashboardTab(idx int) tea.Cmd {
	if idx < 0 || idx >= len(a.dashViews) {
		return nil
	}

	oldActive := a.tabs.Active()
	if oldActive == idx {
		return nil
	}

	// Blur old view
	if oldActive < len(a.dashViews) {
		a.dashViews[oldActive].Blur()
	}

	a.tabs.SetActive(idx)
	a.currentView = a.dashViews[idx]
	return a.currentView.Focus()
}

func (a *UnifiedApp) sendWindowSize() tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{Width: a.width, Height: a.height}
	}
}

// View implements tea.Model
func (a *UnifiedApp) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// In dashboard mode, show tabs
	if a.mode == ModeDashboard {
		b.WriteString(a.tabs.View())
		b.WriteString("\n")
	} else {
		// Onboarding header
		headerStyle := lipgloss.NewStyle().
			Foreground(pkgtui.ColorPrimary).
			Bold(true).
			MarginBottom(1)
		b.WriteString(headerStyle.Render(a.onboardingHeader()))
		b.WriteString("\n\n")
	}

	// Content area
	contentHeight := a.height - 4
	if a.mode == ModeDashboard {
		contentHeight -= 2 // Account for tabs
	}

	var content string
	if a.currentView != nil {
		content = a.currentView.View()
	}

	// Ensure content fills the space
	contentLines := strings.Split(content, "\n")
	for len(contentLines) < contentHeight {
		contentLines = append(contentLines, "")
	}
	if len(contentLines) > contentHeight {
		contentLines = contentLines[:contentHeight]
	}
	b.WriteString(strings.Join(contentLines, "\n"))
	b.WriteString("\n")

	// Footer
	b.WriteString(a.renderFooter())

	// Overlay palette if visible
	if a.palette.Visible() {
		return a.overlay(b.String(), a.palette.View())
	}

	return b.String()
}

func (a *UnifiedApp) onboardingHeader() string {
	switch a.onboardingState {
	case OnboardingKickoff:
		return "New Project"
	case OnboardingInterview:
		return "Project Setup"
	case OnboardingSpecSummary:
		return "Review Spec"
	case OnboardingEpicReview:
		return "Review Epics"
	case OnboardingTaskReview:
		return "Review Tasks"
	default:
		return "Autarch"
	}
}

func (a *UnifiedApp) renderFooter() string {
	help := ""
	if a.currentView != nil {
		help = a.currentView.ShortHelp()
	}

	if a.mode == ModeDashboard {
		help += "  1-4 tabs  ctrl+p palette  q quit"
	} else {
		help += "  ctrl+c cancel"
	}

	style := lipgloss.NewStyle().
		Foreground(pkgtui.ColorMuted).
		Width(a.width)

	return style.Render(help)
}

func (a *UnifiedApp) overlay(base, overlay string) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	startRow := (a.height - len(overlayLines)) / 4
	startCol := (a.width - lipgloss.Width(overlayLines[0])) / 2

	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}

	for i, line := range overlayLines {
		row := startRow + i
		if row >= len(baseLines) {
			break
		}
		baseLines[row] = insertAt(baseLines[row], startCol, line)
	}

	return strings.Join(baseLines, "\n")
}

func insertAt(base string, col int, overlay string) string {
	baseRunes := []rune(base)
	for len(baseRunes) < col {
		baseRunes = append(baseRunes, ' ')
	}

	overlayWidth := lipgloss.Width(overlay)
	var result strings.Builder

	if col > 0 && col < len(baseRunes) {
		result.WriteString(string(baseRunes[:col]))
	}
	result.WriteString(overlay)

	end := col + overlayWidth
	if end < len(baseRunes) {
		result.WriteString(string(baseRunes[end:]))
	}

	return result.String()
}

// RunUnified starts the unified TUI application
func RunUnified(client *autarch.Client, app *UnifiedApp) error {
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
