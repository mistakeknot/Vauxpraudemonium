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

	b, _ = b.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})

	if b.IsNavigating() {
		t.Fatalf("expected back to exit navigation, selected=%d", b.selected)
	}
}
