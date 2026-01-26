package cuj

import (
	"testing"
)

func TestValidate_RequiredFields(t *testing.T) {
	svc := NewService(nil)

	tests := []struct {
		name       string
		cuj        *CUJ
		wantValid  bool
		wantErrors int
	}{
		{
			name:       "empty CUJ",
			cuj:        &CUJ{},
			wantValid:  false,
			wantErrors: 2, // title and spec_id
		},
		{
			name: "missing spec_id",
			cuj: &CUJ{
				Title: "Test CUJ",
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "missing title",
			cuj: &CUJ{
				SpecID: "SPEC-001",
			},
			wantValid:  false,
			wantErrors: 1,
		},
		{
			name: "valid minimal CUJ",
			cuj: &CUJ{
				Title:  "Test CUJ",
				SpecID: "SPEC-001",
			},
			wantValid:  true,
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.Validate(tt.cuj)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if len(result.Errors) != tt.wantErrors {
				t.Errorf("Validate() errors = %v, want %d", result.Errors, tt.wantErrors)
			}
		})
	}
}

func TestValidate_Steps(t *testing.T) {
	svc := NewService(nil)

	tests := []struct {
		name         string
		steps        []Step
		wantValid    bool
		wantWarnings int
	}{
		{
			name:         "no steps",
			steps:        nil,
			wantValid:    true,
			wantWarnings: 1, // warning about no steps
		},
		{
			name: "step without action",
			steps: []Step{
				{Expected: "something happens"},
			},
			wantValid:    false,
			wantWarnings: 0,
		},
		{
			name: "step without expected",
			steps: []Step{
				{Action: "user clicks button"},
			},
			wantValid:    true,
			wantWarnings: 1,
		},
		{
			name: "complete step",
			steps: []Step{
				{Action: "user clicks button", Expected: "form appears"},
			},
			wantValid:    true,
			wantWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cuj := &CUJ{
				Title:  "Test CUJ",
				SpecID: "SPEC-001",
				Steps:  tt.steps,
			}
			result := svc.Validate(cuj)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v, errors: %v", result.Valid, tt.wantValid, result.Errors)
			}
			// Count warnings about steps (not about other missing fields)
			stepWarnings := 0
			for _, w := range result.Warnings {
				if contains(w, "step") || contains(w, "steps") {
					stepWarnings++
				}
			}
			if stepWarnings != tt.wantWarnings {
				t.Errorf("Validate() step warnings = %d, want %d, all warnings: %v", stepWarnings, tt.wantWarnings, result.Warnings)
			}
		})
	}
}

func TestValidate_HighPriorityCUJ(t *testing.T) {
	svc := NewService(nil)

	// High priority without error recovery should warn
	cuj := &CUJ{
		Title:    "Critical Flow",
		SpecID:   "SPEC-001",
		Priority: PriorityHigh,
	}
	result := svc.Validate(cuj)

	hasErrorRecoveryWarning := false
	for _, w := range result.Warnings {
		if contains(w, "error recovery") {
			hasErrorRecoveryWarning = true
			break
		}
	}
	if !hasErrorRecoveryWarning {
		t.Error("expected warning about error recovery for high-priority CUJ")
	}

	// With error recovery, no such warning
	cuj.ErrorRecovery = []string{"Show error message and retry option"}
	result = svc.Validate(cuj)

	hasErrorRecoveryWarning = false
	for _, w := range result.Warnings {
		if contains(w, "error recovery") {
			hasErrorRecoveryWarning = true
			break
		}
	}
	if hasErrorRecoveryWarning {
		t.Error("should not warn about error recovery when it's defined")
	}
}

func TestCreate_WithoutClient(t *testing.T) {
	svc := NewService(nil)

	cuj := &CUJ{
		Title:  "Onboarding Flow",
		SpecID: "SPEC-001",
		Steps: []Step{
			{Action: "User visits homepage", Expected: "Landing page displayed"},
			{Action: "User clicks Get Started", Expected: "Registration form shown"},
		},
	}

	created, err := svc.Create(nil, cuj)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.Status != StatusDraft {
		t.Errorf("Create() status = %v, want %v", created.Status, StatusDraft)
	}
	if created.Priority != PriorityMedium {
		t.Errorf("Create() priority = %v, want %v", created.Priority, PriorityMedium)
	}
	if created.CreatedAt.IsZero() {
		t.Error("Create() should set CreatedAt")
	}
	if created.UpdatedAt.IsZero() {
		t.Error("Create() should set UpdatedAt")
	}

	// Check step orders are normalized
	for i, step := range created.Steps {
		if step.Order != i+1 {
			t.Errorf("Step %d order = %d, want %d", i, step.Order, i+1)
		}
	}
}

func TestCreate_ValidationFails(t *testing.T) {
	svc := NewService(nil)

	cuj := &CUJ{
		// Missing required fields
	}

	_, err := svc.Create(nil, cuj)
	if err == nil {
		t.Error("Create() should fail for invalid CUJ")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
