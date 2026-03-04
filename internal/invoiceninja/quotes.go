package invoiceninja

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/ninjops/ninjops/internal/httpx"
	"github.com/ninjops/ninjops/internal/spec"
)

func (c *Client) ListQuotes(ctx context.Context, page, perPage int) (*QuoteListResponse, error) {
	u := fmt.Sprintf("%s/quotes?page=%d&per_page=%d", c.baseURL, page, perPage)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to list quotes: %w", err)
	}

	var result QuoteListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetQuote(ctx context.Context, id string) (*NinjaQuote, error) {
	u := fmt.Sprintf("%s/quotes/%s", c.baseURL, id)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote %s: %w", id, err)
	}

	var result QuoteResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) FindQuoteByReference(ctx context.Context, reference string) (*NinjaQuote, error) {
	u := fmt.Sprintf("%s/quotes?custom_value1=%s", c.baseURL, url.QueryEscape(reference))

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to find quote by reference: %w", err)
	}

	var result QuoteListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func (c *Client) FindQuotesByClient(ctx context.Context, clientID string, page, perPage int) (*QuoteListResponse, error) {
	u := fmt.Sprintf("%s/quotes?client_id=%s&page=%d&per_page=%d", c.baseURL, clientID, page, perPage)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to find quotes by client: %w", err)
	}

	var result QuoteListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

type CreateQuoteRequest struct {
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

func (c *Client) CreateQuote(ctx context.Context, req CreateQuoteRequest) (*NinjaQuote, error) {
	u := fmt.Sprintf("%s/quotes", c.baseURL)

	resp, err := c.http.Post(ctx, u, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create quote: %w", err)
	}

	var result QuoteResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

type UpdateQuoteRequest struct {
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

func (c *Client) UpdateQuote(ctx context.Context, id string, req UpdateQuoteRequest) (*NinjaQuote, error) {
	req.ID = id
	u := fmt.Sprintf("%s/quotes/%s", c.baseURL, id)

	resp, err := c.http.Put(ctx, u, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update quote %s: %w", id, err)
	}

	var result QuoteResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) UpdateQuoteFields(ctx context.Context, id string, fields map[string]interface{}) (*NinjaQuote, error) {
	existing, err := c.GetQuote(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing quote: %w", err)
	}

	req := UpdateQuoteRequest{
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

	return c.UpdateQuote(ctx, id, req)
}

func (c *Client) ConvertQuoteToInvoice(ctx context.Context, quoteID string) (*NinjaInvoice, error) {
	u := fmt.Sprintf("%s/quotes/%s?action=convert_to_invoice", c.baseURL, quoteID)

	resp, err := c.http.Put(ctx, u, map[string]string{"id": quoteID})
	if err != nil {
		return nil, fmt.Errorf("failed to convert quote %s to invoice: %w", quoteID, err)
	}

	var result InvoiceResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) EmailQuote(ctx context.Context, quoteID string) error {
	u := fmt.Sprintf("%s/quotes/%s?action=email", c.baseURL, quoteID)

	resp, err := c.http.Put(ctx, u, map[string]string{"id": quoteID})
	if err != nil {
		return fmt.Errorf("failed to email quote %s: %w", quoteID, err)
	}

	if err := ensureSuccessResponse(resp); err != nil {
		return fmt.Errorf("failed to email quote %s: %w", quoteID, err)
	}

	return nil
}

func (c *Client) ArchiveQuote(ctx context.Context, quoteID string) error {
	u := fmt.Sprintf("%s/quotes/%s?action=archive", c.baseURL, quoteID)

	resp, err := c.http.Put(ctx, u, map[string]string{"id": quoteID})
	if err != nil {
		return fmt.Errorf("failed to archive quote %s: %w", quoteID, err)
	}

	if err := ensureSuccessResponse(resp); err != nil {
		return fmt.Errorf("failed to archive quote %s: %w", quoteID, err)
	}

	return nil
}

func BuildLineItems(spec *spec.QuoteSpec) []NinjaLineItem {
	items := make([]NinjaLineItem, 0, len(spec.Pricing.LineItems))

	for _, li := range spec.Pricing.LineItems {
		items = append(items, NinjaLineItem{
			Quantity:   li.Quantity,
			Cost:       li.Rate,
			ProductKey: li.Category,
			Notes:      li.Description,
		})
	}

	return items
}

func BuildCreateQuoteRequest(spec *spec.QuoteSpec, clientID string, artifacts *spec.GeneratedArtifacts) CreateQuoteRequest {
	return CreateQuoteRequest{
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

func BuildUpdateQuoteRequest(spec *spec.QuoteSpec, existing *NinjaQuote, artifacts *spec.GeneratedArtifacts) UpdateQuoteRequest {
	return UpdateQuoteRequest{
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
}

func getDiscountAmount(spec *spec.QuoteSpec) float64 {
	if spec.Pricing.Discount == nil {
		return 0
	}
	if spec.Pricing.Discount.Amount > 0 {
		return spec.Pricing.Discount.Amount
	}
	return spec.Pricing.Discount.Percentage
}

func isAmountDiscount(spec *spec.QuoteSpec) bool {
	if spec.Pricing.Discount == nil {
		return false
	}
	return spec.Pricing.Discount.Amount > 0
}

func FormatQuoteSummary(quote *NinjaQuote) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Quote #%s\n", quote.Number))
	sb.WriteString(fmt.Sprintf("  ID: %s\n", quote.ID))
	sb.WriteString(fmt.Sprintf("  Client ID: %s\n", quote.ClientID))
	sb.WriteString(fmt.Sprintf("  Amount: %.2f\n", quote.Amount))
	sb.WriteString(fmt.Sprintf("  Status: %s\n", quote.StatusID))
	if quote.Custom1 != "" {
		sb.WriteString(fmt.Sprintf("  Reference: %s\n", quote.Custom1))
	}
	return sb.String()
}
