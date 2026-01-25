package tui

import (
	"context"
	"testing"
	"time"

	"github.com/mistakeknot/autarch/internal/bigend/aggregator"
	"github.com/mistakeknot/autarch/internal/bigend/tmux"
)

type fakeAggStatus struct {
	state aggregator.State
}

func (f *fakeAggStatus) GetState() aggregator.State { return f.state }
func (f *fakeAggStatus) Refresh(ctx context.Context) error { return nil }
func (f *fakeAggStatus) NewSession(string, string, string) error { return nil }
func (f *fakeAggStatus) RestartSession(string, string, string) error { return nil }
func (f *fakeAggStatus) RenameSession(string, string) error { return nil }
func (f *fakeAggStatus) ForkSession(string, string, string) error { return nil }
func (f *fakeAggStatus) AttachSession(string) error { return nil }
func (f *fakeAggStatus) StartMCP(context.Context, string, string) error { return nil }
func (f *fakeAggStatus) StopMCP(string, string) error { return nil }

type fakeStatusClient struct {
	calls map[string]int
}

func (f *fakeStatusClient) DetectStatus(name string) tmux.Status {
	f.calls[name]++
	return tmux.StatusRunning
}

func TestStatusCacheAvoidsRepeatedDetects(t *testing.T) {
	agg := &fakeAggStatus{state: aggregator.State{
		Sessions: []aggregator.TmuxSession{{Name: "a"}, {Name: "b"}},
	}}
	m := New(agg, "")
	fake := &fakeStatusClient{calls: map[string]int{}}
	m.tmuxClient = fake
	m.statusTTL = time.Minute
	fixed := time.Date(2026, 1, 22, 0, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return fixed }

	m.updateLists()
	m.updateLists()

	if fake.calls["a"] != 1 || fake.calls["b"] != 1 {
		t.Fatalf("expected cached detect status (a=%d b=%d)", fake.calls["a"], fake.calls["b"])
	}
}
