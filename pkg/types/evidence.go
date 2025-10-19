package types

import (
	"fmt"
	"time"
)

// Evidence represents a mapping between an Event and a Control
type Evidence struct {
	ID              string    `json:"id"`
	EventID         string    `json:"event_id"`
	ControlID       string    `json:"control_id"`
	FrameworkID     string    `json:"framework_id"`
	ConfidenceLevel string    `json:"confidence_level"`
	ConfidenceScore float64   `json:"confidence_score"`
	MappedAt        time.Time `json:"mapped_at"`
	Keywords        []string  `json:"keywords"`
	Reasoning       string    `json:"reasoning"`

	// AI analysis metadata (Feature 002: AI Evidence Analysis)
	AIAnalyzed          bool   `json:"ai_analyzed"`          // True if AI was used for this evidence
	AIJustification     string `json:"ai_justification"`     // AI-generated explanation
	AIConfidence        int    `json:"ai_confidence"`        // AI confidence score (0-100)
	AIResidualRisk      string `json:"ai_residual_risk"`     // AI risk notes
	HeuristicConfidence int    `json:"heuristic_confidence"` // Original keyword-based score (0-100)
	CombinedConfidence  int    `json:"combined_confidence"`  // Weighted average (70% AI + 30% heuristic)
	AnalysisMethod      string `json:"analysis_method"`      // "ai+heuristic" | "heuristic-only" | "no-ai"
}

// Confidence level constants
const (
	ConfidenceLevelLow    = "low"
	ConfidenceLevelMedium = "medium"
	ConfidenceLevelHigh   = "high"
)

// Confidence score ranges
const (
	ConfidenceScoreLowMax    = 50.0
	ConfidenceScoreMediumMax = 75.0
	ConfidenceScoreHighMax   = 100.0
)

// ValidateEvidence checks if Evidence meets all validation rules
func ValidateEvidence(e *Evidence) error {
	if e == nil {
		return fmt.Errorf("evidence cannot be nil")
	}

	// Validate IDs are not empty
	if e.ID == "" {
		return fmt.Errorf("evidence ID cannot be empty")
	}
	if e.EventID == "" {
		return fmt.Errorf("event ID cannot be empty")
	}
	if e.ControlID == "" {
		return fmt.Errorf("control ID cannot be empty")
	}
	if e.FrameworkID == "" {
		return fmt.Errorf("framework ID cannot be empty")
	}

	// Validate confidence level
	validLevels := []string{ConfidenceLevelLow, ConfidenceLevelMedium, ConfidenceLevelHigh}
	valid := false
	for _, level := range validLevels {
		if e.ConfidenceLevel == level {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid confidence level: %s, must be one of %v", e.ConfidenceLevel, validLevels)
	}

	// Validate confidence score
	if e.ConfidenceScore < 0 || e.ConfidenceScore > 100 {
		return fmt.Errorf("confidence score must be between 0 and 100, got %f", e.ConfidenceScore)
	}

	return nil
}

// SetConfidenceLevel sets the confidence level based on the confidence score
func (e *Evidence) SetConfidenceLevel() {
	if e.ConfidenceScore <= ConfidenceScoreLowMax {
		e.ConfidenceLevel = ConfidenceLevelLow
	} else if e.ConfidenceScore <= ConfidenceScoreMediumMax {
		e.ConfidenceLevel = ConfidenceLevelMedium
	} else {
		e.ConfidenceLevel = ConfidenceLevelHigh
	}
}

// NewEvidence creates a new Evidence with default values
func NewEvidence(id, eventID, controlID, frameworkID string, score float64) *Evidence {
	e := &Evidence{
		ID:              id,
		EventID:         eventID,
		ControlID:       controlID,
		FrameworkID:     frameworkID,
		ConfidenceScore: score,
		MappedAt:        time.Now(),
		Keywords:        []string{},
		Reasoning:       "",
	}
	e.SetConfidenceLevel()
	return e
}
