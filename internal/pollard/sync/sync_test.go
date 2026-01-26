package sync

import (
	"testing"
	"time"

	"github.com/mistakeknot/autarch/internal/pollard/insights"
)

func TestToIntermuteInsight(t *testing.T) {
	insight := &insights.Insight{
		ID:          "INS-001",
		Title:       "Competitor Analysis",
		Category:    insights.CategoryCompetitive,
		CollectedAt: time.Date(2026, 1, 20, 10, 0, 0, 0, time.UTC),
		Sources: []insights.Source{
			{URL: "https://example.com", Type: "article"},
		},
		Findings: []insights.Finding{
			{Title: "Market Gap", Description: "Gap in market", Relevance: insights.RelevanceHigh},
		},
	}

	iInsight := toIntermuteInsight(insight, "test-project")

	if iInsight.ID != "INS-001" {
		t.Errorf("ID = %v, want %v", iInsight.ID, "INS-001")
	}
	if iInsight.Project != "test-project" {
		t.Errorf("Project = %v, want %v", iInsight.Project, "test-project")
	}
	if iInsight.Title != "Competitor Analysis" {
		t.Errorf("Title = %v, want %v", iInsight.Title, "Competitor Analysis")
	}
	if iInsight.Category != "competitive" {
		t.Errorf("Category = %v, want %v", iInsight.Category, "competitive")
	}
	if iInsight.Source != "pollard" {
		t.Errorf("Source = %v, want %v", iInsight.Source, "pollard")
	}
	if iInsight.URL != "https://example.com" {
		t.Errorf("URL = %v, want %v", iInsight.URL, "https://example.com")
	}
	if iInsight.Score != 1.0 {
		t.Errorf("Score = %v, want %v", iInsight.Score, 1.0)
	}
}

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name     string
		findings []insights.Finding
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "empty findings",
			findings: nil,
			wantMin:  0.5,
			wantMax:  0.5,
		},
		{
			name: "all high",
			findings: []insights.Finding{
				{Relevance: insights.RelevanceHigh},
				{Relevance: insights.RelevanceHigh},
			},
			wantMin: 1.0,
			wantMax: 1.0,
		},
		{
			name: "all low",
			findings: []insights.Finding{
				{Relevance: insights.RelevanceLow},
				{Relevance: insights.RelevanceLow},
			},
			wantMin: 0.3,
			wantMax: 0.3,
		},
		{
			name: "mixed",
			findings: []insights.Finding{
				{Relevance: insights.RelevanceHigh},
				{Relevance: insights.RelevanceMedium},
				{Relevance: insights.RelevanceLow},
			},
			wantMin: 0.63,
			wantMax: 0.64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateScore(tt.findings)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateScore() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestBuildInsightBody(t *testing.T) {
	insight := &insights.Insight{
		Findings: []insights.Finding{
			{Title: "Finding 1", Description: "Description 1"},
			{Title: "Finding 2", Description: "Description 2"},
		},
		Recommendations: []insights.Recommendation{
			{FeatureHint: "Add feature X", Rationale: "Because Y"},
		},
	}

	body := buildInsightBody(insight)

	if body == "" {
		t.Error("buildInsightBody() returned empty string")
	}
	if !contains(body, "Finding 1") {
		t.Error("body should contain finding title")
	}
	if !contains(body, "Add feature X") {
		t.Error("body should contain recommendation")
	}
}

func TestFormatInsightsForInterview(t *testing.T) {
	grouped := map[string][]*insights.Insight{
		"competitive": {
			{
				Title: "Comp Insight",
				Findings: []insights.Finding{
					{Title: "Market Gap", Description: "Gap found", Relevance: insights.RelevanceHigh},
				},
			},
		},
	}

	result := FormatInsightsForInterview(grouped)

	if result == "" {
		t.Error("FormatInsightsForInterview() returned empty string")
	}
	if !contains(result, "Research Insights") {
		t.Error("result should contain header")
	}
	if !contains(result, "Competitive") {
		t.Error("result should contain category")
	}
	if !contains(result, "Market Gap") {
		t.Error("result should contain finding")
	}
}

func TestFormatInsightsForInterview_Empty(t *testing.T) {
	result := FormatInsightsForInterview(nil)
	if result != "No research insights available yet." {
		t.Errorf("unexpected result for empty insights: %v", result)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
