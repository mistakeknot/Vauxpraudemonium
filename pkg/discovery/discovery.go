// Package discovery provides cross-tool data access for Autarch tools.
// It allows each tool to discover and reference data from other tools.
package discovery

import (
	"os"
	"path/filepath"
)

// ToolDir returns the data directory for a tool.
func ToolDir(root, tool string) string {
	return filepath.Join(root, "."+tool)
}

// ToolExists checks if a tool's data directory exists.
func ToolExists(root, tool string) bool {
	_, err := os.Stat(ToolDir(root, tool))
	return err == nil
}

// FindProjectRoot walks up from the given path to find a project root.
// A project root contains at least one of: .praude, .pollard, .tandemonium
func FindProjectRoot(startPath string) (string, error) {
	path := startPath
	for {
		// Check for any tool directory
		for _, tool := range []string{"praude", "pollard", "tandemonium"} {
			if ToolExists(path, tool) {
				return path, nil
			}
		}

		// Move up one directory
		parent := filepath.Dir(path)
		if parent == path {
			// Reached filesystem root
			return startPath, nil // Return start path as fallback
		}
		path = parent
	}
}
