package tui

import (
	"fmt"
	"strings"

	"github.com/mistakeknot/autarch/pkg/signals"
)

// SignalPanel renders aggregated signals from all tools.
type SignalPanel struct {
	signals []signals.Signal
}

// NewSignalPanel creates a signal aggregation panel.
func NewSignalPanel() *SignalPanel {
	return &SignalPanel{}
}

// SetSignals updates the panel with current signals.
func (p *SignalPanel) SetSignals(sigs []signals.Signal) {
	p.signals = sigs
}

// ActiveCount returns the number of undismissed signals.
func (p *SignalPanel) ActiveCount() int {
	count := 0
	for _, s := range p.signals {
		if s.IsActive() {
			count++
		}
	}
	return count
}

// Render returns a string representation of the signal panel.
func (p *SignalPanel) Render(width int) string {
	active := p.activeSignals()
	if len(active) == 0 {
		return "  No active signals"
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("  ⚡ %d active signal(s)", len(active)))
	lines = append(lines, "")

	for _, s := range active {
		icon := severityIcon(s.Severity)
		line := fmt.Sprintf("  %s [%s] %s", icon, s.Source, s.Title)
		if len(line) > width {
			line = line[:width-3] + "..."
		}
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (p *SignalPanel) activeSignals() []signals.Signal {
	var result []signals.Signal
	for _, s := range p.signals {
		if s.IsActive() {
			result = append(result, s)
		}
	}
	return result
}

func severityIcon(sev signals.Severity) string {
	switch sev {
	case signals.SeverityCritical:
		return "🔴"
	case signals.SeverityWarning:
		return "🟡"
	default:
		return "🔵"
	}
}
