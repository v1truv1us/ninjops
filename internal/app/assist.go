package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ninjops/ninjops/internal/agents"
	"github.com/ninjops/ninjops/internal/config"
	"github.com/ninjops/ninjops/internal/spec"
	"github.com/spf13/cobra"
)

func newAssistCmd() *cobra.Command {
	var inputFile string
	var write bool
	var provider string
	var plan string
	var model string

	cmd := &cobra.Command{
		Use:   "assist [role]",
		Short: "Get AI assistance with a QuoteSpec",
		Long: `Assist uses AI to help improve your QuoteSpec. Available roles:
  - clarify: Normalize features, extract responsibilities, identify gaps
  - polish: Improve tone, tighten prose, enhance professionalism  
  - boundaries: Generate scope boundaries, assumptions, and exclusions
  - line-items: Suggest billable line items from features`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			role := args[0]

			if !agents.IsValidRole(role) {
				return fmt.Errorf("invalid role: %s (valid: clarify, polish, boundaries, line-items)", role)
			}

			if inputFile == "" {
				return fmt.Errorf("--input is required")
			}

			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			quoteSpec, err := spec.FromJSON(data)
			if err != nil {
				return fmt.Errorf("failed to parse QuoteSpec: %w", err)
			}

			providerName := provider
			if providerName == "" {
				providerName = cfg.Agent.Provider
			}

			planName := agents.Plan(plan)
			if planName == "" {
				planName = agents.Plan(cfg.Agent.Plan)
			}

			modelName := model
			if modelName == "" {
				modelName = cfg.Agent.Model
			}
			providerName, modelName = config.ResolveProviderModel(providerName, modelName)

			apiKey := config.ResolveProviderAPIKey(providerName, cfg.Agent.ProviderAPIKey)
			agentProvider := agents.GetProvider(providerName, apiKey)

			if !agentProvider.IsAvailable() {
				return fmt.Errorf("%s provider not available - check API key configuration", providerName)
			}

			req := agents.AgentRequest{
				Role:      agents.Role(role),
				Plan:      planName,
				Model:     modelName,
				QuoteSpec: quoteSpec,
			}

			ctx := context.Background()
			response, err := agentProvider.Execute(ctx, req)
			if err != nil {
				return fmt.Errorf("agent execution failed: %w", err)
			}

			fmt.Printf("Role: %s | Provider: %s | Plan: %s | Model: %s\n", role, providerName, planName, modelName)
			fmt.Printf("Confidence: %.2f\n\n", response.Confidence)

			if len(response.Suggestions) > 0 {
				fmt.Println("Suggestions:")
				for _, s := range response.Suggestions {
					fmt.Printf("  • %s\n", s)
				}
				fmt.Println()
			}

			if write {
				updatedJSON, err := response.QuoteSpec.ToJSON()
				if err != nil {
					return fmt.Errorf("failed to marshal updated QuoteSpec: %w", err)
				}

				if err := os.WriteFile(inputFile, updatedJSON, 0644); err != nil {
					return fmt.Errorf("failed to write file: %w", err)
				}

				fmt.Printf("✓ Updated %s\n", inputFile)
			} else {
				fmt.Println("Updated QuoteSpec:")
				updatedJSON, _ := json.MarshalIndent(response.QuoteSpec, "", "  ")
				fmt.Println(string(updatedJSON))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input QuoteSpec JSON file (required)")
	cmd.Flags().BoolVarP(&write, "write", "w", false, "Write changes back to input file")
	cmd.Flags().StringVar(&provider, "provider", "", "Agent provider")
	cmd.Flags().StringVar(&plan, "plan", "", "Agent plan (default, codex-pro, opencode-zen, zai-plan)")
	cmd.Flags().StringVar(&model, "model", "", "Model override (supports alias openai-codex)")

	cmd.MarkFlagRequired("input")

	return cmd
}
