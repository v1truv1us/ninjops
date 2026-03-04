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

func newNinjaDiffCmd() *cobra.Command {
	var inputFile string
	var quoteID string
	var invoiceID string
	var ref string

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Compare local generated content with remote",
		Long:  `Shows field-level differences between locally generated content and what's on Invoice Ninja.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.Ninja.APIToken == "" {
				return fmt.Errorf("Invoice Ninja API token not configured")
			}

			if quoteID != "" && invoiceID != "" {
				return fmt.Errorf("--quote-id and --invoice-id cannot be used together")
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

			if ref != "" {
				quoteSpec.Metadata.Reference = ref
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

			ctx := context.Background()

			entityType := "quote"
			entityID := quoteID
			if invoiceID != "" {
				entityType = "invoice"
				entityID = invoiceID
			}

			textDiff, fieldDiffs, err := syncer.Diff(ctx, quoteSpec, artifacts, entityType, entityID)
			if err != nil {
				return fmt.Errorf("failed to compute diff: %w", err)
			}

			fmt.Printf("=== Field Differences ===\n\n")
			fmt.Println(diff.FormatFieldDiffs(fieldDiffs))

			if textDiff.HasDiff {
				fmt.Printf("\n=== Text Differences ===\n\n")
				fmt.Println(textDiff.String())
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input QuoteSpec JSON file (required)")
	cmd.Flags().StringVar(&quoteID, "quote-id", "", "Quote ID to compare")
	cmd.Flags().StringVar(&invoiceID, "invoice-id", "", "Invoice ID to compare")
	cmd.Flags().StringVar(&ref, "ref", "", "Reference tag to search for")

	cmd.MarkFlagRequired("input")

	return cmd
}
