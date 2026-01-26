package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakeknot/autarch/internal/autarch/drafts"
	"github.com/mistakeknot/autarch/internal/pollard/research"
	"github.com/mistakeknot/autarch/internal/tui"
	"github.com/mistakeknot/autarch/internal/tui/components"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

// InterviewQuestion represents a single interview question with research integration.
type InterviewQuestion struct {
	ID          string   // Stable identifier
	TopicKey    string   // Maps to research topics (e.g., "platform", "storage")
	Title       string   // Short title for navigation
	Prompt      string   // Full question text
	Placeholder string   // Input placeholder
	MultiLine   bool     // Allow multi-line input
	Options     []string // For choice questions (empty = text input)
}

// InterviewView provides an enhanced interview flow with research integration.
type InterviewView struct {
	questions []InterviewQuestion
	current   int
	answers   map[string]string
	states    map[string]*drafts.QuestionState

	// Research integration
	coordinator     *research.Coordinator
	researchOverlay *ResearchOverlay
	teasers         map[string]*components.Teaser
	tradeoffs       map[string]*components.Tradeoff
	showResearch    bool

	// UI components
	input       textinput.Model
	optionIndex int
	width       int
	height      int
	focused     bool

	// Draft persistence
	draft      *drafts.Draft
	draftStore *drafts.Store

	// Callbacks
	onComplete func(answers map[string]string) tea.Cmd
}

// NewInterviewView creates a new interview view with research integration.
func NewInterviewView(questions []InterviewQuestion, coordinator *research.Coordinator) *InterviewView {
	ti := textinput.New()
	ti.CharLimit = 500
	ti.Width = 60

	v := &InterviewView{
		questions:       questions,
		current:         0,
		answers:         make(map[string]string),
		states:          make(map[string]*drafts.QuestionState),
		coordinator:     coordinator,
		teasers:         make(map[string]*components.Teaser),
		tradeoffs:       make(map[string]*components.Tradeoff),
		input:           ti,
		researchOverlay: NewResearchOverlay(coordinator),
	}

	// Initialize teasers for each question
	for _, q := range questions {
		v.teasers[q.ID] = components.NewTeaser(q.TopicKey, components.TeaserHeight)
		v.tradeoffs[q.ID] = components.NewTradeoff()
	}

	return v
}

// SetDraft sets the draft for persistence and restores state.
func (v *InterviewView) SetDraft(draft *drafts.Draft, store *drafts.Store) {
	v.draft = draft
	v.draftStore = store

	// Restore state from draft
	if draft.Interview != nil {
		v.current = draft.Interview.CurrentStep
		for k, val := range draft.Interview.Answers {
			if str, ok := val.(string); ok {
				v.answers[k] = str
			}
		}
		for _, qs := range draft.Interview.Questions {
			v.states[qs.QuestionID] = &qs
		}
	}

	v.loadCurrentInput()
}

// SetCompleteCallback sets the callback for when interview is complete.
func (v *InterviewView) SetCompleteCallback(cb func(answers map[string]string) tea.Cmd) {
	v.onComplete = cb
}

// Init implements View
func (v *InterviewView) Init() tea.Cmd {
	v.loadCurrentInput()
	return textinput.Blink
}

// currentQuestion returns the current interview question.
func (v *InterviewView) currentQuestion() *InterviewQuestion {
	if v.current >= 0 && v.current < len(v.questions) {
		return &v.questions[v.current]
	}
	return nil
}

// Update implements View
func (v *InterviewView) Update(msg tea.Msg) (tui.View, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle research overlay if visible
	if v.showResearch {
		var cmd tea.Cmd
		v.researchOverlay, cmd = v.researchOverlay.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Check for close
		if !v.researchOverlay.Visible() {
			v.showResearch = false
		}
		return v, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height - 4
		v.input.Width = min(60, v.width-10)
		v.researchOverlay.SetSize(v.width, v.height)
		for _, teaser := range v.teasers {
			teaser.SetSize(v.width - 4)
		}
		for _, tradeoff := range v.tradeoffs {
			tradeoff.SetSize(v.width - 4)
		}
		return v, nil

	// Research messages
	case research.HunterUpdateMsg, research.HunterCompletedMsg:
		v.updateResearchForCurrentQuestion()
		return v, nil

	case research.RunCompletedMsg:
		v.updateResearchForCurrentQuestion()
		return v, nil

	// Tradeoff selection
	case components.TradeoffSelectedMsg:
		q := v.currentQuestion()
		if q != nil {
			v.answers[q.ID] = msg.Option.Label
			v.input.SetValue(msg.Option.Label)
			v.markQuestionTouched(q.ID)
		}
		return v, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+r":
			// Toggle research overlay
			v.showResearch = true
			return v, v.researchOverlay.Show()

		case "enter":
			return v, v.handleEnter()

		case "up", "k":
			if v.isChoiceQuestion() && v.optionIndex > 0 {
				v.optionIndex--
			}
			return v, nil

		case "down", "j":
			if v.isChoiceQuestion() {
				q := v.currentQuestion()
				if q != nil && v.optionIndex < len(q.Options)-1 {
					v.optionIndex++
				}
			}
			return v, nil

		case "[":
			// Previous question
			if v.current > 0 {
				v.saveCurrent()
				v.current--
				v.loadCurrentInput()
			}
			return v, nil

		case "]":
			// Next question (without completing)
			if v.current < len(v.questions)-1 {
				v.saveCurrent()
				v.current++
				v.loadCurrentInput()
			}
			return v, nil

		case "1", "2", "3":
			// Handle tradeoff selection
			q := v.currentQuestion()
			if q != nil {
				tradeoff := v.tradeoffs[q.ID]
				if tradeoff != nil && tradeoff.HasOptions() {
					var cmd tea.Cmd
					tradeoff, cmd = tradeoff.Update(msg)
					v.tradeoffs[q.ID] = tradeoff
					if cmd != nil {
						cmds = append(cmds, cmd)
					}
					return v, tea.Batch(cmds...)
				}
			}

			// Also handle as option selection for choice questions
			if v.isChoiceQuestion() {
				idx := int(msg.String()[0] - '1')
				q := v.currentQuestion()
				if q != nil && idx >= 0 && idx < len(q.Options) {
					v.optionIndex = idx
					return v, v.handleEnter()
				}
			}
		}

		// Pass to text input if not a choice question
		if !v.isChoiceQuestion() {
			var cmd tea.Cmd
			v.input, cmd = v.input.Update(msg)
			v.markQuestionTouched(v.currentQuestion().ID)
			return v, cmd
		}
	}

	return v, tea.Batch(cmds...)
}

func (v *InterviewView) handleEnter() tea.Cmd {
	v.saveCurrent()

	// Advance to next question
	if v.current < len(v.questions)-1 {
		v.current++
		v.loadCurrentInput()
		v.maybeApplyResearchDefault()
		return nil
	}

	// Complete interview
	if v.onComplete != nil {
		return v.onComplete(v.answers)
	}
	return nil
}

func (v *InterviewView) saveCurrent() {
	q := v.currentQuestion()
	if q == nil {
		return
	}

	if v.isChoiceQuestion() {
		if v.optionIndex >= 0 && v.optionIndex < len(q.Options) {
			v.answers[q.ID] = q.Options[v.optionIndex]
		}
	} else {
		v.answers[q.ID] = v.input.Value()
	}

	// Persist to draft
	v.saveDraft()
}

func (v *InterviewView) loadCurrentInput() {
	q := v.currentQuestion()
	if q == nil {
		return
	}

	v.input.Placeholder = q.Placeholder
	if answer, ok := v.answers[q.ID]; ok {
		v.input.SetValue(answer)
	} else {
		v.input.SetValue("")
	}

	// Reset option index for choice questions
	if len(q.Options) > 0 {
		v.optionIndex = 0
		for i, opt := range q.Options {
			if opt == v.answers[q.ID] {
				v.optionIndex = i
				break
			}
		}
	}

	// Update research for this question
	v.updateResearchForCurrentQuestion()
}

func (v *InterviewView) isChoiceQuestion() bool {
	q := v.currentQuestion()
	return q != nil && len(q.Options) > 0
}

func (v *InterviewView) markQuestionTouched(questionID string) {
	if v.states[questionID] == nil {
		v.states[questionID] = &drafts.QuestionState{
			QuestionID: questionID,
		}
	}
	v.states[questionID].Touched = true
}

func (v *InterviewView) isQuestionTouched(questionID string) bool {
	if state, ok := v.states[questionID]; ok {
		return state.Touched
	}
	return false
}

func (v *InterviewView) updateResearchForCurrentQuestion() {
	q := v.currentQuestion()
	if q == nil || v.coordinator == nil {
		return
	}

	run := v.coordinator.GetActiveRun()
	if run == nil {
		return
	}

	// Update teaser
	teaser := v.teasers[q.ID]
	if teaser != nil {
		updates := run.GetUpdatesForTopic(q.TopicKey)
		teaser.UpdateFromRun(run.RunID, updates)
	}

	// Update tradeoffs from findings
	tradeoff := v.tradeoffs[q.ID]
	if tradeoff != nil {
		var allFindings []research.Finding
		for _, update := range run.GetUpdatesForTopic(q.TopicKey) {
			allFindings = append(allFindings, update.Findings...)
		}
		options := components.CreateFromFindings(allFindings, q.TopicKey)
		tradeoff.SetOptions(options)
	}
}

func (v *InterviewView) maybeApplyResearchDefault() {
	q := v.currentQuestion()
	if q == nil {
		return
	}

	// Don't apply if already touched or default already applied
	if v.isQuestionTouched(q.ID) {
		return
	}
	if state, ok := v.states[q.ID]; ok && state.DefaultApplied {
		return
	}

	// Check for research suggestions
	tradeoff := v.tradeoffs[q.ID]
	if tradeoff == nil || !tradeoff.HasOptions() {
		return
	}

	// Apply first suggestion as default
	if opt := tradeoff.GetOption(0); opt != nil {
		v.input.SetValue(opt.Label)
		v.answers[q.ID] = opt.Label

		// Mark default as applied
		if v.states[q.ID] == nil {
			v.states[q.ID] = &drafts.QuestionState{
				QuestionID: q.ID,
			}
		}
		v.states[q.ID].DefaultApplied = true
	}
}

func (v *InterviewView) saveDraft() {
	if v.draft == nil || v.draftStore == nil {
		return
	}

	// Update draft interview state
	if v.draft.Interview == nil {
		v.draft.Interview = &drafts.InterviewState{
			Answers: make(map[string]any),
		}
	}

	v.draft.Interview.CurrentStep = v.current
	v.draft.Interview.TotalSteps = len(v.questions)
	for k, val := range v.answers {
		v.draft.Interview.Answers[k] = val
	}

	// Update question states
	v.draft.Interview.Questions = nil
	for id, state := range v.states {
		v.draft.Interview.Questions = append(v.draft.Interview.Questions, drafts.QuestionState{
			QuestionID:     id,
			Touched:        state.Touched,
			DefaultApplied: state.DefaultApplied,
		})
	}

	// Save asynchronously
	go v.draftStore.Save(v.draft)
}

// View implements View
func (v *InterviewView) View() string {
	// Check for research overlay
	if v.showResearch {
		return v.renderWithResearchOverlay()
	}

	return v.renderMainView()
}

func (v *InterviewView) renderMainView() string {
	q := v.currentQuestion()
	if q == nil {
		return pkgtui.LabelStyle.Render("No questions configured")
	}

	var sections []string

	// Progress bar
	progress := v.renderProgress()
	sections = append(sections, progress)
	sections = append(sections, "")

	// Question title
	titleStyle := lipgloss.NewStyle().
		Foreground(pkgtui.ColorPrimary).
		Bold(true)
	sections = append(sections, titleStyle.Render(q.Title))

	// Question prompt
	promptStyle := pkgtui.LabelStyle
	sections = append(sections, promptStyle.Render(q.Prompt))
	sections = append(sections, "")

	// Research teaser (fixed height)
	teaser := v.teasers[q.ID]
	if teaser != nil {
		sections = append(sections, teaser.View())
		sections = append(sections, "")
	}

	// Tradeoff suggestions
	tradeoff := v.tradeoffs[q.ID]
	if tradeoff != nil && tradeoff.HasOptions() {
		if v.isQuestionTouched(q.ID) {
			// Show FYI version for already-answered questions
			sections = append(sections, tradeoff.ViewFYI())
		} else {
			// Show full tradeoff view
			sections = append(sections, tradeoff.View())
		}
		sections = append(sections, "")
	}

	// Input or options
	if len(q.Options) > 0 {
		sections = append(sections, v.renderOptions(q))
	} else {
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(pkgtui.ColorPrimary).
			Padding(0, 1)
		sections = append(sections, inputStyle.Render(v.input.View()))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (v *InterviewView) renderWithResearchOverlay() string {
	// Wide terminals: side-by-side
	if v.width >= 120 {
		mainWidth := v.width * 2 / 3
		overlayWidth := v.width - mainWidth

		mainView := v.renderMainView()
		overlayView := v.researchOverlay.View()

		mainStyle := lipgloss.NewStyle().Width(mainWidth)
		overlayStyle := lipgloss.NewStyle().Width(overlayWidth)

		return lipgloss.JoinHorizontal(lipgloss.Top,
			mainStyle.Render(mainView),
			overlayStyle.Render(overlayView),
		)
	}

	// Narrow terminals: overlay
	return v.researchOverlay.View()
}

func (v *InterviewView) renderProgress() string {
	total := len(v.questions)
	current := v.current + 1

	// Progress bar
	barWidth := min(40, v.width-20)
	filled := (barWidth * current) / total
	empty := barWidth - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	barStyle := lipgloss.NewStyle().Foreground(pkgtui.ColorPrimary)

	return fmt.Sprintf("Step %d/%d  %s", current, total, barStyle.Render(bar))
}

func (v *InterviewView) renderOptions(q *InterviewQuestion) string {
	var lines []string
	for i, opt := range q.Options {
		var line string
		if i == v.optionIndex {
			line = pkgtui.SelectedStyle.Render(fmt.Sprintf("> %d) %s", i+1, opt))
		} else {
			line = pkgtui.UnselectedStyle.Render(fmt.Sprintf("  %d) %s", i+1, opt))
		}
		lines = append(lines, line)
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// Focus implements View
func (v *InterviewView) Focus() tea.Cmd {
	v.focused = true
	v.input.Focus()
	return textinput.Blink
}

// Blur implements View
func (v *InterviewView) Blur() {
	v.focused = false
	v.input.Blur()
	v.saveDraft()
}

// Name implements View
func (v *InterviewView) Name() string {
	return "Interview"
}

// ShortHelp implements View
func (v *InterviewView) ShortHelp() string {
	return "enter next  [/] nav  ctrl+r research"
}

// GetAnswers returns all collected answers.
func (v *InterviewView) GetAnswers() map[string]string {
	return v.answers
}

// DefaultInterviewQuestions returns the standard onboarding questions.
func DefaultInterviewQuestions() []InterviewQuestion {
	return []InterviewQuestion{
		{
			ID:          "vision",
			TopicKey:    "vision",
			Title:       "Vision",
			Prompt:      "What is your project vision? Describe what you want to build.",
			Placeholder: "A tool that...",
			MultiLine:   true,
		},
		{
			ID:          "users",
			TopicKey:    "users",
			Title:       "Users",
			Prompt:      "Who are the primary users of this project?",
			Placeholder: "Developers who...",
			MultiLine:   false,
		},
		{
			ID:          "problem",
			TopicKey:    "problem",
			Title:       "Problem",
			Prompt:      "What problem are you solving?",
			Placeholder: "Currently, users have to...",
			MultiLine:   true,
		},
		{
			ID:          "platform",
			TopicKey:    "platform",
			Title:       "Platform",
			Prompt:      "What platform(s) will this run on?",
			Placeholder: "",
			Options:     []string{"Web", "CLI", "Desktop", "Mobile", "API/Backend"},
		},
		{
			ID:          "language",
			TopicKey:    "language",
			Title:       "Language",
			Prompt:      "What programming language(s) will you use?",
			Placeholder: "",
			Options:     []string{"Go", "TypeScript", "Python", "Rust", "Other"},
		},
		{
			ID:          "requirements",
			TopicKey:    "requirements",
			Title:       "Requirements",
			Prompt:      "List the key requirements (one per line or comma-separated).",
			Placeholder: "User authentication, Real-time updates, ...",
			MultiLine:   true,
		},
	}
}
