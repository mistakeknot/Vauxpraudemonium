# Pollard API Reference

> Programmatic interface for running Pollard hunters from other tools

This document covers the Scanner API for integrating Pollard research capabilities into Gurgeh, Coldwine, and other tools.

---

## Quick Start

```go
import "github.com/mistakeknot/autarch/internal/pollard/api"

// Create scanner
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
    Queries: []string{"task orchestration golang"},
})
```

---

## Scanner

The `Scanner` is the main entry point for programmatic access to Pollard.

### Creating a Scanner

```go
scanner, err := api.NewScanner(projectPath)
if err != nil {
    return fmt.Errorf("failed to create scanner: %w", err)
}
defer scanner.Close()
```

The scanner:
- Loads configuration from `.pollard/config.yaml`
- Creates required directories under `.pollard/`
- Opens the SQLite state database
- Initializes the hunter registry

### Closing

Always close the scanner to release resources:

```go
defer scanner.Close()
```

---

## Core Methods

### Scan

Run hunters with configurable options.

```go
func (s *Scanner) Scan(ctx context.Context, opts ScanOptions) (*ScanResult, error)
```

**ScanOptions:**

```go
type ScanOptions struct {
    // Hunters to run (empty = all enabled hunters)
    Hunters []string

    // Queries to search (overrides config)
    Queries []string

    // Targets for competitor tracking
    Targets []CompetitorTarget

    // Max results per query
    MaxResults int
}
```

**Example - Run all enabled hunters:**

```go
result, err := scanner.Scan(ctx, api.ScanOptions{})
```

**Example - Run specific hunters with custom queries:**

```go
result, err := scanner.Scan(ctx, api.ScanOptions{
    Hunters:    []string{"github-scout", "arxiv"},
    Queries:    []string{"LLM agents", "code generation"},
    MaxResults: 30,
})
```

### ScanResult

```go
type ScanResult struct {
    // Per-hunter results
    HunterResults map[string]*hunters.HuntResult

    // Totals
    TotalSources  int
    TotalInsights int

    // Files written
    OutputFiles []string

    // Any errors encountered
    Errors []error
}
```

---

## Specialized Scan Methods

### ScanGitHub

Run only the GitHub Scout hunter.

```go
func (s *Scanner) ScanGitHub(ctx context.Context, queries []string, maxResults int) (*hunters.HuntResult, error)
```

**Example:**

```go
result, err := scanner.ScanGitHub(ctx, []string{
    "task orchestration golang",
    "AI agent coordination",
}, 50)
```

### ScanTrends

Run only the HackerNews TrendWatcher hunter.

```go
func (s *Scanner) ScanTrends(ctx context.Context, queries []string, minPoints int) (*hunters.HuntResult, error)
```

**Example:**

```go
result, err := scanner.ScanTrends(ctx, []string{
    "AI coding assistant",
    "developer tools",
}, 100) // Minimum 100 points
```

### ScanResearch

Run only the arXiv ResearchScout hunter.

```go
func (s *Scanner) ScanResearch(ctx context.Context, queries []string, categories []string) (*hunters.HuntResult, error)
```

**Example:**

```go
result, err := scanner.ScanResearch(ctx,
    []string{"large language model agents"},
    []string{"cs.AI", "cs.CL", "cs.SE"},
)
```

### ScanCompetitors

Run only the Competitor Tracker hunter.

```go
func (s *Scanner) ScanCompetitors(ctx context.Context, targets []CompetitorTarget) (*hunters.HuntResult, error)
```

**Example:**

```go
result, err := scanner.ScanCompetitors(ctx, []api.CompetitorTarget{
    {
        Name:      "Cursor",
        Changelog: "https://cursor.com/changelog",
        GitHub:    "getcursor/cursor",
    },
    {
        Name:      "Codeium",
        Docs:      "https://codeium.com/docs",
    },
})
```

---

## PRD-Focused Research

These methods are designed for integration with Gurgeh.

### ResearchForPRD

Comprehensive research based on PRD content.

```go
func (s *Scanner) ResearchForPRD(ctx context.Context, vision, problem string, requirements []string) (*ScanResult, error)
```

**Example (from Gurgeh after PRD creation):**

```go
result, err := scanner.ResearchForPRD(ctx,
    "Build a task orchestration tool for AI agents",
    "AI agents can't coordinate effectively on complex tasks",
    []string{
        "Support multiple concurrent agents",
        "Git worktree isolation",
        "Real-time status tracking",
    },
)
```

**Behavior:**
- Generates queries from vision, problem, and requirements
- Limits to 5 most relevant queries
- Runs: `github-scout`, `hackernews`, `arxiv`
- Returns 20 results per query

### ResearchUserPersonas

Research based on user personas and pain points.

```go
func (s *Scanner) ResearchUserPersonas(ctx context.Context, personas []string, painpoints []string) (*ScanResult, error)
```

**Example (from Gurgeh interview flow):**

```go
result, err := scanner.ResearchUserPersonas(ctx,
    []string{"developer", "technical lead"},
    []string{"slow code reviews", "context switching"},
)
```

**Behavior:**
- Generates queries like "developer tools", "developer workflow", "slow code reviews solution"
- Runs: `github-scout`, `hackernews`

---

## Epic-Focused Research

These methods are designed for integration with Coldwine.

### ResearchForEpic

Implementation-focused research for epic planning.

```go
func (s *Scanner) ResearchForEpic(ctx context.Context, epicTitle, description string) (*ScanResult, error)
```

**Example (from Coldwine during epic creation):**

```go
result, err := scanner.ResearchForEpic(ctx,
    "User Authentication System",
    "Implement OAuth2 and JWT-based authentication with role-based access control",
)
```

**Behavior:**
- Generates implementation-focused queries
- Runs: `github-scout` only
- Returns 30 results per query

### GetInsightsForFeature

Get insights linked to a specific feature.

```go
func (s *Scanner) GetInsightsForFeature(ctx context.Context, featureRef string) ([]*insights.Insight, error)
```

**Example (from Coldwine when showing story details):**

```go
insights, err := scanner.GetInsightsForFeature(ctx, "FEAT-001")

for _, insight := range insights {
    for _, finding := range insight.Findings {
        fmt.Printf("- %s: %s\n", finding.Title, finding.Description)
    }
}
```

### GenerateResearchBrief

Create a markdown summary for agent context.

```go
func (s *Scanner) GenerateResearchBrief(ctx context.Context, featureRef string) (string, error)
```

**Example (from Coldwine when assigning work):**

```go
brief, err := scanner.GenerateResearchBrief(ctx, "FEAT-001")
if err != nil {
    return err
}

// Attach to agent mail message
message := fmt.Sprintf("## Task Assignment\n\n%s\n\n%s", task.Description, brief)
```

**Output format:**

```markdown
## Research Context

### Key Insights (HIGH relevance)
- **OAuth2 Best Practices**: Use PKCE for public clients...
- **JWT Security**: Rotate signing keys regularly...

### Additional Context (MEDIUM relevance)
- **Session Management**: Consider Redis for distributed sessions...

### Recommendations
- **Feature: Token Refresh** (high): Implement refresh token rotation...
```

---

## Agent-Native Intelligent Research

These methods use the user's AI agent as the primary research capability.

### IntelligentResearch

Agent-driven research with optional API supplements.

```go
func (s *Scanner) IntelligentResearch(ctx context.Context, vision, problem string, requirements []string) (*ScanResult, error)
```

**Example:**

```go
result, err := scanner.IntelligentResearch(ctx,
    "Build a nutrition tracking app",
    "Users struggle to track nutrient intake accurately",
    []string{"Food barcode scanning", "Meal logging", "Nutrient analysis"},
)
```

**Behavior:**
1. Generates a research brief
2. Runs agent-based research (PRIMARY)
3. Supplements with API hunters if keys available
4. Runs any matching custom hunters

### SuggestHunters

Get recommended hunters based on content analysis.

```go
func (s *Scanner) SuggestHunters(vision, problem string, requirements []string) []HunterSelection
```

**HunterSelection:**

```go
type HunterSelection struct {
    Name      string   // Hunter name
    Score     float64  // Relevance score (0-1)
    Queries   []string // Suggested queries
    Domain    string   // Domain category
    Reasoning string   // Why this hunter
}
```

**Example:**

```go
selections := scanner.SuggestHunters(vision, problem, requirements)
for _, sel := range selections {
    fmt.Printf("%s (%.2f): %s\n", sel.Name, sel.Score, sel.Reasoning)
}
// Output:
// pubmed (0.85): Medical/health domain detected
// usda-nutrition (0.72): Food/nutrition focus
// github-scout (0.65): OSS implementations
```

### SuggestNewHunter

Check if a custom hunter would be beneficial.

```go
func (s *Scanner) SuggestNewHunter(vision, problem string, requirements []string) (string, bool)
```

**Example:**

```go
domain, needed := scanner.SuggestNewHunter(vision, problem, requirements)
if needed {
    fmt.Printf("Consider creating a custom hunter for: %s\n", domain)
}
```

### CreateCustomHunter

Design a new hunter for a domain using the AI agent.

```go
func (s *Scanner) CreateCustomHunter(ctx context.Context, domain, contextInfo string) (*CustomHunterSpec, error)
```

**CustomHunterSpec:**

```go
type CustomHunterSpec struct {
    Name           string
    Description    string
    APIEndpoint    string
    NoAPI          bool   // True if no API available
    Recommendation string // Alternative approach if no API
}
```

### GetResearchBrief

Generate a research brief as a prompt.

```go
func (s *Scanner) GetResearchBrief(vision, problem string, requirements []string) string
```

---

## Intermute Message Handling

Pollard can receive research requests via Intermute messages.

### Message Types

```go
const (
    TypeResearchRequest  = "research_request"
    TypeResearchComplete = "research_complete"
    TypeScanRequest      = "scan_request"
    TypeScanComplete     = "scan_complete"
)
```

### ResearchPayload

```go
type ResearchPayload struct {
    RequestType string   // prd, epic, persona, general
    SourceID    string   // PRD-001, EPIC-001
    SourceType  string   // prd, epic, feature
    Queries     []string
    Personas    []string
    Painpoints  []string
    Vision      string
    Problem     string
    Hunters     []string
    MaxResults  int
    Targets     []CompetitorTarget
}
```

### Sending Research Requests

From other tools (Gurgeh, Coldwine):

```go
msg, err := api.SendResearchRequest(projectPath, api.ResearchPayload{
    RequestType: "prd",
    SourceID:    "PRD-001",
    Vision:      "Build a task orchestration tool",
    Problem:     "Agents can't coordinate",
    Queries:     []string{"agent coordination"},
}, "gurgeh")

if err != nil {
    return err
}
```

### Processing the Inbox

From Pollard (typically run as a background service):

```go
scanner.ProcessInbox(ctx)
```

### Waiting for Responses

```go
response, err := api.WaitForResponse(projectPath, msg.ID, 5*time.Minute)
if err != nil {
    return err
}

fmt.Printf("Research complete: %d sources, %d insights\n",
    response.Response.TotalSources,
    response.Response.TotalInsights)
```

---

## Result Types

### HuntResult

Returned by individual hunters:

```go
type HuntResult struct {
    HunterName       string
    StartedAt        time.Time
    CompletedAt      time.Time
    SourcesCollected int
    InsightsCreated  int
    OutputFiles      []string
    Errors           []error
}

func (r *HuntResult) Success() bool {
    return len(r.Errors) == 0
}

func (r *HuntResult) Duration() time.Duration {
    return r.CompletedAt.Sub(r.StartedAt)
}
```

### CompetitorTarget

```go
type CompetitorTarget struct {
    Name      string // Company/product name
    Changelog string // Changelog URL
    Docs      string // Documentation URL
    GitHub    string // GitHub repo (owner/repo)
}
```

---

## Environment Variables

| Variable | Purpose | Required |
|----------|---------|----------|
| `GITHUB_TOKEN` | GitHub API access | No (rate limited without) |
| `OPENALEX_EMAIL` | OpenAlex polite pool | No |
| `NCBI_API_KEY` | PubMed faster access | No |
| `USDA_API_KEY` | USDA nutrition data | **Yes** for usda-nutrition |
| `COURTLISTENER_API_KEY` | Legal hunter | **Yes** for legal |
| `INTERMUTE_URL` | Intermute server | No |
| `INTERMUTE_API_KEY` | Intermute auth | No |
| `INTERMUTE_PROJECT` | Intermute scope | Required if API key set |

---

## Error Handling

### Graceful Degradation

All methods handle missing API keys gracefully:

```go
result, err := scanner.Scan(ctx, api.ScanOptions{
    Hunters: []string{"usda-nutrition"},
})
// If USDA_API_KEY not set, hunter skips silently
```

### Checking for Errors

```go
result, err := scanner.Scan(ctx, opts)
if err != nil {
    // Fatal error (e.g., context canceled)
    return err
}

// Non-fatal errors are collected
for _, err := range result.Errors {
    log.Printf("Warning: %v", err)
}

// Check specific hunter results
if ghResult, ok := result.HunterResults["github-scout"]; ok {
    if !ghResult.Success() {
        log.Printf("GitHub scout had errors: %v", ghResult.Errors)
    }
}
```

---

## Integration Examples

### Gurgeh Integration

```go
// After PRD creation
func (g *Gurgeh) runPostPRDResearch(ctx context.Context, prd *Spec) error {
    scanner, err := api.NewScanner(g.projectPath)
    if err != nil {
        return err
    }
    defer scanner.Close()

    result, err := scanner.ResearchForPRD(ctx,
        prd.Vision,
        prd.Problem,
        prd.RequirementsList(),
    )
    if err != nil {
        return err
    }

    log.Printf("Research complete: %d sources found", result.TotalSources)
    return nil
}
```

### Coldwine Integration

```go
// When generating epic from PRD
func (c *Coldwine) enrichEpicWithResearch(ctx context.Context, epic *Epic) error {
    scanner, err := api.NewScanner(c.projectPath)
    if err != nil {
        return err
    }
    defer scanner.Close()

    // Get linked insights
    insights, err := scanner.GetInsightsForFeature(ctx, epic.FeatureRef)
    if err != nil {
        return err
    }

    // Generate brief for agent context
    brief, err := scanner.GenerateResearchBrief(ctx, epic.FeatureRef)
    if err != nil {
        return err
    }

    epic.ResearchBrief = brief
    epic.InsightCount = len(insights)
    return nil
}
```

---

## Related Documentation

- [AGENTS.md](./AGENTS.md) - Pollard developer guide
- [HUNTERS.md](./HUNTERS.md) - Hunter reference
- [../INTEGRATION.md](../INTEGRATION.md) - Cross-tool integration
