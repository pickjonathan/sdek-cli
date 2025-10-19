package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
)

func TestLoaderPrecedenceProjectOverridesGlobal(t *testing.T) {
	tmpDir := t.TempDir()
	globalDir := filepath.Join(tmpDir, "global")
	projectDir := filepath.Join(tmpDir, "project")

	os.MkdirAll(filepath.Join(globalDir, ".sdek", "mcp"), 0755)
	os.MkdirAll(filepath.Join(projectDir, ".sdek", "mcp"), 0755)

	// Same tool name in both locations
	globalConfig := `{
  "name": "test-tool",
  "command": "/global/bin/tool",
  "transport": "stdio",
  "capabilities": ["global"],
  "schemaVersion": "1.0.0"
}`
	projectConfig := `{
  "name": "test-tool",
  "command": "/project/bin/tool",
  "transport": "stdio",
  "capabilities": ["project"],
  "schemaVersion": "1.0.0"
}`

	os.WriteFile(filepath.Join(globalDir, ".sdek", "mcp", "test-tool.json"), []byte(globalConfig), 0644)
	os.WriteFile(filepath.Join(projectDir, ".sdek", "mcp", "test-tool.json"), []byte(projectConfig), 0644)

	loader := mcp.NewLoader(globalDir, projectDir)
	configs, err := loader.LoadConfigs()

	if err != nil {
		t.Fatalf("failed to load configs: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("expected 1 config (project should override global), got %d", len(configs))
	}

	if configs[0].Command != "/project/bin/tool" {
		t.Errorf("expected project config to take precedence, got command: %s", configs[0].Command)
	}
}

func TestLoaderEnvVarPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	envDir := filepath.Join(tmpDir, "env")
	projectDir := filepath.Join(tmpDir, "project")

	os.MkdirAll(filepath.Join(envDir, "mcp"), 0755)
	os.MkdirAll(filepath.Join(projectDir, ".sdek", "mcp"), 0755)

	envConfig := `{
  "name": "test-tool",
  "command": "/env/bin/tool",
  "transport": "stdio",
  "capabilities": ["env"],
  "schemaVersion": "1.0.0"
}`
	projectConfig := `{
  "name": "test-tool",
  "command": "/project/bin/tool",
  "transport": "stdio",
  "capabilities": ["project"],
  "schemaVersion": "1.0.0"
}`

	os.WriteFile(filepath.Join(envDir, "mcp", "test-tool.json"), []byte(envConfig), 0644)
	os.WriteFile(filepath.Join(projectDir, ".sdek", "mcp", "test-tool.json"), []byte(projectConfig), 0644)

	// Set env var
	os.Setenv("SDEK_MCP_PATH", envDir)
	defer os.Unsetenv("SDEK_MCP_PATH")

	loader := mcp.NewLoader("", projectDir)
	configs, err := loader.LoadConfigs()

	if err != nil {
		t.Fatalf("failed to load configs: %v", err)
	}

	// With SDEK_MCP_PATH set, expect env config to take precedence
	found := false
	for _, cfg := range configs {
		if cfg.Name == "test-tool" && cfg.Command == "/env/bin/tool" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected SDEK_MCP_PATH config to take precedence")
	}
}

func TestLoaderExpandsEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".sdek", "mcp")
	os.MkdirAll(configDir, 0755)

	config := `{
  "name": "test-tool",
  "command": "/bin/tool",
  "env": {
    "API_TOKEN": "${TEST_TOKEN}",
    "LOG_LEVEL": "info"
  },
  "transport": "stdio",
  "capabilities": ["test"],
  "schemaVersion": "1.0.0"
}`

	os.WriteFile(filepath.Join(configDir, "test-tool.json"), []byte(config), 0644)
	os.Setenv("TEST_TOKEN", "secret123")
	defer os.Unsetenv("TEST_TOKEN")

	loader := mcp.NewLoader("", tmpDir)
	configs, err := loader.LoadConfigs()

	if err != nil {
		t.Fatalf("failed to load configs: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}

	if configs[0].Env["API_TOKEN"] != "secret123" {
		t.Errorf("expected env var expansion, got: %s", configs[0].Env["API_TOKEN"])
	}
}
