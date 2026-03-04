package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ninjops/ninjops/internal/invoiceninja"
	"github.com/spf13/cobra"
)

type quoteConverter interface {
	GetQuote(ctx context.Context, id string) (*invoiceninja.NinjaQuote, error)
	ConvertQuoteToInvoice(ctx context.Context, quoteID string) (*invoiceninja.NinjaInvoice, error)
}

func newConvertCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "convert <quote-id>",
		Short: "Convert a quote to an invoice",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			quoteID := strings.TrimSpace(args[0])
			if quoteID == "" {
				return fmt.Errorf("quote id is required")
			}

			appCfg := activeConfig()
			if strings.TrimSpace(appCfg.Ninja.APIToken) == "" {
				return fmt.Errorf("Invoice Ninja API token not configured")
			}

			client := invoiceninja.NewClient(appCfg.Ninja)
			return runConvertWorkflow(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), client, quoteID, yes)
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func runConvertWorkflow(ctx context.Context, in io.Reader, out io.Writer, client quoteConverter, quoteID string, yes bool) error {
	quote, err := client.GetQuote(ctx, quoteID)
	if err != nil {
		return fmt.Errorf("failed to fetch quote %s: %w", quoteID, err)
	}

	fmt.Fprintln(out, "Quote preview:")
	fmt.Fprintln(out, invoiceninja.FormatQuoteSummary(quote))

	proceed, err := confirmAction(in, out, "Convert this quote to an invoice?", yes)
	if err != nil {
		return err
	}
	if !proceed {
		fmt.Fprintln(out, "Cancelled")
		return nil
	}

	invoice, err := client.ConvertQuoteToInvoice(ctx, quoteID)
	if err != nil {
		return fmt.Errorf("failed to convert quote %s: %w", quoteID, err)
	}

	fmt.Fprintln(out, "✓ Converted quote to invoice")
	fmt.Fprintln(out, invoiceninja.FormatInvoiceSummary(invoice))

	return nil
}

func confirmAction(in io.Reader, out io.Writer, prompt string, yes bool) (bool, error) {
	if yes {
		return true, nil
	}
	return confirmActionWithReader(bufio.NewReader(in), out, prompt)
}

func confirmActionWithReader(reader *bufio.Reader, out io.Writer, prompt string) (bool, error) {
	fmt.Fprintf(out, "%s [y/N]: ", prompt)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	choice := strings.ToLower(strings.TrimSpace(line))
	return choice == "y" || choice == "yes", nil
}

func runConvertWithLiveClient(quoteID string, yes bool) error {
	appCfg := activeConfig()
	if strings.TrimSpace(appCfg.Ninja.APIToken) == "" {
		return fmt.Errorf("Invoice Ninja API token not configured")
	}

	client := invoiceninja.NewClient(appCfg.Ninja)
	return runConvertWorkflow(context.Background(), os.Stdin, os.Stdout, client, quoteID, yes)
}
