package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mistakeknot/vauxpraudemonium/internal/praude/agents"
)

func TestNewKeyStartsInterviewForNewSpec(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".praude", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	_ = os.Chdir(root)

	m := NewModel()
	m = pressKey(m, "n")
	if m.mode != "interview" {
		t.Fatalf("expected interview mode")
	}
	entries, _ := os.ReadDir(filepath.Join(root, ".praude", "specs"))
	if len(entries) != 1 {
		t.Fatalf("expected new spec file")
	}
}

func TestInterviewEnterIteratesDraft(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".praude", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".praude", "briefs"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := `validation_mode = "soft"

[agents.codex]
command = "codex"
args = []
`
	if err := os.WriteFile(filepath.Join(root, ".praude", "config.toml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	_ = os.Chdir(root)

	oldRun := runAgent
	runAgent = func(p agents.Profile, briefPath string) ([]byte, error) {
		return []byte("Drafted vision"), nil
	}
	defer func() { runAgent = oldRun }()

	m := NewModel()
	m = pressKey(m, "n")
	m = pressKey(m, "2")
	m = pressKey(m, "1")
	m = typeText(m, "Initial vision")
	m = pressKey(m, "enter")
	out := m.View()
	if !strings.Contains(out, "Drafted vision") {
		t.Fatalf("expected draft in view")
	}
}
