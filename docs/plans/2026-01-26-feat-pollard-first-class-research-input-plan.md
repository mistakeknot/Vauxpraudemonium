---
title: "feat: Pollard Integration as First-Class Research Input"
type: feat
date: 2026-01-26
---

# Pollard Integration as First-Class Research Input

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Wire Pollard research intelligence into the Arbiter Spec Sprint so research actively informs PRD generation — not just a binary confidence flag.

**Architecture:** Use Intermute as the integration boundary between Pollard and Arbiter. Pollard publishes `Insight` entities to Intermute; Arbiter reads them via the Go client SDK. The embedded Intermute server runs in-process for zero-serialization communication. WebSocket events enable real-time research updates during deep scans.

**Tech Stack:** Go, Intermute client SDK (`github.com/mistakeknot/intermute/client`), embedded Intermute server (`github.com/mistakeknot/intermute/pkg/embedded`), existing Pollard hunters, existing Arbiter orchestrator.

---

## Enhancement Summary

**Deepened on:** 2026-01-26
**Review agents used:** architecture-strategist, performance-oracle, agent-native-reviewer, code-simplicity-reviewer, pattern-recognition-specialist, security-sentinel, learnings-researcher

### Key Improvements from Review

1. **Define `ResearchProvider` interface** — Arbiter should depend on an interface, not concrete `ResearchBridge`. Enables testing with mocks and future provider swaps. (architecture-strategist)
2. **Simplify confidence formula** — Use `clamp(findingCount/5, 0, 1)` instead of 3-factor weighted formula. Over-engineering for an early-stage metric. (code-simplicity-reviewer)
3. **Singleton scanner + skip Fetch stage in quick mode** — Quick scan's 30s budget is consumed by full pipeline; skip network-heavy fetch stage and reuse scanner instance. (performance-oracle)
4. **CRITICAL: CLI/MCP parity gap** — Sprint lifecycle has zero agent-accessible surface. Need at minimum: `gurgeh sprint start`, `sprint accept`, `sprint advance`, `sprint state`, `sprint handoff`. Without these, Pollard/Coldwine agents cannot programmatically trigger sprints. (agent-native-reviewer)
5. **Use narrow interface pattern** — Follow codebase convention: `SpecManager`, `InsightCreator`, `MessageSender` — not broad client wrappers. (pattern-recognition-specialist)
6. **Security: sanitize synthesizer inputs** — `internal/pollard/research/synthesizer.go` passes external API content as CLI args. Validate/escape before use. (security-sentinel)
7. **TUI: account for padding** — Subtract parent Padding(1,3) → 6h, 2v from available dimensions. Use `ansi.StringWidth()`. (learnings-researcher)

### Considerations Not Adopted

- **Collapse to 4 tasks** (simplicity-reviewer): Rejected — the 8-task TDD structure provides better review checkpoints. Each task is already bite-sized.
- **Skip Intermute entirely** (simplicity-reviewer): Rejected — Intermute is the project's coordination layer and provides cross-tool linking that direct function calls cannot.

---

## Context

### What exists today

- **Arbiter Spec Sprint** (`internal/gurgeh/arbiter/orchestrator.go`): 6-phase propose-first PRD flow with consistency engine, confidence scoring, and handoff options.
- **Quick scan stub** (`internal/gurgeh/arbiter/quick/scanner.go`): Returns placeholder text. The real implementation at `internal/pollard/quick/scan.go` exists but isn't wired in.
- **Pollard research coordinator** (`internal/pollard/research/coordinator.go`): Parallel hunter orchestration with 12 registered hunters.
- **Intermute** (`github.com/mistakeknot/intermute`): Full REST + WebSocket + embedded server with first-class Insight, Spec, and CUJ entities. Has `LinkInsightToSpec(insightID, specID)` API.

### 7 gaps this plan closes

1. **Stub scanner** → Wire real Pollard quick scan via Intermute
2. **Data loss** → Populate GitHubHits/HNHits from Intermute Insights
3. **Binary confidence** → Quality-based scoring from finding count, relevance, diversity
4. **Narrow research types** → Read all Intermute Insights (any hunter type)
5. **No deep scan** → Async deep scan handoff via Pollard coordinator + WebSocket events
6. **No pre-sprint import** → Read existing `.pollard/` insights at sprint start
7. **No Intermute linkage** → Create Spec entity, link Insights to it

### Known gotchas (from docs/solutions/)

- **Hunter API returns `OutputFiles []string`**, not structured data — must parse YAML files
- **Import cycles** — solved by using Intermute as boundary (no direct Pollard→Arbiter imports)
- **TUI dimension mismatch** — subtract parent padding (6h, 2v) when rendering async results
- **`SprintState.ID` is empty** — must generate UUID for Intermute Spec linking

---

## Task 1: Generate Sprint ID and Create Intermute Spec

**Files:**
- Modify: `internal/gurgeh/arbiter/types.go` — `NewSprintState` generates UUID
- Create: `internal/gurgeh/arbiter/intermute.go` — Intermute client wrapper
- Test: `internal/gurgeh/arbiter/intermute_test.go`

**What to do:**

1. Add `github.com/google/uuid` dependency (or use `crypto/rand` for simple ID generation)
2. In `NewSprintState`, set `ID` to a generated UUID string
3. Create `intermute.go` with a `ResearchBridge` struct that wraps the Intermute client:

```go
type ResearchBridge struct {
    client *client.Client
    project string
}

func NewResearchBridge(intermuteURL, project string) (*ResearchBridge, error)
func (rb *ResearchBridge) CreateSpec(ctx context.Context, state *SprintState) (string, error)
func (rb *ResearchBridge) LinkInsight(ctx context.Context, insightID, specID string) error
func (rb *ResearchBridge) FetchLinkedInsights(ctx context.Context, specID string) ([]client.Insight, error)
```

4. `CreateSpec` maps SprintState → Intermute Spec (status: "draft", title from Problem content)
5. Write tests: spec creation, insight linking, fetch with no results

### Research Insights (Task 1)

- **Use `ResearchProvider` interface** instead of concrete `ResearchBridge`. Define: `type ResearchProvider interface { CreateSpec(...) (string, error); LinkInsight(...) error; FetchLinkedInsights(...) ([]Insight, error) }`. Orchestrator takes `ResearchProvider` (nil = no-research mode). This follows the codebase's narrow interface pattern (like `SpecManager`, `InsightCreator`).
- **UUID generation**: Prefer `crypto/rand` + hex encoding over `google/uuid` dependency to minimize go.mod churn. 16 random bytes → 32-char hex is unique enough for local sprint IDs.
- **Security**: When creating Intermute Spec from sprint content, sanitize Problem section text before passing as title (no newlines, max 200 chars).

**Step 1:** Write failing test for `NewSprintState` generating non-empty ID
**Step 2:** Run test, verify it fails
**Step 3:** Add UUID generation to `NewSprintState`
**Step 4:** Run test, verify it passes
**Step 5:** Write failing test for `ResearchBridge.CreateSpec`
**Step 6:** Implement `ResearchBridge` with Intermute client
**Step 7:** Run tests, verify pass
**Step 8:** Commit: `feat(arbiter): add sprint ID generation and Intermute research bridge`

---

## Task 2: Wire Real Quick Scan Through Intermute

**Files:**
- Modify: `internal/gurgeh/arbiter/orchestrator.go` — replace stub scanner, use ResearchBridge
- Modify: `internal/gurgeh/arbiter/orchestrator_test.go` — update tests
- Remove: `internal/gurgeh/arbiter/quick/` — delete stub adapter package

**What to do:**

1. Remove the stub `internal/gurgeh/arbiter/quick/` package entirely
2. Add `ResearchBridge` field to `Orchestrator` struct
3. Update `NewOrchestrator` to accept a `ResearchBridge` (or nil for no-research mode)
4. Rewrite `runQuickScan` to:
   a. Use the real `internal/pollard/quick` scanner (it already returns `*arbiter.QuickScanResult`)
   b. After scan, publish each finding as an Intermute Insight via `ResearchBridge`
   c. Link each Insight to the sprint's Spec ID
   d. Populate **all** fields on `QuickScanResult` including `GitHubHits` and `HNHits`
5. Handle errors gracefully — if scan fails, sprint continues (current behavior)
6. Handle 0 results — set `ResearchCtx` with empty slices (not nil) so we can distinguish "scanned but found nothing" from "never scanned"

**Important:** The real scanner at `internal/pollard/quick/scan.go` imports `internal/gurgeh/arbiter` types directly. This is fine — it's a one-way dependency (pollard → arbiter types). The orchestrator imports pollard/quick, not the reverse.

### Research Insights (Task 2)

- **Performance**: The 30s quick scan budget gets consumed by the full Pollard pipeline (Fetch→Parse→Score→Synthesize). In quick mode, skip the Fetch stage entirely — use cached/pre-fetched data only. Make Scanner a singleton to avoid re-initializing hunter registries.
- **Use existing `Coordinator.StartRun()`** for async orchestration instead of raw goroutines. This integrates with the existing Bubble Tea message system (RunStartedMsg, HunterUpdateMsg).
- **Consolidate `mapFindingToInsight`** — this conversion logic exists in multiple places. Create one canonical mapper in the Intermute bridge.

**Step 1:** Write failing test for orchestrator with nil ResearchBridge (backward compat)
**Step 2:** Run test, verify fail
**Step 3:** Update Orchestrator to accept optional ResearchBridge
**Step 4:** Run test, verify pass
**Step 5:** Delete `internal/gurgeh/arbiter/quick/` stub package
**Step 6:** Wire real pollard/quick scanner in runQuickScan, populate all QuickScanResult fields
**Step 7:** Run all tests: `go test ./internal/gurgeh/arbiter/...`
**Step 8:** Commit: `feat(arbiter): wire real Pollard quick scan and Intermute publishing`

---

## Task 3: Quality-Based Research Confidence Scoring

**Files:**
- Modify: `internal/gurgeh/arbiter/confidence/calculator.go` — richer input
- Modify: `internal/gurgeh/arbiter/orchestrator.go` — pass research quality data
- Test: `internal/gurgeh/arbiter/confidence/calculator_test.go` (create if missing)

**What to do:**

1. Change the local confidence calculator's `Calculate` method to accept a richer research input:

```go
type ResearchQuality struct {
    HasResearch   bool
    FindingCount  int
    SourceCount   int     // distinct hunter types that returned results
    AvgRelevance  float64 // 0.0-1.0 average across findings
}
```

2. Compute Research score as:
   - If `!HasResearch`: 0.0
   - Else: `min(1.0, 0.3*clamp(FindingCount/10, 0, 1) + 0.3*clamp(SourceCount/3, 0, 1) + 0.4*AvgRelevance)`
   - This rewards: having findings (30%), diverse sources (30%), and relevant findings (40%)

3. Update `updateConfidence` in orchestrator to compute `ResearchQuality` from `state.ResearchCtx`:
   - `FindingCount` = len(GitHubHits) + len(HNHits) + len(ResearchFindings)
   - `SourceCount` = count of non-empty hit slices
   - `AvgRelevance` = average of finding relevance scores (default 0.5 for GitHub/HN hits which lack scores)

4. **Fix the zero-results false positive**: scanned-but-empty now scores low (FindingCount=0 → 0.3*0 + 0.3*0 + 0.4*0 = 0.0 even though HasResearch=true). This is correct — having scanned but found nothing IS low confidence.

### Research Insights (Task 3)

- **Simplification option**: The 3-factor weighted formula (30% count + 30% sources + 40% relevance) may be over-designed for v1. Consider starting with `clamp(findingCount/5, 0, 1)` and iterating based on real usage data. The current plan formula is fine if you want it, but don't block on getting weights "right" — they're easily tunable.
- **Edge case**: AvgRelevance defaults to 0.5 for GitHub/HN hits. Document this assumption clearly — it means a scan with only GitHub results gets 40% × 0.5 = 0.2 from relevance alone.

**Step 1:** Write failing test: 0 findings → Research score near 0
**Step 2:** Write failing test: 10+ findings from 3 sources → Research score near 1.0
**Step 3:** Run tests, verify they fail
**Step 4:** Implement ResearchQuality struct and updated Calculate method
**Step 5:** Update orchestrator's updateConfidence to compute ResearchQuality
**Step 6:** Run tests, verify pass
**Step 7:** Commit: `feat(arbiter): quality-based research confidence scoring`

---

## Task 4: Expanded Research Types via Intermute Insights

**Files:**
- Modify: `internal/gurgeh/arbiter/types.go` — add ResearchFindings field
- Modify: `internal/gurgeh/arbiter/intermute.go` — fetch and convert insights
- Modify: `internal/gurgeh/arbiter/orchestrator.go` — populate from Intermute
- Test: update existing tests

**What to do:**

1. Add to `SprintState`:

```go
type ResearchFinding struct {
    ID         string
    Title      string
    Summary    string
    Source     string    // URL
    SourceType string    // "github", "hackernews", "arxiv", "competitor", etc.
    Relevance  float64   // 0.0-1.0
    Tags       []string
}

// In SprintState:
ResearchFindings []ResearchFinding  // All findings from any source
```

2. Add `FetchResearchFindings(ctx, specID)` to `ResearchBridge` that:
   - Calls `client.ListInsights` filtered by spec linkage
   - Maps Intermute `Insight` → `ResearchFinding`
   - Returns findings sorted by relevance descending

3. After quick scan completes in `runQuickScan`, also call `FetchResearchFindings` to populate `state.ResearchFindings`

4. Update confidence scoring to include `ResearchFindings` in `FindingCount` and `AvgRelevance`

**Step 1:** Write failing test for ResearchFinding type and FetchResearchFindings
**Step 2:** Run test, verify fail
**Step 3:** Implement types and bridge method
**Step 4:** Wire into orchestrator's runQuickScan
**Step 5:** Update confidence input to count all findings
**Step 6:** Run all tests
**Step 7:** Commit: `feat(arbiter): expanded research types via Intermute insights`

---

## Task 5: Pre-Sprint Research Import (--from-research)

**Files:**
- Modify: `internal/gurgeh/arbiter/orchestrator.go` — add StartWithResearch method
- Modify: `internal/gurgeh/arbiter/intermute.go` — import from .pollard/
- Test: `internal/gurgeh/arbiter/orchestrator_test.go`

**What to do:**

1. Add `StartWithResearch` method to Orchestrator:

```go
func (o *Orchestrator) StartWithResearch(ctx context.Context, userInput string, pollardPath string) (*SprintState, error)
```

2. This method:
   a. Calls `Start()` normally to create sprint and generate Problem draft
   b. Reads `.pollard/insights/*.yaml` files from the given path
   c. Creates Intermute Insights for each and links to sprint's Spec ID
   d. Populates `state.ResearchCtx` and `state.ResearchFindings` from the imported data
   e. Skips the quick scan trigger (research already loaded)

3. Add a `skipQuickScan` flag to SprintState (or check if ResearchFindings is non-empty in `runQuickScan`)

4. Parse YAML insight files using the format from `internal/pollard/insights/insight.go`:
   ```yaml
   id: "insight-xxx"
   title: "..."
   category: "..."
   sources: [...]
   findings: [...]
   recommendations: [...]
   linked_features: [...]
   ```

**Step 1:** Write failing test: StartWithResearch with mock .pollard/ dir
**Step 2:** Run test, verify fail
**Step 3:** Implement StartWithResearch with YAML parsing
**Step 4:** Add skip-quick-scan logic when research pre-loaded
**Step 5:** Run tests, verify pass
**Step 6:** Commit: `feat(arbiter): pre-sprint research import from .pollard/`

---

## Task 6: Async Deep Scan Handoff

**Files:**
- Modify: `internal/gurgeh/arbiter/orchestrator.go` — add ExecuteHandoff method
- Modify: `internal/gurgeh/arbiter/types.go` — add DeepScanStatus to SprintState
- Create: `internal/gurgeh/arbiter/deepscan.go` — deep scan orchestration
- Test: `internal/gurgeh/arbiter/deepscan_test.go`

**What to do:**

1. Add to SprintState:

```go
DeepScanStatus string    // "", "running", "complete", "failed"
DeepScanID     string    // Pollard run ID for tracking
```

2. Create `deepscan.go` with:

```go
type DeepScanConfig struct {
    Hunters []string  // hunter names to use (default: all)
    Topics  []string  // extracted from sprint sections
}

func (o *Orchestrator) StartDeepScan(ctx context.Context, state *SprintState) (*SprintState, error)
func (o *Orchestrator) CheckDeepScan(ctx context.Context, state *SprintState) (*SprintState, error)
func (o *Orchestrator) ImportDeepScanResults(ctx context.Context, state *SprintState) (*SprintState, error)
```

3. `StartDeepScan`:
   - Extracts topics from Problem + Users + Features sections
   - Sends Intermute message to Pollard agent requesting deep research
   - Sets `DeepScanStatus = "running"` and `DeepScanID`
   - Returns immediately (async — sprint session can end)

4. `CheckDeepScan`:
   - Queries Intermute for deep scan status (via message thread or session entity)
   - Returns updated state with current status

5. `ImportDeepScanResults`:
   - Fetches all Insights linked to this sprint's Spec ID
   - Updates `ResearchFindings` and `ResearchCtx`
   - Recalculates confidence score
   - Sets `DeepScanStatus = "complete"`

6. Update `GetHandoffOptions` to wire the "Deep Research" option to `StartDeepScan`

### Research Insights (Task 6)

- **Simplicity consideration**: Deep scan could just shell out to `pollard scan` CLI rather than building a full messaging protocol. The CLI already handles hunter selection, parallel execution, and output writing. An Intermute message saying "scan complete, read .pollard/" is simpler than streaming findings via WebSocket.
- **Reuse `ResearchOverlay` TUI component** if one exists — check for overlay patterns in `pkg/tui/`.

**Step 1:** Write failing test for StartDeepScan setting status
**Step 2:** Write failing test for ImportDeepScanResults updating findings
**Step 3:** Run tests, verify fail
**Step 4:** Implement deepscan.go
**Step 5:** Wire into GetHandoffOptions/ExecuteHandoff
**Step 6:** Run tests, verify pass
**Step 7:** Commit: `feat(arbiter): async deep scan handoff via Intermute messaging`

---

## Task 7: Sprint TUI Research Display

**Files:**
- Modify: `internal/gurgeh/tui/sprint.go` — render research findings
- Test: `internal/gurgeh/tui/sprint_test.go`

**What to do:**

1. Add research panel to sprint TUI View when `state.ResearchCtx` is non-nil:
   - Show "📊 Quick Scan Results" header
   - List GitHub findings: `★ {stars} {name} — {description}`
   - List HN findings: `▲ {points} {title} — {theme}`
   - Show summary text
   - If `ResearchFindings` has additional types, show "📚 {count} additional findings"

2. Add deep scan status indicator when `DeepScanStatus` is non-empty:
   - "running" → spinner with "🔍 Deep scan in progress..."
   - "complete" → "✅ Deep scan complete — {count} new findings"
   - "failed" → "⚠️ Deep scan failed"

3. **TUI dimension gotcha**: Subtract parent padding (6 horizontal, 2 vertical) from window dimensions. Use `ansi.StringWidth()` for styled text width calculations.

4. Add keybinding `r` to refresh deep scan status (calls `CheckDeepScan`)

### Research Insights (Task 7)

- **TUI dimension gotcha**: Parent `Padding(1,3)` subtracts 6 horizontal + 2 vertical pixels. Always compute available width as `windowWidth - 6` and height as `windowHeight - 2`. Use `ansi.StringWidth()` for styled text, NOT `len()`.
- **Import cycle prevention**: Use adapter sub-packages (like `arbiter/confidence/`, `arbiter/consistency/`) to break circular deps between TUI and arbiter.

**Step 1:** Write failing test for research panel rendering
**Step 2:** Run test, verify fail
**Step 3:** Implement research display in sprint.go View()
**Step 4:** Add deep scan status display
**Step 5:** Add `r` keybinding for refresh
**Step 6:** Run tests, verify pass
**Step 7:** Commit: `feat(tui): display research findings in sprint view`

---

## Task 8: Integration Test and Documentation

**Files:**
- Create: `internal/gurgeh/arbiter/integration_test.go` — end-to-end test
- Modify: `AGENTS.md` — update Arbiter section
- Modify: `docs/COMPOUND_INTEGRATION.md` — update research handoff docs

**What to do:**

1. Write integration test that:
   - Creates embedded Intermute server (in-process)
   - Creates Orchestrator with ResearchBridge pointing to embedded server
   - Starts sprint, advances through phases
   - Verifies quick scan creates Intermute Insights linked to Spec
   - Verifies confidence scoring reflects research quality
   - Verifies StartWithResearch reads existing insights

2. Update AGENTS.md:
   - Update Key Files table with new files (intermute.go, deepscan.go)
   - Document ResearchBridge architecture
   - Note Intermute dependency for research features

3. Update COMPOUND_INTEGRATION.md:
   - Update research handoff diagram
   - Document Intermute-mediated Pollard↔Arbiter flow

**Step 1:** Write integration test with embedded Intermute
**Step 2:** Run test, verify pass (or fix issues found)
**Step 3:** Update AGENTS.md
**Step 4:** Update COMPOUND_INTEGRATION.md
**Step 5:** Run full test suite: `go test ./...`
**Step 6:** Commit: `test(arbiter): integration test for Pollard research flow`

---

## Acceptance Criteria

### Functional
- [ ] Sprint generates UUID in `SprintState.ID`
- [ ] Quick scan at PhaseFeaturesGoals uses real Pollard hunters (github-scout + hackernews)
- [ ] Quick scan results populate GitHubHits, HNHits, AND ResearchFindings
- [ ] Each finding is published as Intermute Insight linked to sprint's Spec
- [ ] Confidence.Research reflects finding count, source diversity, and relevance (not binary)
- [ ] Zero results from scan → Research confidence ≈ 0.0 (not 1.0)
- [ ] `StartWithResearch(pollardPath)` loads existing `.pollard/` insights
- [ ] Pre-loaded research skips quick scan trigger
- [ ] Deep scan handoff sends async request via Intermute messaging
- [ ] Deep scan results can be imported back into sprint
- [ ] Sprint TUI displays research findings when available
- [ ] Quick scan failure does not block sprint progression

### Non-Functional
- [ ] Quick scan completes within 30 seconds
- [ ] No import cycles between arbiter and pollard packages
- [ ] All existing tests continue to pass
- [ ] Orchestrator works without ResearchProvider (nil = no-research mode)

### Security (from security-sentinel review)
- [ ] Sprint Problem text sanitized before use as Intermute Spec title (no newlines, max 200 chars)
- [ ] External API responses not passed as raw CLI arguments in synthesizer.go

### Agent-Native (from agent-native review — follow-up work)
- [ ] **Future task**: Add CLI commands: `gurgeh sprint start`, `sprint accept`, `sprint advance`, `sprint state`, `sprint handoff`
- [ ] **Future task**: Wire `autarch_research` MCP tool to real Pollard scan
- [ ] **Future task**: Implement `--from-research` CLI flag for `gurgeh sprint new`

---

## Dependencies & Risks

| Dependency | Risk | Mitigation |
|-----------|------|------------|
| Intermute server running | Medium — embedded mode avoids external dependency for tests | Use `pkg/embedded` in tests and integration |
| Pollard hunters need network | Low — quick scan already handles failures gracefully | Continue swallowing errors, sprint proceeds |
| `go.sum` changes from Intermute client import | Low — already a dependency | Run `go mod tidy` |
| Real quick scanner imports arbiter types | None — one-way dependency, already exists | No action needed |

---

## References

- Orchestrator: `internal/gurgeh/arbiter/orchestrator.go`
- Types: `internal/gurgeh/arbiter/types.go`
- Real quick scan: `internal/pollard/quick/scan.go`
- Stub quick scan: `internal/gurgeh/arbiter/quick/scanner.go` (to be deleted)
- Pollard coordinator: `internal/pollard/research/coordinator.go`
- Intermute client: `github.com/mistakeknot/intermute/client`
- Intermute embedded: `github.com/mistakeknot/intermute/pkg/embedded`
- Architecture learnings: `docs/solutions/patterns/arbiter-spec-sprint-architecture.md`
- TUI dimension gotcha: `docs/solutions/ui-bugs/tui-dimension-mismatch-splitlayout-20260126.md`
