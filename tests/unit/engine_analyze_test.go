package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T008: Contract test for Engine.Analyze (context mode)
// These tests define the contract for AI analysis with context injection
// EXPECTED: These tests MUST FAIL until Engine.Analyze is extended in Phase 3.3

func TestAnalyze_ValidContextMode(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
			Redaction: types.RedactionConfig{
				Enabled: true,
			},
		},
	}
	engine := ai.NewEngine(cfg, ai.NewMockProvider())

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data. This includes implementing role-based access controls, multi-factor authentication, and regular access reviews.",
		[]string{"CC6.1", "CC6.2"},
	)
	require.NoError(t, err)

	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:        "evt-1",
				Source:    "github",
				Timestamp: time.Now().Add(-24 * time.Hour),
				Type:      "commit",
				Content:   "Added MFA authentication to login endpoint",
			},
			{
				ID:        "evt-2",
				Source:    "jira",
				Timestamp: time.Now().Add(-12 * time.Hour),
				Type:      "ticket",
				Content:   "SEC-123: Implement role-based access control",
			},
		},
	}

	// Act
	ctx := context.Background()
	finding, err := engine.Analyze(ctx, *preamble, evidence)

	// Assert
	require.NoError(t, err, "Analyze should succeed with valid inputs")
	require.NotNil(t, finding)

	// Verify required fields
	assert.NotEmpty(t, finding.Summary, "Finding should have summary")
	assert.NotEmpty(t, finding.MappedControls, "Finding should have mapped controls")
	assert.GreaterOrEqual(t, finding.ConfidenceScore, 0.0, "Confidence should be >= 0.0")
	assert.LessOrEqual(t, finding.ConfidenceScore, 1.0, "Confidence should be <= 1.0")
	assert.Contains(t, []string{"low", "medium", "high"}, finding.ResidualRisk, "Risk should be valid")
	assert.NotEmpty(t, finding.Justification, "Finding should have justification")
	assert.NotEmpty(t, finding.Citations, "Finding should have citations")
	assert.Equal(t, "ai", finding.Mode, "Mode should be 'ai'")

	// Verify review flag logic
	if finding.ConfidenceScore < 0.6 {
		assert.True(t, finding.ReviewRequired, "ReviewRequired should be true when confidence < 0.6")
	}
}

func TestAnalyze_SetsModeToAI(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
		},
	}
	engine := ai.NewEngine(cfg, ai.NewMockProvider())

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:      "evt-1",
				Source:  "github",
				Content: "Added authentication",
			},
		},
	}

	// Act
	finding, err := engine.Analyze(context.Background(), *preamble, evidence)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "ai", finding.Mode, "Mode should be set to 'ai' on success")
}

func TestAnalyze_SetsReviewRequiredWhenLowConfidence(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
		},
	}
	mockProvider := ai.NewMockProvider()
	// Configure mock to return low confidence
	mockProvider.SetConfidenceScore(0.4) // Below 0.6 threshold
	engine := ai.NewEngine(cfg, mockProvider)

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:      "evt-1",
				Source:  "github",
				Content: "Added authentication",
			},
		},
	}

	// Act
	finding, err := engine.Analyze(context.Background(), *preamble, evidence)

	// Assert
	require.NoError(t, err)
	assert.True(t, finding.ReviewRequired, "ReviewRequired should be true when confidence < 0.6")
	assert.Less(t, finding.ConfidenceScore, 0.6, "Confidence should be < 0.6")
}

func TestAnalyze_RedactsPIIBeforeSending(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
			Redaction: types.RedactionConfig{
				Enabled: true,
			},
		},
	}
	mockProvider := ai.NewMockProvider()
	engine := ai.NewEngine(cfg, mockProvider)

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:      "evt-1",
				Source:  "github",
				Content: "User admin@company.com added authentication at IP 192.168.1.1",
			},
		},
	}

	// Act
	_, err = engine.Analyze(context.Background(), *preamble, evidence)

	// Assert
	require.NoError(t, err)

	// Verify provider received redacted content
	lastPrompt := mockProvider.GetLastPrompt()
	assert.NotContains(t, lastPrompt, "admin@company.com", "Email should be redacted")
	assert.NotContains(t, lastPrompt, "192.168.1.1", "IP should be redacted")
	assert.Contains(t, lastPrompt, "[REDACTED:PII:", "Redaction placeholder should be present")
}

func TestAnalyze_UsesCacheOnSecondCall(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
			CacheDir: t.TempDir(),
		},
	}
	mockProvider := ai.NewMockProvider()
	engine := ai.NewEngine(cfg, mockProvider)

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:      "evt-1",
				Source:  "github",
				Content: "Added authentication",
			},
		},
	}

	// Act - first call
	finding1, err := engine.Analyze(context.Background(), *preamble, evidence)
	require.NoError(t, err)
	callCount1 := mockProvider.GetCallCount()

	// Act - second call with same inputs
	finding2, err := engine.Analyze(context.Background(), *preamble, evidence)
	require.NoError(t, err)
	callCount2 := mockProvider.GetCallCount()

	// Assert
	assert.Equal(t, finding1.Summary, finding2.Summary, "Cached result should be identical")
	assert.Equal(t, callCount1, callCount2, "Provider should not be called again (cache hit)")
}

func TestAnalyze_RespectsNoCacheFlag(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
			CacheDir: t.TempDir(),
			NoCache:  true, // Cache disabled
		},
	}
	mockProvider := ai.NewMockProvider()
	engine := ai.NewEngine(cfg, mockProvider)

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:      "evt-1",
				Source:  "github",
				Content: "Added authentication",
			},
		},
	}

	// Act - two calls with same inputs
	_, err = engine.Analyze(context.Background(), *preamble, evidence)
	require.NoError(t, err)
	callCount1 := mockProvider.GetCallCount()

	_, err = engine.Analyze(context.Background(), *preamble, evidence)
	require.NoError(t, err)
	callCount2 := mockProvider.GetCallCount()

	// Assert
	assert.Greater(t, callCount2, callCount1, "Provider should be called again when cache disabled")
}

func TestAnalyze_InjectsPreambleIntoPrompt(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
		},
	}
	mockProvider := ai.NewMockProvider()
	engine := ai.NewEngine(cfg, mockProvider)

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:      "evt-1",
				Source:  "github",
				Content: "Added authentication",
			},
		},
	}

	// Act
	_, err = engine.Analyze(context.Background(), *preamble, evidence)

	// Assert
	require.NoError(t, err)
	lastPrompt := mockProvider.GetLastPrompt()
	assert.Contains(t, lastPrompt, "SOC2", "Prompt should contain framework")
	assert.Contains(t, lastPrompt, "CC6.1", "Prompt should contain section")
	assert.Contains(t, lastPrompt, "Access controls shall be implemented", "Prompt should contain excerpt")
}

func TestAnalyze_EmptyEvidenceReturnsLowConfidence(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
		},
	}
	mockProvider := ai.NewMockProvider()
	engine := ai.NewEngine(cfg, mockProvider)

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{}, // Empty evidence
	}

	// Act
	finding, err := engine.Analyze(context.Background(), *preamble, evidence)

	// Assert
	require.NoError(t, err)
	assert.Less(t, finding.ConfidenceScore, 0.6, "Empty evidence should result in low confidence")
	assert.True(t, finding.ReviewRequired, "Empty evidence should require review")
}

func TestAnalyze_InvalidPreambleReturnsError(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
		},
	}
	engine := ai.NewEngine(cfg, ai.NewMockProvider())

	// Invalid preamble (empty framework)
	invalidPreamble := types.ContextPreamble{
		Framework: "", // Invalid
		Version:   "2017",
		Section:   "CC6.1",
		Excerpt:   "Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
	}

	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:      "evt-1",
				Source:  "github",
				Content: "Added authentication",
			},
		},
	}

	// Act
	_, err := engine.Analyze(context.Background(), invalidPreamble, evidence)

	// Assert
	require.Error(t, err, "Should return error for invalid preamble")
	assert.Contains(t, err.Error(), "preamble", "Error should mention preamble")
}

func TestAnalyze_ProviderErrorReturnsFallbackError(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
		},
	}
	mockProvider := ai.NewMockProvider()
	mockProvider.SetError(ai.ErrProviderUnavailable)
	engine := ai.NewEngine(cfg, mockProvider)

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:      "evt-1",
				Source:  "github",
				Content: "Added authentication",
			},
		},
	}

	// Act
	_, err = engine.Analyze(context.Background(), *preamble, evidence)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ai.ErrProviderUnavailable, "Should return provider error")
}

func TestAnalyze_ContextCancellationReturnsError(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
		},
	}
	engine := ai.NewEngine(cfg, ai.NewMockProvider())

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:      "evt-1",
				Source:  "github",
				Content: "Added authentication",
			},
		},
	}

	// Act - cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = engine.Analyze(ctx, *preamble, evidence)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context", "Error should mention context")
}

func TestAnalyze_PerformanceTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
		},
	}
	engine := ai.NewEngine(cfg, ai.NewMockProvider())

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	// Create typical evidence bundle (~100 events)
	events := make([]types.EvidenceEvent, 100)
	for i := 0; i < 100; i++ {
		events[i] = types.EvidenceEvent{
			ID:      fmt.Sprintf("evt-%d", i),
			Source:  "github",
			Content: fmt.Sprintf("Event %d: security update", i),
		}
	}
	evidence := types.EvidenceBundle{Events: events}

	// Act
	start := time.Now()
	_, err = engine.Analyze(context.Background(), *preamble, evidence)
	duration := time.Since(start)

	// Assert
	require.NoError(t, err)
	assert.Less(t, duration, 30*time.Second, "Analyze should complete in <30s for 100 events")
	t.Logf("Analyze took %v for 100 events", duration)
}

func TestAnalyze_CacheHitPerformance(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeContext,
			CacheDir: t.TempDir(),
		},
	}
	engine := ai.NewEngine(cfg, ai.NewMockProvider())

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:      "evt-1",
				Source:  "github",
				Content: "Added authentication",
			},
		},
	}

	// Prime cache
	_, err = engine.Analyze(context.Background(), *preamble, evidence)
	require.NoError(t, err)

	// Act - cached call
	start := time.Now()
	_, err = engine.Analyze(context.Background(), *preamble, evidence)
	duration := time.Since(start)

	// Assert
	require.NoError(t, err)
	assert.Less(t, duration, 100*time.Millisecond, "Cache hit should take <100ms")
	t.Logf("Cache hit took %v", duration)
}
