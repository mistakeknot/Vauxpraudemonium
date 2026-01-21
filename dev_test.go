package main

import (
	"os"
	"testing"
)

func TestDevScriptExistsAndExecutable(t *testing.T) {
	info, err := os.Stat("dev")
	if err != nil {
		t.Fatalf("expected dev script: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("expected dev script to be executable")
	}
}
