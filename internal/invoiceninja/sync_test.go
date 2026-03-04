package invoiceninja

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ninjops/ninjops/internal/config"
	"github.com/ninjops/ninjops/internal/spec"
	"github.com/ninjops/ninjops/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCreateProjectRequest(t *testing.T) {
	quoteSpec := &spec.QuoteSpec{
		Project: spec.ProjectInfo{
			Name:        "Test Project",
			Description: "Test Description",
			Deadline:    "2024-12-31",
		},
	}

	req := BuildCreateProjectRequest(quoteSpec, "client-1")

	assert.Equal(t, "client-1", req.ClientID)
	assert.Equal(t, "Test Project", req.Name)
	assert.Equal(t, "Test Description", req.Description)
	assert.Equal(t, "2024-12-31", req.DueDate)
}

func TestBuildUpdateProjectRequest(t *testing.T) {
	quoteSpec := &spec.QuoteSpec{
		Project: spec.ProjectInfo{
			Name:        "Updated Project",
			Description: "Updated Description",
			Deadline:    "2025-01-31",
		},
	}

	existing := &NinjaProject{
		ID:          "project-1",
		ClientID:    "client-1",
		Name:        "Old Project",
		Description: "Old Description",
		DueDate:     "2024-06-30",
	}

	req := BuildUpdateProjectRequest(quoteSpec, existing)

	assert.Equal(t, "project-1", req.ID)
	assert.Equal(t, "client-1", req.ClientID)
	assert.Equal(t, "Updated Project", req.Name)
	assert.Equal(t, "Updated Description", req.Description)
	assert.Equal(t, "2025-01-31", req.DueDate)
}

func TestBuildUpdateProjectRequest_KeepsExistingWhenEmpty(t *testing.T) {
	quoteSpec := &spec.QuoteSpec{
		Project: spec.ProjectInfo{
			Name:        "",
			Description: "",
			Deadline:    "",
		},
	}

	existing := &NinjaProject{
		ID:          "project-1",
		ClientID:    "client-1",
		Name:        "Existing Project",
		Description: "Existing Description",
		DueDate:     "2024-06-30",
	}

	req := BuildUpdateProjectRequest(quoteSpec, existing)

	assert.Equal(t, "Existing Project", req.Name)
	assert.Equal(t, "Existing Description", req.Description)
	assert.Equal(t, "2024-06-30", req.DueDate)
}

func TestSyncer_ComputeProjectDiffs_DetectsChanges(t *testing.T) {
	syncer := &Syncer{}

	existing := &NinjaProject{
		ID:          "project-1",
		Name:        "Old Name",
		Description: "Old Description",
	}

	quoteSpec := &spec.QuoteSpec{
		Project: spec.ProjectInfo{
			Name:        "New Name",
			Description: "New Description",
		},
	}

	diffs := syncer.computeProjectDiffs(existing, quoteSpec)

	assert.Len(t, diffs, 2)
}

func TestSyncer_ComputeProjectDiffs_NoChanges(t *testing.T) {
	syncer := &Syncer{}

	existing := &NinjaProject{
		ID:          "project-1",
		Name:        "Same Name",
		Description: "Same Description",
	}

	quoteSpec := &spec.QuoteSpec{
		Project: spec.ProjectInfo{
			Name:        "Same Name",
			Description: "Same Description",
		},
	}

	diffs := syncer.computeProjectDiffs(existing, quoteSpec)

	assert.Len(t, diffs, 0)
}

func TestSyncer_ComputeQuoteDiffs(t *testing.T) {
	syncer := &Syncer{}

	existing := &NinjaQuote{
		ID:          "quote-1",
		PublicNotes: "Old notes",
		Terms:       "Old terms",
	}

	artifacts := &spec.GeneratedArtifacts{
		PublicNotesText: "New notes",
		TermsMarkdown:   "Old terms",
	}

	diffs := syncer.computeQuoteDiffs(existing, artifacts)

	assert.Len(t, diffs, 1)
	assert.Equal(t, "public_notes", diffs[0].Field)
}

func TestSyncer_ComputeInvoiceDiffs(t *testing.T) {
	syncer := &Syncer{}

	existing := &NinjaInvoice{
		ID:          "invoice-1",
		PublicNotes: "Old notes",
		Terms:       "Old terms",
	}

	artifacts := &spec.GeneratedArtifacts{
		PublicNotesText: "New notes",
		TermsMarkdown:   "New terms",
	}

	diffs := syncer.computeInvoiceDiffs(existing, artifacts)

	assert.Len(t, diffs, 2)
}

func TestSyncer_SyncProject_CreatesNew(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			// FindProjectByName returns nothing
			json.NewEncoder(w).Encode(ProjectListResponse{
				Data: []NinjaProject{},
			})
			return
		}

		if r.Method == http.MethodPost {
			// CreateProject
			var req CreateProjectRequest
			json.NewDecoder(r.Body).Decode(&req)
			json.NewEncoder(w).Encode(ProjectResponse{
				Data: NinjaProject{
					ID:          "new-project-1",
					ClientID:    req.ClientID,
					Name:        req.Name,
					Description: req.Description,
				},
			})
		}
	}))
	defer server.Close()

	client := NewClient(config.NinjaConfig{BaseURL: server.URL, APIToken: "token"})
	tmpDir := t.TempDir()
	st, err := store.NewStore(tmpDir)
	require.NoError(t, err)

	syncer := NewSyncer(client, st)

	quoteSpec := &spec.QuoteSpec{
		Project: spec.ProjectInfo{
			Name:        "New Project",
			Description: "A new project",
		},
	}

	opts := SyncOptions{}

	projectID, created, updated, err := syncer.syncProject(context.Background(), quoteSpec, "client-1", nil, opts)

	require.NoError(t, err)
	assert.Equal(t, "new-project-1", projectID)
	assert.True(t, created)
	assert.False(t, updated)
}

func TestSyncer_SyncProject_UpdatesExisting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			// FindProjectByName returns existing project
			json.NewEncoder(w).Encode(ProjectListResponse{
				Data: []NinjaProject{
					{ID: "existing-project-1", ClientID: "client-1", Name: "Existing Project", Description: "Old desc"},
				},
			})
			return
		}

		if r.Method == http.MethodPut {
			// UpdateProject
			var req UpdateProjectRequest
			json.NewDecoder(r.Body).Decode(&req)
			json.NewEncoder(w).Encode(ProjectResponse{
				Data: NinjaProject{
					ID:          req.ID,
					ClientID:    req.ClientID,
					Name:        req.Name,
					Description: req.Description,
				},
			})
		}
	}))
	defer server.Close()

	client := NewClient(config.NinjaConfig{BaseURL: server.URL, APIToken: "token"})
	tmpDir := t.TempDir()
	st, err := store.NewStore(tmpDir)
	require.NoError(t, err)

	syncer := NewSyncer(client, st)

	quoteSpec := &spec.QuoteSpec{
		Project: spec.ProjectInfo{
			Name:        "Existing Project",
			Description: "Updated description",
		},
	}

	opts := SyncOptions{}

	projectID, created, updated, err := syncer.syncProject(context.Background(), quoteSpec, "client-1", nil, opts)

	require.NoError(t, err)
	assert.Equal(t, "existing-project-1", projectID)
	assert.False(t, created)
	assert.True(t, updated)
}

func TestSyncer_SyncProject_SkipsWhenNoName(t *testing.T) {
	syncer := &Syncer{}

	quoteSpec := &spec.QuoteSpec{
		Project: spec.ProjectInfo{
			Name: "",
		},
	}

	opts := SyncOptions{}

	projectID, created, updated, err := syncer.syncProject(context.Background(), quoteSpec, "client-1", nil, opts)

	require.NoError(t, err)
	assert.Empty(t, projectID)
	assert.False(t, created)
	assert.False(t, updated)
}

func TestSyncer_SyncProject_DryRun(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Only allow GET requests (FindProjectByName) - no POST/PUT
		if r.Method != http.MethodGet {
			t.Fatalf("should not make %s requests in dry run mode", r.Method)
		}

		// Return no existing project
		json.NewEncoder(w).Encode(ProjectListResponse{
			Data: []NinjaProject{},
		})
	}))
	defer server.Close()

	client := NewClient(config.NinjaConfig{BaseURL: server.URL, APIToken: "token"})
	tmpDir := t.TempDir()
	st, err := store.NewStore(tmpDir)
	require.NoError(t, err)

	syncer := NewSyncer(client, st)

	quoteSpec := &spec.QuoteSpec{
		Project: spec.ProjectInfo{
			Name:        "New Project",
			Description: "A new project",
		},
	}

	opts := SyncOptions{DryRun: true}

	projectID, created, updated, err := syncer.syncProject(context.Background(), quoteSpec, "client-1", nil, opts)

	require.NoError(t, err)
	assert.Empty(t, projectID)
	assert.True(t, created) // Dry run reports what it WOULD create
	assert.False(t, updated)
}
