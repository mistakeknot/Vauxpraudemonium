package watch

import (
	"context"
	"testing"
	"time"

	"github.com/mistakeknot/autarch/pkg/signals"
)

type testPublisher struct {
	sigs []signals.Signal
	err  error
}

func (p *testPublisher) Publish(_ context.Context, sig signals.Signal) error {
	p.sigs = append(p.sigs, sig)
	return p.err
}

func TestEmitSignalsPublishesWhenChanged(t *testing.T) {
	pub := &testPublisher{}
	w := &Watcher{
		watchCfg:  WatchConfig{NotifyOn: []string{string(signals.SignalCompetitorShipped)}},
		publisher: pub,
	}

	diff := &WatchDiff{NewSources: 1}
	w.emitSignals(context.Background(), diff)

	if len(pub.sigs) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(pub.sigs))
	}
	sig := pub.sigs[0]
	if sig.Type != signals.SignalCompetitorShipped {
		t.Fatalf("expected competitor_shipped, got %s", sig.Type)
	}
	if sig.Source != "pollard" {
		t.Fatalf("expected source pollard, got %s", sig.Source)
	}
	if sig.AffectedField != "watch" {
		t.Fatalf("expected affected_field watch, got %s", sig.AffectedField)
	}
	if sig.CreatedAt.IsZero() {
		t.Fatalf("expected created_at to be set")
	}
}

func TestEmitSignalsRespectsNotifyOn(t *testing.T) {
	pub := &testPublisher{}
	w := &Watcher{
		watchCfg:  WatchConfig{NotifyOn: []string{"assumption_decayed"}},
		publisher: pub,
	}

	diff := &WatchDiff{NewSources: 1}
	w.emitSignals(context.Background(), diff)

	if len(pub.sigs) != 0 {
		t.Fatalf("expected no signals, got %d", len(pub.sigs))
	}
}

func TestEmitSignalsNoChangesNoPublish(t *testing.T) {
	pub := &testPublisher{}
	w := &Watcher{publisher: pub}

	w.emitSignals(context.Background(), &WatchDiff{})
	if len(pub.sigs) != 0 {
		t.Fatalf("expected no signals, got %d", len(pub.sigs))
	}

	w.emitSignals(context.Background(), nil)
	if len(pub.sigs) != 0 {
		t.Fatalf("expected no signals, got %d", len(pub.sigs))
	}
}

func TestNotifyEnabledDefaultsTrue(t *testing.T) {
	w := &Watcher{watchCfg: WatchConfig{}}
	if !w.notifyEnabled(signals.SignalCompetitorShipped) {
		t.Fatalf("expected notifyEnabled true when NotifyOn empty")
	}
}

func TestNotifyEnabledMatchesCaseInsensitive(t *testing.T) {
	w := &Watcher{watchCfg: WatchConfig{NotifyOn: []string{"CompEtItOr_ShiPPed"}}}
	if !w.notifyEnabled(signals.SignalCompetitorShipped) {
		t.Fatalf("expected notifyEnabled true for case-insensitive match")
	}
}

func TestEmitSignalsAddsDetailWhenFilesPresent(t *testing.T) {
	pub := &testPublisher{}
	w := &Watcher{publisher: pub}
	diff := &WatchDiff{NewSources: 1, NewFiles: []string{"a", "b"}}

	w.emitSignals(context.Background(), diff)
	if len(pub.sigs) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(pub.sigs))
	}
	if pub.sigs[0].Detail == "" {
		t.Fatalf("expected detail to be set")
	}
	if pub.sigs[0].CreatedAt.Before(time.Now().Add(-1 * time.Minute)) {
		t.Fatalf("created_at looks too old")
	}
}
