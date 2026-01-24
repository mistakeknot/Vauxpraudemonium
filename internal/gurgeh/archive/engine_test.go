package archive

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mistakeknot/vauxpraudemonium/internal/gurgeh/project"
)

func TestArchiveMovesSpec(t *testing.T) {
	root := t.TempDir()
	_ = project.Init(root)
	src := filepath.Join(project.SpecsDir(root), "PRD-001.yaml")
	if err := os.WriteFile(src, []byte("id: \"PRD-001\"\nstatus: \"draft\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Archive(root, "PRD-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.To) == 0 {
		t.Fatalf("expected move paths")
	}
	if _, err := os.Stat(filepath.Join(project.ArchivedSpecsDir(root), "PRD-001.yaml")); err != nil {
		t.Fatalf("expected archived spec")
	}
}
