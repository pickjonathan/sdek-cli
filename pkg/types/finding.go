package types

import (
	"fmt"
	"time"
)

// Finding represents a compliance gap or issue
type Finding struct {
	ID          string    `json:"id"`
	ControlID   string    `json:"control_id"`
	FrameworkID string    `json:"framework_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AssignedTo  string    `json:"assigned_to"`
}

// Severity constants
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// Status constants
const (
	StatusOpen       = "open"
	StatusInProgress = "in_progress"
	StatusResolved   = "resolved"
)

// ValidateFinding checks if a Finding meets all validation rules
func ValidateFinding(f *Finding) error {
	if f == nil {
		return fmt.Errorf("finding cannot be nil")
	}

	// Validate IDs
	if f.ID == "" {
		return fmt.Errorf("finding ID cannot be empty")
	}
	if f.ControlID == "" {
		return fmt.Errorf("control ID cannot be empty")
	}
	if f.FrameworkID == "" {
		return fmt.Errorf("framework ID cannot be empty")
	}

	// Validate severity
	validSeverities := []string{SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical}
	valid := false
	for _, s := range validSeverities {
		if f.Severity == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid severity: %s, must be one of %v", f.Severity, validSeverities)
	}

	// Validate status
	validStatuses := []string{StatusOpen, StatusInProgress, StatusResolved}
	valid = false
	for _, s := range validStatuses {
		if f.Status == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid status: %s, must be one of %v", f.Status, validStatuses)
	}

	// Validate title
	if f.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	return nil
}

// NewFinding creates a new Finding with default values
func NewFinding(id, controlID, frameworkID, title, severity string) *Finding {
	return &Finding{
		ID:          id,
		ControlID:   controlID,
		FrameworkID: frameworkID,
		Title:       title,
		Description: "",
		Severity:    severity,
		Status:      StatusOpen,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		AssignedTo:  "",
	}
}

// UpdateStatus updates the finding status and UpdatedAt timestamp
func (f *Finding) UpdateStatus(status string) error {
	validStatuses := []string{StatusOpen, StatusInProgress, StatusResolved}
	valid := false
	for _, s := range validStatuses {
		if status == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid status: %s", status)
	}

	f.Status = status
	f.UpdatedAt = time.Now()
	return nil
}
