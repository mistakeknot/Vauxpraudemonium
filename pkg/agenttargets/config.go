package agenttargets

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type legacyAgent struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
}

type fileConfig struct {
	Targets map[string]Target      `toml:"targets"`
	Agents  map[string]legacyAgent `toml:"agents"`
}

func Load(globalPath, projectRoot string) (Registry, Registry, error) {
	global, err := loadRegistry(globalPath)
	if err != nil {
		return Registry{}, Registry{}, err
	}
	project, err := loadProjectRegistry(projectRoot)
	if err != nil {
		return Registry{}, Registry{}, err
	}
	return global, project, nil
}

func loadRegistry(path string) (Registry, error) {
	if path == "" {
		return Registry{Targets: map[string]Target{}}, nil
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return Registry{Targets: map[string]Target{}}, nil
		}
		return Registry{}, err
	}
	var cfg fileConfig
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Registry{}, err
	}
	return registryFromConfig(cfg), nil
}

func loadProjectRegistry(projectRoot string) (Registry, error) {
	if projectRoot == "" {
		return Registry{Targets: map[string]Target{}}, nil
	}
	agentsPath := filepath.Join(projectRoot, ".praude", "agents.toml")
	if _, err := os.Stat(agentsPath); err == nil {
		return loadRegistry(agentsPath)
	} else if err != nil && !os.IsNotExist(err) {
		return Registry{}, err
	}

	compatPath := filepath.Join(projectRoot, ".praude", "config.toml")
	if _, err := os.Stat(compatPath); err != nil {
		if os.IsNotExist(err) {
			return Registry{Targets: map[string]Target{}}, nil
		}
		return Registry{}, err
	}
	var cfg fileConfig
	if _, err := toml.DecodeFile(compatPath, &cfg); err != nil {
		return Registry{}, err
	}
	return registryFromConfig(cfg), nil
}

func registryFromConfig(cfg fileConfig) Registry {
	reg := Registry{Targets: map[string]Target{}}
	for name, target := range cfg.Targets {
		if target.Name == "" {
			target.Name = name
		}
		reg.Targets[name] = target
	}
	for name, agent := range cfg.Agents {
		reg.Targets[name] = Target{
			Name:    name,
			Type:    TargetCommand,
			Command: agent.Command,
			Args:    agent.Args,
		}
	}
	return reg
}
