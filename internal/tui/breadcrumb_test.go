package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestBreadcrumbBackExitsNavigation(t *testing.T) {
	b := NewBreadcrumb()
	b.SetCurrent(OnboardingInterview)
	b.StartNavigation()

	if !b.IsNavigating() {
		t.Fatal("expected breadcrumb to be navigating")
	}

	b, _ = b.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if b.IsNavigating() {
		t.Fatalf("expected back to exit navigation, selected=%d", b.selected)
	}
}

func TestBreadcrumbIncludesScanSteps(t *testing.T) {
	b := NewBreadcrumb()
	labels := b.LabelsForTest()
	want := []string{"Project", "Vision", "Problem", "Users"}
	for _, w := range want {
		found := false
		for _, label := range labels {
			if label == w {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected breadcrumb to include %q", w)
		}
	}
}
