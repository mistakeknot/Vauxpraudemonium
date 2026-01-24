package selector

import (
	"strings"
)

// HunterSelection represents a selected hunter with relevance information.
type HunterSelection struct {
	Name      string
	Score     float64
	Queries   []string
	Domain    string
	Reasoning string
}

// Selector determines which hunters to use based on PRD content.
type Selector struct {
	domains []Domain
}

// NewSelector creates a new hunter selector.
func NewSelector() *Selector {
	return &Selector{domains: Domains}
}

// SelectForPRD analyzes PRD content and returns recommended hunters.
func (s *Selector) SelectForPRD(vision, problem string, requirements []string) []HunterSelection {
	// Combine all text for domain detection
	allText := strings.ToLower(vision + " " + problem + " " + strings.Join(requirements, " "))

	// Detect primary domain
	domain := FindDomain(allText)

	// Build selections from domain hunters
	selections := make([]HunterSelection, 0, len(domain.Hunters))
	for i, hunterName := range domain.Hunters {
		// Calculate score based on position and domain match strength
		score := 1.0 - (float64(i) * 0.1)

		// Build queries specific to this hunter
		queries := s.buildQueriesForHunter(hunterName, vision, problem, requirements, domain)

		selections = append(selections, HunterSelection{
			Name:      hunterName,
			Score:     score,
			Queries:   queries,
			Domain:    domain.Name,
			Reasoning: s.explainSelection(hunterName, domain),
		})
	}

	return selections
}

// buildQueriesForHunter creates domain-specific queries for a hunter.
func (s *Selector) buildQueriesForHunter(hunter, vision, problem string, requirements []string, domain Domain) []string {
	var queries []string

	// Base queries from PRD
	if vision != "" {
		queries = append(queries, vision)
	}
	if problem != "" {
		queries = append(queries, problem)
	}

	// Add domain-specific query modifications
	switch hunter {
	case "pubmed":
		// For medical/nutrition research, focus on clinical aspects
		if problem != "" {
			queries = append(queries, problem+" treatment")
			queries = append(queries, problem+" clinical study")
		}
	case "github-scout":
		// For technology, focus on implementations
		if vision != "" {
			queries = append(queries, vision+" implementation")
			queries = append(queries, vision+" library")
		}
	case "arxiv":
		// For academic, focus on research papers
		if problem != "" {
			queries = append(queries, problem+" algorithm")
			queries = append(queries, problem+" approach")
		}
	case "usda-nutrition":
		// Extract food-related terms
		for _, req := range requirements {
			reqLower := strings.ToLower(req)
			if containsWord(reqLower, "allergy") || containsWord(reqLower, "allergen") ||
				containsWord(reqLower, "ingredient") || containsWord(reqLower, "nutrition") {
				queries = append(queries, req)
			}
		}
	case "legal":
		// Focus on regulatory aspects
		if problem != "" {
			queries = append(queries, problem+" regulation")
			queries = append(queries, problem+" compliance")
		}
	case "economics":
		// Focus on market/labor aspects
		if vision != "" {
			queries = append(queries, vision+" market")
		}
	}

	// Limit queries to prevent overwhelming the hunter
	if len(queries) > 5 {
		queries = queries[:5]
	}

	return queries
}

// explainSelection provides reasoning for why a hunter was selected.
func (s *Selector) explainSelection(hunter string, domain Domain) string {
	switch hunter {
	case "pubmed":
		return "Selected for medical/clinical research from PubMed database"
	case "openalex":
		return "Selected for broad academic research across 260M+ works"
	case "github-scout":
		return "Selected for finding open source implementations and libraries"
	case "hackernews":
		return "Selected for technology trends and community discussions"
	case "arxiv":
		return "Selected for cutting-edge research papers"
	case "usda-nutrition":
		return "Selected for nutritional data and food composition"
	case "legal":
		return "Selected for legal research and case law"
	case "economics":
		return "Selected for economic data and labor statistics"
	case "wiki":
		return "Selected for general reference and background knowledge"
	default:
		return "Selected based on domain: " + domain.Name
	}
}

// SuggestNewHunter determines if a new custom hunter might be beneficial.
func (s *Selector) SuggestNewHunter(vision, problem string, requirements []string) (string, bool) {
	allText := strings.ToLower(vision + " " + problem + " " + strings.Join(requirements, " "))
	domain := FindDomain(allText)

	// Specific domain gaps that might warrant custom hunters
	suggestions := map[string]struct {
		keywords []string
		hunter   string
	}{
		"culinary": {
			keywords: []string{"recipe", "cooking", "cuisine", "chef", "restaurant"},
			hunter:   "recipe-hunter",
		},
		"real-estate": {
			keywords: []string{"property", "real estate", "housing", "mortgage", "rental"},
			hunter:   "property-hunter",
		},
		"weather": {
			keywords: []string{"weather", "forecast", "climate", "temperature", "precipitation"},
			hunter:   "weather-hunter",
		},
		"sports": {
			keywords: []string{"sports", "athlete", "game", "team", "score", "league"},
			hunter:   "sports-hunter",
		},
	}

	for _, suggestion := range suggestions {
		matchCount := 0
		for _, keyword := range suggestion.keywords {
			if containsWord(allText, keyword) {
				matchCount++
			}
		}
		if matchCount >= 2 {
			return suggestion.hunter, true
		}
	}

	// If domain is general but has specific patterns, suggest custom hunter
	if domain.Name == "general" && len(requirements) > 3 {
		// Many requirements might indicate a specialized domain
		return "domain-specific-hunter", true
	}

	return "", false
}
