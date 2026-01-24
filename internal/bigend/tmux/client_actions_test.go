package tmux

import "testing"

type fakeRunner struct {
	calls [][]string
}

func (f *fakeRunner) Run(name string, args ...string) (string, string, error) {
	f.calls = append(f.calls, append([]string{name}, args...))
	return "", "", nil
}

func TestClientNewSessionCommand(t *testing.T) {
	fr := &fakeRunner{}
	c := NewClientWithRunner(fr)
	err := c.NewSession("claude-demo", "/root/projects/demo", []string{"claude"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fr.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fr.calls))
	}
	want := []string{c.tmuxPath, "new-session", "-d", "-s", "claude-demo", "-c", "/root/projects/demo", "claude"}
	got := fr.calls[0]
	if len(got) != len(want) {
		t.Fatalf("unexpected arg count: %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestClientRenameSessionCommand(t *testing.T) {
	fr := &fakeRunner{}
	c := NewClientWithRunner(fr)
	if err := c.RenameSession("old", "new"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fr.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fr.calls))
	}
	want := []string{c.tmuxPath, "rename-session", "-t", "old", "new"}
	got := fr.calls[0]
	if len(got) != len(want) {
		t.Fatalf("unexpected arg count: %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestClientKillSessionCommand(t *testing.T) {
	fr := &fakeRunner{}
	c := NewClientWithRunner(fr)
	if err := c.KillSession("dead"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fr.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(fr.calls))
	}
	want := []string{c.tmuxPath, "kill-session", "-t", "dead"}
	got := fr.calls[0]
	if len(got) != len(want) {
		t.Fatalf("unexpected arg count: %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d: got %q want %q", i, got[i], want[i])
		}
	}
}
