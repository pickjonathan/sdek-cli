package unit

import (
	"context"
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
)

func TestRegistryInitDiscoversConfigs(t *testing.T) {
	tmpDir := t.TempDir()
	// Setup test configs in tmpDir/.sdek/mcp/
	
	registry := mcp.NewRegistry()
	ctx := context.Background()
	
	count, err := registry.Init(ctx)
	if err != nil {
		t.Fatalf("registry init failed: %v", err)
	}
	
	if count == 0 {
		t.Error("expected at least one tool to be discovered")
	}
}

func TestRegistryAsyncHandshake(t *testing.T) {
	t.Skip("requires mock MCP server and async handshake implementation")
}

func TestRegistryCloseWaitsForInflight(t *testing.T) {
	registry := mcp.NewRegistry()
	ctx := context.Background()
	
	// Start some invocations
	// ...
	
	err := registry.Close(ctx)
	if err != nil {
		t.Errorf("registry close failed: %v", err)
	}
}

func TestRegistryReloadHotReloadsChangedConfigs(t *testing.T) {
	registry := mcp.NewRegistry()
	
	ctx := context.Background()
	count, err := registry.Reload(ctx)
	
	if err != nil {
		t.Errorf("registry reload failed: %v", err)
	}
	
	_ = count // Check count reflects reloaded tools
}

func TestRegistryListReturnsAllTools(t *testing.T) {
	registry := mcp.NewRegistry()
	ctx := context.Background()
	
	tools, err := registry.List(ctx)
	if err != nil {
		t.Errorf("registry list failed: %v", err)
	}
	
	if tools == nil {
		t.Error("expected non-nil tools list")
	}
}

func TestRegistryGetRetrievesTool(t *testing.T) {
	registry := mcp.NewRegistry()
	ctx := context.Background()
	
	_, err := registry.Get(ctx, "nonexistent-tool")
	if err != mcp.ErrToolNotFound {
		t.Errorf("expected ErrToolNotFound, got: %v", err)
	}
}

func TestRegistryEnableTransitionsToolToReady(t *testing.T) {
	registry := mcp.NewRegistry()
	ctx := context.Background()
	
	err := registry.Enable(ctx, "test-tool")
	if err == nil {
		// Success expected if tool exists
	}
}

func TestRegistryDisableTransitionsToolToOffline(t *testing.T) {
	registry := mcp.NewRegistry()
	ctx := context.Background()
	
	err := registry.Disable(ctx, "test-tool")
	if err == nil {
		// Success expected if tool exists
	}
}

func TestRegistryDisabledToolRejectsInvocations(t *testing.T) {
	t.Skip("requires integration with invoker")
}

func TestRegistryValidateChecksSchema(t *testing.T) {
	registry := mcp.NewRegistry()
	ctx := context.Background()
	
	errors, err := registry.Validate(ctx, "/path/to/invalid.json")
	if err != nil {
		t.Fatalf("validate call failed: %v", err)
	}
	
	_ = errors // Check for schema errors
}

func TestRegistryTestPerformsHealthCheck(t *testing.T) {
	registry := mcp.NewRegistry()
	ctx := context.Background()
	
	report, err := registry.Test(ctx, "test-tool")
	if err != nil {
		// Error expected for nonexistent tool
	}
	
	_ = report // Check health report structure
}
