package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mistakeknot/autarch/pkg/agenttargets"
	pkgtui "github.com/mistakeknot/autarch/pkg/tui"
)

// LoadAgentOptions returns agent options from detected binaries and config.
func LoadAgentOptions(projectRoot string) ([]pkgtui.AgentOption, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	globalPath := filepath.Join(configDir, "autarch", "agents.toml")

	global, project, err := agenttargets.Load(globalPath, projectRoot)
	if err != nil {
		return nil, err
	}

	detected := agenttargets.DetectAvailableTargets(exec.LookPath)
	merged := agenttargets.MergeDetected(detected, global, project)
	return buildAgentOptionsFromRegistry(merged), nil
}

func buildAgentOptionsFromRegistry(reg agenttargets.Registry) []pkgtui.AgentOption {
	options := make([]pkgtui.AgentOption, 0, len(reg.Targets))
	for name, target := range reg.Targets {
		optName := strings.TrimSpace(target.Name)
		if optName == "" {
			optName = name
		}
		if optName == "" {
			continue
		}
		options = append(options, pkgtui.AgentOption{
			Name:   optName,
			Source: string(target.Type),
		})
	}

	priority := map[string]int{
		"codex":  0,
		"claude": 1,
		"gemini": 2,
	}

	sort.Slice(options, func(i, j int) bool {
		ai := strings.ToLower(options[i].Name)
		aj := strings.ToLower(options[j].Name)
		pi, okI := priority[ai]
		pj, okJ := priority[aj]
		if okI && okJ && pi != pj {
			return pi < pj
		}
		if okI != okJ {
			return okI
		}
		return ai < aj
	})

	return options
}
