package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mistakeknot/autarch/internal/tui"
)

func TestKickoffSeedsChatHistory(t *testing.T) {
	v := NewKickoffView()
	msgs := v.ChatMessagesForTest()
	if len(msgs) == 0 {
		t.Fatal("expected seeded chat messages")
	}
	if msgs[0].Role != "system" {
		t.Fatalf("expected system role, got %q", msgs[0].Role)
	}
	if !strings.Contains(msgs[0].Content, "What do you want to build") {
		t.Fatalf("expected prompt message, got %q", msgs[0].Content)
	}
}

func TestKickoffScanPreparingMessageRoutesToChat(t *testing.T) {
	v := NewKickoffView()
	v.loading = true
	v.scanning = true
	v.loadingMsg = "Scanning codebase..."

	_, _ = v.Update(tui.ScanProgressMsg{Step: "Preparing", Details: "Building analysis prompt..."})

	if v.loadingMsg == "Building analysis prompt..." {
		t.Fatalf("expected preparing detail not to render in main view")
	}

	msgs := v.ChatMessagesForTest()
	found := false
	for _, msg := range msgs {
		if strings.Contains(msg.Content, "Building analysis prompt...") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected preparing detail in chat messages")
	}
}

func TestKickoffAcceptsVisionStepAndAdvances(t *testing.T) {
	v := NewKickoffView()
	v.scanResult = &tui.CodebaseScanResultMsg{Vision: "Vision text"}
	v.SetScanStepForTest(tui.OnboardingScanVision)

	_, _ = v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if v.ScanStepForTest() != tui.OnboardingScanProblem {
		t.Fatalf("expected step advance to problem")
	}
}

func TestKickoffAcceptTriggersResuggest(t *testing.T) {
	v := NewKickoffView()
	v.scanResult = &tui.CodebaseScanResultMsg{Vision: "Vision text"}
	v.scanPath = "/tmp/project"
	v.SetScanStepForTest(tui.OnboardingScanVision)

	called := false
	v.SetScanCodebaseCallback(func(path string) tea.Cmd {
		if path != "/tmp/project" {
			t.Fatalf("expected resuggest path %q, got %q", "/tmp/project", path)
		}
		called = true
		return nil
	})

	_, _ = v.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !called {
		t.Fatalf("expected resuggest callback to fire")
	}
}
