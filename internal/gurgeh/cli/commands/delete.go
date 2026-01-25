package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mistakeknot/autarch/internal/gurgeh/archive"
	"github.com/mistakeknot/autarch/internal/gurgeh/project"
	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
	"github.com/mistakeknot/autarch/internal/gurgeh/tui"
	"github.com/spf13/cobra"
)

func DeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a PRD spec (reversible)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			id := strings.TrimSpace(args[0])
			if err := deleteCmdRun(root, id); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Deleted", id)
			return nil
		},
	}
}

func deleteCmdRun(root, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("missing id")
	}
	prevStatus := ""
	path := filepath.Join(project.SpecsDir(root), id+".yaml")
	if spec, err := specs.LoadSpec(path); err == nil {
		prevStatus = spec.Status
	}
	res, err := archive.Delete(root, id)
	if err != nil {
		return err
	}
	return recordLastAction(root, tui.LastAction{Type: "delete", ID: id, PrevStatus: prevStatus, From: res.From, To: res.To})
}
