package unit

import (
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T005: Contract test for ContextPreamble builder
// These tests define the contract for creating valid ContextPreamble objects
// EXPECTED: These tests MUST FAIL until the builder is implemented in Phase 3.3

func TestNewContextPreamble_ValidInput(t *testing.T) {
	// Arrange
	framework := "SOC2"
	version := "2017"
	section := "CC6.1"
	excerpt := "Access controls shall be implemented to ensure that only authorized individuals can access sensitive data. This includes implementing role-based access controls, multi-factor authentication, and regular access reviews."
	controlIDs := []string{"CC6.1", "CC6.2"}

	// Act
	preamble, err := types.NewContextPreamble(
		framework,
		version,
		section,
		excerpt,
		controlIDs,
	)

	// Assert
	require.NoError(t, err, "NewContextPreamble should succeed with valid input")
	assert.Equal(t, framework, preamble.Framework)
	assert.Equal(t, version, preamble.Version)
	assert.Equal(t, section, preamble.Section)
	assert.Equal(t, excerpt, preamble.Excerpt)
	assert.Equal(t, controlIDs, preamble.ControlIDs)

	// Verify defaults are set
	assert.Equal(t, 0.6, preamble.Rubrics.ConfidenceThreshold, "Default confidence threshold should be 0.6")
	assert.Equal(t, 3, preamble.Rubrics.RequiredCitations, "Default required citations should be 3")
	assert.Equal(t, []string{"low", "medium", "high"}, preamble.Rubrics.RiskLevels)

	// Verify timestamp is set
	assert.False(t, preamble.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.WithinDuration(t, time.Now(), preamble.CreatedAt, 5*time.Second)
}

func TestNewContextPreamble_WithCustomRubrics(t *testing.T) {
	// Arrange
	framework := "ISO27001"
	version := "2013"
	section := "A.9.4.2"
	excerpt := "Secure log-on procedures shall control access to information systems. Requirements include unique user IDs, password complexity, session timeouts, and login attempt restrictions."
	customRubrics := types.AnalysisRubrics{
		ConfidenceThreshold: 0.8,
		RiskLevels:          []string{"low", "medium", "high", "critical"},
		RequiredCitations:   5,
	}

	// Act
	preamble, err := types.NewContextPreambleWithRubrics(
		framework,
		version,
		section,
		excerpt,
		nil, // No control IDs
		customRubrics,
	)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0.8, preamble.Rubrics.ConfidenceThreshold)
	assert.Equal(t, 5, preamble.Rubrics.RequiredCitations)
	assert.Equal(t, []string{"low", "medium", "high", "critical"}, preamble.Rubrics.RiskLevels)
}

func TestNewContextPreamble_ErrorEmptyFramework(t *testing.T) {
	// Arrange
	framework := "" // Invalid: empty
	version := "2017"
	section := "CC6.1"
	excerpt := "Access controls shall be implemented to ensure that only authorized individuals can access sensitive data. This includes implementing role-based access controls, multi-factor authentication, and regular access reviews."

	// Act
	_, err := types.NewContextPreamble(framework, version, section, excerpt, nil)

	// Assert
	require.Error(t, err, "Should return error for empty framework")
	assert.Contains(t, err.Error(), "framework")
	assert.Contains(t, err.Error(), "empty")
}

func TestNewContextPreamble_ErrorEmptyVersion(t *testing.T) {
	// Arrange
	framework := "SOC2"
	version := "" // Invalid: empty
	section := "CC6.1"
	excerpt := "Access controls shall be implemented to ensure that only authorized individuals can access sensitive data. This includes implementing role-based access controls, multi-factor authentication, and regular access reviews."

	// Act
	_, err := types.NewContextPreamble(framework, version, section, excerpt, nil)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version")
	assert.Contains(t, err.Error(), "empty")
}

func TestNewContextPreamble_ErrorEmptySection(t *testing.T) {
	// Arrange
	framework := "SOC2"
	version := "2017"
	section := "" // Invalid: empty
	excerpt := "Access controls shall be implemented to ensure that only authorized individuals can access sensitive data. This includes implementing role-based access controls, multi-factor authentication, and regular access reviews."

	// Act
	_, err := types.NewContextPreamble(framework, version, section, excerpt, nil)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "section")
	assert.Contains(t, err.Error(), "empty")
}

func TestNewContextPreamble_ErrorEmptyExcerpt(t *testing.T) {
	// Arrange
	framework := "SOC2"
	version := "2017"
	section := "CC6.1"
	excerpt := "" // Invalid: empty

	// Act
	_, err := types.NewContextPreamble(framework, version, section, excerpt, nil)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "excerpt")
	assert.Contains(t, err.Error(), "empty")
}

func TestNewContextPreamble_ErrorExcerptTooShort(t *testing.T) {
	// Arrange
	framework := "SOC2"
	version := "2017"
	section := "CC6.1"
	excerpt := "Too short" // Invalid: < 50 characters

	// Act
	_, err := types.NewContextPreamble(framework, version, section, excerpt, nil)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "excerpt")
	assert.Contains(t, err.Error(), "50 characters")
}

func TestNewContextPreamble_ErrorExcerptTooLong(t *testing.T) {
	// Arrange
	framework := "SOC2"
	version := "2017"
	section := "CC6.1"
	// Generate excerpt > 10,000 characters
	excerpt := ""
	for i := 0; i < 11000; i++ {
		excerpt += "x"
	}

	// Act
	_, err := types.NewContextPreamble(framework, version, section, excerpt, nil)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "excerpt")
	assert.Contains(t, err.Error(), "10000 characters")
}

func TestNewContextPreamble_ErrorInvalidControlID(t *testing.T) {
	// Arrange
	framework := "SOC2"
	version := "2017"
	section := "CC6.1"
	excerpt := "Access controls shall be implemented to ensure that only authorized individuals can access sensitive data. This includes implementing role-based access controls, multi-factor authentication, and regular access reviews."
	controlIDs := []string{"CC6.1", "invalid id!"} // Invalid: contains space and exclamation

	// Act
	_, err := types.NewContextPreamble(framework, version, section, excerpt, controlIDs)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "control_id")
	assert.Contains(t, err.Error(), "invalid id!")
}

func TestNewContextPreamble_ErrorInvalidConfidenceThreshold(t *testing.T) {
	// Arrange
	framework := "SOC2"
	version := "2017"
	section := "CC6.1"
	excerpt := "Access controls shall be implemented to ensure that only authorized individuals can access sensitive data. This includes implementing role-based access controls, multi-factor authentication, and regular access reviews."
	invalidRubrics := types.AnalysisRubrics{
		ConfidenceThreshold: 1.5, // Invalid: > 1.0
		RiskLevels:          []string{"low", "medium", "high"},
		RequiredCitations:   3,
	}

	// Act
	_, err := types.NewContextPreambleWithRubrics(
		framework,
		version,
		section,
		excerpt,
		nil,
		invalidRubrics,
	)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "confidence_threshold")
	assert.Contains(t, err.Error(), "0.0")
	assert.Contains(t, err.Error(), "1.0")
}

func TestNewContextPreamble_ErrorNegativeConfidenceThreshold(t *testing.T) {
	// Arrange
	framework := "SOC2"
	version := "2017"
	section := "CC6.1"
	excerpt := "Access controls shall be implemented to ensure that only authorized individuals can access sensitive data. This includes implementing role-based access controls, multi-factor authentication, and regular access reviews."
	invalidRubrics := types.AnalysisRubrics{
		ConfidenceThreshold: -0.1, // Invalid: < 0.0
		RiskLevels:          []string{"low", "medium", "high"},
		RequiredCitations:   3,
	}

	// Act
	_, err := types.NewContextPreambleWithRubrics(
		framework,
		version,
		section,
		excerpt,
		nil,
		invalidRubrics,
	)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "confidence_threshold")
	assert.Contains(t, err.Error(), "0.0")
	assert.Contains(t, err.Error(), "1.0")
}

func TestNewContextPreamble_MinimumValidExcerpt(t *testing.T) {
	// Arrange
	framework := "SOC2"
	version := "2017"
	section := "CC6.1"
	// Exactly 50 characters (minimum valid)
	excerpt := "Access controls must be properly configured always"

	// Act
	preamble, err := types.NewContextPreamble(framework, version, section, excerpt, nil)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, excerpt, preamble.Excerpt)
	assert.Equal(t, 50, len(excerpt), "Test data should be exactly 50 chars")
}

func TestNewContextPreamble_MaximumValidExcerpt(t *testing.T) {
	// Arrange
	framework := "SOC2"
	version := "2017"
	section := "CC6.1"
	// Exactly 10,000 characters (maximum valid)
	excerpt := ""
	for i := 0; i < 10000; i++ {
		excerpt += "x"
	}

	// Act
	preamble, err := types.NewContextPreamble(framework, version, section, excerpt, nil)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 10000, len(preamble.Excerpt))
}

func TestValidateContextPreamble_ValidPreamble(t *testing.T) {
	// Arrange
	preamble := &types.ContextPreamble{
		Framework:  "SOC2",
		Version:    "2017",
		Section:    "CC6.1",
		Excerpt:    "Access controls shall be implemented to ensure that only authorized individuals can access sensitive data. This includes implementing role-based access controls, multi-factor authentication, and regular access reviews.",
		ControlIDs: []string{"CC6.1", "CC6.2"},
		Rubrics: types.AnalysisRubrics{
			ConfidenceThreshold: 0.6,
			RiskLevels:          []string{"low", "medium", "high"},
			RequiredCitations:   3,
		},
		CreatedAt: time.Now(),
	}

	// Act
	err := preamble.Validate()

	// Assert
	require.NoError(t, err, "Valid preamble should pass validation")
}

func TestValidateContextPreamble_InvalidPreamble(t *testing.T) {
	// Arrange
	preamble := &types.ContextPreamble{
		Framework:  "", // Invalid: empty
		Version:    "2017",
		Section:    "CC6.1",
		Excerpt:    "Too short",
		ControlIDs: []string{"CC6.1", "invalid id!"},
		Rubrics: types.AnalysisRubrics{
			ConfidenceThreshold: 1.5, // Invalid: > 1.0
			RiskLevels:          []string{"low", "medium", "high"},
			RequiredCitations:   3,
		},
		CreatedAt: time.Now(),
	}

	// Act
	err := preamble.Validate()

	// Assert
	require.Error(t, err, "Invalid preamble should fail validation")
}
