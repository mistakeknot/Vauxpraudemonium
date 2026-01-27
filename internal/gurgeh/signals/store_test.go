package signals

import (
	"os"
	"testing"
	"time"

	"github.com/mistakeknot/autarch/pkg/signals"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestStoreEmitAndActive(t *testing.T) {
	s := testStore(t)

	sig := signals.Signal{
		ID:            "sig-001",
		Type:          signals.SignalAssumptionDecayed,
		Source:        "gurgeh",
		SpecID:        "SPEC-1",
		AffectedField: "assumptions",
		Severity:      signals.SeverityWarning,
		Title:         "Assumption decayed",
		Detail:        "ASSM-001 confidence dropped",
		CreatedAt:     time.Now(),
	}

	if err := s.Emit(sig); err != nil {
		t.Fatalf("Emit: %v", err)
	}

	active, err := s.Active("SPEC-1")
	if err != nil {
		t.Fatalf("Active: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("expected 1 active signal, got %d", len(active))
	}
	if active[0].ID != "sig-001" {
		t.Errorf("expected ID sig-001, got %s", active[0].ID)
	}
	if active[0].AffectedField != "assumptions" {
		t.Errorf("expected affected_field assumptions, got %s", active[0].AffectedField)
	}
}

func TestStoreDeduplication(t *testing.T) {
	s := testStore(t)

	sig := signals.Signal{
		ID:            "sig-001",
		Type:          signals.SignalAssumptionDecayed,
		Source:        "gurgeh",
		SpecID:        "SPEC-1",
		AffectedField: "assumptions",
		Severity:      signals.SeverityWarning,
		Detail:        "first",
		CreatedAt:     time.Now(),
	}

	if err := s.Emit(sig); err != nil {
		t.Fatalf("Emit 1: %v", err)
	}

	// Same spec_id, type, affected_field but different ID — should be ignored
	sig.ID = "sig-002"
	sig.Detail = "second"
	if err := s.Emit(sig); err != nil {
		t.Fatalf("Emit 2: %v", err)
	}

	if count := s.Count("SPEC-1"); count != 1 {
		t.Errorf("expected count 1 after dedup, got %d", count)
	}
}

func TestStoreDismiss(t *testing.T) {
	s := testStore(t)

	sig := signals.Signal{
		ID:            "sig-001",
		Type:          signals.SignalAssumptionDecayed,
		Source:        "gurgeh",
		SpecID:        "SPEC-1",
		AffectedField: "assumptions",
		Severity:      signals.SeverityWarning,
		Detail:        "test",
		CreatedAt:     time.Now(),
	}
	s.Emit(sig)

	if err := s.Dismiss("sig-001"); err != nil {
		t.Fatalf("Dismiss: %v", err)
	}

	if count := s.Count("SPEC-1"); count != 0 {
		t.Errorf("expected count 0 after dismiss, got %d", count)
	}

	active, _ := s.Active("SPEC-1")
	if len(active) != 0 {
		t.Errorf("expected 0 active after dismiss, got %d", len(active))
	}

	// Dismiss again should error
	if err := s.Dismiss("sig-001"); err == nil {
		t.Error("expected error dismissing already dismissed signal")
	}
}

func TestStoreCount(t *testing.T) {
	s := testStore(t)

	// Empty
	if count := s.Count("SPEC-1"); count != 0 {
		t.Errorf("expected 0 for empty, got %d", count)
	}

	for i, field := range []string{"goals", "assumptions", "scope"} {
		s.Emit(signals.Signal{
			ID:            generateID(),
			Type:          signals.SignalAssumptionDecayed,
			Source:        "gurgeh",
			SpecID:        "SPEC-1",
			AffectedField: field,
			Severity:      signals.SeverityWarning,
			Detail:        "test",
			CreatedAt:     time.Now(),
		})
		if count := s.Count("SPEC-1"); count != i+1 {
			t.Errorf("expected count %d, got %d", i+1, count)
		}
	}
}

func TestStoreDescriptionCap(t *testing.T) {
	s := testStore(t)

	longDesc := make([]byte, maxDescriptionLen+1000)
	for i := range longDesc {
		longDesc[i] = 'x'
	}

	sig := signals.Signal{
		ID:            "sig-cap",
		Type:          signals.SignalSpecHealthLow,
		Source:        "gurgeh",
		SpecID:        "SPEC-1",
		AffectedField: "health",
		Severity:      signals.SeverityCritical,
		Detail:        string(longDesc),
		CreatedAt:     time.Now(),
	}
	if err := s.Emit(sig); err != nil {
		t.Fatalf("Emit long description: %v", err)
	}

	active, _ := s.Active("SPEC-1")
	if len(active) != 1 {
		t.Fatalf("expected 1 active, got %d", len(active))
	}
	if len(active[0].Detail) > maxDescriptionLen {
		t.Errorf("description not capped: len=%d", len(active[0].Detail))
	}
}

func TestStoreFilePermissions(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer s.Close()

	dbPath := dir + "/.gurgeh/signals/signals.db"
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("Stat db: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected 0600 permissions, got %o", perm)
	}
}
