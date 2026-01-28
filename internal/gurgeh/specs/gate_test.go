package specs

import "testing"

func TestNeedsVisionSpec(t *testing.T) {
	tests := []struct {
		name      string
		summaries []Summary
		want      bool
	}{
		{
			name:      "no specs",
			summaries: nil,
			want:      false,
		},
		{
			name:      "one PRD no vision",
			summaries: []Summary{{ID: "PRD-001", Type: ""}},
			want:      true,
		},
		{
			name:      "one PRD with vision",
			summaries: []Summary{{ID: "PRD-001", Type: ""}, {ID: "VIS-001", Type: "vision"}},
			want:      false,
		},
		{
			name:      "only vision",
			summaries: []Summary{{ID: "VIS-001", Type: "vision"}},
			want:      false,
		},
		{
			name:      "explicit prd type no vision",
			summaries: []Summary{{ID: "PRD-001", Type: "prd"}},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsVisionSpec(tt.summaries)
			if got != tt.want {
				t.Errorf("NeedsVisionSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}
