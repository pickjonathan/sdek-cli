package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Cache manages AI analysis result caching with event-driven invalidation
type Cache struct {
	dir         string
	mu          sync.RWMutex
	eventHashes map[string]string // eventID -> hash for invalidation tracking
}

// NewCache creates a new cache instance
func NewCache(cacheDir string) (*Cache, error) {
	if cacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		cacheDir = filepath.Join(homeDir, ".sdek", "cache", "ai")
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{
		dir:         cacheDir,
		eventHashes: make(map[string]string),
	}, nil
}

// Get retrieves a cached result by key
func (c *Cache) Get(key string) (*CachedResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cachePath := c.getCachePath(key)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}

	var result CachedResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	return &result, nil
}

// Set stores a result in the cache
func (c *Cache) Set(key string, result *CachedResult) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cachePath := c.getCachePath(key)
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// Delete removes a cached result by key
func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cachePath := c.getCachePath(key)
	err := os.Remove(cachePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

// Clear removes all cached results
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			path := filepath.Join(c.dir, entry.Name())
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove cache file %s: %w", entry.Name(), err)
			}
		}
	}

	c.eventHashes = make(map[string]string)
	return nil
}

// GenerateKey creates a cache key from a request
func (c *Cache) GenerateKey(req *AnalysisRequest) string {
	h := sha256.New()

	// Include control ID and framework
	h.Write([]byte(req.ControlID))
	h.Write([]byte(req.Framework))
	h.Write([]byte(req.PolicyExcerpt))

	// Include all event IDs and content in sorted order
	for _, event := range req.Events {
		h.Write([]byte(event.EventID))
		h.Write([]byte(event.Source))
		h.Write([]byte(event.EventType))
		h.Write([]byte(event.Content))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// InvalidateByEvents removes cache entries that reference specific events
func (c *Cache) InvalidateByEvents(eventIDs []string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	invalidated := 0
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	eventSet := make(map[string]bool)
	for _, id := range eventIDs {
		eventSet[id] = true
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		cachePath := filepath.Join(c.dir, entry.Name())
		data, err := os.ReadFile(cachePath)
		if err != nil {
			continue // Skip files we can't read
		}

		var result CachedResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue // Skip files we can't parse
		}

		// Check if any evidence links reference the invalidated events
		shouldInvalidate := false
		for _, link := range result.Response.EvidenceLinks {
			if eventSet[link] {
				shouldInvalidate = true
				break
			}
		}

		if shouldInvalidate {
			if err := os.Remove(cachePath); err == nil {
				invalidated++
			}
		}
	}

	return invalidated, nil
}

// TrackEvent tracks an event's hash for invalidation detection
func (c *Cache) TrackEvent(event *types.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := c.hashEvent(event)
	c.eventHashes[event.ID] = hash
}

// DetectChangedEvents compares current events with tracked hashes
func (c *Cache) DetectChangedEvents(events []*types.Event) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var changed []string
	for _, event := range events {
		currentHash := c.hashEvent(event)
		if oldHash, exists := c.eventHashes[event.ID]; exists {
			if oldHash != currentHash {
				changed = append(changed, event.ID)
			}
		}
	}

	return changed
}

// Stats returns cache statistics
func (c *Cache) Stats() (CacheStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{}

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return stats, fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		stats.TotalEntries++

		info, err := entry.Info()
		if err != nil {
			continue
		}

		stats.TotalSize += info.Size()

		// Check if entry is older than 7 days
		if time.Since(info.ModTime()) > 7*24*time.Hour {
			stats.OldEntries++
		}
	}

	return stats, nil
}

// getCachePath returns the filesystem path for a cache key
func (c *Cache) getCachePath(key string) string {
	return filepath.Join(c.dir, key+".json")
}

// hashEvent creates a hash of an event's content for change detection
func (c *Cache) hashEvent(event *types.Event) string {
	h := sha256.New()
	h.Write([]byte(event.SourceID))
	h.Write([]byte(event.EventType))
	h.Write([]byte(event.Content))
	h.Write([]byte(event.Title))
	return hex.EncodeToString(h.Sum(nil))
}

// CacheStats contains cache statistics
type CacheStats struct {
	TotalEntries int
	TotalSize    int64
	OldEntries   int
}
