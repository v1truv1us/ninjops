package invoiceninja

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ninjops/ninjops/internal/config"
	"github.com/ninjops/ninjops/internal/httpx"
)

const (
	APIVersion = "/api/v1"
)

type Client struct {
	http    *httpx.Client
	baseURL string
}

func NewClient(cfg config.NinjaConfig) *Client {
	httpConfig := httpx.DefaultClientConfig()
	httpClient := httpx.NewClientWithAuth(httpConfig, cfg.APIToken, cfg.APISecret)

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")

	return &Client{
		http:    httpClient,
		baseURL: baseURL + APIVersion,
	}
}

func (c *Client) apiURL(path string) string {
	return c.baseURL + path
}

func (c *Client) TestConnection(ctx context.Context) error {
	resp, err := c.http.Get(ctx, c.apiURL("/clients?per_page=1"))
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection test failed: HTTP %d", resp.StatusCode)
	}

	return nil
}

func ensureSuccessResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		resp.Body.Close()
		return nil
	}

	body, err := httpx.ReadBody(resp)
	if err != nil {
		return fmt.Errorf("HTTP %d: failed to read response body: %w", resp.StatusCode, err)
	}

	if len(body) == 0 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

type NinjaClient struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	Email       string          `json:"email"`
	Phone       string          `json:"phone"`
	Address1    string          `json:"address1"`
	Address2    string          `json:"address2"`
	City        string          `json:"city"`
	State       string          `json:"state"`
	PostalCode  string          `json:"postal_code"`
	CountryID   string          `json:"country_id"`
	Custom1     string          `json:"custom_value1"`
	Custom2     string          `json:"custom_value2"`
	CreatedAt   int64           `json:"created_at"`
	UpdatedAt   int64           `json:"updated_at"`
	Contacts    []ClientContact `json:"contacts,omitempty"`
}

type ClientContact struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	IsPrimary bool   `json:"is_primary"`
}

type NinjaQuote struct {
	ID               string          `json:"id"`
	Number           string          `json:"number"`
	ClientID         string          `json:"client_id"`
	ProjectID        string          `json:"project_id"`
	Discount         float64         `json:"discount"`
	IsAmountDiscount bool            `json:"is_amount_discount"`
	LineItems        []NinjaLineItem `json:"line_items"`
	PublicNotes      string          `json:"public_notes"`
	PrivateNotes     string          `json:"private_notes"`
	Terms            string          `json:"terms"`
	Footer           string          `json:"footer"`
	Custom1          string          `json:"custom_value1"`
	Custom2          string          `json:"custom_value2"`
	Custom3          string          `json:"custom_value3"`
	Custom4          string          `json:"custom_value4"`
	StatusID         string          `json:"status_id"`
	Amount           float64         `json:"amount"`
	Balance          float64         `json:"balance"`
	CreatedAt        int64           `json:"created_at"`
	UpdatedAt        int64           `json:"updated_at"`
}

type NinjaInvoice struct {
	ID               string          `json:"id"`
	Number           string          `json:"number"`
	ClientID         string          `json:"client_id"`
	ProjectID        string          `json:"project_id"`
	Discount         float64         `json:"discount"`
	IsAmountDiscount bool            `json:"is_amount_discount"`
	LineItems        []NinjaLineItem `json:"line_items"`
	PublicNotes      string          `json:"public_notes"`
	PrivateNotes     string          `json:"private_notes"`
	Terms            string          `json:"terms"`
	Footer           string          `json:"footer"`
	Custom1          string          `json:"custom_value1"`
	Custom2          string          `json:"custom_value2"`
	Custom3          string          `json:"custom_value3"`
	Custom4          string          `json:"custom_value4"`
	StatusID         string          `json:"status_id"`
	Amount           float64         `json:"amount"`
	Balance          float64         `json:"balance"`
	CreatedAt        int64           `json:"created_at"`
	UpdatedAt        int64           `json:"updated_at"`
}

type NinjaProject struct {
	ID            string  `json:"id"`
	ClientID      string  `json:"client_id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	PublicNotes   string  `json:"public_notes"`
	PrivateNotes  string  `json:"private_notes"`
	DueDate       string  `json:"due_date"`
	TaskRate      float64 `json:"task_rate"`
	BudgetedHours float64 `json:"budgeted_hours"`
	CreatedAt     int64   `json:"created_at"`
	UpdatedAt     int64   `json:"updated_at"`
}

type NinjaTask struct {
	ID          string  `json:"id"`
	ClientID    string  `json:"client_id"`
	ProjectID   string  `json:"project_id"`
	Number      string  `json:"number"`
	Description string  `json:"description"`
	Rate        float64 `json:"rate"`
	Duration    int64   `json:"duration"`
	TimeLog     string  `json:"time_log"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`
}

type NinjaLineItem struct {
	Quantity         float64 `json:"quantity"`
	Cost             float64 `json:"cost"`
	ProductCost      float64 `json:"product_cost,omitempty"`
	ProductKey       string  `json:"product_key,omitempty"`
	Notes            string  `json:"notes"`
	Discount         float64 `json:"discount"`
	IsAmountDiscount bool    `json:"is_amount_discount"`
}

type APIResponse struct {
	Data interface{} `json:"data"`
	Meta APIMeta     `json:"meta"`
}

type APIMeta struct {
	Pagination APIPagination `json:"pagination"`
}

type APIPagination struct {
	Total       int               `json:"total"`
	Count       int               `json:"count"`
	PerPage     int               `json:"per_page"`
	CurrentPage int               `json:"current_page"`
	TotalPages  int               `json:"total_pages"`
	Links       map[string]string `json:"links"`
}

type ClientListResponse struct {
	Data []NinjaClient `json:"data"`
	Meta APIMeta       `json:"meta"`
}

type QuoteListResponse struct {
	Data []NinjaQuote `json:"data"`
	Meta APIMeta      `json:"meta"`
}

type InvoiceListResponse struct {
	Data []NinjaInvoice `json:"data"`
	Meta APIMeta        `json:"meta"`
}

type ProjectListResponse struct {
	Data []NinjaProject `json:"data"`
	Meta APIMeta        `json:"meta"`
}

type TaskListResponse struct {
	Data []NinjaTask `json:"data"`
	Meta APIMeta     `json:"meta"`
}

type ClientResponse struct {
	Data NinjaClient `json:"data"`
}

type QuoteResponse struct {
	Data NinjaQuote `json:"data"`
}

type InvoiceResponse struct {
	Data NinjaInvoice `json:"data"`
}

type ProjectResponse struct {
	Data NinjaProject `json:"data"`
}

type TaskResponse struct {
	Data NinjaTask `json:"data"`
}

type NinjaError struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors"`
}

func (e *NinjaError) Error() string {
	if len(e.Errors) > 0 {
		var errs []string
		for field, messages := range e.Errors {
			errs = append(errs, fmt.Sprintf("%s: %s", field, strings.Join(messages, ", ")))
		}
		return fmt.Sprintf("%s: %s", e.Message, strings.Join(errs, "; "))
	}
	return e.Message
}

func timestampToTime(ts int64) time.Time {
	return time.Unix(ts, 0)
}

func timeToTimestamp(t time.Time) int64 {
	return t.Unix()
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
