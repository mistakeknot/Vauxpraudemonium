package commands

import (
	"fmt"
	"os"

	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
	"github.com/spf13/cobra"
)

// DiffCmd shows structured differences between two spec versions.
func DiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff <spec-id> <v1> <v2>",
		Short: "Compare two spec versions",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}

			specID := args[0]
			v1, err := specs.ParseVersion(args[1])
			if err != nil {
				return fmt.Errorf("invalid version %q: %w", args[1], err)
			}
			v2, err := specs.ParseVersion(args[2])
			if err != nil {
				return fmt.Errorf("invalid version %q: %w", args[2], err)
			}

			spec1, err := specs.LoadRevisionSpec(root, specID, v1)
			if err != nil {
				return fmt.Errorf("loading v%d: %w", v1, err)
			}
			spec2, err := specs.LoadRevisionSpec(root, specID, v2)
			if err != nil {
				return fmt.Errorf("loading v%d: %w", v2, err)
			}

			entries := specs.DiffSpecs(&spec1, &spec2)
			fmt.Fprintf(cmd.OutOrStdout(), "Diff %s v%d → v%d:\n", specID, v1, v2)
			fmt.Fprintln(cmd.OutOrStdout(), specs.FormatDiff(entries))
			return nil
		},
	}
}
