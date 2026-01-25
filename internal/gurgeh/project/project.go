package project

import (
	"os"
	"path/filepath"

	"github.com/mistakeknot/autarch/internal/gurgeh/config"
)

const GurgDir = ".gurgeh"
const LegacyPraudeDir = ".praude" // backward compatibility

// RootDir returns the gurgeh data directory, falling back to .praude if it exists.
func RootDir(root string) string {
	gurgPath := filepath.Join(root, GurgDir)
	if _, err := os.Stat(gurgPath); err == nil {
		return gurgPath
	}
	// Fallback to legacy path
	praudePath := filepath.Join(root, LegacyPraudeDir)
	if _, err := os.Stat(praudePath); err == nil {
		return praudePath
	}
	// Default to new path for new projects
	return gurgPath
}

func SpecsDir(root string) string {
	return filepath.Join(RootDir(root), "specs")
}

func ResearchDir(root string) string {
	return filepath.Join(RootDir(root), "research")
}

func SuggestionsDir(root string) string {
	return filepath.Join(RootDir(root), "suggestions")
}

func BriefsDir(root string) string {
	return filepath.Join(RootDir(root), "briefs")
}

func ConfigPath(root string) string {
	return filepath.Join(RootDir(root), "config.toml")
}

func StatePath(root string) string {
	return filepath.Join(RootDir(root), "state.json")
}

func ArchivedDir(root string) string {
	return filepath.Join(RootDir(root), "archived")
}

func ArchivedSpecsDir(root string) string {
	return filepath.Join(ArchivedDir(root), "specs")
}

func ArchivedResearchDir(root string) string {
	return filepath.Join(ArchivedDir(root), "research")
}

func ArchivedSuggestionsDir(root string) string {
	return filepath.Join(ArchivedDir(root), "suggestions")
}

func ArchivedBriefsDir(root string) string {
	return filepath.Join(ArchivedDir(root), "briefs")
}

func TrashDir(root string) string {
	return filepath.Join(RootDir(root), "trash")
}

func TrashSpecsDir(root string) string {
	return filepath.Join(TrashDir(root), "specs")
}

func TrashResearchDir(root string) string {
	return filepath.Join(TrashDir(root), "research")
}

func TrashSuggestionsDir(root string) string {
	return filepath.Join(TrashDir(root), "suggestions")
}

func TrashBriefsDir(root string) string {
	return filepath.Join(TrashDir(root), "briefs")
}

func Init(root string) error {
	dirs := []string{
		RootDir(root),
		SpecsDir(root),
		ResearchDir(root),
		SuggestionsDir(root),
		BriefsDir(root),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	if _, err := os.Stat(ConfigPath(root)); os.IsNotExist(err) {
		if err := os.WriteFile(ConfigPath(root), []byte(config.DefaultConfigToml), 0o644); err != nil {
			return err
		}
	}
	return nil
}
