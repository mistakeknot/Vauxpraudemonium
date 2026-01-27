---
module: gurgeh
date: 2026-01-26
problem_type: architecture
component: arbiter
symptoms:
  - "Import cycle between arbiter and consistency/confidence/quick packages"
  - "Hunter API returns file paths, not structured data"
  - "Plan color constants don't match actual API"
root_cause: "Cross-package dependencies create cycles in Go; plan assumptions diverged from implementation reality"
severity: medium
tags: [arbiter, import-cycle, architecture, hunter-api, tui]
---

# Arbiter Spec Sprint Architecture Patterns

## Problem Statement

Implementing the Arbiter Spec Sprint required 3 new packages (consistency, confidence, quick) that import arbiter types, while the orchestrator in the arbiter package needs to use all three — creating a Go import cycle.

Additionally, the implementation plan made assumptions about APIs (hunter output format, TUI color constants) that didn't match reality.

## Investigation

### Import Cycle
- `arbiter` defines `SprintState`, `Conflict`, `ConfidenceScore`
- `consistency` imports `arbiter` for `SprintState` input and `Conflict` output
- `confidence` imports `arbiter` for `SprintState` input and `ConfidenceScore` output
- `quick` imports `arbiter` for `QuickScanResult` output
- `arbiter/orchestrator.go` needs all three → circular dependency

### Hunter API
- Plan assumed `HuntResult.Items []Item` field
- Actual API: `HuntResult.OutputFiles []string` (YAML file paths)
- Hunters write results to YAML files and return paths

### Color API
- Plan referenced `tui.TokyoNight.Cyan`, `tui.TokyoNight.Blue`, etc.
- Actual: `tui.ColorPrimary`, `tui.ColorSecondary`, `tui.ColorSuccess`, etc.

## Root Cause

Go's strict prohibition on import cycles requires careful package design. When types flow bidirectionally between packages, adapter patterns are needed.

Plan-to-implementation drift is normal — the plan was written before exploring actual APIs.

## Solution

### Import Cycle: Local Adapter Sub-Packages

Created lightweight sub-packages under `arbiter/` that define local types:

```
internal/gurgeh/arbiter/
├── types.go            # Core types (SprintState, Phase, etc.)
├── orchestrator.go     # Uses local adapters
├── consistency/
│   └── engine.go       # Local Engine with own SectionInfo type
├── confidence/
│   └── calculator.go   # Local Calculator with own Score type
└── quick/
    └── scanner.go      # Local Scanner with own ScanResult type
```

The orchestrator converts between adapter types and arbiter types via helper methods (`checkConsistency`, `updateConfidence`).

**Trade-off:** Some type duplication, but clean dependency graph. The existing top-level packages (`internal/gurgeh/consistency/`, `internal/gurgeh/confidence/`) remain for direct CLI use.

### Hunter API: Parse YAML Output Files

```go
// Quick scanner parses hunter output files instead of accessing Items
for _, path := range result.OutputFiles {
    data, err := os.ReadFile(path)
    // Parse YAML into local structs
    var output gitHubOutput
    yaml.Unmarshal(data, &output)
}
```

### Color API: Use Actual Constants

```go
// Wrong (from plan):
lipgloss.NewStyle().Foreground(tui.TokyoNight.Cyan)

// Correct (actual API):
lipgloss.NewStyle().Foreground(tui.ColorPrimary)
```

## Prevention

1. **Before implementing cross-package orchestrators**: Map the dependency graph first. If package A's types are used by B, C, D, and A needs to call B, C, D — use adapter patterns or interfaces.
2. **Before implementing against external APIs**: Run an exploration subagent to verify actual function signatures and return types.
3. **Keep a mapping of plan assumptions vs reality**: When adapting, document the divergence for future plan writers.
