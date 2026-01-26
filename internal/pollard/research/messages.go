package research

import tea "github.com/charmbracelet/bubbletea"

// Msg types for communicating research updates via Bubble Tea.
// All messages include RunID to prevent stale updates from affecting UI.

// RunStartedMsg is sent when a new research run begins.
type RunStartedMsg struct {
	RunID     string
	ProjectID string
	Hunters   []string // Names of hunters that will run
}

// HunterStartedMsg is sent when an individual hunter begins.
type HunterStartedMsg struct {
	RunID      string
	HunterName string
}

// HunterUpdateMsg is sent when a hunter produces findings.
type HunterUpdateMsg struct {
	RunID      string
	HunterName string
	TopicKey   string
	Findings   []Finding
}

// HunterCompletedMsg is sent when a hunter finishes successfully.
type HunterCompletedMsg struct {
	RunID        string
	HunterName   string
	FindingCount int
}

// HunterErrorMsg is sent when a hunter fails.
type HunterErrorMsg struct {
	RunID      string
	HunterName string
	Error      error
}

// RunCompletedMsg is sent when all hunters have finished.
type RunCompletedMsg struct {
	RunID         string
	TotalFindings int
	Duration      string
}

// RunCancelledMsg is sent when a run is cancelled.
type RunCancelledMsg struct {
	RunID  string
	Reason string
}

// CreateRunStartedCmd creates a command that sends a RunStartedMsg.
func CreateRunStartedCmd(run *Run, hunters []string) tea.Cmd {
	return func() tea.Msg {
		return RunStartedMsg{
			RunID:     run.RunID,
			ProjectID: run.ProjectID,
			Hunters:   hunters,
		}
	}
}

// CreateHunterStartedCmd creates a command that sends a HunterStartedMsg.
func CreateHunterStartedCmd(runID, hunterName string) tea.Cmd {
	return func() tea.Msg {
		return HunterStartedMsg{
			RunID:      runID,
			HunterName: hunterName,
		}
	}
}

// CreateHunterUpdateCmd creates a command that sends a HunterUpdateMsg.
func CreateHunterUpdateCmd(runID, hunterName, topicKey string, findings []Finding) tea.Cmd {
	return func() tea.Msg {
		return HunterUpdateMsg{
			RunID:      runID,
			HunterName: hunterName,
			TopicKey:   topicKey,
			Findings:   findings,
		}
	}
}

// CreateHunterCompletedCmd creates a command that sends a HunterCompletedMsg.
func CreateHunterCompletedCmd(runID, hunterName string, findingCount int) tea.Cmd {
	return func() tea.Msg {
		return HunterCompletedMsg{
			RunID:        runID,
			HunterName:   hunterName,
			FindingCount: findingCount,
		}
	}
}

// CreateHunterErrorCmd creates a command that sends a HunterErrorMsg.
func CreateHunterErrorCmd(runID, hunterName string, err error) tea.Cmd {
	return func() tea.Msg {
		return HunterErrorMsg{
			RunID:      runID,
			HunterName: hunterName,
			Error:      err,
		}
	}
}

// CreateRunCompletedCmd creates a command that sends a RunCompletedMsg.
func CreateRunCompletedCmd(run *Run) tea.Cmd {
	return func() tea.Msg {
		return RunCompletedMsg{
			RunID:         run.RunID,
			TotalFindings: run.TotalFindings(),
			Duration:      run.Duration().String(),
		}
	}
}

// CreateRunCancelledCmd creates a command that sends a RunCancelledMsg.
func CreateRunCancelledCmd(runID, reason string) tea.Cmd {
	return func() tea.Msg {
		return RunCancelledMsg{
			RunID:  runID,
			Reason: reason,
		}
	}
}
