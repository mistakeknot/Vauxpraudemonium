package brief

import "fmt"

type Input struct {
	ID               string
	Title            string
	Summary          string
	Requirements     []string
	Acceptance       []string
	ResearchFiles    []string
	ThinkingPreamble string // Optional thinking shape preamble prepended to brief
}

func Compose(in Input) string {
	var preamble string
	if in.ThinkingPreamble != "" {
		preamble = in.ThinkingPreamble + "\n\n"
	}
	return fmt.Sprintf(`%sPRD: %s
Title: %s

Summary:
%s

Requirements:
%v

Acceptance Criteria:
%v

Research:
%v
`, preamble, in.ID, in.Title, in.Summary, in.Requirements, in.Acceptance, in.ResearchFiles)
}
