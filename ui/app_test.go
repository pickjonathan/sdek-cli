package ui_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/pickjonathan/sdek-cli/ui"
	"github.com/pickjonathan/sdek-cli/ui/models"
)

// getTestState creates a deterministic test state for golden file tests
func getTestState() *store.State {
	return &store.State{
		Sources: []types.Source{
			{
				ID:         "src-001",
				Type:       "git",
				Name:       "main-repo",
				Status:     "active",
				Enabled:    true,
				EventCount: 150,
			},
			{
				ID:         "src-002",
				Type:       "jira",
				Name:       "project-tracker",
				Status:     "active",
				Enabled:    true,
				EventCount: 75,
			},
		},
		Frameworks: []types.Framework{
			{
				ID:                   "soc2",
				Name:                 "SOC 2 Type II",
				Description:          "Service Organization Control 2",
				Version:              "2017",
				CompliancePercentage: 85.5,
			},
			{
				ID:                   "iso27001",
				Name:                 "ISO 27001",
				Description:          "Information Security Management",
				Version:              "2013",
				CompliancePercentage: 72.3,
			},
		},
		Controls: []types.Control{
			{
				ID:          "CC6.1",
				FrameworkID: "soc2",
				Title:       "Logical Access Controls",
				Description: "The entity implements logical access security software.",
				Category:    "Access Control",
				RiskStatus:  "green",
			},
			{
				ID:          "CC6.2",
				FrameworkID: "soc2",
				Title:       "Authentication",
				Description: "Prior to issuing credentials, the entity registers users.",
				Category:    "Access Control",
				RiskStatus:  "yellow",
			},
			{
				ID:          "CC6.3",
				FrameworkID: "soc2",
				Title:       "Authorization",
				Description: "The entity authorizes users based on their roles.",
				Category:    "Access Control",
				RiskStatus:  "red",
			},
		},
		Evidence: []types.Evidence{
			{
				EventID:         "evt-001",
				ControlID:       "CC6.1",
				ConfidenceLevel: "high",
				Reasoning:       "Git commit shows implementation of access control",
			},
			{
				EventID:         "evt-002",
				ControlID:       "CC6.1",
				ConfidenceLevel: "medium",
				Reasoning:       "JIRA ticket documents access control requirements",
			},
			{
				EventID:         "evt-003",
				ControlID:       "CC6.2",
				ConfidenceLevel: "low",
				Reasoning:       "Partial evidence of authentication implementation",
			},
		},
		Events: []types.Event{
			{
				ID:        "evt-001",
				SourceID:  "src-001",
				EventType: "commit",
				Title:     "implemented access control",
				Content:   "Added logical access control to auth service",
				Author:    "developer@example.com",
				Metadata:  map[string]interface{}{"repo": "main-repo"},
			},
			{
				ID:        "evt-002",
				SourceID:  "src-002",
				EventType: "ticket",
				Title:     "Access control requirements",
				Content:   "Document and implement access control requirements",
				Author:    "pm@example.com",
				Metadata:  map[string]interface{}{"ticket": "PROJ-123"},
			},
		},
		Findings: []types.Finding{
			{
				ID:          "finding-001",
				ControlID:   "CC6.3",
				Severity:    "high",
				Title:       "Missing authorization checks",
				Description: "Authorization logic not fully implemented",
				Status:      "open",
			},
		},
	}
}

// TestHomeView tests the home screen rendering
func TestHomeView(t *testing.T) {
	state := getTestState()
	model := models.NewHomeModel(state)
	model.SetSize(80, 24)

	view := model.View()

	// Check for key elements
	if view == "" {
		t.Error("Home view should not be empty")
	}

	// Test with golden file
	goldenFile := filepath.Join("testdata", "golden", "home_view.txt")
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
		t.Errorf("Home view does not match golden file.\nRun 'UPDATE_GOLDEN=1 go test' to update.")
		// Optionally write the diff to a file for inspection
		diffFile := filepath.Join("testdata", "golden", "home_view.diff")
		os.WriteFile(diffFile, []byte(view), 0644)
		t.Logf("Current output written to: %s", diffFile)
	}
}

// TestSourcesView tests the sources list rendering
func TestSourcesView(t *testing.T) {
	state := getTestState()
	model := models.NewSourcesModel(state)
	model.SetSize(80, 24)

	view := model.View()

	if view == "" {
		t.Error("Sources view should not be empty")
	}

	// Check that both sources are present
	if !contains(view, "git") {
		t.Error("Sources view should contain 'git' source")
	}
	if !contains(view, "jira") {
		t.Error("Sources view should contain 'jira' source")
	}

	// Test with golden file
	goldenFile := filepath.Join("testdata", "golden", "sources_view.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		dir := filepath.Dir(goldenFile)
		os.MkdirAll(dir, 0755)
		os.WriteFile(goldenFile, []byte(view), 0644)
		t.Log("Updated golden file:", goldenFile)
		return
	}

	golden, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Logf("Golden file not found, skipping comparison.")
		return
	}

	if string(golden) != view {
		t.Errorf("Sources view does not match golden file.")
		diffFile := filepath.Join("testdata", "golden", "sources_view.diff")
		os.WriteFile(diffFile, []byte(view), 0644)
		t.Logf("Current output written to: %s", diffFile)
	}
}

// TestControlsView tests the controls list rendering
func TestControlsView(t *testing.T) {
	state := getTestState()
	model := models.NewControlsModel(state, "")
	model.SetSize(80, 24)

	view := model.View()

	if view == "" {
		t.Error("Controls view should not be empty")
	}

	// Check for risk summary header
	if !contains(view, "Risk Summary") {
		t.Error("Controls view should show risk distribution summary")
	}

	// Test with golden file
	goldenFile := filepath.Join("testdata", "golden", "controls_view.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		dir := filepath.Dir(goldenFile)
		os.MkdirAll(dir, 0755)
		os.WriteFile(goldenFile, []byte(view), 0644)
		t.Log("Updated golden file:", goldenFile)
		return
	}

	golden, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Logf("Golden file not found, skipping comparison.")
		return
	}

	if string(golden) != view {
		t.Errorf("Controls view does not match golden file.")
		diffFile := filepath.Join("testdata", "golden", "controls_view.diff")
		os.WriteFile(diffFile, []byte(view), 0644)
		t.Logf("Current output written to: %s", diffFile)
	}
}

// TestEvidenceView tests the evidence list rendering
func TestEvidenceView(t *testing.T) {
	state := getTestState()
	model := models.NewEvidenceModel(state, "")
	model.SetSize(80, 24)

	view := model.View()

	if view == "" {
		t.Error("Evidence view should not be empty")
	}

	// Check for confidence header
	if !contains(view, "Confidence:") {
		t.Error("Evidence view should show confidence distribution")
	}

	// Test with golden file
	goldenFile := filepath.Join("testdata", "golden", "evidence_view.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		dir := filepath.Dir(goldenFile)
		os.MkdirAll(dir, 0755)
		os.WriteFile(goldenFile, []byte(view), 0644)
		t.Log("Updated golden file:", goldenFile)
		return
	}

	golden, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Logf("Golden file not found, skipping comparison.")
		return
	}

	if string(golden) != view {
		t.Errorf("Evidence view does not match golden file.")
		diffFile := filepath.Join("testdata", "golden", "evidence_view.diff")
		os.WriteFile(diffFile, []byte(view), 0644)
		t.Logf("Current output written to: %s", diffFile)
	}
}

// TestAppModelIntegration tests the main app model integration
func TestAppModelIntegration(t *testing.T) {
	state := getTestState()
	model := ui.NewModel(state, "")

	// Test initial state
	if model.View() == "" {
		t.Error("App view should not be empty")
	}

	// We can't easily test interactive updates without a full Bubble Tea test harness,
	// but we can at least verify the model initializes correctly
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		len(s) >= len(substr) &&
		findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
