package agenttargets

import "testing"

func TestResolveUsesProjectInProjectContext(t *testing.T) {
	global := Registry{Targets: map[string]Target{
		"custom": {Name: "custom", Type: TargetCommand, Command: "/bin/global"},
	}}
	project := Registry{Targets: map[string]Target{
		"custom": {Name: "custom", Type: TargetCommand, Command: "/bin/project"},
	}}
	r := NewResolver(global, project)
	got, err := r.Resolve(ProjectContext, "custom")
	if err != nil {
		t.Fatal(err)
	}
	if got.Command != "/bin/project" {
		t.Fatalf("expected project override")
	}
}

func TestResolveFallsBackToDetected(t *testing.T) {
	r := NewResolver(Registry{}, Registry{})
	got, err := r.Resolve(GlobalContext, "codex")
	if err != nil {
		t.Fatal(err)
	}
	if got.Command != "codex" {
		t.Fatalf("expected detected fallback")
	}
}
