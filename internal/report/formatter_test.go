package report

import (
	"encoding/json"
	"testing"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// TestNewFormatter verifies formatter initialization
func TestNewFormatter(t *testing.T) {
	formatter := NewFormatter()
	if formatter == nil {
		t.Fatal("Expected formatter to be created")
	}
}

// TestFilterByRoleComplianceManager verifies compliance manager gets full access
func TestFilterByRoleComplianceManager(t *testing.T) {
	formatter := NewFormatter()
	
	report := createTestReport()
	
	filtered := formatter.FilterByRole(report, types.RoleComplianceManager)
	
	// Compliance manager should see everything
	if len(filtered.Sources) != len(report.Sources) {
		t.Errorf("Expected %d sources, got %d", len(report.Sources), len(filtered.Sources))
	}
	if len(filtered.Events) != len(report.Events) {
		t.Errorf("Expected %d events, got %d", len(report.Events), len(filtered.Events))
	}
	if len(filtered.Findings) != len(report.Findings) {
		t.Errorf("Expected %d findings, got %d", len(report.Findings), len(filtered.Findings))
	}
	if len(filtered.Frameworks) != len(report.Frameworks) {
		t.Errorf("Expected %d frameworks, got %d", len(report.Frameworks), len(filtered.Frameworks))
	}
	if filtered.Metadata.Role != types.RoleComplianceManager {
		t.Errorf("Expected role %s, got %s", types.RoleComplianceManager, filtered.Metadata.Role)
	}
}

// TestFilterByRoleEngineer verifies engineer gets filtered view
func TestFilterByRoleEngineer(t *testing.T) {
	formatter := NewFormatter()
	
	report := createTestReport()
	
	filtered := formatter.FilterByRole(report, types.RoleEngineer)
	
	// Engineers should not see sources and events
	if filtered.Sources != nil {
		t.Error("Expected sources to be nil for engineer role")
	}
	if filtered.Events != nil {
		t.Error("Expected events to be nil for engineer role")
	}
	
	// Engineers should only see critical and high findings
	if len(filtered.Findings) != 2 { // 1 critical + 1 high from test data
		t.Errorf("Expected 2 critical/high findings, got %d", len(filtered.Findings))
	}
	
	for _, finding := range filtered.Findings {
		if finding.Severity != types.SeverityCritical && finding.Severity != types.SeverityHigh {
			t.Errorf("Expected only critical/high findings, got %s", finding.Severity)
		}
	}
	
	// Engineers should only see frameworks with findings
	if len(filtered.Frameworks) == 0 {
		t.Error("Expected at least one framework with findings")
	}
	
	for _, fw := range filtered.Frameworks {
		hasFindings := false
		for _, ctrl := range fw.Controls {
			if len(ctrl.Findings) > 0 {
				hasFindings = true
				break
			}
		}
		if !hasFindings {
			t.Error("Expected frameworks to only contain controls with findings")
		}
	}
	
	if filtered.Metadata.Role != types.RoleEngineer {
		t.Errorf("Expected role %s, got %s", types.RoleEngineer, filtered.Metadata.Role)
	}
}

// TestFilterByRoleUnknown verifies unknown role gets minimal view
func TestFilterByRoleUnknown(t *testing.T) {
	formatter := NewFormatter()
	
	report := createTestReport()
	
	filtered := formatter.FilterByRole(report, "unknown")
	
	// Unknown role should see minimal data
	if filtered.Sources != nil {
		t.Error("Expected sources to be nil for unknown role")
	}
	if filtered.Events != nil {
		t.Error("Expected events to be nil for unknown role")
	}
	if filtered.Findings != nil {
		t.Error("Expected findings to be nil for unknown role")
	}
	
	// Should see framework summaries only (no control details)
	for _, fw := range filtered.Frameworks {
		if fw.Controls != nil {
			t.Error("Expected no control details for unknown role")
		}
	}
}

// TestFilterCriticalAndHighFindings verifies finding filtering
func TestFilterCriticalAndHighFindings(t *testing.T) {
	formatter := NewFormatter()
	
	findings := []types.Finding{
		{ID: "f1", Severity: types.SeverityCritical},
		{ID: "f2", Severity: types.SeverityHigh},
		{ID: "f3", Severity: types.SeverityMedium},
		{ID: "f4", Severity: types.SeverityLow},
		{ID: "f5", Severity: types.SeverityHigh},
	}
	
	filtered := formatter.filterCriticalAndHighFindings(findings)
	
	if len(filtered) != 3 {
		t.Errorf("Expected 3 critical/high findings, got %d", len(filtered))
	}
	
	for _, finding := range filtered {
		if finding.Severity != types.SeverityCritical && finding.Severity != types.SeverityHigh {
			t.Errorf("Expected only critical/high findings, got %s", finding.Severity)
		}
	}
}

// TestFilterControlsWithFindings verifies control filtering
func TestFilterControlsWithFindings(t *testing.T) {
	formatter := NewFormatter()
	
	frameworks := []FrameworkReport{
		{
			Framework: types.Framework{ID: types.FrameworkSOC2},
			Controls: []ControlReport{
				{
					Control:  types.Control{ID: "c1"},
					Findings: []types.Finding{{ID: "f1"}},
				},
				{
					Control:  types.Control{ID: "c2"},
					Findings: nil, // No findings
				},
			},
		},
		{
			Framework: types.Framework{ID: types.FrameworkISO27001},
			Controls: []ControlReport{
				{
					Control:  types.Control{ID: "c3"},
					Findings: nil, // No findings
				},
			},
		},
	}
	
	filtered := formatter.filterControlsWithFindings(frameworks)
	
	// Should only have SOC2 framework (has findings)
	if len(filtered) != 1 {
		t.Errorf("Expected 1 framework with findings, got %d", len(filtered))
	}
	
	if filtered[0].Framework.ID != types.FrameworkSOC2 {
		t.Errorf("Expected framework %s, got %s", types.FrameworkSOC2, filtered[0].Framework.ID)
	}
	
	// Should only have c1 control (has findings)
	if len(filtered[0].Controls) != 1 {
		t.Errorf("Expected 1 control with findings, got %d", len(filtered[0].Controls))
	}
	
	if filtered[0].Controls[0].Control.ID != "c1" {
		t.Errorf("Expected control c1, got %s", filtered[0].Controls[0].Control.ID)
	}
}

// TestFormatJSON verifies JSON formatting
func TestFormatJSON(t *testing.T) {
	formatter := NewFormatter()
	report := createTestReport()
	
	// Test indented JSON
	indented, err := formatter.FormatJSON(report, true)
	if err != nil {
		t.Fatalf("Failed to format indented JSON: %v", err)
	}
	if len(indented) == 0 {
		t.Error("Expected non-empty JSON")
	}
	
	// Test compact JSON
	compact, err := formatter.FormatJSON(report, false)
	if err != nil {
		t.Fatalf("Failed to format compact JSON: %v", err)
	}
	if len(compact) == 0 {
		t.Error("Expected non-empty JSON")
	}
	
	// Indented should be longer
	if len(indented) <= len(compact) {
		t.Error("Expected indented JSON to be longer than compact")
	}
	
	// Verify both are valid JSON
	var r1, r2 Report
	if err := json.Unmarshal(indented, &r1); err != nil {
		t.Errorf("Indented JSON is invalid: %v", err)
	}
	if err := json.Unmarshal(compact, &r2); err != nil {
		t.Errorf("Compact JSON is invalid: %v", err)
	}
}

// TestFormatSummary verifies summary-only formatting
func TestFormatSummary(t *testing.T) {
	formatter := NewFormatter()
	report := createTestReport()
	
	data, err := formatter.FormatSummary(report)
	if err != nil {
		t.Fatalf("Failed to format summary: %v", err)
	}
	
	var summary ReportSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		t.Fatalf("Failed to unmarshal summary: %v", err)
	}
	
	if summary.TotalSources != report.Summary.TotalSources {
		t.Error("Summary mismatch")
	}
}

// TestFormatMetadata verifies metadata-only formatting
func TestFormatMetadata(t *testing.T) {
	formatter := NewFormatter()
	report := createTestReport()
	
	data, err := formatter.FormatMetadata(report)
	if err != nil {
		t.Fatalf("Failed to format metadata: %v", err)
	}
	
	var metadata ReportMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}
	
	if metadata.Version != report.Metadata.Version {
		t.Error("Metadata mismatch")
	}
}

// TestGetFrameworkSummaries verifies framework summary extraction
func TestGetFrameworkSummaries(t *testing.T) {
	formatter := NewFormatter()
	
	report := &Report{
		Frameworks: []FrameworkReport{
			{
				Framework: types.Framework{
					ID:                   types.FrameworkSOC2,
					Name:                 "SOC 2",
					CompliancePercentage: 75.0,
				},
				Controls: []ControlReport{
					{
						Control:  types.Control{ID: "c1", RiskStatus: "green"},
						Evidence: []types.Evidence{{ID: "e1"}, {ID: "e2"}},
						Findings: []types.Finding{{ID: "f1"}},
					},
					{
						Control:  types.Control{ID: "c2", RiskStatus: "yellow"},
						Evidence: []types.Evidence{{ID: "e3"}},
						Findings: nil,
					},
					{
						Control:  types.Control{ID: "c3", RiskStatus: "red"},
						Evidence: nil,
						Findings: []types.Finding{{ID: "f2"}, {ID: "f3"}},
					},
				},
			},
		},
	}
	
	summaries := formatter.GetFrameworkSummaries(report)
	
	if len(summaries) != 1 {
		t.Fatalf("Expected 1 summary, got %d", len(summaries))
	}
	
	summary := summaries[0]
	if summary.ID != types.FrameworkSOC2 {
		t.Errorf("Expected framework %s, got %s", types.FrameworkSOC2, summary.ID)
	}
	if summary.TotalControls != 3 {
		t.Errorf("Expected 3 controls, got %d", summary.TotalControls)
	}
	if summary.GreenControls != 1 {
		t.Errorf("Expected 1 green control, got %d", summary.GreenControls)
	}
	if summary.YellowControls != 1 {
		t.Errorf("Expected 1 yellow control, got %d", summary.YellowControls)
	}
	if summary.RedControls != 1 {
		t.Errorf("Expected 1 red control, got %d", summary.RedControls)
	}
	if summary.TotalEvidence != 3 {
		t.Errorf("Expected 3 evidence, got %d", summary.TotalEvidence)
	}
	if summary.TotalFindings != 3 {
		t.Errorf("Expected 3 findings, got %d", summary.TotalFindings)
	}
	if summary.CompliancePercentage != 75.0 {
		t.Errorf("Expected 75%% compliance, got %.2f%%", summary.CompliancePercentage)
	}
}

// Helper function to create test report
func createTestReport() *Report {
	return &Report{
		Metadata: ReportMetadata{
			Version: "1.0.0",
			Role:    types.RoleComplianceManager,
		},
		Summary: ReportSummary{
			TotalSources:    2,
			TotalEvents:     10,
			TotalFrameworks: 1,
		},
		Sources: []types.Source{
			{ID: "src-1", Name: "Git"},
			{ID: "src-2", Name: "Jira"},
		},
		Events: []types.Event{
			{ID: "evt-1", SourceID: "src-1"},
		},
		Findings: []types.Finding{
			{ID: "f1", Severity: types.SeverityCritical},
			{ID: "f2", Severity: types.SeverityHigh},
			{ID: "f3", Severity: types.SeverityMedium},
			{ID: "f4", Severity: types.SeverityLow},
		},
		Frameworks: []FrameworkReport{
			{
				Framework: types.Framework{ID: types.FrameworkSOC2},
				Controls: []ControlReport{
					{
						Control:  types.Control{ID: "c1"},
						Findings: []types.Finding{{ID: "f1"}},
					},
					{
						Control:  types.Control{ID: "c2"},
						Findings: nil,
					},
				},
			},
		},
	}
}
