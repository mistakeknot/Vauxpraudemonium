package tui

import "github.com/mistakeknot/autarch/internal/coldwine/git"

func LoadDiffFiles(r git.Runner, rev string) ([]string, error) {
	return git.DiffNameOnly(r, rev)
}
