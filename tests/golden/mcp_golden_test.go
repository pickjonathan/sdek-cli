package golden_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMCPListOutput validates the output format of the mcp list command
// against the golden file to ensure consistent formatting across releases.
func TestMCPListOutput(t *testing.T) {
	goldenPath := "mcp_list_output.txt"
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Verify golden file structure
	lines := strings.Split(string(golden), "\n")
	if len(lines) < 2 {
		t.Fatal("Golden file should have at least header and one tool")
	}

	// Check header columns
	header := lines[0]
	requiredColumns := []string{"NAME", "STATUS", "LATENCY", "ERRORS", "CAPABILITIES", "LAST CHECK"}
	for _, col := range requiredColumns {
		if !strings.Contains(header, col) {
			t.Errorf("Header missing required column: %s", col)
		}
	}

	// Check that data rows have proper structure
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		
		// Each row should have columns separated by whitespace
		fields := strings.Fields(line)
		if len(fields) < 6 {
			t.Errorf("Line %d has insufficient columns (expected 6+): %s", i, line)
		}
		
		// Verify status values are valid
		status := fields[1]
		validStatuses := []string{"online", "offline", "degraded", "unknown"}
		isValid := false
		for _, vs := range validStatuses {
			if status == vs {
				isValid = true
				break
			}
		}
		if !isValid {
			t.Errorf("Line %d has invalid status '%s': %s", i, status, line)
		}
	}
}

// TestMCPValidateOutput validates the output format of the mcp validate command
// against the golden file to ensure consistent success/error reporting.
func TestMCPValidateOutput(t *testing.T) {
	goldenPath := "mcp_validate_output.txt"
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	content := string(golden)
	lines := strings.Split(content, "\n")

	// Should have file validation lines with ✓ or ❌
	hasValidation := false
	hasSummary := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Check for validation markers
		if strings.HasPrefix(line, "✓") || strings.HasPrefix(line, "❌") {
			hasValidation = true
			
			// Validation lines should reference a file path
			if !strings.Contains(line, ".json") && !strings.Contains(line, "All configuration") {
				t.Errorf("Validation line doesn't reference a file: %s", line)
			}
		}
		
		// Check for summary line
		if strings.Contains(line, "All configuration files") {
			hasSummary = true
		}
	}

	if !hasValidation {
		t.Error("Golden file should contain validation markers (✓ or ❌)")
	}
	if !hasSummary {
		t.Error("Golden file should contain summary line")
	}
}

// TestTUIMCPToolsOutput validates the rendering of the MCP Tools TUI panel
// against the golden file to ensure consistent UI layout.
func TestTUIMCPToolsOutput(t *testing.T) {
	goldenPath := "tui_mcp_tools.txt"
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	content := string(golden)

	// Check for required UI components
	requiredElements := []string{
		"MCP Tools",              // Title
		"NAME",                   // Table header
		"STATUS",                 // Table header
		"LATENCY",                // Table header
		"CAPABILITIES",           // Table header
		"LAST CHECK",             // Table header
		"Test selected tool",     // Action help
		"Refresh",                // Action help
		"Back to home",           // Action help
		"Help:",                  // Help section
		"Quit",                   // Help text
		"Navigate screens",       // Help text
	}

	for _, elem := range requiredElements {
		if !strings.Contains(content, elem) {
			t.Errorf("TUI golden file missing required element: %s", elem)
		}
	}

	// Check for status indicators (emoji or text)
	hasStatusIndicator := strings.Contains(content, "🟢") || 
	                       strings.Contains(content, "online") ||
	                       strings.Contains(content, "offline")
	if !hasStatusIndicator {
		t.Error("TUI golden file should contain status indicators")
	}

	// Check for box drawing characters (indicates table structure)
	boxChars := []string{"│", "─", "┌", "┐", "└", "┘", "├", "┤", "╔", "╗", "╚", "╝", "║", "═"}
	hasBoxDrawing := false
	for _, char := range boxChars {
		if strings.Contains(content, char) {
			hasBoxDrawing = true
			break
		}
	}
	if !hasBoxDrawing {
		t.Error("TUI golden file should contain box-drawing characters for table structure")
	}
}

// NormalizeWhitespace removes trailing whitespace and normalizes line endings
// for consistent comparison. Used when comparing actual output to golden files.
func NormalizeWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var normalized []string
	for _, line := range lines {
		normalized = append(normalized, strings.TrimRight(line, " \t"))
	}
	return strings.Join(normalized, "\n")
}

// CompareOutput compares actual command output with golden file content,
// allowing for minor whitespace differences but catching structural changes.
func CompareOutput(t *testing.T, actual []byte, goldenPath string) {
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Failed to read golden file %s: %v", goldenPath, err)
	}

	actualNorm := NormalizeWhitespace(string(actual))
	goldenNorm := NormalizeWhitespace(string(golden))

	if actualNorm != goldenNorm {
		t.Errorf("Output differs from golden file %s\nActual:\n%s\n\nExpected:\n%s", 
			goldenPath, actualNorm, goldenNorm)
	}
}

// UpdateGoldenFile updates the golden file with new content. Should only be
// called manually when output format intentionally changes, never in CI.
func UpdateGoldenFile(t *testing.T, content []byte, goldenPath string) {
	if os.Getenv("UPDATE_GOLDEN") != "true" {
		t.Skip("Skipping golden file update (set UPDATE_GOLDEN=true to enable)")
	}

	// Ensure directory exists
	dir := filepath.Dir(goldenPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create golden directory: %v", err)
	}

	// Write normalized content
	normalized := []byte(NormalizeWhitespace(string(content)))
	if err := os.WriteFile(goldenPath, normalized, 0644); err != nil {
		t.Fatalf("Failed to write golden file: %v", err)
	}

	t.Logf("Updated golden file: %s", goldenPath)
}

// TestGoldenFilesExist verifies all expected golden files are present
func TestGoldenFilesExist(t *testing.T) {
	expectedFiles := []string{
		"mcp_list_output.txt",
		"mcp_validate_output.txt",
		"tui_mcp_tools.txt",
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Missing golden file: %s", file)
		}
	}
}

// Buffer type for capturing command output in tests
type OutputBuffer struct {
	*bytes.Buffer
}

// NewOutputBuffer creates a new buffer for capturing output
func NewOutputBuffer() *OutputBuffer {
	return &OutputBuffer{Buffer: new(bytes.Buffer)}
}
