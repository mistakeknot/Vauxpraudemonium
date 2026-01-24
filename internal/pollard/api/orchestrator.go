package api

import (
	"context"
	"fmt"
	"os"

	"github.com/mistakeknot/vauxpraudemonium/internal/pollard/hunters"
	"github.com/mistakeknot/vauxpraudemonium/internal/pollard/selector"
)

// ResearchOrchestrator coordinates agent-driven research with optional API supplements.
type ResearchOrchestrator struct {
	scanner     *Scanner
	briefGen    *selector.BriefGenerator
	selector    *selector.Selector
	agentHunter *hunters.AgentHunter
}

// NewResearchOrchestrator creates an orchestrator.
func NewResearchOrchestrator(scanner *Scanner) *ResearchOrchestrator {
	return &ResearchOrchestrator{
		scanner:     scanner,
		briefGen:    selector.NewBriefGenerator(),
		selector:    selector.NewSelector(),
		agentHunter: hunters.NewAgentHunter(),
	}
}

// Research conducts intelligent research for a PRD.
// The primary research is done by the user's AI agent.
// API hunters are used as optional supplements if keys are available.
func (o *ResearchOrchestrator) Research(ctx context.Context, vision, problem string, requirements []string) (*ScanResult, error) {
	result := &ScanResult{
		HunterResults: make(map[string]*hunters.HuntResult),
	}

	// 1. Generate research brief
	brief := o.briefGen.GenerateForPRD(vision, problem, requirements)

	// 2. Run agent-based research (PRIMARY)
	agentCfg := hunters.HunterConfig{
		Queries:     brief.Questions,
		ProjectPath: o.scanner.projectPath,
	}

	agentResult, err := o.agentHunter.Hunt(ctx, agentCfg)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("agent research: %w", err))
	} else {
		result.HunterResults["agent-research"] = agentResult
		result.TotalSources += agentResult.SourcesCollected
		result.OutputFiles = append(result.OutputFiles, agentResult.OutputFiles...)
		if len(agentResult.Errors) > 0 {
			result.Errors = append(result.Errors, agentResult.Errors...)
		}
	}

	// 3. Optionally supplement with API hunters (if keys available)
	for _, hunterName := range brief.SuggestedAPIs {
		if !o.hasAPIKey(hunterName) {
			continue // Skip if no API key
		}

		hunterResult, err := o.runAPIHunter(ctx, hunterName, brief)
		if err != nil {
			result.Errors = append(result.Errors, err)
			continue
		}

		result.HunterResults[hunterName] = hunterResult
		result.TotalSources += hunterResult.SourcesCollected
		result.OutputFiles = append(result.OutputFiles, hunterResult.OutputFiles...)
	}

	// 4. Also run any custom hunters that match the domain
	customResults, customErrs := o.runCustomHunters(ctx, brief)
	for name, hr := range customResults {
		result.HunterResults[name] = hr
		result.TotalSources += hr.SourcesCollected
		result.OutputFiles = append(result.OutputFiles, hr.OutputFiles...)
	}
	result.Errors = append(result.Errors, customErrs...)

	return result, nil
}

// hasAPIKey checks if the required API key is configured.
func (o *ResearchOrchestrator) hasAPIKey(hunter string) bool {
	switch hunter {
	case "pubmed":
		// PubMed works without key, but faster with one
		return true
	case "usda-nutrition":
		return os.Getenv("USDA_API_KEY") != ""
	case "legal":
		return os.Getenv("COURTLISTENER_API_KEY") != ""
	case "github-scout":
		// Works without token but with lower rate limit
		return true
	case "openalex":
		// No key required, email optional for higher rate limit
		return true
	case "economics", "wiki", "hackernews", "arxiv":
		// No key required
		return true
	default:
		return false
	}
}

// runAPIHunter executes an API-based hunter.
func (o *ResearchOrchestrator) runAPIHunter(ctx context.Context, name string, brief *selector.ResearchBrief) (*hunters.HuntResult, error) {
	hunter, ok := o.scanner.registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("hunter %s not found", name)
	}

	// Build queries from brief questions
	queries := make([]string, 0)
	for i, q := range brief.Questions {
		if i >= 3 {
			break // Limit API queries
		}
		queries = append(queries, q)
	}

	cfg := hunters.HunterConfig{
		Queries:     queries,
		MaxResults:  20, // Limit API calls
		ProjectPath: o.scanner.projectPath,
	}

	return hunter.Hunt(ctx, cfg)
}

// runCustomHunters executes any matching custom hunters.
func (o *ResearchOrchestrator) runCustomHunters(ctx context.Context, brief *selector.ResearchBrief) (map[string]*hunters.HuntResult, []error) {
	results := make(map[string]*hunters.HuntResult)
	var errors []error

	customHunters, err := hunters.LoadCustomHunters(o.scanner.projectPath)
	if err != nil {
		errors = append(errors, fmt.Errorf("load custom hunters: %w", err))
		return results, errors
	}

	for _, ch := range customHunters {
		cfg := hunters.HunterConfig{
			Queries:     brief.Questions[:min(3, len(brief.Questions))],
			MaxResults:  20,
			ProjectPath: o.scanner.projectPath,
		}

		result, err := ch.Hunt(ctx, cfg)
		if err != nil {
			errors = append(errors, fmt.Errorf("custom hunter %s: %w", ch.Name(), err))
			continue
		}

		results[ch.Name()] = result
	}

	return results, errors
}

// ResearchForEpic conducts research tailored for implementation.
func (o *ResearchOrchestrator) ResearchForEpic(ctx context.Context, title, description string) (*ScanResult, error) {
	result := &ScanResult{
		HunterResults: make(map[string]*hunters.HuntResult),
	}

	// Generate epic-focused brief
	brief := o.briefGen.GenerateForEpic(title, description)

	// Run agent research
	agentCfg := hunters.HunterConfig{
		Queries:     brief.Questions,
		ProjectPath: o.scanner.projectPath,
	}

	agentResult, err := o.agentHunter.Hunt(ctx, agentCfg)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("agent research: %w", err))
	} else {
		result.HunterResults["agent-research"] = agentResult
		result.TotalSources += agentResult.SourcesCollected
		result.OutputFiles = append(result.OutputFiles, agentResult.OutputFiles...)
	}

	// For epics, also run GitHub scout if available
	if o.hasAPIKey("github-scout") {
		ghResult, err := o.runAPIHunter(ctx, "github-scout", brief)
		if err != nil {
			result.Errors = append(result.Errors, err)
		} else if ghResult != nil {
			result.HunterResults["github-scout"] = ghResult
			result.TotalSources += ghResult.SourcesCollected
			result.OutputFiles = append(result.OutputFiles, ghResult.OutputFiles...)
		}
	}

	return result, nil
}

// SuggestHunters returns recommended hunters for the given PRD content.
func (o *ResearchOrchestrator) SuggestHunters(vision, problem string, requirements []string) []selector.HunterSelection {
	return o.selector.SelectForPRD(vision, problem, requirements)
}

// SuggestNewHunter determines if a custom hunter would be beneficial.
func (o *ResearchOrchestrator) SuggestNewHunter(vision, problem string, requirements []string) (string, bool) {
	return o.selector.SuggestNewHunter(vision, problem, requirements)
}

// CreateCustomHunter uses the AI agent to design a new hunter.
func (o *ResearchOrchestrator) CreateCustomHunter(ctx context.Context, domain, contextInfo string) (*hunters.CustomHunterSpec, error) {
	creator := hunters.NewHunterCreator()
	return creator.CreateHunterForDomain(ctx, o.scanner.projectPath, domain, contextInfo)
}

// GetResearchBrief generates a research brief for the given PRD content.
func (o *ResearchOrchestrator) GetResearchBrief(vision, problem string, requirements []string) *selector.ResearchBrief {
	return o.briefGen.GenerateForPRD(vision, problem, requirements)
}
