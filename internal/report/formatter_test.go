package report

import (
	"encoding/json"
	"testing"
	"time"

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

// TestFormatCSV_WithAIMetadata verifies CSV export includes AI analysis fields
func TestFormatCSV_WithAIMetadata(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	formatter := NewFormatter()

	report := &Report{
		Frameworks: []FrameworkReport{
			{
				Framework: types.Framework{
					ID:   types.FrameworkSOC2,
					Name: "SOC 2",
				},
				Controls: []ControlReport{
					{
						Control: types.Control{
							ID:    "CC6.1",
							Title: "Logical Access Controls",
						},
						Evidence: []types.Evidence{
							{
								ID:                  "ev-1",
								EventID:             "evt-1",
								ControlID:           "CC6.1",
								ConfidenceScore:     85.5,
								ConfidenceLevel:     "high",
								AIAnalyzed:          true,
								AIConfidence:        90,
								HeuristicConfidence: 75,
								CombinedConfidence:  86,
								AIJustification:     "Strong evidence of MFA implementation",
								AIResidualRisk:      "No multi-region support",
								AnalysisMethod:      "ai+heuristic",
								Keywords:            []string{"authentication", "MFA"},
								Reasoning:           "Matches authentication keywords",
							},
						},
					},
				},
			},
		},
	}

	csv := formatter.FormatCSV(report)

	// Verify CSV header includes AI fields
	if !contains(csv, "AI Analyzed") {
		t.Error("CSV should include 'AI Analyzed' column")
	}
	if !contains(csv, "AI Confidence") {
		t.Error("CSV should include 'AI Confidence' column")
	}
	if !contains(csv, "Heuristic Confidence") {
		t.Error("CSV should include 'Heuristic Confidence' column")
	}
	if !contains(csv, "Combined Confidence") {
		t.Error("CSV should include 'Combined Confidence' column")
	}
	if !contains(csv, "AI Justification") {
		t.Error("CSV should include 'AI Justification' column")
	}
	if !contains(csv, "Residual Risk") {
		t.Error("CSV should include 'Residual Risk' column")
	}
	if !contains(csv, "Analysis Method") {
		t.Error("CSV should include 'Analysis Method' column")
	}

	// Verify AI-analyzed evidence has populated fields
	if !contains(csv, "Yes") {
		t.Error("CSV should show 'Yes' for AI Analyzed")
	}
	if !contains(csv, "90") {
		t.Error("CSV should include AI confidence value")
	}
	if !contains(csv, "75") {
		t.Error("CSV should include heuristic confidence value")
	}
	if !contains(csv, "86") {
		t.Error("CSV should include combined confidence value")
	}
	if !contains(csv, "Strong evidence of MFA implementation") {
		t.Error("CSV should include AI justification")
	}
	if !contains(csv, "No multi-region support") {
		t.Error("CSV should include residual risk")
	}
	if !contains(csv, "ai+heuristic") {
		t.Error("CSV should include analysis method")
	}
}

// TestFormatCSV_HeuristicOnly verifies CSV shows empty AI fields for heuristic-only evidence
func TestFormatCSV_HeuristicOnly(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	formatter := NewFormatter()

	report := &Report{
		Frameworks: []FrameworkReport{
			{
				Framework: types.Framework{
					ID:   types.FrameworkSOC2,
					Name: "SOC 2",
				},
				Controls: []ControlReport{
					{
						Control: types.Control{
							ID:    "CC6.1",
							Title: "Logical Access Controls",
						},
						Evidence: []types.Evidence{
							{
								ID:                  "ev-1",
								EventID:             "evt-1",
								ControlID:           "CC6.1",
								ConfidenceScore:     60.0,
								ConfidenceLevel:     "medium",
								AIAnalyzed:          false,
								HeuristicConfidence: 60,
								CombinedConfidence:  60,
								AnalysisMethod:      "heuristic-only",
								Keywords:            []string{"authentication"},
								Reasoning:           "Matches authentication keyword",
							},
						},
					},
				},
			},
		},
	}

	csv := formatter.FormatCSV(report)

	// Verify CSV shows "No" for AI Analyzed
	lines := splitLines(csv)
	if len(lines) < 2 {
		t.Fatal("CSV should have header and data lines")
	}

	dataLine := lines[1]
	if !contains(dataLine, "No") {
		t.Error("CSV should show 'No' for non-AI analyzed evidence")
	}
	if !contains(dataLine, "heuristic-only") {
		t.Error("CSV should show 'heuristic-only' analysis method")
	}

	// AI-specific fields should be empty (no values between commas or quoted empty strings)
	// The data line should have empty AI confidence, justification, and residual risk
}

// TestFormatMarkdown_WithAIAnalysis verifies Markdown includes AI analysis section
func TestFormatMarkdown_WithAIAnalysis(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	formatter := NewFormatter()

	report := &Report{
		Metadata: ReportMetadata{
			GeneratedAt: testTime(),
			Version:     "1.0.0",
			Role:        types.RoleComplianceManager,
		},
		Summary: ReportSummary{
			TotalSources:    2,
			TotalEvents:     5,
			TotalFrameworks: 1,
			TotalControls:   3,
			TotalEvidence:   4,
		},
		Frameworks: []FrameworkReport{
			{
				Framework: types.Framework{
					ID:                   types.FrameworkSOC2,
					Name:                 "SOC 2",
					CompliancePercentage: 85.5,
				},
				Controls: []ControlReport{
					{
						Control: types.Control{
							ID:          "CC6.1",
							Title:       "Logical Access Controls",
							Description: "Implement authentication controls",
							RiskStatus:  "green",
						},
						Evidence: []types.Evidence{
							{
								ID:                  "ev-1",
								EventID:             "evt-1",
								ControlID:           "CC6.1",
								ConfidenceScore:     85.5,
								ConfidenceLevel:     "high",
								AIAnalyzed:          true,
								AIConfidence:        90,
								HeuristicConfidence: 75,
								CombinedConfidence:  86,
								AIJustification:     "Strong evidence of MFA implementation with OAuth",
								AIResidualRisk:      "No multi-region failover support yet",
								AnalysisMethod:      "ai+heuristic",
								Keywords:            []string{"authentication", "MFA"},
								Reasoning:           "Matches authentication keywords",
							},
						},
					},
				},
			},
		},
	}

	md := formatter.FormatMarkdown(report)

	// Verify markdown structure
	if !contains(md, "# Compliance Report") {
		t.Error("Markdown should have report title")
	}
	if !contains(md, "## Report Metadata") {
		t.Error("Markdown should have metadata section")
	}
	if !contains(md, "## Summary") {
		t.Error("Markdown should have summary section")
	}
	if !contains(md, "## Framework: SOC 2") {
		t.Error("Markdown should have framework section")
	}

	// Verify AI Analysis section
	if !contains(md, "**AI Analysis:**") {
		t.Error("Markdown should have AI Analysis section for AI-analyzed evidence")
	}
	if !contains(md, "AI Confidence: 90%") {
		t.Error("Markdown should show AI confidence")
	}
	if !contains(md, "Heuristic Confidence: 75%") {
		t.Error("Markdown should show heuristic confidence")
	}
	if !contains(md, "Combined Confidence: 86%") {
		t.Error("Markdown should show combined confidence")
	}
	if !contains(md, "Justification: Strong evidence of MFA implementation with OAuth") {
		t.Error("Markdown should show AI justification")
	}
	if !contains(md, "Residual Risk: No multi-region failover support yet") {
		t.Error("Markdown should show residual risk")
	}
	if !contains(md, "Analysis Method: ai+heuristic") {
		t.Error("Markdown should show analysis method")
	}
}

// TestFormatMarkdown_HeuristicOnly verifies Markdown excludes AI section for heuristic-only evidence
func TestFormatMarkdown_HeuristicOnly(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	formatter := NewFormatter()

	report := &Report{
		Metadata: ReportMetadata{
			GeneratedAt: testTime(),
			Version:     "1.0.0",
		},
		Summary: ReportSummary{
			TotalEvidence: 1,
		},
		Frameworks: []FrameworkReport{
			{
				Framework: types.Framework{
					ID:   types.FrameworkSOC2,
					Name: "SOC 2",
				},
				Controls: []ControlReport{
					{
						Control: types.Control{
							ID:          "CC6.1",
							Title:       "Logical Access Controls",
							Description: "Implement authentication controls",
						},
						Evidence: []types.Evidence{
							{
								ID:              "ev-1",
								EventID:         "evt-1",
								ControlID:       "CC6.1",
								ConfidenceScore: 60.0,
								ConfidenceLevel: "medium",
								AIAnalyzed:      false,
								AnalysisMethod:  "heuristic-only",
								Keywords:        []string{"authentication"},
								Reasoning:       "Matches authentication keyword",
							},
						},
					},
				},
			},
		},
	}

	md := formatter.FormatMarkdown(report)

	// Verify no AI Analysis section for heuristic-only evidence
	if contains(md, "**AI Analysis:**") {
		t.Error("Markdown should not have AI Analysis section for heuristic-only evidence")
	}
	if contains(md, "AI Confidence:") {
		t.Error("Markdown should not show AI confidence for heuristic-only evidence")
	}

	// But should still show analysis method
	if !contains(md, "Analysis Method: heuristic-only") {
		t.Error("Markdown should show heuristic-only analysis method")
	}
}

// Helper functions for tests

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	lines := []string{}
	current := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func testTime() time.Time {
	// Return a fixed time for consistent testing
	return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
}
