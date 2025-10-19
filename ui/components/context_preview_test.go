package components

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// TestContextPreviewView tests the rendering of the context preview component
func TestContextPreviewView(b *testing.T) {
	preamble := types.ContextPreamble{
		Framework:  "SOC2",
		Version:    "2017",
		Section:    "CC6.1",
		Excerpt:    "The entity restricts logical and physical access to systems and programs to authorized users. The entity implements controls to prevent or detect and act upon the introduction of unauthorized or malicious software.",
		ControlIDs: []string{"CC6.1", "CC6.2", "CC6.3"},
		Rubrics: types.AnalysisRubrics{
			ConfidenceThreshold: 0.6,
			RiskLevels:          []string{"low", "medium", "high"},
			RequiredCitations:   3,
		},
		CreatedAt: time.Date(2025, 10, 18, 10, 0, 0, 0, time.UTC),
	}

	model := NewContextPreview(preamble, 42)
	model = model.SetSize(80, 24)

	output := model.View()

	// Check that output contains expected elements
	expectedStrings := []string{
		"AI Context Preview",
		"Framework: SOC2 2017",
		"Section:   CC6.1",
		"Evidence:  42 events",
		"Policy Excerpt:",
		"restricts logical and physical access",
		"Related Controls: CC6.1, CC6.2, CC6.3",
		"Press Enter or Y to proceed",
		"Press N or Q to cancel",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			b.Errorf("Output missing expected string: %q", expected)
		}
	}
}

// TestContextPreviewTruncation tests that long excerpts are truncated
func TestContextPreviewTruncation(b *testing.T) {
	// Create a very long excerpt (>500 chars)
	longExcerpt := strings.Repeat("The entity implements comprehensive security controls. ", 20) // ~1000 chars

	preamble := types.ContextPreamble{
		Framework: "ISO27001",
		Version:   "2013",
		Section:   "A.9.4.2",
		Excerpt:   longExcerpt,
		CreatedAt: time.Now(),
	}

	model := NewContextPreview(preamble, 10)
	model = model.SetSize(80, 24)

	output := model.View()

	// Should contain ellipsis for truncation
	if !strings.Contains(output, "...") {
		b.Error("Expected excerpt to be truncated with ellipsis")
	}

	// Should not contain the full original text
	if strings.Contains(output, longExcerpt) {
		b.Error("Excerpt was not truncated as expected")
	}
}

// TestContextPreviewNoRelatedControls tests rendering without related controls
func TestContextPreviewNoRelatedControls(b *testing.T) {
	preamble := types.ContextPreamble{
		Framework:  "PCI-DSS",
		Version:    "4.0",
		Section:    "8.2.1",
		Excerpt:    "Use strong cryptographic techniques to render authentication credentials unreadable.",
		ControlIDs: []string{}, // Empty related controls
		CreatedAt:  time.Now(),
	}

	model := NewContextPreview(preamble, 5)
	model = model.SetSize(80, 24)

	output := model.View()

	// Should NOT contain "Related Controls" section
	if strings.Contains(output, "Related Controls:") {
		b.Error("Should not display 'Related Controls' when list is empty")
	}

	// Should still contain other expected elements
	if !strings.Contains(output, "PCI-DSS 4.0") {
		b.Error("Missing framework information")
	}
}

// TestContextPreviewConfirm tests confirmation action
func TestContextPreviewConfirm(b *testing.T) {
	preamble := types.ContextPreamble{
		Framework: "SOC2",
		Version:   "2017",
		Section:   "CC6.1",
		Excerpt:   "Sample excerpt",
		CreatedAt: time.Now(),
	}

	model := NewContextPreview(preamble, 10)

	// Initially not confirmed
	if model.Confirmed() {
		b.Error("Model should not be confirmed initially")
	}

	// Simulate Enter key press
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	finalModel := updatedModel.(ContextPreviewModel)

	if !finalModel.Confirmed() {
		b.Error("Model should be confirmed after Enter key")
	}
	if finalModel.Cancelled() {
		b.Error("Model should not be cancelled after Enter key")
	}
}

// TestContextPreviewCancel tests cancellation action
func TestContextPreviewCancel(b *testing.T) {
	preamble := types.ContextPreamble{
		Framework: "SOC2",
		Version:   "2017",
		Section:   "CC6.1",
		Excerpt:   "Sample excerpt",
		CreatedAt: time.Now(),
	}

	model := NewContextPreview(preamble, 10)

	// Initially not cancelled
	if model.Cancelled() {
		b.Error("Model should not be cancelled initially")
	}

	// Simulate 'q' key press
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	finalModel := updatedModel.(ContextPreviewModel)

	if !finalModel.Cancelled() {
		b.Error("Model should be cancelled after 'q' key")
	}
	if finalModel.Confirmed() {
		b.Error("Model should not be confirmed after 'q' key")
	}
}

// TestContextPreviewGoldenFile tests against golden file
func TestContextPreviewGoldenFile(b *testing.T) {
	preamble := types.ContextPreamble{
		Framework:  "SOC2",
		Version:    "2017",
		Section:    "CC6.1",
		Excerpt:    "The entity restricts logical and physical access to systems and programs to authorized users. The entity implements controls to prevent or detect and act upon the introduction of unauthorized or malicious software to meet the entity's objectives.",
		ControlIDs: []string{"CC6.1", "CC6.2", "CC6.3"},
		Rubrics: types.AnalysisRubrics{
			ConfidenceThreshold: 0.6,
			RiskLevels:          []string{"low", "medium", "high"},
			RequiredCitations:   3,
		},
		CreatedAt: time.Date(2025, 10, 18, 10, 0, 0, 0, time.UTC),
	}

	model := NewContextPreview(preamble, 42)
	model = model.SetSize(80, 24)

	output := model.View()

	// Path to golden file
	goldenPath := filepath.Join("..", "..", "tests", "golden", "fixtures", "context_preview_soc2.txt")

	// Update golden file if updateGolden is true
	if updateGolden {
		err := os.MkdirAll(filepath.Dir(goldenPath), 0755)
		if err != nil {
			b.Fatalf("Failed to create golden directory: %v", err)
		}
		err = os.WriteFile(goldenPath, []byte(output), 0644)
		if err != nil {
			b.Fatalf("Failed to write golden file: %v", err)
		}
		b.Logf("Updated golden file: %s", goldenPath)
		return
	}

	// Read golden file
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		b.Fatalf("Failed to read golden file: %v (run with -update to create)", err)
	}

	// Compare output with golden file
	if string(expected) != output {
		b.Errorf("Output does not match golden file.\nExpected:\n%s\n\nGot:\n%s", string(expected), output)
		b.Logf("Run with -update to update the golden file")
	}
}

// Flag for updating golden files
var updateGolden = false // Set to true to update golden files

func init() {
	// In a real test, this would be a command-line flag
	// For now, we set it to false after generating the golden file
}
