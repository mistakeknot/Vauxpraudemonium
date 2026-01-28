package commands

import (
	"fmt"
	"os"

	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
	"github.com/spf13/cobra"
)

// HistoryCmd shows the revision history for a spec.
func HistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history <spec-id>",
		Short: "Show spec revision history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}

			revisions, err := specs.LoadHistory(root, args[0])
			if err != nil {
				return fmt.Errorf("loading history: %w", err)
			}

			if len(revisions) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No revisions found for "+args[0])
				return nil
			}

			for _, rev := range revisions {
				fmt.Fprintf(cmd.OutOrStdout(), "v%d  %s  by %s  trigger: %s\n",
					rev.Version, rev.Timestamp.Format("2006-01-02 15:04"), rev.Author, rev.Trigger)
				for _, c := range rev.Changes {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s: %q → %q", c.Field, c.Before, c.After)
					if c.Reason != "" {
						fmt.Fprintf(cmd.OutOrStdout(), " (%s)", c.Reason)
					}
					fmt.Fprintln(cmd.OutOrStdout())
				}
			}
			return nil
		},
	}
}
