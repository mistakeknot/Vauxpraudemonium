package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakeknot/autarch/internal/pollard/research"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

// ResearchOverlay displays full Pollard research results.
// Can be shown as modal overlay or side panel depending on terminal width.
type ResearchOverlay struct {
	coordinator *research.Coordinator
	runID       string // Track which run we're showing

	// UI state
	visible      bool
	width        int
	height       int
	scrollOffset int
	selected     int
	expanded     map[int]bool // Which findings are expanded
	searchMode   bool
	searchInput  textinput.Model

	// Cached data
	hunterStatuses map[string]research.HunterStatus
	findings       []research.Finding
	filteredIdx    []int // Indices into findings after search filter
}

// NewResearchOverlay creates a new research overlay.
func NewResearchOverlay(coordinator *research.Coordinator) *ResearchOverlay {
	si := textinput.New()
	si.Placeholder = "Search findings..."
	si.CharLimit = 100

	return &ResearchOverlay{
		coordinator: coordinator,
		expanded:    make(map[int]bool),
		searchInput: si,
	}
}

// Show makes the overlay visible and refreshes data.
func (o *ResearchOverlay) Show() tea.Cmd {
	o.visible = true
	o.scrollOffset = 0
	o.selected = 0
	o.searchMode = false
	o.searchInput.SetValue("")
	return o.refresh()
}

// Hide makes the overlay invisible.
func (o *ResearchOverlay) Hide() {
	o.visible = false
	o.searchMode = false
	o.searchInput.Blur()
}

// Toggle switches visibility.
func (o *ResearchOverlay) Toggle() tea.Cmd {
	if o.visible {
		o.Hide()
		return nil
	}
	return o.Show()
}

// Visible returns whether the overlay is visible.
func (o *ResearchOverlay) Visible() bool {
	return o.visible
}

// SetSize updates the overlay dimensions.
func (o *ResearchOverlay) SetSize(width, height int) {
	o.width = width
	o.height = height
	o.searchInput.Width = width - 10
}

// refresh updates cached data from the coordinator.
func (o *ResearchOverlay) refresh() tea.Cmd {
	return func() tea.Msg {
		return researchRefreshMsg{}
	}
}

type researchRefreshMsg struct{}

// Update handles messages for the overlay.
func (o *ResearchOverlay) Update(msg tea.Msg) (*ResearchOverlay, tea.Cmd) {
	if !o.visible {
		return o, nil
	}

	switch msg := msg.(type) {
	case researchRefreshMsg:
		o.loadFromCoordinator()
		return o, nil

	case research.HunterStartedMsg, research.HunterCompletedMsg,
		research.HunterErrorMsg, research.HunterUpdateMsg,
		research.RunStartedMsg, research.RunCompletedMsg:
		// Refresh on any research update
		o.loadFromCoordinator()
		return o, nil

	case tea.KeyMsg:
		// Handle search mode
		if o.searchMode {
			switch msg.String() {
			case "esc":
				o.searchMode = false
				o.searchInput.Blur()
				o.searchInput.SetValue("")
				o.applyFilter("")
				return o, nil
			case "enter":
				o.searchMode = false
				o.searchInput.Blur()
				return o, nil
			default:
				var cmd tea.Cmd
				o.searchInput, cmd = o.searchInput.Update(msg)
				o.applyFilter(o.searchInput.Value())
				return o, cmd
			}
		}

		switch msg.String() {
		case "ctrl+r", "esc":
			o.Hide()
			return o, nil

		case "j", "down":
			maxIdx := len(o.filteredIdx) - 1
			if o.selected < maxIdx {
				o.selected++
				o.ensureVisible()
			}
			return o, nil

		case "k", "up":
			if o.selected > 0 {
				o.selected--
				o.ensureVisible()
			}
			return o, nil

		case "enter":
			if len(o.filteredIdx) > 0 && o.selected < len(o.filteredIdx) {
				idx := o.filteredIdx[o.selected]
				o.expanded[idx] = !o.expanded[idx]
			}
			return o, nil

		case "/":
			o.searchMode = true
			o.searchInput.Focus()
			return o, textinput.Blink

		case "r":
			return o, o.refresh()
		}
	}

	return o, nil
}

func (o *ResearchOverlay) loadFromCoordinator() {
	if o.coordinator == nil {
		return
	}

	run := o.coordinator.GetActiveRun()
	if run == nil {
		o.hunterStatuses = nil
		o.findings = nil
		o.filteredIdx = nil
		return
	}

	o.runID = run.RunID
	o.hunterStatuses = run.GetHunterStatuses()

	// Collect all findings from all updates
	o.findings = nil
	for _, update := range run.GetAllUpdates() {
		o.findings = append(o.findings, update.Findings...)
	}

	// Initialize filter to show all
	o.applyFilter(o.searchInput.Value())
}

func (o *ResearchOverlay) applyFilter(query string) {
	o.filteredIdx = nil
	query = strings.ToLower(query)

	for i, f := range o.findings {
		if query == "" {
			o.filteredIdx = append(o.filteredIdx, i)
			continue
		}
		// Search in title and summary
		if strings.Contains(strings.ToLower(f.Title), query) ||
			strings.Contains(strings.ToLower(f.Summary), query) {
			o.filteredIdx = append(o.filteredIdx, i)
		}
	}

	// Reset selection if out of bounds
	if o.selected >= len(o.filteredIdx) {
		o.selected = max(0, len(o.filteredIdx)-1)
	}
}

func (o *ResearchOverlay) ensureVisible() {
	visibleHeight := o.height - 8 // Account for header, footer, etc.
	if o.selected < o.scrollOffset {
		o.scrollOffset = o.selected
	} else if o.selected >= o.scrollOffset+visibleHeight {
		o.scrollOffset = o.selected - visibleHeight + 1
	}
}

// View renders the overlay.
func (o *ResearchOverlay) View() string {
	if !o.visible {
		return ""
	}

	// Determine if we're a side panel or full overlay
	isSidePanel := o.width >= 120

	var panelWidth int
	if isSidePanel {
		panelWidth = o.width / 3
	} else {
		panelWidth = min(o.width-4, 80)
	}

	var sections []string

	// Header
	headerStyle := lipgloss.NewStyle().
		Background(pkgtui.ColorBgLight).
		Foreground(pkgtui.ColorPrimary).
		Bold(true).
		Padding(0, 1).
		Width(panelWidth)

	header := headerStyle.Render("RESEARCH (Ctrl+R/Esc to return)")
	sections = append(sections, header)

	// Hunter status
	sections = append(sections, o.renderHunterStatus(panelWidth))

	// Search bar if in search mode
	if o.searchMode {
		searchStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(pkgtui.ColorPrimary).
			Width(panelWidth - 4)
		sections = append(sections, searchStyle.Render(o.searchInput.View()))
	}

	// Findings
	if len(o.filteredIdx) == 0 {
		if len(o.findings) == 0 {
			sections = append(sections, pkgtui.LabelStyle.Render("No findings yet..."))
		} else {
			sections = append(sections, pkgtui.LabelStyle.Render("No matches for search"))
		}
	} else {
		findingsView := o.renderFindings(panelWidth)
		sections = append(sections, findingsView)
	}

	// Footer help
	footerStyle := pkgtui.LabelStyle
	footer := footerStyle.Render("j/k scroll  Enter expand  / search  r refresh")
	sections = append(sections, footer)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Wrap in panel
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(pkgtui.ColorPrimary).
		Background(pkgtui.ColorBg).
		Padding(1).
		Width(panelWidth).
		Height(o.height - 2)

	return panelStyle.Render(content)
}

func (o *ResearchOverlay) renderHunterStatus(width int) string {
	if len(o.hunterStatuses) == 0 {
		return pkgtui.LabelStyle.Render("No hunters registered")
	}

	var parts []string
	for name, status := range o.hunterStatuses {
		var icon string
		var style lipgloss.Style

		switch status.Status {
		case research.StatusRunning:
			icon = "↻"
			style = lipgloss.NewStyle().Foreground(pkgtui.ColorWarning)
		case research.StatusComplete:
			icon = "✓"
			style = lipgloss.NewStyle().Foreground(pkgtui.ColorSuccess)
		case research.StatusError:
			icon = "✗"
			style = lipgloss.NewStyle().Foreground(pkgtui.ColorError)
		case research.StatusPending:
			icon = "○"
			style = lipgloss.NewStyle().Foreground(pkgtui.ColorMuted)
		default:
			icon = "?"
			style = lipgloss.NewStyle().Foreground(pkgtui.ColorMuted)
		}

		line := style.Render(fmt.Sprintf("%s %s", icon, name))
		if status.Findings > 0 {
			line += pkgtui.LabelStyle.Render(fmt.Sprintf(" (%d)", status.Findings))
		}
		parts = append(parts, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (o *ResearchOverlay) renderFindings(width int) string {
	visibleHeight := o.height - 12 // Account for header, status, footer
	if visibleHeight < 3 {
		visibleHeight = 3
	}

	var lines []string
	endIdx := min(o.scrollOffset+visibleHeight, len(o.filteredIdx))

	for i := o.scrollOffset; i < endIdx; i++ {
		if i >= len(o.filteredIdx) {
			break
		}
		findingIdx := o.filteredIdx[i]
		if findingIdx >= len(o.findings) {
			continue
		}
		finding := o.findings[findingIdx]
		isSelected := i == o.selected
		isExpanded := o.expanded[findingIdx]

		line := o.renderFinding(finding, width-4, isSelected, isExpanded)
		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (o *ResearchOverlay) renderFinding(f research.Finding, width int, selected, expanded bool) string {
	// Source badge
	sourceStyle := lipgloss.NewStyle().
		Background(pkgtui.ColorBgLight).
		Foreground(pkgtui.ColorFgDim).
		Padding(0, 1)

	sourceBadge := sourceStyle.Render(f.SourceType)

	// Title
	titleStyle := lipgloss.NewStyle().Foreground(pkgtui.ColorFg)
	if selected {
		titleStyle = titleStyle.Bold(true).Foreground(pkgtui.ColorPrimary)
	}

	// Truncate title if needed
	title := f.Title
	maxTitleLen := width - lipgloss.Width(sourceBadge) - 4
	if len(title) > maxTitleLen {
		title = title[:maxTitleLen-3] + "..."
	}

	// Selector indicator
	selector := "  "
	if selected {
		selector = "> "
	}

	header := fmt.Sprintf("%s%s  %s", selector, titleStyle.Render(title), sourceBadge)

	if !expanded {
		return header
	}

	// Expanded view
	var sections []string
	sections = append(sections, header)

	// Summary
	summaryStyle := pkgtui.LabelStyle.Width(width - 4).MarginLeft(4)
	sections = append(sections, summaryStyle.Render(f.Summary))

	// Source link
	if f.Source != "" {
		sourceStyle := pkgtui.LabelStyle.Foreground(pkgtui.ColorSecondary).MarginLeft(4)
		sections = append(sections, sourceStyle.Render(f.Source))
	}

	// Relevance
	relevanceStyle := pkgtui.LabelStyle.MarginLeft(4)
	relevanceBar := renderRelevanceBar(f.Relevance)
	sections = append(sections, relevanceStyle.Render(fmt.Sprintf("Relevance: %s %.0f%%", relevanceBar, f.Relevance*100)))

	sections = append(sections, "") // Spacer

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func renderRelevanceBar(score float64) string {
	filled := int(score * 10)
	empty := 10 - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	var style lipgloss.Style
	switch {
	case score >= 0.7:
		style = lipgloss.NewStyle().Foreground(pkgtui.ColorSuccess)
	case score >= 0.4:
		style = lipgloss.NewStyle().Foreground(pkgtui.ColorWarning)
	default:
		style = lipgloss.NewStyle().Foreground(pkgtui.ColorMuted)
	}

	return style.Render(bar)
}

// ViewAsPanel returns the view formatted as a side panel (right side).
func (o *ResearchOverlay) ViewAsPanel() string {
	return o.View()
}
