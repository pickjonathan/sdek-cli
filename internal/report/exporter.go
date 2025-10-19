package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Report represents a complete compliance report
type Report struct {
	Metadata   ReportMetadata    `json:"metadata"`
	Summary    ReportSummary     `json:"summary"`
	Frameworks []FrameworkReport `json:"frameworks"`
	Sources    []types.Source    `json:"sources,omitempty"`
	Events     []types.Event     `json:"events,omitempty"`
	Findings   []types.Finding   `json:"findings,omitempty"`
}

// ReportMetadata contains report generation information
type ReportMetadata struct {
	GeneratedAt time.Time `json:"generated_at"`
	Version     string    `json:"version"`
	Role        string    `json:"role,omitempty"`
}

// ReportSummary contains high-level statistics
type ReportSummary struct {
	TotalSources      int     `json:"total_sources"`
	TotalEvents       int     `json:"total_events"`
	TotalFrameworks   int     `json:"total_frameworks"`
	TotalControls     int     `json:"total_controls"`
	TotalEvidence     int     `json:"total_evidence"`
	TotalFindings     int     `json:"total_findings"`
	OverallCompliance float64 `json:"overall_compliance_percentage"`
	CriticalFindings  int     `json:"critical_findings"`
	HighFindings      int     `json:"high_findings"`
	MediumFindings    int     `json:"medium_findings"`
	LowFindings       int     `json:"low_findings"`
}

// FrameworkReport contains framework-specific analysis
type FrameworkReport struct {
	Framework types.Framework `json:"framework"`
	Controls  []ControlReport `json:"controls"`
}

// ControlReport contains control-specific details
type ControlReport struct {
	Control  types.Control    `json:"control"`
	Evidence []types.Evidence `json:"evidence"`
	Findings []types.Finding  `json:"findings"`
}

// Exporter generates compliance reports
type Exporter struct {
	version string
}

// NewExporter creates a new report exporter
func NewExporter(version string) *Exporter {
	return &Exporter{
		version: version,
	}
}

// GenerateReport creates a complete compliance report from state data
func (e *Exporter) GenerateReport(
	sources []types.Source,
	events []types.Event,
	frameworks []types.Framework,
	controls []types.Control,
	evidence []types.Evidence,
	findings []types.Finding,
	role string,
) (*Report, error) {
	// Calculate summary statistics
	summary := e.calculateSummary(sources, events, frameworks, controls, evidence, findings)

	// Group data by framework
	frameworkReports := e.groupByFramework(frameworks, controls, evidence, findings)

	// Create report
	report := &Report{
		Metadata: ReportMetadata{
			GeneratedAt: time.Now(),
			Version:     e.version,
			Role:        role,
		},
		Summary:    summary,
		Frameworks: frameworkReports,
		Sources:    sources,
		Events:     events,
		Findings:   findings,
	}

	return report, nil
}

// calculateSummary computes summary statistics
func (e *Exporter) calculateSummary(
	sources []types.Source,
	events []types.Event,
	frameworks []types.Framework,
	controls []types.Control,
	evidence []types.Evidence,
	findings []types.Finding,
) ReportSummary {
	summary := ReportSummary{
		TotalSources:    len(sources),
		TotalEvents:     len(events),
		TotalFrameworks: len(frameworks),
		TotalControls:   len(controls),
		TotalEvidence:   len(evidence),
		TotalFindings:   len(findings),
	}

	// Count findings by severity
	for _, finding := range findings {
		switch finding.Severity {
		case types.SeverityCritical:
			summary.CriticalFindings++
		case types.SeverityHigh:
			summary.HighFindings++
		case types.SeverityMedium:
			summary.MediumFindings++
		case types.SeverityLow:
			summary.LowFindings++
		}
	}

	// Calculate overall compliance
	if len(controls) > 0 {
		greenCount := 0
		for _, control := range controls {
			if control.RiskStatus == "green" {
				greenCount++
			}
		}
		summary.OverallCompliance = float64(greenCount) / float64(len(controls)) * 100
	}

	return summary
}

// groupByFramework organizes controls, evidence, and findings by framework
func (e *Exporter) groupByFramework(
	frameworks []types.Framework,
	controls []types.Control,
	evidence []types.Evidence,
	findings []types.Finding,
) []FrameworkReport {
	frameworkReports := make([]FrameworkReport, 0, len(frameworks))

	for _, fw := range frameworks {
		// Find controls for this framework
		var frameworkControls []types.Control
		for _, control := range controls {
			if control.FrameworkID == fw.ID {
				frameworkControls = append(frameworkControls, control)
			}
		}

		// Build control reports
		controlReports := make([]ControlReport, 0, len(frameworkControls))
		for _, control := range frameworkControls {
			// Find evidence for this control
			var controlEvidence []types.Evidence
			for _, ev := range evidence {
				if ev.ControlID == control.ID && ev.FrameworkID == fw.ID {
					controlEvidence = append(controlEvidence, ev)
				}
			}

			// Find findings for this control
			var controlFindings []types.Finding
			for _, finding := range findings {
				if finding.ControlID == control.ID && finding.FrameworkID == fw.ID {
					controlFindings = append(controlFindings, finding)
				}
			}

			controlReports = append(controlReports, ControlReport{
				Control:  control,
				Evidence: controlEvidence,
				Findings: controlFindings,
			})
		}

		frameworkReports = append(frameworkReports, FrameworkReport{
			Framework: fw,
			Controls:  controlReports,
		})
	}

	return frameworkReports
}

// ExportToFile saves the report to a JSON file
func (e *Exporter) ExportToFile(report *Report, filePath string) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportToJSON converts the report to JSON bytes
func (e *Exporter) ExportToJSON(report *Report, pretty bool) ([]byte, error) {
	var data []byte
	var err error

	if pretty {
		data, err = json.MarshalIndent(report, "", "  ")
	} else {
		data, err = json.Marshal(report)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal report: %w", err)
	}

	return data, nil
}
