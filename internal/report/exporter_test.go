package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// TestNewExporter verifies exporter initialization
func TestNewExporter(t *testing.T) {
	exporter := NewExporter("1.0.0")
	if exporter == nil {
		t.Fatal("Expected exporter to be created")
	}
	if exporter.version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", exporter.version)
	}
}

// TestGenerateReport verifies report generation
func TestGenerateReport(t *testing.T) {
	exporter := NewExporter("1.0.0")

	// Create test data
	sources := []types.Source{
		{ID: "src-1", Name: "Git Repo", Type: types.SourceTypeGit},
	}

	events := []types.Event{
		{ID: "evt-1", SourceID: "src-1", Title: "Test Event", Timestamp: time.Now()},
	}

	frameworks := []types.Framework{
		{ID: types.FrameworkSOC2, Name: "SOC 2", ControlCount: 45, CompliancePercentage: 80.0},
	}

	controls := []types.Control{
		{ID: "CC6.1", FrameworkID: types.FrameworkSOC2, Title: "Test Control", RiskStatus: "green"},
		{ID: "CC6.2", FrameworkID: types.FrameworkSOC2, Title: "Test Control 2", RiskStatus: "red"},
	}

	evidence := []types.Evidence{
		{ID: "ev-1", ControlID: "CC6.1", FrameworkID: types.FrameworkSOC2, EventID: "evt-1"},
	}

	findings := []types.Finding{
		{ID: "f-1", ControlID: "CC6.2", FrameworkID: types.FrameworkSOC2, Severity: types.SeverityCritical},
	}

	report, err := exporter.GenerateReport(sources, events, frameworks, controls, evidence, findings, types.RoleComplianceManager)
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Verify metadata
	if report.Metadata.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", report.Metadata.Version)
	}
	if report.Metadata.Role != types.RoleComplianceManager {
		t.Errorf("Expected role %s, got %s", types.RoleComplianceManager, report.Metadata.Role)
	}
	if report.Metadata.GeneratedAt.IsZero() {
		t.Error("Expected GeneratedAt to be set")
	}

	// Verify summary
	if report.Summary.TotalSources != 1 {
		t.Errorf("Expected 1 source, got %d", report.Summary.TotalSources)
	}
	if report.Summary.TotalEvents != 1 {
		t.Errorf("Expected 1 event, got %d", report.Summary.TotalEvents)
	}
	if report.Summary.TotalFrameworks != 1 {
		t.Errorf("Expected 1 framework, got %d", report.Summary.TotalFrameworks)
	}
	if report.Summary.TotalControls != 2 {
		t.Errorf("Expected 2 controls, got %d", report.Summary.TotalControls)
	}
	if report.Summary.TotalEvidence != 1 {
		t.Errorf("Expected 1 evidence, got %d", report.Summary.TotalEvidence)
	}
	if report.Summary.TotalFindings != 1 {
		t.Errorf("Expected 1 finding, got %d", report.Summary.TotalFindings)
	}
	if report.Summary.CriticalFindings != 1 {
		t.Errorf("Expected 1 critical finding, got %d", report.Summary.CriticalFindings)
	}

	// Verify frameworks
	if len(report.Frameworks) != 1 {
		t.Fatalf("Expected 1 framework report, got %d", len(report.Frameworks))
	}
	if len(report.Frameworks[0].Controls) != 2 {
		t.Errorf("Expected 2 controls in framework, got %d", len(report.Frameworks[0].Controls))
	}
}

// TestCalculateSummary verifies summary statistics calculation
func TestCalculateSummary(t *testing.T) {
	exporter := NewExporter("1.0.0")

	controls := []types.Control{
		{ID: "c1", RiskStatus: "green"},
		{ID: "c2", RiskStatus: "green"},
		{ID: "c3", RiskStatus: "yellow"},
		{ID: "c4", RiskStatus: "red"},
	}

	findings := []types.Finding{
		{ID: "f1", Severity: types.SeverityCritical},
		{ID: "f2", Severity: types.SeverityHigh},
		{ID: "f3", Severity: types.SeverityHigh},
		{ID: "f4", Severity: types.SeverityMedium},
		{ID: "f5", Severity: types.SeverityLow},
	}

	summary := exporter.calculateSummary(nil, nil, nil, controls, nil, findings)

	if summary.TotalControls != 4 {
		t.Errorf("Expected 4 controls, got %d", summary.TotalControls)
	}
	if summary.TotalFindings != 5 {
		t.Errorf("Expected 5 findings, got %d", summary.TotalFindings)
	}
	if summary.CriticalFindings != 1 {
		t.Errorf("Expected 1 critical finding, got %d", summary.CriticalFindings)
	}
	if summary.HighFindings != 2 {
		t.Errorf("Expected 2 high findings, got %d", summary.HighFindings)
	}
	if summary.MediumFindings != 1 {
		t.Errorf("Expected 1 medium finding, got %d", summary.MediumFindings)
	}
	if summary.LowFindings != 1 {
		t.Errorf("Expected 1 low finding, got %d", summary.LowFindings)
	}

	// Verify compliance calculation (2 green out of 4 = 50%)
	expectedCompliance := 50.0
	if summary.OverallCompliance != expectedCompliance {
		t.Errorf("Expected compliance %.2f%%, got %.2f%%", expectedCompliance, summary.OverallCompliance)
	}
}

// TestGroupByFramework verifies framework grouping
func TestGroupByFramework(t *testing.T) {
	exporter := NewExporter("1.0.0")

	frameworks := []types.Framework{
		{ID: types.FrameworkSOC2, Name: "SOC 2"},
		{ID: types.FrameworkISO27001, Name: "ISO 27001"},
	}

	controls := []types.Control{
		{ID: "CC6.1", FrameworkID: types.FrameworkSOC2},
		{ID: "CC6.2", FrameworkID: types.FrameworkSOC2},
		{ID: "A.5.1", FrameworkID: types.FrameworkISO27001},
	}

	evidence := []types.Evidence{
		{ID: "ev-1", ControlID: "CC6.1", FrameworkID: types.FrameworkSOC2},
		{ID: "ev-2", ControlID: "A.5.1", FrameworkID: types.FrameworkISO27001},
	}

	findings := []types.Finding{
		{ID: "f-1", ControlID: "CC6.2", FrameworkID: types.FrameworkSOC2},
	}

	reports := exporter.groupByFramework(frameworks, controls, evidence, findings)

	if len(reports) != 2 {
		t.Fatalf("Expected 2 framework reports, got %d", len(reports))
	}

	// Verify SOC 2 framework
	soc2Report := reports[0]
	if soc2Report.Framework.ID != types.FrameworkSOC2 {
		t.Errorf("Expected framework %s, got %s", types.FrameworkSOC2, soc2Report.Framework.ID)
	}
	if len(soc2Report.Controls) != 2 {
		t.Errorf("Expected 2 controls in SOC 2, got %d", len(soc2Report.Controls))
	}

	// Verify control evidence
	if len(soc2Report.Controls[0].Evidence) != 1 {
		t.Errorf("Expected 1 evidence for CC6.1, got %d", len(soc2Report.Controls[0].Evidence))
	}
	if len(soc2Report.Controls[1].Findings) != 1 {
		t.Errorf("Expected 1 finding for CC6.2, got %d", len(soc2Report.Controls[1].Findings))
	}

	// Verify ISO 27001 framework
	isoReport := reports[1]
	if isoReport.Framework.ID != types.FrameworkISO27001 {
		t.Errorf("Expected framework %s, got %s", types.FrameworkISO27001, isoReport.Framework.ID)
	}
	if len(isoReport.Controls) != 1 {
		t.Errorf("Expected 1 control in ISO 27001, got %d", len(isoReport.Controls))
	}
}

// TestExportToFile verifies file export functionality
func TestExportToFile(t *testing.T) {
	exporter := NewExporter("1.0.0")

	// Create a test report
	report := &Report{
		Metadata: ReportMetadata{
			GeneratedAt: time.Now(),
			Version:     "1.0.0",
			Role:        types.RoleComplianceManager,
		},
		Summary: ReportSummary{
			TotalSources:    1,
			TotalEvents:     10,
			TotalFrameworks: 1,
		},
	}

	// Create temporary directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "subdir", "report.json")

	// Export to file
	err := exporter.ExportToFile(report, filePath)
	if err != nil {
		t.Fatalf("Failed to export to file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Expected file to exist")
	}

	// Read and verify content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var loaded Report
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal report: %v", err)
	}

	if loaded.Metadata.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", loaded.Metadata.Version)
	}
	if loaded.Summary.TotalSources != 1 {
		t.Errorf("Expected 1 source, got %d", loaded.Summary.TotalSources)
	}
}

// TestExportToJSON verifies JSON export
func TestExportToJSON(t *testing.T) {
	exporter := NewExporter("1.0.0")

	report := &Report{
		Metadata: ReportMetadata{
			GeneratedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Version:     "1.0.0",
		},
		Summary: ReportSummary{
			TotalSources: 1,
		},
	}

	// Test pretty JSON
	prettyData, err := exporter.ExportToJSON(report, true)
	if err != nil {
		t.Fatalf("Failed to export pretty JSON: %v", err)
	}
	if len(prettyData) == 0 {
		t.Error("Expected non-empty JSON")
	}

	// Test compact JSON
	compactData, err := exporter.ExportToJSON(report, false)
	if err != nil {
		t.Fatalf("Failed to export compact JSON: %v", err)
	}
	if len(compactData) == 0 {
		t.Error("Expected non-empty JSON")
	}

	// Pretty should be longer due to indentation
	if len(prettyData) <= len(compactData) {
		t.Error("Expected pretty JSON to be longer than compact JSON")
	}

	// Verify both are valid JSON
	var prettyReport, compactReport Report
	if err := json.Unmarshal(prettyData, &prettyReport); err != nil {
		t.Errorf("Pretty JSON is invalid: %v", err)
	}
	if err := json.Unmarshal(compactData, &compactReport); err != nil {
		t.Errorf("Compact JSON is invalid: %v", err)
	}
}

// TestEmptyReport verifies handling of empty data
func TestEmptyReport(t *testing.T) {
	exporter := NewExporter("1.0.0")

	report, err := exporter.GenerateReport(nil, nil, nil, nil, nil, nil, "")
	if err != nil {
		t.Fatalf("Failed to generate empty report: %v", err)
	}

	if report.Summary.TotalSources != 0 {
		t.Errorf("Expected 0 sources, got %d", report.Summary.TotalSources)
	}
	if report.Summary.OverallCompliance != 0 {
		t.Errorf("Expected 0 compliance, got %.2f", report.Summary.OverallCompliance)
	}
}
