package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMCPEndToEnd validates the complete MCP workflow from configuration to invocation.
// This test covers all quickstart scenarios (AC-01 through AC-06) from the spec.
func TestMCPEndToEnd(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// AC-01: Configuration File Discovery
	t.Run("AC01_ConfigDiscovery", func(t *testing.T) {
		// Verify example configs exist
		examples := []string{
			"docs/examples/mcp/github.json",
			"docs/examples/mcp/slack.json",
			"docs/examples/mcp/jira.json",
		}

		for _, example := range examples {
			if _, err := os.Stat(example); os.IsNotExist(err) {
				t.Errorf("Example config not found: %s", example)
			} else {
				t.Logf("✓ Found example config: %s", example)
			}
		}

		// Verify example README exists
		readmePath := "docs/examples/mcp/README.md"
		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			t.Errorf("Example README not found: %s", readmePath)
		} else {
			t.Logf("✓ Found example README: %s", readmePath)
		}
	})

	// AC-02: JSON Schema Validation
	t.Run("AC02_SchemaValidation", func(t *testing.T) {
		// Test sdek mcp validate command exists
		cmd := exec.Command("./sdek", "mcp", "validate", "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("mcp validate command not available: %v\nOutput: %s", err, output)
		}

		if !strings.Contains(string(output), "validate") {
			t.Error("mcp validate help output unexpected")
		} else {
			t.Log("✓ mcp validate command available")
		}

		// Validate example configs
		examples := []string{
			"docs/examples/mcp/github.json",
			"docs/examples/mcp/slack.json",
			"docs/examples/mcp/jira.json",
		}

		for _, example := range examples {
			cmd = exec.Command("./sdek", "mcp", "validate", example)
			output, err = cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Validation failed for %s: %v\nOutput: %s", example, err, output)
			} else {
				t.Logf("✓ Validated %s", example)
			}
		}
	})

	// AC-03: CLI Command Availability
	t.Run("AC03_CLICommands", func(t *testing.T) {
		// Verify all MCP CLI commands exist
		commands := []string{
			"list",
			"validate",
			"test",
			"enable",
			"disable",
		}

		for _, cmd := range commands {
			cmdArgs := []string{"mcp", cmd, "--help"}
			output, err := exec.Command("./sdek", cmdArgs...).CombinedOutput()
			if err != nil {
				t.Errorf("mcp %s command not available: %v\nOutput: %s", cmd, err, output)
			} else {
				t.Logf("✓ mcp %s command available", cmd)
			}
		}
	})

	// AC-04: TUI Panel Structure
	t.Run("AC04_TUIPanel", func(t *testing.T) {
		// Verify TUI components exist
		requiredFiles := []string{
			"ui/models/mcp_tools.go",
			"ui/app.go",
		}

		for _, file := range requiredFiles {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("TUI component not found: %s", file)
			} else {
				t.Logf("✓ Found TUI component: %s", file)
			}
		}

		// Verify golden file for TUI rendering
		goldenFile := "tests/golden/tui_mcp_tools.txt"
		if _, err := os.Stat(goldenFile); os.IsNotExist(err) {
			t.Errorf("TUI golden file not found: %s", goldenFile)
		} else {
			t.Logf("✓ Found TUI golden file: %s", goldenFile)
		}
	})

	// AC-05: Documentation Completeness
	t.Run("AC05_Documentation", func(t *testing.T) {
		// Verify MCP documentation exists
		docs := []string{
			"docs/commands.md",
			"README.md",
			"docs/examples/mcp/README.md",
		}

		for _, doc := range docs {
			content, err := os.ReadFile(doc)
			if err != nil {
				t.Errorf("Failed to read documentation: %s: %v", doc, err)
				continue
			}

			// Check for MCP-related content
			if !strings.Contains(string(content), "MCP") && !strings.Contains(string(content), "mcp") {
				t.Errorf("Documentation %s doesn't mention MCP", doc)
			} else {
				t.Logf("✓ Documentation %s contains MCP content", doc)
			}
		}
	})

	// AC-06: Golden File Tests
	t.Run("AC06_GoldenTests", func(t *testing.T) {
		// Verify golden test file exists and passes
		goldenTestFile := "tests/golden/mcp_golden_test.go"
		if _, err := os.Stat(goldenTestFile); os.IsNotExist(err) {
			t.Fatalf("Golden test file not found: %s", goldenTestFile)
		}

		// Run golden tests
		cmd := exec.Command("go", "test", "-v", "./tests/golden/...")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Golden tests failed: %v\nOutput: %s", err, output)
		} else {
			t.Log("✓ All golden tests passing")
			// Count passing tests in output
			passCount := strings.Count(string(output), "PASS:")
			t.Logf("  %d golden tests passed", passCount)
		}
	})
}
