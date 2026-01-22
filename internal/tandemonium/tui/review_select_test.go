package tui

import "testing"

func TestClampSelectionAfterRefresh(t *testing.T) {
	m := NewModel()
	m.Review.Queue = []string{"T1", "T2"}
	m.Review.Selected = 1
	m.Review.Queue = []string{"T1"}
	m.ClampReviewSelection()
	if m.Review.Selected != 0 {
		t.Fatalf("expected selection 0, got %d", m.Review.Selected)
	}
}
