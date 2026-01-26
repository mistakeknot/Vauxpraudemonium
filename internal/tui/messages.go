package tui

import (
	"github.com/mistakeknot/autarch/internal/coldwine/epics"
	"github.com/mistakeknot/autarch/internal/coldwine/tasks"
	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
)

// View transition messages

// ProjectCreatedMsg is sent when a new project is created from kickoff
type ProjectCreatedMsg struct {
	ProjectID   string
	ProjectName string
	Description string
}

// SpecCompletedMsg is sent when the interview/spec is complete
type SpecCompletedMsg struct {
	Spec *specs.PRD
}

// EpicsGeneratedMsg is sent when epics are auto-generated
type EpicsGeneratedMsg struct {
	Epics []epics.EpicProposal
}

// EpicsAcceptedMsg is sent when user accepts the epics
type EpicsAcceptedMsg struct {
	Epics []epics.EpicProposal
}

// TasksGeneratedMsg is sent when tasks are generated from epics
type TasksGeneratedMsg struct {
	Tasks []tasks.TaskProposal
}

// TasksAcceptedMsg is sent when user accepts the tasks
type TasksAcceptedMsg struct {
	Tasks []tasks.TaskProposal
}

// OnboardingCompleteMsg signals onboarding is done
type OnboardingCompleteMsg struct {
	ProjectID   string
	ProjectName string
}

// NavigateToTaskDetailMsg requests navigation to task detail view
type NavigateToTaskDetailMsg struct {
	Task tasks.TaskProposal
}

// NavigateBackMsg requests navigation back to previous view
type NavigateBackMsg struct{}

// StartAgentMsg requests starting an agent for a task
type StartAgentMsg struct {
	Task      tasks.TaskProposal
	Agent     string
	Worktree  bool
}
