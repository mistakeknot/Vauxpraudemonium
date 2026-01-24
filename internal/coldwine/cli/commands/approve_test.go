package commands

import (
	"strings"
	"testing"
)

func TestApproveCmd(t *testing.T) {
	if ApproveCmd().Use != "approve" {
		t.Fatal("expected approve")
	}
}

func TestApproveCmdRejectsInvalidTaskID(t *testing.T) {
	cmd := ApproveCmd()
	err := cmd.Args(cmd, []string{"bad/id", "branch"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid task id") {
		t.Fatalf("unexpected error: %v", err)
	}
}
