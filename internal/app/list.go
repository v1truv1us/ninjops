package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/ninjops/ninjops/internal/invoiceninja"
	"github.com/spf13/cobra"
)

const (
	defaultListLimit  = 20
	defaultListFormat = "table"
	listMaxPerPage    = 100
)

type listEntitySource interface {
	ListClients(ctx context.Context, page, perPage int) (*invoiceninja.ClientListResponse, error)
	ListProjects(ctx context.Context, page, perPage int) (*invoiceninja.ProjectListResponse, error)
	FindProjectsByClient(ctx context.Context, clientID string, page, perPage int) (*invoiceninja.ProjectListResponse, error)
	ListTasks(ctx context.Context, page, perPage int) (*invoiceninja.TaskListResponse, error)
	FindTasksByProject(ctx context.Context, projectID string, page, perPage int) (*invoiceninja.TaskListResponse, error)
	FindTasksByClient(ctx context.Context, clientID string, page, perPage int) (*invoiceninja.TaskListResponse, error)
	ListQuotes(ctx context.Context, page, perPage int) (*invoiceninja.QuoteListResponse, error)
	FindQuotesByClient(ctx context.Context, clientID string, page, perPage int) (*invoiceninja.QuoteListResponse, error)
	ListInvoices(ctx context.Context, page, perPage int) (*invoiceninja.InvoiceListResponse, error)
	FindInvoicesByClient(ctx context.Context, clientID string, page, perPage int) (*invoiceninja.InvoiceListResponse, error)
}

func newListCmd() *cobra.Command {
	var clientID string
	var projectID string
	var limit int
	var format string

	cmd := &cobra.Command{
		Use:   "list <entity>",
		Short: "List Invoice Ninja entities",
		Long:  "List clients, projects, tasks, quotes, or invoices from Invoice Ninja.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			entity := strings.ToLower(strings.TrimSpace(args[0]))
			if !isSupportedListEntity(entity) {
				return fmt.Errorf("unsupported entity %q (supported: clients, projects, tasks, quotes, invoices)", entity)
			}

			resolvedFormat := strings.ToLower(strings.TrimSpace(format))
			if !isSupportedListFormat(resolvedFormat) {
				return fmt.Errorf("unsupported --format %q (supported: table, json, simple)", format)
			}

			if limit <= 0 {
				return fmt.Errorf("--limit must be greater than 0")
			}

			if err := validateListFilters(entity, clientID, projectID); err != nil {
				return err
			}

			appCfg := activeConfig()
			if strings.TrimSpace(appCfg.Ninja.APIToken) == "" {
				return fmt.Errorf("Invoice Ninja API token not configured")
			}

			client := invoiceninja.NewClient(appCfg.Ninja)
			return runListEntity(cmd.Context(), cmd.OutOrStdout(), client, entity, strings.TrimSpace(clientID), strings.TrimSpace(projectID), limit, resolvedFormat)
		},
	}

	cmd.Flags().StringVar(&clientID, "client-id", "", "Filter by Invoice Ninja client ID")
	cmd.Flags().StringVar(&projectID, "project-id", "", "Filter by Invoice Ninja project ID")
	cmd.Flags().IntVar(&limit, "limit", defaultListLimit, "Maximum number of records to return")
	cmd.Flags().StringVar(&format, "format", defaultListFormat, "Output format: table|json|simple")

	return cmd
}

func runListEntity(ctx context.Context, w io.Writer, client listEntitySource, entity, clientID, projectID string, limit int, format string) error {
	switch entity {
	case "clients":
		rows, err := listClientsWithLimit(ctx, client, limit)
		if err != nil {
			return err
		}
		return renderClients(w, rows, format)
	case "projects":
		rows, err := listProjectsWithLimit(ctx, client, clientID, limit)
		if err != nil {
			return err
		}
		return renderProjects(w, rows, format)
	case "tasks":
		tasks, err := listTasksForFilters(ctx, client, clientID, projectID, limit)
		if err != nil {
			return err
		}
		return renderTasks(w, tasks, format)
	case "quotes":
		rows, err := listQuotesWithLimit(ctx, client, clientID, limit)
		if err != nil {
			return err
		}
		return renderQuotes(w, rows, format)
	case "invoices":
		rows, err := listInvoicesWithLimit(ctx, client, clientID, limit)
		if err != nil {
			return err
		}
		return renderInvoices(w, rows, format)
	default:
		return fmt.Errorf("unsupported entity %q", entity)
	}
}

func listClientsWithLimit(ctx context.Context, source listEntitySource, limit int) ([]invoiceninja.NinjaClient, error) {
	return collectPages(limit, func(page, perPage int) ([]invoiceninja.NinjaClient, invoiceninja.APIPagination, error) {
		resp, err := source.ListClients(ctx, page, perPage)
		if err != nil {
			return nil, invoiceninja.APIPagination{}, err
		}
		return resp.Data, resp.Meta.Pagination, nil
	})
}

func listProjectsWithLimit(ctx context.Context, source listEntitySource, clientID string, limit int) ([]invoiceninja.NinjaProject, error) {
	if strings.TrimSpace(clientID) != "" {
		return collectPages(limit, func(page, perPage int) ([]invoiceninja.NinjaProject, invoiceninja.APIPagination, error) {
			resp, err := source.FindProjectsByClient(ctx, clientID, page, perPage)
			if err != nil {
				return nil, invoiceninja.APIPagination{}, err
			}
			return resp.Data, resp.Meta.Pagination, nil
		})
	}

	return collectPages(limit, func(page, perPage int) ([]invoiceninja.NinjaProject, invoiceninja.APIPagination, error) {
		resp, err := source.ListProjects(ctx, page, perPage)
		if err != nil {
			return nil, invoiceninja.APIPagination{}, err
		}
		return resp.Data, resp.Meta.Pagination, nil
	})
}

func listQuotesWithLimit(ctx context.Context, source listEntitySource, clientID string, limit int) ([]invoiceninja.NinjaQuote, error) {
	if strings.TrimSpace(clientID) != "" {
		return collectPages(limit, func(page, perPage int) ([]invoiceninja.NinjaQuote, invoiceninja.APIPagination, error) {
			resp, err := source.FindQuotesByClient(ctx, clientID, page, perPage)
			if err != nil {
				return nil, invoiceninja.APIPagination{}, err
			}
			return resp.Data, resp.Meta.Pagination, nil
		})
	}

	return collectPages(limit, func(page, perPage int) ([]invoiceninja.NinjaQuote, invoiceninja.APIPagination, error) {
		resp, err := source.ListQuotes(ctx, page, perPage)
		if err != nil {
			return nil, invoiceninja.APIPagination{}, err
		}
		return resp.Data, resp.Meta.Pagination, nil
	})
}

func listInvoicesWithLimit(ctx context.Context, source listEntitySource, clientID string, limit int) ([]invoiceninja.NinjaInvoice, error) {
	if strings.TrimSpace(clientID) != "" {
		return collectPages(limit, func(page, perPage int) ([]invoiceninja.NinjaInvoice, invoiceninja.APIPagination, error) {
			resp, err := source.FindInvoicesByClient(ctx, clientID, page, perPage)
			if err != nil {
				return nil, invoiceninja.APIPagination{}, err
			}
			return resp.Data, resp.Meta.Pagination, nil
		})
	}

	return collectPages(limit, func(page, perPage int) ([]invoiceninja.NinjaInvoice, invoiceninja.APIPagination, error) {
		resp, err := source.ListInvoices(ctx, page, perPage)
		if err != nil {
			return nil, invoiceninja.APIPagination{}, err
		}
		return resp.Data, resp.Meta.Pagination, nil
	})
}

func listTasksForFilters(ctx context.Context, source listEntitySource, clientID, projectID string, limit int) ([]invoiceninja.NinjaTask, error) {
	if limit <= 0 {
		return nil, nil
	}

	perPage := int(math.Min(float64(limit), listMaxPerPage))
	if perPage <= 0 {
		perPage = defaultListLimit
	}

	collected := make([]invoiceninja.NinjaTask, 0, limit)
	page := 1
	for len(collected) < limit {
		var (
			resp *invoiceninja.TaskListResponse
			err  error
		)

		switch {
		case projectID != "":
			resp, err = source.FindTasksByProject(ctx, projectID, page, perPage)
		case clientID != "":
			resp, err = source.FindTasksByClient(ctx, clientID, page, perPage)
		default:
			resp, err = source.ListTasks(ctx, page, perPage)
		}
		if err != nil {
			return nil, err
		}

		pageRows := resp.Data
		if projectID != "" && clientID != "" {
			pageRows = filterTasksForList(pageRows, clientID, projectID)
		}

		remaining := limit - len(collected)
		collected = append(collected, limitRows(pageRows, remaining)...)

		if shouldStopPagination(resp.Meta.Pagination, len(resp.Data), perPage, page) {
			break
		}
		page++
	}

	return collected, nil
}

func collectPages[T any](limit int, fetch func(page, perPage int) ([]T, invoiceninja.APIPagination, error)) ([]T, error) {
	if limit <= 0 {
		return nil, nil
	}

	perPage := int(math.Min(float64(limit), listMaxPerPage))
	if perPage <= 0 {
		perPage = defaultListLimit
	}

	rows := make([]T, 0, limit)
	page := 1
	for len(rows) < limit {
		pageRows, pagination, err := fetch(page, perPage)
		if err != nil {
			return nil, err
		}

		remaining := limit - len(rows)
		rows = append(rows, limitRows(pageRows, remaining)...)

		if shouldStopPagination(pagination, len(pageRows), perPage, page) {
			break
		}
		page++
	}

	return rows, nil
}

func limitRows[T any](rows []T, limit int) []T {
	if limit <= 0 || len(rows) <= limit {
		return rows
	}
	return rows[:limit]
}

func shouldStopPagination(pagination invoiceninja.APIPagination, fetchedRows int, perPage, page int) bool {
	if pagination.TotalPages > 0 && page >= pagination.TotalPages {
		return true
	}
	if pagination.TotalPages > 0 {
		return false
	}
	if pagination.Total > 0 && perPage > 0 && page*perPage >= pagination.Total {
		return true
	}
	if fetchedRows == 0 {
		return true
	}
	if fetchedRows < perPage {
		return true
	}
	return false
}

func isSupportedListEntity(entity string) bool {
	switch entity {
	case "clients", "projects", "tasks", "quotes", "invoices":
		return true
	default:
		return false
	}
}

func isSupportedListFormat(format string) bool {
	switch format {
	case "table", "json", "simple":
		return true
	default:
		return false
	}
}

func validateListFilters(entity, clientID, projectID string) error {
	hasClient := strings.TrimSpace(clientID) != ""
	hasProject := strings.TrimSpace(projectID) != ""

	switch entity {
	case "clients":
		if hasClient || hasProject {
			return fmt.Errorf("--client-id/--project-id are not supported for list clients")
		}
	case "projects":
		if hasProject {
			return fmt.Errorf("--project-id is not supported for list projects")
		}
	case "tasks":
		return nil
	case "quotes", "invoices":
		if hasProject {
			return fmt.Errorf("--project-id is not supported for list %s", entity)
		}
	}

	return nil
}

func renderClients(w io.Writer, clients []invoiceninja.NinjaClient, format string) error {
	if format == "json" {
		return writeJSON(w, clients)
	}
	if len(clients) == 0 {
		_, err := fmt.Fprintln(w, "No clients found.")
		return err
	}

	if format == "simple" {
		for _, c := range clients {
			if _, err := fmt.Fprintf(w, "%s\t%s\n", c.ID, listClientLabel(c)); err != nil {
				return err
			}
		}
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tNAME\tEMAIL"); err != nil {
		return err
	}
	for _, c := range clients {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", c.ID, listClientName(c), firstNonEmpty(primaryContactEmail(c.Contacts), c.Email, "-")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func renderProjects(w io.Writer, projects []invoiceninja.NinjaProject, format string) error {
	if format == "json" {
		return writeJSON(w, projects)
	}
	if len(projects) == 0 {
		_, err := fmt.Fprintln(w, "No projects found.")
		return err
	}

	if format == "simple" {
		for _, p := range projects {
			if _, err := fmt.Fprintf(w, "%s\t%s\n", p.ID, firstNonEmpty(strings.TrimSpace(p.Name), "(unnamed project)")); err != nil {
				return err
			}
		}
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tNAME\tCLIENT ID"); err != nil {
		return err
	}
	for _, p := range projects {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\n", p.ID, firstNonEmpty(strings.TrimSpace(p.Name), "(unnamed project)"), p.ClientID); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func renderTasks(w io.Writer, tasks []invoiceninja.NinjaTask, format string) error {
	if format == "json" {
		return writeJSON(w, tasks)
	}
	if len(tasks) == 0 {
		_, err := fmt.Fprintln(w, "No tasks found.")
		return err
	}

	if format == "simple" {
		for _, task := range tasks {
			if _, err := fmt.Fprintf(w, "%s\t%s\n", task.ID, listTaskLabel(task)); err != nil {
				return err
			}
		}
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tDESCRIPTION\tPROJECT ID\tCLIENT ID"); err != nil {
		return err
	}
	for _, task := range tasks {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", task.ID, listTaskLabel(task), task.ProjectID, task.ClientID); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func renderQuotes(w io.Writer, quotes []invoiceninja.NinjaQuote, format string) error {
	if format == "json" {
		return writeJSON(w, quotes)
	}
	if len(quotes) == 0 {
		_, err := fmt.Fprintln(w, "No quotes found.")
		return err
	}

	if format == "simple" {
		for _, quote := range quotes {
			if _, err := fmt.Fprintf(w, "%s\t#%s\t%s\n", quote.ID, quote.Number, listAmount(quote.Amount)); err != nil {
				return err
			}
		}
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tNUMBER\tCLIENT ID\tAMOUNT\tSTATUS"); err != nil {
		return err
	}
	for _, quote := range quotes {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", quote.ID, quote.Number, quote.ClientID, listAmount(quote.Amount), quote.StatusID); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func renderInvoices(w io.Writer, invoices []invoiceninja.NinjaInvoice, format string) error {
	if format == "json" {
		return writeJSON(w, invoices)
	}
	if len(invoices) == 0 {
		_, err := fmt.Fprintln(w, "No invoices found.")
		return err
	}

	if format == "simple" {
		for _, invoice := range invoices {
			if _, err := fmt.Fprintf(w, "%s\t#%s\t%s\n", invoice.ID, invoice.Number, listAmount(invoice.Amount)); err != nil {
				return err
			}
		}
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tNUMBER\tCLIENT ID\tAMOUNT\tSTATUS"); err != nil {
		return err
	}
	for _, invoice := range invoices {
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", invoice.ID, invoice.Number, invoice.ClientID, listAmount(invoice.Amount), invoice.StatusID); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func filterProjectsForList(projects []invoiceninja.NinjaProject, clientID string) []invoiceninja.NinjaProject {
	if strings.TrimSpace(clientID) == "" {
		return projects
	}

	filtered := make([]invoiceninja.NinjaProject, 0, len(projects))
	for _, project := range projects {
		if strings.TrimSpace(project.ClientID) == strings.TrimSpace(clientID) {
			filtered = append(filtered, project)
		}
	}
	return filtered
}

func filterTasksForList(tasks []invoiceninja.NinjaTask, clientID, projectID string) []invoiceninja.NinjaTask {
	hasClient := strings.TrimSpace(clientID) != ""
	hasProject := strings.TrimSpace(projectID) != ""
	if !hasClient && !hasProject {
		return tasks
	}

	filtered := make([]invoiceninja.NinjaTask, 0, len(tasks))
	for _, task := range tasks {
		if hasClient && strings.TrimSpace(task.ClientID) != strings.TrimSpace(clientID) {
			continue
		}
		if hasProject && strings.TrimSpace(task.ProjectID) != strings.TrimSpace(projectID) {
			continue
		}
		filtered = append(filtered, task)
	}
	return filtered
}

func filterQuotesForList(quotes []invoiceninja.NinjaQuote, clientID string) []invoiceninja.NinjaQuote {
	if strings.TrimSpace(clientID) == "" {
		return quotes
	}

	filtered := make([]invoiceninja.NinjaQuote, 0, len(quotes))
	for _, quote := range quotes {
		if strings.TrimSpace(quote.ClientID) == strings.TrimSpace(clientID) {
			filtered = append(filtered, quote)
		}
	}
	return filtered
}

func filterInvoicesForList(invoices []invoiceninja.NinjaInvoice, clientID string) []invoiceninja.NinjaInvoice {
	if strings.TrimSpace(clientID) == "" {
		return invoices
	}

	filtered := make([]invoiceninja.NinjaInvoice, 0, len(invoices))
	for _, invoice := range invoices {
		if strings.TrimSpace(invoice.ClientID) == strings.TrimSpace(clientID) {
			filtered = append(filtered, invoice)
		}
	}
	return filtered
}

func listClientName(client invoiceninja.NinjaClient) string {
	return firstNonEmpty(strings.TrimSpace(client.DisplayName), strings.TrimSpace(client.Name), "(unnamed client)")
}

func listClientLabel(client invoiceninja.NinjaClient) string {
	return firstNonEmpty(listClientName(client), firstNonEmpty(primaryContactEmail(client.Contacts), client.Email))
}

func listTaskLabel(task invoiceninja.NinjaTask) string {
	if strings.TrimSpace(task.Description) != "" {
		return strings.TrimSpace(task.Description)
	}
	if strings.TrimSpace(task.Number) != "" {
		return "Task " + strings.TrimSpace(task.Number)
	}
	return "(unnamed task)"
}

func listAmount(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func writeJSON(w io.Writer, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, string(data)); err != nil {
		return err
	}
	return nil
}
