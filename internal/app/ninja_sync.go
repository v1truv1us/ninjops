package app

import (
	"context"
	"fmt"
	"os"

	"github.com/ninjops/ninjops/internal/diff"
	"github.com/ninjops/ninjops/internal/generate"
	"github.com/ninjops/ninjops/internal/invoiceninja"
	"github.com/ninjops/ninjops/internal/spec"
	"github.com/ninjops/ninjops/internal/store"
	"github.com/spf13/cobra"
)

func newNinjaSyncCmd() *cobra.Command {
	var inputFile string
	var mode string
	var dryRun bool
	var showDiff bool
	var confirm bool
	var yes bool
	var quoteID string
	var invoiceID string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync QuoteSpec with Invoice Ninja",
		Long:  `Synchronizes a QuoteSpec with Invoice Ninja: ensures client exists, creates or updates quote/invoice.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.Ninja.APIToken == "" {
				return fmt.Errorf("Invoice Ninja API token not configured")
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

			if err := spec.Validate(quoteSpec); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			generator := generate.NewGenerator()
			artifacts, err := generator.Generate(quoteSpec)
			if err != nil {
				return fmt.Errorf("failed to generate artifacts: %w", err)
			}

			client := invoiceninja.NewClient(cfg.Ninja)
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to determine working directory: %w", err)
			}
			st, err := store.NewStore(wd)
			if err != nil {
				return fmt.Errorf("failed to initialize state store: %w", err)
			}

			syncer := invoiceninja.NewSyncer(client, st)

			syncMode := invoiceninja.SyncMode(mode)
			if syncMode == "" {
				syncMode = invoiceninja.SyncModeQuote
			}

			opts := invoiceninja.SyncOptions{
				Mode:      syncMode,
				DryRun:    dryRun,
				ShowDiff:  showDiff,
				Confirm:   confirm,
				QuoteID:   quoteID,
				InvoiceID: invoiceID,
			}

			if !dryRun && !yes {
				fmt.Printf("Ready to sync: %s\n", quoteSpec.Project.Name)
				fmt.Printf("  Client: %s\n", quoteSpec.Client.Name)
				fmt.Printf("  Mode: %s\n", mode)
				fmt.Printf("  Reference: %s\n", quoteSpec.Metadata.Reference)
				fmt.Printf("\nProceed? [y/N]: ")

				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Cancelled")
					return nil
				}
			}

			ctx := context.Background()
			result, err := syncer.Sync(ctx, quoteSpec, artifacts, opts)
			if err != nil {
				return fmt.Errorf("sync failed: %w", err)
			}

			fmt.Println("\n=== Sync Results ===")

			if result.ClientCreated {
				fmt.Printf("✓ Created client: %s\n", result.ClientID)
			} else if result.ClientID != "" {
				fmt.Printf("✓ Found existing client: %s\n", result.ClientID)
			}

			if result.ProjectCreated {
				fmt.Printf("✓ Created project: %s\n", result.ProjectID)
			} else if result.ProjectUpdated {
				fmt.Printf("✓ Updated project: %s\n", result.ProjectID)
			} else if result.ProjectID != "" {
				fmt.Printf("✓ Project unchanged: %s\n", result.ProjectID)
			}

			if syncMode == invoiceninja.SyncModeQuote || syncMode == invoiceninja.SyncModeBoth {
				if result.QuoteCreated {
					fmt.Printf("✓ Created quote: %s\n", result.QuoteID)
				} else if result.QuoteUpdated {
					fmt.Printf("✓ Updated quote: %s\n", result.QuoteID)
				} else if result.QuoteID != "" {
					fmt.Printf("✓ Quote unchanged: %s\n", result.QuoteID)
				}
			}

			if syncMode == invoiceninja.SyncModeInvoice || syncMode == invoiceninja.SyncModeBoth {
				if result.InvoiceCreated {
					fmt.Printf("✓ Created invoice: %s\n", result.InvoiceID)
				} else if result.InvoiceUpdated {
					fmt.Printf("✓ Updated invoice: %s\n", result.InvoiceID)
				} else if result.InvoiceID != "" {
					fmt.Printf("✓ Invoice unchanged: %s\n", result.InvoiceID)
				}
			}

			if len(result.Diffs) > 0 {
				fmt.Println("\n" + diff.FormatFieldDiffs(result.Diffs))
			}

			if dryRun {
				fmt.Println("\n(Dry run - no changes made)")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input QuoteSpec JSON file (required)")
	cmd.Flags().StringVarP(&mode, "mode", "m", "quote", "Sync mode: quote, invoice, or both")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	cmd.Flags().BoolVarP(&showDiff, "diff", "d", false, "Show field-level differences")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Require confirmation before making changes")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&quoteID, "quote-id", "", "Specific quote ID to update")
	cmd.Flags().StringVar(&invoiceID, "invoice-id", "", "Specific invoice ID to update")

	cmd.MarkFlagRequired("input")

	return cmd
}
