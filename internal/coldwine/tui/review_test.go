package tui

import "testing"

func TestModelHasReviewQueue(t *testing.T) {
	m := NewModel()
	if m.Review.Queue == nil {
		t.Fatal("expected review queue")
	}
}
