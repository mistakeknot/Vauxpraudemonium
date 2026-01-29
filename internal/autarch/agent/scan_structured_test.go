package agent

import (
	"strings"
	"testing"
)

func TestParseStructuredScanResponse(t *testing.T) {
	content := `{
  "project_name": "Autarch",
  "description": "Autarch is a suite...",
  "artifacts": {
    "vision": {
      "phase": "vision",
      "version": "v1",
      "summary": "A long enough vision summary for validation.",
      "goals": ["Goal 1"],
      "non_goals": [],
      "evidence": [
        {"type":"file","path":"README.md","quote":"Autarch","confidence":0.9},
        {"type":"doc","path":"docs/ARCHITECTURE.md","quote":"Architecture","confidence":0.8}
      ],
      "open_questions": [],
      "quality": {"clarity":0.7,"completeness":0.7,"grounding":0.7,"consistency":0.7}
    },
    "problem": {
      "phase": "problem",
      "version": "v1",
      "summary": "A long enough problem summary.",
      "pain_points": ["Pain"],
      "impact": "Impact text",
      "evidence": [
        {"type":"file","path":"README.md","quote":"Autarch","confidence":0.9},
        {"type":"doc","path":"docs/ARCHITECTURE.md","quote":"Architecture","confidence":0.8}
      ],
      "open_questions": [],
      "quality": {"clarity":0.7,"completeness":0.7,"grounding":0.7,"consistency":0.7}
    },
    "users": {
      "phase": "users",
      "version": "v1",
      "personas": [{"name":"Builders","needs":["Need"],"context":"Context text"}],
      "evidence": [
        {"type":"file","path":"README.md","quote":"Autarch","confidence":0.9},
        {"type":"doc","path":"docs/ARCHITECTURE.md","quote":"Architecture","confidence":0.8}
      ],
      "open_questions": [],
      "quality": {"clarity":0.7,"completeness":0.7,"grounding":0.7,"consistency":0.7}
    }
  }
}`

	res, err := parseScanResponse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.PhaseArtifacts == nil {
		t.Fatal("expected artifacts")
	}
	if res.PhaseArtifacts.Vision.Summary == "" {
		t.Fatal("expected vision artifact")
	}
}

func TestStructuredScanMapsLegacyFields(t *testing.T) {
	content := `{
  "project_name": "Autarch",
  "description": "Autarch is a suite...",
  "artifacts": {
    "vision": {
      "phase": "vision",
      "version": "v1",
      "summary": "A long enough vision summary for validation.",
      "goals": ["Goal 1"],
      "non_goals": [],
      "evidence": [
        {"type":"file","path":"README.md","quote":"Autarch","confidence":0.9},
        {"type":"doc","path":"docs/ARCHITECTURE.md","quote":"Architecture","confidence":0.8}
      ],
      "open_questions": [],
      "quality": {"clarity":0.7,"completeness":0.7,"grounding":0.7,"consistency":0.7}
    },
    "problem": {
      "phase": "problem",
      "version": "v1",
      "summary": "A long enough problem summary.",
      "pain_points": ["Pain"],
      "impact": "Impact text",
      "evidence": [
        {"type":"file","path":"README.md","quote":"Autarch","confidence":0.9},
        {"type":"doc","path":"docs/ARCHITECTURE.md","quote":"Architecture","confidence":0.8}
      ],
      "open_questions": [],
      "quality": {"clarity":0.7,"completeness":0.7,"grounding":0.7,"consistency":0.7}
    },
    "users": {
      "phase": "users",
      "version": "v1",
      "personas": [{"name":"Builders","needs":["Need"],"context":"Context text"}],
      "evidence": [
        {"type":"file","path":"README.md","quote":"Autarch","confidence":0.9},
        {"type":"doc","path":"docs/ARCHITECTURE.md","quote":"Architecture","confidence":0.8}
      ],
      "open_questions": [],
      "quality": {"clarity":0.7,"completeness":0.7,"grounding":0.7,"consistency":0.7}
    }
  }
}`

	res, err := parseScanResponse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Vision == "" || res.Problem == "" || res.Users == "" {
		t.Fatal("expected legacy fields populated from artifacts")
	}
	if res.Users != "Builders" {
		t.Fatalf("expected users to map from persona name, got %q", res.Users)
	}
}

func TestScanPromptRequestsStructuredArtifacts(t *testing.T) {
	prompt := buildScanPrompt("/tmp/project", map[string]string{"README.md": "Autarch"})
	if !containsAll(prompt, []string{"\"artifacts\"", "\"vision\"", "\"evidence\"", "\"quality\""}) {
		t.Fatalf("prompt missing structured artifact output")
	}
}

func TestStructuredScanValidation(t *testing.T) {
	content := `{
  "project_name": "Autarch",
  "description": "Autarch is a suite...",
  "artifacts": {
    "vision": {
      "phase": "vision",
      "version": "v1",
      "summary": "A long enough vision summary for validation.",
      "goals": ["Goal 1"],
      "non_goals": [],
      "evidence": [
        {"type":"file","path":"README.md","quote":"Autarch","confidence":0.9},
        {"type":"doc","path":"docs/ARCHITECTURE.md","quote":"Architecture","confidence":0.8}
      ],
      "open_questions": [],
      "quality": {"clarity":0.7,"completeness":0.7,"grounding":0.7,"consistency":0.7}
    }
  }
}`
	res, err := parseScanResponse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	errs := ValidateStructuredScanArtifacts(res, map[string]string{
		"README.md":            "Autarch",
		"docs/ARCHITECTURE.md": "Architecture",
	})
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %d", len(errs))
	}
}

func containsAll(s string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}
