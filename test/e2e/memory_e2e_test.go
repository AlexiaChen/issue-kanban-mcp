package e2e

import (
	"fmt"
	"os"
	"testing"
)

func TestE2E_MemoryCRUD(t *testing.T) {
	if serverURL == "" {
		t.Skip("E2E_SERVER_URL not set")
	}
	client := NewE2EClient(serverURL)

	// Create project
	proj, err := client.CreateProject(map[string]interface{}{
		"name": fmt.Sprintf("Memory E2E %d", os.Getpid()),
	})
	if err != nil {
		t.Fatalf("create project failed: %v", err)
	}
	projectID := int64(proj["id"].(float64))
	defer client.DeleteProject(projectID)

	// Store memory
	mem, err := client.StoreMemory(projectID, map[string]interface{}{
		"content":    "Go uses goroutines for lightweight concurrency",
		"category":   "fact",
		"importance": 4,
		"tags":       "go,concurrency",
	})
	if err != nil {
		t.Fatalf("store memory failed: %v", err)
	}
	memoryID := int64(mem["id"].(float64))
	t.Logf("Stored memory ID: %d", memoryID)

	if mem["category"] != "fact" {
		t.Errorf("expected category 'fact', got %v", mem["category"])
	}

	// Store another memory for search
	_, err = client.StoreMemory(projectID, map[string]interface{}{
		"content":  "Python uses GIL for thread safety",
		"category": "fact",
	})
	if err != nil {
		t.Fatalf("store second memory failed: %v", err)
	}

	// List memories
	mems, err := client.ListMemories(projectID, "")
	if err != nil {
		t.Fatalf("list memories failed: %v", err)
	}
	if len(mems) != 2 {
		t.Errorf("expected 2 memories, got %d", len(mems))
	}

	// Search memories
	results, err := client.SearchMemories(projectID, "goroutines", "")
	if err != nil {
		t.Fatalf("search memories failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results for 'goroutines'")
	}

	// Delete memory
	err = client.DeleteMemory(projectID, memoryID)
	if err != nil {
		t.Fatalf("delete memory failed: %v", err)
	}

	// Verify deleted
	mems, err = client.ListMemories(projectID, "")
	if err != nil {
		t.Fatalf("list after delete failed: %v", err)
	}
	if len(mems) != 1 {
		t.Errorf("expected 1 memory after delete, got %d", len(mems))
	}
}

// Memory E2E Client methods

func (c *E2EClient) StoreMemory(projectID int64, data map[string]interface{}) (map[string]interface{}, error) {
	return c.doObject("POST", fmt.Sprintf("/api/projects/%d/memories", projectID), data)
}

func (c *E2EClient) ListMemories(projectID int64, category string) ([]map[string]interface{}, error) {
	path := fmt.Sprintf("/api/projects/%d/memories", projectID)
	if category != "" {
		path += "?category=" + category
	}
	return c.doArray("GET", path, nil)
}

func (c *E2EClient) SearchMemories(projectID int64, query, category string) ([]map[string]interface{}, error) {
	path := fmt.Sprintf("/api/projects/%d/memories/search?q=%s", projectID, query)
	if category != "" {
		path += "&category=" + category
	}
	return c.doArray("GET", path, nil)
}

func (c *E2EClient) DeleteMemory(projectID, memoryID int64) error {
	_, err := c.doObject("DELETE", fmt.Sprintf("/api/projects/%d/memories/%d", projectID, memoryID), nil)
	return err
}
