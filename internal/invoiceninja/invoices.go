package invoiceninja

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ninjops/ninjops/internal/httpx"
	"github.com/ninjops/ninjops/internal/spec"
)

func (c *Client) ListInvoices(ctx context.Context, page, perPage int) (*InvoiceListResponse, error) {
	u := fmt.Sprintf("%s/invoices?page=%d&per_page=%d", c.baseURL, page, perPage)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}

	var result InvoiceListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetInvoice(ctx context.Context, id string) (*NinjaInvoice, error) {
	u := fmt.Sprintf("%s/invoices/%s", c.baseURL, id)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice %s: %w", id, err)
	}

	var result InvoiceResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) FindInvoiceByReference(ctx context.Context, reference string) (*NinjaInvoice, error) {
	u := fmt.Sprintf("%s/invoices?custom_value1=%s", c.baseURL, url.QueryEscape(reference))

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to find invoice by reference: %w", err)
	}

	var result InvoiceListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func (c *Client) FindInvoicesByClient(ctx context.Context, clientID string, page, perPage int) (*InvoiceListResponse, error) {
	u := fmt.Sprintf("%s/invoices?client_id=%s&page=%d&per_page=%d", c.baseURL, clientID, page, perPage)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to find invoices by client: %w", err)
	}

	var result InvoiceListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

type CreateInvoiceRequest struct {
	ClientID         string          `json:"client_id"`
	ProjectID        string          `json:"project_id,omitempty"`
	Number           string          `json:"number,omitempty"`
	Discount         float64         `json:"discount,omitempty"`
	IsAmountDiscount bool            `json:"is_amount_discount,omitempty"`
	LineItems        []NinjaLineItem `json:"line_items"`
	PublicNotes      string          `json:"public_notes,omitempty"`
	PrivateNotes     string          `json:"private_notes,omitempty"`
	Terms            string          `json:"terms,omitempty"`
	Footer           string          `json:"footer,omitempty"`
	Custom1          string          `json:"custom_value1,omitempty"`
	Custom2          string          `json:"custom_value2,omitempty"`
	Custom3          string          `json:"custom_value3,omitempty"`
	Custom4          string          `json:"custom_value4,omitempty"`
}

func (c *Client) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*NinjaInvoice, error) {
	u := fmt.Sprintf("%s/invoices", c.baseURL)

	resp, err := c.http.Post(ctx, u, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	var result InvoiceResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

type UpdateInvoiceRequest struct {
	ID               string          `json:"id"`
	ClientID         string          `json:"client_id,omitempty"`
	ProjectID        string          `json:"project_id,omitempty"`
	Number           string          `json:"number,omitempty"`
	Discount         float64         `json:"discount,omitempty"`
	IsAmountDiscount bool            `json:"is_amount_discount,omitempty"`
	LineItems        []NinjaLineItem `json:"line_items,omitempty"`
	PublicNotes      string          `json:"public_notes,omitempty"`
	PrivateNotes     string          `json:"private_notes,omitempty"`
	Terms            string          `json:"terms,omitempty"`
	Footer           string          `json:"footer,omitempty"`
	Custom1          string          `json:"custom_value1,omitempty"`
	Custom2          string          `json:"custom_value2,omitempty"`
	Custom3          string          `json:"custom_value3,omitempty"`
	Custom4          string          `json:"custom_value4,omitempty"`
}

func (c *Client) UpdateInvoice(ctx context.Context, id string, req UpdateInvoiceRequest) (*NinjaInvoice, error) {
	req.ID = id
	u := fmt.Sprintf("%s/invoices/%s", c.baseURL, id)

	resp, err := c.http.Put(ctx, u, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update invoice %s: %w", id, err)
	}

	var result InvoiceResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) UpdateInvoiceFields(ctx context.Context, id string, fields map[string]interface{}) (*NinjaInvoice, error) {
	existing, err := c.GetInvoice(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing invoice: %w", err)
	}

	req := UpdateInvoiceRequest{
		ID:               id,
		ClientID:         existing.ClientID,
		ProjectID:        existing.ProjectID,
		Number:           existing.Number,
		Discount:         existing.Discount,
		IsAmountDiscount: existing.IsAmountDiscount,
		LineItems:        existing.LineItems,
		PublicNotes:      existing.PublicNotes,
		PrivateNotes:     existing.PrivateNotes,
		Terms:            existing.Terms,
		Footer:           existing.Footer,
		Custom1:          existing.Custom1,
		Custom2:          existing.Custom2,
		Custom3:          existing.Custom3,
		Custom4:          existing.Custom4,
	}

	if v, ok := fields["public_notes"].(string); ok {
		req.PublicNotes = v
	}
	if v, ok := fields["terms"].(string); ok {
		req.Terms = v
	}
	if v, ok := fields["private_notes"].(string); ok {
		req.PrivateNotes = v
	}
	if v, ok := fields["custom_value1"].(string); ok {
		req.Custom1 = v
	}
	if v, ok := fields["custom_value2"].(string); ok {
		req.Custom2 = v
	}
	if v, ok := fields["line_items"].([]NinjaLineItem); ok {
		req.LineItems = v
	}

	return c.UpdateInvoice(ctx, id, req)
}

func (c *Client) EmailInvoice(ctx context.Context, invoiceID string) error {
	u := fmt.Sprintf("%s/invoices/%s?action=email", c.baseURL, invoiceID)

	resp, err := c.http.Put(ctx, u, map[string]string{"id": invoiceID})
	if err != nil {
		return fmt.Errorf("failed to email invoice %s: %w", invoiceID, err)
	}

	if err := ensureSuccessResponse(resp); err != nil {
		return fmt.Errorf("failed to email invoice %s: %w", invoiceID, err)
	}

	return nil
}

func (c *Client) MarkInvoiceSent(ctx context.Context, invoiceID string) error {
	u := fmt.Sprintf("%s/invoices/%s?action=mark_sent", c.baseURL, invoiceID)

	resp, err := c.http.Put(ctx, u, map[string]string{"id": invoiceID})
	if err != nil {
		return fmt.Errorf("failed to mark invoice %s as sent: %w", invoiceID, err)
	}

	if err := ensureSuccessResponse(resp); err != nil {
		return fmt.Errorf("failed to mark invoice %s as sent: %w", invoiceID, err)
	}

	return nil
}

func (c *Client) ArchiveInvoice(ctx context.Context, invoiceID string) error {
	u := fmt.Sprintf("%s/invoices/%s?action=archive", c.baseURL, invoiceID)

	resp, err := c.http.Put(ctx, u, map[string]string{"id": invoiceID})
	if err != nil {
		return fmt.Errorf("failed to archive invoice %s: %w", invoiceID, err)
	}

	if err := ensureSuccessResponse(resp); err != nil {
		return fmt.Errorf("failed to archive invoice %s: %w", invoiceID, err)
	}

	return nil
}

func (c *Client) CancelInvoice(ctx context.Context, invoiceID string) error {
	u := fmt.Sprintf("%s/invoices/%s?action=cancel", c.baseURL, invoiceID)

	resp, err := c.http.Put(ctx, u, map[string]string{"id": invoiceID})
	if err != nil {
		return fmt.Errorf("failed to cancel invoice %s: %w", invoiceID, err)
	}

	if err := ensureSuccessResponse(resp); err != nil {
		return fmt.Errorf("failed to cancel invoice %s: %w", invoiceID, err)
	}

	return nil
}

func BuildCreateInvoiceRequest(spec *spec.QuoteSpec, clientID string, artifacts *spec.GeneratedArtifacts) CreateInvoiceRequest {
	return CreateInvoiceRequest{
		ClientID:         clientID,
		ProjectID:        spec.Project.ID,
		PublicNotes:      artifacts.PublicNotesText,
		Terms:            artifacts.TermsMarkdown,
		PrivateNotes:     fmt.Sprintf("Generated by ninjops\nReference: %s\nHash: %s", spec.Metadata.Reference, artifacts.Meta.Hash),
		LineItems:        BuildLineItems(spec),
		Discount:         getDiscountAmount(spec),
		IsAmountDiscount: isAmountDiscount(spec),
		Custom1:          spec.Metadata.Reference,
	}
}

func BuildUpdateInvoiceRequest(spec *spec.QuoteSpec, existing *NinjaInvoice, artifacts *spec.GeneratedArtifacts) UpdateInvoiceRequest {
	return UpdateInvoiceRequest{
		ID:               existing.ID,
		ClientID:         existing.ClientID,
		ProjectID:        firstNonEmpty(spec.Project.ID, existing.ProjectID),
		Number:           existing.Number,
		PublicNotes:      artifacts.PublicNotesText,
		Terms:            artifacts.TermsMarkdown,
		PrivateNotes:     existing.PrivateNotes + "\n\nUpdated by ninjops at " + spec.Metadata.UpdatedAt.Format("2006-01-02 15:04:05"),
		LineItems:        BuildLineItems(spec),
		Discount:         getDiscountAmount(spec),
		IsAmountDiscount: isAmountDiscount(spec),
		Custom1:          spec.Metadata.Reference,
	}
}

func FormatInvoiceSummary(invoice *NinjaInvoice) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Invoice #%s\n", invoice.Number))
	sb.WriteString(fmt.Sprintf("  ID: %s\n", invoice.ID))
	sb.WriteString(fmt.Sprintf("  Client ID: %s\n", invoice.ClientID))
	sb.WriteString(fmt.Sprintf("  Amount: %.2f\n", invoice.Amount))
	sb.WriteString(fmt.Sprintf("  Balance: %.2f\n", invoice.Balance))
	sb.WriteString(fmt.Sprintf("  Status: %s\n", invoice.StatusID))
	if invoice.Custom1 != "" {
		sb.WriteString(fmt.Sprintf("  Reference: %s\n", invoice.Custom1))
	}
	return sb.String()
}
