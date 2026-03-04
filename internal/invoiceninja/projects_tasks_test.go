package invoiceninja

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ninjops/ninjops/internal/config"
	"github.com/ninjops/ninjops/internal/spec"
	"github.com/stretchr/testify/assert"
)

func TestFilterProjectsByClient(t *testing.T) {
	projects := []NinjaProject{
		{ID: "p1", ClientID: "c1", Name: "Website"},
		{ID: "p2", ClientID: "c2", Name: "Mobile"},
		{ID: "p3", ClientID: "c1", Name: "API"},
	}

	filtered := filterProjectsByClient(projects, "c1")
	assert.Len(t, filtered, 2)
	assert.Equal(t, "p1", filtered[0].ID)
	assert.Equal(t, "p3", filtered[1].ID)
}

func TestFindProjectByName(t *testing.T) {
	projects := []NinjaProject{
		{ID: "p1", Name: "Website Revamp"},
		{ID: "p2", Name: "API Platform"},
	}

	match := findProjectByName(projects, "  website revamp  ")
	if assert.NotNil(t, match) {
		assert.Equal(t, "p1", match.ID)
	}

	assert.Nil(t, findProjectByName(projects, "missing"))
}

func TestFilterTasksByProject(t *testing.T) {
	tasks := []NinjaTask{
		{ID: "t1", ProjectID: "p1", Description: "Setup"},
		{ID: "t2", ProjectID: "p2", Description: "Deploy"},
		{ID: "t3", ProjectID: "p1", Description: "QA"},
	}

	filtered := filterTasksByProject(tasks, "p1")
	assert.Len(t, filtered, 2)
	assert.Equal(t, "t1", filtered[0].ID)
	assert.Equal(t, "t3", filtered[1].ID)
}

func TestFindTaskByDescription(t *testing.T) {
	tasks := []NinjaTask{
		{ID: "t1", Description: "Discovery"},
		{ID: "t2", Description: "Implementation"},
	}

	match := findTaskByDescription(tasks, " implementation ")
	if assert.NotNil(t, match) {
		assert.Equal(t, "t2", match.ID)
	}

	assert.Nil(t, findTaskByDescription(tasks, "none"))
}

func TestFilterTasksByClient(t *testing.T) {
	tasks := []NinjaTask{
		{ID: "t1", ClientID: "c1", ProjectID: "p1"},
		{ID: "t2", ClientID: "c2", ProjectID: "p1"},
		{ID: "t3", ClientID: "c1", ProjectID: "p2"},
	}

	filtered := filterTasksByClient(tasks, "c1")
	assert.Len(t, filtered, 2)
	assert.Equal(t, "t1", filtered[0].ID)
	assert.Equal(t, "t3", filtered[1].ID)
}

func TestFindTasksByClient_PassesFilterQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/tasks", r.URL.Path)
		assert.Equal(t, "client-1", r.URL.Query().Get("client_id"))
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "3", r.URL.Query().Get("per_page"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TaskListResponse{
			Data: []NinjaTask{
				{ID: "task-1", ClientID: "client-1"},
				{ID: "task-2", ClientID: "client-2"},
			},
			Meta: APIMeta{Pagination: APIPagination{Count: 2, TotalPages: 5}},
		})
	}))
	defer server.Close()

	client := NewClient(config.NinjaConfig{BaseURL: server.URL, APIToken: "token"})
	resp, err := client.FindTasksByClient(context.Background(), "client-1", 2, 3)
	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Len(t, resp.Data, 2)
		assert.Equal(t, "task-1", resp.Data[0].ID)
		assert.Equal(t, 2, resp.Meta.Pagination.Count)
	}
}

func TestBuildCreateQuoteRequest_SetsProjectID(t *testing.T) {
	quoteSpec := spec.NewQuoteSpec()
	quoteSpec.Project.ID = "project-1"

	artifacts := &spec.GeneratedArtifacts{
		PublicNotesText: "public notes",
		TermsMarkdown:   "terms",
		Meta: spec.GenMeta{
			Hash: "hash-1",
		},
	}

	req := BuildCreateQuoteRequest(quoteSpec, "client-1", artifacts)
	assert.Equal(t, "project-1", req.ProjectID)
}

func TestBuildUpdateQuoteRequest_ProjectIDFallback(t *testing.T) {
	quoteSpec := spec.NewQuoteSpec()
	quoteSpec.Project.ID = ""
	quoteSpec.Metadata.UpdatedAt = time.Now().UTC()

	artifacts := &spec.GeneratedArtifacts{
		PublicNotesText: "public notes",
		TermsMarkdown:   "terms",
	}

	existing := &NinjaQuote{
		ID:        "q1",
		ClientID:  "client-1",
		ProjectID: "project-existing",
	}

	req := BuildUpdateQuoteRequest(quoteSpec, existing, artifacts)
	assert.Equal(t, "project-existing", req.ProjectID)
}

func TestBuildCreateInvoiceRequest_SetsProjectID(t *testing.T) {
	quoteSpec := spec.NewQuoteSpec()
	quoteSpec.Project.ID = "project-2"

	artifacts := &spec.GeneratedArtifacts{
		PublicNotesText: "public notes",
		TermsMarkdown:   "terms",
		Meta: spec.GenMeta{
			Hash: "hash-2",
		},
	}

	req := BuildCreateInvoiceRequest(quoteSpec, "client-1", artifacts)
	assert.Equal(t, "project-2", req.ProjectID)
}

func TestBuildUpdateInvoiceRequest_ProjectIDFallback(t *testing.T) {
	quoteSpec := spec.NewQuoteSpec()
	quoteSpec.Project.ID = ""
	quoteSpec.Metadata.UpdatedAt = time.Now().UTC()

	artifacts := &spec.GeneratedArtifacts{
		PublicNotesText: "public notes",
		TermsMarkdown:   "terms",
	}

	existing := &NinjaInvoice{
		ID:        "i1",
		ClientID:  "client-1",
		ProjectID: "project-existing",
	}

	req := BuildUpdateInvoiceRequest(quoteSpec, existing, artifacts)
	assert.Equal(t, "project-existing", req.ProjectID)
}
