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

// T009: Contract test for Engine.ProposePlan
// These tests define the contract for autonomous evidence plan generation
// EXPECTED: These tests MUST FAIL until Engine.ProposePlan is implemented in Phase 3.3

func TestProposePlan_ValidPreamble(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
			Budgets: types.BudgetLimits{
				MaxSources:  50,
				MaxAPICalls: 500,
				MaxTokens:   250000,
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

	// Act
	ctx := context.Background()
	plan, err := engine.ProposePlan(ctx, *preamble)

	// Assert
	require.NoError(t, err, "ProposePlan should succeed with valid preamble")
	require.NotNil(t, plan)

	// Verify plan metadata
	assert.NotEmpty(t, plan.ID, "Plan should have ID")
	assert.Equal(t, "SOC2", plan.Framework)
	assert.Equal(t, "CC6.1", plan.Section)
	assert.Equal(t, types.PlanPending, plan.Status, "New plan should have 'pending' status")

	// Verify plan has items
	assert.NotEmpty(t, plan.Items, "Plan should have at least one item")

	// Verify budget estimates
	assert.GreaterOrEqual(t, plan.EstimatedSources, 1, "Should estimate at least 1 source")
	assert.GreaterOrEqual(t, plan.EstimatedCalls, 1, "Should estimate at least 1 API call")
	assert.GreaterOrEqual(t, plan.EstimatedTokens, 0, "Should estimate token usage")
}

func TestProposePlan_PlanItemsHaveRequiredFields(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
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

	// Act
	plan, err := engine.ProposePlan(context.Background(), *preamble)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, plan.Items)

	// Verify each item has required fields
	for i, item := range plan.Items {
		assert.NotEmpty(t, item.Source, "Item %d should have source", i)
		assert.NotEmpty(t, item.Query, "Item %d should have query", i)
		assert.GreaterOrEqual(t, item.SignalStrength, 0.0, "Item %d signal strength >= 0.0", i)
		assert.LessOrEqual(t, item.SignalStrength, 1.0, "Item %d signal strength <= 1.0", i)
		assert.NotEmpty(t, item.Rationale, "Item %d should have rationale", i)
		assert.Equal(t, types.ApprovalPending, item.ApprovalStatus, "Item %d should be pending", i)
	}
}

func TestProposePlan_AutoApproveMarking(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"auth*", "*login*"},
					"aws":    {"iam:*"},
				},
			},
		},
	}
	mockProvider := ai.NewMockProvider()
	// Configure mock to return plan with matching queries
	mockProvider.SetPlanItems([]types.PlanItem{
		{Source: "github", Query: "authentication"},
		{Source: "github", Query: "payment"},
		{Source: "aws", Query: "iam:CreateUser"},
	})
	engine := ai.NewEngine(cfg, mockProvider)

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	// Act
	plan, err := engine.ProposePlan(context.Background(), *preamble)

	// Assert
	require.NoError(t, err)
	require.Len(t, plan.Items, 3)

	// After sorting: aws/iam:CreateUser, github/authentication, github/payment
	// Item 0: aws/iam:CreateUser should match "iam:*"
	assert.Equal(t, "aws", plan.Items[0].Source)
	assert.Equal(t, "iam:CreateUser", plan.Items[0].Query)
	assert.True(t, plan.Items[0].AutoApproved, "aws/iam:CreateUser should be auto-approved")
	assert.Equal(t, types.ApprovalAutoApproved, plan.Items[0].ApprovalStatus)

	// Item 1: github/authentication should match "auth*"
	assert.Equal(t, "github", plan.Items[1].Source)
	assert.Equal(t, "authentication", plan.Items[1].Query)
	assert.True(t, plan.Items[1].AutoApproved, "github/authentication should be auto-approved")
	assert.Equal(t, types.ApprovalAutoApproved, plan.Items[1].ApprovalStatus)

	// Item 2: github/payment should NOT match
	assert.Equal(t, "github", plan.Items[2].Source)
	assert.Equal(t, "payment", plan.Items[2].Query)
	assert.False(t, plan.Items[2].AutoApproved, "github/payment should not be auto-approved")
	assert.Equal(t, types.ApprovalPending, plan.Items[2].ApprovalStatus)
}

func TestProposePlan_DeterministicSorting(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
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

	// Act - generate plan twice
	plan1, err := engine.ProposePlan(context.Background(), *preamble)
	require.NoError(t, err)

	plan2, err := engine.ProposePlan(context.Background(), *preamble)
	require.NoError(t, err)

	// Assert - plans should have identical item order
	require.Equal(t, len(plan1.Items), len(plan2.Items), "Plans should have same item count")

	for i := range plan1.Items {
		assert.Equal(t, plan1.Items[i].Source, plan2.Items[i].Source,
			"Item %d source should match", i)
		assert.Equal(t, plan1.Items[i].Query, plan2.Items[i].Query,
			"Item %d query should match", i)
	}

	// Verify items are sorted (source asc, then query asc)
	for i := 1; i < len(plan1.Items); i++ {
		prev := plan1.Items[i-1]
		curr := plan1.Items[i]

		if prev.Source == curr.Source {
			assert.LessOrEqual(t, prev.Query, curr.Query,
				"Within same source, queries should be sorted")
		} else {
			assert.Less(t, prev.Source, curr.Source,
				"Sources should be sorted alphabetically")
		}
	}
}

func TestProposePlan_EnforcesBudgetLimits(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
			Budgets: types.BudgetLimits{
				MaxSources:  5, // Strict limit
				MaxAPICalls: 50,
				MaxTokens:   1000,
			},
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

	// Act
	plan, err := engine.ProposePlan(context.Background(), *preamble)

	// Assert
	require.NoError(t, err)

	// Verify budget constraints are respected
	assert.LessOrEqual(t, plan.EstimatedSources, 5, "Should not exceed max sources")
	assert.LessOrEqual(t, plan.EstimatedCalls, 50, "Should not exceed max API calls")
	assert.LessOrEqual(t, plan.EstimatedTokens, 1000, "Should not exceed max tokens")
	assert.LessOrEqual(t, len(plan.Items), 5, "Should not have more items than max sources")
}

func TestProposePlan_DiverseSources(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockProvider := ai.NewMockProvider()
	// Configure mock to return diverse sources
	mockProvider.SetPlanItems([]types.PlanItem{
		{Source: "github", Query: "authentication"},
		{Source: "jira", Query: "SEC-*"},
		{Source: "aws", Query: "iam:*"},
		{Source: "slack", Query: "#security-*"},
		{Source: "docs", Query: "security-policy"},
	})
	engine := ai.NewEngine(cfg, mockProvider)

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	// Act
	plan, err := engine.ProposePlan(context.Background(), *preamble)

	// Assert
	require.NoError(t, err)

	// Count unique sources
	sources := make(map[string]bool)
	for _, item := range plan.Items {
		sources[item.Source] = true
	}

	assert.GreaterOrEqual(t, len(sources), 3, "Plan should include diverse sources (at least 3)")
}

func TestProposePlan_InvalidPreambleReturnsError(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
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

	// Act
	_, err := engine.ProposePlan(context.Background(), invalidPreamble)

	// Assert
	require.Error(t, err, "Should return error for invalid preamble")
	assert.Contains(t, err.Error(), "preamble", "Error should mention preamble")
}

func TestProposePlan_ProviderErrorReturnsError(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
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

	// Act
	_, err = engine.ProposePlan(context.Background(), *preamble)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ai.ErrProviderUnavailable, "Should return provider error")
}

func TestProposePlan_NoPlanItemsReturnsError(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockProvider := ai.NewMockProvider()
	mockProvider.SetPlanItems([]types.PlanItem{}) // No items
	engine := ai.NewEngine(cfg, mockProvider)

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	// Act
	_, err = engine.ProposePlan(context.Background(), *preamble)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ai.ErrNoPlanItems, "Should return ErrNoPlanItems")
}

func TestProposePlan_BudgetExceededReturnsError(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
			Budgets: types.BudgetLimits{
				MaxSources:  2, // Very strict limit
				MaxAPICalls: 10,
				MaxTokens:   100,
			},
		},
	}
	mockProvider := ai.NewMockProvider()
	// Configure mock to return plan that exceeds budget
	items := make([]types.PlanItem, 5) // Exceeds MaxSources
	for i := 0; i < 5; i++ {
		items[i] = types.PlanItem{
			Source: "github",
			Query:  fmt.Sprintf("query-%d", i),
		}
	}
	mockProvider.SetPlanItems(items)
	engine := ai.NewEngine(cfg, mockProvider)

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	// Act
	_, err = engine.ProposePlan(context.Background(), *preamble)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, ai.ErrBudgetExceeded, "Should return ErrBudgetExceeded")
}

func TestProposePlan_ContextCancellationReturnsError(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
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

	// Act - cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = engine.ProposePlan(ctx, *preamble)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context", "Error should mention context")
}

func TestProposePlan_InjectionsPreambleIntoPrompt(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
		},
	}
	mockProvider := ai.NewMockProvider()
	engine := ai.NewEngine(cfg, mockProvider)

	preamble, err := types.NewContextPreamble(
		"ISO27001",
		"2013",
		"A.9.4.2",
		"Secure log-on procedures shall control access to information systems.",
		nil,
	)
	require.NoError(t, err)

	// Act
	_, err = engine.ProposePlan(context.Background(), *preamble)

	// Assert
	require.NoError(t, err)
	lastPrompt := mockProvider.GetLastPrompt()
	assert.Contains(t, lastPrompt, "ISO27001", "Prompt should contain framework")
	assert.Contains(t, lastPrompt, "A.9.4.2", "Prompt should contain section")
	assert.Contains(t, lastPrompt, "Secure log-on procedures", "Prompt should contain excerpt")
}

func TestProposePlan_PerformanceTarget(t *testing.T) {
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
	engine := ai.NewEngine(cfg, ai.NewMockProvider())

	preamble, err := types.NewContextPreamble(
		"SOC2",
		"2017",
		"CC6.1",
		"Access controls shall be implemented to ensure that only authorized individuals can access sensitive data.",
		nil,
	)
	require.NoError(t, err)

	// Act
	start := time.Now()
	_, err = engine.ProposePlan(context.Background(), *preamble)
	duration := time.Since(start)

	// Assert
	require.NoError(t, err)
	assert.Less(t, duration, 10*time.Second, "ProposePlan should complete in <10s")
	t.Logf("ProposePlan took %v", duration)
}

func TestProposePlan_NoCaching(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Enabled:  true,
			Provider: "mock",
			Mode:     types.AIModeAutonomous,
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

	// Act - call twice with same inputs
	_, err = engine.ProposePlan(context.Background(), *preamble)
	require.NoError(t, err)
	callCount1 := mockProvider.GetCallCount()

	_, err = engine.ProposePlan(context.Background(), *preamble)
	require.NoError(t, err)
	callCount2 := mockProvider.GetCallCount()

	// Assert - should NOT use cache (plans should be fresh)
	assert.Greater(t, callCount2, callCount1, "ProposePlan should not use cache")
}
