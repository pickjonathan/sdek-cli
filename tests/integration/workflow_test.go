package integration

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/analyze"
	"github.com/pickjonathan/sdek-cli/internal/report"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// TestCompleteWorkflow tests the core workflow: data → analyze → report
func TestCompleteWorkflow(t *testing.T) {
	// Setup temporary state directory
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Step 1: Create test data
	t.Log("Step 1: Creating test data...")
	state := &store.State{
		Sources: []types.Source{
			{ID: "src-001", Type: "git", Name: "test-repo", Status: "active", Enabled: true, EventCount: 2},
		},
		Frameworks: []types.Framework{
			{ID: "soc2", Name: "SOC 2", Version: "2017"},
		},
		Controls: []types.Control{
			{ID: "CC6.1", FrameworkID: "soc2", Title: "Access Control", Category: "Security"},
		},
		Events: []types.Event{
			{ID: "evt-001", SourceID: "src-001", EventType: "commit", Title: "Add access control", Content: "Implemented access control logic"},
		},
		Evidence: []types.Evidence{},
		Findings: []types.Finding{},
	}

	if err := state.Save(); err != nil {
		t.Fatalf("Failed to save initial state: %v", err)
	}

	// Step 2: Map events to controls to generate evidence
	t.Log("Step 2: Running analysis...")
	mapper := analyze.NewMapper()
	state.Evidence = mapper.MapEventsToControls(state.Events)

	if len(state.Evidence) == 0 {
		t.Error("Mapping should generate evidence")
	}
	t.Logf("✓ Generated %d evidence items", len(state.Evidence))

	// Step 3: Calculate risk and generate findings
	riskScorer := analyze.NewRiskScorer()
	for i := range state.Controls {
		control := &state.Controls[i]

		// Filter evidence for this control
		var controlEvidence []types.Evidence
		for _, ev := range state.Evidence {
			if ev.ControlID == control.ID {
				controlEvidence = append(controlEvidence, ev)
			}
		}

		// Generate findings
		findings := riskScorer.GenerateFindingsForControl(*control, controlEvidence)
		state.Findings = append(state.Findings, findings...)

		// Calculate risk status
		riskStatus := riskScorer.CalculateControlRisk(findings, len(controlEvidence))
		control.RiskStatus = riskStatus
	}

	t.Logf("✓ Generated %d findings", len(state.Findings))

	// Step 4: Calculate compliance percentage
	for i := range state.Frameworks {
		fw := &state.Frameworks[i]
		var frameworkControls []types.Control

		for _, ctrl := range state.Controls {
			if ctrl.FrameworkID == fw.ID {
				frameworkControls = append(frameworkControls, ctrl)
			}
		}

		fw.CompliancePercentage = riskScorer.CalculateOverallCompliance(frameworkControls)
	}

	t.Logf("✓ Compliance: %.1f%%", state.Frameworks[0].CompliancePercentage)

	// Save analyzed state
	if err := state.Save(); err != nil {
		t.Fatalf("Failed to save analyzed state: %v", err)
	}

	// Step 5: Generate report
	t.Log("Step 3: Generating report...")
	exporter := report.NewExporter("1.0")

	reportObj, err := exporter.GenerateReport(
		state.Sources,
		state.Events,
		state.Frameworks,
		state.Controls,
		state.Evidence,
		state.Findings,
		"",
	)
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Verify report structure
	if reportObj.Metadata.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", reportObj.Metadata.Version)
	}
	if reportObj.Summary.TotalSources != 1 {
		t.Errorf("Expected 1 source, got %d", reportObj.Summary.TotalSources)
	}
	if reportObj.Summary.TotalEvents != 1 {
		t.Errorf("Expected 1 event, got %d", reportObj.Summary.TotalEvents)
	}

	t.Log("✓ Complete workflow test passed!")
}

// TestStatePersistence tests state save and load
func TestStatePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create and save state
	state1 := &store.State{
		Sources: []types.Source{
			{ID: "src-001", Name: "test-source", Type: "git"},
		},
		Events: []types.Event{
			{ID: "evt-001", SourceID: "src-001", Title: "Test Event"},
		},
	}

	if err := state1.Save(); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Load state
	state2, err := store.Load()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	// Verify
	if len(state2.Sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(state2.Sources))
	}
	if len(state2.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(state2.Events))
	}
	if state2.Sources[0].ID != "src-001" {
		t.Error("Source ID not persisted correctly")
	}
	if state2.Events[0].ID != "evt-001" {
		t.Error("Event ID not persisted correctly")
	}

	t.Log("✓ State persistence works correctly")
}

// TestEvidenceMapping tests the mapper logic
func TestEvidenceMapping(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create test events
	events := []types.Event{
		{
			ID:        "evt-001",
			EventType: "commit",
			Title:     "Implement access control and authentication",
			Content:   "Added role-based access control and multi-factor authentication",
		},
		{
			ID:        "evt-002",
			EventType: "commit",
			Title:     "Add encryption for data at rest",
			Content:   "Implemented AES-256 encryption for sensitive data storage",
		},
	}

	// Run mapping
	mapper := analyze.NewMapper()
	evidence := mapper.MapEventsToControls(events)

	if len(evidence) == 0 {
		t.Error("Mapper should generate evidence from events")
	}

	// Verify evidence structure
	for _, ev := range evidence {
		if ev.ID == "" {
			t.Error("Evidence missing ID")
		}
		if ev.EventID == "" {
			t.Error("Evidence missing event ID")
		}
		if ev.ControlID == "" {
			t.Error("Evidence missing control ID")
		}
		if ev.ConfidenceLevel == "" {
			t.Error("Evidence missing confidence level")
		}

		validLevels := map[string]bool{"high": true, "medium": true, "low": true}
		if !validLevels[ev.ConfidenceLevel] {
			t.Errorf("Invalid confidence level: %s", ev.ConfidenceLevel)
		}

		if len(ev.Keywords) == 0 {
			t.Error("Evidence should have matched keywords")
		}
	}

	t.Logf("✓ Mapped %d evidence items", len(evidence))
}

// TestRiskCalculation tests risk scoring logic
func TestRiskCalculation(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	riskScorer := analyze.NewRiskScorer()

	// Test case 1: No evidence or findings (red risk)
	riskStatus := riskScorer.CalculateControlRisk([]types.Finding{}, 0)
	if riskStatus != "red" {
		t.Errorf("Expected red risk for no evidence, got %s", riskStatus)
	}

	// Test case 2: Insufficient evidence (yellow risk)
	findings := []types.Finding{
		{Severity: "low"},
	}
	riskStatus = riskScorer.CalculateControlRisk(findings, 2)
	if riskStatus != "yellow" {
		t.Errorf("Expected yellow risk with insufficient evidence, got %s", riskStatus)
	}

	// Test case 3: Multiple controls for compliance calculation
	controls := []types.Control{
		{ID: "C1", RiskStatus: "green"},
		{ID: "C2", RiskStatus: "green"},
		{ID: "C3", RiskStatus: "yellow"},
		{ID: "C4", RiskStatus: "red"},
	}

	compliance := riskScorer.CalculateOverallCompliance(controls)
	expectedCompliance := 50.0 // 2 green out of 4 total

	if compliance != expectedCompliance {
		t.Errorf("Expected %.1f%% compliance, got %.1f%%", expectedCompliance, compliance)
	}

	t.Log("✓ Risk calculation works correctly")
}

// TestReportGeneration tests report generation
func TestReportGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create test state
	state := store.NewState()
	state.Frameworks = []types.Framework{
		{ID: "soc2", Name: "SOC 2", CompliancePercentage: 75.5},
	}
	state.Controls = []types.Control{
		{ID: "CC6.1", FrameworkID: "soc2", Title: "Test Control", RiskStatus: "green"},
	}
	state.Evidence = []types.Evidence{
		{EventID: "evt-001", ControlID: "CC6.1", ConfidenceLevel: "high"},
	}
	state.Findings = []types.Finding{
		{ID: "finding-001", ControlID: "CC6.1", Severity: "low", Title: "Test Finding"},
	}

	// Generate report
	exporter := report.NewExporter("1.0")
	reportObj, err := exporter.GenerateReport(
		state.Sources,
		state.Events,
		state.Frameworks,
		state.Controls,
		state.Evidence,
		state.Findings,
		"",
	)
	if err != nil {
		t.Fatalf("Report generation failed: %v", err)
	}

	// Verify report structure
	if reportObj.Metadata.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", reportObj.Metadata.Version)
	}
	if len(reportObj.Frameworks) != 1 {
		t.Errorf("Expected 1 framework, got %d", len(reportObj.Frameworks))
	}
	if reportObj.Summary.TotalControls != 1 {
		t.Errorf("Expected 1 control, got %d", reportObj.Summary.TotalControls)
	}
	if reportObj.Summary.TotalEvidence != 1 {
		t.Errorf("Expected 1 evidence, got %d", reportObj.Summary.TotalEvidence)
	}

	// Verify JSON serialization
	jsonData, err := json.MarshalIndent(reportObj, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal report: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Report JSON is empty")
	}

	t.Logf("✓ Report generation successful: %d bytes", len(jsonData))
}

// TestConfidenceCalculation tests confidence level determination
func TestConfidenceCalculation(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	calc := analyze.NewConfidenceCalculator()

	// Test different confidence levels
	testCases := []struct {
		score    int
		expected string
	}{
		{90, "high"},
		{70, "medium"},
		{40, "low"},
	}

	for _, tc := range testCases {
		level := calc.GetLevel(tc.score)
		if level != tc.expected {
			t.Errorf("Score %d: expected %s, got %s", tc.score, tc.expected, level)
		}
	}

	t.Log("✓ Confidence calculation works correctly")
}

// TestRiskSummary tests risk summary aggregation
func TestRiskSummary(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	riskScorer := analyze.NewRiskScorer()

	controls := []types.Control{
		{ID: "C1", RiskStatus: "green"},
		{ID: "C2", RiskStatus: "green"},
		{ID: "C3", RiskStatus: "yellow"},
		{ID: "C4", RiskStatus: "red"},
	}

	summary := riskScorer.GetRiskSummary(controls)

	if summary.GreenCount != 2 {
		t.Errorf("Expected 2 green, got %d", summary.GreenCount)
	}
	if summary.YellowCount != 1 {
		t.Errorf("Expected 1 yellow, got %d", summary.YellowCount)
	}
	if summary.RedCount != 1 {
		t.Errorf("Expected 1 red, got %d", summary.RedCount)
	}
	if summary.TotalCount != 4 {
		t.Errorf("Expected 4 total, got %d", summary.TotalCount)
	}

	t.Log("✓ Risk summary works correctly")
}
