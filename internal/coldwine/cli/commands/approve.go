package commands

import (
	"fmt"

	"github.com/mistakeknot/autarch/internal/coldwine/git"
	"github.com/mistakeknot/autarch/internal/coldwine/project"
	"github.com/mistakeknot/autarch/internal/coldwine/storage"
	"github.com/spf13/cobra"
)

func ApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve a task by merging its branch and marking done",
		Args: wrapArgs("approve", func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return err
			}
			return project.ValidateTaskID(args[0])
		}),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer func() {
				if err != nil {
					err = wrapCommandError("approve", err)
				}
			}()
			taskID := args[0]
			branch := args[1]

			if err := git.MergeBranch(&git.ExecRunner{}, branch); err != nil {
				return err
			}
			root, err := project.FindRoot(".")
			if err != nil {
				return err
			}
			db, err := storage.Open(project.StateDBPath(root))
			if err != nil {
				return err
			}
			defer db.Close()
			if err := storage.ApproveTask(db, taskID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Approved %s (merged %s)\n", taskID, branch)
			return nil
		},
	}
	return cmd
}
