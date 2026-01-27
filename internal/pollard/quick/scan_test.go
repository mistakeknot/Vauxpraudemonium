package quick

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewScanner(t *testing.T) {
	scanner := NewScanner()
	if scanner == nil {
		t.Fatal("expected non-nil scanner")
	}
}

func TestQuickScanWithTimeout(t *testing.T) {
	// This test uses real network - skip in short mode
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "quickscan-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	scanner := NewScanner()
	result, err := scanner.Scan(ctx, "todo app", tmpDir)

	if err != nil {
		t.Skipf("skipping (network required): %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Topic != "todo app" {
		t.Errorf("expected topic 'todo app', got %q", result.Topic)
	}
}

func TestQuickScanReturnsPartialOnTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "quickscan-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Very short timeout - should return partial/empty results, not error
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	scanner := NewScanner()
	result, err := scanner.Scan(ctx, "test", tmpDir)

	// Should not error - returns empty/partial results
	if err != nil {
		t.Fatalf("expected no error on timeout, got: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result even on timeout")
	}
}
