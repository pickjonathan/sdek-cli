package types

import (
	"fmt"
	"regexp"
	"time"
)

// ContextPreamble represents framework metadata and control excerpt used to ground AI prompts.
// This is injected into AI prompts to provide domain-specific context for evidence analysis.
type ContextPreamble struct {
	// Framework metadata
	Framework string `json:"framework"` // e.g., "SOC2", "ISO27001"
	Version   string `json:"version"`   // e.g., "2017", "2013"
	Section   string `json:"section"`   // e.g., "CC6.1", "A.9.4.2"

	// Control context
	Excerpt    string   `json:"excerpt"`     // Full control text
	ControlIDs []string `json:"control_ids"` // Related control identifiers

	// Analysis configuration
	Rubrics AnalysisRubrics `json:"rubrics"` // Confidence/risk criteria

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
}

// AnalysisRubrics defines confidence and risk evaluation criteria for AI analysis.
type AnalysisRubrics struct {
	ConfidenceThreshold float64  `json:"confidence_threshold"` // Default: 0.6
	RiskLevels          []string `json:"risk_levels"`          // ["low", "medium", "high"]
	RequiredCitations   int      `json:"required_citations"`   // Min citations for high confidence
}

// Validation constants
const (
	MinExcerptLength = 50
	MaxExcerptLength = 10000
)

var controlIDPattern = regexp.MustCompile(`^[A-Z0-9.-]+$`)

// NewContextPreamble creates a new ContextPreamble with default rubrics.
func NewContextPreamble(
	framework string,
	version string,
	section string,
	excerpt string,
	controlIDs []string,
) (*ContextPreamble, error) {
	// Use default rubrics
	defaultRubrics := AnalysisRubrics{
		ConfidenceThreshold: 0.6,
		RiskLevels:          []string{"low", "medium", "high"},
		RequiredCitations:   3,
	}

	return NewContextPreambleWithRubrics(framework, version, section, excerpt, controlIDs, defaultRubrics)
}

// NewContextPreambleWithRubrics creates a new ContextPreamble with custom rubrics.
func NewContextPreambleWithRubrics(
	framework string,
	version string,
	section string,
	excerpt string,
	controlIDs []string,
	rubrics AnalysisRubrics,
) (*ContextPreamble, error) {
	preamble := &ContextPreamble{
		Framework:  framework,
		Version:    version,
		Section:    section,
		Excerpt:    excerpt,
		ControlIDs: controlIDs,
		Rubrics:    rubrics,
		CreatedAt:  time.Now(),
	}

	// Validate
	if err := preamble.Validate(); err != nil {
		return nil, err
	}

	return preamble, nil
}

// Validate checks if the ContextPreamble meets all validation rules.
func (cp *ContextPreamble) Validate() error {
	// Framework validation
	if cp.Framework == "" {
		return fmt.Errorf("framework cannot be empty")
	}

	// Version validation
	if cp.Version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	// Section validation
	if cp.Section == "" {
		return fmt.Errorf("section cannot be empty")
	}

	// Excerpt validation
	if cp.Excerpt == "" {
		return fmt.Errorf("excerpt cannot be empty")
	}

	excerptLen := len(cp.Excerpt)
	if excerptLen < MinExcerptLength {
		return fmt.Errorf("excerpt must be at least 50 characters, got %d", excerptLen)
	}

	if excerptLen > MaxExcerptLength {
		return fmt.Errorf("excerpt must not exceed 10000 characters, got %d", excerptLen)
	}

	// Control IDs validation
	for _, controlID := range cp.ControlIDs {
		if !controlIDPattern.MatchString(controlID) {
			return fmt.Errorf("invalid control_id: %s (must match pattern ^[A-Z0-9.-]+$)", controlID)
		}
	}

	// Rubrics validation
	if cp.Rubrics.ConfidenceThreshold < 0.0 || cp.Rubrics.ConfidenceThreshold > 1.0 {
		return fmt.Errorf("confidence_threshold must be between 0.0 and 1.0, got %f", cp.Rubrics.ConfidenceThreshold)
	}

	return nil
}
