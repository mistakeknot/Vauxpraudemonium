package tui

import (
	"testing"

	"github.com/mistakeknot/autarch/pkg/agenttargets"
)

func TestBuildAgentOptionsDedupePrefersConfig(t *testing.T) {
	options := buildAgentOptionsFromRegistry(
		agenttargets.Registry{Targets: map[string]agenttargets.Target{"codex": {Name: "codex", Command: "/bin/codex"}}},
	)
	if len(options) != 1 || options[0].Name != "codex" {
		t.Fatalf("expected codex option, got %v", options)
	}
}
