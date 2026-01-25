package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pollardPlan "github.com/mistakeknot/autarch/internal/pollard/plan"
	"github.com/mistakeknot/autarch/internal/pollard/reports"
)

var (
	reportType     string
	reportStdout   bool
	reportPlanMode bool
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a research report",
	Long: `Generate a research report summarizing collected intelligence.

Report types:
  landscape   - Comprehensive overview of all collected data (default)
  competitive - Focus on competitor activity and threats
  trends      - Industry trends from HackerNews and other sources
  research    - Academic papers from arXiv and research sources

Examples:
  pollard report                    # Generate landscape report
  pollard report --type competitive # Generate competitive analysis
  pollard report --type trends      # Generate trends report
  pollard report --stdout           # Output to stdout instead of file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		// Plan mode - generate JSON plan
		if reportPlanMode {
			p, err := pollardPlan.GenerateReportPlan(pollardPlan.ReportPlanOptions{
				Root:       cwd,
				ReportType: reportType,
			})
			if err != nil {
				return err
			}

			planPath, err := p.Save(cwd)
			if err != nil {
				return err
			}

			data, err := json.MarshalIndent(p, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			fmt.Printf("\nPlan saved to: %s\n", planPath)
			fmt.Println("Run 'pollard apply' to execute this plan.")
			return nil
		}

		generator := reports.NewGenerator(cwd)

		var rType reports.ReportType
		switch reportType {
		case "landscape":
			rType = reports.TypeLandscape
		case "competitive":
			rType = reports.TypeCompetitive
		case "trends":
			rType = reports.TypeTrends
		case "research":
			rType = reports.TypeResearch
		default:
			rType = reports.TypeLandscape
		}

		filePath, err := generator.Generate(rType)
		if err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}

		if reportStdout {
			// Read and print the file
			content, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read report: %w", err)
			}
			fmt.Print(string(content))
		} else {
			fmt.Printf("Report generated: %s\n", filePath)
		}

		return nil
	},
}

func init() {
	reportCmd.Flags().StringVar(&reportType, "type", "landscape", "Report type: landscape, competitive, trends, research")
	reportCmd.Flags().BoolVar(&reportStdout, "stdout", false, "Output report to stdout instead of file")
	reportCmd.Flags().BoolVar(&reportPlanMode, "plan", false, "Generate plan JSON instead of executing")
}
