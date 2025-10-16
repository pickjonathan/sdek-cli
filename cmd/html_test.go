package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHTMLCommand(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a minimal test report JSON
	testReport := `{
  "metadata": {
    "generated_at": "2025-10-16T21:00:00Z",
    "version": "1.0"
  },
  "summary": {
    "total_sources": 2,
    "total_events": 10,
    "total_frameworks": 1,
    "total_controls": 5,
    "total_evidence": 8,
    "total_findings": 2,
    "overall_compliance_percentage": 75.0
  },
  "frameworks": [
    {
      "framework": {
        "id": "soc2",
        "name": "SOC 2",
        "version": "2017",
        "control_count": 5,
        "compliance_percentage": 75.0
      },
      "controls": [
        {
          "control": {
            "id": "CC6.1",
            "framework_id": "soc2",
            "title": "Logical Access Controls",
            "description": "Test control",
            "risk_status": "green",
            "confidence_level": 85.0
          },
          "evidence": [
            {
              "id": "ev-1",
              "control_id": "CC6.1",
              "framework_id": "soc2",
              "event_id": "evt-1",
              "confidence_score": 85.0,
              "reasoning": "Test reasoning",
              "analysis_method": "heuristic",
              "ai_analyzed": false
            }
          ],
          "findings": []
        }
      ]
    }
  ]
}`

	inputPath := filepath.Join(tmpDir, "test-report.json")
	outputPath := filepath.Join(tmpDir, "test-dashboard.html")

	// Write test report
	if err := os.WriteFile(inputPath, []byte(testReport), 0644); err != nil {
		t.Fatalf("Failed to write test report: %v", err)
	}

	// Set flags
	htmlInputFile = inputPath
	htmlOutputFile = outputPath

	// Run command
	if err := runHTML(htmlCmd, []string{}); err != nil {
		t.Fatalf("runHTML failed: %v", err)
	}

	// Verify output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Output file was not created: %s", outputPath)
	}

	// Verify output file contains expected HTML
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Check for key HTML elements
	htmlStr := string(content)
	checks := []string{
		"<!DOCTYPE html>",
		"<title>SDEK Compliance Report</title>",
		"SDEK Compliance Dashboard",
		"reportData = JSON.parse",
		"function renderSummary",
		"function renderFrameworks",
		"function renderFindings",
		"function renderEvidence",
	}

	for _, check := range checks {
		if !contains(htmlStr, check) {
			t.Errorf("Output HTML missing expected content: %s", check)
		}
	}

	// Verify file size is reasonable (should be > 10KB for a full HTML dashboard)
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if info.Size() < 10000 {
		t.Errorf("Output file too small (%d bytes), expected > 10KB", info.Size())
	}
}

func TestHTMLCommandMissingInput(t *testing.T) {
	tmpDir := t.TempDir()

	// Set flags to non-existent file
	htmlInputFile = filepath.Join(tmpDir, "nonexistent.json")
	htmlOutputFile = filepath.Join(tmpDir, "output.html")

	// Run command - should fail
	err := runHTML(htmlCmd, []string{})
	if err == nil {
		t.Error("Expected error for missing input file, got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
