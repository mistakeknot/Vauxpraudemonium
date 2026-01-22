package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mistakeknot/vauxpraudemonium/internal/praude/project"
)

func TestArchiveKeyMovesSpec(t *testing.T) {
	root := t.TempDir()
	_ = project.Init(root)
	src := filepath.Join(project.SpecsDir(root), "PRD-001.yaml")
	_ = os.WriteFile(src, []byte("id: \"PRD-001\"\nsummary: \"S\"\n"), 0o644)
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	_ = os.Chdir(root)
	m := NewModel()
	m = pressKey(m, "a")
	m = pressKey(m, "enter")
	if _, err := os.Stat(filepath.Join(project.ArchivedSpecsDir(root), "PRD-001.yaml")); err != nil {
		t.Fatalf("expected archived spec")
	}
}
