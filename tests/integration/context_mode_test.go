package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// TestContextModeE2E tests the complete context mode analysis workflow
// This corresponds to Scenario 1 in quickstart.md
func TestContextModeE2E(t *testing.T) {
	// Setup: Load SOC2 CC6.1 excerpt
	excerptsPath := filepath.Join("../../testdata/ai/policies/soc2_excerpts.json")
	excerptsData, err := os.ReadFile(excerptsPath)
	if err != nil {
		t.Fatalf("Failed to load excerpts: %v", err)
	}

	type excerptEntry struct {
		ControlID string `json:"control_id"`
		Title     string `json:"title"`
		Excerpt   string `json:"excerpt"`
	}

	var excerpts map[string]excerptEntry
	if err := json.Unmarshal(excerptsData, &excerpts); err != nil {
		t.Fatalf("Failed to parse excerpts: %v", err)
	}

	// Create context preamble
	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		excerpts["CC6.1"].Excerpt,
		[]string{"CC6.1"},
	)
	if err != nil {
		t.Fatalf("Failed to create preamble: %v", err)
	}

	// Parse timestamps
	ts1, _ := time.Parse(time.RFC3339, "2024-01-15T10:00:00Z")
	ts2, _ := time.Parse(time.RFC3339, "2024-01-10T14:30:00Z")

	// Create evidence bundle
	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:        "evt-001",
				Source:    "github",
				Type:      "commit",
				Timestamp: ts1,
				Content:   "Implemented MFA for admin access",
				Metadata: map[string]interface{}{
					"author": "user@example.com",
					"repo":   "auth-service",
				},
			},
			{
				ID:        "evt-002",
				Source:    "aws",
				Type:      "iam_policy",
				Timestamp: ts2,
				Content:   "Created IAM policy for least privilege access",
				Metadata: map[string]interface{}{
					"policy_name": "SecurityControls",
				},
			},
		},
	}

	// Create mock provider with high-confidence response
	mockProvider := ai.NewMockProvider()
	mockProvider.SetResponse(`{
		"confidence_score": 0.85,
		"residual_risk": "low",
		"finding_summary": "Access controls are properly implemented with MFA and least privilege",
		"citations": ["evt-001", "evt-002"],
		"justification": "Evidence shows MFA implementation and IAM policies follow security best practices"
	}`)

	// Create AI config
	cfg := &types.Config{
		AI: types.AIConfig{
			Provider: "anthropic",
			Mode:     "context",
			Concurrency: types.ConcurrencyLimits{
				MaxAnalyses: 25,
			},
			Budgets: types.BudgetLimits{
				MaxSources:  50,
				MaxAPICalls: 500,
				MaxTokens:   250000,
			},
			Redaction: types.RedactionConfig{
				Enabled:  true,
				Denylist: []string{},
			},
		},
	}

	// Create engine with mock provider
	engine := ai.NewEngine(cfg, mockProvider)

	// Execute context mode analysis
	ctx := context.Background()
	finding, err := engine.Analyze(ctx, *preamble, evidence)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	// Verify finding structure
	if finding == nil {
		t.Fatal("Expected finding, got nil")
	}

	// Verify finding has AI mode
	if finding.Mode != "ai" {
		t.Errorf("Expected mode 'ai', got '%s'", finding.Mode)
	}

	// Verify confidence score
	if finding.ConfidenceScore < 0.8 {
		t.Errorf("Expected high confidence (>= 0.8), got %.2f", finding.ConfidenceScore)
	}

	// Verify review not required for high confidence
	if finding.ReviewRequired {
		t.Error("Expected ReviewRequired to be false for high confidence")
	}

	// Verify residual risk
	if finding.ResidualRisk != "low" {
		t.Errorf("Expected residual risk 'low', got '%s'", finding.ResidualRisk)
	}

	// Verify citations exist
	if len(finding.Citations) == 0 {
		t.Error("Expected citations, got none")
	}
}

// TestContextModeCacheHit tests cache reuse on repeated analysis
func TestContextModeCacheHit(t *testing.T) {
	// Load excerpts
	excerptsPath := filepath.Join("../../testdata/ai/policies/soc2_excerpts.json")
	excerptsData, err := os.ReadFile(excerptsPath)
	if err != nil {
		t.Fatalf("Failed to load excerpts: %v", err)
	}

	type excerptEntry struct {
		ControlID string `json:"control_id"`
		Title     string `json:"title"`
		Excerpt   string `json:"excerpt"`
	}

	var excerpts map[string]excerptEntry
	if err := json.Unmarshal(excerptsData, &excerpts); err != nil {
		t.Fatalf("Failed to parse excerpts: %v", err)
	}

	// Create preamble
	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC1.1",
		excerpts["CC1.1"].Excerpt,
		[]string{"CC1.1"},
	)
	if err != nil {
		t.Fatalf("Failed to create preamble: %v", err)
	}

	// Parse timestamp
	ts, _ := time.Parse(time.RFC3339, "2024-01-20T09:00:00Z")

	// Create evidence
	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:        "evt-cache-001",
				Source:    "jira",
				Type:      "ticket",
				Timestamp: ts,
				Content:   "Security policy review ticket",
			},
		},
	}

	// Create mock provider
	mockProvider := ai.NewMockProvider()
	mockProvider.SetResponse(`{
		"confidence_score": 0.75,
		"residual_risk": "medium",
		"finding_summary": "Policy documentation exists",
		"citations": ["evt-cache-001"],
		"justification": "Evidence shows policy review process"
	}`)

	// Create engine
	cfg := &types.Config{
		AI: types.AIConfig{
			Provider: "anthropic",
			Mode:     "context",
			Redaction: types.RedactionConfig{
				Enabled:  true,
				Denylist: []string{},
			},
		},
	}

	engine := ai.NewEngine(cfg, mockProvider)

	ctx := context.Background()

	// First analysis (cache miss)
	finding1, err := engine.Analyze(ctx, *preamble, evidence)
	if err != nil {
		t.Fatalf("First analysis failed: %v", err)
	}

	// Verify provider was called
	callCount1 := mockProvider.GetCallCount()
	if callCount1 != 1 {
		t.Errorf("Expected 1 provider call, got %d", callCount1)
	}

	// Second analysis with same inputs (cache hit)
	finding2, err := engine.Analyze(ctx, *preamble, evidence)
	if err != nil {
		t.Fatalf("Second analysis failed: %v", err)
	}

	// Verify provider was NOT called again (cache hit)
	callCount2 := mockProvider.GetCallCount()
	if callCount2 != 1 {
		t.Errorf("Expected same call count (cache hit), got %d calls", callCount2)
	}

	// Verify findings are equivalent
	if finding1.ConfidenceScore != finding2.ConfidenceScore {
		t.Errorf("Expected same confidence scores, got %.2f and %.2f",
			finding1.ConfidenceScore, finding2.ConfidenceScore)
	}

	if finding1.ResidualRisk != finding2.ResidualRisk {
		t.Errorf("Expected same residual risk, got '%s' and '%s'",
			finding1.ResidualRisk, finding2.ResidualRisk)
	}
}

// TestContextModeRedaction verifies PII/secrets are redacted before AI transmission
func TestContextModeRedaction(t *testing.T) {
	// Load excerpts
	excerptsPath := filepath.Join("../../testdata/ai/policies/iso27001_excerpts.json")
	excerptsData, err := os.ReadFile(excerptsPath)
	if err != nil {
		t.Fatalf("Failed to load excerpts: %v", err)
	}

	type excerptEntry struct {
		ControlID string `json:"control_id"`
		Title     string `json:"title"`
		Excerpt   string `json:"excerpt"`
	}

	var excerpts map[string]excerptEntry
	if err := json.Unmarshal(excerptsData, &excerpts); err != nil {
		t.Fatalf("Failed to parse excerpts: %v", err)
	}

	// Create preamble
	preamble, err := types.NewContextPreamble(
		"ISO27001",
		"2013",
		"A.9.1",
		excerpts["A.9.1"].Excerpt,
		[]string{"A.9.1"},
	)
	if err != nil {
		t.Fatalf("Failed to create preamble: %v", err)
	}

	// Parse timestamp
	ts, _ := time.Parse(time.RFC3339, "2024-01-25T11:00:00Z")

	// Create evidence with PII
	evidence := types.EvidenceBundle{
		Events: []types.EvidenceEvent{
			{
				ID:        "evt-pii-001",
				Source:    "slack",
				Type:      "message",
				Timestamp: ts,
				Content:   "User john.doe@company.com requested access from IP 192.168.1.100",
				Metadata: map[string]interface{}{
					"channel": "#security",
				},
			},
		},
	}

	// Create mock provider that will receive redacted evidence
	mockProvider := ai.NewMockProvider()
	mockProvider.SetResponse(`{
		"confidence_score": 0.70,
		"residual_risk": "medium",
		"finding_summary": "Access request process documented",
		"citations": ["evt-pii-001"],
		"justification": "Evidence shows access request workflow"
	}`)

	// Create engine with redaction enabled
	cfg := &types.Config{
		AI: types.AIConfig{
			Provider: "anthropic",
			Mode:     "context",
			Redaction: types.RedactionConfig{
				Enabled:  true,
				Denylist: []string{"secret-token"},
			},
		},
	}

	engine := ai.NewEngine(cfg, mockProvider)

	ctx := context.Background()

	// Execute analysis
	finding, err := engine.Analyze(ctx, *preamble, evidence)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}

	// Verify finding was generated
	if finding == nil {
		t.Fatal("Expected finding, got nil")
	}

	// Verify the last prompt sent to provider doesn't contain PII
	lastPrompt := mockProvider.GetLastPrompt()
	if lastPrompt == "" {
		t.Fatal("Expected prompt to be captured, got empty string")
	}

	// Check that email and IP were redacted
	if containsPII(lastPrompt, "john.doe@company.com") {
		t.Error("Expected email to be redacted, but found in prompt")
	}

	if containsPII(lastPrompt, "192.168.1.100") {
		t.Error("Expected IP address to be redacted, but found in prompt")
	}

	// Verify finding was still generated successfully despite redaction
	if finding.ConfidenceScore == 0 {
		t.Error("Expected non-zero confidence score after redaction")
	}
}

// Helper function to check if PII exists in text
func containsPII(text, pii string) bool {
	return strings.Contains(text, pii)
}
