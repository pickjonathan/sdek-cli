package report

import (
	"encoding/json"
	"fmt"

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

// FormatCSV exports evidence data as CSV with AI analysis fields
func (f *Formatter) FormatCSV(report *Report) string {
	csv := "Framework,Control ID,Control Name,Event ID,Confidence Score,Confidence Level,Analysis Method,AI Analyzed,AI Confidence,Heuristic Confidence,Combined Confidence,AI Justification,Residual Risk,Keywords,Reasoning,Mapped At\n"

	for _, fw := range report.Frameworks {
		for _, ctrl := range fw.Controls {
			for _, evidence := range ctrl.Evidence {
				// Format boolean as Yes/No
				aiAnalyzed := "No"
				if evidence.AIAnalyzed {
					aiAnalyzed = "Yes"
				}

				// Format AI fields (empty if not analyzed)
				aiConfidence := ""
				aiJustification := ""
				residualRisk := ""
				if evidence.AIAnalyzed {
					aiConfidence = fmt.Sprintf("%d", evidence.AIConfidence)
					aiJustification = escapeCSV(evidence.AIJustification)
					residualRisk = escapeCSV(evidence.AIResidualRisk)
				}

				// Format heuristic and combined confidence
				heuristicConfidence := fmt.Sprintf("%d", evidence.HeuristicConfidence)
				combinedConfidence := fmt.Sprintf("%d", evidence.CombinedConfidence)

				// Format keywords
				keywords := ""
				if len(evidence.Keywords) > 0 {
					keywordBytes, _ := json.Marshal(evidence.Keywords)
					keywords = escapeCSV(string(keywordBytes))
				}

				csv += fmt.Sprintf("%s,%s,%s,%s,%.2f,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
					escapeCSV(fw.Framework.Name),
					escapeCSV(evidence.ControlID),
					escapeCSV(ctrl.Control.Title),
					escapeCSV(evidence.EventID),
					evidence.ConfidenceScore,
					escapeCSV(evidence.ConfidenceLevel),
					escapeCSV(evidence.AnalysisMethod),
					aiAnalyzed,
					aiConfidence,
					heuristicConfidence,
					combinedConfidence,
					aiJustification,
					residualRisk,
					keywords,
					escapeCSV(evidence.Reasoning),
					evidence.MappedAt.Format("2006-01-02 15:04:05"),
				)
			}
		}
	}

	return csv
}

// FormatMarkdown exports report as Markdown with AI analysis details
func (f *Formatter) FormatMarkdown(report *Report) string {
	md := "# Compliance Report\n\n"

	// Metadata
	md += "## Report Metadata\n\n"
	md += fmt.Sprintf("- **Generated:** %s\n", report.Metadata.GeneratedAt.Format("2006-01-02 15:04:05"))
	md += fmt.Sprintf("- **Version:** %s\n", report.Metadata.Version)
	if report.Metadata.Role != "" {
		md += fmt.Sprintf("- **Role:** %s\n", report.Metadata.Role)
	}
	md += "\n"

	// Summary
	md += "## Summary\n\n"
	md += fmt.Sprintf("- **Total Sources:** %d\n", report.Summary.TotalSources)
	md += fmt.Sprintf("- **Total Events:** %d\n", report.Summary.TotalEvents)
	md += fmt.Sprintf("- **Total Frameworks:** %d\n", report.Summary.TotalFrameworks)
	md += fmt.Sprintf("- **Total Controls:** %d\n", report.Summary.TotalControls)
	md += fmt.Sprintf("- **Total Evidence:** %d\n", report.Summary.TotalEvidence)
	md += fmt.Sprintf("- **Overall Compliance:** %.2f%%\n", report.Summary.OverallCompliance)
	md += fmt.Sprintf("- **Findings:** %d Critical, %d High, %d Medium, %d Low\n\n",
		report.Summary.CriticalFindings,
		report.Summary.HighFindings,
		report.Summary.MediumFindings,
		report.Summary.LowFindings,
	)

	// Frameworks
	for _, fw := range report.Frameworks {
		md += fmt.Sprintf("## Framework: %s\n\n", fw.Framework.Name)
		md += fmt.Sprintf("**Compliance:** %.2f%%\n\n", fw.Framework.CompliancePercentage)

		for _, ctrl := range fw.Controls {
			md += fmt.Sprintf("### Control: %s - %s\n\n", ctrl.Control.ID, ctrl.Control.Title)
			md += fmt.Sprintf("**Description:** %s\n\n", ctrl.Control.Description)
			md += fmt.Sprintf("**Risk Status:** %s\n\n", ctrl.Control.RiskStatus)

			if len(ctrl.Evidence) > 0 {
				md += "#### Evidence\n\n"
				for i, evidence := range ctrl.Evidence {
					md += fmt.Sprintf("%d. **Event ID:** %s\n", i+1, evidence.EventID)
					md += fmt.Sprintf("   - **Confidence:** %.2f%% (%s)\n", evidence.ConfidenceScore, evidence.ConfidenceLevel)
					md += fmt.Sprintf("   - **Analysis Method:** %s\n", evidence.AnalysisMethod)
					md += fmt.Sprintf("   - **Mapped At:** %s\n", evidence.MappedAt.Format("2006-01-02 15:04:05"))

					// AI Analysis section (if applicable)
					if evidence.AIAnalyzed {
						md += "\n   **AI Analysis:**\n"
						md += fmt.Sprintf("   - AI Confidence: %d%%\n", evidence.AIConfidence)
						md += fmt.Sprintf("   - Heuristic Confidence: %d%%\n", evidence.HeuristicConfidence)
						md += fmt.Sprintf("   - Combined Confidence: %d%%\n", evidence.CombinedConfidence)
						if evidence.AIJustification != "" {
							md += fmt.Sprintf("   - Justification: %s\n", evidence.AIJustification)
						}
						if evidence.AIResidualRisk != "" {
							md += fmt.Sprintf("   - Residual Risk: %s\n", evidence.AIResidualRisk)
						}
					}

					if evidence.Reasoning != "" {
						md += fmt.Sprintf("   - **Reasoning:** %s\n", evidence.Reasoning)
					}
					if len(evidence.Keywords) > 0 {
						md += fmt.Sprintf("   - **Keywords:** %v\n", evidence.Keywords)
					}
					md += "\n"
				}
			}

			if len(ctrl.Findings) > 0 {
				md += "#### Findings\n\n"
				for i, finding := range ctrl.Findings {
					md += fmt.Sprintf("%d. **[%s]** %s\n", i+1, finding.Severity, finding.Title)
					md += fmt.Sprintf("   - **Description:** %s\n", finding.Description)
					md += fmt.Sprintf("   - **Status:** %s\n", finding.Status)
					if finding.AssignedTo != "" {
						md += fmt.Sprintf("   - **Assigned To:** %s\n", finding.AssignedTo)
					}
					md += "\n"
				}
			}
		}
	}

	return md
}

// escapeCSV escapes special characters in CSV fields
func escapeCSV(s string) string {
	// If the field contains comma, quote, or newline, wrap it in quotes
	// and escape any internal quotes by doubling them
	needsQuotes := false
	for _, c := range s {
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			needsQuotes = true
			break
		}
	}

	if !needsQuotes {
		return s
	}

	// Escape internal quotes
	escaped := ""
	for _, c := range s {
		if c == '"' {
			escaped += "\"\""
		} else {
			escaped += string(c)
		}
	}

	return "\"" + escaped + "\""
}
