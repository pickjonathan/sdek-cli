package components

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// getTestPlan creates a deterministic test plan for golden file tests
func getTestPlan() *types.EvidencePlan {
	plan := &types.EvidencePlan{
		ID:        "test-plan-001",
		Framework: "ISO27001",
		Section:   "A.9.4.2",
		Items: []types.PlanItem{
			{
				Source:         "github",
				Query:          "authentication",
				Filters:        []string{"repo:myorg/auth-service"},
				SignalStrength: 0.92,
				Rationale:      "Authentication code reviews are highly relevant for access control requirements",
				ApprovalStatus: types.ApprovalAutoApproved,
				AutoApproved:   true,
			},
			{
				Source:         "github",
				Query:          "mfa",
				Filters:        []string{"repo:myorg/auth-service", "path:src/"},
				SignalStrength: 0.88,
				Rationale:      "Multi-factor authentication implementation details",
				ApprovalStatus: types.ApprovalPending,
				AutoApproved:   false,
			},
			{
				Source:         "jira",
				Query:          "INFOSEC-*",
				Filters:        []string{"status=Done"},
				SignalStrength: 0.75,
				Rationale:      "Security-related tickets demonstrate compliance activities",
				ApprovalStatus: types.ApprovalApproved,
				AutoApproved:   false,
			},
			{
				Source:         "aws",
				Query:          "iam",
				Filters:        []string{"action:GetPolicy", "action:ListPolicies"},
				SignalStrength: 0.65,
				Rationale:      "IAM policy configurations relevant to access control",
				ApprovalStatus: types.ApprovalPending,
				AutoApproved:   false,
			},
			{
				Source:         "slack",
				Query:          "security incident",
				Filters:        []string{"channel:security"},
				SignalStrength: 0.45,
				Rationale:      "Security incident discussions may show response processes",
				ApprovalStatus: types.ApprovalDenied,
				AutoApproved:   false,
			},
		},
		EstimatedSources: 150,
		EstimatedCalls:   25,
		EstimatedTokens:  50000,
		Status:           types.PlanPending,
		CreatedAt:        time.Date(2025, 10, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2025, 10, 18, 12, 0, 0, 0, time.UTC),
	}
	return plan
}

// TestPlanApprovalView tests the plan approval rendering
func TestPlanApprovalView(t *testing.T) {
	plan := getTestPlan()
	model := NewPlanApproval(plan)
	model = model.SetSize(100, 30)

	view := model.View()

	if view == "" {
		t.Error("Plan approval view should not be empty")
	}

	// Test with golden file
	goldenFile := filepath.Join("..", "..", "tests", "golden", "fixtures", "plan_approval_iso27001.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		// Update golden file
		dir := filepath.Dir(goldenFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create golden file directory: %v", err)
		}
		if err := os.WriteFile(goldenFile, []byte(view), 0644); err != nil {
			t.Fatalf("Failed to write golden file: %v", err)
		}
		t.Log("Updated golden file:", goldenFile)
		return
	}

	// Compare with golden file
	golden, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Logf("Golden file not found, skipping comparison. Run with UPDATE_GOLDEN=1 to create it.")
		return
	}

	if string(golden) != view {
		t.Errorf("Plan approval view does not match golden file.\nRun 'UPDATE_GOLDEN=1 go test' to update.")
		// Optionally write the diff to a file for inspection
		diffFile := filepath.Join("..", "..", "tests", "golden", "fixtures", "plan_approval_iso27001.diff")
		os.WriteFile(diffFile, []byte(view), 0644)
		t.Logf("Current output written to: %s", diffFile)
	}
}

// TestPlanApprovalNavigation tests keyboard navigation
func TestPlanApprovalNavigation(t *testing.T) {
	tests := []struct {
		name          string
		key           tea.KeyMsg
		expectedIndex int
	}{
		{"initial state", tea.KeyMsg{}, 0},
		{"down arrow", tea.KeyMsg{Type: tea.KeyDown}, 1},
		{"up arrow", tea.KeyMsg{Type: tea.KeyUp}, 0},
		{"j key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, 1},
		{"k key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := getTestPlan()
			model := NewPlanApproval(plan)
			model = model.SetSize(100, 30)

			if tt.key.Type != tea.KeyType(0) || len(tt.key.Runes) > 0 {
				updated, _ := model.Update(tt.key)
				model = updated.(PlanApprovalModel)
			}

			if model.selectedIndex != tt.expectedIndex {
				t.Errorf("Expected selectedIndex %d, got %d", tt.expectedIndex, model.selectedIndex)
			}
		})
	}
}

// TestPlanApprovalToggle tests approval toggling
func TestPlanApprovalToggle(t *testing.T) {
	plan := getTestPlan()
	model := NewPlanApproval(plan)
	model = model.SetSize(100, 30)

	// Item 1 starts as Pending
	initialStatus := model.plan.Items[1].ApprovalStatus
	if initialStatus != types.ApprovalPending {
		t.Errorf("Expected initial status Pending, got %s", initialStatus)
	}

	// Move to item 1 using KeyDown
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(PlanApprovalModel)

	// Toggle with space bar
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(PlanApprovalModel)

	// Should now be Approved
	if model.plan.Items[1].ApprovalStatus != types.ApprovalApproved {
		t.Errorf("Expected status Approved after toggle, got %s", model.plan.Items[1].ApprovalStatus)
	}

	// Toggle again
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(PlanApprovalModel)

	// Should be back to Pending
	if model.plan.Items[1].ApprovalStatus != types.ApprovalPending {
		t.Errorf("Expected status Pending after second toggle, got %s", model.plan.Items[1].ApprovalStatus)
	}
}

// TestPlanApprovalApproveAll tests the 'a' key to approve all
func TestPlanApprovalApproveAll(t *testing.T) {
	plan := getTestPlan()
	model := NewPlanApproval(plan)
	model = model.SetSize(100, 30)

	// Press 'a' to approve all
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updated.(PlanApprovalModel)

	// All items should be approved
	for i, item := range model.plan.Items {
		if item.ApprovalStatus != types.ApprovalApproved {
			t.Errorf("Item %d: expected Approved, got %s", i, item.ApprovalStatus)
		}
	}
}

// TestPlanApprovalRejectAll tests the 'r' key to reject all
func TestPlanApprovalRejectAll(t *testing.T) {
	plan := getTestPlan()
	model := NewPlanApproval(plan)
	model = model.SetSize(100, 30)

	// Press 'r' to reject all
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model = updated.(PlanApprovalModel)

	// All items should be denied
	for i, item := range model.plan.Items {
		if item.ApprovalStatus != types.ApprovalDenied {
			t.Errorf("Item %d: expected Denied, got %s", i, item.ApprovalStatus)
		}
	}
}

// TestPlanApprovalConfirm tests confirmation
func TestPlanApprovalConfirm(t *testing.T) {
	plan := getTestPlan()
	model := NewPlanApproval(plan)

	// Press enter to confirm
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(PlanApprovalModel)

	if !model.Confirmed() {
		t.Error("Expected model to be confirmed after Enter key")
	}

	if cmd == nil {
		t.Error("Expected tea.Quit command after confirmation")
	}
}

// TestPlanApprovalCancel tests cancellation
func TestPlanApprovalCancel(t *testing.T) {
	plan := getTestPlan()
	model := NewPlanApproval(plan)

	// Press 'q' to cancel
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updated.(PlanApprovalModel)

	if !model.Cancelled() {
		t.Error("Expected model to be cancelled after 'q' key")
	}

	if cmd == nil {
		t.Error("Expected tea.Quit command after cancellation")
	}
}

// TestPlanApprovalBudgetCalculation tests budget calculations
func TestPlanApprovalBudgetCalculation(t *testing.T) {
	plan := getTestPlan()
	model := NewPlanApproval(plan)

	// Initially 2 approved items (1 auto-approved + 1 approved)
	approved := model.countApproved()
	if approved != 2 {
		t.Errorf("Expected 2 approved items, got %d", approved)
	}

	// Approve all items
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updated.(PlanApprovalModel)

	approved = model.countApproved()
	if approved != 5 {
		t.Errorf("Expected 5 approved items after approve all, got %d", approved)
	}

	// Check total events calculation
	totalEvents := model.calculateTotalEvents()
	expectedEvents := plan.EstimatedSources // All items approved = full budget
	if totalEvents != expectedEvents {
		t.Errorf("Expected %d total events, got %d", expectedEvents, totalEvents)
	}
}

// TestPlanApprovalAutoApprovedBadges tests that auto-approved items are shown correctly
func TestPlanApprovalAutoApprovedBadges(t *testing.T) {
	plan := getTestPlan()
	model := NewPlanApproval(plan)
	model = model.SetSize(100, 30)

	view := model.View()

	// Should contain checkmarks for auto-approved items
	if !contains(view, "[✓]") {
		t.Error("Expected view to contain checkmark for approved items")
	}

	// First item should be auto-approved
	if plan.Items[0].ApprovalStatus != types.ApprovalAutoApproved {
		t.Error("Expected first item to be auto-approved")
	}
}

// TestPlanApprovalDeniedBadges tests that denied items show correct badges
func TestPlanApprovalDeniedBadges(t *testing.T) {
	plan := getTestPlan()
	model := NewPlanApproval(plan)
	model = model.SetSize(100, 30)

	view := model.View()

	// Should contain X mark for denied item (item 4)
	if !contains(view, "[✗]") {
		t.Error("Expected view to contain X mark for denied items")
	}

	// Last item should be denied
	if plan.Items[4].ApprovalStatus != types.ApprovalDenied {
		t.Error("Expected last item to be denied")
	}
}

// TestPlanApprovalPendingItems tests pending items display
func TestPlanApprovalPendingItems(t *testing.T) {
	plan := getTestPlan()
	model := NewPlanApproval(plan)
	model = model.SetSize(100, 30)

	view := model.View()

	// Should contain empty checkboxes for pending items
	if !contains(view, "[ ]") {
		t.Error("Expected view to contain empty checkboxes for pending items")
	}

	// Items 1 and 3 should be pending
	pendingCount := 0
	for _, item := range plan.Items {
		if item.ApprovalStatus == types.ApprovalPending {
			pendingCount++
		}
	}

	if pendingCount != 2 {
		t.Errorf("Expected 2 pending items, got %d", pendingCount)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
