package views

import (
	"strings"
	"testing"

	"github.com/mistakeknot/autarch/internal/tui"
)

func TestKickoffDocPanelIncludesAutarchCopy(t *testing.T) {
	v := NewKickoffView()
	v.docPanel.SetSize(80, 20)

	view := v.docPanel.View()
	expected := "Autarch is a platform for a suite of agentic tools"

	if !strings.Contains(view, expected) {
		t.Fatalf("expected doc panel to include kickoff copy, got %q", view)
	}
}

func TestKickoffScanResultUpdatesDocPanel(t *testing.T) {
	v := NewKickoffView()
	v.docPanel.SetSize(80, 30)
	v.chatPanel.SetValue("before")

	_, _ = v.Update(tui.CodebaseScanResultMsg{
		Description:  "Scanned description",
		Vision:       "Vision text",
		Users:        "Users text",
		Problem:      "Problem text",
		Platform:     "Web",
		Language:     "Go",
		Requirements: []string{"Req one", "Req two"},
	})

	if v.chatPanel.Value() != "before" {
		t.Fatalf("expected chat composer unchanged")
	}

	view := v.docPanel.View()
	if !strings.Contains(view, "Scanned description") {
		t.Fatalf("expected doc panel to include scan description")
	}
	if !strings.Contains(view, "Vision text") {
		t.Fatalf("expected doc panel to include scan vision")
	}
}
