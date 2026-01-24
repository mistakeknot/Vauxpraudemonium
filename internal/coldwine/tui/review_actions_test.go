package tui

import "testing"

func TestSubmitFeedbackClearsInput(t *testing.T) {
	m := NewModel()
	m.ViewMode = ViewReview
	m.Review.InputMode = ReviewInputFeedback
	m.Review.Input = "Looks good"
	m.Review.Detail = ReviewDetail{TaskID: "T1"}
	m.Review.ActionWriter = func(taskID, text string) error { return nil }
	m.handleReviewSubmit()
	if m.Review.InputMode != ReviewInputNone || m.Review.Input != "" {
		t.Fatalf("expected input cleared")
	}
}

func TestSubmitRejectRequeues(t *testing.T) {
	m := NewModel()
	m.ViewMode = ViewReview
	m.Review.InputMode = ReviewInputFeedback
	m.Review.PendingReject = true
	m.Review.Input = "Needs work"
	m.Review.Detail = ReviewDetail{TaskID: "T1"}
	m.Review.ActionWriter = func(taskID, text string) error { return nil }
	m.Review.Rejecter = func(taskID string) error { return nil }
	m.handleReviewSubmit()
	if m.Review.PendingReject {
		t.Fatalf("expected reject cleared")
	}
}

func TestReviewStateHoldsDetailAndDiff(t *testing.T) {
	m := NewModel()
	m.Review.Detail.TaskID = "TAND-002"
	m.Review.Diff.Files = []string{"a.txt"}
	if m.Review.Detail.TaskID != "TAND-002" || len(m.Review.Diff.Files) != 1 {
		t.Fatalf("expected review detail/diff to be stored")
	}
}
