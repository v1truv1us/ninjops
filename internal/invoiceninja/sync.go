package invoiceninja

import (
	"context"
	"fmt"
	"strings"

	"github.com/ninjops/ninjops/internal/diff"
	"github.com/ninjops/ninjops/internal/spec"
	"github.com/ninjops/ninjops/internal/store"
)

type SyncMode string

const (
	SyncModeQuote   SyncMode = "quote"
	SyncModeInvoice SyncMode = "invoice"
	SyncModeBoth    SyncMode = "both"
)

type SyncOptions struct {
	Mode       SyncMode
	DryRun     bool
	ShowDiff   bool
	Confirm    bool
	AllowFuzzy bool
	QuoteID    string
	InvoiceID  string
}

type SyncResult struct {
	ClientID       string
	ClientCreated  bool
	ProjectID      string
	ProjectCreated bool
	ProjectUpdated bool
	QuoteID        string
	QuoteCreated   bool
	QuoteUpdated   bool
	InvoiceID      string
	InvoiceCreated bool
	InvoiceUpdated bool
	Diffs          []diff.FieldDiff
}

type Syncer struct {
	client *Client
	store  *store.Store
}

func NewSyncer(client *Client, store *store.Store) *Syncer {
	return &Syncer{
		client: client,
		store:  store,
	}
}

func (s *Syncer) Sync(ctx context.Context, quoteSpec *spec.QuoteSpec, artifacts *spec.GeneratedArtifacts, opts SyncOptions) (*SyncResult, error) {
	result := &SyncResult{}

	clientID, created, err := s.ensureClient(ctx, quoteSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure client: %w", err)
	}
	result.ClientID = clientID
	result.ClientCreated = created

	refID, err := spec.ExtractReferenceID(quoteSpec.Metadata.Reference)
	if err != nil {
		return nil, fmt.Errorf("invalid reference: %w", err)
	}

	entry, _ := s.store.GetEntry(refID)

	projectID, projectCreated, projectUpdated, err := s.syncProject(ctx, quoteSpec, clientID, entry, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to sync project: %w", err)
	}
	result.ProjectID = projectID
	result.ProjectCreated = projectCreated
	result.ProjectUpdated = projectUpdated

	if projectID != "" && quoteSpec.Project.ID == "" {
		quoteSpec.Project.ID = projectID
	}

	if projectID != "" {
		_ = s.store.UpdateProjectID(refID, projectID)
	}

	if opts.Mode == SyncModeQuote || opts.Mode == SyncModeBoth {
		quoteID, quoteCreated, quoteUpdated, diffs, err := s.syncQuote(ctx, quoteSpec, artifacts, clientID, entry, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to sync quote: %w", err)
		}
		result.QuoteID = quoteID
		result.QuoteCreated = quoteCreated
		result.QuoteUpdated = quoteUpdated
		result.Diffs = diffs

		if quoteID != "" {
			_ = s.store.UpdateQuoteID(refID, quoteID)
		}
	}

	if opts.Mode == SyncModeInvoice || opts.Mode == SyncModeBoth {
		invoiceID, invoiceCreated, invoiceUpdated, diffs, err := s.syncInvoice(ctx, quoteSpec, artifacts, clientID, entry, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to sync invoice: %w", err)
		}
		result.InvoiceID = invoiceID
		result.InvoiceCreated = invoiceCreated
		result.InvoiceUpdated = invoiceUpdated
		if len(result.Diffs) == 0 {
			result.Diffs = diffs
		}

		if invoiceID != "" {
			_ = s.store.UpdateInvoiceID(refID, invoiceID)
		}
	}

	_ = s.store.UpdateClientID(refID, clientID)
	_ = s.store.UpdateSyncHash(refID, artifacts.Meta.Hash)

	return result, nil
}

func (s *Syncer) ensureClient(ctx context.Context, quoteSpec *spec.QuoteSpec) (string, bool, error) {
	req := CreateClientRequest{
		Name:       quoteSpec.Client.Name,
		Email:      quoteSpec.Client.Email,
		Phone:      quoteSpec.Client.Phone,
		Address1:   quoteSpec.Client.Address.Line1,
		Address2:   quoteSpec.Client.Address.Line2,
		City:       quoteSpec.Client.Address.City,
		State:      quoteSpec.Client.Address.State,
		PostalCode: quoteSpec.Client.Address.PostalCode,
	}

	client, created, err := s.client.UpsertClient(ctx, req)
	if err != nil {
		return "", false, err
	}

	return client.ID, created, nil
}

func (s *Syncer) syncProject(ctx context.Context, quoteSpec *spec.QuoteSpec, clientID string, entry *store.StateEntry, opts SyncOptions) (string, bool, bool, error) {
	if strings.TrimSpace(quoteSpec.Project.Name) == "" {
		return "", false, false, nil
	}

	var existingProject *NinjaProject
	var err error

	if entry != nil && entry.ProjectID != "" {
		existingProject, err = s.client.GetProject(ctx, entry.ProjectID)
		if err != nil {
			existingProject = nil
		}
	}

	if existingProject == nil {
		existingProject, err = s.client.FindProjectByName(ctx, clientID, quoteSpec.Project.Name)
		if err != nil {
			return "", false, false, err
		}
	}

	if existingProject != nil {
		diffs := s.computeProjectDiffs(existingProject, quoteSpec)

		if opts.DryRun {
			return existingProject.ID, false, false, nil
		}

		if len(diffs) == 0 {
			return existingProject.ID, false, false, nil
		}

		updated, err := s.client.UpdateProject(ctx, existingProject.ID, BuildUpdateProjectRequest(quoteSpec, existingProject))
		if err != nil {
			return "", false, false, err
		}

		return updated.ID, false, true, nil
	}

	if opts.DryRun {
		return "", true, false, nil
	}

	newProject, err := s.client.CreateProject(ctx, BuildCreateProjectRequest(quoteSpec, clientID))
	if err != nil {
		return "", false, false, err
	}

	return newProject.ID, true, false, nil
}

func (s *Syncer) computeProjectDiffs(existing *NinjaProject, quoteSpec *spec.QuoteSpec) []diff.FieldDiff {
	oldFields := map[string]string{
		"name":        existing.Name,
		"description": existing.Description,
	}

	newFields := map[string]string{
		"name":        quoteSpec.Project.Name,
		"description": quoteSpec.Project.Description,
	}

	return diff.ComputeFields(oldFields, newFields)
}

func (s *Syncer) syncQuote(ctx context.Context, quoteSpec *spec.QuoteSpec, artifacts *spec.GeneratedArtifacts, clientID string, entry *store.StateEntry, opts SyncOptions) (string, bool, bool, []diff.FieldDiff, error) {
	var existingQuote *NinjaQuote
	var err error

	if opts.QuoteID != "" {
		existingQuote, err = s.client.GetQuote(ctx, opts.QuoteID)
		if err != nil {
			return "", false, false, nil, err
		}
	} else if entry != nil && entry.QuoteID != "" {
		existingQuote, err = s.client.GetQuote(ctx, entry.QuoteID)
		if err != nil {
			existingQuote = nil
		}
	}

	if existingQuote == nil {
		existingQuote, err = s.client.FindQuoteByReference(ctx, quoteSpec.Metadata.Reference)
		if err != nil {
			return "", false, false, nil, err
		}
	}

	if existingQuote != nil {
		diffs := s.computeQuoteDiffs(existingQuote, artifacts)

		if opts.DryRun {
			return existingQuote.ID, false, false, diffs, nil
		}

		if opts.ShowDiff && len(diffs) > 0 {
			fmt.Println(diff.FormatFieldDiffs(diffs))
		}

		if len(diffs) == 0 {
			return existingQuote.ID, false, false, nil, nil
		}

		updated, err := s.client.UpdateQuote(ctx, existingQuote.ID, BuildUpdateQuoteRequest(quoteSpec, existingQuote, artifacts))
		if err != nil {
			return "", false, false, nil, err
		}

		return updated.ID, false, true, diffs, nil
	}

	if opts.DryRun {
		return "", true, false, nil, nil
	}

	newQuote, err := s.client.CreateQuote(ctx, BuildCreateQuoteRequest(quoteSpec, clientID, artifacts))
	if err != nil {
		return "", false, false, nil, err
	}

	return newQuote.ID, true, false, nil, nil
}

func (s *Syncer) syncInvoice(ctx context.Context, quoteSpec *spec.QuoteSpec, artifacts *spec.GeneratedArtifacts, clientID string, entry *store.StateEntry, opts SyncOptions) (string, bool, bool, []diff.FieldDiff, error) {
	var existingInvoice *NinjaInvoice
	var err error

	if opts.InvoiceID != "" {
		existingInvoice, err = s.client.GetInvoice(ctx, opts.InvoiceID)
		if err != nil {
			return "", false, false, nil, err
		}
	} else if entry != nil && entry.InvoiceID != "" {
		existingInvoice, err = s.client.GetInvoice(ctx, entry.InvoiceID)
		if err != nil {
			existingInvoice = nil
		}
	}

	if existingInvoice == nil {
		existingInvoice, err = s.client.FindInvoiceByReference(ctx, quoteSpec.Metadata.Reference)
		if err != nil {
			return "", false, false, nil, err
		}
	}

	if existingInvoice != nil {
		diffs := s.computeInvoiceDiffs(existingInvoice, artifacts)

		if opts.DryRun {
			return existingInvoice.ID, false, false, diffs, nil
		}

		if opts.ShowDiff && len(diffs) > 0 {
			fmt.Println(diff.FormatFieldDiffs(diffs))
		}

		if len(diffs) == 0 {
			return existingInvoice.ID, false, false, nil, nil
		}

		updated, err := s.client.UpdateInvoice(ctx, existingInvoice.ID, BuildUpdateInvoiceRequest(quoteSpec, existingInvoice, artifacts))
		if err != nil {
			return "", false, false, nil, err
		}

		return updated.ID, false, true, diffs, nil
	}

	if opts.DryRun {
		return "", true, false, nil, nil
	}

	newInvoice, err := s.client.CreateInvoice(ctx, BuildCreateInvoiceRequest(quoteSpec, clientID, artifacts))
	if err != nil {
		return "", false, false, nil, err
	}

	return newInvoice.ID, true, false, nil, nil
}

func (s *Syncer) computeQuoteDiffs(existing *NinjaQuote, artifacts *spec.GeneratedArtifacts) []diff.FieldDiff {
	oldFields := map[string]string{
		"public_notes": existing.PublicNotes,
		"terms":        existing.Terms,
	}

	newFields := map[string]string{
		"public_notes": artifacts.PublicNotesText,
		"terms":        artifacts.TermsMarkdown,
	}

	return diff.ComputeFields(oldFields, newFields)
}

func (s *Syncer) computeInvoiceDiffs(existing *NinjaInvoice, artifacts *spec.GeneratedArtifacts) []diff.FieldDiff {
	oldFields := map[string]string{
		"public_notes": existing.PublicNotes,
		"terms":        existing.Terms,
	}

	newFields := map[string]string{
		"public_notes": artifacts.PublicNotesText,
		"terms":        artifacts.TermsMarkdown,
	}

	return diff.ComputeFields(oldFields, newFields)
}

func (s *Syncer) Pull(ctx context.Context, quoteSpec *spec.QuoteSpec, entityType string, entityID string) (interface{}, error) {
	refID, err := spec.ExtractReferenceID(quoteSpec.Metadata.Reference)
	if err != nil {
		return nil, fmt.Errorf("invalid reference: %w", err)
	}

	entry, _ := s.store.GetEntry(refID)

	switch entityType {
	case "quote":
		if entityID == "" && entry != nil {
			entityID = entry.QuoteID
		}
		if entityID != "" {
			return s.client.GetQuote(ctx, entityID)
		}
		return s.client.FindQuoteByReference(ctx, quoteSpec.Metadata.Reference)

	case "invoice":
		if entityID == "" && entry != nil {
			entityID = entry.InvoiceID
		}
		if entityID != "" {
			return s.client.GetInvoice(ctx, entityID)
		}
		return s.client.FindInvoiceByReference(ctx, quoteSpec.Metadata.Reference)

	default:
		return nil, fmt.Errorf("unknown entity type: %s", entityType)
	}
}

func (s *Syncer) Diff(ctx context.Context, quoteSpec *spec.QuoteSpec, artifacts *spec.GeneratedArtifacts, entityType string, entityID string) (*diff.DiffResult, []diff.FieldDiff, error) {
	entity, err := s.Pull(ctx, quoteSpec, entityType, entityID)
	if err != nil {
		return nil, nil, err
	}

	switch e := entity.(type) {
	case *NinjaQuote:
		if e == nil {
			return nil, nil, fmt.Errorf("quote not found")
		}
		fieldDiffs := s.computeQuoteDiffs(e, artifacts)
		textDiff := diff.Compute(e.PublicNotes+"\n\n"+e.Terms, artifacts.PublicNotesText+"\n\n"+artifacts.TermsMarkdown)
		return textDiff, fieldDiffs, nil

	case *NinjaInvoice:
		if e == nil {
			return nil, nil, fmt.Errorf("invoice not found")
		}
		fieldDiffs := s.computeInvoiceDiffs(e, artifacts)
		textDiff := diff.Compute(e.PublicNotes+"\n\n"+e.Terms, artifacts.PublicNotesText+"\n\n"+artifacts.TermsMarkdown)
		return textDiff, fieldDiffs, nil

	default:
		return nil, nil, fmt.Errorf("unexpected entity type")
	}
}
