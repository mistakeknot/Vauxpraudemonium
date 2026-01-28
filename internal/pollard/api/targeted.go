package api

import (
	"context"
	"fmt"
)

// ScanMode controls research depth.
type ScanMode string

const (
	ScanModeQuick    ScanMode = "quick"    // ~30s, minimal hunters
	ScanModeBalanced ScanMode = "balanced" // ~2min, moderate depth
	ScanModeDeep     ScanMode = "deep"     // ~5min, full depth
)

// TargetedScanOpts configures a phase-specific research scan.
type TargetedScanOpts struct {
	SpecID  string
	Hunters []string // hunter names to run
	Mode    ScanMode
	Query   string // research query extracted from spec phase
}

// TargetedScanResult holds results from a targeted scan.
type TargetedScanResult struct {
	SpecID        string
	Mode          ScanMode
	Hunters       []string
	TotalSources  int
	TotalInsights int
	OutputFiles   []string
	Errors        []error
}

// RunTargetedScan executes a phase-specific research scan using the existing
// Scanner pipeline but filtered to only the specified hunters and mode.
func (s *Scanner) RunTargetedScan(ctx context.Context, opts TargetedScanOpts) (*TargetedScanResult, error) {
	if len(opts.Hunters) == 0 {
		return &TargetedScanResult{SpecID: opts.SpecID, Mode: opts.Mode}, nil
	}

	// Build scan options filtering to requested hunters
	scanOpts := ScanOptions{
		Hunters: opts.Hunters,
	}
	if opts.Query != "" {
		scanOpts.Queries = []string{opts.Query}
	}

	result, err := s.Scan(ctx, scanOpts)
	if err != nil {
		return nil, fmt.Errorf("targeted scan: %w", err)
	}

	return &TargetedScanResult{
		SpecID:        opts.SpecID,
		Mode:          opts.Mode,
		Hunters:       opts.Hunters,
		TotalSources:  result.TotalSources,
		TotalInsights: result.TotalInsights,
		OutputFiles:   result.OutputFiles,
		Errors:        result.Errors,
	}, nil
}
