package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/AlexiaChen/issue-kanban-mcp/internal/queue"
)

// QueueWithStats embeds Queue and includes per-queue statistics.
type QueueWithStats struct {
	queue.Queue
	Stats queue.QueueStats `json:"stats"`
}

// Client communicates with the issue-kanban-mcp REST API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a Client pointed at baseURL (e.g. "http://localhost:9292").
func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type serverError struct {
	Message string `json:"error"`
}

func (e *serverError) Error() string { return e.Message }

func (c *Client) doRequest(ctx context.Context, method, path string, body, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var se serverError
		_ = json.NewDecoder(resp.Body).Decode(&se)
		if se.Message == "" {
			se.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return &se
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// ListProjects returns all queues with their statistics.
func (c *Client) ListProjects(ctx context.Context) ([]QueueWithStats, error) {
	var result []QueueWithStats
	if err := c.doRequest(ctx, http.MethodGet, "/api/projects", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateProject creates a new queue.
func (c *Client) CreateProject(ctx context.Context, input queue.CreateQueueInput) (*queue.Queue, error) {
	var result queue.Queue
	if err := c.doRequest(ctx, http.MethodPost, "/api/projects", input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteProject deletes a queue by ID.
func (c *Client) DeleteProject(ctx context.Context, id int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/projects/%d", id), nil, nil)
}

// GetProjectStats returns statistics for a single queue.
func (c *Client) GetProjectStats(ctx context.Context, queueID int64) (*queue.QueueStats, error) {
	var result queue.QueueStats
	if err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/projects/%d/stats", queueID), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListIssues lists tasks in a queue, optionally filtered by status.
func (c *Client) ListIssues(ctx context.Context, queueID int64, status string) ([]queue.Task, error) {
	path := fmt.Sprintf("/api/projects/%d/issues", queueID)
	if status != "" {
		path += "?status=" + url.QueryEscape(status)
	}
	var result []queue.Task
	if err := c.doRequest(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetIssue returns a task by ID.
func (c *Client) GetIssue(ctx context.Context, id int64) (*queue.Task, error) {
	var result queue.Task
	if err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/issues/%d", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateIssue creates a new task.
func (c *Client) CreateIssue(ctx context.Context, input queue.CreateTaskInput) (*queue.Task, error) {
	var result queue.Task
	if err := c.doRequest(ctx, http.MethodPost, "/api/issues", input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateIssueStatus updates a task's status.
func (c *Client) UpdateIssueStatus(ctx context.Context, id int64, status queue.TaskStatus) (*queue.Task, error) {
	input := queue.UpdateTaskInput{Status: &status}
	var result queue.Task
	if err := c.doRequest(ctx, http.MethodPatch, fmt.Sprintf("/api/issues/%d", id), input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// EditIssue updates the title, description, and/or priority of a pending task.
func (c *Client) EditIssue(ctx context.Context, id int64, title, desc *string, priority *queue.Priority) (*queue.Task, error) {
	input := queue.EditTaskInput{Title: title, Description: desc, Priority: priority}
	var result queue.Task
	if err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/issues/%d", id), input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteIssue deletes a task by ID.
func (c *Client) DeleteIssue(ctx context.Context, id int64) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/issues/%d", id), nil, nil)
}

// PrioritizeIssue moves a task ahead of lower-priority pending tasks.
func (c *Client) PrioritizeIssue(ctx context.Context, id int64) (*queue.Task, error) {
	var result queue.Task
	if err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/api/issues/%d/prioritize", id), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
