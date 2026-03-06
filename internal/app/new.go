package app

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ninjops/ninjops/internal/generate"
	"github.com/ninjops/ninjops/internal/invoiceninja"
	"github.com/ninjops/ninjops/internal/spec"
	"github.com/spf13/cobra"
)

const (
	defaultClientName        = "Client Name"
	defaultClientEmail       = "client@example.com"
	defaultProjectName       = "Project Name"
	defaultProjectDesc       = "Project description"
	defaultProjectType       = "web"
	defaultTermsFilename     = "terms.md"
	defaultPublicNotesFile   = "public_notes.txt"
	interactiveSelectMaxRows = 50
	previewSnippetLimit      = 320
)

type clientSource interface {
	ListClients(ctx context.Context, page, perPage int) (*invoiceninja.ClientListResponse, error)
	ListProjects(ctx context.Context, page, perPage int) (*invoiceninja.ProjectListResponse, error)
	GetClient(ctx context.Context, id string) (*invoiceninja.NinjaClient, error)
	FindClientByEmail(ctx context.Context, email string) (*invoiceninja.NinjaClient, error)
	FindClientByName(ctx context.Context, name string) (*invoiceninja.NinjaClient, error)
}

type quoteWorkflowSource interface {
	clientSource
	ListTasks(ctx context.Context, page, perPage int) (*invoiceninja.TaskListResponse, error)
	GetProject(ctx context.Context, id string) (*invoiceninja.NinjaProject, error)
	GetTask(ctx context.Context, id string) (*invoiceninja.NinjaTask, error)
	CreateClient(ctx context.Context, req invoiceninja.CreateClientRequest) (*invoiceninja.NinjaClient, error)
	CreateProject(ctx context.Context, req invoiceninja.CreateProjectRequest) (*invoiceninja.NinjaProject, error)
	CreateTask(ctx context.Context, req invoiceninja.CreateTaskRequest) (*invoiceninja.NinjaTask, error)
	CreateQuote(ctx context.Context, req invoiceninja.CreateQuoteRequest) (*invoiceninja.NinjaQuote, error)
	ConvertQuoteToInvoice(ctx context.Context, quoteID string) (*invoiceninja.NinjaInvoice, error)
}

type quoteCreateDecision struct {
	CreateQuote      bool
	ConvertToInvoice bool
}

func newNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create new resources",
	}

	cmd.AddCommand(newNewQuoteCmd())
	cmd.AddCommand(newNewInvoiceCmd())

	return cmd
}

func newNewQuoteCmd() *cobra.Command {
	var output string
	var artifactsDir string
	var clientID string
	var clientEmail string
	var clientName string
	var projectID string
	var taskIDs string
	var nonInteractive bool
	var yes bool
	var createQuote bool
	var convertToInvoice bool
	var projectName string
	var projectDescription string
	var projectType string
	var projectTimeline string

	cmd := &cobra.Command{
		Use:   "quote",
		Short: "Create a new QuoteSpec JSON template",
		Long:  `Generates a new QuoteSpec JSON with a fresh reference ID.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			in := cmd.InOrStdin()
			out := cmd.OutOrStdout()
			bufferedIn := asBufferedReader(in)

			if err := validateClientSelectors(clientID, clientEmail, clientName); err != nil {
				return err
			}

			parsedTaskIDs := parseCSVIdentifiers(taskIDs)
			decision, err := validateCreateConvertDecision(strings.TrimSpace(activeConfig().Ninja.APIToken) != "", createQuote, convertToInvoice)
			if err != nil {
				return err
			}

			quoteSpec := newQuoteTemplate()
			quoteSpec.Project.ID = strings.TrimSpace(projectID)
			quoteSpec.TaskIDs = parsedTaskIDs

			applyProjectFlags(quoteSpec, projectName, projectDescription, projectType, projectTimeline)

			interactive := isInteractiveSession(nonInteractive, in)
			if err := hydrateClientInfo(cmd, quoteSpec, clientID, clientEmail, clientName, interactive, bufferedIn, out); err != nil {
				return err
			}

			var selectedTasks []invoiceninja.NinjaTask
			hasToken := strings.TrimSpace(activeConfig().Ninja.APIToken) != ""
			if hasToken {
				source := invoiceninja.NewClient(activeConfig().Ninja)
				selectedTasks, err = enrichQuoteWithNinjaSelections(cmd.Context(), source, quoteSpec, quoteSelectionInput{
					ProjectID:          strings.TrimSpace(projectID),
					TaskIDs:            parsedTaskIDs,
					Interactive:        interactive,
					Yes:                yes,
					CreateQuoteFlag:    decision.CreateQuote,
					ConvertInvoiceFlag: decision.ConvertToInvoice,
				}, bufferedIn, out)
				if err != nil {
					return err
				}
			}

			applyProjectFlags(quoteSpec, projectName, projectDescription, projectType, projectTimeline)

			if interactive {
				if err := promptProjectFields(quoteSpec, bufferedIn, out); err != nil {
					return err
				}
			}

			appendTaskLineItems := mapTasksToLineItems(selectedTasks)
			if err := mergeTaskLineItems(quoteSpec, appendTaskLineItems, interactive, yes, bufferedIn, out); err != nil {
				return err
			}

			generator := generate.NewGenerator()
			artifacts, err := generator.Generate(quoteSpec)
			if err != nil {
				return fmt.Errorf("failed to generate quote artifacts: %w", err)
			}

			if interactive {
				if err := editQuoteArtifactsInline(bufferedIn, out, artifacts); err != nil {
					return err
				}
				printQuotePreview(out, quoteSpec, selectedTasks, artifacts)
			}

			if interactive {
				decision, err = resolveCreateConvertDecision(decision, hasToken, yes, bufferedIn, out)
				if err != nil {
					return err
				}
			}

			jsonData, err := quoteSpec.ToJSON()
			if err != nil {
				return fmt.Errorf("failed to marshal QuoteSpec: %w", err)
			}

			if output != "" {
				if err := writeFile(output, jsonData); err != nil {
					return err
				}

				if err := writeNewQuoteArtifacts(output, artifactsDir, quoteSpec, artifacts); err != nil {
					return err
				}

				fmt.Fprintf(out, "✓ Created %s\n", output)
				fmt.Fprintf(out, "  Reference: %s\n", quoteSpec.Metadata.Reference)
			} else {
				fmt.Fprintln(out, string(jsonData))
			}

			if decision.CreateQuote {
				if strings.TrimSpace(quoteSpec.Client.ID) == "" {
					return fmt.Errorf("cannot create quote: client ID is required (select a client or pass --client-id)")
				}

				ninjaClient := invoiceninja.NewClient(activeConfig().Ninja)
				createdQuote, createErr := ninjaClient.CreateQuote(cmd.Context(), invoiceninja.BuildCreateQuoteRequest(quoteSpec, quoteSpec.Client.ID, artifacts))
				if createErr != nil {
					return fmt.Errorf("failed to create quote in Invoice Ninja: %w", createErr)
				}

				fmt.Fprintf(out, "✓ Created quote in Invoice Ninja: id=%s number=%s\n", createdQuote.ID, createdQuote.Number)

				if decision.ConvertToInvoice {
					invoice, convertErr := ninjaClient.ConvertQuoteToInvoice(cmd.Context(), createdQuote.ID)
					if convertErr != nil {
						return fmt.Errorf("failed to convert quote to invoice: %w", convertErr)
					}
					fmt.Fprintf(out, "✓ Converted to invoice: id=%s number=%s\n", invoice.ID, invoice.Number)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	cmd.Flags().StringVar(&artifactsDir, "artifacts-dir", "", "Directory to write generated terms/public notes when --output is set")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Invoice Ninja client ID to hydrate client fields")
	cmd.Flags().StringVar(&clientEmail, "client-email", "", "Invoice Ninja client email to hydrate client fields")
	cmd.Flags().StringVar(&clientName, "client-name", "", "Invoice Ninja client name to hydrate client fields")
	cmd.Flags().StringVar(&projectID, "project-id", "", "Invoice Ninja project ID to link to this quote")
	cmd.Flags().StringVar(&taskIDs, "task-ids", "", "Comma-separated Invoice Ninja task IDs to link/use as line items")
	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Disable prompts for automation/tests")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompts")
	cmd.Flags().BoolVar(&createQuote, "create-quote", false, "Create quote in Invoice Ninja after preparing quote")
	cmd.Flags().BoolVar(&convertToInvoice, "convert-to-invoice", false, "Convert created quote to invoice")
	cmd.Flags().StringVar(&projectName, "project-name", "", "Project name")
	cmd.Flags().StringVar(&projectDescription, "project-description", "", "Project description")
	cmd.Flags().StringVar(&projectType, "project-type", "", "Project type")
	cmd.Flags().StringVar(&projectTimeline, "project-timeline", "", "Project timeline")

	return cmd
}

func newQuoteTemplate() *spec.QuoteSpec {
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

	quoteSpec.Work = spec.WorkDefinition{
		Features: []spec.Feature{
			{
				Name:        "Feature 1",
				Description: "Description of feature 1",
				Priority:    "high",
			},
		},
	}

	quoteSpec.Pricing = spec.PricingInfo{
		Currency:  "USD",
		LineItems: []spec.LineItem{},
	}

	quoteSpec.Settings = spec.QuoteSettings{
		Tone:           spec.ToneProfessional,
		IncludePricing: true,
	}

	return quoteSpec
}

func applyProjectFlags(quoteSpec *spec.QuoteSpec, projectName, projectDescription, projectType, projectTimeline string) {
	if strings.TrimSpace(projectName) != "" {
		quoteSpec.Project.Name = strings.TrimSpace(projectName)
	}
	if strings.TrimSpace(projectDescription) != "" {
		quoteSpec.Project.Description = strings.TrimSpace(projectDescription)
	}
	if strings.TrimSpace(projectType) != "" {
		quoteSpec.Project.Type = strings.TrimSpace(projectType)
	}
	if strings.TrimSpace(projectTimeline) != "" {
		quoteSpec.Project.Timeline = strings.TrimSpace(projectTimeline)
	}
}

func validateClientSelectors(clientID, clientEmail, clientName string) error {
	selectors := 0
	if strings.TrimSpace(clientID) != "" {
		selectors++
	}
	if strings.TrimSpace(clientEmail) != "" {
		selectors++
	}
	if strings.TrimSpace(clientName) != "" {
		selectors++
	}

	if selectors > 1 {
		return fmt.Errorf("only one client selector may be provided (--client-id, --client-email, or --client-name)")
	}

	return nil
}

func hydrateClientInfo(cmd *cobra.Command, quoteSpec *spec.QuoteSpec, clientID, clientEmail, clientName string, interactive bool, reader io.Reader, writer io.Writer) error {
	clientCfg := activeConfig().Ninja
	hasToken := strings.TrimSpace(clientCfg.APIToken) != ""

	hasSelector := strings.TrimSpace(clientID) != "" || strings.TrimSpace(clientEmail) != "" || strings.TrimSpace(clientName) != ""
	if !hasSelector {
		return nil
	}

	if !hasToken {
		if hasSelector {
			return fmt.Errorf("client selector requires Invoice Ninja API token configuration")
		}
		return nil
	}

	source := invoiceninja.NewClient(clientCfg)

	selected, err := resolveClientSelection(cmd.Context(), source, clientID, clientEmail, clientName)
	if err != nil {
		return err
	}
	mapNinjaClientToQuote(quoteSpec, selected)

	return nil
}

func maybeSelectClientProject(ctx context.Context, source clientSource, reader io.Reader, writer io.Writer, selectedClient *invoiceninja.NinjaClient) (*invoiceninja.NinjaProject, error) {
	if selectedClient == nil || strings.TrimSpace(selectedClient.ID) == "" {
		return nil, nil
	}

	projectList, err := source.ListProjects(ctx, 1, interactiveSelectMaxRows)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	projects := filterProjectsByClientID(projectList.Data, selectedClient.ID)
	if len(projects) == 0 {
		return nil, nil
	}

	return promptProjectSelection(reader, writer, projects)
}

func filterProjectsByClientID(projects []invoiceninja.NinjaProject, clientID string) []invoiceninja.NinjaProject {
	trimmedClientID := strings.TrimSpace(clientID)
	if trimmedClientID == "" {
		return nil
	}

	filtered := make([]invoiceninja.NinjaProject, 0, len(projects))
	for _, project := range projects {
		if strings.TrimSpace(project.ClientID) == trimmedClientID {
			filtered = append(filtered, project)
		}
	}

	return filtered
}

func resolveClientSelection(ctx context.Context, source clientSource, clientID, clientEmail, clientName string) (*invoiceninja.NinjaClient, error) {
	if strings.TrimSpace(clientID) != "" {
		client, err := source.GetClient(ctx, strings.TrimSpace(clientID))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve --client-id %q: %w", strings.TrimSpace(clientID), err)
		}
		if client == nil || strings.TrimSpace(client.ID) == "" {
			return nil, fmt.Errorf("no Invoice Ninja client found for --client-id %q", strings.TrimSpace(clientID))
		}
		return client, nil
	}

	if strings.TrimSpace(clientEmail) != "" {
		client, err := source.FindClientByEmail(ctx, strings.TrimSpace(clientEmail))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve --client-email %q: %w", strings.TrimSpace(clientEmail), err)
		}
		if client == nil {
			return nil, fmt.Errorf("no Invoice Ninja client found for --client-email %q", strings.TrimSpace(clientEmail))
		}
		return client, nil
	}

	if strings.TrimSpace(clientName) != "" {
		client, err := source.FindClientByName(ctx, strings.TrimSpace(clientName))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve --client-name %q: %w", strings.TrimSpace(clientName), err)
		}
		if client == nil {
			return nil, fmt.Errorf("no Invoice Ninja client found for --client-name %q", strings.TrimSpace(clientName))
		}
		return client, nil
	}

	return nil, nil
}

func mapNinjaClientToQuote(quoteSpec *spec.QuoteSpec, ninjaClient *invoiceninja.NinjaClient) {
	if ninjaClient == nil {
		return
	}

	quoteSpec.Client.ID = strings.TrimSpace(ninjaClient.ID)

	quoteSpec.Client.Name = firstNonEmpty(ninjaClient.DisplayName, ninjaClient.Name, quoteSpec.Client.Name)
	quoteSpec.Client.Email = firstNonEmpty(primaryContactEmail(ninjaClient.Contacts), ninjaClient.Email, quoteSpec.Client.Email)
	quoteSpec.Client.Phone = firstNonEmpty(primaryContactPhone(ninjaClient.Contacts), ninjaClient.Phone)
	quoteSpec.Client.Address = spec.Address{
		Line1:      strings.TrimSpace(ninjaClient.Address1),
		Line2:      strings.TrimSpace(ninjaClient.Address2),
		City:       strings.TrimSpace(ninjaClient.City),
		State:      strings.TrimSpace(ninjaClient.State),
		PostalCode: strings.TrimSpace(ninjaClient.PostalCode),
	}
}

func mapNinjaProjectToQuote(quoteSpec *spec.QuoteSpec, ninjaProject *invoiceninja.NinjaProject) {
	if ninjaProject == nil {
		return
	}

	quoteSpec.Project.ID = strings.TrimSpace(ninjaProject.ID)

	quoteSpec.Project.Name = firstNonEmpty(ninjaProject.Name, quoteSpec.Project.Name)

	description := firstNonEmpty(ninjaProject.Description, ninjaProject.PublicNotes, ninjaProject.PrivateNotes)
	if description != "" {
		quoteSpec.Project.Description = description
	}

	deadline := strings.TrimSpace(ninjaProject.DueDate)
	if deadline != "" {
		quoteSpec.Project.Deadline = deadline
	}
}

func primaryContactEmail(contacts []invoiceninja.ClientContact) string {
	if len(contacts) == 0 {
		return ""
	}

	for _, contact := range contacts {
		if contact.IsPrimary && strings.TrimSpace(contact.Email) != "" {
			return strings.TrimSpace(contact.Email)
		}
	}

	for _, contact := range contacts {
		if strings.TrimSpace(contact.Email) != "" {
			return strings.TrimSpace(contact.Email)
		}
	}

	return ""
}

func primaryContactPhone(contacts []invoiceninja.ClientContact) string {
	if len(contacts) == 0 {
		return ""
	}

	for _, contact := range contacts {
		if contact.IsPrimary && strings.TrimSpace(contact.Phone) != "" {
			return strings.TrimSpace(contact.Phone)
		}
	}

	for _, contact := range contacts {
		if strings.TrimSpace(contact.Phone) != "" {
			return strings.TrimSpace(contact.Phone)
		}
	}

	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func isInteractiveSession(nonInteractive bool, reader io.Reader) bool {
	if nonInteractive {
		return false
	}

	file, ok := reader.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

func promptProjectFields(quoteSpec *spec.QuoteSpec, reader io.Reader, writer io.Writer) error {
	buffered := asBufferedReader(reader)

	if isMissingProjectField(quoteSpec.Project.Name, defaultProjectName) {
		value, err := promptField(buffered, writer, "Project name", quoteSpec.Project.Name, true)
		if err != nil {
			return err
		}
		quoteSpec.Project.Name = value
	}

	if isMissingProjectField(quoteSpec.Project.Description, defaultProjectDesc) {
		value, err := promptField(buffered, writer, "Project description", quoteSpec.Project.Description, true)
		if err != nil {
			return err
		}
		quoteSpec.Project.Description = value
	}

	if isMissingProjectField(quoteSpec.Project.Type, defaultProjectType) {
		value, err := promptField(buffered, writer, "Project type", quoteSpec.Project.Type, true)
		if err != nil {
			return err
		}
		quoteSpec.Project.Type = value
	}

	if strings.TrimSpace(quoteSpec.Project.Timeline) == "" {
		value, err := promptField(buffered, writer, "Project timeline (optional)", "", false)
		if err != nil {
			return err
		}
		quoteSpec.Project.Timeline = value
	}

	return nil
}

func isMissingProjectField(value, placeholder string) bool {
	trimmed := strings.TrimSpace(value)
	return trimmed == "" || trimmed == placeholder
}

func promptField(reader *bufio.Reader, writer io.Writer, label string, current string, required bool) (string, error) {
	for {
		if strings.TrimSpace(current) != "" {
			fmt.Fprintf(writer, "%s [%s]: ", label, current)
		} else {
			fmt.Fprintf(writer, "%s: ", label)
		}

		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read %s: %w", strings.ToLower(label), err)
		}

		input := strings.TrimSpace(line)
		if input != "" {
			return input, nil
		}

		if err == io.EOF {
			if strings.TrimSpace(current) != "" {
				return current, nil
			}
			if required {
				return "", fmt.Errorf("%s is required", strings.ToLower(label))
			}
			return "", nil
		}

		if strings.TrimSpace(current) != "" {
			return current, nil
		}

		if !required {
			return "", nil
		}

		fmt.Fprintf(writer, "%s is required.\n", label)
	}
}

func promptClientSelection(reader io.Reader, writer io.Writer, clients []invoiceninja.NinjaClient) (*invoiceninja.NinjaClient, error) {
	if len(clients) == 0 {
		return nil, nil
	}

	fmt.Fprintln(writer, "Select a client from Invoice Ninja:")
	fmt.Fprintln(writer, "  0) Keep template defaults")
	for i, client := range clients {
		displayName := firstNonEmpty(client.DisplayName, client.Name, "(unnamed client)")
		email := firstNonEmpty(primaryContactEmail(client.Contacts), client.Email, "no-email")
		fmt.Fprintf(writer, "  %d) %s <%s> [id=%s]\n", i+1, displayName, email, client.ID)
	}

	buffered := asBufferedReader(reader)
	for {
		fmt.Fprintf(writer, "Enter selection (0-%d): ", len(clients))
		line, err := buffered.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read client selection: %w", err)
		}

		choiceText := strings.TrimSpace(line)
		if choiceText == "" {
			if err == io.EOF {
				return nil, nil
			}
			fmt.Fprintln(writer, "Selection is required.")
			continue
		}

		choice, convErr := strconv.Atoi(choiceText)
		if convErr != nil {
			fmt.Fprintln(writer, "Please enter a valid number.")
			continue
		}

		if choice == 0 {
			return nil, nil
		}
		if choice < 0 || choice > len(clients) {
			fmt.Fprintln(writer, "Selection out of range.")
			continue
		}

		selected := clients[choice-1]
		return &selected, nil
	}
}

func promptProjectSelection(reader io.Reader, writer io.Writer, projects []invoiceninja.NinjaProject) (*invoiceninja.NinjaProject, error) {
	if len(projects) == 0 {
		return nil, nil
	}

	fmt.Fprintln(writer, "Select a project from Invoice Ninja:")
	fmt.Fprintln(writer, "  0) Keep current/manual project details")
	for i, project := range projects {
		fmt.Fprintf(writer, "  %d) %s [id=%s]\n", i+1, firstNonEmpty(project.Name, "(unnamed project)"), project.ID)
	}

	buffered := asBufferedReader(reader)
	for {
		fmt.Fprintf(writer, "Enter selection (0-%d): ", len(projects))
		line, err := buffered.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read project selection: %w", err)
		}

		choiceText := strings.TrimSpace(line)
		if choiceText == "" {
			if err == io.EOF {
				return nil, nil
			}
			fmt.Fprintln(writer, "Selection is required.")
			continue
		}

		choice, convErr := strconv.Atoi(choiceText)
		if convErr != nil {
			fmt.Fprintln(writer, "Please enter a valid number.")
			continue
		}

		if choice == 0 {
			return nil, nil
		}
		if choice < 0 || choice > len(projects) {
			fmt.Fprintln(writer, "Selection out of range.")
			continue
		}

		selected := projects[choice-1]
		return &selected, nil
	}
}

type quoteSelectionInput struct {
	ProjectID          string
	TaskIDs            []string
	Interactive        bool
	Yes                bool
	CreateQuoteFlag    bool
	ConvertInvoiceFlag bool
}

func enrichQuoteWithNinjaSelections(ctx context.Context, source quoteWorkflowSource, quoteSpec *spec.QuoteSpec, input quoteSelectionInput, reader io.Reader, writer io.Writer) ([]invoiceninja.NinjaTask, error) {
	if source == nil {
		return nil, nil
	}

	if quoteSpec.Client.ID == "" && input.Interactive {
		client, err := selectOrCreateClient(ctx, source, reader, writer)
		if err != nil {
			return nil, err
		}
		mapNinjaClientToQuote(quoteSpec, client)
	}

	if strings.TrimSpace(input.ProjectID) != "" {
		project, err := source.GetProject(ctx, strings.TrimSpace(input.ProjectID))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve --project-id %q: %w", strings.TrimSpace(input.ProjectID), err)
		}
		if project != nil {
			projectClientID := strings.TrimSpace(project.ClientID)
			if strings.TrimSpace(quoteSpec.Client.ID) != "" && projectClientID != "" && strings.TrimSpace(quoteSpec.Client.ID) != projectClientID {
				return nil, fmt.Errorf("project %q belongs to client %q but quote is set to client %q", strings.TrimSpace(project.ID), projectClientID, strings.TrimSpace(quoteSpec.Client.ID))
			}
			mapNinjaProjectToQuote(quoteSpec, project)
			if quoteSpec.Client.ID == "" {
				quoteSpec.Client.ID = strings.TrimSpace(project.ClientID)
			}
		}
	} else if input.Interactive && strings.TrimSpace(quoteSpec.Client.ID) != "" {
		project, err := selectOrCreateProject(ctx, source, reader, writer, quoteSpec.Client.ID)
		if err != nil {
			return nil, err
		}
		mapNinjaProjectToQuote(quoteSpec, project)
	}

	selectedTasks, err := resolveSelectedTasks(ctx, source, quoteSpec, input, reader, writer)
	if err != nil {
		return nil, err
	}

	if err := validateAndDeriveTaskRelationships(quoteSpec, selectedTasks); err != nil {
		return nil, err
	}

	quoteSpec.TaskIDs = extractTaskIDs(selectedTasks)
	if len(quoteSpec.TaskIDs) == 0 && len(input.TaskIDs) > 0 {
		quoteSpec.TaskIDs = append([]string(nil), input.TaskIDs...)
	}

	return selectedTasks, nil
}

func selectOrCreateClient(ctx context.Context, source quoteWorkflowSource, reader io.Reader, writer io.Writer) (*invoiceninja.NinjaClient, error) {
	buffered := asBufferedReader(reader)

	clientList, err := source.ListClients(ctx, 1, interactiveSelectMaxRows)
	if err != nil {
		return nil, fmt.Errorf("failed to list clients: %w", err)
	}

	clients := clientList.Data
	fmt.Fprintln(writer, "Select a client from Invoice Ninja:")
	fmt.Fprintln(writer, "  0) Keep template defaults")
	fmt.Fprintln(writer, "  1) Create new client")
	for i, client := range clients {
		displayName := firstNonEmpty(client.DisplayName, client.Name, "(unnamed client)")
		email := firstNonEmpty(primaryContactEmail(client.Contacts), client.Email, "no-email")
		fmt.Fprintf(writer, "  %d) %s <%s> [id=%s]\n", i+2, displayName, email, client.ID)
	}

	choice, err := promptNumericChoice(buffered, writer, "Enter selection", 0, len(clients)+1)
	if err != nil {
		return nil, err
	}

	if choice == 0 {
		return nil, nil
	}
	if choice == 1 {
		name, promptErr := promptField(buffered, writer, "Client name", "", true)
		if promptErr != nil {
			return nil, promptErr
		}
		email, promptErr := promptField(buffered, writer, "Client email (optional)", "", false)
		if promptErr != nil {
			return nil, promptErr
		}

		created, createErr := source.CreateClient(ctx, invoiceninja.CreateClientRequest{Name: name, Email: email})
		if createErr != nil {
			return nil, fmt.Errorf("failed to create client: %w", createErr)
		}
		fmt.Fprintf(writer, "✓ Created client %s [id=%s]\n", firstNonEmpty(created.DisplayName, created.Name), created.ID)
		return created, nil
	}

	selected := clients[choice-2]
	return &selected, nil
}

func selectOrCreateProject(ctx context.Context, source quoteWorkflowSource, reader io.Reader, writer io.Writer, clientID string) (*invoiceninja.NinjaProject, error) {
	buffered := asBufferedReader(reader)

	projectList, err := source.ListProjects(ctx, 1, interactiveSelectMaxRows)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	projects := filterProjectsByClientID(projectList.Data, clientID)
	fmt.Fprintln(writer, "Select a project from Invoice Ninja:")
	fmt.Fprintln(writer, "  0) Keep current/manual project details")
	fmt.Fprintln(writer, "  1) Create new project")
	for i, project := range projects {
		fmt.Fprintf(writer, "  %d) %s [id=%s]\n", i+2, firstNonEmpty(project.Name, "(unnamed project)"), project.ID)
	}

	choice, err := promptNumericChoice(buffered, writer, "Enter selection", 0, len(projects)+1)
	if err != nil {
		return nil, err
	}

	if choice == 0 {
		return nil, nil
	}
	if choice == 1 {
		name, promptErr := promptField(buffered, writer, "Project name", "", true)
		if promptErr != nil {
			return nil, promptErr
		}
		description, promptErr := promptField(buffered, writer, "Project description", "", false)
		if promptErr != nil {
			return nil, promptErr
		}
		created, createErr := source.CreateProject(ctx, invoiceninja.CreateProjectRequest{ClientID: clientID, Name: name, Description: description})
		if createErr != nil {
			return nil, fmt.Errorf("failed to create project: %w", createErr)
		}
		fmt.Fprintf(writer, "✓ Created project %s [id=%s]\n", firstNonEmpty(created.Name, "(unnamed project)"), created.ID)
		return created, nil
	}

	selected := projects[choice-2]
	return &selected, nil
}

func resolveSelectedTasks(ctx context.Context, source quoteWorkflowSource, quoteSpec *spec.QuoteSpec, input quoteSelectionInput, reader io.Reader, writer io.Writer) ([]invoiceninja.NinjaTask, error) {
	if len(input.TaskIDs) > 0 {
		return fetchTasksByID(ctx, source, input.TaskIDs)
	}

	if !input.Interactive || strings.TrimSpace(quoteSpec.Project.ID) == "" {
		return nil, nil
	}

	taskList, err := source.ListTasks(ctx, 1, interactiveSelectMaxRows)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	projectTasks := filterTasksByProjectID(taskList.Data, quoteSpec.Project.ID)
	selected, err := promptTaskSelection(reader, writer, projectTasks)
	if err != nil {
		return nil, err
	}

	if !input.Yes {
		for {
			addTask, askErr := promptYesNo(reader, writer, "Add a new task for this project?", false)
			if askErr != nil {
				return nil, askErr
			}
			if !addTask {
				break
			}

			created, createErr := promptCreateTask(ctx, source, quoteSpec.Client.ID, quoteSpec.Project.ID, reader, writer)
			if createErr != nil {
				return nil, createErr
			}
			if created != nil {
				selected = append(selected, *created)
			}
		}
	}

	return selected, nil
}

func fetchTasksByID(ctx context.Context, source quoteWorkflowSource, ids []string) ([]invoiceninja.NinjaTask, error) {
	result := make([]invoiceninja.NinjaTask, 0, len(ids))
	for _, id := range ids {
		task, err := source.GetTask(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve task %q: %w", id, err)
		}
		if task == nil || strings.TrimSpace(task.ID) == "" {
			return nil, fmt.Errorf("no Invoice Ninja task found for --task-ids value %q", strings.TrimSpace(id))
		}
		result = append(result, *task)
	}
	return result, nil
}

func validateAndDeriveTaskRelationships(quoteSpec *spec.QuoteSpec, tasks []invoiceninja.NinjaTask) error {
	if quoteSpec == nil || len(tasks) == 0 {
		return nil
	}

	trimmedClientID := strings.TrimSpace(quoteSpec.Client.ID)
	trimmedProjectID := strings.TrimSpace(quoteSpec.Project.ID)

	if trimmedClientID != "" {
		for _, task := range tasks {
			taskClientID := strings.TrimSpace(task.ClientID)
			if taskClientID != "" && taskClientID != trimmedClientID {
				return fmt.Errorf("task %q belongs to client %q but quote is set to client %q", strings.TrimSpace(task.ID), taskClientID, trimmedClientID)
			}
		}
	}

	if trimmedProjectID != "" {
		for _, task := range tasks {
			taskProjectID := strings.TrimSpace(task.ProjectID)
			if taskProjectID != "" && taskProjectID != trimmedProjectID {
				return fmt.Errorf("task %q belongs to project %q but quote is set to project %q", strings.TrimSpace(task.ID), taskProjectID, trimmedProjectID)
			}
		}
	}

	if trimmedClientID == "" {
		derivedClientID, err := deriveSingleTaskRelationshipID(tasks, func(task invoiceninja.NinjaTask) string {
			return task.ClientID
		}, "client")
		if err != nil {
			return err
		}
		if derivedClientID != "" {
			quoteSpec.Client.ID = derivedClientID
		}
	}

	if trimmedProjectID == "" {
		derivedProjectID, err := deriveSingleTaskRelationshipID(tasks, func(task invoiceninja.NinjaTask) string {
			return task.ProjectID
		}, "project")
		if err != nil {
			return err
		}
		if derivedProjectID != "" {
			quoteSpec.Project.ID = derivedProjectID
		}
	}

	return nil
}

func deriveSingleTaskRelationshipID(tasks []invoiceninja.NinjaTask, extract func(invoiceninja.NinjaTask) string, label string) (string, error) {
	values := make(map[string]struct{})
	for _, task := range tasks {
		value := strings.TrimSpace(extract(task))
		if value == "" {
			continue
		}
		values[value] = struct{}{}
	}

	if len(values) == 0 {
		return "", nil
	}
	if len(values) > 1 {
		resolved := make([]string, 0, len(values))
		for value := range values {
			resolved = append(resolved, value)
		}
		sort.Strings(resolved)
		return "", fmt.Errorf("selected tasks span multiple %s IDs (%s); select tasks from a single %s or pass --%s-id", label, strings.Join(resolved, ", "), label, label)
	}

	for value := range values {
		return value, nil
	}

	return "", nil
}

func promptTaskSelection(reader io.Reader, writer io.Writer, tasks []invoiceninja.NinjaTask) ([]invoiceninja.NinjaTask, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	fmt.Fprintln(writer, "Select existing tasks (comma-separated indexes, blank for none):")
	for i, task := range tasks {
		hours := formatHours(task.Duration)
		fmt.Fprintf(writer, "  %d) %s [id=%s, hours=%s, rate=%.2f]\n", i+1, firstNonEmpty(strings.TrimSpace(task.Description), "(unnamed task)"), task.ID, hours, task.Rate)
	}
	fmt.Fprint(writer, "Task selections: ")

	line, err := asBufferedReader(reader).ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read task selection: %w", err)
	}

	indexes, parseErr := parseNumericSelections(line, len(tasks))
	if parseErr != nil {
		return nil, parseErr
	}

	selected := make([]invoiceninja.NinjaTask, 0, len(indexes))
	for _, idx := range indexes {
		selected = append(selected, tasks[idx-1])
	}

	return selected, nil
}

func promptCreateTask(ctx context.Context, source quoteWorkflowSource, clientID, projectID string, reader io.Reader, writer io.Writer) (*invoiceninja.NinjaTask, error) {
	buffered := asBufferedReader(reader)
	description, err := promptField(buffered, writer, "Task description", "", true)
	if err != nil {
		return nil, err
	}
	hoursText, err := promptField(buffered, writer, "Task hours (optional)", "", false)
	if err != nil {
		return nil, err
	}
	rateText, err := promptField(buffered, writer, "Task rate (optional)", "", false)
	if err != nil {
		return nil, err
	}

	hours, err := parseOptionalFloat(hoursText)
	if err != nil {
		return nil, fmt.Errorf("invalid task hours: %w", err)
	}
	rate, err := parseOptionalFloat(rateText)
	if err != nil {
		return nil, fmt.Errorf("invalid task rate: %w", err)
	}

	request := invoiceninja.CreateTaskRequest{
		ClientID:    strings.TrimSpace(clientID),
		ProjectID:   strings.TrimSpace(projectID),
		Description: strings.TrimSpace(description),
		Rate:        rate,
	}
	if hours > 0 {
		request.Duration = int64(math.Round(hours * 3600))
	}

	created, createErr := source.CreateTask(ctx, request)
	if createErr != nil {
		return nil, fmt.Errorf("failed to create task: %w", createErr)
	}

	fmt.Fprintf(writer, "✓ Created task %s [id=%s]\n", firstNonEmpty(strings.TrimSpace(created.Description), "(unnamed task)"), created.ID)
	return created, nil
}

func promptNumericChoice(reader io.Reader, writer io.Writer, label string, min, max int) (int, error) {
	buffered := asBufferedReader(reader)
	for {
		fmt.Fprintf(writer, "%s (%d-%d): ", label, min, max)
		line, err := buffered.ReadString('\n')
		if err != nil && err != io.EOF {
			return 0, err
		}

		choiceText := strings.TrimSpace(line)
		if choiceText == "" {
			if err == io.EOF {
				return min, nil
			}
			fmt.Fprintln(writer, "Selection is required.")
			continue
		}

		choice, convErr := strconv.Atoi(choiceText)
		if convErr != nil || choice < min || choice > max {
			fmt.Fprintln(writer, "Please enter a valid number.")
			continue
		}

		return choice, nil
	}
}

func promptYesNo(reader io.Reader, writer io.Writer, prompt string, defaultYes bool) (bool, error) {
	defaultHint := "y/N"
	if defaultYes {
		defaultHint = "Y/n"
	}

	fmt.Fprintf(writer, "%s [%s]: ", prompt, defaultHint)
	line, err := readLine(reader)
	if err != nil && err != io.EOF {
		return false, err
	}

	value := strings.TrimSpace(strings.ToLower(line))
	if value == "" {
		return defaultYes, nil
	}
	return value == "y" || value == "yes", nil
}

func readLine(reader io.Reader) (string, error) {
	return asBufferedReader(reader).ReadString('\n')
}

func asBufferedReader(reader io.Reader) *bufio.Reader {
	if buffered, ok := reader.(*bufio.Reader); ok {
		return buffered
	}
	return bufio.NewReader(reader)
}

func parseOptionalFloat(value string) (float64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func parseCSVIdentifiers(csv string) []string {
	parts := strings.Split(csv, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func parseNumericSelections(raw string, max int) ([]int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	parts := strings.Split(trimmed, ",")
	selected := make([]int, 0, len(parts))
	seen := make(map[int]struct{}, len(parts))
	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		idx, err := strconv.Atoi(candidate)
		if err != nil {
			return nil, fmt.Errorf("invalid selection %q", candidate)
		}
		if idx < 1 || idx > max {
			return nil, fmt.Errorf("selection %d out of range", idx)
		}
		if _, exists := seen[idx]; exists {
			continue
		}
		seen[idx] = struct{}{}
		selected = append(selected, idx)
	}
	return selected, nil
}

func filterTasksByProjectID(tasks []invoiceninja.NinjaTask, projectID string) []invoiceninja.NinjaTask {
	trimmedProjectID := strings.TrimSpace(projectID)
	if trimmedProjectID == "" {
		return tasks
	}

	filtered := make([]invoiceninja.NinjaTask, 0, len(tasks))
	for _, task := range tasks {
		if strings.TrimSpace(task.ProjectID) == trimmedProjectID {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

func extractTaskIDs(tasks []invoiceninja.NinjaTask) []string {
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		if trimmed := strings.TrimSpace(task.ID); trimmed != "" {
			ids = append(ids, trimmed)
		}
	}
	return ids
}

func mapTasksToLineItems(tasks []invoiceninja.NinjaTask) []spec.LineItem {
	lineItems := make([]spec.LineItem, 0, len(tasks))
	for _, task := range tasks {
		description := firstNonEmpty(strings.TrimSpace(task.Description), "Task "+strings.TrimSpace(task.Number), "Task")
		quantity := taskDurationHours(task.Duration)
		rate := task.Rate
		amount := quantity * rate
		lineItems = append(lineItems, spec.LineItem{
			Description: description,
			Quantity:    quantity,
			Rate:        rate,
			Amount:      amount,
			Category:    "task",
		})
	}
	return lineItems
}

func mergeTaskLineItems(quoteSpec *spec.QuoteSpec, taskItems []spec.LineItem, interactive bool, yes bool, reader io.Reader, writer io.Writer) error {
	if len(taskItems) == 0 {
		return nil
	}

	if len(quoteSpec.Pricing.LineItems) == 0 {
		quoteSpec.Pricing.LineItems = append([]spec.LineItem(nil), taskItems...)
		return nil
	}

	replace := false
	if interactive && !yes {
		confirmReplace, err := promptYesNo(reader, writer, "Replace existing line items with selected task items? (default appends)", false)
		if err != nil {
			return err
		}
		replace = confirmReplace
	}

	if replace {
		quoteSpec.Pricing.LineItems = append([]spec.LineItem(nil), taskItems...)
		return nil
	}

	quoteSpec.Pricing.LineItems = append(quoteSpec.Pricing.LineItems, taskItems...)
	return nil
}

func taskDurationHours(durationSeconds int64) float64 {
	if durationSeconds <= 0 {
		return 0
	}
	return float64(durationSeconds) / 3600
}

func formatHours(durationSeconds int64) string {
	hours := taskDurationHours(durationSeconds)
	return strconv.FormatFloat(hours, 'f', 2, 64)
}

func editQuoteArtifactsInline(reader io.Reader, writer io.Writer, artifacts *spec.GeneratedArtifacts) error {
	if artifacts == nil {
		return nil
	}

	editedTerms, err := maybeEditArtifactText(reader, writer, "terms", artifacts.TermsMarkdown)
	if err != nil {
		return err
	}
	editedNotes, err := maybeEditArtifactText(reader, writer, "public notes", artifacts.PublicNotesText)
	if err != nil {
		return err
	}

	artifacts.TermsMarkdown = editedTerms
	artifacts.PublicNotesText = editedNotes
	artifacts.Meta.Hash = recomputeArtifactHash(artifacts)
	return nil
}

func maybeEditArtifactText(reader io.Reader, writer io.Writer, label, current string) (string, error) {
	buffered := asBufferedReader(reader)
	edit, err := promptYesNo(buffered, writer, fmt.Sprintf("Edit %s now?", label), false)
	if err != nil {
		return "", err
	}
	if !edit {
		return current, nil
	}

	fmt.Fprintf(writer, "Enter %s. Finish input with a single line containing only .\n", label)
	lines := make([]string, 0)
	for {
		line, readErr := buffered.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return "", readErr
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if strings.TrimSpace(trimmed) == "." {
			break
		}
		if readErr == io.EOF {
			if strings.TrimSpace(trimmed) != "." {
				lines = append(lines, trimmed)
			}
			break
		}
		lines = append(lines, trimmed)
	}

	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}

func recomputeArtifactHash(artifacts *spec.GeneratedArtifacts) string {
	h := sha256.New()
	h.Write([]byte(artifacts.ProposalMarkdown))
	h.Write([]byte(artifacts.TermsMarkdown))
	h.Write([]byte(artifacts.PublicNotesText))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func printQuotePreview(writer io.Writer, quoteSpec *spec.QuoteSpec, selectedTasks []invoiceninja.NinjaTask, artifacts *spec.GeneratedArtifacts) {
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, "--- Quote Preview ---")
	fmt.Fprintf(writer, "Client: %s [id=%s]\n", firstNonEmpty(quoteSpec.Client.Name, "(unset)"), firstNonEmpty(quoteSpec.Client.ID, "-"))
	fmt.Fprintf(writer, "Project: %s [id=%s]\n", firstNonEmpty(quoteSpec.Project.Name, "(unset)"), firstNonEmpty(quoteSpec.Project.ID, "-"))
	fmt.Fprintf(writer, "Tasks selected: %d\n", len(selectedTasks))
	if len(selectedTasks) > 0 {
		for _, task := range selectedTasks {
			fmt.Fprintf(writer, "  - %s [id=%s]\n", firstNonEmpty(strings.TrimSpace(task.Description), "(unnamed task)"), task.ID)
		}
	}

	total := quoteSpec.CalculateTotal()
	fmt.Fprintf(writer, "Line items: %d (total %.2f %s)\n", len(quoteSpec.Pricing.LineItems), total, firstNonEmpty(quoteSpec.Pricing.Currency, "USD"))
	for _, item := range quoteSpec.Pricing.LineItems {
		fmt.Fprintf(writer, "  - %s | qty %.2f @ %.2f = %.2f\n", item.Description, item.Quantity, item.Rate, item.Amount)
	}

	if artifacts != nil {
		fmt.Fprintf(writer, "Terms: %s\n", previewText(artifacts.TermsMarkdown, previewSnippetLimit))
		fmt.Fprintf(writer, "Public notes: %s\n", previewText(artifacts.PublicNotesText, previewSnippetLimit))
	}
	fmt.Fprintln(writer, "---------------------")
	fmt.Fprintln(writer, "")
}

func previewText(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= max {
		return trimmed
	}
	if max < 4 {
		return trimmed[:max]
	}
	return trimmed[:max-3] + "..."
}

func validateCreateConvertDecision(hasToken bool, createQuote bool, convertToInvoice bool) (quoteCreateDecision, error) {
	decision := quoteCreateDecision{CreateQuote: createQuote, ConvertToInvoice: convertToInvoice}
	if decision.ConvertToInvoice {
		decision.CreateQuote = true
	}

	if (decision.CreateQuote || decision.ConvertToInvoice) && !hasToken {
		return decision, fmt.Errorf("creating or converting quotes requires Invoice Ninja API token configuration")
	}

	return decision, nil
}

func resolveCreateConvertDecision(decision quoteCreateDecision, hasToken bool, yes bool, reader io.Reader, writer io.Writer) (quoteCreateDecision, error) {
	if !hasToken || yes {
		return decision, nil
	}

	if !decision.CreateQuote {
		createQuote, err := promptYesNo(reader, writer, "Create quote in Invoice Ninja?", false)
		if err != nil {
			return decision, err
		}
		decision.CreateQuote = createQuote
	}

	if decision.CreateQuote && !decision.ConvertToInvoice {
		convert, err := promptYesNo(reader, writer, "Convert created quote to invoice?", false)
		if err != nil {
			return decision, err
		}
		decision.ConvertToInvoice = convert
	}

	return decision, nil
}

func writeNewQuoteArtifacts(outputPath string, artifactsDir string, quoteSpec *spec.QuoteSpec, prebuiltArtifacts ...*spec.GeneratedArtifacts) error {
	var artifacts *spec.GeneratedArtifacts
	if len(prebuiltArtifacts) > 0 && prebuiltArtifacts[0] != nil {
		artifacts = prebuiltArtifacts[0]
	} else {
		generator := generate.NewGenerator()
		generated, err := generator.Generate(quoteSpec)
		if err != nil {
			return fmt.Errorf("failed to generate quote artifacts: %w", err)
		}
		artifacts = generated
	}

	finalDir := strings.TrimSpace(artifactsDir)
	if finalDir == "" {
		finalDir = filepath.Dir(outputPath)
	}

	if err := os.MkdirAll(finalDir, 0750); err != nil {
		return fmt.Errorf("failed to create artifacts directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(finalDir, defaultTermsFilename), []byte(artifacts.TermsMarkdown), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", defaultTermsFilename, err)
	}

	if err := os.WriteFile(filepath.Join(finalDir, defaultPublicNotesFile), []byte(artifacts.PublicNotesText), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", defaultPublicNotesFile, err)
	}

	return nil
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
