package invoiceninja

import (
	"context"
	"fmt"
	"strings"

	"github.com/ninjops/ninjops/internal/httpx"
)

func (c *Client) ListTasks(ctx context.Context, page, perPage int) (*TaskListResponse, error) {
	u := fmt.Sprintf("%s/tasks?page=%d&per_page=%d", c.baseURL, page, perPage)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	var result TaskListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetTask(ctx context.Context, id string) (*NinjaTask, error) {
	u := fmt.Sprintf("%s/tasks/%s", c.baseURL, id)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to get task %s: %w", id, err)
	}

	var result TaskResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

type CreateTaskRequest struct {
	ClientID    string  `json:"client_id,omitempty"`
	ProjectID   string  `json:"project_id,omitempty"`
	Description string  `json:"description"`
	Rate        float64 `json:"rate,omitempty"`
	Duration    int64   `json:"duration,omitempty"`
	TimeLog     string  `json:"time_log,omitempty"`
}

func (c *Client) CreateTask(ctx context.Context, req CreateTaskRequest) (*NinjaTask, error) {
	u := fmt.Sprintf("%s/tasks", c.baseURL)

	resp, err := c.http.Post(ctx, u, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	var result TaskResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

type UpdateTaskRequest struct {
	ID          string  `json:"id"`
	ClientID    string  `json:"client_id,omitempty"`
	ProjectID   string  `json:"project_id,omitempty"`
	Description string  `json:"description,omitempty"`
	Rate        float64 `json:"rate,omitempty"`
	Duration    int64   `json:"duration,omitempty"`
	TimeLog     string  `json:"time_log,omitempty"`
}

func (c *Client) UpdateTask(ctx context.Context, id string, req UpdateTaskRequest) (*NinjaTask, error) {
	req.ID = id
	u := fmt.Sprintf("%s/tasks/%s", c.baseURL, id)

	resp, err := c.http.Put(ctx, u, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update task %s: %w", id, err)
	}

	var result TaskResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) FindTasksByProject(ctx context.Context, projectID string, page, perPage int) (*TaskListResponse, error) {
	u := fmt.Sprintf("%s/tasks?project_id=%s&page=%d&per_page=%d", c.baseURL, projectID, page, perPage)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by project: %w", err)
	}

	var result TaskListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) FindTasksByClient(ctx context.Context, clientID string, page, perPage int) (*TaskListResponse, error) {
	u := fmt.Sprintf("%s/tasks?client_id=%s&page=%d&per_page=%d", c.baseURL, clientID, page, perPage)

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks by client: %w", err)
	}

	var result TaskListResponse
	if err := httpx.ParseJSONResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) FindTaskByDescription(ctx context.Context, projectID, description string) (*NinjaTask, error) {
	tasks, err := c.listAllTasksByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return findTaskByDescription(tasks, description), nil
}

func (c *Client) listAllTasksByProject(ctx context.Context, projectID string) ([]NinjaTask, error) {
	const perPage = 100
	page := 1
	tasks := make([]NinjaTask, 0)

	for {
		resp, err := c.FindTasksByProject(ctx, projectID, page, perPage)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, resp.Data...)

		if len(resp.Data) < perPage {
			break
		}

		if resp.Meta.Pagination.TotalPages > 0 && page >= resp.Meta.Pagination.TotalPages {
			break
		}

		page++
	}

	return tasks, nil
}

func filterTasksByProject(tasks []NinjaTask, projectID string) []NinjaTask {
	if projectID == "" {
		return tasks
	}

	filtered := make([]NinjaTask, 0, len(tasks))
	for _, task := range tasks {
		if task.ProjectID == projectID {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

func filterTasksByClient(tasks []NinjaTask, clientID string) []NinjaTask {
	if clientID == "" {
		return tasks
	}

	filtered := make([]NinjaTask, 0, len(tasks))
	for _, task := range tasks {
		if task.ClientID == clientID {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

func findTaskByDescription(tasks []NinjaTask, description string) *NinjaTask {
	target := strings.TrimSpace(strings.ToLower(description))
	if target == "" {
		return nil
	}

	for _, task := range tasks {
		if strings.TrimSpace(strings.ToLower(task.Description)) == target {
			match := task
			return &match
		}
	}

	return nil
}
