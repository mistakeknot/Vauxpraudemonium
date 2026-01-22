package tui

import (
	"strings"
	"testing"
)

func TestReviewViewShowsMVPWarning(t *testing.T) {
	m := NewModel()
	m.ViewMode = ViewReview
	m.Review.Detail = ReviewDetail{Alignment: "out"}
	out := m.View()
	if !strings.Contains(out, "MVP SCOPE WARNING") {
		t.Fatalf("expected mvp warning")
	}
}

func TestMVPExplainClearsInput(t *testing.T) {
	m := NewModel()
	m.ViewMode = ViewReview
	m.Review.InputMode = ReviewInputFeedback
	m.Review.Input = "Reason"
	m.Review.Detail = ReviewDetail{TaskID: "T1", Alignment: "out"}
	m.Review.MVPExplainWriter = func(taskID, text string) error { return nil }
	m.handleMVPExplainSubmit()
	if m.Review.InputMode != ReviewInputNone {
		t.Fatalf("expected input cleared")
	}
}

func TestMVPAcceptUpdatesScope(t *testing.T) {
	m := NewModel()
	m.ViewMode = ViewReview
	m.Review.Detail = ReviewDetail{TaskID: "T1", Alignment: "out"}
	m.Review.MVPAcceptor = func(taskID string) error { return nil }
	m.handleMVPAccept()
	if m.Status == "" {
		t.Fatalf("expected status")
	}
}

func TestMVPExplainUsesWriterWhenPending(t *testing.T) {
	called := false
	m := NewModel()
	m.ViewMode = ViewReview
	m.Review.Detail = ReviewDetail{TaskID: "T1", Alignment: "out"}
	m.Review.InputMode = ReviewInputFeedback
	m.Review.Input = "Reason"
	m.Review.MVPExplainPending = true
	m.Review.MVPExplainWriter = func(taskID, text string) error {
		called = true
		return nil
	}
	m.handleReviewSubmit()
	if !called {
		t.Fatalf("expected explain writer to be called")
	}
}

func TestMVPRevertCallsReverter(t *testing.T) {
	called := false
	var gotPath string
	m := NewModel()
	m.ViewMode = ViewReview
	m.Review.Detail = ReviewDetail{
		TaskID:    "T1",
		Alignment: "out",
		Files:     []ReviewFile{{Path: "a.txt"}, {Path: "b.txt"}},
	}
	m.Review.MVPReverter = func(taskID, path string) error {
		called = true
		gotPath = path
		return nil
	}
	m.handleMVPRevertStart()
	m.handleMVPRevertConfirm()
	if !called || gotPath != "a.txt" {
		t.Fatalf("expected revert on a.txt")
	}
}
