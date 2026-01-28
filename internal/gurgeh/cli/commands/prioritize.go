package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mistakeknot/autarch/internal/gurgeh/prioritize"
	"github.com/mistakeknot/autarch/internal/gurgeh/project"
	"github.com/mistakeknot/autarch/internal/gurgeh/specs"
	"github.com/spf13/cobra"
)

// PrioritizeCmd ranks features using agent-powered synthesis.
func PrioritizeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prioritize <spec-id>",
		Short: "Rank features by what to build next",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}

			specPath := filepath.Join(project.SpecsDir(root), args[0]+".yaml")
			spec, err := specs.LoadSpec(specPath)
			if err != nil {
				return fmt.Errorf("loading spec: %w", err)
			}

			input := prioritize.RankingInput{
				Spec: &spec,
			}

			ranker := prioritize.NewRanker()
			result, err := ranker.Rank(input)
			if err != nil {
				return fmt.Errorf("ranking: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Feature Priority Ranking")
			fmt.Fprintln(cmd.OutOrStdout(), "========================")
			for _, item := range result.Items {
				fmt.Fprintf(cmd.OutOrStdout(), "\n#%d %s (%s)\n", item.Rank, item.Title, item.FeatureID)
				fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", item.Reasoning)
				fmt.Fprintf(cmd.OutOrStdout(), "    Confidence: %.0f%%\n", item.Confidence*100)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", result.Summary)
			return nil
		},
	}
}
