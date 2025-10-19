package analyze

import "github.com/pickjonathan/sdek-cli/pkg/types"

// RiskScorer calculates risk scores for controls based on findings
type RiskScorer struct{}

// NewRiskScorer creates a new risk scorer
func NewRiskScorer() *RiskScorer {
	return &RiskScorer{}
}

// CalculateControlRisk calculates the risk status for a control
// Formula: 3 High = 1 Critical, 6 Medium = 1 Critical, 12 Low = 1 Critical
// Returns risk status (Green, Yellow, Red) based on severity-weighted findings
func (r *RiskScorer) CalculateControlRisk(findings []types.Finding, evidenceCount int) string {
	// If no evidence, cannot determine risk
	if evidenceCount == 0 {
		return types.RiskStatusRed // No evidence = highest risk
	}

	// Count findings by severity
	var criticalCount, highCount, mediumCount, lowCount int

	for _, finding := range findings {
		switch finding.Severity {
		case types.SeverityCritical:
			criticalCount++
		case types.SeverityHigh:
			highCount++
		case types.SeverityMedium:
			mediumCount++
		case types.SeverityLow:
			lowCount++
		}
	}

	// Calculate critical equivalents using the formula
	criticalEquivalents := r.calculateCriticalEquivalents(criticalCount, highCount, mediumCount, lowCount)

	// Determine risk status
	if criticalEquivalents >= 1.0 {
		return types.RiskStatusRed // Critical issues present
	}

	if criticalEquivalents >= 0.5 || highCount >= 2 {
		return types.RiskStatusYellow // Moderate risk
	}

	// Check if sufficient evidence exists (at least 3 pieces of evidence for green)
	if evidenceCount < 3 {
		return types.RiskStatusYellow // Insufficient evidence
	}

	return types.RiskStatusGreen // Low risk, sufficient evidence
}

// calculateCriticalEquivalents converts findings to critical equivalents
// 3 High = 1 Critical, 6 Medium = 1 Critical, 12 Low = 1 Critical
func (r *RiskScorer) calculateCriticalEquivalents(critical, high, medium, low int) float64 {
	equivalents := float64(critical)
	equivalents += float64(high) / 3.0
	equivalents += float64(medium) / 6.0
	equivalents += float64(low) / 12.0

	return equivalents
}

// GenerateFindingsForControl generates findings for controls based on risk status
func (r *RiskScorer) GenerateFindingsForControl(control types.Control, evidenceList []types.Evidence) []types.Finding {
	var findings []types.Finding

	// If control is red, generate a finding
	if control.RiskStatus == types.RiskStatusRed {
		// Check why it's red
		if len(evidenceList) == 0 {
			// No evidence found
			finding := types.Finding{
				ID:          control.ID + "-insufficient-evidence",
				ControlID:   control.ID,
				FrameworkID: control.FrameworkID,
				Title:       "Insufficient Evidence",
				Description: "No evidence found to support compliance with this control",
				Severity:    types.SeverityHigh,
				Status:      types.StatusOpen,
			}
			findings = append(findings, finding)
		} else {
			// Has evidence but still red (low confidence)
			lowConfidenceCount := 0
			for _, evidence := range evidenceList {
				if evidence.ConfidenceLevel == types.ConfidenceLevelLow {
					lowConfidenceCount++
				}
			}

			if lowConfidenceCount > 0 {
				finding := types.Finding{
					ID:          control.ID + "-low-confidence",
					ControlID:   control.ID,
					FrameworkID: control.FrameworkID,
					Title:       "Low Confidence Evidence",
					Description: "Evidence exists but confidence level is insufficient",
					Severity:    types.SeverityMedium,
					Status:      types.StatusOpen,
				}
				findings = append(findings, finding)
			}
		}
	}

	// If control is yellow, generate a medium finding
	if control.RiskStatus == types.RiskStatusYellow {
		finding := types.Finding{
			ID:          control.ID + "-moderate-risk",
			ControlID:   control.ID,
			FrameworkID: control.FrameworkID,
			Title:       "Moderate Risk",
			Description: "Control has moderate risk - additional evidence or remediation needed",
			Severity:    types.SeverityMedium,
			Status:      types.StatusOpen,
		}
		findings = append(findings, finding)
	}

	return findings
}

// CalculateOverallCompliance calculates compliance percentage for a framework
func (r *RiskScorer) CalculateOverallCompliance(controls []types.Control) float64 {
	if len(controls) == 0 {
		return 0.0
	}

	greenCount := 0
	for _, control := range controls {
		if control.RiskStatus == types.RiskStatusGreen {
			greenCount++
		}
	}

	return (float64(greenCount) / float64(len(controls))) * 100.0
}

// GetRiskSummary generates a risk summary for a set of controls
func (r *RiskScorer) GetRiskSummary(controls []types.Control) RiskSummary {
	summary := RiskSummary{}

	for _, control := range controls {
		switch control.RiskStatus {
		case types.RiskStatusGreen:
			summary.GreenCount++
		case types.RiskStatusYellow:
			summary.YellowCount++
		case types.RiskStatusRed:
			summary.RedCount++
		}
	}

	summary.TotalCount = len(controls)
	summary.CompliancePercentage = r.CalculateOverallCompliance(controls)

	return summary
}

// RiskSummary contains summary statistics about risk across controls
type RiskSummary struct {
	TotalCount           int
	GreenCount           int
	YellowCount          int
	RedCount             int
	CompliancePercentage float64
}
