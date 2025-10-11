package report

import (
	"encoding/json"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Formatter handles report formatting and filtering
type Formatter struct{}

// NewFormatter creates a new report formatter
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FilterByRole filters report content based on user role
func (f *Formatter) FilterByRole(report *Report, role string) *Report {
	// Create a copy of the report
	filtered := &Report{
		Metadata:   report.Metadata,
		Summary:    report.Summary,
		Frameworks: make([]FrameworkReport, len(report.Frameworks)),
	}

	// Update role in metadata
	filtered.Metadata.Role = role

	// Copy framework reports
	copy(filtered.Frameworks, report.Frameworks)

	switch role {
	case types.RoleComplianceManager:
		// Compliance managers see everything
		filtered.Sources = report.Sources
		filtered.Events = report.Events
		filtered.Findings = report.Findings
		return filtered

	case types.RoleEngineer:
		// Engineers see filtered view - no raw sources/events, only findings and evidence
		filtered.Sources = nil
		filtered.Events = nil
		filtered.Findings = f.filterCriticalAndHighFindings(report.Findings)
		filtered.Frameworks = f.filterControlsWithFindings(report.Frameworks)
		return filtered

	default:
		// Unknown role - minimal view with just summary
		filtered.Sources = nil
		filtered.Events = nil
		filtered.Findings = nil
		filtered.Frameworks = f.filterToSummaryOnly(report.Frameworks)
		return filtered
	}
}

// filterCriticalAndHighFindings returns only critical and high severity findings
func (f *Formatter) filterCriticalAndHighFindings(findings []types.Finding) []types.Finding {
	filtered := make([]types.Finding, 0)
	for _, finding := range findings {
		if finding.Severity == types.SeverityCritical || finding.Severity == types.SeverityHigh {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

// filterControlsWithFindings returns only controls that have findings
func (f *Formatter) filterControlsWithFindings(frameworks []FrameworkReport) []FrameworkReport {
	filtered := make([]FrameworkReport, 0, len(frameworks))

	for _, fw := range frameworks {
		controls := make([]ControlReport, 0)
		for _, ctrl := range fw.Controls {
			if len(ctrl.Findings) > 0 {
				controls = append(controls, ctrl)
			}
		}

		if len(controls) > 0 {
			filtered = append(filtered, FrameworkReport{
				Framework: fw.Framework,
				Controls:  controls,
			})
		}
	}

	return filtered
}

// filterToSummaryOnly removes all detailed data, keeping only framework summaries
func (f *Formatter) filterToSummaryOnly(frameworks []FrameworkReport) []FrameworkReport {
	filtered := make([]FrameworkReport, 0, len(frameworks))

	for _, fw := range frameworks {
		filtered = append(filtered, FrameworkReport{
			Framework: fw.Framework,
			Controls:  nil, // Remove all control details
		})
	}

	return filtered
}

// FormatJSON formats the report as pretty-printed JSON
func (f *Formatter) FormatJSON(report *Report, indent bool) ([]byte, error) {
	if indent {
		return json.MarshalIndent(report, "", "  ")
	}
	return json.Marshal(report)
}

// FormatSummary returns just the summary section as JSON
func (f *Formatter) FormatSummary(report *Report) ([]byte, error) {
	return json.MarshalIndent(report.Summary, "", "  ")
}

// FormatMetadata returns just the metadata section as JSON
func (f *Formatter) FormatMetadata(report *Report) ([]byte, error) {
	return json.MarshalIndent(report.Metadata, "", "  ")
}

// GetFrameworkSummaries extracts framework summary information
func (f *Formatter) GetFrameworkSummaries(report *Report) []FrameworkSummary {
	summaries := make([]FrameworkSummary, 0, len(report.Frameworks))

	for _, fw := range report.Frameworks {
		summary := FrameworkSummary{
			ID:                   fw.Framework.ID,
			Name:                 fw.Framework.Name,
			TotalControls:        len(fw.Controls),
			CompliancePercentage: fw.Framework.CompliancePercentage,
		}

		// Count controls by risk status
		for _, ctrl := range fw.Controls {
			switch ctrl.Control.RiskStatus {
			case "green":
				summary.GreenControls++
			case "yellow":
				summary.YellowControls++
			case "red":
				summary.RedControls++
			}

			summary.TotalEvidence += len(ctrl.Evidence)
			summary.TotalFindings += len(ctrl.Findings)
		}

		summaries = append(summaries, summary)
	}

	return summaries
}

// FrameworkSummary provides high-level framework statistics
type FrameworkSummary struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	TotalControls        int     `json:"total_controls"`
	GreenControls        int     `json:"green_controls"`
	YellowControls       int     `json:"yellow_controls"`
	RedControls          int     `json:"red_controls"`
	TotalEvidence        int     `json:"total_evidence"`
	TotalFindings        int     `json:"total_findings"`
	CompliancePercentage float64 `json:"compliance_percentage"`
}
