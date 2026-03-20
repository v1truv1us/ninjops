package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ninjops/ninjops/internal/agents"
	"github.com/ninjops/ninjops/internal/config"
	"github.com/spf13/cobra"
)

func newConfigureCmd() *cobra.Command {
	defaults := activeConfig()

	var baseURL string
	var apiToken string
	var apiSecret string
	var provider string
	var plan string
	var model string
	var listen string
	var port int
	var output string
	var authCredsOutput string
	var format string
	var nonInteractive bool
	var providerAPIKey string
	var skipProviderTest bool

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Write ninjops onboarding config",
		Long:  "Writes non-secret runtime settings to config and secrets to auth-creds.json.",
		RunE: func(cmd *cobra.Command, args []string) error {
			appCfg := activeConfig()

			if !slices.Contains([]string{"json", "jsonc"}, format) {
				return fmt.Errorf("invalid format %q (valid: json, jsonc)", format)
			}

			if !slices.Contains(config.ValidPlans, plan) {
				return fmt.Errorf("invalid plan %q (valid: %s)", plan, strings.Join(config.ValidPlans, ", "))
			}

			resolvedProvider, resolvedModel := config.ResolveProviderModel(provider, model)
			if !config.IsValidProvider(resolvedProvider) {
				return fmt.Errorf("invalid provider %q (valid: %s)", resolvedProvider, strings.Join(config.ValidProviders, ", "))
			}

			if port < 1 || port > 65535 {
				return fmt.Errorf("invalid port %d (must be 1-65535)", port)
			}

			resolvedProviderAPIKey := providerAPIKey
			if resolvedProviderAPIKey == "" {
				resolvedProviderAPIKey = config.ResolveProviderAPIKey(resolvedProvider, appCfg.Agent.ProviderAPIKey)
			}

			if resolvedProvider != "offline" {
				envVar := config.ProviderAPIKeyEnvHint(resolvedProvider)
				if resolvedProviderAPIKey == "" {
					if envVar != "" {
						return fmt.Errorf("%s provider requires an API key; set --provider-api-key, %s, or agent.provider_api_key in config", resolvedProvider, envVar)
					}
					return fmt.Errorf("%s provider requires an API key; set --provider-api-key or agent.provider_api_key in config", resolvedProvider)
				}

				if !skipProviderTest {
					if err := agents.CheckProviderConnectionWithModel(cmd.Context(), resolvedProvider, resolvedModel, resolvedProviderAPIKey); err != nil {
						return fmt.Errorf("provider connection check failed: %w", err)
					}
				}
			}

			configPath, err := resolveConfigureOutputPath(output, format)
			if err != nil {
				return err
			}

			authCredsPath, err := resolveConfigureAuthCredsPath(configPath, authCredsOutput)
			if err != nil {
				return err
			}

			cfgData := map[string]interface{}{
				"ninja": map[string]string{
					"base_url": baseURL,
				},
				"agent": map[string]string{
					"provider": resolvedProvider,
					"plan":     plan,
					"model":    resolvedModel,
				},
				"serve": map[string]interface{}{
					"listen": listen,
					"port":   port,
				},
				"auth_creds_file": authCredsPath,
			}

			authData := map[string]interface{}{
				"ninja": map[string]string{
					"api_token":  apiToken,
					"api_secret": apiSecret,
				},
				"agent": map[string]string{
					"provider_api_key": resolvedProviderAPIKey,
				},
			}

			configJSONData, err := json.MarshalIndent(cfgData, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize config: %w", err)
			}

			authJSONData, err := json.MarshalIndent(authData, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize auth credentials: %w", err)
			}

			if err := os.MkdirAll(filepath.Dir(configPath), 0750); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
			if err := os.MkdirAll(filepath.Dir(authCredsPath), 0750); err != nil {
				return fmt.Errorf("failed to create auth credentials directory: %w", err)
			}

			if err := os.WriteFile(configPath, configJSONData, 0600); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
			if err := os.WriteFile(authCredsPath, authJSONData, 0600); err != nil {
				return fmt.Errorf("failed to write auth credentials file: %w", err)
			}

			fmt.Printf("✓ Wrote config to %s\n", configPath)
			fmt.Printf("✓ Wrote auth credentials to %s\n", authCredsPath)
			if resolvedProvider != "offline" {
				if skipProviderTest {
					fmt.Printf("✓ Skipped %s provider connection check\n", resolvedProvider)
				} else {
					fmt.Printf("✓ Verified %s provider connection\n", resolvedProvider)
				}
			}
			fmt.Printf("✓ Selected model %s\n", resolvedModel)
			if nonInteractive {
				fmt.Println("✓ Non-interactive onboarding complete")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&baseURL, "base-url", defaults.Ninja.BaseURL, "Invoice Ninja base URL")
	cmd.Flags().StringVar(&apiToken, "api-token", defaults.Ninja.APIToken, "Invoice Ninja API token")
	cmd.Flags().StringVar(&apiSecret, "api-secret", defaults.Ninja.APISecret, "Invoice Ninja API secret")
	cmd.Flags().StringVar(&provider, "provider", defaults.Agent.Provider, "Agent provider")
	cmd.Flags().StringVar(&plan, "plan", defaults.Agent.Plan, "Agent plan")
	cmd.Flags().StringVar(&model, "model", "", "Agent model (supports alias openai-codex)")
	cmd.Flags().StringVar(&listen, "listen", defaults.Serve.Listen, "Serve listen address")
	cmd.Flags().IntVar(&port, "port", defaults.Serve.Port, "Serve port")
	cmd.Flags().StringVar(&format, "format", "json", "Output format: json or jsonc")
	cmd.Flags().StringVar(&output, "output", "", "Output path override")
	cmd.Flags().StringVar(&authCredsOutput, "auth-creds-output", "", "Auth credentials output path override")
	cmd.Flags().StringVar(&providerAPIKey, "provider-api-key", defaults.Agent.ProviderAPIKey, "Provider API key (OpenAI/Anthropic/future providers)")
	cmd.Flags().BoolVar(&skipProviderTest, "skip-provider-test", false, "Skip provider connection test during configure")
	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Disable prompts for automation")

	return cmd
}

func resolveConfigureOutputPath(output string, format string) (string, error) {
	if output != "" {
		return output, nil
	}

	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("HOME is not set")
	}

	baseDir := filepath.Join(home, ".config", "ninjops")
	if format == "jsonc" {
		return filepath.Join(baseDir, "ninjops.jsonc"), nil
	}

	return filepath.Join(baseDir, "config.json"), nil
}

func resolveConfigureAuthCredsPath(configPath string, authCredsOutput string) (string, error) {
	if authCredsOutput != "" {
		return authCredsOutput, nil
	}

	if configPath == "" {
		return "", fmt.Errorf("config path is required to resolve auth credentials path")
	}

	return filepath.Join(filepath.Dir(configPath), "auth-creds.json"), nil
}
