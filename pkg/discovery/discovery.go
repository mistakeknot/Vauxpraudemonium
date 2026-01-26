// Package discovery provides cross-tool data access for Autarch tools.
// It allows each tool to discover and reference data from other tools.
package discovery

import (
	"fmt"
	"os"
	"path/filepath"
)

// ParseError represents a failure to parse a single file during discovery.
// Discovery functions that return ParseError slices continue processing
// other files, allowing partial results even when some files are malformed.
type ParseError struct {
	Path string // Absolute path to the file that failed
	Err  error  // The underlying parse error
}

// Error implements the error interface
func (e ParseError) Error() string {
	return fmt.Sprintf("parse %s: %v", e.Path, e.Err)
}

// Unwrap returns the underlying error for errors.Is/As support
func (e ParseError) Unwrap() error {
	return e.Err
}

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
// A project root contains at least one of: .gurgeh, .pollard, .coldwine
func FindProjectRoot(startPath string) (string, error) {
	path := startPath
	for {
		// Check for any tool directory (current names)
		for _, tool := range []string{"gurgeh", "pollard", "coldwine"} {
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
