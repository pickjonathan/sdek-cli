package types

import "fmt"

// Control represents a specific compliance requirement within a framework
type Control struct {
	ID               string   `json:"id"`
	FrameworkID      string   `json:"framework_id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Category         string   `json:"category"`
	RiskStatus       string   `json:"risk_status"`
	RiskScore        float64  `json:"risk_score"`
	EvidenceCount    int      `json:"evidence_count"`
	ConfidenceLevel  float64  `json:"confidence_level"`
	Keywords         []string `json:"keywords"`
	RequiredEvidence int      `json:"required_evidence"`
	CriticalSeverity int      `json:"critical_severity"`
	HighSeverity     int      `json:"high_severity"`
	MediumSeverity   int      `json:"medium_severity"`
	LowSeverity      int      `json:"low_severity"`
}

// Risk status constants
const (
	RiskStatusGreen  = "green"
	RiskStatusYellow = "yellow"
	RiskStatusRed    = "red"
)

// ValidateControl checks if a Control meets all validation rules
func ValidateControl(c *Control) error {
	if c == nil {
		return fmt.Errorf("control cannot be nil")
	}

	// Validate ID is not empty
	if c.ID == "" {
		return fmt.Errorf("control ID cannot be empty")
	}

	// Validate framework ID is not empty
	if c.FrameworkID == "" {
		return fmt.Errorf("framework ID cannot be empty")
	}

	// Validate risk status
	validStatuses := []string{RiskStatusGreen, RiskStatusYellow, RiskStatusRed}
	valid := false
	for _, status := range validStatuses {
		if c.RiskStatus == status {
			valid = true
			break
		}
	}
	if !valid && c.RiskStatus != "" {
		return fmt.Errorf("invalid risk status: %s, must be one of %v", c.RiskStatus, validStatuses)
	}

	// Validate confidence level
	if c.ConfidenceLevel < 0 || c.ConfidenceLevel > 100 {
		return fmt.Errorf("confidence level must be between 0 and 100, got %f", c.ConfidenceLevel)
	}

	return nil
}

// CalculateRiskScore calculates risk score based on severity-weighted findings
// Formula: 3 high = 1 critical, 6 medium = 1 critical, 12 low = 1 critical
func (c *Control) CalculateRiskScore() {
	criticalEquivalent := float64(c.CriticalSeverity) +
		float64(c.HighSeverity)/3.0 +
		float64(c.MediumSeverity)/6.0 +
		float64(c.LowSeverity)/12.0

	c.RiskScore = criticalEquivalent

	// Determine risk status based on score and evidence
	if c.EvidenceCount >= c.RequiredEvidence && criticalEquivalent == 0 {
		c.RiskStatus = RiskStatusGreen
	} else if criticalEquivalent >= 1 {
		c.RiskStatus = RiskStatusRed
	} else {
		c.RiskStatus = RiskStatusYellow
	}
}

// NewControl creates a new Control with default values
func NewControl(id, frameworkID, title string, keywords []string) *Control {
	return &Control{
		ID:               id,
		FrameworkID:      frameworkID,
		Title:            title,
		Description:      "",
		Category:         "",
		RiskStatus:       "",
		RiskScore:        0,
		EvidenceCount:    0,
		ConfidenceLevel:  0,
		Keywords:         keywords,
		RequiredEvidence: 1,
	}
}
