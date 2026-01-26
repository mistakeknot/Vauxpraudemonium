# Pollard - Development Guide

> Continuous research intelligence with pluggable hunters

Named after Cayce Pollard from William Gibson's *Pattern Recognition* - someone with an uncanny ability to recognize patterns and trends before they become obvious.

## Quick Reference

| Item | Value |
|------|-------|
| Entry Point | `cmd/pollard/main.go` |
| CLI Framework | Cobra |
| Data Directory | `.pollard/` (per-project) |
| State Database | `.pollard/state.db` (SQLite) |

```bash
# Initialize Pollard in a project
go run ./cmd/pollard init

# Run all enabled hunters
go run ./cmd/pollard scan

# Run specific hunter
go run ./cmd/pollard scan --hunter github-scout

# Generate research report
go run ./cmd/pollard report
go run ./cmd/pollard report --type competitive
```

---

## Key Paths

| Path | Purpose |
|------|---------|
| `cmd/pollard/` | Entry point |
| `internal/pollard/api/` | Programmatic API (Scanner, Orchestrator) |
| `internal/pollard/cli/` | CLI commands (Cobra) |
| `internal/pollard/config/` | YAML configuration |
| `internal/pollard/hunters/` | Research agents (GitHub, arXiv, etc.) |
| `internal/pollard/insights/` | Synthesized findings |
| `internal/pollard/intermute/` | Intermute bridge (Publisher) |
| `internal/pollard/patterns/` | Implementation patterns |
| `internal/pollard/proposal/` | Research agenda proposals |
| `internal/pollard/reports/` | Markdown report generation |
| `internal/pollard/research/` | Research coordinator |
| `internal/pollard/sources/` | Raw collected data |
| `internal/pollard/state/` | SQLite state management |
| `internal/pollard/weaver/` | Insight synthesis |

---

## Architecture

```
                    ┌─────────────────────────────┐
                    │        Pollard CLI          │
                    │   (scan, report, propose)   │
                    └──────────┬──────────────────┘
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                ▼
    ┌─────────────────┐ ┌───────────┐ ┌─────────────────┐
    │     Scanner     │ │ Reporter  │ │    Proposer     │
    │  (api/scanner)  │ │           │ │  (agenda gen)   │
    └────────┬────────┘ └─────┬─────┘ └────────┬────────┘
             │                │                │
             ▼                │                │
    ┌─────────────────┐       │                │
    │  Orchestrator   │       │                │
    │ (intelligent    │       │                │
    │  hunter select) │       │                │
    └────────┬────────┘       │                │
             │                │                │
    ┌────────┴────────────────┴────────────────┴────────┐
    │                    Registry                        │
    │     github | arxiv | hackernews | openalex | ...  │
    └────────────────────────┬──────────────────────────┘
                             │
                    ┌────────┴────────┐
                    ▼                 ▼
              ┌──────────┐     ┌───────────┐
              │ Sources  │     │  Insights │
              │ (raw)    │     │(synthesis)│
              └──────────┘     └───────────┘
```

---

## Hunter System

### Hunter Interface

Every hunter implements this interface:

```go
// internal/pollard/hunters/hunter.go
type Hunter interface {
    // Name returns the hunter's identifier (e.g., "github-scout")
    Name() string

    // Hunt performs the research collection
    Hunt(ctx context.Context, cfg HunterConfig) (*HuntResult, error)
}
```

### HunterConfig

Configuration passed to each hunt:

```go
type HunterConfig struct {
    Queries     []string           // Search queries
    MaxResults  int                // Results per query
    MinStars    int                // GitHub: minimum stars
    MinPoints   int                // HackerNews: minimum points
    Categories  []string           // arXiv: category filters
    Targets     []CompetitorTarget // Competitor tracking targets
    OutputDir   string             // Where to write outputs
    ProjectPath string             // Root project path
    APIToken    string             // Optional auth token
    Mode        string             // quick, balanced, deep
    Pipeline    PipelineOptions    // Fetch/synthesize options
}
```

### HuntResult

Results returned from each hunt:

```go
type HuntResult struct {
    HunterName       string
    StartedAt        time.Time
    CompletedAt      time.Time
    SourcesCollected int       // Raw items found
    InsightsCreated  int       // Synthesized insights
    OutputFiles      []string  // Files written
    Errors           []error
}
```

### Registry

All hunters register with the default registry:

```go
// internal/pollard/hunters/hunter.go
func DefaultRegistry() *Registry {
    reg := NewRegistry()
    // Tech-focused (enabled by default)
    reg.Register(NewGitHubScout())
    reg.Register(NewHackerNewsHunter())
    reg.Register(NewArxivHunter())
    reg.Register(NewCompetitorTracker())
    // General-purpose (disabled by default)
    reg.Register(NewOpenAlexHunter())
    reg.Register(NewPubMedHunter())
    reg.Register(NewUSDAHunter())
    reg.Register(NewLegalHunter())
    reg.Register(NewEconomicsHunter())
    reg.Register(NewWikiHunter())
    // Agent-native
    reg.Register(NewAgentHunter())
    return reg
}
```

---

## Adding a New Hunter

### 1. Create Hunter File

```go
// internal/pollard/hunters/myapi.go
package hunters

import (
    "context"
    "fmt"
)

type MyAPIHunter struct {
    rateLimiter *RateLimiter
}

func NewMyAPIHunter() *MyAPIHunter {
    return &MyAPIHunter{
        rateLimiter: NewRateLimiter(10, time.Second, false), // 10 req/s
    }
}

func (h *MyAPIHunter) Name() string {
    return "myapi"
}

func (h *MyAPIHunter) Hunt(ctx context.Context, cfg HunterConfig) (*HuntResult, error) {
    result := &HuntResult{
        HunterName: h.Name(),
        StartedAt:  time.Now(),
    }

    for _, query := range cfg.Queries {
        // Rate limit
        if err := h.rateLimiter.Wait(ctx); err != nil {
            result.Errors = append(result.Errors, err)
            continue
        }

        // Make API call
        items, err := h.search(ctx, query, cfg.MaxResults)
        if err != nil {
            result.Errors = append(result.Errors, err)
            continue
        }

        // Save sources
        for _, item := range items {
            if err := h.saveSource(cfg, item); err != nil {
                result.Errors = append(result.Errors, err)
                continue
            }
            result.SourcesCollected++
        }
    }

    result.CompletedAt = time.Now()
    return result, nil
}

func (h *MyAPIHunter) search(ctx context.Context, query string, limit int) ([]Item, error) {
    // Implement API call
}

func (h *MyAPIHunter) saveSource(cfg HunterConfig, item Item) error {
    // Save to .pollard/sources/myapi/
}
```

### 2. Register in Default Registry

```go
// internal/pollard/hunters/hunter.go
func DefaultRegistry() *Registry {
    reg := NewRegistry()
    // ... existing hunters
    reg.Register(NewMyAPIHunter())  // Add here
    return reg
}
```

### 3. Add to Config Schema

```go
// internal/pollard/config/config.go
type Config struct {
    // ... existing
    MyAPI HunterConfigEntry `yaml:"myapi,omitempty"`
}
```

### 4. Add CLI Support (optional)

```go
// internal/pollard/cli/scan.go
// The hunter will automatically be available via:
// pollard scan --hunter myapi
```

---

## Configuration

### Project Config (`.pollard/config.yaml`)

```yaml
# Tech-focused hunters (enabled by default)
github-scout:
  enabled: true
  queries:
    - "task orchestration golang"
    - "AI agent coordination"
  max_results: 50
  min_stars: 100
  output: sources/github

hackernews:
  enabled: true
  queries:
    - "AI coding assistant"
  min_points: 50
  output: sources/hackernews

arxiv:
  enabled: true
  queries:
    - "large language model agents"
  categories:
    - cs.AI
    - cs.CL
  max_results: 30
  output: sources/research

competitor-tracker:
  enabled: true
  targets:
    - name: "Cursor"
      changelog: "https://cursor.com/changelog"
      github: "getcursor/cursor"
    - name: "Codeium"
      docs: "https://codeium.com/docs"
  output: insights/competitive

# General-purpose hunters (disabled by default)
openalex:
  enabled: false
  queries: []
  output: sources/openalex

pubmed:
  enabled: false
  queries: []
  output: sources/pubmed

usda-nutrition:
  enabled: false
  queries: []
  output: sources/nutrition

legal:
  enabled: false
  queries: []
  output: sources/legal

economics:
  enabled: false
  queries: []
  output: sources/economics

wiki:
  enabled: false
  queries: []
  output: sources/wiki
```

---

## Programmatic API

The Scanner API allows other tools (Gurgeh, Coldwine) to trigger research:

### Basic Usage

```go
import "github.com/mistakeknot/autarch/internal/pollard/api"

scanner, err := api.NewScanner(projectPath)
if err != nil {
    return err
}
defer scanner.Close()

// Run all enabled hunters
result, err := scanner.Scan(ctx, api.ScanOptions{})

// Run specific hunters
result, err := scanner.Scan(ctx, api.ScanOptions{
    Hunters: []string{"github-scout", "hackernews"},
    Queries: []string{"custom query"},
})
```

### PRD-Focused Research

```go
// Called from Gurgeh after PRD creation
result, err := scanner.ResearchForPRD(ctx, vision, problem, requirements)

// Called from Gurgeh during persona interview
result, err := scanner.ResearchUserPersonas(ctx, personas, painpoints)
```

### Epic-Focused Research

```go
// Called from Coldwine for epic planning
result, err := scanner.ResearchForEpic(ctx, epicTitle, description)
```

### Feature Insights

```go
// Get insights linked to a specific feature
insights, err := scanner.GetInsightsForFeature(ctx, "FEAT-001")

// Generate a research brief for agent context
brief, err := scanner.GenerateResearchBrief(ctx, "FEAT-001")
```

### Intelligent Research (Agent-Native)

```go
// Uses agent + selective hunters based on content analysis
result, err := scanner.IntelligentResearch(ctx, vision, problem, requirements)

// Get hunter recommendations
selections := scanner.SuggestHunters(vision, problem, requirements)
for _, sel := range selections {
    fmt.Printf("%s (score: %.2f): %s\n", sel.Name, sel.Score, sel.Reasoning)
}

// Check if a custom hunter would help
domain, needed := scanner.SuggestNewHunter(vision, problem, requirements)
if needed {
    spec, _ := scanner.CreateCustomHunter(ctx, domain, contextInfo)
}
```

---

## Intermute Integration

Pollard integrates with Intermute for cross-tool coordination:

### Publisher

```go
// internal/pollard/intermute/publisher.go
pub := intermute.NewPublisher(client, "autarch")

// Optionally link to a spec
pub = pub.WithSpecID("spec-123")

// Publish a single finding as Intermute Insight
insight, err := pub.PublishFinding(ctx, finding)

// Publish multiple findings
insights, err := pub.PublishFindings(ctx, findings)
```

### Category Mapping

Pollard tags map to Intermute insight categories:

| Pollard Tag | Intermute Category |
|-------------|-------------------|
| `competitive` | `competitive` |
| `trend` | `trends` |
| `user` | `user` |
| (default) | `research` |

### Message-Based Research Requests

Pollard can receive research requests via Intermute messages:

```go
// Other tools send requests
api.SendResearchRequest(projectPath, api.ResearchPayload{
    RequestType: "prd",
    Vision:      "Build a task orchestration tool",
    Problem:     "Agents can't coordinate",
    Queries:     []string{"agent coordination"},
}, "gurgeh")

// Pollard processes its inbox
scanner.ProcessInbox(ctx)

// Or wait for response
response, err := api.WaitForResponse(projectPath, msgID, 5*time.Minute)
```

### Graceful Degradation

All Intermute integrations work without Intermute configured:

```go
pub := intermute.NewPublisher(nil, "autarch")  // nil client
insight, err := pub.PublishFinding(ctx, finding)
// Returns empty insight, nil error - no-op
```

---

## State Management

### SQLite Database (`.pollard/state.db`)

Tracks run history and freshness:

```go
// internal/pollard/state/db.go
type DB struct {
    db *sql.DB
}

// Record run start
runID, err := db.StartRun("github-scout")

// Record run completion
db.CompleteRun(runID, success, sourcesCount, insightsCount, errorMsg)

// Get last run for a hunter
run, err := db.LastRun("github-scout")

// Check freshness
stale := db.IsStale("github-scout", 24*time.Hour)
```

### Schema

```sql
CREATE TABLE runs (
    id INTEGER PRIMARY KEY,
    hunter TEXT NOT NULL,
    started_at DATETIME NOT NULL,
    completed_at DATETIME,
    success BOOLEAN,
    sources_count INTEGER DEFAULT 0,
    insights_count INTEGER DEFAULT 0,
    error_message TEXT
);

CREATE INDEX idx_runs_hunter ON runs(hunter);
CREATE INDEX idx_runs_started ON runs(started_at);
```

---

## Report Generation

### Report Types

| Type | Description | Hunters Used |
|------|-------------|--------------|
| `landscape` | Full landscape overview | All |
| `competitive` | Competitor analysis | competitor-tracker |
| `trends` | Industry trends | hackernews |
| `research` | Academic papers | arxiv, openalex, pubmed |

### Usage

```bash
# Generate landscape report (default)
pollard report

# Specific report type
pollard report --type competitive

# Output to stdout instead of file
pollard report --stdout

# Custom output path
pollard report --output custom-report.md
```

### Programmatic

```go
// internal/pollard/reports/generator.go
gen := reports.NewGenerator(projectPath)
report, err := gen.Generate(ctx, reports.TypeLandscape)

// Write to file
err = gen.WriteReport(report, outputPath)
```

---

## Testing

### Unit Tests

```bash
# Test hunters
go test ./internal/pollard/hunters -v

# Test API
go test ./internal/pollard/api -v

# Test config
go test ./internal/pollard/config -v
```

### Integration Tests

```bash
# Initialize test project
mkdir /tmp/pollard-test && cd /tmp/pollard-test
go run ./cmd/pollard init

# Run scan (dry-run first)
go run ./cmd/pollard scan --dry-run

# Run actual scan (requires API access)
go run ./cmd/pollard scan --hunter hackernews

# Generate report
go run ./cmd/pollard report --stdout
```

### Mock API Testing

For CI/CD, use mock responses:

```go
// internal/pollard/hunters/github_test.go
func TestGitHubScout_Hunt(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(mockSearchResponse)
    }))
    defer server.Close()

    hunter := NewGitHubScoutWithURL(server.URL)
    result, err := hunter.Hunt(ctx, cfg)
    // Assert results
}
```

---

## Environment Variables

| Variable | Hunter | Required | Description |
|----------|--------|----------|-------------|
| `GITHUB_TOKEN` | github-scout | No | Higher rate limits (5000/hr vs 60/hr) |
| `POLLARD_GITHUB_TOKEN` | github-scout | No | Alternative to GITHUB_TOKEN |
| `OPENALEX_EMAIL` | openalex | No | Polite pool access (100k/day) |
| `NCBI_API_KEY` | pubmed | No | Higher rate limits (10/s vs 3/s) |
| `USDA_API_KEY` | usda-nutrition | **Yes** | Required for USDA access |
| `COURTLISTENER_API_KEY` | legal | **Yes** | Required for CourtListener |

---

## Troubleshooting

### Hunter Not Found

```bash
# List available hunters
go run ./cmd/pollard hunter list

# Check if registered
grep -r "reg.Register" internal/pollard/hunters/
```

### Rate Limited

Check the rate limit settings in each hunter:

```go
// Unauthenticated GitHub: 60 req/hr
rateLimiter: NewRateLimiter(60, time.Hour, false)

// Authenticated GitHub: 5000 req/hr
rateLimiter: NewRateLimiter(5000, time.Hour, true)
```

### No Results

- Check query syntax for the specific API
- Verify API is accessible (network, auth)
- Check `--dry-run` output for what would be searched
- Look at `.pollard/sources/` for raw outputs

### Config Not Loading

```bash
# Validate YAML
cat .pollard/config.yaml | yq

# Check for syntax errors
go run ./cmd/pollard status
```

---

## Related Documentation

- [HUNTERS.md](./HUNTERS.md) - Complete hunter reference
- [API.md](./API.md) - Programmatic API reference
- Root [AGENTS.md](../../AGENTS.md) - Full Autarch development guide
