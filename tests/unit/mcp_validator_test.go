package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
)

func TestValidatorAcceptsValidConfig(t *testing.T) {
	// Create temp dir for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "valid.json")

	validConfig := `{
  "name": "github",
  "command": "/usr/local/bin/mcp-github",
  "args": ["--verbose"],
  "env": {
    "GITHUB_TOKEN": "${GITHUB_TOKEN}"
  },
  "transport": "stdio",
  "capabilities": ["read", "commits.list", "pr.list"],
  "timeout": "30s",
  "schemaVersion": "1.0.0"
}`

	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	validator := mcp.NewValidator()
	errors := validator.Validate(configPath)

	if len(errors) > 0 {
		t.Errorf("expected no validation errors for valid config, got: %v", errors)
	}
}

func TestValidatorRejectsMissingRequiredFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "missing_command.json")

	invalidConfig := `{
  "name": "invalid",
  "transport": "stdio",
  "capabilities": ["read"],
  "schemaVersion": "1.0.0"
}`

	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	validator := mcp.NewValidator()
	errors := validator.Validate(configPath)

	if len(errors) == 0 {
		t.Error("expected validation errors for config missing 'command' field")
	}

	// Check that error includes file path, line, and property
	foundCommandError := false
	for _, err := range errors {
		if err.JSONPath == "/command" {
			foundCommandError = true
			if err.FilePath == "" || err.Line == 0 {
				t.Errorf("error missing file/line context: %+v", err)
			}
		}
	}

	if !foundCommandError {
		t.Error("expected error about missing 'command' property")
	}
}

func TestValidatorRejectsInvalidTransport(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid_transport.json")

	invalidConfig := `{
  "name": "test",
  "command": "/bin/test",
  "transport": "grpc",
  "capabilities": ["read"],
  "schemaVersion": "1.0.0"
}`

	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	validator := mcp.NewValidator()
	errors := validator.Validate(configPath)

	if len(errors) == 0 {
		t.Error("expected validation errors for invalid transport type 'grpc'")
	}
}

func TestValidatorRejectsInvalidCapabilityFormat(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid_capability.json")

	invalidConfig := `{
  "name": "test",
  "command": "/bin/test",
  "transport": "stdio",
  "capabilities": ["Invalid-Capability!"],
  "schemaVersion": "1.0.0"
}`

	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	validator := mcp.NewValidator()
	errors := validator.Validate(configPath)

	if len(errors) == 0 {
		t.Error("expected validation errors for invalid capability format")
	}
}
