package agenttargets

import "testing"

func TestMergeDetectedPrefersConfig(t *testing.T) {
	detected := Registry{Targets: map[string]Target{
		"codex":  {Name: "codex", Type: TargetDetected, Command: "codex"},
		"claude": {Name: "claude", Type: TargetDetected, Command: "claude"},
	}}
	global := Registry{Targets: map[string]Target{
		"codex": {Name: "codex", Type: TargetCommand, Command: "/bin/codex"},
	}}
	project := Registry{Targets: map[string]Target{
		"claude": {Name: "claude", Type: TargetCommand, Command: "/bin/claude"},
	}}

	merged := MergeDetected(detected, global, project)

	if merged.Targets["codex"].Command != "/bin/codex" {
		t.Fatalf("expected codex from global, got %q", merged.Targets["codex"].Command)
	}
	if merged.Targets["claude"].Command != "/bin/claude" {
		t.Fatalf("expected claude from project, got %q", merged.Targets["claude"].Command)
	}
}
