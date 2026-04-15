package memory_test

import (
	"context"
	"os"
	"testing"

	"github.com/AlexiaChen/issue-kanban-mcp/internal/memory"
	"github.com/AlexiaChen/issue-kanban-mcp/internal/queue"
	"github.com/AlexiaChen/issue-kanban-mcp/internal/storage"
)

func TestIntegration_MemoryWorkflow(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "memory-integration-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	store, err := storage.NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer store.Close()

	queueMgr := queue.NewManager(store)
	memMgr := memory.NewMemoryManager(store)
	ctx := context.Background()

	// Create project
	proj, err := queueMgr.CreateProject(ctx, queue.CreateQueueInput{
		Name:        "Memory Integration Project",
		Description: "Testing memory system end-to-end",
	})
	if err != nil {
		t.Fatalf("failed to create project: %v", err)
	}
	pid := proj.ID

	// Store memories
	memories := []memory.StoreMemoryInput{
		{ProjectID: pid, Content: "Go uses goroutines for lightweight concurrency", Category: "fact", Importance: 5, Tags: "go,concurrency"},
		{ProjectID: pid, Content: "Always use context.WithTimeout for HTTP client calls", Category: "advice", Importance: 4, Tags: "go,http,timeout"},
		{ProjectID: pid, Content: "Decided to use FTS5 for search instead of vector embeddings", Category: "decision", Importance: 5, Tags: "search,fts5"},
		{ProjectID: pid, Content: "User prefers minimal dependencies in Go projects", Category: "preference", Importance: 3, Tags: "go,dependencies"},
		{ProjectID: pid, Content: "Release v1.0 shipped on 2024-01-15", Category: "event", Importance: 2, Tags: "release"},
	}

	var storedIDs []int64
	for i, m := range memories {
		mem, err := memMgr.Store(ctx, m)
		if err != nil {
			t.Fatalf("failed to store memory %d: %v", i, err)
		}
		storedIDs = append(storedIDs, mem.ID)
		t.Logf("Stored memory %d: ID=%d, category=%s", i, mem.ID, mem.Category)
	}

	// Dedup test: store same content again → should return existing
	dupMem, err := memMgr.Store(ctx, memories[0])
	if err != nil {
		t.Fatalf("dedup store failed: %v", err)
	}
	if dupMem.ID != storedIDs[0] {
		t.Errorf("dedup failed: expected ID %d, got %d", storedIDs[0], dupMem.ID)
	}

	// Search
	results, err := memMgr.Search(ctx, pid, "goroutines", memory.SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results for 'goroutines'")
	}
	t.Logf("Search 'goroutines': %d results, top=%q", len(results), results[0].Content)

	// Search with category filter
	results, err = memMgr.Search(ctx, pid, "timeout", memory.SearchOptions{Category: "advice", Limit: 10})
	if err != nil {
		t.Fatalf("search with category failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results for 'timeout' in advice category")
	}

	// List all
	mems, err := memMgr.List(ctx, pid, memory.ListOptions{Limit: 50})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(mems) != 5 {
		t.Errorf("expected 5 memories, got %d", len(mems))
	}

	// List by category
	mems, err = memMgr.List(ctx, pid, memory.ListOptions{Category: "fact", Limit: 50})
	if err != nil {
		t.Fatalf("list by category failed: %v", err)
	}
	if len(mems) != 1 {
		t.Errorf("expected 1 fact memory, got %d", len(mems))
	}

	// Delete single memory
	err = memMgr.Delete(ctx, pid, storedIDs[4])
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	mems, err = memMgr.List(ctx, pid, memory.ListOptions{Limit: 50})
	if err != nil {
		t.Fatalf("list after delete failed: %v", err)
	}
	if len(mems) != 4 {
		t.Errorf("expected 4 memories after delete, got %d", len(mems))
	}

	// Project deletion cascades to memories
	err = queueMgr.DeleteProject(ctx, pid)
	if err != nil {
		t.Fatalf("delete project failed: %v", err)
	}
	mems, err = memMgr.List(ctx, pid, memory.ListOptions{Limit: 50})
	if err != nil {
		t.Fatalf("list after project delete failed: %v", err)
	}
	if len(mems) != 0 {
		t.Errorf("expected 0 memories after project delete, got %d", len(mems))
	}

	t.Log("Memory integration workflow completed successfully")
}
