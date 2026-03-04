package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ninjops/ninjops/internal/invoiceninja"
	"github.com/ninjops/ninjops/internal/spec"
)

func TestParseCSVIdentifiers_TrimAndDeduplicate(t *testing.T) {
	ids := parseCSVIdentifiers(" t1, t2, t1 ,,t3 ")
	if len(ids) != 3 {
		t.Fatalf("expected 3 ids, got %d", len(ids))
	}
	if ids[0] != "t1" || ids[1] != "t2" || ids[2] != "t3" {
		t.Fatalf("unexpected parsed ids: %#v", ids)
	}
}

func TestFilterTasksByProjectID(t *testing.T) {
	tasks := []invoiceninja.NinjaTask{
		{ID: "t1", ProjectID: "p1"},
		{ID: "t2", ProjectID: "p2"},
		{ID: "t3", ProjectID: "p1"},
	}

	filtered := filterTasksByProjectID(tasks, "p1")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(filtered))
	}
	if filtered[0].ID != "t1" || filtered[1].ID != "t3" {
		t.Fatalf("unexpected filtered tasks: %#v", filtered)
	}
}

func TestPromptTaskSelection_SelectsMultiple(t *testing.T) {
	tasks := []invoiceninja.NinjaTask{
		{ID: "t1", Description: "Design"},
		{ID: "t2", Description: "Build"},
		{ID: "t3", Description: "QA"},
	}

	in := bytes.NewBufferString("1,3\n")
	out := &bytes.Buffer{}

	selected, err := promptTaskSelection(in, out, tasks)
	if err != nil {
		t.Fatalf("promptTaskSelection returned error: %v", err)
	}
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected tasks, got %d", len(selected))
	}
	if selected[0].ID != "t1" || selected[1].ID != "t3" {
		t.Fatalf("unexpected selected tasks: %#v", selected)
	}
}

func TestMapTasksToLineItems(t *testing.T) {
	tasks := []invoiceninja.NinjaTask{{
		ID:          "t1",
		Description: "Implementation",
		Rate:        125,
		Duration:    7200,
	}}

	items := mapTasksToLineItems(tasks)
	if len(items) != 1 {
		t.Fatalf("expected one line item, got %d", len(items))
	}
	if items[0].Description != "Implementation" {
		t.Fatalf("unexpected description: %q", items[0].Description)
	}
	if items[0].Quantity != 2 {
		t.Fatalf("unexpected quantity: %f", items[0].Quantity)
	}
	if items[0].Rate != 125 {
		t.Fatalf("unexpected rate: %f", items[0].Rate)
	}
	if items[0].Amount != 250 {
		t.Fatalf("unexpected amount: %f", items[0].Amount)
	}
}

func TestMergeTaskLineItems_AppendsByDefault(t *testing.T) {
	quoteSpec := newQuoteTemplate()
	quoteSpec.Pricing.LineItems = []spec.LineItem{{Description: "Existing", Quantity: 1, Rate: 10, Amount: 10}}

	taskItems := []spec.LineItem{{Description: "Task", Quantity: 2, Rate: 50, Amount: 100}}
	if err := mergeTaskLineItems(quoteSpec, taskItems, false, false, bytes.NewBufferString(""), &bytes.Buffer{}); err != nil {
		t.Fatalf("mergeTaskLineItems returned error: %v", err)
	}

	if len(quoteSpec.Pricing.LineItems) != 2 {
		t.Fatalf("expected append behavior with 2 items, got %d", len(quoteSpec.Pricing.LineItems))
	}
}

func TestMaybeEditArtifactText_UsesEditedInput(t *testing.T) {
	in := bytes.NewBufferString("y\nFirst line\nSecond line\n.\n")
	out := &bytes.Buffer{}

	edited, err := maybeEditArtifactText(in, out, "terms", "original")
	if err != nil {
		t.Fatalf("maybeEditArtifactText returned error: %v", err)
	}

	if edited != "First line\nSecond line" {
		t.Fatalf("unexpected edited content: %q", edited)
	}
}

func TestMaybeEditArtifactText_DeclineKeepsCurrent(t *testing.T) {
	in := bytes.NewBufferString("n\n")
	out := &bytes.Buffer{}

	edited, err := maybeEditArtifactText(in, out, "public notes", "keep me")
	if err != nil {
		t.Fatalf("maybeEditArtifactText returned error: %v", err)
	}
	if edited != "keep me" {
		t.Fatalf("expected original text, got %q", edited)
	}
}

func TestValidateCreateConvertDecision_ConvertImpliesCreate(t *testing.T) {
	decision, err := validateCreateConvertDecision(true, false, true)
	if err != nil {
		t.Fatalf("validateCreateConvertDecision returned error: %v", err)
	}
	if !decision.CreateQuote {
		t.Fatalf("expected convert-to-invoice to imply create-quote")
	}
	if !decision.ConvertToInvoice {
		t.Fatalf("expected convert flag to remain true")
	}
}

func TestValidateCreateConvertDecision_RequiresTokenForCreate(t *testing.T) {
	_, err := validateCreateConvertDecision(false, true, false)
	if err == nil {
		t.Fatalf("expected token validation error")
	}
	if !strings.Contains(err.Error(), "API token") {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestEnrichQuoteWithNinjaSelections_ProjectClientMismatch(t *testing.T) {
	quoteSpec := newQuoteTemplate()
	quoteSpec.Client.ID = "client-a"

	stub := &stubQuoteWorkflowSource{
		project: &invoiceninja.NinjaProject{ID: "project-x", ClientID: "client-b"},
	}

	_, err := enrichQuoteWithNinjaSelections(
		context.Background(),
		stub,
		quoteSpec,
		quoteSelectionInput{ProjectID: "project-x", Interactive: false},
		bytes.NewBufferString(""),
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatalf("expected mismatch validation error")
	}

	if !strings.Contains(err.Error(), "belongs to client") {
		t.Fatalf("unexpected mismatch error: %v", err)
	}
}

func TestWriteFile_CreatesParentDirectories(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "nested", "dir", "quote.json")

	if err := writeFile(path, []byte("{\"ok\":true}")); err != nil {
		t.Fatalf("writeFile returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile returned error: %v", err)
	}

	if string(data) != "{\"ok\":true}" {
		t.Fatalf("unexpected file contents: %s", string(data))
	}
}

func TestNewQuote_NonInteractiveProjectFlagsAndArtifacts(t *testing.T) {
	tmp := t.TempDir()
	outputPath := filepath.Join(tmp, "quotes", "quote.json")
	artifactsPath := filepath.Join(tmp, "artifacts")

	cmd := newNewQuoteCmd()
	cmd.SetArgs([]string{
		"--non-interactive",
		"--project-name", "CLI Project",
		"--project-description", "Project description from flags",
		"--project-type", "integration",
		"--project-timeline", "6 weeks",
		"--output", outputPath,
		"--artifacts-dir", artifactsPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("new quote command failed: %v", err)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output quote spec: %v", err)
	}

	quoteSpec, err := spec.FromJSON(raw)
	if err != nil {
		t.Fatalf("failed to parse generated quote spec: %v", err)
	}

	if quoteSpec.Project.Name != "CLI Project" {
		t.Fatalf("unexpected project name: %q", quoteSpec.Project.Name)
	}
	if quoteSpec.Project.Description != "Project description from flags" {
		t.Fatalf("unexpected project description: %q", quoteSpec.Project.Description)
	}
	if quoteSpec.Project.Type != "integration" {
		t.Fatalf("unexpected project type: %q", quoteSpec.Project.Type)
	}
	if quoteSpec.Project.Timeline != "6 weeks" {
		t.Fatalf("unexpected project timeline: %q", quoteSpec.Project.Timeline)
	}

	termsPath := filepath.Join(artifactsPath, defaultTermsFilename)
	notesPath := filepath.Join(artifactsPath, defaultPublicNotesFile)

	terms, err := os.ReadFile(termsPath)
	if err != nil {
		t.Fatalf("failed to read generated terms file: %v", err)
	}
	notes, err := os.ReadFile(notesPath)
	if err != nil {
		t.Fatalf("failed to read generated notes file: %v", err)
	}

	if len(strings.TrimSpace(string(terms))) == 0 {
		t.Fatalf("terms file is empty")
	}
	if len(strings.TrimSpace(string(notes))) == 0 {
		t.Fatalf("public notes file is empty")
	}
}

func TestNewQuote_DefaultArtifactsDirUsesOutputDirectory(t *testing.T) {
	tmp := t.TempDir()
	outputPath := filepath.Join(tmp, "quotes", "quote.json")

	cmd := newNewQuoteCmd()
	cmd.SetArgs([]string{"--non-interactive", "--output", outputPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("new quote command failed: %v", err)
	}

	outputDir := filepath.Dir(outputPath)
	if _, err := os.Stat(filepath.Join(outputDir, defaultTermsFilename)); err != nil {
		t.Fatalf("expected terms file beside output: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, defaultPublicNotesFile)); err != nil {
		t.Fatalf("expected public notes file beside output: %v", err)
	}
}

func TestResolveClientSelection_WithSelectorLookup(t *testing.T) {
	stub := &stubClientSource{
		findByEmailResult: &invoiceninja.NinjaClient{ID: "42", Name: "Acme Corp", Email: "billing@acme.test"},
	}

	selected, err := resolveClientSelection(context.Background(), stub, "", "billing@acme.test", "")
	if err != nil {
		t.Fatalf("resolveClientSelection returned error: %v", err)
	}
	if selected == nil {
		t.Fatalf("expected a selected client")
	}
	if selected.ID != "42" {
		t.Fatalf("unexpected selected client id: %q", selected.ID)
	}
	if stub.findByEmailCalls != 1 {
		t.Fatalf("expected one email lookup call, got %d", stub.findByEmailCalls)
	}
}

func TestResolveClientSelection_ReturnsHelpfulNotFoundError(t *testing.T) {
	stub := &stubClientSource{}

	_, err := resolveClientSelection(context.Background(), stub, "", "missing@example.com", "")
	if err == nil {
		t.Fatalf("expected error when email selector has no match")
	}

	if !strings.Contains(err.Error(), "--client-email") {
		t.Fatalf("expected error to reference --client-email, got %q", err.Error())
	}
}

func TestPromptClientSelection_ChoosesNumberedClient(t *testing.T) {
	clients := []invoiceninja.NinjaClient{
		{ID: "1", Name: "Alpha", Email: "a@example.com"},
		{ID: "2", Name: "Beta", Email: "b@example.com"},
	}

	in := bytes.NewBufferString("2\n")
	out := &bytes.Buffer{}

	selected, err := promptClientSelection(in, out, clients)
	if err != nil {
		t.Fatalf("promptClientSelection returned error: %v", err)
	}
	if selected == nil {
		t.Fatalf("expected client selection")
	}
	if selected.ID != "2" {
		t.Fatalf("unexpected selection id: %q", selected.ID)
	}
}

func TestPromptProjectSelection_ChoosesNumberedProject(t *testing.T) {
	projects := []invoiceninja.NinjaProject{
		{ID: "p1", Name: "Website Refresh"},
		{ID: "p2", Name: "Mobile API"},
	}

	in := bytes.NewBufferString("2\n")
	out := &bytes.Buffer{}

	selected, err := promptProjectSelection(in, out, projects)
	if err != nil {
		t.Fatalf("promptProjectSelection returned error: %v", err)
	}
	if selected == nil {
		t.Fatalf("expected project selection")
	}
	if selected.ID != "p2" {
		t.Fatalf("unexpected selection id: %q", selected.ID)
	}
}

func TestMapNinjaProjectToQuote_MapsProjectDetails(t *testing.T) {
	quoteSpec := newQuoteTemplate()

	mapNinjaProjectToQuote(quoteSpec, &invoiceninja.NinjaProject{
		Name:         "Portal Rebuild",
		PublicNotes:  "Rebuild customer portal",
		PrivateNotes: "Internal-only notes",
		DueDate:      "2026-07-15",
	})

	if quoteSpec.Project.Name != "Portal Rebuild" {
		t.Fatalf("unexpected project name: %q", quoteSpec.Project.Name)
	}
	if quoteSpec.Project.Description != "Rebuild customer portal" {
		t.Fatalf("unexpected project description: %q", quoteSpec.Project.Description)
	}
	if quoteSpec.Project.Deadline != "2026-07-15" {
		t.Fatalf("unexpected project deadline: %q", quoteSpec.Project.Deadline)
	}
}

func TestMaybeSelectClientProject_UsesStubSourceAndFiltersByClient(t *testing.T) {
	stub := &stubClientSource{
		listProjectsResult: []invoiceninja.NinjaProject{
			{ID: "p1", ClientID: "c1", Name: "Client One Alpha"},
			{ID: "p2", ClientID: "c2", Name: "Client Two"},
			{ID: "p3", ClientID: "c1", Name: "Client One Beta"},
		},
	}

	in := bytes.NewBufferString("2\n")
	out := &bytes.Buffer{}

	selected, err := maybeSelectClientProject(context.Background(), stub, in, out, &invoiceninja.NinjaClient{ID: "c1"})
	if err != nil {
		t.Fatalf("maybeSelectClientProject returned error: %v", err)
	}
	if selected == nil {
		t.Fatalf("expected project selection")
	}
	if selected.ID != "p3" {
		t.Fatalf("unexpected selected project id: %q", selected.ID)
	}
}

func TestSelectOrCreateClient_UsesSharedBufferedReader(t *testing.T) {
	stub := &stubQuoteWorkflowSource{}
	in := bytes.NewBufferString("1\nAcme Client\nbilling@acme.test\n")
	out := &bytes.Buffer{}

	selected, err := selectOrCreateClient(context.Background(), stub, in, out)
	if err != nil {
		t.Fatalf("selectOrCreateClient returned error: %v", err)
	}
	if selected == nil {
		t.Fatalf("expected created client")
	}
	if selected.Name != "Acme Client" {
		t.Fatalf("unexpected created client name: %q", selected.Name)
	}
	if selected.Email != "billing@acme.test" {
		t.Fatalf("unexpected created client email: %q", selected.Email)
	}
}

func TestEnrichQuoteWithNinjaSelections_TaskIDClientMismatch(t *testing.T) {
	quoteSpec := newQuoteTemplate()
	quoteSpec.Client.ID = "client-a"

	stub := &stubQuoteWorkflowSource{
		tasksByID: map[string]*invoiceninja.NinjaTask{
			"task-1": {ID: "task-1", ClientID: "client-b", ProjectID: "project-1", Description: "Build"},
		},
	}

	_, err := enrichQuoteWithNinjaSelections(
		context.Background(),
		stub,
		quoteSpec,
		quoteSelectionInput{TaskIDs: []string{"task-1"}, Interactive: false},
		bytes.NewBufferString(""),
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatalf("expected client relationship mismatch error")
	}
	if !strings.Contains(err.Error(), "belongs to client") {
		t.Fatalf("unexpected mismatch error: %v", err)
	}
}

func TestEnrichQuoteWithNinjaSelections_DerivesClientAndProjectFromTasks(t *testing.T) {
	quoteSpec := newQuoteTemplate()
	quoteSpec.Client.ID = ""
	quoteSpec.Project.ID = ""

	stub := &stubQuoteWorkflowSource{
		tasksByID: map[string]*invoiceninja.NinjaTask{
			"task-1": {ID: "task-1", ClientID: "client-a", ProjectID: "project-a", Description: "Build"},
			"task-2": {ID: "task-2", ClientID: "client-a", ProjectID: "project-a", Description: "QA"},
		},
	}

	selected, err := enrichQuoteWithNinjaSelections(
		context.Background(),
		stub,
		quoteSpec,
		quoteSelectionInput{TaskIDs: []string{"task-1", "task-2"}, Interactive: false},
		bytes.NewBufferString(""),
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatalf("enrichQuoteWithNinjaSelections returned error: %v", err)
	}
	if len(selected) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(selected))
	}
	if quoteSpec.Client.ID != "client-a" {
		t.Fatalf("expected derived client id, got %q", quoteSpec.Client.ID)
	}
	if quoteSpec.Project.ID != "project-a" {
		t.Fatalf("expected derived project id, got %q", quoteSpec.Project.ID)
	}
}

func TestEnrichQuoteWithNinjaSelections_TaskIDsConflictAcrossProject(t *testing.T) {
	quoteSpec := newQuoteTemplate()
	quoteSpec.Client.ID = ""
	quoteSpec.Project.ID = ""

	stub := &stubQuoteWorkflowSource{
		tasksByID: map[string]*invoiceninja.NinjaTask{
			"task-1": {ID: "task-1", ClientID: "client-a", ProjectID: "project-a"},
			"task-2": {ID: "task-2", ClientID: "client-a", ProjectID: "project-b"},
		},
	}

	_, err := enrichQuoteWithNinjaSelections(
		context.Background(),
		stub,
		quoteSpec,
		quoteSelectionInput{TaskIDs: []string{"task-1", "task-2"}, Interactive: false},
		bytes.NewBufferString(""),
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatalf("expected conflict error")
	}
	if !strings.Contains(err.Error(), "span multiple project IDs") {
		t.Fatalf("unexpected conflict error: %v", err)
	}
}

func TestResolveSelectedTasks_PropagatesListError(t *testing.T) {
	stub := &stubQuoteWorkflowSource{listTasksErr: errors.New("tasks unavailable")}
	quoteSpec := newQuoteTemplate()
	quoteSpec.Project.ID = "project-1"

	_, err := resolveSelectedTasks(
		context.Background(),
		stub,
		quoteSpec,
		quoteSelectionInput{Interactive: true},
		bytes.NewBufferString(""),
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatalf("expected list task error")
	}
	if !strings.Contains(err.Error(), "failed to list tasks") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMaybeSelectClientProject_PropagatesListError(t *testing.T) {
	stub := &stubClientSource{listProjectsErr: errors.New("projects unavailable")}

	_, err := maybeSelectClientProject(context.Background(), stub, bytes.NewBufferString(""), &bytes.Buffer{}, &invoiceninja.NinjaClient{ID: "client-1"})
	if err == nil {
		t.Fatalf("expected list project error")
	}
	if !strings.Contains(err.Error(), "failed to list projects") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type stubClientSource struct {
	listClientsResult  []*invoiceninja.NinjaClient
	listProjectsResult []invoiceninja.NinjaProject
	getClientResult    *invoiceninja.NinjaClient
	findByEmailResult  *invoiceninja.NinjaClient
	findByNameResult   *invoiceninja.NinjaClient

	listClientsErr  error
	listProjectsErr error
	getClientErr    error
	findByEmailErr  error
	findByNameErr   error

	findByEmailCalls int
}

type stubQuoteWorkflowSource struct {
	project       *invoiceninja.NinjaProject
	projects      []invoiceninja.NinjaProject
	tasks         []invoiceninja.NinjaTask
	tasksByID     map[string]*invoiceninja.NinjaTask
	listClients   []invoiceninja.NinjaClient
	listTasksErr  error
	createdClient *invoiceninja.NinjaClient
}

func (s *stubQuoteWorkflowSource) ListClients(context.Context, int, int) (*invoiceninja.ClientListResponse, error) {
	return &invoiceninja.ClientListResponse{Data: append([]invoiceninja.NinjaClient(nil), s.listClients...)}, nil
}

func (s *stubQuoteWorkflowSource) ListProjects(context.Context, int, int) (*invoiceninja.ProjectListResponse, error) {
	return &invoiceninja.ProjectListResponse{Data: append([]invoiceninja.NinjaProject(nil), s.projects...)}, nil
}

func (s *stubQuoteWorkflowSource) GetClient(context.Context, string) (*invoiceninja.NinjaClient, error) {
	return nil, nil
}

func (s *stubQuoteWorkflowSource) FindClientByEmail(context.Context, string) (*invoiceninja.NinjaClient, error) {
	return nil, nil
}

func (s *stubQuoteWorkflowSource) FindClientByName(context.Context, string) (*invoiceninja.NinjaClient, error) {
	return nil, nil
}

func (s *stubQuoteWorkflowSource) ListTasks(context.Context, int, int) (*invoiceninja.TaskListResponse, error) {
	if s.listTasksErr != nil {
		return nil, s.listTasksErr
	}
	return &invoiceninja.TaskListResponse{Data: append([]invoiceninja.NinjaTask(nil), s.tasks...)}, nil
}

func (s *stubQuoteWorkflowSource) GetProject(context.Context, string) (*invoiceninja.NinjaProject, error) {
	return s.project, nil
}

func (s *stubQuoteWorkflowSource) GetTask(_ context.Context, id string) (*invoiceninja.NinjaTask, error) {
	if s.tasksByID == nil {
		return nil, nil
	}
	task, ok := s.tasksByID[id]
	if !ok {
		return nil, nil
	}
	return task, nil
}

func (s *stubQuoteWorkflowSource) CreateClient(_ context.Context, req invoiceninja.CreateClientRequest) (*invoiceninja.NinjaClient, error) {
	created := s.createdClient
	if created == nil {
		created = &invoiceninja.NinjaClient{ID: "created-client", Name: req.Name, Email: req.Email}
	}
	if strings.TrimSpace(created.Name) == "" {
		created.Name = req.Name
	}
	if strings.TrimSpace(created.Email) == "" {
		created.Email = req.Email
	}
	return created, nil
}

func (s *stubQuoteWorkflowSource) CreateProject(context.Context, invoiceninja.CreateProjectRequest) (*invoiceninja.NinjaProject, error) {
	return nil, nil
}

func (s *stubQuoteWorkflowSource) CreateTask(context.Context, invoiceninja.CreateTaskRequest) (*invoiceninja.NinjaTask, error) {
	return nil, nil
}

func (s *stubQuoteWorkflowSource) CreateQuote(context.Context, invoiceninja.CreateQuoteRequest) (*invoiceninja.NinjaQuote, error) {
	return nil, nil
}

func (s *stubQuoteWorkflowSource) ConvertQuoteToInvoice(context.Context, string) (*invoiceninja.NinjaInvoice, error) {
	return nil, nil
}

func (s *stubClientSource) ListClients(context.Context, int, int) (*invoiceninja.ClientListResponse, error) {
	if s.listClientsErr != nil {
		return nil, s.listClientsErr
	}

	list := make([]invoiceninja.NinjaClient, 0, len(s.listClientsResult))
	for _, c := range s.listClientsResult {
		if c != nil {
			list = append(list, *c)
		}
	}

	return &invoiceninja.ClientListResponse{Data: list}, nil
}

func (s *stubClientSource) ListProjects(context.Context, int, int) (*invoiceninja.ProjectListResponse, error) {
	if s.listProjectsErr != nil {
		return nil, s.listProjectsErr
	}

	list := make([]invoiceninja.NinjaProject, len(s.listProjectsResult))
	copy(list, s.listProjectsResult)

	return &invoiceninja.ProjectListResponse{Data: list}, nil
}

func (s *stubClientSource) GetClient(context.Context, string) (*invoiceninja.NinjaClient, error) {
	if s.getClientErr != nil {
		return nil, s.getClientErr
	}
	return s.getClientResult, nil
}

func (s *stubClientSource) FindClientByEmail(context.Context, string) (*invoiceninja.NinjaClient, error) {
	s.findByEmailCalls++
	if s.findByEmailErr != nil {
		return nil, s.findByEmailErr
	}
	return s.findByEmailResult, nil
}

func (s *stubClientSource) FindClientByName(context.Context, string) (*invoiceninja.NinjaClient, error) {
	if s.findByNameErr != nil {
		return nil, s.findByNameErr
	}
	return s.findByNameResult, nil
}
