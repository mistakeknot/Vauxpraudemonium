// Package sync provides synchronization between Pollard's file-based insights
// and the Intermute coordination server. This enables research to flow into
// the PRD creation process via Gurgeh.
package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/mistakeknot/autarch/internal/pollard/insights"
	"github.com/mistakeknot/autarch/pkg/intermute"
)

// Syncer provides synchronization between Pollard insights and Intermute.
type Syncer struct {
	client  *intermute.Client
	project string
}

// NewSyncer creates a new insight syncer with an Intermute client.
func NewSyncer(client *intermute.Client, project string) *Syncer {
	return &Syncer{
		client:  client,
		project: project,
	}
}

// PushInsight uploads an insight to Intermute.
func (s *Syncer) PushInsight(ctx context.Context, insight *insights.Insight) (intermute.Insight, error) {
	if s.client == nil {
		return intermute.Insight{}, fmt.Errorf("Intermute client not configured")
	}

	iInsight := toIntermuteInsight(insight, s.project)
	return s.client.CreateInsight(ctx, iInsight)
}

// PushInsights uploads multiple insights to Intermute.
func (s *Syncer) PushInsights(ctx context.Context, insts []*insights.Insight) ([]intermute.Insight, error) {
	if s.client == nil {
		return nil, fmt.Errorf("Intermute client not configured")
	}

	var results []intermute.Insight
	for _, insight := range insts {
		iInsight, err := s.PushInsight(ctx, insight)
		if err != nil {
			// Log error but continue with other insights
			continue
		}
		results = append(results, iInsight)
	}
	return results, nil
}

// PullInsights downloads insights from Intermute filtered by spec or category.
func (s *Syncer) PullInsights(ctx context.Context, specID, category string) ([]*insights.Insight, error) {
	if s.client == nil {
		return nil, fmt.Errorf("Intermute client not configured")
	}

	iInsights, err := s.client.ListInsights(ctx, specID, category)
	if err != nil {
		return nil, err
	}

	result := make([]*insights.Insight, len(iInsights))
	for i, iInsight := range iInsights {
		result[i] = fromIntermuteInsight(&iInsight)
	}
	return result, nil
}

// LinkInsightToSpec links an insight to a specification.
func (s *Syncer) LinkInsightToSpec(ctx context.Context, insightID, specID string) error {
	if s.client == nil {
		return fmt.Errorf("Intermute client not configured")
	}
	return s.client.LinkInsightToSpec(ctx, insightID, specID)
}

// SyncAllInsights syncs all local insights from a project directory to Intermute.
func (s *Syncer) SyncAllInsights(ctx context.Context, projectPath string) (int, error) {
	if s.client == nil {
		return 0, fmt.Errorf("Intermute client not configured")
	}

	localInsights, err := insights.LoadAll(projectPath)
	if err != nil {
		return 0, fmt.Errorf("failed to load local insights: %w", err)
	}

	count := 0
	for _, insight := range localInsights {
		_, err := s.PushInsight(ctx, insight)
		if err != nil {
			continue // Skip failed insights
		}
		count++
	}
	return count, nil
}

// GetInsightsForInterview retrieves insights relevant for a PRD interview.
// It returns insights grouped by category for easy surfacing in the TUI.
func (s *Syncer) GetInsightsForInterview(ctx context.Context, specID string) (map[string][]*insights.Insight, error) {
	if s.client == nil {
		return nil, fmt.Errorf("Intermute client not configured")
	}

	// Get all insights linked to this spec
	allInsights, err := s.PullInsights(ctx, specID, "")
	if err != nil {
		return nil, err
	}

	// Group by category
	grouped := make(map[string][]*insights.Insight)
	for _, insight := range allInsights {
		category := string(insight.Category)
		if category == "" {
			category = "general"
		}
		grouped[category] = append(grouped[category], insight)
	}

	return grouped, nil
}

// FormatInsightsForInterview formats insights for display in the interview TUI.
func FormatInsightsForInterview(groupedInsights map[string][]*insights.Insight) string {
	if len(groupedInsights) == 0 {
		return "No research insights available yet."
	}

	var sb strings.Builder
	sb.WriteString("# Research Insights\n\n")

	categoryOrder := []string{"competitive", "trends", "user", "general"}
	for _, category := range categoryOrder {
		insts, ok := groupedInsights[category]
		if !ok || len(insts) == 0 {
			continue
		}

		sb.WriteString("## ")
		sb.WriteString(strings.Title(category))
		sb.WriteString("\n\n")

		for _, insight := range insts {
			sb.WriteString("### ")
			sb.WriteString(insight.Title)
			sb.WriteString("\n")

			for _, finding := range insight.Findings {
				sb.WriteString("- **")
				sb.WriteString(finding.Title)
				sb.WriteString("** (")
				sb.WriteString(string(finding.Relevance))
				sb.WriteString("): ")
				sb.WriteString(finding.Description)
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// --- Conversion helpers ---

func toIntermuteInsight(insight *insights.Insight, project string) intermute.Insight {
	// Extract first source URL if available
	var url string
	if len(insight.Sources) > 0 {
		url = insight.Sources[0].URL
	}

	// Map Pollard category to Intermute category
	category := string(insight.Category)
	if category == "" {
		category = "general"
	}

	// Calculate score from findings relevance
	score := calculateScore(insight.Findings)

	// Build body from findings
	body := buildInsightBody(insight)

	return intermute.Insight{
		ID:        insight.ID,
		Project:   project,
		Source:    "pollard",
		Category:  category,
		Title:     insight.Title,
		Body:      body,
		URL:       url,
		Score:     score,
		CreatedAt: insight.CollectedAt,
	}
}

func fromIntermuteInsight(iInsight *intermute.Insight) *insights.Insight {
	// Map Intermute category back to Pollard category
	var category insights.Category
	switch iInsight.Category {
	case "competitive":
		category = insights.CategoryCompetitive
	case "trends":
		category = insights.CategoryTrends
	case "user":
		category = insights.CategoryUser
	default:
		category = insights.CategoryCompetitive // Default
	}

	// Parse body back to findings (simplified)
	findings := parseBodyToFindings(iInsight.Body)

	return &insights.Insight{
		ID:          iInsight.ID,
		Title:       iInsight.Title,
		Category:    category,
		CollectedAt: iInsight.CreatedAt,
		Sources:     []insights.Source{{URL: iInsight.URL, Type: iInsight.Source}},
		Findings:    findings,
	}
}

func calculateScore(findings []insights.Finding) float64 {
	if len(findings) == 0 {
		return 0.5
	}

	var total float64
	for _, f := range findings {
		switch f.Relevance {
		case insights.RelevanceHigh:
			total += 1.0
		case insights.RelevanceMedium:
			total += 0.6
		case insights.RelevanceLow:
			total += 0.3
		default:
			total += 0.5
		}
	}
	return total / float64(len(findings))
}

func buildInsightBody(insight *insights.Insight) string {
	var sb strings.Builder

	for _, finding := range insight.Findings {
		sb.WriteString("**")
		sb.WriteString(finding.Title)
		sb.WriteString("**: ")
		sb.WriteString(finding.Description)
		sb.WriteString("\n")
	}

	if len(insight.Recommendations) > 0 {
		sb.WriteString("\n_Recommendations:_\n")
		for _, rec := range insight.Recommendations {
			sb.WriteString("- ")
			sb.WriteString(rec.FeatureHint)
			sb.WriteString(": ")
			sb.WriteString(rec.Rationale)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func parseBodyToFindings(body string) []insights.Finding {
	// Simplified parsing - in production would need more robust parsing
	if body == "" {
		return nil
	}

	// For now, create a single finding from the body
	return []insights.Finding{
		{
			Title:       "Insight",
			Description: body,
			Relevance:   insights.RelevanceMedium,
		},
	}
}
