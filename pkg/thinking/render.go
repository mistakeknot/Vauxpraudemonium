package thinking

import "strings"

// RenderPreamble produces a thinking preamble for the given shape and phase.
// If shape is ShapeNone or the phase has no template, returns an empty string.
func RenderPreamble(shape Shape, phase string) string {
	if shape == ShapeNone {
		return ""
	}
	tmpl, ok := preambleTemplates[phase]
	if !ok {
		return ""
	}
	return tmpl
}

// RenderForPhase uses the default shape for a phase.
// Returns empty string if the phase has no default shape.
func RenderForPhase(phase string) string {
	shape, ok := PhaseDefault[phase]
	if !ok {
		return ""
	}
	return RenderPreamble(shape, phase)
}

// FormatPreamble wraps a preamble in a delimiter block for inclusion in drafts.
// Returns empty string if preamble is empty.
func FormatPreamble(preamble string) string {
	if strings.TrimSpace(preamble) == "" {
		return ""
	}
	return "---\n**Thinking Preamble:**\n\n" + preamble + "\n\n---\n\n"
}
