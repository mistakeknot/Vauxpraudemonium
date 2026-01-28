// Package watch provides continuous competitor monitoring for Pollard.
// It runs configured hunters on a schedule, diffs against previous findings,
// and emits signals for new discoveries.
package watch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mistakeknot/autarch/internal/pollard/api"
	"github.com/mistakeknot/autarch/internal/pollard/config"
)

// WatchConfig extends Pollard config with watch-specific settings.
type WatchConfig struct {
	Enabled  bool     `yaml:"enabled"`
	Interval string   `yaml:"interval"` // e.g., "24h"
	Hunters  []string `yaml:"hunters"`
	NotifyOn []string `yaml:"notify_on"` // signal types to emit
}

// Watcher runs periodic scans and diffs results.
type Watcher struct {
	projectPath string
	scanner     *api.Scanner
	config      *config.Config
	watchCfg    WatchConfig
}

// NewWatcher creates a Watcher from project config.
func NewWatcher(projectPath string, scanner *api.Scanner, cfg *config.Config, watchCfg WatchConfig) *Watcher {
	return &Watcher{
		projectPath: projectPath,
		scanner:     scanner,
		config:      cfg,
		watchCfg:    watchCfg,
	}
}

// RunOnce performs a single watch cycle: scan + diff + emit signals.
func (w *Watcher) RunOnce(ctx context.Context) (*WatchResult, error) {
	hunters := w.watchCfg.Hunters
	if len(hunters) == 0 {
		hunters = []string{"competitor-tracker", "hackernews-trendwatcher"}
	}

	// Run scan
	result, err := w.scanner.Scan(ctx, api.ScanOptions{
		Hunters: hunters,
	})
	if err != nil {
		return nil, fmt.Errorf("watch scan: %w", err)
	}

	// Build current snapshot
	current := &WatchSnapshot{
		ScannedAt:     time.Now(),
		TotalSources:  result.TotalSources,
		TotalInsights: result.TotalInsights,
		OutputFiles:   result.OutputFiles,
	}

	// Load previous snapshot
	previous, _ := loadSnapshot(w.projectPath)

	// Diff
	diff := diffSnapshots(previous, current)

	// Save current as new baseline
	if err := saveSnapshot(w.projectPath, current); err != nil {
		return nil, fmt.Errorf("saving watch snapshot: %w", err)
	}

	return &WatchResult{
		Snapshot: current,
		Diff:     diff,
		IsFirst:  previous == nil,
	}, nil
}

// Run starts a continuous watch loop. Blocks until context is cancelled.
func (w *Watcher) Run(ctx context.Context) error {
	interval, err := time.ParseDuration(w.watchCfg.Interval)
	if err != nil {
		interval = 24 * time.Hour
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start
	if _, err := w.RunOnce(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "watch cycle error: %v\n", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if _, err := w.RunOnce(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "watch cycle error: %v\n", err)
			}
		}
	}
}

// WatchSnapshot captures the state of a watch scan.
type WatchSnapshot struct {
	ScannedAt     time.Time `json:"scanned_at"`
	TotalSources  int       `json:"total_sources"`
	TotalInsights int       `json:"total_insights"`
	OutputFiles   []string  `json:"output_files"`
}

// WatchResult is the output of a single watch cycle.
type WatchResult struct {
	Snapshot *WatchSnapshot
	Diff     *WatchDiff
	IsFirst  bool
}

func watchDir(projectPath string) string {
	return filepath.Join(projectPath, ".pollard", "watch")
}

func snapshotPath(projectPath string) string {
	return filepath.Join(watchDir(projectPath), "last_scan.json")
}

func loadSnapshot(projectPath string) (*WatchSnapshot, error) {
	data, err := os.ReadFile(snapshotPath(projectPath))
	if err != nil {
		return nil, err
	}
	var s WatchSnapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func saveSnapshot(projectPath string, s *WatchSnapshot) error {
	dir := watchDir(projectPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(snapshotPath(projectPath), data, 0644)
}
