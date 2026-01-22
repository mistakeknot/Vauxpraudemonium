package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type fakeKeyApprover struct {
	called bool
	taskID string
	branch string
}

func (f *fakeKeyApprover) Approve(taskID, branch string) error {
	f.called = true
	f.taskID = taskID
	f.branch = branch
	return nil
}

func TestApproveKeyCallsApprover(t *testing.T) {
	m := NewModel()
	m.ConfirmApprove = false
	m.Review.Queue = []string{"TAND-001"}
	m.Review.BranchLookup = func(taskID string) (string, error) {
		if taskID != "TAND-001" {
			t.Fatalf("unexpected task ID: %s", taskID)
		}
		return "feature/TAND-001", nil
	}
	fake := &fakeKeyApprover{}
	m.Review.Approver = fake

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if !fake.called {
		t.Fatal("expected approve call")
	}
	if fake.taskID != "TAND-001" {
		t.Fatalf("expected task ID, got %q", fake.taskID)
	}
	if fake.branch != "feature/TAND-001" {
		t.Fatalf("expected branch, got %q", fake.branch)
	}
}

func TestApproveKeyRefreshesReviewQueue(t *testing.T) {
	m := NewModel()
	m.ConfirmApprove = false
	m.Review.Queue = []string{"TAND-001"}
	m.Review.BranchLookup = func(taskID string) (string, error) {
		return "feature/TAND-001", nil
	}
	m.Review.Loader = func() ([]string, error) {
		return []string{}, nil
	}
	fake := &fakeKeyApprover{}
	m.Review.Approver = fake

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated := next.(Model)

	if len(updated.Review.Queue) != 0 {
		t.Fatalf("expected review queue to refresh, got %v", updated.Review.Queue)
	}
	if updated.StatusLevel != StatusInfo {
		t.Fatalf("expected status info, got %v", updated.StatusLevel)
	}
}

func TestApproveKeySetsErrorStatus(t *testing.T) {
	m := NewModel()
	m.ConfirmApprove = false
	m.Review.Queue = []string{"TAND-001"}
	m.Review.BranchLookup = func(taskID string) (string, error) {
		return "", errors.New("boom")
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updated := next.(Model)

	if updated.StatusLevel != StatusError {
		t.Fatalf("expected error status, got %v", updated.StatusLevel)
	}
	if !strings.Contains(updated.Status, "branch lookup failed") {
		t.Fatalf("expected branch lookup error, got %q", updated.Status)
	}
}

func TestApproveKeyUsesSelectedItem(t *testing.T) {
	m := NewModel()
	m.ConfirmApprove = false
	m.Review.Queue = []string{"T1", "T2"}
	m.Review.Selected = 1
	m.Review.BranchLookup = func(taskID string) (string, error) {
		return "feature/" + taskID, nil
	}
	m.Review.Loader = func() ([]string, error) {
		return []string{}, nil
	}
	fake := &fakeKeyApprover{}
	m.Review.Approver = fake

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !fake.called {
		t.Fatal("expected approve call")
	}
	if fake.taskID != "T2" {
		t.Fatalf("expected selected task, got %q", fake.taskID)
	}
	if fake.branch != "feature/T2" {
		t.Fatalf("expected selected branch, got %q", fake.branch)
	}
}

func TestApproveEnterRequiresConfirmation(t *testing.T) {
	m := NewModel()
	m.ConfirmApprove = true
	m.Review.Queue = []string{"T1"}
	m.Review.BranchLookup = func(taskID string) (string, error) {
		return "feature/" + taskID, nil
	}
	m.Review.Loader = func() ([]string, error) {
		return []string{}, nil
	}
	fake := &fakeKeyApprover{}
	m.Review.Approver = fake

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)

	if fake.called {
		t.Fatal("expected approve to be deferred")
	}
	if updated.Review.PendingApproveTask != "T1" {
		t.Fatalf("expected pending task, got %q", updated.Review.PendingApproveTask)
	}
}

func TestApproveConfirmationYRunsApprove(t *testing.T) {
	m := NewModel()
	m.ConfirmApprove = true
	m.Review.PendingApproveTask = "T1"
	m.Review.BranchLookup = func(taskID string) (string, error) {
		return "feature/" + taskID, nil
	}
	m.Review.Loader = func() ([]string, error) {
		return []string{}, nil
	}
	fake := &fakeKeyApprover{}
	m.Review.Approver = fake

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	updated := next.(Model)

	if !fake.called {
		t.Fatal("expected approve call on confirmation")
	}
	if updated.Review.PendingApproveTask != "" {
		t.Fatalf("expected pending cleared, got %q", updated.Review.PendingApproveTask)
	}
}

func TestApproveConfirmationNCancels(t *testing.T) {
	m := NewModel()
	m.ConfirmApprove = true
	m.Review.PendingApproveTask = "T1"

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	updated := next.(Model)

	if updated.Review.PendingApproveTask != "" {
		t.Fatalf("expected pending cleared, got %q", updated.Review.PendingApproveTask)
	}
	if updated.StatusLevel != StatusInfo {
		t.Fatalf("expected status info, got %v", updated.StatusLevel)
	}
}

func TestReviewStatePlumbsPendingApprove(t *testing.T) {
	m := NewModel()
	m.Review.PendingApproveTask = "TAND-001"
	if m.Review.PendingApproveTask != "TAND-001" {
		t.Fatalf("expected pending approve task")
	}
}
