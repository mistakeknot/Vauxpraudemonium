package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
	"github.com/sahilm/fuzzy"
)

// Palette is a command palette with fuzzy search
type Palette struct {
	input    textinput.Model
	commands []Command
	matches  []fuzzy.Match
	selected int
	width    int
	height   int
	visible  bool
}

// NewPalette creates a new command palette
func NewPalette() *Palette {
	input := textinput.New()
	input.Placeholder = "Type a command..."
	input.Prompt = "> "
	input.CharLimit = 64

	return &Palette{
		input: input,
	}
}

// SetCommands sets the available commands
func (p *Palette) SetCommands(cmds []Command) {
	p.commands = cmds
	p.updateMatches()
}

// SetSize sets the palette dimensions
func (p *Palette) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.input.Width = width - 6
}

// Show shows the palette and focuses input
func (p *Palette) Show() tea.Cmd {
	p.visible = true
	p.input.Reset()
	p.selected = 0
	p.updateMatches()
	return p.input.Focus()
}

// Hide hides the palette
func (p *Palette) Hide() {
	p.visible = false
}

// Visible returns whether the palette is visible
func (p *Palette) Visible() bool {
	return p.visible
}

// Selected returns the currently selected command, if any
func (p *Palette) Selected() *Command {
	if len(p.matches) == 0 {
		return nil
	}
	if p.selected >= len(p.matches) {
		return nil
	}
	idx := p.matches[p.selected].Index
	if idx >= len(p.commands) {
		return nil
	}
	return &p.commands[idx]
}

// Update handles input
func (p *Palette) Update(msg tea.Msg) (*Palette, tea.Cmd) {
	if !p.visible {
		return p, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			p.Hide()
			return p, nil

		case "enter":
			if cmd := p.Selected(); cmd != nil {
				p.Hide()
				return p, cmd.Action()
			}
			return p, nil

		case "up", "ctrl+p":
			if p.selected > 0 {
				p.selected--
			}
			return p, nil

		case "down", "ctrl+n":
			if p.selected < len(p.matches)-1 {
				p.selected++
			}
			return p, nil

		case "ctrl+c":
			p.Hide()
			return p, nil
		}
	}

	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)

	// Update matches on input change
	p.updateMatches()
	// Reset selection on input change
	p.selected = 0

	return p, cmd
}

func (p *Palette) updateMatches() {
	query := strings.TrimSpace(p.input.Value())
	if query == "" {
		// Show all commands when query is empty
		p.matches = make([]fuzzy.Match, len(p.commands))
		for i := range p.commands {
			p.matches[i] = fuzzy.Match{Index: i}
		}
		return
	}

	// Build searchable list
	names := make([]string, len(p.commands))
	for i, cmd := range p.commands {
		names[i] = cmd.Name
	}

	p.matches = fuzzy.Find(query, names)
}

// View renders the palette
func (p *Palette) View() string {
	if !p.visible {
		return ""
	}

	// Calculate dimensions
	width := p.width
	if width > 60 {
		width = 60
	}

	var b strings.Builder

	// Title
	title := pkgtui.TitleStyle.Render("Command Palette")
	b.WriteString(title + "\n")

	// Input
	b.WriteString(p.input.View() + "\n")

	// Separator
	b.WriteString(strings.Repeat("─", width-4) + "\n")

	// Results
	maxResults := 8
	if p.height > 0 {
		maxResults = min(maxResults, p.height-6)
	}

	for i, match := range p.matches {
		if i >= maxResults {
			break
		}

		cmd := p.commands[match.Index]
		name := cmd.Name
		desc := cmd.Description

		if i == p.selected {
			name = pkgtui.SelectedStyle.Render(name)
		} else {
			name = pkgtui.UnselectedStyle.Render(name)
		}

		desc = pkgtui.LabelStyle.Render(desc)

		line := "  " + name
		if desc != "" {
			line += "  " + desc
		}
		b.WriteString(line + "\n")
	}

	if len(p.matches) == 0 {
		b.WriteString(pkgtui.LabelStyle.Render("  No matching commands\n"))
	}

	// Style the container
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(pkgtui.ColorPrimary).
		Padding(1, 2).
		Width(width)

	return style.Render(b.String())
}
