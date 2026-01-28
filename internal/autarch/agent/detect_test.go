package agent

import "testing"

func TestDetectAgentByNamePrefersName(t *testing.T) {
	agent, err := DetectAgentByName("codex", func(name string) (string, error) { return "/bin/codex", nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent.Type != TypeCodex {
		t.Fatalf("expected codex type, got %v", agent.Type)
	}
}
