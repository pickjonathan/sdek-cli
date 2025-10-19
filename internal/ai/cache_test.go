package ai

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestCache_GetSet(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	tmpDir := t.TempDir()
	cache, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	// Test cache miss
	result, err := cache.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result != nil {
		t.Errorf("Expected cache miss, got result: %+v", result)
	}

	// Test cache hit
	key := "test-key"
	expected := &CachedResult{
		CacheKey: key,
		Response: AnalysisResponse{RequestID: "req-1", Confidence: 85},
		CachedAt: time.Now(),
	}

	if err := cache.Set(key, expected); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	result, err = cache.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected cache hit, got nil")
	}
	if result.Response.Confidence != 85 {
		t.Errorf("Expected confidence 85, got %d", result.Response.Confidence)
	}
}

func TestCache_GenerateKey(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	tmpDir := t.TempDir()
	cache, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	req1 := &AnalysisRequest{
		ControlID:     "SOC2-CC1.1",
		Framework:     "SOC2",
		PolicyExcerpt: "Test policy",
		Events: []AnalysisEvent{
			{EventID: "event-1", Content: "content-1"},
		},
	}

	req2 := &AnalysisRequest{
		ControlID:     "SOC2-CC1.1",
		Framework:     "SOC2",
		PolicyExcerpt: "Test policy",
		Events: []AnalysisEvent{
			{EventID: "event-1", Content: "content-1"},
		},
	}

	req3 := &AnalysisRequest{
		ControlID:     "SOC2-CC1.1",
		Framework:     "SOC2",
		PolicyExcerpt: "Test policy",
		Events: []AnalysisEvent{
			{EventID: "event-1", Content: "content-2"}, // Different content
		},
	}

	key1 := cache.GenerateKey(req1)
	key2 := cache.GenerateKey(req2)
	key3 := cache.GenerateKey(req3)

	if key1 != key2 {
		t.Errorf("Expected identical keys for identical requests")
	}
	if key1 == key3 {
		t.Errorf("Expected different keys for different content")
	}
}

func TestCache_InvalidateByEvents(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	tmpDir := t.TempDir()
	cache, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	// Store two results
	result1 := &CachedResult{
		CacheKey: "key-1",
		Response: AnalysisResponse{
			RequestID:     "req-1",
			EvidenceLinks: []string{"event-1", "event-2"},
		},
		CachedAt: time.Now(),
	}

	result2 := &CachedResult{
		CacheKey: "key-2",
		Response: AnalysisResponse{
			RequestID:     "req-2",
			EvidenceLinks: []string{"event-3"},
		},
		CachedAt: time.Now(),
	}

	if err := cache.Set("key-1", result1); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if err := cache.Set("key-2", result2); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Invalidate event-1
	count, err := cache.InvalidateByEvents([]string{"event-1"})
	if err != nil {
		t.Fatalf("InvalidateByEvents failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 invalidation, got %d", count)
	}

	// Check result1 is gone, result2 remains
	res1, _ := cache.Get("key-1")
	if res1 != nil {
		t.Errorf("Expected key-1 to be invalidated")
	}

	res2, _ := cache.Get("key-2")
	if res2 == nil {
		t.Errorf("Expected key-2 to remain")
	}
}

func TestCache_Clear(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	tmpDir := t.TempDir()
	cache, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	// Store multiple results
	for i := 0; i < 3; i++ {
		key := string(rune('a' + i))
		result := &CachedResult{
			CacheKey: key,
			Response: AnalysisResponse{RequestID: key},
			CachedAt: time.Now(),
		}
		if err := cache.Set(key, result); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	// Clear all
	if err := cache.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Check all are gone
	stats, err := cache.Stats()
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}
	if stats.TotalEntries != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", stats.TotalEntries)
	}
}

func TestCache_TrackEvent(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	tmpDir := t.TempDir()
	cache, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	event := &types.Event{
		ID:        "event-1",
		SourceID:  "git",
		EventType: "commit",
		Title:     "Test commit",
		Content:   "Test content",
	}

	// Track event
	cache.TrackEvent(event)

	// Verify no changes detected
	changed := cache.DetectChangedEvents([]*types.Event{event})
	if len(changed) != 0 {
		t.Errorf("Expected no changes, got %d", len(changed))
	}

	// Modify event
	event.Content = "Modified content"

	// Verify change detected
	changed = cache.DetectChangedEvents([]*types.Event{event})
	if len(changed) != 1 {
		t.Errorf("Expected 1 change, got %d", len(changed))
	}
	if changed[0] != "event-1" {
		t.Errorf("Expected event-1 to be changed, got %s", changed[0])
	}
}

func TestCache_Stats(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	tmpDir := t.TempDir()
	cache, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	// Empty cache
	stats, err := cache.Stats()
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}
	if stats.TotalEntries != 0 {
		t.Errorf("Expected 0 entries, got %d", stats.TotalEntries)
	}

	// Add entries
	for i := 0; i < 5; i++ {
		key := string(rune('a' + i))
		result := &CachedResult{
			CacheKey: key,
			Response: AnalysisResponse{RequestID: key},
			CachedAt: time.Now(),
		}
		if err := cache.Set(key, result); err != nil {
			t.Fatalf("Set failed: %v", err)
		}
	}

	stats, err = cache.Stats()
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}
	if stats.TotalEntries != 5 {
		t.Errorf("Expected 5 entries, got %d", stats.TotalEntries)
	}
	if stats.TotalSize == 0 {
		t.Errorf("Expected non-zero total size")
	}
}

func TestCache_Delete(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	tmpDir := t.TempDir()
	cache, err := NewCache(tmpDir)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	key := "test-key"
	result := &CachedResult{
		CacheKey: key,
		Response: AnalysisResponse{RequestID: "req-1"},
		CachedAt: time.Now(),
	}

	// Set and verify
	if err := cache.Set(key, result); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	res, _ := cache.Get(key)
	if res == nil {
		t.Fatal("Expected cache hit")
	}

	// Delete
	if err := cache.Delete(key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	res, _ = cache.Get(key)
	if res != nil {
		t.Errorf("Expected cache miss after delete")
	}

	// Delete non-existent should not error
	if err := cache.Delete("nonexistent"); err != nil {
		t.Errorf("Delete of nonexistent key should not error: %v", err)
	}
}

func TestCache_DefaultDirectory(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	// Create cache with empty dir (should use default)
	cache, err := NewCache("")
	if err != nil {
		t.Fatalf("NewCache with default dir failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, ".sdek", "cache", "ai")

	if cache.dir != expectedDir {
		t.Errorf("Expected default dir %s, got %s", expectedDir, cache.dir)
	}
}
