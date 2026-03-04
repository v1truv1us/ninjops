package app

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ninjops/ninjops/internal/config"
	"github.com/ninjops/ninjops/internal/invoiceninja"
)

func TestListCmd_RejectsUnsupportedEntity(t *testing.T) {
	cmd := newListCmd()
	cmd.SetArgs([]string{"widgets"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "unsupported entity \"widgets\""; !strings.Contains(got, want) {
		t.Fatalf("unexpected error %q, want substring %q", got, want)
	}
}

func TestListCmd_RejectsUnsupportedFormat(t *testing.T) {
	cmd := newListCmd()
	cmd.SetArgs([]string{"clients", "--format", "xml"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "unsupported --format \"xml\""; !strings.Contains(got, want) {
		t.Fatalf("unexpected error %q, want substring %q", got, want)
	}
}

func TestListCmd_RejectsUnsupportedFilterForClients(t *testing.T) {
	cmd := newListCmd()
	cmd.SetArgs([]string{"clients", "--client-id", "c1"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "not supported for list clients"; !strings.Contains(got, want) {
		t.Fatalf("unexpected error %q, want substring %q", got, want)
	}
}

func TestListCmd_MissingTokenReturnsClearError(t *testing.T) {
	prevCfg := cfg
	cfg = &config.Config{Ninja: config.NinjaConfig{}}
	t.Cleanup(func() {
		cfg = prevCfg
	})

	cmd := newListCmd()
	cmd.SetArgs([]string{"projects"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "Invoice Ninja API token not configured"; got != want {
		t.Fatalf("unexpected error %q, want %q", got, want)
	}
}

func TestShowCmd_RejectsUnsupportedEntity(t *testing.T) {
	cmd := newShowCmd()
	cmd.SetArgs([]string{"widgets", "123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "unsupported entity \"widgets\""; !strings.Contains(got, want) {
		t.Fatalf("unexpected error %q, want substring %q", got, want)
	}
}

func TestShowCmd_MissingTokenReturnsClearError(t *testing.T) {
	prevCfg := cfg
	cfg = &config.Config{Ninja: config.NinjaConfig{}}
	t.Cleanup(func() {
		cfg = prevCfg
	})

	cmd := newShowCmd()
	cmd.SetArgs([]string{"client", "c1"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "Invoice Ninja API token not configured"; got != want {
		t.Fatalf("unexpected error %q, want %q", got, want)
	}
}

func TestFilterTasksForList_FiltersByClientAndProject(t *testing.T) {
	tasks := []invoiceninja.NinjaTask{
		{ID: "t1", ClientID: "c1", ProjectID: "p1"},
		{ID: "t2", ClientID: "c1", ProjectID: "p2"},
		{ID: "t3", ClientID: "c2", ProjectID: "p1"},
	}

	filtered := filterTasksForList(tasks, "c1", "p2")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 task, got %d", len(filtered))
	}
	if filtered[0].ID != "t2" {
		t.Fatalf("unexpected task selected: %+v", filtered[0])
	}
}

func TestRenderClients_TableOutput(t *testing.T) {
	clients := []invoiceninja.NinjaClient{{ID: "c1", Name: "Acme", Email: "billing@acme.test"}}

	var out bytes.Buffer
	if err := renderClients(&out, clients, "table"); err != nil {
		t.Fatalf("renderClients returned error: %v", err)
	}

	rendered := out.String()
	if !strings.Contains(rendered, "ID") || !strings.Contains(rendered, "NAME") {
		t.Fatalf("missing table header in output: %s", rendered)
	}
	if !strings.Contains(rendered, "c1") || !strings.Contains(rendered, "Acme") {
		t.Fatalf("missing expected row data in output: %s", rendered)
	}
}

func TestRenderQuotes_SimpleOutput(t *testing.T) {
	quotes := []invoiceninja.NinjaQuote{{ID: "q1", Number: "Q-100", Amount: 1250.5}}

	var out bytes.Buffer
	if err := renderQuotes(&out, quotes, "simple"); err != nil {
		t.Fatalf("renderQuotes returned error: %v", err)
	}

	if got, want := out.String(), "q1\t#Q-100\t1250.50\n"; got != want {
		t.Fatalf("unexpected simple output %q, want %q", got, want)
	}
}

func TestRunListEntity_UsesFindHelpersForClientFilters(t *testing.T) {
	stub := &stubListEntitySource{
		projectsByClient: []invoiceninja.NinjaProject{{ID: "p1", ClientID: "c1", Name: "Project One"}},
		quotesByClient:   []invoiceninja.NinjaQuote{{ID: "q1", ClientID: "c1", Number: "Q-1"}},
		invoicesByClient: []invoiceninja.NinjaInvoice{{ID: "i1", ClientID: "c1", Number: "I-1"}},
	}

	for _, tc := range []struct {
		entity string
		label  string
	}{
		{entity: "projects", label: "Project One"},
		{entity: "quotes", label: "q1"},
		{entity: "invoices", label: "i1"},
	} {
		var out bytes.Buffer
		if err := runListEntity(context.Background(), &out, stub, tc.entity, "c1", "", 5, "simple"); err != nil {
			t.Fatalf("runListEntity(%s) returned error: %v", tc.entity, err)
		}
		if !strings.Contains(out.String(), tc.label) {
			t.Fatalf("expected output to contain %q for %s, got %q", tc.label, tc.entity, out.String())
		}
	}

	if stub.findProjectsByClientCalls == 0 || stub.findQuotesByClientCalls == 0 || stub.findInvoicesByClientCalls == 0 {
		t.Fatalf("expected find helpers to be used for client filters, calls=%+v", stub)
	}
	if stub.listProjectsCalls != 0 || stub.listQuotesCalls != 0 || stub.listInvoicesCalls != 0 {
		t.Fatalf("did not expect unfiltered list calls when client filter is set, calls=%+v", stub)
	}
}

func TestRunListEntity_TasksByClientPaginatesToLimit(t *testing.T) {
	stub := &stubListEntitySource{
		tasksByClientPages: map[int][]invoiceninja.NinjaTask{
			1: {{ID: "t1", ClientID: "c1", Description: "Task 1"}},
			2: {{ID: "t2", ClientID: "c1", Description: "Task 2"}},
			3: {{ID: "t3", ClientID: "c1", Description: "Task 3"}},
		},
		tasksTotalPages: 3,
	}

	var out bytes.Buffer
	if err := runListEntity(context.Background(), &out, stub, "tasks", "c1", "", 3, "simple"); err != nil {
		t.Fatalf("runListEntity returned error: %v", err)
	}

	rendered := out.String()
	if !strings.Contains(rendered, "t1") || !strings.Contains(rendered, "t2") || !strings.Contains(rendered, "t3") {
		t.Fatalf("expected paginated task output, got %q", rendered)
	}
	if stub.findTasksByClientCalls != 3 {
		t.Fatalf("expected 3 paginated FindTasksByClient calls, got %d", stub.findTasksByClientCalls)
	}
}

type stubListEntitySource struct {
	clients          []invoiceninja.NinjaClient
	projects         []invoiceninja.NinjaProject
	projectsByClient []invoiceninja.NinjaProject
	tasks            []invoiceninja.NinjaTask
	tasksByProject   []invoiceninja.NinjaTask
	quotes           []invoiceninja.NinjaQuote
	quotesByClient   []invoiceninja.NinjaQuote
	invoices         []invoiceninja.NinjaInvoice
	invoicesByClient []invoiceninja.NinjaInvoice

	tasksByClientPages map[int][]invoiceninja.NinjaTask
	tasksTotalPages    int

	listProjectsCalls         int
	findProjectsByClientCalls int
	listTasksCalls            int
	findTasksByProjectCalls   int
	findTasksByClientCalls    int
	listQuotesCalls           int
	findQuotesByClientCalls   int
	listInvoicesCalls         int
	findInvoicesByClientCalls int
}

func (s *stubListEntitySource) ListClients(context.Context, int, int) (*invoiceninja.ClientListResponse, error) {
	return &invoiceninja.ClientListResponse{Data: append([]invoiceninja.NinjaClient(nil), s.clients...)}, nil
}

func (s *stubListEntitySource) ListProjects(context.Context, int, int) (*invoiceninja.ProjectListResponse, error) {
	s.listProjectsCalls++
	return &invoiceninja.ProjectListResponse{Data: append([]invoiceninja.NinjaProject(nil), s.projects...)}, nil
}

func (s *stubListEntitySource) FindProjectsByClient(context.Context, string, int, int) (*invoiceninja.ProjectListResponse, error) {
	s.findProjectsByClientCalls++
	return &invoiceninja.ProjectListResponse{Data: append([]invoiceninja.NinjaProject(nil), s.projectsByClient...)}, nil
}

func (s *stubListEntitySource) ListTasks(context.Context, int, int) (*invoiceninja.TaskListResponse, error) {
	s.listTasksCalls++
	return &invoiceninja.TaskListResponse{Data: append([]invoiceninja.NinjaTask(nil), s.tasks...)}, nil
}

func (s *stubListEntitySource) FindTasksByProject(context.Context, string, int, int) (*invoiceninja.TaskListResponse, error) {
	s.findTasksByProjectCalls++
	return &invoiceninja.TaskListResponse{Data: append([]invoiceninja.NinjaTask(nil), s.tasksByProject...)}, nil
}

func (s *stubListEntitySource) FindTasksByClient(_ context.Context, _ string, page, _ int) (*invoiceninja.TaskListResponse, error) {
	s.findTasksByClientCalls++
	rows := s.tasksByClientPages[page]
	return &invoiceninja.TaskListResponse{
		Data: append([]invoiceninja.NinjaTask(nil), rows...),
		Meta: invoiceninja.APIMeta{Pagination: invoiceninja.APIPagination{TotalPages: s.tasksTotalPages}},
	}, nil
}

func (s *stubListEntitySource) ListQuotes(context.Context, int, int) (*invoiceninja.QuoteListResponse, error) {
	s.listQuotesCalls++
	return &invoiceninja.QuoteListResponse{Data: append([]invoiceninja.NinjaQuote(nil), s.quotes...)}, nil
}

func (s *stubListEntitySource) FindQuotesByClient(context.Context, string, int, int) (*invoiceninja.QuoteListResponse, error) {
	s.findQuotesByClientCalls++
	return &invoiceninja.QuoteListResponse{Data: append([]invoiceninja.NinjaQuote(nil), s.quotesByClient...)}, nil
}

func (s *stubListEntitySource) ListInvoices(context.Context, int, int) (*invoiceninja.InvoiceListResponse, error) {
	s.listInvoicesCalls++
	return &invoiceninja.InvoiceListResponse{Data: append([]invoiceninja.NinjaInvoice(nil), s.invoices...)}, nil
}

func (s *stubListEntitySource) FindInvoicesByClient(context.Context, string, int, int) (*invoiceninja.InvoiceListResponse, error) {
	s.findInvoicesByClientCalls++
	return &invoiceninja.InvoiceListResponse{Data: append([]invoiceninja.NinjaInvoice(nil), s.invoicesByClient...)}, nil
}
