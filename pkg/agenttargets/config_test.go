package agenttargets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRegistriesWithCompat(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "agents.toml")
	projectDir := filepath.Join(dir, "proj")
	if err := os.MkdirAll(filepath.Join(projectDir, ".praude"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(globalPath, []byte("[targets.codex]\ncommand=\"codex\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".praude", "config.toml"), []byte("[agents.custom]\ncommand=\"/bin/custom\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	global, project, err := Load(globalPath, projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := global.Targets["codex"]; !ok {
		t.Fatalf("expected global codex")
	}
	if _, ok := project.Targets["custom"]; !ok {
		t.Fatalf("expected project custom")
	}
}
