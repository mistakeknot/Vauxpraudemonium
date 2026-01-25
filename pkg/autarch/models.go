// Package autarch provides client types for the Intermute domain API.
package autarch

import "time"

// SpecStatus represents the status of a specification
type SpecStatus string

const (
	SpecStatusDraft     SpecStatus = "draft"
	SpecStatusResearch  SpecStatus = "research"
	SpecStatusValidated SpecStatus = "validated"
	SpecStatusArchived  SpecStatus = "archived"
)

// Spec represents a product specification (PRD)
type Spec struct {
	ID        string     `json:"id"`
	Project   string     `json:"project"`
	Title     string     `json:"title"`
	Vision    string     `json:"vision,omitempty"`
	Users     string     `json:"users,omitempty"`
	Problem   string     `json:"problem,omitempty"`
	Status    SpecStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// EpicStatus represents the status of an epic
type EpicStatus string

const (
	EpicStatusOpen       EpicStatus = "open"
	EpicStatusInProgress EpicStatus = "in_progress"
	EpicStatusDone       EpicStatus = "done"
)

// Epic represents a large feature or initiative
type Epic struct {
	ID          string     `json:"id"`
	Project     string     `json:"project"`
	SpecID      string     `json:"spec_id,omitempty"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Status      EpicStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// StoryStatus represents the status of a story
type StoryStatus string

const (
	StoryStatusTodo       StoryStatus = "todo"
	StoryStatusInProgress StoryStatus = "in_progress"
	StoryStatusReview     StoryStatus = "review"
	StoryStatusDone       StoryStatus = "done"
)

// Story represents a user story within an epic
type Story struct {
	ID                 string      `json:"id"`
	Project            string      `json:"project"`
	EpicID             string      `json:"epic_id"`
	Title              string      `json:"title"`
	AcceptanceCriteria []string    `json:"acceptance_criteria,omitempty"`
	Status             StoryStatus `json:"status"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending TaskStatus = "pending"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusBlocked TaskStatus = "blocked"
	TaskStatusDone    TaskStatus = "done"
)

// Task represents an execution unit assigned to an agent
type Task struct {
	ID        string     `json:"id"`
	Project   string     `json:"project"`
	StoryID   string     `json:"story_id,omitempty"`
	Title     string     `json:"title"`
	Agent     string     `json:"agent,omitempty"`
	SessionID string     `json:"session_id,omitempty"`
	Status    TaskStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Insight represents a research insight from Pollard
type Insight struct {
	ID        string    `json:"id"`
	Project   string    `json:"project"`
	SpecID    string    `json:"spec_id,omitempty"`
	Source    string    `json:"source"`
	Category  string    `json:"category"`
	Title     string    `json:"title"`
	Body      string    `json:"body,omitempty"`
	URL       string    `json:"url,omitempty"`
	Score     float64   `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}

// SessionStatus represents the status of an agent session
type SessionStatus string

const (
	SessionStatusRunning SessionStatus = "running"
	SessionStatusIdle    SessionStatus = "idle"
	SessionStatusError   SessionStatus = "error"
)

// Session represents an agent session (tmux session)
type Session struct {
	ID        string        `json:"id"`
	Project   string        `json:"project"`
	Name      string        `json:"name"`
	Agent     string        `json:"agent"`
	TaskID    string        `json:"task_id,omitempty"`
	Status    SessionStatus `json:"status"`
	StartedAt time.Time     `json:"started_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// DomainEvent represents a real-time event from the server
type DomainEvent struct {
	Type     string `json:"type"`
	Project  string `json:"project"`
	EntityID string `json:"entity_id"`
	Data     any    `json:"data,omitempty"`
}
