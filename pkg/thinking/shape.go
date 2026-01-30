package thinking

// Shape represents a thinking strategy applied before content generation.
// Each shape forces a different kind of deliberation preamble.
type Shape int

const (
	ShapeNone           Shape = iota // No preamble, preserves current behavior
	ShapeDeductive                   // State quality criteria first, then execute
	ShapeInductive                   // Show examples, infer pattern (few-shot)
	ShapeAbductive                   // Show examples, extract principles, then apply
	ShapeContrapositive              // Enumerate failures, then avoid them all
	ShapeDSL                         // Build domain vocabulary first, then fill it
)

// String returns the display name for a shape.
func (s Shape) String() string {
	names := []string{
		"None",
		"Deductive",
		"Inductive",
		"Abductive",
		"Contrapositive",
		"DSL",
	}
	if s >= 0 && int(s) < len(names) {
		return names[s]
	}
	return "Unknown"
}

// Example holds a few-shot example for inductive/abductive shapes.
type Example struct {
	Label   string // Short label (e.g. "Developer persona")
	Content string // The example text
}

// Preamble holds a rendered thinking preamble ready to prepend to a draft.
type Preamble struct {
	Shape    Shape
	Phase    string
	Rendered string // Final rendered preamble text
}
