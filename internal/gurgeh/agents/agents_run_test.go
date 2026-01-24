package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunReturnsStdout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := Run(Profile{Command: "cat", Args: []string{}}, path)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "hello" {
		t.Fatalf("expected stdout")
	}
}
