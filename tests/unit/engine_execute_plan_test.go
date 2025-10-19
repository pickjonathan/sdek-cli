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

// T010: Contract test for Engine.ExecutePlan
// These tests define the contract for executing evidence collection plans via MCP connectors
// EXPECTED: These tests MUST FAIL until Engine.ExecutePlan is implemented in Phase 3.3

func TestExecutePlan_ApprovedPlan(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockConnector := ai.NewMockMCPConnector()
	mockConnector.SetEvents("github", []types.EvidenceEvent{
		{ID: "evt-1", Source: "github", Content: "Added MFA authentication"},
		{ID: "evt-2", Source: "github", Content: "Updated access control policy"},
	})
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), mockConnector)
	
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{
				Source:          "github",
				Query:           "authentication",
				SignalStrength:  0.8,
				Rationale:       "Look for auth changes",
				ApprovalStatus:  types.ApprovalApproved,
				ExecutionStatus: types.ExecPending,
			},
		},
	}

	// Act
	ctx := context.Background()
	bundle, err := engine.ExecutePlan(ctx, plan)

	// Assert
	require.NoError(t, err, "ExecutePlan should succeed with approved plan")
	require.NotNil(t, bundle)
	
	// Verify events were collected
	assert.Len(t, bundle.Events, 2, "Should collect events from github")
	assert.Equal(t, "github", bundle.Events[0].Source)
	
	// Verify plan item status updated
	assert.Equal(t, types.ExecComplete, plan.Items[0].ExecutionStatus)
	assert.Equal(t, 2, plan.Items[0].EventsCollected)
	assert.Empty(t, plan.Items[0].Error)
}

func TestExecutePlan_SkipsPendingItems(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockConnector := ai.NewMockMCPConnector()
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), mockConnector)
	
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{
				Source:          "github",
				Query:           "authentication",
				ApprovalStatus:  types.ApprovalPending, // Should be skipped
				ExecutionStatus: types.ExecPending,
			},
			{
				Source:          "jira",
				Query:           "SEC-*",
				ApprovalStatus:  types.ApprovalApproved, // Should be executed
				ExecutionStatus: types.ExecPending,
			},
		},
	}

	// Act
	_, err := engine.ExecutePlan(context.Background(), plan)

	// Assert
	require.NoError(t, err)
	
	// Verify pending item was skipped
	assert.Equal(t, types.ExecPending, plan.Items[0].ExecutionStatus, 
		"Pending item should remain pending")
	
	// Verify approved item was executed
	assert.Equal(t, types.ExecComplete, plan.Items[1].ExecutionStatus,
		"Approved item should be executed")
}

func TestExecutePlan_SkipsDeniedItems(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockConnector := ai.NewMockMCPConnector()
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), mockConnector)
	
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{
				Source:          "github",
				Query:           "authentication",
				ApprovalStatus:  types.ApprovalDenied, // Should be skipped
				ExecutionStatus: types.ExecPending,
			},
		},
	}

	// Act
	_, err := engine.ExecutePlan(context.Background(), plan)

	// Assert
	require.NoError(t, err)
	
	// Verify denied item was skipped
	assert.Equal(t, types.ExecPending, plan.Items[0].ExecutionStatus,
		"Denied item should remain pending (not executed)")
}

func TestExecutePlan_ExecutesAutoApprovedItems(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockConnector := ai.NewMockMCPConnector()
	mockConnector.SetEvents("github", []types.EvidenceEvent{
		{ID: "evt-1", Source: "github", Content: "Auth update"},
	})
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), mockConnector)
	
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{
				Source:          "github",
				Query:           "authentication",
				ApprovalStatus:  types.ApprovalAutoApproved, // Auto-approved
				ExecutionStatus: types.ExecPending,
			},
		},
	}

	// Act
	bundle, err := engine.ExecutePlan(context.Background(), plan)

	// Assert
	require.NoError(t, err)
	assert.Len(t, bundle.Events, 1, "Auto-approved item should be executed")
	assert.Equal(t, types.ExecComplete, plan.Items[0].ExecutionStatus)
}

func TestExecutePlan_ParallelExecution(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockConnector := ai.NewMockMCPConnector()
	// Configure connector with delay to test parallelism
	mockConnector.SetDelay(100 * time.Millisecond)
	mockConnector.SetEvents("github", []types.EvidenceEvent{{ID: "evt-1"}})
	mockConnector.SetEvents("jira", []types.EvidenceEvent{{ID: "evt-2"}})
	mockConnector.SetEvents("aws", []types.EvidenceEvent{{ID: "evt-3"}})
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), mockConnector)
	
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{Source: "github", Query: "auth", ApprovalStatus: types.ApprovalApproved},
			{Source: "jira", Query: "SEC-*", ApprovalStatus: types.ApprovalApproved},
			{Source: "aws", Query: "iam:*", ApprovalStatus: types.ApprovalApproved},
		},
	}

	// Act
	start := time.Now()
	bundle, err := engine.ExecutePlan(context.Background(), plan)
	duration := time.Since(start)

	// Assert
	require.NoError(t, err)
	assert.Len(t, bundle.Events, 3)
	
	// Verify parallel execution (should be ~100ms, not 300ms)
	assert.Less(t, duration, 200*time.Millisecond, 
		"Parallel execution should be faster than sequential")
}

func TestExecutePlan_HandlesPartialFailures(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockConnector := ai.NewMockMCPConnector()
	mockConnector.SetEvents("github", []types.EvidenceEvent{{ID: "evt-1"}})
	mockConnector.SetError("jira", fmt.Errorf("jira connector timeout"))
	mockConnector.SetEvents("aws", []types.EvidenceEvent{{ID: "evt-3"}})
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), mockConnector)
	
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{Source: "github", Query: "auth", ApprovalStatus: types.ApprovalApproved},
			{Source: "jira", Query: "SEC-*", ApprovalStatus: types.ApprovalApproved},
			{Source: "aws", Query: "iam:*", ApprovalStatus: types.ApprovalApproved},
		},
	}

	// Act
	bundle, err := engine.ExecutePlan(context.Background(), plan)

	// Assert
	require.NoError(t, err, "Partial failure should not return error")
	assert.Len(t, bundle.Events, 2, "Should collect from successful sources")
	
	// Verify item statuses
	assert.Equal(t, types.ExecComplete, plan.Items[0].ExecutionStatus, "github should succeed")
	assert.Equal(t, types.ExecFailed, plan.Items[1].ExecutionStatus, "jira should fail")
	assert.NotEmpty(t, plan.Items[1].Error, "jira error should be recorded")
	assert.Equal(t, types.ExecComplete, plan.Items[2].ExecutionStatus, "aws should succeed")
}

func TestExecutePlan_NormalizesEventsToSchema(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockConnector := ai.NewMockMCPConnector()
	mockConnector.SetEvents("github", []types.EvidenceEvent{
		{
			ID:        "commit-123",
			Source:    "github",
			Type:      "commit",
			Timestamp: time.Now(),
			Content:   "Added authentication",
			Metadata: map[string]interface{}{
				"author": "john@example.com",
				"repo":   "security-service",
			},
		},
	})
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), mockConnector)
	
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{Source: "github", Query: "auth", ApprovalStatus: types.ApprovalApproved},
		},
	}

	// Act
	bundle, err := engine.ExecutePlan(context.Background(), plan)

	// Assert
	require.NoError(t, err)
	require.Len(t, bundle.Events, 1)
	
	event := bundle.Events[0]
	assert.NotEmpty(t, event.ID, "Event should have ID")
	assert.Equal(t, "github", event.Source)
	assert.Equal(t, "commit", event.Type)
	assert.NotZero(t, event.Timestamp, "Event should have timestamp")
	assert.NotEmpty(t, event.Content, "Event should have content")
}

func TestExecutePlan_UpdatesExecutionStatus(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockConnector := ai.NewMockMCPConnector()
	mockConnector.SetEvents("github", []types.EvidenceEvent{
		{ID: "evt-1"},
		{ID: "evt-2"},
		{ID: "evt-3"},
	})
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), mockConnector)
	
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{Source: "github", Query: "auth", ApprovalStatus: types.ApprovalApproved},
		},
	}

	// Act
	_, err := engine.ExecutePlan(context.Background(), plan)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, types.ExecComplete, plan.Items[0].ExecutionStatus)
	assert.Equal(t, 3, plan.Items[0].EventsCollected)
	assert.Empty(t, plan.Items[0].Error)
}

func TestExecutePlan_PlanNotApprovedReturnsError(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), ai.NewMockMCPConnector())
	
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanPending, // Not approved
		Items: []types.PlanItem{
			{Source: "github", Query: "auth", ApprovalStatus: types.ApprovalApproved},
		},
	}

	// Act
	_, err := engine.ExecutePlan(context.Background(), plan)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ai.ErrPlanNotApproved, "Should return ErrPlanNotApproved")
}

func TestExecutePlan_NoApprovedItemsReturnsError(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), ai.NewMockMCPConnector())
	
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{Source: "github", Query: "auth", ApprovalStatus: types.ApprovalPending}, // Not approved
			{Source: "jira", Query: "SEC-*", ApprovalStatus: types.ApprovalDenied},   // Denied
		},
	}

	// Act
	_, err := engine.ExecutePlan(context.Background(), plan)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ai.ErrNoApprovedItems, "Should return ErrNoApprovedItems")
}

func TestExecutePlan_AllConnectorsFailReturnsError(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockConnector := ai.NewMockMCPConnector()
	mockConnector.SetError("github", fmt.Errorf("connector timeout"))
	mockConnector.SetError("jira", fmt.Errorf("connector auth failed"))
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), mockConnector)
	
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{Source: "github", Query: "auth", ApprovalStatus: types.ApprovalApproved},
			{Source: "jira", Query: "SEC-*", ApprovalStatus: types.ApprovalApproved},
		},
	}

	// Act
	_, err := engine.ExecutePlan(context.Background(), plan)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ai.ErrMCPConnectorFailed, "Should return ErrMCPConnectorFailed")
}

func TestExecutePlan_ContextCancellationReturnsError(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), ai.NewMockMCPConnector())
	
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items: []types.PlanItem{
			{Source: "github", Query: "auth", ApprovalStatus: types.ApprovalApproved},
		},
	}

	// Act - cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := engine.ExecutePlan(ctx, plan)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context", "Error should mention context")
}

func TestExecutePlan_PerformanceTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockConnector := ai.NewMockMCPConnector()
	// Configure 10 sources with events
	for i := 0; i < 10; i++ {
		source := fmt.Sprintf("source-%d", i)
		mockConnector.SetEvents(source, []types.EvidenceEvent{
			{ID: fmt.Sprintf("evt-%d", i), Source: source},
		})
	}
	engine := ai.NewEngineWithConnector(cfg, ai.NewMockProvider(), mockConnector)
	
	// Create plan with 10 sources
	items := make([]types.PlanItem, 10)
	for i := 0; i < 10; i++ {
		items[i] = types.PlanItem{
			Source:         fmt.Sprintf("source-%d", i),
			Query:          "test",
			ApprovalStatus: types.ApprovalApproved,
		}
	}
	plan := &types.EvidencePlan{
		ID:        "plan-001",
		Framework: "SOC2",
		Section:   "CC6.1",
		Status:    types.PlanApproved,
		Items:     items,
	}

	// Act
	start := time.Now()
	bundle, err := engine.ExecutePlan(context.Background(), plan)
	duration := time.Since(start)

	// Assert
	require.NoError(t, err)
	assert.Len(t, bundle.Events, 10)
	assert.Less(t, duration, 5*time.Minute, "ExecutePlan should complete in <5min for 10 sources")
	t.Logf("ExecutePlan took %v for 10 sources", duration)
}
