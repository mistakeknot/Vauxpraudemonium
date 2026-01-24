package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/mistakeknot/vauxpraudemonium/internal/pollard/api"
	"github.com/mistakeknot/vauxpraudemonium/internal/pollard/hunters"
)

var (
	hunterDomain  string
	hunterContext string
)

var hunterCmd = &cobra.Command{
	Use:   "hunter",
	Short: "Manage custom hunters",
	Long:  `Create, list, and manage custom hunters for specialized domains.`,
}

var hunterCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a custom hunter using AI",
	Long: `Create a custom hunter for a specific domain using your AI agent.

The AI agent will analyze the domain and suggest an appropriate API
or recommend using agent-based research if no suitable API exists.

Examples:
  pollard hunter create recipe-hunter --domain culinary
  pollard hunter create weather-hunter --domain weather --context "need historical data"
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		if hunterDomain == "" {
			hunterDomain = name // Use name as domain if not specified
		}

		// Set up context with cancellation
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			fmt.Println("\nInterrupted...")
			cancel()
		}()

		scanner, err := api.NewScanner(cwd)
		if err != nil {
			return fmt.Errorf("failed to create scanner: %w", err)
		}
		defer scanner.Close()

		fmt.Printf("Creating hunter for domain: %s\n", hunterDomain)
		fmt.Println("Invoking AI agent to design hunter...")

		spec, err := scanner.CreateCustomHunter(ctx, hunterDomain, hunterContext)
		if err != nil {
			return fmt.Errorf("failed to create hunter: %w", err)
		}

		fmt.Printf("\nCreated hunter: %s\n", spec.Name)
		fmt.Printf("Description: %s\n", spec.Description)

		if spec.NoAPI {
			fmt.Println("\nNo suitable API found.")
			fmt.Printf("Recommendation: %s\n", spec.Recommendation)
		} else {
			fmt.Printf("API Endpoint: %s\n", spec.APIEndpoint)
			fmt.Printf("\nHunter saved to: .pollard/hunters/custom/%s.yaml\n", spec.Name)
		}

		return nil
	},
}

var hunterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List custom hunters",
	Long:  `List all custom hunters configured for this project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		names, err := hunters.ListCustomHunters(cwd)
		if err != nil {
			return fmt.Errorf("failed to list hunters: %w", err)
		}

		if len(names) == 0 {
			fmt.Println("No custom hunters configured.")
			fmt.Println("Use 'pollard hunter create' to create one.")
			return nil
		}

		fmt.Println("Custom Hunters:")
		for _, name := range names {
			spec, err := hunters.GetCustomHunterSpec(cwd, name)
			if err != nil {
				fmt.Printf("  - %s (error loading spec)\n", name)
				continue
			}

			if spec.NoAPI {
				fmt.Printf("  - %s (agent-based)\n", name)
			} else {
				fmt.Printf("  - %s (%s)\n", name, spec.APIEndpoint)
			}
		}

		return nil
	},
}

var hunterDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a custom hunter",
	Long:  `Delete a custom hunter configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		if err := hunters.DeleteCustomHunter(cwd, name); err != nil {
			return fmt.Errorf("failed to delete hunter: %w", err)
		}

		fmt.Printf("Deleted hunter: %s\n", name)
		return nil
	},
}

var hunterShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show details of a custom hunter",
	Long:  `Display the configuration of a custom hunter.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		spec, err := hunters.GetCustomHunterSpec(cwd, name)
		if err != nil {
			return fmt.Errorf("failed to load hunter: %w", err)
		}

		fmt.Printf("Name: %s\n", spec.Name)
		fmt.Printf("Description: %s\n", spec.Description)

		if spec.NoAPI {
			fmt.Println("Type: Agent-based (no API)")
			fmt.Printf("Recommendation: %s\n", spec.Recommendation)
		} else {
			fmt.Println("Type: API-based")
			fmt.Printf("Endpoint: %s\n", spec.APIEndpoint)
		}

		return nil
	},
}

func init() {
	hunterCreateCmd.Flags().StringVar(&hunterDomain, "domain", "", "Domain for the hunter")
	hunterCreateCmd.Flags().StringVar(&hunterContext, "context", "", "Additional context for the AI agent")

	hunterCmd.AddCommand(hunterCreateCmd)
	hunterCmd.AddCommand(hunterListCmd)
	hunterCmd.AddCommand(hunterDeleteCmd)
	hunterCmd.AddCommand(hunterShowCmd)
}
