package selector

import (
	"fmt"
	"strings"
)

// ResearchBrief contains instructions for an AI agent to conduct research.
type ResearchBrief struct {
	Title         string
	Domain        string
	Questions     []string
	Scope         []string
	SuggestedAPIs []string // Optional API hunters to supplement
	OutputFormat  string
}

// BriefGenerator creates research briefs from PRD content.
type BriefGenerator struct {
	domains []Domain
}

// NewBriefGenerator creates a new brief generator.
func NewBriefGenerator() *BriefGenerator {
	return &BriefGenerator{domains: Domains}
}

// GenerateForPRD creates a research brief tailored to the PRD.
func (g *BriefGenerator) GenerateForPRD(vision, problem string, requirements []string) *ResearchBrief {
	// Detect primary domain
	allText := strings.ToLower(vision + " " + problem + " " + strings.Join(requirements, " "))
	domain := FindDomain(allText)

	// Generate research questions
	questions := g.generateQuestions(vision, problem, requirements, domain)

	// Determine scope based on domain
	scope := g.generateScope(domain)

	// Suggest supplementary API hunters (if user has keys)
	suggestedAPIs := g.suggestAPIs(domain)

	return &ResearchBrief{
		Title:         fmt.Sprintf("Research for: %s", truncate(vision, 50)),
		Domain:        domain.Name,
		Questions:     questions,
		Scope:         scope,
		SuggestedAPIs: suggestedAPIs,
		OutputFormat:  g.outputFormat(),
	}
}

// ToPrompt converts the research brief to a prompt for the AI agent.
func (b *ResearchBrief) ToPrompt() string {
	var sb strings.Builder

	sb.WriteString("# Research Brief\n\n")
	sb.WriteString("Conduct research on the following topics and return structured findings.\n\n")

	sb.WriteString("## Research Questions\n")
	for i, q := range b.Questions {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, q))
	}

	sb.WriteString("\n## Scope\n")
	for _, s := range b.Scope {
		sb.WriteString(fmt.Sprintf("- %s\n", s))
	}

	sb.WriteString("\n## Expected Output Format\n")
	sb.WriteString("Return findings as YAML with this structure:\n")
	sb.WriteString("```yaml\n")
	sb.WriteString(b.OutputFormat)
	sb.WriteString("```\n\n")

	sb.WriteString("## Guidelines\n")
	sb.WriteString("- Focus on authoritative sources (academic papers, official docs, reputable sites)\n")
	sb.WriteString("- Prioritize recent information (last 2-3 years when relevant)\n")
	sb.WriteString("- Include both technical and domain-specific perspectives\n")
	sb.WriteString("- Be thorough but concise in summaries\n")

	return sb.String()
}

// generateQuestions creates research questions based on PRD content.
func (g *BriefGenerator) generateQuestions(vision, problem string, requirements []string, domain Domain) []string {
	var questions []string

	// Core questions from PRD
	if problem != "" {
		questions = append(questions, fmt.Sprintf("What existing solutions address: %s", problem))
	}
	if vision != "" {
		questions = append(questions, fmt.Sprintf("What are the key technical approaches for: %s", vision))
	}

	// Domain-specific questions
	switch domain.Name {
	case "medical":
		questions = append(questions, "What clinical research supports this approach?")
		questions = append(questions, "What regulatory considerations apply (FDA, HIPAA)?")
	case "nutrition":
		questions = append(questions, "What nutrition databases and APIs exist for this use case?")
		questions = append(questions, "What allergen detection methods are currently used?")
	case "legal":
		questions = append(questions, "What regulations and compliance requirements apply?")
		questions = append(questions, "What legal precedents are relevant?")
	case "technology":
		questions = append(questions, "What open source projects implement similar features?")
		questions = append(questions, "What are the recommended frameworks and libraries?")
	case "economics":
		questions = append(questions, "What economic data sources are available?")
		questions = append(questions, "What market trends are relevant?")
	default:
		questions = append(questions, "What are the industry best practices?")
		questions = append(questions, "What academic research exists on this topic?")
	}

	// Questions from requirements (limit to first 3)
	for i, req := range requirements {
		if i >= 3 {
			break
		}
		if req != "" {
			questions = append(questions, fmt.Sprintf("How do existing solutions handle: %s", req))
		}
	}

	return questions
}

// generateScope creates scope guidelines based on domain.
func (g *BriefGenerator) generateScope(domain Domain) []string {
	base := []string{
		"Focus on authoritative sources (official docs, academic papers)",
		"Prioritize recent information (last 2-3 years)",
		"Include both technical implementation and domain expertise",
	}

	switch domain.Name {
	case "medical":
		base = append(base, "Include peer-reviewed research from PubMed/medical journals")
		base = append(base, "Note any FDA or regulatory requirements")
	case "nutrition":
		base = append(base, "Reference USDA or equivalent nutritional databases")
		base = append(base, "Include allergen labeling standards")
	case "legal":
		base = append(base, "Cite relevant statutes and case law")
		base = append(base, "Note jurisdiction-specific requirements")
	case "technology":
		base = append(base, "Evaluate GitHub projects by stars, activity, and maintenance")
		base = append(base, "Check for security advisories and known issues")
	case "economics":
		base = append(base, "Use official government data sources (BLS, Census, Fed)")
		base = append(base, "Include recent market analysis")
	}

	return base
}

// suggestAPIs returns API hunters that could supplement agent research.
func (g *BriefGenerator) suggestAPIs(domain Domain) []string {
	return domain.Hunters
}

// outputFormat returns the expected YAML output format.
func (g *BriefGenerator) outputFormat() string {
	return `sources:
  - title: "Source title"
    url: "https://..."
    type: "article|paper|repository|documentation"
    summary: "Brief summary of key insights"
    relevance: "high|medium|low"
    collected_at: "2026-01-24"
insights:
  - finding: "Key finding or insight"
    evidence: "Supporting evidence"
    relevance: "high|medium|low"
`
}

// truncate shortens a string to the specified length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// GenerateForEpic creates a research brief for an epic/story.
func (g *BriefGenerator) GenerateForEpic(title, description string) *ResearchBrief {
	allText := strings.ToLower(title + " " + description)
	domain := FindDomain(allText)

	questions := []string{
		fmt.Sprintf("What are the best implementation patterns for: %s", title),
		fmt.Sprintf("What open source examples exist for: %s", title),
		"What are common pitfalls and anti-patterns to avoid?",
		"What testing strategies are recommended?",
	}

	scope := []string{
		"Focus on implementation examples and code patterns",
		"Prioritize well-maintained, production-quality examples",
		"Include testing and error handling approaches",
	}

	return &ResearchBrief{
		Title:         fmt.Sprintf("Implementation research: %s", truncate(title, 40)),
		Domain:        domain.Name,
		Questions:     questions,
		Scope:         scope,
		SuggestedAPIs: []string{"github-scout"},
		OutputFormat:  g.outputFormat(),
	}
}
