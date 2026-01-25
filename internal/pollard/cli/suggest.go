package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mistakeknot/autarch/internal/pollard/api"
)

var (
	suggestVision       string
	suggestProblem      string
	suggestRequirements string
)

var suggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Suggest hunters for a PRD",
	Long: `Analyze PRD content and suggest the best hunters to use.

This command analyzes the domain and content of a PRD and recommends
which hunters would provide the most relevant research.

Examples:
  pollard suggest --vision "Allergy-safe recipe platform" \
      --problem "People with food allergies can't find safe recipes" \
      --requirements "allergen detection,ingredient substitution"
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		if suggestVision == "" {
			return fmt.Errorf("--vision is required")
		}

		// Parse requirements
		var requirements []string
		if suggestRequirements != "" {
			requirements = strings.Split(suggestRequirements, ",")
			for i := range requirements {
				requirements[i] = strings.TrimSpace(requirements[i])
			}
		}

		scanner, err := api.NewScanner(cwd)
		if err != nil {
			return fmt.Errorf("failed to create scanner: %w", err)
		}
		defer scanner.Close()

		// Get hunter suggestions
		selections := scanner.SuggestHunters(suggestVision, suggestProblem, requirements)

		fmt.Println("Selected Hunters:")
		for i, sel := range selections {
			fmt.Printf("%d. %s (score: %.2f)\n", i+1, sel.Name, sel.Score)
			fmt.Printf("   Domain: %s\n", sel.Domain)
			fmt.Printf("   Reasoning: %s\n", sel.Reasoning)
			if len(sel.Queries) > 0 {
				fmt.Printf("   Queries: %s\n", strings.Join(sel.Queries, ", "))
			}
			fmt.Println()
		}

		// Check for suggested new hunter
		if suggestedHunter, ok := scanner.SuggestNewHunter(suggestVision, suggestProblem, requirements); ok {
			fmt.Println("Suggested New Hunters:")
			fmt.Printf("  - %s\n", suggestedHunter)
			fmt.Println("    Run 'pollard hunter create' to generate this hunter")
		}

		return nil
	},
}

func init() {
	suggestCmd.Flags().StringVar(&suggestVision, "vision", "", "Project vision statement (required)")
	suggestCmd.Flags().StringVar(&suggestProblem, "problem", "", "Problem being solved")
	suggestCmd.Flags().StringVar(&suggestRequirements, "requirements", "", "Comma-separated requirements")
}
