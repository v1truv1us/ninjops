package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/ninjops/ninjops/internal/invoiceninja"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <entity> <id>",
		Short: "Show Invoice Ninja entity details",
		Long:  "Show one client, project, task, quote, or invoice as JSON.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			entity := strings.ToLower(strings.TrimSpace(args[0]))
			id := strings.TrimSpace(args[1])

			if !isSupportedShowEntity(entity) {
				return fmt.Errorf("unsupported entity %q (supported: client, project, task, quote, invoice)", entity)
			}
			if id == "" {
				return fmt.Errorf("entity id is required")
			}

			appCfg := activeConfig()
			if strings.TrimSpace(appCfg.Ninja.APIToken) == "" {
				return fmt.Errorf("Invoice Ninja API token not configured")
			}

			client := invoiceninja.NewClient(appCfg.Ninja)
			entityData, err := fetchShowEntity(cmd.Context(), client, entity, id)
			if err != nil {
				return err
			}

			return writeJSON(cmd.OutOrStdout(), entityData)
		},
	}

	return cmd
}

func isSupportedShowEntity(entity string) bool {
	switch entity {
	case "client", "project", "task", "quote", "invoice":
		return true
	default:
		return false
	}
}

func fetchShowEntity(ctx context.Context, client *invoiceninja.Client, entity, id string) (interface{}, error) {
	switch entity {
	case "client":
		return client.GetClient(ctx, id)
	case "project":
		return client.GetProject(ctx, id)
	case "task":
		return client.GetTask(ctx, id)
	case "quote":
		return client.GetQuote(ctx, id)
	case "invoice":
		return client.GetInvoice(ctx, id)
	default:
		return nil, fmt.Errorf("unsupported entity %q", entity)
	}
}
