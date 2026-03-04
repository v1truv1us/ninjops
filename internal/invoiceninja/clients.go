package invoiceninja

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ninjops/ninjops/internal/httpx"
)

func (c *Client) ListClients(ctx context.Context, page, perPage int) (*ClientListResponse, error) {
	u := fmt.Sprintf("%s/clients?page=%d&per_page=%d", c.baseURL, page, perPage)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to list clients: %w", err)
	}

	var result ClientListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetClient(ctx context.Context, id string) (*NinjaClient, error) {
	u := fmt.Sprintf("%s/clients/%s", c.baseURL, id)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to get client %s: %w", id, err)
	}

	var result ClientResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) FindClientByEmail(ctx context.Context, email string) (*NinjaClient, error) {
	u := fmt.Sprintf("%s/clients?email=%s", c.baseURL, url.QueryEscape(email))

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to find client by email: %w", err)
	}

	var result ClientListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func (c *Client) FindClientByName(ctx context.Context, name string) (*NinjaClient, error) {
	u := fmt.Sprintf("%s/clients?name=%s", c.baseURL, url.QueryEscape(name))

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to find client by name: %w", err)
	}

	var result ClientListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

type CreateClientRequest struct {
	Name       string          `json:"name"`
	Email      string          `json:"email,omitempty"`
	Phone      string          `json:"phone,omitempty"`
	Address1   string          `json:"address1,omitempty"`
	Address2   string          `json:"address2,omitempty"`
	City       string          `json:"city,omitempty"`
	State      string          `json:"state,omitempty"`
	PostalCode string          `json:"postal_code,omitempty"`
	CountryID  string          `json:"country_id,omitempty"`
	Custom1    string          `json:"custom_value1,omitempty"`
	Custom2    string          `json:"custom_value2,omitempty"`
	Contacts   []ClientContact `json:"contacts,omitempty"`
}

func (c *Client) CreateClient(ctx context.Context, req CreateClientRequest) (*NinjaClient, error) {
	u := fmt.Sprintf("%s/clients", c.baseURL)

	resp, err := c.http.Post(ctx, u, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	var result ClientResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

type UpdateClientRequest struct {
	ID         string          `json:"id"`
	Name       string          `json:"name,omitempty"`
	Email      string          `json:"email,omitempty"`
	Phone      string          `json:"phone,omitempty"`
	Address1   string          `json:"address1,omitempty"`
	Address2   string          `json:"address2,omitempty"`
	City       string          `json:"city,omitempty"`
	State      string          `json:"state,omitempty"`
	PostalCode string          `json:"postal_code,omitempty"`
	CountryID  string          `json:"country_id,omitempty"`
	Custom1    string          `json:"custom_value1,omitempty"`
	Custom2    string          `json:"custom_value2,omitempty"`
	Contacts   []ClientContact `json:"contacts,omitempty"`
}

func (c *Client) UpdateClient(ctx context.Context, id string, req UpdateClientRequest) (*NinjaClient, error) {
	req.ID = id
	u := fmt.Sprintf("%s/clients/%s", c.baseURL, id)

	resp, err := c.http.Put(ctx, u, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update client %s: %w", id, err)
	}

	var result ClientResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) UpsertClient(ctx context.Context, req CreateClientRequest) (*NinjaClient, bool, error) {
	if req.Email != "" {
		existing, err := c.FindClientByEmail(ctx, req.Email)
		if err != nil {
			return nil, false, err
		}
		if existing != nil {
			updateReq := UpdateClientRequest{
				Name:       req.Name,
				Email:      req.Email,
				Phone:      req.Phone,
				Address1:   req.Address1,
				Address2:   req.Address2,
				City:       req.City,
				State:      req.State,
				PostalCode: req.PostalCode,
				CountryID:  req.CountryID,
				Custom1:    req.Custom1,
				Custom2:    req.Custom2,
				Contacts:   req.Contacts,
			}
			updated, err := c.UpdateClient(ctx, existing.ID, updateReq)
			return updated, false, err
		}
	}

	if req.Name != "" {
		existing, err := c.FindClientByName(ctx, req.Name)
		if err != nil {
			return nil, false, err
		}
		if existing != nil {
			updateReq := UpdateClientRequest{
				Name:       req.Name,
				Email:      req.Email,
				Phone:      req.Phone,
				Address1:   req.Address1,
				Address2:   req.Address2,
				City:       req.City,
				State:      req.State,
				PostalCode: req.PostalCode,
				CountryID:  req.CountryID,
				Custom1:    req.Custom1,
				Custom2:    req.Custom2,
				Contacts:   req.Contacts,
			}
			updated, err := c.UpdateClient(ctx, existing.ID, updateReq)
			return updated, false, err
		}
	}

	client, err := c.CreateClient(ctx, req)
	return client, true, err
}

func (c *Client) ArchiveClient(ctx context.Context, id string) error {
	u := fmt.Sprintf("%s/clients/%s?action=archive", c.baseURL, id)

	resp, err := c.http.Put(ctx, u, map[string]string{"id": id})
	if err != nil {
		return fmt.Errorf("failed to archive client %s: %w", id, err)
	}

	if err := ensureSuccessResponse(resp); err != nil {
		return fmt.Errorf("failed to archive client %s: %w", id, err)
	}

	return nil
}

func (c *Client) DeleteClient(ctx context.Context, id string) error {
	u := fmt.Sprintf("%s/clients/%s?action=delete", c.baseURL, id)

	resp, err := c.http.Put(ctx, u, map[string]string{"id": id})
	if err != nil {
		return fmt.Errorf("failed to delete client %s: %w", id, err)
	}

	if err := ensureSuccessResponse(resp); err != nil {
		return fmt.Errorf("failed to delete client %s: %w", id, err)
	}

	return nil
}
