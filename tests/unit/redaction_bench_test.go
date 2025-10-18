package unit

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// BenchmarkRedact_1KB tests redaction performance on a 1KB text
// Target: <10ms per event (10,000 µs)
func BenchmarkRedact_1KB(b *testing.B) {
	// Create a 1KB sample text with various PII types
	text := generateSampleText(1024)

	config := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:  true,
				Denylist: []string{"secret123", "password456"},
			},
		},
	}

	redactor := ai.NewRedactor(config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = redactor.Redact(text)
	}
}

// BenchmarkRedact_10KB tests redaction performance on a 10KB text
func BenchmarkRedact_10KB(b *testing.B) {
	text := generateSampleText(10240)

	config := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:  true,
				Denylist: []string{"secret123", "password456"},
			},
		},
	}

	redactor := ai.NewRedactor(config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = redactor.Redact(text)
	}
}

// BenchmarkRedact_100Events tests redaction of 100 typical events
// Target: <1s total (100 events * 10ms = 1000ms)
func BenchmarkRedact_100Events(b *testing.B) {
	events := generateSampleEvents(100)

	config := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:  true,
				Denylist: []string{"secret123", "password456"},
			},
		},
	}

	redactor := ai.NewRedactor(config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, event := range events {
			eventJSON, _ := json.Marshal(event)
			_, _, _ = redactor.Redact(string(eventJSON))
		}
	}
}

// BenchmarkRedact_EmailOnly tests email-only redaction performance
func BenchmarkRedact_EmailOnly(b *testing.B) {
	text := strings.Repeat("Contact us at user@example.com or admin@company.org. ", 100)

	config := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:  true,
				Denylist: []string{},
			},
		},
	}

	redactor := ai.NewRedactor(config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = redactor.Redact(text)
	}
}

// BenchmarkRedact_MixedPII tests redaction with mixed PII types
func BenchmarkRedact_MixedPII(b *testing.B) {
	text := `
		User: john.doe@example.com
		Phone: +1-555-123-4567
		IP: 192.168.1.100
		AWS Key: AKIAIOSFODNN7EXAMPLE
		API Key: sk_test_REDACTED_FOR_TESTING
		Secret: password123
	`
	text = strings.Repeat(text, 10)

	config := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:  true,
				Denylist: []string{"password123"},
			},
		},
	}

	redactor := ai.NewRedactor(config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = redactor.Redact(text)
	}
}

// BenchmarkRedact_NoMatches tests performance when no PII is present
func BenchmarkRedact_NoMatches(b *testing.B) {
	text := strings.Repeat("This is clean text with no PII or secrets. ", 100)

	config := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:  true,
				Denylist: []string{},
			},
		},
	}

	redactor := ai.NewRedactor(config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = redactor.Redact(text)
	}
}

// BenchmarkRedactor_Creation tests the cost of creating a new redactor
func BenchmarkRedactor_Creation(b *testing.B) {
	config := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:  true,
				Denylist: []string{"secret1", "secret2", "secret3"},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ai.NewRedactor(config)
	}
}

// Helper functions

func generateSampleText(sizeBytes int) string {
	// Generate text with various PII patterns
	template := `
		User Report #%d:
		Name: John Doe
		Email: john.doe%d@example.com
		Phone: +1-555-%03d-%04d
		IP Address: 192.168.%d.%d
		AWS Access Key: AKIAIOSFODNN7EXAMPLE%d
		Comments: This is a sample comment about the user's activity.
		Timestamp: 2025-10-18T10:00:00Z
		Status: active
		Notes: Regular user with no issues.
	`

	var sb strings.Builder
	count := 0

	for sb.Len() < sizeBytes {
		count++
		sb.WriteString(template)
	}

	result := sb.String()
	if len(result) > sizeBytes {
		return result[:sizeBytes]
	}
	return result
}

func generateSampleEvents(count int) []types.EvidenceEvent {
	events := make([]types.EvidenceEvent, count)

	baseTime := time.Date(2025, 10, 18, 10, 0, 0, 0, time.UTC)

	for i := 0; i < count; i++ {
		events[i] = types.EvidenceEvent{
			ID:     generateEventID(i),
			Source: "github",
			Type:   "commit",
			Metadata: map[string]interface{}{
				"author":  "john.doe@example.com",
				"message": "Fixed authentication bug in login flow",
				"sha":     generateSHA(i),
				"repo":    "company/auth-service",
			},
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
		}
	}

	return events
}

func generateEventID(index int) string {
	return strings.Repeat("e", 32) + string(rune('0'+index%10))
}

func generateSHA(index int) string {
	return strings.Repeat("a", 40) + string(rune('0'+index%10))
}
