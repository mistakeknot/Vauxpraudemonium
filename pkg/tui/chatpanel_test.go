package tui

import (
	"strings"
	"testing"
)

func TestChatPanelRendersSelector(t *testing.T) {
	panel := NewChatPanel()
	panel.SetSize(40, 20)
	selector := NewAgentSelector([]AgentOption{{Name: "codex"}, {Name: "claude"}})
	selector.Open = true
	panel.SetAgentSelector(selector)

	view := panel.View()
	if !strings.Contains(view, "codex") {
		t.Fatalf("expected selector in view")
	}
}
