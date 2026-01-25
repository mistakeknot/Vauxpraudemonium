package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/mistakeknot/autarch/internal/pollard/api"
)

var (
	researchVision       string
	researchProblem      string
	researchRequirements string
	researchBriefOnly    bool
)

var researchCmd = &cobra.Command{
	Use:   "research",
	Short: "Conduct intelligent agent-driven research",
	Long: `Conduct research using your AI agent as the primary capability.

This command uses the agent-native architecture where your existing AI agent
(Claude, Codex, etc.) conducts research via web search and document analysis.
API-based hunters supplement the agent research when keys are configured.

Examples:
  # Research for a PRD
  pollard research --vision "Allergy-safe recipe platform" \
      --problem "People with food allergies can't find safe recipes" \
      --requirements "allergen detection,ingredient substitution"

  # Generate research brief only (don't invoke agent)
  pollard research --vision "Allergy-safe recipe platform" --brief-only

Environment Variables:
  POLLARD_AGENT_COMMAND  - AI agent command (default: claude)
  POLLARD_AGENT_ARGS     - Arguments for agent (default: --print)
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		if researchVision == "" {
			return fmt.Errorf("--vision is required")
		}

		// Parse requirements
		var requirements []string
		if researchRequirements != "" {
			requirements = strings.Split(researchRequirements, ",")
			for i := range requirements {
				requirements[i] = strings.TrimSpace(requirements[i])
			}
		}

		scanner, err := api.NewScanner(cwd)
		if err != nil {
			return fmt.Errorf("failed to create scanner: %w", err)
		}
		defer scanner.Close()

		// Brief-only mode: just generate and display the research brief
		if researchBriefOnly {
			brief := scanner.GetResearchBrief(researchVision, researchProblem, requirements)
			fmt.Println(brief)
			return nil
		}

		// Set up context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle interrupt signals
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			fmt.Println("\nInterrupted, stopping research...")
			cancel()
		}()

		fmt.Println("Starting intelligent research...")
		fmt.Printf("Vision: %s\n", researchVision)
		if researchProblem != "" {
			fmt.Printf("Problem: %s\n", researchProblem)
		}
		if len(requirements) > 0 {
			fmt.Printf("Requirements: %s\n", strings.Join(requirements, ", "))
		}
		fmt.Println()

		// Suggest hunters
		selections := scanner.SuggestHunters(researchVision, researchProblem, requirements)
		if len(selections) > 0 {
			fmt.Println("Recommended API hunters (will run if keys available):")
			for _, sel := range selections {
				fmt.Printf("  - %s (score: %.2f) - %s\n", sel.Name, sel.Score, sel.Reasoning)
			}
			fmt.Println()
		}

		// Check for suggested new hunter
		if suggestedHunter, ok := scanner.SuggestNewHunter(researchVision, researchProblem, requirements); ok {
			fmt.Printf("Suggested custom hunter: %s\n", suggestedHunter)
			fmt.Println("  Run 'pollard hunter create' to generate this hunter")
			fmt.Println()
		}

		// Run intelligent research
		fmt.Println("Invoking AI agent for research...")
		result, err := scanner.IntelligentResearch(ctx, researchVision, researchProblem, requirements)
		if err != nil {
			return fmt.Errorf("research failed: %w", err)
		}

		// Print results
		fmt.Printf("\nResearch complete: %d sources collected\n", result.TotalSources)
		fmt.Println("\nHunters run:")
		for name, hr := range result.HunterResults {
			status := "✓"
			if !hr.Success() {
				status = "✗"
			}
			fmt.Printf("  %s %s: %d sources\n", status, name, hr.SourcesCollected)
		}

		if len(result.OutputFiles) > 0 {
			fmt.Println("\nOutput files:")
			for _, f := range result.OutputFiles {
				fmt.Printf("  - %s\n", f)
			}
		}

		if len(result.Errors) > 0 {
			fmt.Println("\nWarnings/Errors:")
			for _, e := range result.Errors {
				fmt.Printf("  - %v\n", e)
			}
		}

		return nil
	},
}

func init() {
	researchCmd.Flags().StringVar(&researchVision, "vision", "", "Project vision statement (required)")
	researchCmd.Flags().StringVar(&researchProblem, "problem", "", "Problem being solved")
	researchCmd.Flags().StringVar(&researchRequirements, "requirements", "", "Comma-separated requirements")
	researchCmd.Flags().BoolVar(&researchBriefOnly, "brief-only", false, "Generate research brief without invoking agent")
}
