package thinking

// PhaseDefault maps a phase name to its default thinking shape.
var PhaseDefault = map[string]Shape{
	"Vision":                 ShapeDeductive,
	"Problem":                ShapeContrapositive,
	"Users":                  ShapeInductive,
	"Features + Goals":       ShapeDSL,
	"Requirements":           ShapeDSL,
	"Scope + Assumptions":    ShapeContrapositive,
	"Critical User Journeys": ShapeAbductive,
	"Acceptance Criteria":    ShapeDeductive,
}

// preambleTemplates holds the template text for each phase.
// Templates use simple {{.Phase}} and {{.Context}} placeholders.
var preambleTemplates = map[string]string{
	// Deductive: state criteria, then produce
	"Vision": `Before writing the vision statement, state the criteria a strong vision must satisfy:
- Describes a future state, not current state
- Focuses on user outcomes, not features
- Is ambitious but achievable
- Fits in one paragraph

Now produce a vision statement that meets ALL of the above criteria.`,

	// Contrapositive: enumerate failures, then avoid
	"Problem": `Before writing the problem statement, list the ways problem statements typically fail:
- Too vague: no specific audience named
- Unmeasurable: no way to know if it's solved
- Solution-shaped: describes a feature, not a pain
- No urgency: doesn't convey cost of inaction

Now produce a problem statement that avoids ALL of the above failures.`,

	// Inductive: examples, then pattern
	"Users": `Here are examples of strong user personas:

**Example 1 — Developer persona:**
Primary: Full-stack developers (3-7 years experience) building SaaS products.
Demographics: 25-35, comfortable with CLI tools, values speed over polish.
Workflow: Currently uses spreadsheets + Notion to track specs, losing context switching.

**Example 2 — Product manager persona:**
Primary: Technical PMs at Series A-C startups (50-200 employees).
Demographics: 28-40, reads code but doesn't write it daily, manages 2-4 engineers.
Workflow: Writes PRDs in Google Docs, manually tracks feature coverage gaps.

Following this pattern, produce user personas for the current project.`,

	// DSL: define schema, then populate
	"Features + Goals": `Define features using this schema before writing content:
- Feature ID (F-001, F-002, ...)
- Hypothesis: "If we build [feature], then [metric] will [change] by [amount] within [timeframe]"
- Success metric: quantitative measure
- Priority: P0 (must-have) / P1 (should-have) / P2 (nice-to-have)

Populate the schema for each feature, then write the narrative.`,

	// DSL: Given/When/Then schema
	"Requirements": `Each requirement must follow this schema:
- Requirement ID (REQ-001, REQ-002, ...)
- Format: Given [precondition], When [action], Then [expected outcome]
- Each MUST have at least one measurable constraint (latency, accuracy, count, etc.)

Produce requirements that strictly follow this format.`,

	// Contrapositive: scope creep indicators
	"Scope + Assumptions": `Before defining scope, list scope creep indicators to avoid:
- "And also..." additions that aren't core to the problem
- Features that serve a different user segment than the primary
- Optimizations before the happy path works
- Integrations that add complexity without validating the core hypothesis

Now produce scope boundaries that trigger NONE of the above indicators.`,

	// Abductive: extract principles from examples
	"Critical User Journeys": `Consider these example user journeys:

**Journey A — File upload:**
1. User drags file onto the page → instant visual feedback (< 100ms)
2. Progress bar shows upload status → user can continue other work
3. Upload completes → toast notification, file appears in list without refresh

**Journey B — Search:**
1. User types in search box → results appear after 2 keystrokes (< 200ms)
2. Results highlight matching terms → user scans visually
3. User clicks result → lands directly at the relevant section

Extract the UX principles these journeys share (responsiveness, progressive disclosure, zero-reload).
Apply those principles to produce user journeys for the current project.`,

	// Deductive: testability criteria
	"Acceptance Criteria": `Before writing acceptance criteria, state what makes criteria testable:
- Each criterion has exactly one expected outcome
- The outcome is observable (visible state change, measurable metric, or API response)
- Pass/fail is unambiguous — no "should be reasonable" or "performs well"
- Edge cases are explicit, not implied

Now write acceptance criteria that meet ALL testability standards above.`,
}
