package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCreatesGurgehLayout(t *testing.T) {
	root := t.TempDir()
	if err := Init(root); err != nil {
		t.Fatal(err)
	}
	mustDir := []string{
		filepath.Join(root, ".gurgeh"),
		filepath.Join(root, ".gurgeh", "specs"),
		filepath.Join(root, ".gurgeh", "research"),
		filepath.Join(root, ".gurgeh", "suggestions"),
		filepath.Join(root, ".gurgeh", "briefs"),
	}
	for _, dir := range mustDir {
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			t.Fatalf("missing dir %s", dir)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".gurgeh", "config.toml")); err != nil {
		t.Fatalf("expected config.toml")
	}
}
