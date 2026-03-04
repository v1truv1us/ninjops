package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ninjops/ninjops/internal/generate"
	"github.com/ninjops/ninjops/internal/invoiceninja"
	"github.com/ninjops/ninjops/internal/spec"
	"github.com/spf13/cobra"
)

func newNewInvoiceCmd() *cobra.Command {
	var fromQuote string
	var input string
	var yes bool
	var clientID string
	var clientEmail string
	var clientName string
	var projectID string
	var taskIDs string
	var nonInteractive bool

	cmd := &cobra.Command{
		Use:   "invoice",
		Short: "Create a new invoice",
		Long:  "Create an invoice by converting a quote, from a QuoteSpec JSON file, or interactively.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateNewInvoiceMode(fromQuote, input, clientID, clientEmail, clientName, projectID, taskIDs); err != nil {
				return err
			}

			appCfg := activeConfig()
			if strings.TrimSpace(appCfg.Ninja.APIToken) == "" {
				return fmt.Errorf("Invoice Ninja API token not configured")
			}

			if strings.TrimSpace(fromQuote) != "" {
				client := invoiceninja.NewClient(appCfg.Ninja)
				return runConvertWorkflow(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), client, strings.TrimSpace(fromQuote), yes)
			}

			if strings.TrimSpace(input) != "" {
				return runNewInvoiceFromInput(cmd, strings.TrimSpace(input), yes)
			}

			return runNewInvoiceInteractive(cmd, clientID, clientEmail, clientName, projectID, taskIDs, nonInteractive, yes)
		},
	}

	cmd.Flags().StringVar(&fromQuote, "from-quote", "", "Convert the given quote ID to an invoice")
	cmd.Flags().StringVar(&input, "input", "", "Create an invoice directly from QuoteSpec JSON")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Invoice Ninja client ID")
	cmd.Flags().StringVar(&clientEmail, "client-email", "", "Invoice Ninja client email")
	cmd.Flags().StringVar(&clientName, "client-name", "", "Invoice Ninja client name")
	cmd.Flags().StringVar(&projectID, "project-id", "", "Invoice Ninja project ID")
	cmd.Flags().StringVar(&taskIDs, "task-ids", "", "Comma-separated Invoice Ninja task IDs")
	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Disable prompts for automation/tests")

	return cmd
}

func validateNewInvoiceMode(fromQuote, input, clientID, clientEmail, clientName, projectID, taskIDs string) error {
	hasFromQuote := strings.TrimSpace(fromQuote) != ""
	hasInput := strings.TrimSpace(input) != ""
	hasEntitySelectors := strings.TrimSpace(clientID) != "" ||
		strings.TrimSpace(clientEmail) != "" ||
		strings.TrimSpace(clientName) != "" ||
		strings.TrimSpace(projectID) != "" ||
		strings.TrimSpace(taskIDs) != ""

	modeCount := 0
	if hasFromQuote {
		modeCount++
	}
	if hasInput {
		modeCount++
	}
	if hasEntitySelectors {
		modeCount++
	}

	if modeCount > 1 {
		return fmt.Errorf("--from-quote, --input, and entity selection flags cannot be used together")
	}

	return nil
}

func runNewInvoiceInteractive(cmd *cobra.Command, clientID, clientEmail, clientName, projectID, taskIDs string, nonInteractive, yes bool) error {
	ctx := cmd.Context()
	in := cmd.InOrStdin()
	out := cmd.OutOrStdout()

	appCfg := activeConfig()
	ninjaClient := invoiceninja.NewClient(appCfg.Ninja)

	quoteSpec := newInvoiceTemplate()

	input := quoteSelectionInput{
		ProjectID:   strings.TrimSpace(projectID),
		TaskIDs:     parseCSVIdentifiers(taskIDs),
		Interactive: !nonInteractive,
		Yes:         yes,
	}

	if strings.TrimSpace(clientID) != "" {
		client, err := resolveClientByID(ctx, ninjaClient, strings.TrimSpace(clientID))
		if err != nil {
			return err
		}
		if client != nil {
			mapNinjaClientToQuote(quoteSpec, client)
		}
	} else if strings.TrimSpace(clientEmail) != "" {
		client, err := ninjaClient.FindClientByEmail(ctx, strings.TrimSpace(clientEmail))
		if err != nil {
			return fmt.Errorf("failed to find client by email: %w", err)
		}
		if client != nil {
			mapNinjaClientToQuote(quoteSpec, client)
		}
	} else if strings.TrimSpace(clientName) != "" {
		client, err := ninjaClient.FindClientByName(ctx, strings.TrimSpace(clientName))
		if err != nil {
			return fmt.Errorf("failed to find client by name: %w", err)
		}
		if client != nil {
			mapNinjaClientToQuote(quoteSpec, client)
		}
	}

	selectedTasks, err := enrichQuoteWithNinjaSelections(ctx, ninjaClient, quoteSpec, input, in, out)
	if err != nil {
		return fmt.Errorf("failed to enrich invoice with selections: %w", err)
	}

	if strings.TrimSpace(quoteSpec.Client.ID) == "" {
		return fmt.Errorf("client is required - use --client-id, --client-email, --client-name, or run interactively")
	}

	taskLineItems := mapTasksToLineItems(selectedTasks)
	if err := mergeTaskLineItems(quoteSpec, taskLineItems, !nonInteractive, yes, in, out); err != nil {
		return fmt.Errorf("failed to merge line items: %w", err)
	}

	generator := generate.NewGenerator()
	artifacts, err := generator.Generate(quoteSpec)
	if err != nil {
		return fmt.Errorf("failed to generate artifacts: %w", err)
	}

	fmt.Fprintf(out, "Invoice creation preview:\n")
	fmt.Fprintf(out, "  Client: %s [id=%s]\n", firstNonEmpty(quoteSpec.Client.Name, "(unnamed client)"), quoteSpec.Client.ID)
	if quoteSpec.Project.ID != "" {
		fmt.Fprintf(out, "  Project: %s [id=%s]\n", firstNonEmpty(quoteSpec.Project.Name, "(unnamed project)"), quoteSpec.Project.ID)
	}
	fmt.Fprintf(out, "  Line items: %d\n", len(quoteSpec.Pricing.LineItems))
	fmt.Fprintf(out, "  Reference: %s\n", quoteSpec.Metadata.Reference)

	proceed, err := confirmAction(in, out, "Create invoice?", yes)
	if err != nil {
		return err
	}
	if !proceed {
		fmt.Fprintln(out, "Cancelled")
		return nil
	}

	invoice, err := ninjaClient.CreateInvoice(ctx, invoiceninja.BuildCreateInvoiceRequest(quoteSpec, quoteSpec.Client.ID, artifacts))
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	fmt.Fprintln(out, "✓ Created invoice")
	fmt.Fprint(out, invoiceninja.FormatInvoiceSummary(invoice))

	return nil
}

func runNewInvoiceFromInput(cmd *cobra.Command, input string, yes bool) error {
	data, err := os.ReadFile(input)
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

	appCfg := activeConfig()
	client := invoiceninja.NewClient(appCfg.Ninja)

	fmt.Fprintf(cmd.OutOrStdout(), "Invoice creation preview:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Source: %s\n", input)
	fmt.Fprintf(cmd.OutOrStdout(), "  Client: %s\n", firstNonEmpty(quoteSpec.Client.Name, "(unnamed client)"))
	fmt.Fprintf(cmd.OutOrStdout(), "  Project: %s\n", firstNonEmpty(quoteSpec.Project.Name, "(unnamed project)"))
	fmt.Fprintf(cmd.OutOrStdout(), "  Line items: %d\n", len(quoteSpec.Pricing.LineItems))
	fmt.Fprintf(cmd.OutOrStdout(), "  Reference: %s\n", quoteSpec.Metadata.Reference)

	proceed, err := confirmAction(cmd.InOrStdin(), cmd.OutOrStdout(), "Create invoice from this spec?", yes)
	if err != nil {
		return err
	}
	if !proceed {
		fmt.Fprintln(cmd.OutOrStdout(), "Cancelled")
		return nil
	}

	ninjaClient, created, err := ensureClientForInvoiceFromSpec(cmd.Context(), client, quoteSpec)
	if err != nil {
		return fmt.Errorf("failed to ensure client: %w", err)
	}

	invoice, err := client.CreateInvoice(cmd.Context(), invoiceninja.BuildCreateInvoiceRequest(quoteSpec, ninjaClient.ID, artifacts))
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	if created {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Created client: %s\n", ninjaClient.ID)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Reused client: %s\n", ninjaClient.ID)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "✓ Created invoice")
	fmt.Fprintln(cmd.OutOrStdout(), invoiceninja.FormatInvoiceSummary(invoice))

	return nil
}

func ensureClientForInvoiceFromSpec(ctx context.Context, client *invoiceninja.Client, quoteSpec *spec.QuoteSpec) (*invoiceninja.NinjaClient, bool, error) {
	req := invoiceninja.CreateClientRequest{
		Name:       quoteSpec.Client.Name,
		Email:      quoteSpec.Client.Email,
		Phone:      quoteSpec.Client.Phone,
		Address1:   quoteSpec.Client.Address.Line1,
		Address2:   quoteSpec.Client.Address.Line2,
		City:       quoteSpec.Client.Address.City,
		State:      quoteSpec.Client.Address.State,
		PostalCode: quoteSpec.Client.Address.PostalCode,
	}

	return client.UpsertClient(ctx, req)
}

func newInvoiceTemplate() *spec.QuoteSpec {
	quoteSpec := spec.NewQuoteSpec()

	quoteSpec.Client = spec.ClientInfo{
		Name:    defaultClientName,
		Email:   defaultClientEmail,
		OrgType: spec.OrgTypeBusiness,
	}

	quoteSpec.Project = spec.ProjectInfo{
		Name:        defaultProjectName,
		Description: defaultProjectDesc,
		Type:        defaultProjectType,
	}

	quoteSpec.Pricing = spec.PricingInfo{
		Currency:  "USD",
		LineItems: []spec.LineItem{},
	}

	return quoteSpec
}

func resolveClientByID(ctx context.Context, client *invoiceninja.Client, clientID string) (*invoiceninja.NinjaClient, error) {
	if client == nil {
		return nil, fmt.Errorf("invoice ninja client is required")
	}

	c, err := client.GetClient(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client %q: %w", clientID, err)
	}
	return c, nil
}
