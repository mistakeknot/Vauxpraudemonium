package agenttargets

// DetectAvailableTargets returns a registry of detected targets whose command
// exists in the current environment.
func DetectAvailableTargets(lookPath func(string) (string, error)) Registry {
	reg := Registry{Targets: map[string]Target{}}
	for name, target := range DefaultDetectedRegistry().Targets {
		if _, err := lookPath(target.Command); err == nil {
			reg.Targets[name] = target
		}
	}
	return reg
}

// MergeDetected merges registries in order: detected → global → project.
// Later registries override earlier ones.
func MergeDetected(detected, global, project Registry) Registry {
	merged := Merge(detected, global)
	return Merge(merged, project)
}
