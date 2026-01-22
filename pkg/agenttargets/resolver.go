package agenttargets

import (
	"fmt"
	"strings"
)

type Context string

const (
	GlobalContext  Context = "global"
	ProjectContext Context = "project"
	SpawnContext   Context = "spawn"
)

type Source string

const (
	SourceDetected Source = "detected"
	SourceGlobal   Source = "global"
	SourceProject  Source = "project"
)

type ResolvedTarget struct {
	Name    string
	Type    TargetType
	Command string
	Args    []string
	Env     map[string]string
	Source  Source
}

type Resolver struct {
	global   Registry
	project  Registry
	detected Registry
}

func NewResolver(global, project Registry) Resolver {
	return Resolver{
		global:   global,
		project:  project,
		detected: DefaultDetectedRegistry(),
	}
}

func (r Resolver) Resolve(ctx Context, name string) (ResolvedTarget, error) {
	key := strings.ToLower(name)
	if key == "" {
		return ResolvedTarget{}, fmt.Errorf("target name required")
	}

	if ctx == ProjectContext || ctx == SpawnContext {
		if t, ok := r.project.Targets[key]; ok {
			return resolvedFromTarget(t, SourceProject), nil
		}
	}

	if t, ok := r.global.Targets[key]; ok {
		return resolvedFromTarget(t, SourceGlobal), nil
	}

	if t, ok := r.detected.Targets[key]; ok {
		return resolvedFromTarget(t, SourceDetected), nil
	}

	return ResolvedTarget{}, fmt.Errorf("unknown target %q", name)
}

func DefaultDetectedRegistry() Registry {
	return Registry{Targets: map[string]Target{
		"codex":  {Name: "codex", Type: TargetDetected, Command: "codex"},
		"claude": {Name: "claude", Type: TargetDetected, Command: "claude"},
		"gemini": {Name: "gemini", Type: TargetDetected, Command: "gemini"},
	}}
}

func resolvedFromTarget(t Target, source Source) ResolvedTarget {
	return ResolvedTarget{
		Name:    t.Name,
		Type:    t.Type,
		Command: t.Command,
		Args:    t.Args,
		Env:     t.Env,
		Source:  source,
	}
}
