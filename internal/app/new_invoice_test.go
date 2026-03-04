package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ninjops/ninjops/internal/config"
	"github.com/ninjops/ninjops/internal/invoiceninja"
)

func TestNewInvoiceTemplate_HasDefaults(t *testing.T) {
	quoteSpec := newInvoiceTemplate()

	if quoteSpec.Client.Name == "" {
		t.Fatal("expected default client name")
	}
	if quoteSpec.Client.Email == "" {
		t.Fatal("expected default client email")
	}
	if quoteSpec.Project.Name == "" {
		t.Fatal("expected default project name")
	}
	if quoteSpec.Pricing.Currency == "" {
		t.Fatal("expected default currency")
	}
}

func TestMapNinjaClientToQuote_NilClient(t *testing.T) {
	quoteSpec := newInvoiceTemplate()
	originalName := quoteSpec.Client.Name

	mapNinjaClientToQuote(quoteSpec, nil)

	if quoteSpec.Client.Name != originalName {
		t.Fatal("expected client name to remain unchanged with nil client")
	}
}

func TestMapNinjaClientToQuote_PopulatesFields(t *testing.T) {
	quoteSpec := newInvoiceTemplate()
	client := &invoiceninja.NinjaClient{
		ID:          "client-1",
		Name:        "Updated Client",
		DisplayName: "Display Name",
		Email:       "client@example.com",
		Contacts: []invoiceninja.ClientContact{
			{Email: "contact@example.com", IsPrimary: true},
		},
	}

	mapNinjaClientToQuote(quoteSpec, client)

	if quoteSpec.Client.ID != "client-1" {
		t.Fatalf("expected client ID client-1, got %s", quoteSpec.Client.ID)
	}
	if quoteSpec.Client.Name != "Display Name" {
		t.Fatalf("expected display name, got %s", quoteSpec.Client.Name)
	}
	if quoteSpec.Client.Email != "contact@example.com" {
		t.Fatalf("expected primary contact email, got %s", quoteSpec.Client.Email)
	}
}

func TestMapNinjaProjectToQuote_NilProject(t *testing.T) {
	quoteSpec := newInvoiceTemplate()
	originalName := quoteSpec.Project.Name

	mapNinjaProjectToQuote(quoteSpec, nil)

	if quoteSpec.Project.Name != originalName {
		t.Fatal("expected project name to remain unchanged with nil project")
	}
}

func TestMapNinjaProjectToQuote_PopulatesFields(t *testing.T) {
	quoteSpec := newInvoiceTemplate()
	project := &invoiceninja.NinjaProject{
		ID:          "project-1",
		Name:        "Updated Project",
		Description: "Updated Description",
	}

	mapNinjaProjectToQuote(quoteSpec, project)

	if quoteSpec.Project.ID != "project-1" {
		t.Fatalf("expected project ID project-1, got %s", quoteSpec.Project.ID)
	}
	if quoteSpec.Project.Name != "Updated Project" {
		t.Fatalf("expected updated name, got %s", quoteSpec.Project.Name)
	}
	if quoteSpec.Project.Description != "Updated Description" {
		t.Fatalf("expected updated description, got %s", quoteSpec.Project.Description)
	}
}

func TestFilterProjectsByClientID_FiltersCorrectly(t *testing.T) {
	projects := []invoiceninja.NinjaProject{
		{ID: "p1", ClientID: "c1"},
		{ID: "p2", ClientID: "c2"},
		{ID: "p3", ClientID: "c1"},
	}

	filtered := filterProjectsByClientID(projects, "c1")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(filtered))
	}
	if filtered[0].ID != "p1" || filtered[1].ID != "p3" {
		t.Fatalf("unexpected filtered projects: %v", filtered)
	}
}

func TestFilterProjectsByClientID_EmptyClientID(t *testing.T) {
	projects := []invoiceninja.NinjaProject{
		{ID: "p1", ClientID: "c1"},
		{ID: "p2", ClientID: "c2"},
	}

	filtered := filterProjectsByClientID(projects, "")
	if filtered != nil {
		t.Fatalf("expected nil when client ID is empty, got %d projects", len(filtered))
	}
}

func TestFilterTasksByProjectID_FiltersCorrectly(t *testing.T) {
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
		t.Fatalf("unexpected filtered tasks: %v", filtered)
	}
}

func TestExtractTaskIDs_ExtractsIDs(t *testing.T) {
	tasks := []invoiceninja.NinjaTask{
		{ID: "t1"},
		{ID: "t2"},
		{ID: ""},
	}

	ids := extractTaskIDs(tasks)
	if len(ids) != 2 {
		t.Fatalf("expected 2 IDs (empty ID excluded), got %d", len(ids))
	}
	if ids[0] != "t1" || ids[1] != "t2" {
		t.Fatalf("unexpected IDs: %v", ids)
	}
}

func TestPromptNumericChoice_ValidSelection(t *testing.T) {
	in := bytes.NewBufferString("2\n")
	out := &bytes.Buffer{}

	choice, err := promptNumericChoice(in, out, "Select", 0, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if choice != 2 {
		t.Fatalf("expected choice 2, got %d", choice)
	}
}

func TestPromptNumericChoice_EOF(t *testing.T) {
	in := bytes.NewBufferString("")
	out := &bytes.Buffer{}

	choice, err := promptNumericChoice(in, out, "Select", 0, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if choice != 0 {
		t.Fatalf("expected choice 0 on EOF, got %d", choice)
	}
}

func TestResolveClientByID_WithNilClient_ReturnsError(t *testing.T) {
	_, err := resolveClientByID(context.Background(), nil, "test-id")
	if err == nil {
		t.Fatal("expected error when client is nil")
	}
}

func TestResolveClientByID_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/clients/client-123" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(invoiceninja.ClientResponse{Data: invoiceninja.NinjaClient{ID: "client-123", Name: "Acme"}})
	}))
	defer server.Close()

	client := invoiceninja.NewClient(config.NinjaConfig{BaseURL: server.URL, APIToken: "token"})

	resolved, err := resolveClientByID(context.Background(), client, "client-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved == nil || resolved.ID != "client-123" {
		t.Fatalf("unexpected resolved client: %#v", resolved)
	}
}

func TestValidateNewInvoiceMode_AllowsEntitySelectors(t *testing.T) {
	err := validateNewInvoiceMode("", "", "client1", "", "", "", "task1")
	if err != nil {
		t.Fatalf("unexpected error when using entity selectors: %v", err)
	}

	err = validateNewInvoiceMode("", "", "", "email@test.com", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error when using client-email: %v", err)
	}

	err = validateNewInvoiceMode("", "", "", "", "Client Name", "", "")
	if err != nil {
		t.Fatalf("unexpected error when using client-name: %v", err)
	}

	err = validateNewInvoiceMode("", "", "", "", "", "project1", "")
	if err != nil {
		t.Fatalf("unexpected error when using project-id: %v", err)
	}
}

func TestValidateNewInvoiceMode_AllowsInteractiveMode(t *testing.T) {
	err := validateNewInvoiceMode("", "", "", "", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error when no flags provided (interactive mode): %v", err)
	}
}

func TestEnsureClientForInvoiceFromSpec_CreatesClient(t *testing.T) {
	quoteSpec := newInvoiceTemplate()
	quoteSpec.Client.Name = "Acme Inc"
	quoteSpec.Client.Email = "billing@acme.test"

	createdCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/clients" && r.URL.Query().Get("email") == quoteSpec.Client.Email:
			_ = json.NewEncoder(w).Encode(invoiceninja.ClientListResponse{Data: []invoiceninja.NinjaClient{}})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/clients" && r.URL.Query().Get("name") == quoteSpec.Client.Name:
			_ = json.NewEncoder(w).Encode(invoiceninja.ClientListResponse{Data: []invoiceninja.NinjaClient{}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/clients":
			createdCalled = true
			var req invoiceninja.CreateClientRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode create request: %v", err)
			}
			if req.Name != quoteSpec.Client.Name || req.Email != quoteSpec.Client.Email {
				t.Fatalf("unexpected create request payload: %#v", req)
			}
			_ = json.NewEncoder(w).Encode(invoiceninja.ClientResponse{Data: invoiceninja.NinjaClient{ID: "client-1", Name: req.Name, Email: req.Email}})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := invoiceninja.NewClient(config.NinjaConfig{BaseURL: server.URL, APIToken: "token"})
	resolved, created, err := ensureClientForInvoiceFromSpec(context.Background(), client, quoteSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected client to be created")
	}
	if !createdCalled {
		t.Fatal("expected create endpoint to be called")
	}
	if resolved == nil || resolved.ID != "client-1" {
		t.Fatalf("unexpected resolved client: %#v", resolved)
	}
}

func TestEnsureClientForInvoiceFromSpec_ReusesClient(t *testing.T) {
	quoteSpec := newInvoiceTemplate()
	quoteSpec.Client.Name = "Acme Inc"
	quoteSpec.Client.Email = "billing@acme.test"

	updatedCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/clients" && r.URL.Query().Get("email") == quoteSpec.Client.Email:
			_ = json.NewEncoder(w).Encode(invoiceninja.ClientListResponse{Data: []invoiceninja.NinjaClient{{ID: "existing-1", Name: "Acme Inc", Email: quoteSpec.Client.Email}}})
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/clients/existing-1":
			updatedCalled = true
			var req invoiceninja.UpdateClientRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode update request: %v", err)
			}
			if req.Name != quoteSpec.Client.Name || req.Email != quoteSpec.Client.Email {
				t.Fatalf("unexpected update request payload: %#v", req)
			}
			_ = json.NewEncoder(w).Encode(invoiceninja.ClientResponse{Data: invoiceninja.NinjaClient{ID: "existing-1", Name: req.Name, Email: req.Email}})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	client := invoiceninja.NewClient(config.NinjaConfig{BaseURL: server.URL, APIToken: "token"})
	resolved, created, err := ensureClientForInvoiceFromSpec(context.Background(), client, quoteSpec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created {
		t.Fatal("expected existing client to be reused")
	}
	if !updatedCalled {
		t.Fatal("expected update endpoint to be called")
	}
	if resolved == nil || resolved.ID != "existing-1" {
		t.Fatalf("unexpected resolved client: %#v", resolved)
	}
}
