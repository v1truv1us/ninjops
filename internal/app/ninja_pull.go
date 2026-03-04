package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ninjops/ninjops/internal/invoiceninja"
	"github.com/ninjops/ninjops/internal/spec"
	"github.com/ninjops/ninjops/internal/store"
	"github.com/spf13/cobra"
)

func newNinjaPullCmd() *cobra.Command {
	var inputFile string
	var quoteID string
	var invoiceID string
	var ref string
	var output string

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull quote or invoice from Invoice Ninja",
		Long:  `Fetches a quote or invoice from Invoice Ninja and prints/displays the data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.Ninja.APIToken == "" {
				return fmt.Errorf("Invoice Ninja API token not configured")
			}

			if quoteID != "" && invoiceID != "" {
				return fmt.Errorf("--quote-id and --invoice-id cannot be used together")
			}

			if inputFile == "" && quoteID == "" && invoiceID == "" && ref == "" {
				return fmt.Errorf("one of --input, --quote-id, --invoice-id, or --ref is required")
			}

			client := invoiceninja.NewClient(cfg.Ninja)
			ctx := context.Background()

			entityType := "quote"
			var entityID string

			if quoteID != "" {
				entityType = "quote"
				entityID = quoteID
			} else if invoiceID != "" {
				entityType = "invoice"
				entityID = invoiceID
			}

			if inputFile == "" && ref == "" {
				entity, err := pullByEntityID(ctx, client, entityType, entityID)
				if err != nil {
					return err
				}
				return writeOrPrintEntity(output, entity)
			}

			lookupSpec, err := loadLookupSpec(inputFile, ref)
			if err != nil {
				return err
			}

			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to determine working directory: %w", err)
			}

			st, err := store.NewStore(wd)
			if err != nil {
				return fmt.Errorf("failed to initialize state store: %w", err)
			}

			syncer := invoiceninja.NewSyncer(client, st)
			entity, err := syncer.Pull(ctx, lookupSpec, entityType, entityID)
			if err != nil {
				return fmt.Errorf("failed to pull: %w", err)
			}

			return writeOrPrintEntity(output, entity)
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input QuoteSpec JSON file")
	cmd.Flags().StringVar(&quoteID, "quote-id", "", "Quote ID to pull")
	cmd.Flags().StringVar(&invoiceID, "invoice-id", "", "Invoice ID to pull")
	cmd.Flags().StringVar(&ref, "ref", "", "Reference tag to search for")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file for pulled data")

	return cmd
}

func loadLookupSpec(inputFile, ref string) (*spec.QuoteSpec, error) {
	if inputFile != "" {
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		quoteSpec, err := spec.FromJSON(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse QuoteSpec: %w", err)
		}

		if ref != "" {
			quoteSpec.Metadata.Reference = ref
		}

		return quoteSpec, nil
	}

	return &spec.QuoteSpec{
		Metadata: spec.Metadata{
			Reference: ref,
		},
	}, nil
}

func pullByEntityID(ctx context.Context, client *invoiceninja.Client, entityType, entityID string) (interface{}, error) {
	switch entityType {
	case "quote":
		quote, err := client.GetQuote(ctx, entityID)
		if err != nil {
			return nil, fmt.Errorf("failed to get quote: %w", err)
		}
		return quote, nil
	case "invoice":
		invoice, err := client.GetInvoice(ctx, entityID)
		if err != nil {
			return nil, fmt.Errorf("failed to get invoice: %w", err)
		}
		return invoice, nil
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}
}

func writeOrPrintEntity(output string, entity interface{}) error {
	if output != "" {
		jsonData, err := json.MarshalIndent(entity, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal output: %w", err)
		}
		if err := writeFile(output, jsonData); err != nil {
			return err
		}
		fmt.Printf("✓ Written to %s\n", output)
		return nil
	}

	switch e := entity.(type) {
	case *invoiceninja.NinjaQuote:
		fmt.Println(invoiceninja.FormatQuoteSummary(e))
	case *invoiceninja.NinjaInvoice:
		fmt.Println(invoiceninja.FormatInvoiceSummary(e))
	default:
		return fmt.Errorf("unexpected pull response type %T", entity)
	}

	return nil
}
