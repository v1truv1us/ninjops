package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/ninjops/ninjops/internal/invoiceninja"
	"github.com/spf13/cobra"
)

const multilineSentinel = ".end"

type editableEntityClient interface {
	GetQuote(ctx context.Context, id string) (*invoiceninja.NinjaQuote, error)
	GetInvoice(ctx context.Context, id string) (*invoiceninja.NinjaInvoice, error)
	UpdateQuoteFields(ctx context.Context, id string, fields map[string]interface{}) (*invoiceninja.NinjaQuote, error)
	UpdateInvoiceFields(ctx context.Context, id string, fields map[string]interface{}) (*invoiceninja.NinjaInvoice, error)
}

type editFieldSelection struct {
	publicNotes bool
	terms       bool
}

func newEditCmd() *cobra.Command {
	var field string
	var yes bool

	cmd := &cobra.Command{
		Use:   "edit <entity> <id>",
		Short: "Edit quote or invoice fields",
		Long:  "Edit public_notes and/or terms for a quote or invoice.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			entity := strings.ToLower(strings.TrimSpace(args[0]))
			id := strings.TrimSpace(args[1])

			if entity != "quote" && entity != "invoice" {
				return fmt.Errorf("unsupported entity %q (supported: quote, invoice)", entity)
			}
			if id == "" {
				return fmt.Errorf("entity id is required")
			}

			selection, err := parseEditFieldSelection(field)
			if err != nil {
				return err
			}

			appCfg := activeConfig()
			if strings.TrimSpace(appCfg.Ninja.APIToken) == "" {
				return fmt.Errorf("Invoice Ninja API token not configured")
			}

			client := invoiceninja.NewClient(appCfg.Ninja)
			return runEditWorkflow(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), client, entity, id, selection, yes)
		},
	}

	cmd.Flags().StringVar(&field, "field", "both", "Fields to edit: public_notes|terms|both")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")

	return cmd
}

func parseEditFieldSelection(raw string) (editFieldSelection, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "public_notes":
		return editFieldSelection{publicNotes: true}, nil
	case "terms":
		return editFieldSelection{terms: true}, nil
	case "both", "":
		return editFieldSelection{publicNotes: true, terms: true}, nil
	default:
		return editFieldSelection{}, fmt.Errorf("unsupported --field %q (supported: public_notes, terms, both)", raw)
	}
}

func runEditWorkflow(ctx context.Context, in io.Reader, out io.Writer, client editableEntityClient, entity, id string, selection editFieldSelection, yes bool) error {
	var currentPublicNotes string
	var currentTerms string

	switch entity {
	case "quote":
		quote, err := client.GetQuote(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to fetch quote %s: %w", id, err)
		}
		currentPublicNotes = quote.PublicNotes
		currentTerms = quote.Terms
	case "invoice":
		invoice, err := client.GetInvoice(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to fetch invoice %s: %w", id, err)
		}
		currentPublicNotes = invoice.PublicNotes
		currentTerms = invoice.Terms
	default:
		return fmt.Errorf("unsupported entity %q", entity)
	}

	fmt.Fprintf(out, "Current %s values:\n", entity)
	printFieldBlock(out, "public_notes", currentPublicNotes)
	printFieldBlock(out, "terms", currentTerms)

	updatedPublicNotes := currentPublicNotes
	updatedTerms := currentTerms

	reader := bufio.NewReader(in)
	var err error
	if selection.publicNotes {
		updatedPublicNotes, err = promptMultilineReplacement(reader, out, "public_notes", currentPublicNotes)
		if err != nil {
			return err
		}
	}
	if selection.terms {
		updatedTerms, err = promptMultilineReplacement(reader, out, "terms", currentTerms)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Updated %s values:\n", entity)
	printFieldBlock(out, "public_notes", updatedPublicNotes)
	printFieldBlock(out, "terms", updatedTerms)

	var proceed bool
	if yes {
		proceed = true
	} else {
		proceed, err = confirmActionWithReader(reader, out, "Apply these updates?")
		if err != nil {
			return err
		}
	}
	if !proceed {
		fmt.Fprintln(out, "Cancelled")
		return nil
	}

	fields := map[string]interface{}{}
	if selection.publicNotes {
		fields["public_notes"] = updatedPublicNotes
	}
	if selection.terms {
		fields["terms"] = updatedTerms
	}

	switch entity {
	case "quote":
		updatedQuote, err := client.UpdateQuoteFields(ctx, id, fields)
		if err != nil {
			return fmt.Errorf("failed to update quote %s: %w", id, err)
		}
		fmt.Fprintln(out, "✓ Updated quote fields")
		fmt.Fprintln(out, invoiceninja.FormatQuoteSummary(updatedQuote))
	case "invoice":
		updatedInvoice, err := client.UpdateInvoiceFields(ctx, id, fields)
		if err != nil {
			return fmt.Errorf("failed to update invoice %s: %w", id, err)
		}
		fmt.Fprintln(out, "✓ Updated invoice fields")
		fmt.Fprintln(out, invoiceninja.FormatInvoiceSummary(updatedInvoice))
	}

	return nil
}

func promptMultilineReplacement(reader *bufio.Reader, out io.Writer, fieldName, current string) (string, error) {
	fmt.Fprintf(out, "Enter new %s (finish with %s on its own line).\n", fieldName, multilineSentinel)
	fmt.Fprintln(out, "Leave empty and type .end to keep current value.")

	lines := make([]string, 0)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read %s: %w", fieldName, err)
		}

		trimmedLine := strings.TrimRight(line, "\r\n")
		if trimmedLine == multilineSentinel {
			break
		}

		if err == io.EOF {
			if strings.TrimSpace(trimmedLine) == "" && len(lines) == 0 {
				return current, nil
			}
			lines = append(lines, trimmedLine)
			break
		}

		lines = append(lines, trimmedLine)
	}

	if len(lines) == 0 {
		return current, nil
	}

	return strings.Join(lines, "\n"), nil
}

func printFieldBlock(out io.Writer, name, value string) {
	fmt.Fprintf(out, "%s:\n", name)
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		fmt.Fprintln(out, "  (empty)")
		return
	}

	for _, line := range strings.Split(value, "\n") {
		fmt.Fprintf(out, "  %s\n", line)
	}
}
