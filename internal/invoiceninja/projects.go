package invoiceninja

import (
	"context"
	"fmt"
	"strings"

	"github.com/ninjops/ninjops/internal/httpx"
	"github.com/ninjops/ninjops/internal/spec"
)

func (c *Client) ListProjects(ctx context.Context, page, perPage int) (*ProjectListResponse, error) {
	u := fmt.Sprintf("%s/projects?page=%d&per_page=%d", c.baseURL, page, perPage)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var result ProjectListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetProject(ctx context.Context, id string) (*NinjaProject, error) {
	u := fmt.Sprintf("%s/projects/%s", c.baseURL, id)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", id, err)
	}

	var result ProjectResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

type CreateProjectRequest struct {
	ClientID      string  `json:"client_id"`
	Name          string  `json:"name"`
	Description   string  `json:"description,omitempty"`
	PublicNotes   string  `json:"public_notes,omitempty"`
	PrivateNotes  string  `json:"private_notes,omitempty"`
	DueDate       string  `json:"due_date,omitempty"`
	TaskRate      float64 `json:"task_rate,omitempty"`
	BudgetedHours float64 `json:"budgeted_hours,omitempty"`
}

func (c *Client) CreateProject(ctx context.Context, req CreateProjectRequest) (*NinjaProject, error) {
	u := fmt.Sprintf("%s/projects", c.baseURL)

	resp, err := c.http.Post(ctx, u, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	var result ProjectResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

type UpdateProjectRequest struct {
	ID            string  `json:"id"`
	ClientID      string  `json:"client_id,omitempty"`
	Name          string  `json:"name,omitempty"`
	Description   string  `json:"description,omitempty"`
	PublicNotes   string  `json:"public_notes,omitempty"`
	PrivateNotes  string  `json:"private_notes,omitempty"`
	DueDate       string  `json:"due_date,omitempty"`
	TaskRate      float64 `json:"task_rate,omitempty"`
	BudgetedHours float64 `json:"budgeted_hours,omitempty"`
}

func (c *Client) UpdateProject(ctx context.Context, id string, req UpdateProjectRequest) (*NinjaProject, error) {
	req.ID = id
	u := fmt.Sprintf("%s/projects/%s", c.baseURL, id)

	resp, err := c.http.Put(ctx, u, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update project %s: %w", id, err)
	}

	var result ProjectResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) FindProjectsByClient(ctx context.Context, clientID string, page, perPage int) (*ProjectListResponse, error) {
	u := fmt.Sprintf("%s/projects?client_id=%s&page=%d&per_page=%d", c.baseURL, clientID, page, perPage)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to find projects by client: %w", err)
	}

	var result ProjectListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	result.Data = filterProjectsByClient(result.Data, clientID)
	result.Meta.Pagination.Count = len(result.Data)

	return &result, nil
}

func (c *Client) FindProjectByName(ctx context.Context, clientID, name string) (*NinjaProject, error) {
	projects, err := c.listAllProjectsByClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	return findProjectByName(projects, name), nil
}

func (c *Client) listAllProjectsByClient(ctx context.Context, clientID string) ([]NinjaProject, error) {
	const perPage = 100
	page := 1
	projects := make([]NinjaProject, 0)

	for {
		resp, err := c.FindProjectsByClient(ctx, clientID, page, perPage)
		if err != nil {
			return nil, err
		}

		projects = append(projects, resp.Data...)

		if len(resp.Data) < perPage {
			break
		}

		if resp.Meta.Pagination.TotalPages > 0 && page >= resp.Meta.Pagination.TotalPages {
			break
		}

		page++
	}

	return projects, nil
}

func filterProjectsByClient(projects []NinjaProject, clientID string) []NinjaProject {
	if clientID == "" {
		return projects
	}

	filtered := make([]NinjaProject, 0, len(projects))
	for _, project := range projects {
		if project.ClientID == clientID {
			filtered = append(filtered, project)
		}
	}

	return filtered
}

func findProjectByName(projects []NinjaProject, name string) *NinjaProject {
	target := strings.TrimSpace(strings.ToLower(name))
	if target == "" {
		return nil
	}

	for _, project := range projects {
		if strings.TrimSpace(strings.ToLower(project.Name)) == target {
			match := project
			return &match
		}
	}

	return nil
}

func BuildCreateProjectRequest(quoteSpec *spec.QuoteSpec, clientID string) CreateProjectRequest {
	return CreateProjectRequest{
		ClientID:    clientID,
		Name:        quoteSpec.Project.Name,
		Description: quoteSpec.Project.Description,
		DueDate:     quoteSpec.Project.Deadline,
	}
}

func BuildUpdateProjectRequest(quoteSpec *spec.QuoteSpec, existing *NinjaProject) UpdateProjectRequest {
	return UpdateProjectRequest{
		ID:          existing.ID,
		ClientID:    existing.ClientID,
		Name:        firstNonEmpty(quoteSpec.Project.Name, existing.Name),
		Description: firstNonEmpty(quoteSpec.Project.Description, existing.Description),
		DueDate:     firstNonEmpty(quoteSpec.Project.Deadline, existing.DueDate),
	}
}
