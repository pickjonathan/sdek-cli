package analyze

import (
	"testing"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// TestNewRiskScorer verifies scorer initialization
func TestNewRiskScorer(t *testing.T) {
	scorer := NewRiskScorer()
	if scorer == nil {
		t.Fatal("Scorer should not be nil")
	}
}

// TestCalculateControlRisk verifies risk calculation
func TestCalculateControlRisk(t *testing.T) {
	scorer := NewRiskScorer()
	
	tests := []struct {
		name          string
		findings      []types.Finding
		evidenceCount int
		expected      string
	}{
		{
			name:          "No evidence - highest risk",
			findings:      []types.Finding{},
			evidenceCount: 0,
			expected:      types.RiskStatusRed,
		},
		{
			name: "Critical finding present",
			findings: []types.Finding{
				{Severity: types.SeverityCritical},
			},
			evidenceCount: 5,
			expected:      types.RiskStatusRed,
		},
		{
			name: "3 high findings = 1 critical",
			findings: []types.Finding{
				{Severity: types.SeverityHigh},
				{Severity: types.SeverityHigh},
				{Severity: types.SeverityHigh},
			},
			evidenceCount: 5,
			expected:      types.RiskStatusRed,
		},
		{
			name: "2 high findings = moderate risk",
			findings: []types.Finding{
				{Severity: types.SeverityHigh},
				{Severity: types.SeverityHigh},
			},
			evidenceCount: 5,
			expected:      types.RiskStatusYellow,
		},
		{
			name: "6 medium findings = 1 critical",
			findings: []types.Finding{
				{Severity: types.SeverityMedium},
				{Severity: types.SeverityMedium},
				{Severity: types.SeverityMedium},
				{Severity: types.SeverityMedium},
				{Severity: types.SeverityMedium},
				{Severity: types.SeverityMedium},
			},
			evidenceCount: 5,
			expected:      types.RiskStatusRed,
		},
		{
			name: "12 low findings = 1 critical",
			findings: []types.Finding{
				{Severity: types.SeverityLow},
				{Severity: types.SeverityLow},
				{Severity: types.SeverityLow},
				{Severity: types.SeverityLow},
				{Severity: types.SeverityLow},
				{Severity: types.SeverityLow},
				{Severity: types.SeverityLow},
				{Severity: types.SeverityLow},
				{Severity: types.SeverityLow},
				{Severity: types.SeverityLow},
				{Severity: types.SeverityLow},
				{Severity: types.SeverityLow},
			},
			evidenceCount: 5,
			expected:      types.RiskStatusRed,
		},
		{
			name:          "Sufficient evidence, no findings",
			findings:      []types.Finding{},
			evidenceCount: 5,
			expected:      types.RiskStatusGreen,
		},
		{
			name: "Few low findings with evidence",
			findings: []types.Finding{
				{Severity: types.SeverityLow},
				{Severity: types.SeverityLow},
			},
			evidenceCount: 5,
			expected:      types.RiskStatusGreen,
		},
		{
			name:          "Insufficient evidence (< 3)",
			findings:      []types.Finding{},
			evidenceCount: 2,
			expected:      types.RiskStatusYellow,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risk := scorer.CalculateControlRisk(tt.findings, tt.evidenceCount)
			if risk != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, risk)
			}
		})
	}
}

// TestCalculateCriticalEquivalents verifies the conversion formula
func TestCalculateCriticalEquivalents(t *testing.T) {
	scorer := NewRiskScorer()
	
	tests := []struct {
		name     string
		critical int
		high     int
		medium   int
		low      int
		expected float64
	}{
		{
			name:     "1 critical = 1.0",
			critical: 1,
			high:     0,
			medium:   0,
			low:      0,
			expected: 1.0,
		},
		{
			name:     "3 high = 1.0",
			critical: 0,
			high:     3,
			medium:   0,
			low:      0,
			expected: 1.0,
		},
		{
			name:     "6 medium = 1.0",
			critical: 0,
			high:     0,
			medium:   6,
			low:      0,
			expected: 1.0,
		},
		{
			name:     "12 low = 1.0",
			critical: 0,
			high:     0,
			medium:   0,
			low:      12,
			expected: 1.0,
		},
		{
			name:     "Mixed findings",
			critical: 1,
			high:     3,
			medium:   6,
			low:      12,
			expected: 4.0, // 1 + 1 + 1 + 1
		},
		{
			name:     "Partial equivalents",
			critical: 0,
			high:     1,
			medium:   3,
			low:      6,
			expected: 1.33333, // 0.333 + 0.5 + 0.5
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scorer.calculateCriticalEquivalents(tt.critical, tt.high, tt.medium, tt.low)
			
			// Allow small floating point differences
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("Expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

// TestGenerateFindingsForControl verifies finding generation
func TestGenerateFindingsForControl(t *testing.T) {
	scorer := NewRiskScorer()
	
	tests := []struct {
		name             string
		control          types.Control
		evidence         []types.Evidence
		expectedFindings int
	}{
		{
			name: "Red control with no evidence",
			control: types.Control{
				ID:          "TEST-1",
				FrameworkID: string(types.FrameworkSOC2),
				RiskStatus:  types.RiskStatusRed,
			},
			evidence:         []types.Evidence{},
			expectedFindings: 1, // Insufficient evidence finding
		},
		{
			name: "Red control with low confidence evidence",
			control: types.Control{
				ID:          "TEST-2",
				FrameworkID: string(types.FrameworkSOC2),
				RiskStatus:  types.RiskStatusRed,
			},
			evidence: []types.Evidence{
				{ConfidenceLevel: types.ConfidenceLevelLow},
			},
			expectedFindings: 1, // Low confidence finding
		},
		{
			name: "Yellow control",
			control: types.Control{
				ID:          "TEST-3",
				FrameworkID: string(types.FrameworkSOC2),
				RiskStatus:  types.RiskStatusYellow,
			},
			evidence:         []types.Evidence{},
			expectedFindings: 1, // Moderate risk finding
		},
		{
			name: "Green control",
			control: types.Control{
				ID:          "TEST-4",
				FrameworkID: string(types.FrameworkSOC2),
				RiskStatus:  types.RiskStatusGreen,
			},
			evidence:         []types.Evidence{},
			expectedFindings: 0, // No findings
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := scorer.GenerateFindingsForControl(tt.control, tt.evidence)
			
			if len(findings) != tt.expectedFindings {
				t.Errorf("Expected %d findings, got %d", tt.expectedFindings, len(findings))
			}
			
			// Verify finding structure
			for _, finding := range findings {
				if finding.ID == "" {
					t.Error("Finding ID should not be empty")
				}
				if finding.ControlID != tt.control.ID {
					t.Errorf("Finding control ID mismatch: expected %s, got %s", 
						tt.control.ID, finding.ControlID)
				}
				if finding.FrameworkID != tt.control.FrameworkID {
					t.Errorf("Finding framework ID mismatch")
				}
				if finding.Status != types.StatusOpen {
					t.Errorf("Finding status should be open, got %s", finding.Status)
				}
			}
		})
	}
}

// TestCalculateOverallCompliance verifies compliance percentage
func TestCalculateOverallCompliance(t *testing.T) {
	scorer := NewRiskScorer()
	
	tests := []struct {
		name     string
		controls []types.Control
		expected float64
	}{
		{
			name:     "No controls",
			controls: []types.Control{},
			expected: 0.0,
		},
		{
			name: "All green",
			controls: []types.Control{
				{RiskStatus: types.RiskStatusGreen},
				{RiskStatus: types.RiskStatusGreen},
				{RiskStatus: types.RiskStatusGreen},
				{RiskStatus: types.RiskStatusGreen},
			},
			expected: 100.0,
		},
		{
			name: "Half green",
			controls: []types.Control{
				{RiskStatus: types.RiskStatusGreen},
				{RiskStatus: types.RiskStatusGreen},
				{RiskStatus: types.RiskStatusRed},
				{RiskStatus: types.RiskStatusYellow},
			},
			expected: 50.0,
		},
		{
			name: "One third green",
			controls: []types.Control{
				{RiskStatus: types.RiskStatusGreen},
				{RiskStatus: types.RiskStatusYellow},
				{RiskStatus: types.RiskStatusRed},
			},
			expected: 33.33,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compliance := scorer.CalculateOverallCompliance(tt.controls)
			
			// Allow small floating point differences
			if compliance < tt.expected-0.1 || compliance > tt.expected+0.1 {
				t.Errorf("Expected %.2f%%, got %.2f%%", tt.expected, compliance)
			}
		})
	}
}

// TestGetRiskSummary verifies risk summary generation
func TestGetRiskSummary(t *testing.T) {
	scorer := NewRiskScorer()
	
	controls := []types.Control{
		{RiskStatus: types.RiskStatusGreen},
		{RiskStatus: types.RiskStatusGreen},
		{RiskStatus: types.RiskStatusGreen},
		{RiskStatus: types.RiskStatusYellow},
		{RiskStatus: types.RiskStatusYellow},
		{RiskStatus: types.RiskStatusRed},
	}
	
	summary := scorer.GetRiskSummary(controls)
	
	if summary.TotalCount != 6 {
		t.Errorf("Expected total count 6, got %d", summary.TotalCount)
	}
	if summary.GreenCount != 3 {
		t.Errorf("Expected green count 3, got %d", summary.GreenCount)
	}
	if summary.YellowCount != 2 {
		t.Errorf("Expected yellow count 2, got %d", summary.YellowCount)
	}
	if summary.RedCount != 1 {
		t.Errorf("Expected red count 1, got %d", summary.RedCount)
	}
	
	expectedCompliance := 50.0 // 3 out of 6 green
	if summary.CompliancePercentage < expectedCompliance-0.1 || summary.CompliancePercentage > expectedCompliance+0.1 {
		t.Errorf("Expected compliance %.2f%%, got %.2f%%", 
			expectedCompliance, summary.CompliancePercentage)
	}
}
