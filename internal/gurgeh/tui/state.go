package tui

import "github.com/mistakeknot/autarch/internal/gurgeh/specs"

type SharedState struct {
	Summaries []specs.Summary
	Selected  int
	Focus     string
	Filter    string
}

func NewSharedState() *SharedState {
	return &SharedState{Focus: "LIST"}
}
