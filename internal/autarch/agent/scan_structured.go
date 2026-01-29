package agent

import "strings"

type PhaseArtifacts struct {
	Vision  *VisionArtifact  `json:"vision,omitempty"`
	Problem *ProblemArtifact `json:"problem,omitempty"`
	Users   *UsersArtifact   `json:"users,omitempty"`
}

type structuredScanResponse struct {
	ProjectName  string          `json:"project_name"`
	Description  string          `json:"description"`
	Vision       string          `json:"vision,omitempty"`
	Users        string          `json:"users,omitempty"`
	Problem      string          `json:"problem,omitempty"`
	Platform     string          `json:"platform,omitempty"`
	Language     string          `json:"language,omitempty"`
	Requirements []string        `json:"requirements,omitempty"`
	Artifacts    *PhaseArtifacts `json:"artifacts,omitempty"`
}

func ValidateStructuredScanArtifacts(result *ScanResult, files map[string]string) []ValidationError {
	if result == nil || result.PhaseArtifacts == nil {
		return nil
	}
	lookup := fileEvidenceLookup{files: files}
	errs := []ValidationError{}
	if result.PhaseArtifacts.Vision != nil {
		errs = append(errs, validatePhaseArtifact("vision", result.PhaseArtifacts.Vision, lookup)...)
	}
	if result.PhaseArtifacts.Problem != nil {
		errs = append(errs, validatePhaseArtifact("problem", result.PhaseArtifacts.Problem, lookup)...)
	}
	if result.PhaseArtifacts.Users != nil {
		errs = append(errs, validatePhaseArtifact("users", result.PhaseArtifacts.Users, lookup)...)
	}
	return errs
}

func applyStructuredDefaults(result *ScanResult) {
	if result == nil || result.PhaseArtifacts == nil {
		return
	}
	if result.Vision == "" && result.PhaseArtifacts.Vision != nil {
		result.Vision = result.PhaseArtifacts.Vision.Summary
	}
	if result.Problem == "" && result.PhaseArtifacts.Problem != nil {
		result.Problem = result.PhaseArtifacts.Problem.Summary
	}
	if result.Users == "" && result.PhaseArtifacts.Users != nil {
		names := []string{}
		for _, persona := range result.PhaseArtifacts.Users.Personas {
			if persona.Name != "" {
				names = append(names, persona.Name)
			}
		}
		if len(names) > 0 {
			result.Users = strings.Join(names, ", ")
		} else if len(result.PhaseArtifacts.Users.Personas) > 0 {
			result.Users = result.PhaseArtifacts.Users.Personas[0].Context
		}
	}
}
