package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
)

func TestHotReloadOnConfigChange(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".sdek", "mcp")
	os.MkdirAll(configDir, 0755)
	
	registry := mcp.NewRegistry()
	ctx := context.Background()
	
	// Initial load
	registry.Init(ctx)
	
	// Create new config file
	newConfig := `{
  "name": "new-tool",
  "command": "/bin/test",
  "transport": "stdio",
  "capabilities": ["test"],
  "schemaVersion": "1.0.0"
}`
	configPath := filepath.Join(configDir, "new-tool.json")
	os.WriteFile(configPath, []byte(newConfig), 0644)
	
	// Wait for file watcher to detect change
	time.Sleep(200 * time.Millisecond)
	
	// Verify tool was loaded
	tools, _ := registry.List(ctx)
	found := false
	for _, tool := range tools {
		if tool.Name == "new-tool" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("expected new tool to be hot-reloaded")
	}
}

func TestHotReloadOnConfigDelete(t *testing.T) {
	t.Skip("requires file watcher implementation")
}
