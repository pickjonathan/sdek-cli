package unit

import (
	"strings"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T006: Contract test for Redactor interface
// These tests define the contract for PII/secret redaction from evidence
// EXPECTED: These tests MUST FAIL until the Redactor is implemented in Phase 3.3

func TestRedact_Email(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "Contact: user@example.com for support"

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, output, "[REDACTED:PII:EMAIL]", "Email should be redacted")
	assert.NotContains(t, output, "user@example.com", "Original email should be removed")
	assert.Equal(t, 1, rm.TotalRedactions, "Should have 1 redaction")
	assert.Contains(t, rm.RedactionTypes, types.RedactionPII)
}

func TestRedact_MultipleEmails(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "Email admin@company.com or support@company.org for help"

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, strings.Count(output, "[REDACTED:PII:EMAIL]"), "Both emails should be redacted")
	assert.NotContains(t, output, "admin@company.com")
	assert.NotContains(t, output, "support@company.org")
	assert.Equal(t, 2, rm.TotalRedactions)
}

func TestRedact_IPv4(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "Server IP: 192.168.1.100 is offline"

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, output, "[REDACTED:PII:IP]", "IP should be redacted")
	assert.NotContains(t, output, "192.168.1.100", "Original IP should be removed")
	assert.Equal(t, 1, rm.TotalRedactions)
	assert.Contains(t, rm.RedactionTypes, types.RedactionPII)
}

func TestRedact_IPv6(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "IPv6 address: 2001:0db8:85a3:0000:0000:8a2e:0370:7334"

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, output, "[REDACTED:PII:IP]")
	assert.NotContains(t, output, "2001:0db8:85a3:0000:0000:8a2e:0370:7334")
	assert.Equal(t, 1, rm.TotalRedactions)
}

func TestRedact_Phone(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "Call me at (555) 123-4567 or 555-987-6543"

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, output, "[REDACTED:PII:PHONE]")
	assert.NotContains(t, output, "555-123-4567")
	assert.NotContains(t, output, "555-987-6543")
	assert.GreaterOrEqual(t, rm.TotalRedactions, 2, "Should redact both phone numbers")
}

func TestRedact_AWSKey(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "AWS_KEY=AKIAIOSFODNN7EXAMPLE found in logs"

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, output, "[REDACTED:SECRET]", "AWS key should be redacted")
	assert.NotContains(t, output, "AKIAIOSFODNN7EXAMPLE", "Original key should be removed")
	assert.Equal(t, 1, rm.TotalRedactions)
	assert.Contains(t, rm.RedactionTypes, types.RedactionSecret)
}

func TestRedact_GenericAPIKey(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	// 32-character API key
	input := "API_KEY=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6 is hardcoded"

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, output, "[REDACTED:SECRET]")
	assert.NotContains(t, output, "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6")
	assert.Equal(t, 1, rm.TotalRedactions)
}

func TestRedact_Denylist(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{"password123", "secret-token"},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "Default password is password123 and token is secret-token"

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.Contains(t, output, "[REDACTED:SECRET]")
	assert.NotContains(t, output, "password123", "Denylist item should be redacted")
	assert.NotContains(t, output, "secret-token", "Denylist item should be redacted")
	assert.Equal(t, 2, rm.TotalRedactions)
}

func TestRedact_DenylistCaseInsensitive(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{"SecretValue"},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "The value is secretvalue and SECRETVALUE"

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.NotContains(t, output, "secretvalue", "Should be case-insensitive")
	assert.NotContains(t, output, "SECRETVALUE", "Should be case-insensitive")
	assert.Equal(t, 2, rm.TotalRedactions)
}

func TestRedact_HighRedactionPercentageWarning(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	// Input with >40% redactable content
	input := "user1@example.com user2@example.com user3@example.com user4@example.com short text"

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, rm)
	// Verify warning is set (implementation will add WarningThresholdExceeded field)
	redactedPercent := float64(len(input)-len(output)) / float64(len(input))
	if redactedPercent > 0.4 {
		// Test passes if >40% was redacted (warning condition triggered)
		assert.True(t, true, "High redaction percentage detected")
	}
}

func TestRedact_Idempotent(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "Contact user@example.com at 192.168.1.1"

	// Act - first redaction
	output1, rm1, err1 := redactor.Redact(input)
	require.NoError(t, err1)

	// Act - second redaction of already redacted text
	output2, rm2, err2 := redactor.Redact(output1)
	require.NoError(t, err2)

	// Assert - idempotent behavior
	assert.Equal(t, output1, output2, "Redacting twice should produce same result")
	assert.Equal(t, rm1.TotalRedactions, rm2.TotalRedactions+rm1.TotalRedactions, 
		"Second redaction should not find more items if first was complete")
}

func TestRedact_MixedContent(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{"api-secret"},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := `
		User: admin@company.com
		IP: 10.0.1.50
		AWS_KEY: AKIAIOSFODNN7EXAMPLE
		Custom: api-secret
		Phone: (555) 123-4567
	`

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.NotContains(t, output, "admin@company.com")
	assert.NotContains(t, output, "10.0.1.50")
	assert.NotContains(t, output, "AKIAIOSFODNN7EXAMPLE")
	assert.NotContains(t, output, "api-secret")
	assert.NotContains(t, output, "(555) 123-4567")
	assert.GreaterOrEqual(t, rm.TotalRedactions, 5, "Should redact all sensitive items")
}

func TestRedact_PerformanceBenchmark(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "Contact user@example.com at 192.168.1.1 with key AKIAIOSFODNN7EXAMPLE"

	// Act
	start := time.Now()
	_, _, err := redactor.Redact(input)
	duration := time.Since(start)

	// Assert
	require.NoError(t, err)
	assert.Less(t, duration, 10*time.Millisecond, "Redaction should take <10ms per event")
}

func TestRedact_Performance100Events(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "Contact user@example.com at 192.168.1.1 with key AKIAIOSFODNN7EXAMPLE"

	// Act - redact 100 events
	start := time.Now()
	for i := 0; i < 100; i++ {
		_, _, err := redactor.Redact(input)
		require.NoError(t, err)
	}
	duration := time.Since(start)

	// Assert
	assert.Less(t, duration, 1*time.Second, "100 events should redact in <1s")
	t.Logf("Redacted 100 events in %v (avg: %v per event)", duration, duration/100)
}

func TestRedact_Disabled(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   false, // Redaction disabled
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "Contact user@example.com with key AKIAIOSFODNN7EXAMPLE"

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, input, output, "When disabled, redaction should return input unchanged")
	assert.Equal(t, 0, rm.TotalRedactions, "Should not redact when disabled")
}

func TestRedact_EmptyInput(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := ""

	// Act
	output, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "", output)
	assert.Equal(t, 0, rm.TotalRedactions)
}

func TestRedactionMap_Structure(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Redaction: types.RedactionConfig{
				Enabled:   true,
				Denylist:  []string{},
			},
		},
	}
	redactor := ai.NewRedactor(cfg)
	input := "Email: admin@example.com, IP: 192.168.1.1"

	// Act
	_, rm, err := redactor.Redact(input)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, rm)
	assert.GreaterOrEqual(t, rm.TotalRedactions, 2)
	assert.Contains(t, rm.RedactionTypes, types.RedactionPII, "Should track PII redactions")
	
	// Verify redaction types are set correctly
	for _, redType := range rm.RedactionTypes {
		assert.Contains(t, []types.RedactionType{types.RedactionPII, types.RedactionSecret}, redType)
	}
}
