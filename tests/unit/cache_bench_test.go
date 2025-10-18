package unit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai"
)

// BenchmarkGenerateKey tests cache key generation performance
// Target: <20µs per key
func BenchmarkGenerateKey(b *testing.B) {
	cache := createBenchmarkCache(b)
	req := createSampleAnalysisRequest(10)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = cache.GenerateKey(req)
	}
}

// BenchmarkGenerateKey_LargeRequest tests with 100 events
func BenchmarkGenerateKey_LargeRequest(b *testing.B) {
	cache := createBenchmarkCache(b)
	req := createSampleAnalysisRequest(100)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = cache.GenerateKey(req)
	}
}

// BenchmarkCacheGet tests cache read performance
// Target: <100ms (but should be much faster)
func BenchmarkCacheGet(b *testing.B) {
	cache := createBenchmarkCache(b)
	req := createSampleAnalysisRequest(10)
	key := cache.GenerateKey(req)

	// Pre-populate cache
	result := &ai.CachedResult{
		CacheKey: key,
		Response: ai.AnalysisResponse{
			RequestID:     "test-request",
			EvidenceLinks: []string{"event1", "event2"},
			Justification: "Sample finding",
			Confidence:    85,
			ResidualRisk:  "Low",
			Provider:      "test",
			Model:         "test-model",
			TokensUsed:    100,
			Latency:       50,
			Timestamp:     time.Now(),
			CacheHit:      false,
		},
		CachedAt:     time.Now(),
		EventIDs:     []string{"event1", "event2"},
		ControlID:    "CC6.1",
		Provider:     "test",
		ModelVersion: "1.0",
	}
	_ = cache.Set(key, result)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(key)
	}
}

// BenchmarkCacheSet tests cache write performance
func BenchmarkCacheSet(b *testing.B) {
	cache := createBenchmarkCache(b)
	req := createSampleAnalysisRequest(10)
	key := cache.GenerateKey(req)

	result := &ai.CachedResult{
		CacheKey: key,
		Response: ai.AnalysisResponse{
			RequestID:     "test-request",
			EvidenceLinks: []string{"event1", "event2"},
			Justification: "Sample finding",
			Confidence:    85,
			ResidualRisk:  "Low",
			Provider:      "test",
			Model:         "test-model",
			TokensUsed:    100,
			Latency:       50,
			Timestamp:     time.Now(),
			CacheHit:      false,
		},
		CachedAt:     time.Now(),
		EventIDs:     []string{"event1", "event2"},
		ControlID:    "CC6.1",
		Provider:     "test",
		ModelVersion: "1.0",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = cache.Set(key, result)
	}
}

// BenchmarkCacheMiss tests performance when cache key doesn't exist
func BenchmarkCacheMiss(b *testing.B) {
	cache := createBenchmarkCache(b)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(fmt.Sprintf("nonexistent-key-%d", i))
	}
}

// BenchmarkSHA256_SmallInput tests SHA256 performance on small inputs
func BenchmarkSHA256_SmallInput(b *testing.B) {
	input := []byte("SOC2CC6.1The entity restricts logical access...")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		h := sha256.Sum256(input)
		_ = hex.EncodeToString(h[:])
	}
}

// BenchmarkSHA256_LargeInput tests SHA256 performance on large inputs (10KB)
func BenchmarkSHA256_LargeInput(b *testing.B) {
	input := make([]byte, 10240)
	for i := range input {
		input[i] = byte(i % 256)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		h := sha256.Sum256(input)
		_ = hex.EncodeToString(h[:])
	}
}

// BenchmarkCacheInvalidation tests performance of event-based cache invalidation
func BenchmarkCacheInvalidation(b *testing.B) {
	cache := createBenchmarkCache(b)

	// Pre-populate cache with 100 entries
	for i := 0; i < 100; i++ {
		req := createSampleAnalysisRequest(5)
		key := cache.GenerateKey(req)
		result := &ai.CachedResult{
			CacheKey: key,
			Response: ai.AnalysisResponse{
				RequestID:     fmt.Sprintf("request-%d", i),
				EvidenceLinks: []string{fmt.Sprintf("event%d", i)},
				Justification: fmt.Sprintf("Finding %d", i),
				Confidence:    85,
				Provider:      "test",
				Model:         "test-model",
				Timestamp:     time.Now(),
			},
			CachedAt:     time.Now(),
			EventIDs:     []string{fmt.Sprintf("event%d", i)},
			ControlID:    "CC6.1",
			Provider:     "test",
			ModelVersion: "1.0",
		}
		_ = cache.Set(key, result)
	}

	eventIDs := []string{"event10", "event20", "event30"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = cache.InvalidateByEvents(eventIDs)
	}
}

// Helper functions

func createBenchmarkCache(b *testing.B) *ai.Cache {
	tmpDir, err := os.MkdirTemp("", "cache-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	b.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	cache, err := ai.NewCache(tmpDir)
	if err != nil {
		b.Fatalf("Failed to create cache: %v", err)
	}

	return cache
}

func createSampleAnalysisRequest(numEvents int) *ai.AnalysisRequest {
	events := make([]ai.AnalysisEvent, numEvents)
	for i := 0; i < numEvents; i++ {
		events[i] = ai.AnalysisEvent{
			EventID:     fmt.Sprintf("event-%d", i),
			EventType:   "commit",
			Source:      "git",
			Description: "Authentication changes",
			Content:     fmt.Sprintf("Commit %d: Fixed authentication logic", i),
			Timestamp:   time.Now(),
		}
	}

	return &ai.AnalysisRequest{
		RequestID:     "test-request",
		ControlID:     "CC6.1",
		ControlName:   "Logical Access Controls",
		Framework:     "SOC2",
		PolicyExcerpt: "The entity restricts logical and physical access to systems...",
		Events:        events,
		Timestamp:     time.Now(),
	}
}
