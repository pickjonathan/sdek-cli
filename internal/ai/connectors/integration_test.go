package connectors_test

import (
	"context"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/internal/ai/connectors"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// TestMockConnectorIntegration demonstrates the full connector framework
// working with Engine.ExecutePlan using a mock connector.
func TestMockConnectorIntegration(t *testing.T) {
	// Step 1: Create a registry with mock connectors
	registry := connectors.NewRegistry()

	// Register GitHub mock connector
	githubCfg := connectors.DefaultConfig()
	githubCfg.APIKey = "mock-github-token"
	githubCfg.Extra = map[string]interface{}{"name": "github"}

	githubConnector, err := connectors.NewMockConnector(githubCfg)
	if err != nil {
		t.Fatalf("failed to create github mock: %v", err)
	}

	// Configure mock to return specific events
	mockGitHub := githubConnector.(*connectors.MockConnector)
	mockGitHub.SetEvents("auth* security", []types.EvidenceEvent{
		{
			ID:        "github-pr-1234",
			Source:    "github",
			Type:      "pr",
			Timestamp: time.Now(),
			Content:   "Added OAuth2 authentication with MFA support",
			Metadata: map[string]interface{}{
				"title":  "Implement OAuth2 authentication",
				"labels": []string{"security", "authentication"},
				"state":  "merged",
			},
		},
		{
			ID:        "github-commit-5678",
			Source:    "github",
			Type:      "commit",
			Timestamp: time.Now(),
			Content:   "Added security headers to API responses",
			Metadata: map[string]interface{}{
				"sha":    "abc123def456",
				"author": "security-team",
			},
		},
	})

	if err := registry.Register(githubConnector); err != nil {
		t.Fatalf("failed to register github: %v", err)
	}

	// Register Jira mock connector
	jiraCfg := connectors.DefaultConfig()
	jiraCfg.APIKey = "mock-jira-token"
	jiraCfg.Extra = map[string]interface{}{"name": "jira"}

	jiraConnector, err := connectors.NewMockConnector(jiraCfg)
	if err != nil {
		t.Fatalf("failed to create jira mock: %v", err)
	}

	mockJira := jiraConnector.(*connectors.MockConnector)
	mockJira.SetEvents("project=INFOSEC", []types.EvidenceEvent{
		{
			ID:        "jira-INFOSEC-123",
			Source:    "jira",
			Type:      "ticket",
			Timestamp: time.Now(),
			Content:   "Implement password rotation policy for service accounts",
			Metadata: map[string]interface{}{
				"key":      "INFOSEC-123",
				"status":   "Done",
				"priority": "High",
			},
		},
	})

	if err := registry.Register(jiraConnector); err != nil {
		t.Fatalf("failed to register jira: %v", err)
	}

	// Step 2: Create a config for the engine
	config := &types.Config{
		AI: types.AIConfig{
			Provider: "mock",
			Model:    "mock-model",
		},
	}

	// Step 3: Create a mock AI provider
	mockProvider := ai.NewMockProvider()

	// Step 4: Create engine with the connector registry
	engine := ai.NewEngineWithConnector(config, mockProvider, registry)

	// Step 5: Create an evidence collection plan
	plan := &types.EvidencePlan{
		ID:        "test-plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{
				Source:          "github",
				Query:           "auth* security",
				SignalStrength:  0.9,
				Rationale:       "Authentication changes show access control implementation",
				ApprovalStatus:  types.ApprovalApproved,
				ExecutionStatus: types.ExecPending,
			},
			{
				Source:          "jira",
				Query:           "project=INFOSEC",
				SignalStrength:  0.85,
				Rationale:       "Security tickets document policy implementations",
				ApprovalStatus:  types.ApprovalApproved,
				ExecutionStatus: types.ExecPending,
			},
			{
				Source:          "github",
				Query:           "label:documentation",
				SignalStrength:  0.6,
				Rationale:       "Documentation updates show policy awareness",
				ApprovalStatus:  types.ApprovalPending, // This should be skipped
				ExecutionStatus: types.ExecPending,
			},
		},
	}

	// Step 6: Execute the plan
	bundle, err := engine.ExecutePlan(context.Background(), plan)
	if err != nil {
		t.Fatalf("ExecutePlan failed: %v", err)
	}

	// Step 7: Verify results
	if bundle == nil {
		t.Fatal("expected bundle to be non-nil")
	}

	// Should have collected 3 events (2 from github, 1 from jira)
	// The pending item should be skipped
	expectedEvents := 3
	if len(bundle.Events) != expectedEvents {
		t.Errorf("expected %d events, got %d", expectedEvents, len(bundle.Events))
		for i, event := range bundle.Events {
			t.Logf("Event %d: ID=%s Source=%s Type=%s", i+1, event.ID, event.Source, event.Type)
		}
	}

	// Verify event sources
	sourceCount := make(map[string]int)
	for _, event := range bundle.Events {
		sourceCount[event.Source]++
	}

	if sourceCount["github"] != 2 {
		t.Errorf("expected 2 github events, got %d", sourceCount["github"])
	}
	if sourceCount["jira"] != 1 {
		t.Errorf("expected 1 jira event, got %d", sourceCount["jira"])
	}

	// Verify event content
	for _, event := range bundle.Events {
		if event.ID == "" {
			t.Error("event ID should not be empty")
		}
		if event.Source == "" {
			t.Error("event source should not be empty")
		}
		if event.Content == "" {
			t.Error("event content should not be empty")
		}
		if event.Timestamp.IsZero() {
			t.Error("event timestamp should not be zero")
		}
	}

	// Step 8: Verify execution status was updated
	approvedCount := 0
	completedCount := 0
	for _, item := range plan.Items {
		if item.ApprovalStatus == types.ApprovalApproved {
			approvedCount++
			if item.ExecutionStatus == types.ExecComplete {
				completedCount++
			}
		}
	}

	if approvedCount != 2 {
		t.Errorf("expected 2 approved items, got %d", approvedCount)
	}
	if completedCount != 2 {
		t.Errorf("expected 2 completed items, got %d", completedCount)
	}

	t.Log("✅ Successfully executed plan with mock connectors")
	t.Logf("   - Collected %d events from %d sources", len(bundle.Events), len(sourceCount))
	t.Logf("   - Skipped %d pending items", len(plan.Items)-approvedCount)
}

// TestMockConnectorErrorHandling demonstrates error handling in the framework.
func TestMockConnectorErrorHandling(t *testing.T) {
	registry := connectors.NewRegistry()

	// Create mock connector that returns errors
	cfg := connectors.DefaultConfig()
	cfg.APIKey = "mock-token"
	cfg.Extra = map[string]interface{}{"name": "failing-source"}

	connector, err := connectors.NewMockConnector(cfg)
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	mock := connector.(*connectors.MockConnector)
	mock.SetError("bad-query", connectors.ErrInvalidQuery)

	if err := registry.Register(connector); err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	// Create engine
	config := &types.Config{
		AI: types.AIConfig{
			Provider: "mock",
			Model:    "mock-model",
		},
	}
	engine := ai.NewEngineWithConnector(config, ai.NewMockProvider(), registry)

	// Create plan with failing item
	plan := &types.EvidencePlan{
		ID:        "error-test-plan",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{
				Source:          "failing-source",
				Query:           "bad-query",
				SignalStrength:  0.9,
				ApprovalStatus:  types.ApprovalApproved,
				ExecutionStatus: types.ExecPending,
			},
		},
	}

	// Execute - should fail since all items fail
	_, err = engine.ExecutePlan(context.Background(), plan)
	if err == nil {
		t.Error("expected error when all connectors fail")
	}

	// Verify execution status was updated to failed
	if plan.Items[0].ExecutionStatus != types.ExecFailed {
		t.Errorf("expected ExecFailed status, got %v", plan.Items[0].ExecutionStatus)
	}
	if plan.Items[0].Error == "" {
		t.Error("expected error message to be set")
	}

	t.Log("✅ Error handling works correctly")
}

// TestMockConnectorPartialSuccess demonstrates partial failure handling.
func TestMockConnectorPartialSuccess(t *testing.T) {
	registry := connectors.NewRegistry()

	// Create successful connector
	successCfg := connectors.DefaultConfig()
	successCfg.APIKey = "mock-token"
	successCfg.Extra = map[string]interface{}{"name": "success-source"}

	successConnector, err := connectors.NewMockConnector(successCfg)
	if err != nil {
		t.Fatalf("failed to create success mock: %v", err)
	}

	successMock := successConnector.(*connectors.MockConnector)
	successMock.SetEvents("good-query", []types.EvidenceEvent{
		{
			ID:        "success-1",
			Source:    "success-source",
			Type:      "test",
			Timestamp: time.Now(),
			Content:   "Successful event",
		},
	})

	if err := registry.Register(successConnector); err != nil {
		t.Fatalf("failed to register success: %v", err)
	}

	// Create failing connector
	failCfg := connectors.DefaultConfig()
	failCfg.APIKey = "mock-token"
	failCfg.Extra = map[string]interface{}{"name": "fail-source"}

	failConnector, err := connectors.NewMockConnector(failCfg)
	if err != nil {
		t.Fatalf("failed to create fail mock: %v", err)
	}

	failMock := failConnector.(*connectors.MockConnector)
	failMock.SetError("bad-query", connectors.ErrTimeout)

	if err := registry.Register(failConnector); err != nil {
		t.Fatalf("failed to register fail: %v", err)
	}

	// Create engine
	config := &types.Config{
		AI: types.AIConfig{
			Provider: "mock",
			Model:    "mock-model",
		},
	}
	engine := ai.NewEngineWithConnector(config, ai.NewMockProvider(), registry)

	// Create plan with mixed success/failure
	plan := &types.EvidencePlan{
		ID:        "partial-test-plan",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{
				Source:          "success-source",
				Query:           "good-query",
				SignalStrength:  0.9,
				ApprovalStatus:  types.ApprovalApproved,
				ExecutionStatus: types.ExecPending,
			},
			{
				Source:          "fail-source",
				Query:           "bad-query",
				SignalStrength:  0.8,
				ApprovalStatus:  types.ApprovalApproved,
				ExecutionStatus: types.ExecPending,
			},
		},
	}

	// Execute - should succeed with partial results
	bundle, err := engine.ExecutePlan(context.Background(), plan)
	if err != nil {
		t.Fatalf("ExecutePlan should succeed with partial results: %v", err)
	}

	// Should have 1 event from successful connector
	if len(bundle.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(bundle.Events))
	}

	// Verify first item succeeded
	if plan.Items[0].ExecutionStatus != types.ExecComplete {
		t.Errorf("expected first item to be complete, got %v", plan.Items[0].ExecutionStatus)
	}

	// Verify second item failed
	if plan.Items[1].ExecutionStatus != types.ExecFailed {
		t.Errorf("expected second item to fail, got %v", plan.Items[1].ExecutionStatus)
	}

	t.Log("✅ Partial success handling works correctly")
}

// BenchmarkConnectorRegistry measures registry performance.
func BenchmarkConnectorRegistry(b *testing.B) {
	registry := connectors.NewRegistry()

	// Register 10 mock connectors
	for i := 0; i < 10; i++ {
		cfg := connectors.DefaultConfig()
		cfg.APIKey = "mock-token"
		cfg.Extra = map[string]interface{}{"name": string(rune('a' + i))}

		connector, _ := connectors.NewMockConnector(cfg)
		registry.Register(connector)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Lookup connector (should be O(1))
		_ = registry.Get("e")
	}
}
