package app

import (
	"context"
	"fmt"

	"github.com/ninjops/ninjops/internal/config"
	"github.com/ninjops/ninjops/internal/invoiceninja"
	"github.com/spf13/cobra"
)

func newNinjaTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test Invoice Ninja API connection",
		Long:  `Tests the connection to Invoice Ninja by making a safe read request.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.Ninja.APIToken == "" {
				return fmt.Errorf("Invoice Ninja API token not configured. Set NINJOPS_NINJA_API_TOKEN environment variable")
			}

			client := invoiceninja.NewClient(cfg.Ninja)

			fmt.Printf("Testing connection to: %s\n", cfg.Ninja.BaseURL)
			fmt.Printf("Using token: %s...\n", config.RedactToken(cfg.Ninja.APIToken))

			ctx := context.Background()
			if err := client.TestConnection(ctx); err != nil {
				fmt.Printf("✗ Connection failed: %v\n", err)
				return err
			}

			fmt.Println("✓ Connection successful!")
			fmt.Println("  API is reachable and authentication is valid")

			return nil
		},
	}

	return cmd
}
