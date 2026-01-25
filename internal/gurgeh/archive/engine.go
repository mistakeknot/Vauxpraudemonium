package archive

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mistakeknot/autarch/internal/gurgeh/project"
	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
)

type Result struct {
	From []string
	To   []string
}

func Archive(root, id string) (Result, error) {
	return movePRD(root, id, true)
}

func Delete(root, id string) (Result, error) {
	return movePRD(root, id, false)
}

func Undo(root string, from, to []string) error {
	if len(from) != len(to) {
		return fmt.Errorf("undo path mismatch")
	}
	for i := range to {
		if err := moveFile(to[i], from[i]); err != nil {
			return err
		}
	}
	return nil
}

func movePRD(root, id string, archived bool) (Result, error) {
	if strings.TrimSpace(id) == "" {
		return Result{}, fmt.Errorf("missing id")
	}
	moves := [][2]string{}
	if from, to, err := moveSpec(root, id, archived); err == nil {
		moves = append(moves, [2]string{from, to})
	} else {
		return Result{}, err
	}
	artifactMoves, err := moveArtifacts(root, id, archived)
	if err != nil {
		return Result{}, err
	}
	moves = append(moves, artifactMoves...)
	res := Result{From: []string{}, To: []string{}}
	for _, pair := range moves {
		res.From = append(res.From, pair[0])
		res.To = append(res.To, pair[1])
	}
	return res, nil
}

func moveSpec(root, id string, archived bool) (string, string, error) {
	srcSpec := filepath.Join(project.SpecsDir(root), id+".yaml")
	dstSpec := filepath.Join(project.ArchivedSpecsDir(root), id+".yaml")
	if !archived {
		dstSpec = filepath.Join(project.TrashSpecsDir(root), id+".yaml")
	}
	if err := moveFile(srcSpec, dstSpec); err != nil {
		return "", "", err
	}
	if archived {
		_ = specs.UpdateStatus(dstSpec, "archived")
	}
	return srcSpec, dstSpec, nil
}

func moveArtifacts(root, id string, archived bool) ([][2]string, error) {
	moves := [][2]string{}
	researchPrefix := id + "-"
	appendMoves := func(srcDir, dstDir string) error {
		entries, err := os.ReadDir(srcDir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !(strings.HasPrefix(name, researchPrefix) || strings.HasPrefix(name, id+".")) {
				continue
			}
			src := filepath.Join(srcDir, name)
			dst := filepath.Join(dstDir, name)
			if err := moveFile(src, dst); err != nil {
				return err
			}
			moves = append(moves, [2]string{src, dst})
		}
		return nil
	}
	var researchDst, suggestionsDst, briefsDst string
	if archived {
		researchDst = project.ArchivedResearchDir(root)
		suggestionsDst = project.ArchivedSuggestionsDir(root)
		briefsDst = project.ArchivedBriefsDir(root)
	} else {
		researchDst = project.TrashResearchDir(root)
		suggestionsDst = project.TrashSuggestionsDir(root)
		briefsDst = project.TrashBriefsDir(root)
	}
	if err := appendMoves(project.ResearchDir(root), researchDst); err != nil {
		return nil, err
	}
	if err := appendMoves(project.SuggestionsDir(root), suggestionsDst); err != nil {
		return nil, err
	}
	if err := appendMoves(project.BriefsDir(root), briefsDst); err != nil {
		return nil, err
	}
	return moves, nil
}

func moveFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Rename(src, dst)
}
