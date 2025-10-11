package types

import "fmt"

// Framework represents a compliance standard
type Framework struct {
	ID                   string  `json:"id"`
	Name                 string  `json:"name"`
	Version              string  `json:"version"`
	ControlCount         int     `json:"control_count"`
	CompliancePercentage float64 `json:"compliance_percentage"`
	Description          string  `json:"description"`
	Category             string  `json:"category"`
}

// Framework ID constants
const (
	FrameworkSOC2     = "soc2"
	FrameworkISO27001 = "iso27001"
	FrameworkPCIDSS   = "pci_dss"
)

// ValidFrameworkIDs contains all valid framework identifiers
var ValidFrameworkIDs = []string{
	FrameworkSOC2,
	FrameworkISO27001,
	FrameworkPCIDSS,
}

// ValidateFramework checks if a Framework meets all validation rules
func ValidateFramework(f *Framework) error {
	if f == nil {
		return fmt.Errorf("framework cannot be nil")
	}

	// Validate ID
	valid := false
	for _, id := range ValidFrameworkIDs {
		if f.ID == id {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid framework ID: %s, must be one of %v", f.ID, ValidFrameworkIDs)
	}

	// Validate control count
	if f.ControlCount <= 0 {
		return fmt.Errorf("control count must be positive, got %d", f.ControlCount)
	}

	// Validate compliance percentage
	if f.CompliancePercentage < 0 || f.CompliancePercentage > 100 {
		return fmt.Errorf("compliance percentage must be between 0 and 100, got %f", f.CompliancePercentage)
	}

	return nil
}

// CalculateCompliance calculates compliance percentage from control counts
func (f *Framework) CalculateCompliance(greenControls, totalControls int) {
	if totalControls == 0 {
		f.CompliancePercentage = 0
		return
	}
	f.CompliancePercentage = float64(greenControls) / float64(totalControls) * 100
}

// NewFramework creates a new Framework with default values
func NewFramework(id, name, version string, controlCount int) *Framework {
	return &Framework{
		ID:                   id,
		Name:                 name,
		Version:              version,
		ControlCount:         controlCount,
		CompliancePercentage: 0,
		Description:          "",
		Category:             "security",
	}
}
