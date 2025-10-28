package tools

import (
	"context"
	"testing"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestToolRegistry_Register(t *testing.T) {
	tests := []struct {
		name    string
		tool    *types.Tool
		wantErr bool
	}{
		{
			name: "register builtin tool",
			tool: &types.Tool{
				Name:        "kubectl",
				Description: "Kubernetes CLI",
				Source:      types.ToolSourceBuiltin,
			},
			wantErr: false,
		},
		{
			name: "register MCP tool",
			tool: &types.Tool{
				Name:        "call_aws",
				Description: "AWS CLI wrapper",
				Source:      types.ToolSourceMCP,
				ServerName:  "aws-api",
			},
			wantErr: false,
		},
		{
			name: "register legacy tool",
			tool: &types.Tool{
				Name:        "github_legacy",
				Description: "Legacy GitHub connector",
				Source:      types.ToolSourceLegacy,
			},
			wantErr: false,
		},
		{
			name:    "register nil tool",
			tool:    nil,
			wantErr: true,
		},
		{
			name: "register tool with empty name",
			tool: &types.Tool{
				Name:        "",
				Description: "Invalid tool",
				Source:      types.ToolSourceBuiltin,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewToolRegistry(true, nil)
			err := registry.Register(tt.tool)

			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If registration succeeded, verify tool was added
			if err == nil {
				tool, err := registry.Get(tt.tool.Name)
				if err != nil {
					t.Errorf("Get() failed to retrieve registered tool: %v", err)
				}
				if tool.Name != tt.tool.Name {
					t.Errorf("Get() returned wrong tool: got %s, want %s", tool.Name, tt.tool.Name)
				}
			}
		})
	}
}

func TestToolRegistry_Get(t *testing.T) {
	registry := NewToolRegistry(true, nil) // preferMCP = true

	// Register tools from different sources with same name
	builtinTool := &types.Tool{
		Name:        "test_tool",
		Description: "Builtin version",
		Source:      types.ToolSourceBuiltin,
	}
	mcpTool := &types.Tool{
		Name:        "test_tool",
		Description: "MCP version",
		Source:      types.ToolSourceMCP,
		ServerName:  "test-server",
	}
	legacyTool := &types.Tool{
		Name:        "test_tool",
		Description: "Legacy version",
		Source:      types.ToolSourceLegacy,
	}

	registry.Register(builtinTool)
	registry.Register(mcpTool)
	registry.Register(legacyTool)

	// With preferMCP=true, should return MCP tool
	tool, err := registry.Get("test_tool")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if tool.Source != types.ToolSourceMCP {
		t.Errorf("Get() returned wrong source: got %s, want %s", tool.Source, types.ToolSourceMCP)
	}

	// Test with preferMCP=false
	registry2 := NewToolRegistry(false, nil)
	registry2.Register(builtinTool)
	registry2.Register(mcpTool)
	registry2.Register(legacyTool)

	tool2, err := registry2.Get("test_tool")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if tool2.Source != types.ToolSourceBuiltin {
		t.Errorf("Get() returned wrong source: got %s, want %s", tool2.Source, types.ToolSourceBuiltin)
	}

	// Test non-existent tool
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Get() should return error for non-existent tool")
	}
	if _, ok := err.(ErrToolNotFound); !ok {
		t.Errorf("Get() should return ErrToolNotFound, got %T", err)
	}
}

func TestToolRegistry_List(t *testing.T) {
	registry := NewToolRegistry(true, nil)

	// Register multiple tools
	tools := []*types.Tool{
		{Name: "tool1", Source: types.ToolSourceBuiltin},
		{Name: "tool2", Source: types.ToolSourceMCP, ServerName: "server1"},
		{Name: "tool3", Source: types.ToolSourceLegacy},
	}

	for _, tool := range tools {
		if err := registry.Register(tool); err != nil {
			t.Fatalf("Register() failed: %v", err)
		}
	}

	// List should return all tools
	allTools := registry.List()
	if len(allTools) != 3 {
		t.Errorf("List() returned %d tools, want 3", len(allTools))
	}

	// Register duplicate name (different source)
	dupeTool := &types.Tool{Name: "tool1", Source: types.ToolSourceMCP, ServerName: "server2"}
	registry.Register(dupeTool)

	// List should still return 3 tools (MCP shadows builtin)
	allTools = registry.List()
	if len(allTools) != 3 {
		t.Errorf("List() returned %d tools after dupe, want 3", len(allTools))
	}

	// The tool1 should be MCP version
	var tool1 *types.Tool
	for _, t := range allTools {
		if t.Name == "tool1" {
			tool1 = t
			break
		}
	}
	if tool1 == nil {
		t.Fatal("List() did not return tool1")
	}
	if tool1.Source != types.ToolSourceMCP {
		t.Errorf("tool1 has wrong source: got %s, want %s", tool1.Source, types.ToolSourceMCP)
	}
}

func TestToolRegistry_Count(t *testing.T) {
	registry := NewToolRegistry(true, nil)

	// Initially empty
	builtin, mcp, legacy := registry.Count()
	if builtin != 0 || mcp != 0 || legacy != 0 {
		t.Errorf("Count() for empty registry: got (%d, %d, %d), want (0, 0, 0)", builtin, mcp, legacy)
	}

	// Register tools
	registry.Register(&types.Tool{Name: "b1", Source: types.ToolSourceBuiltin})
	registry.Register(&types.Tool{Name: "b2", Source: types.ToolSourceBuiltin})
	registry.Register(&types.Tool{Name: "m1", Source: types.ToolSourceMCP, ServerName: "s1"})
	registry.Register(&types.Tool{Name: "l1", Source: types.ToolSourceLegacy})

	builtin, mcp, legacy = registry.Count()
	if builtin != 2 || mcp != 1 || legacy != 1 {
		t.Errorf("Count() got (%d, %d, %d), want (2, 1, 1)", builtin, mcp, legacy)
	}
}

func TestToolRegistry_Clear(t *testing.T) {
	registry := NewToolRegistry(true, nil)

	// Register tools
	registry.Register(&types.Tool{Name: "tool1", Source: types.ToolSourceBuiltin})
	registry.Register(&types.Tool{Name: "tool2", Source: types.ToolSourceMCP, ServerName: "s1"})

	// Verify tools exist
	if len(registry.List()) != 2 {
		t.Fatal("Setup failed: tools not registered")
	}

	// Clear
	registry.Clear()

	// Verify empty
	if len(registry.List()) != 0 {
		t.Errorf("Clear() did not remove all tools")
	}

	builtin, mcp, legacy := registry.Count()
	if builtin != 0 || mcp != 0 || legacy != 0 {
		t.Errorf("Count() after Clear(): got (%d, %d, %d), want (0, 0, 0)", builtin, mcp, legacy)
	}
}

func TestToolRegistry_Analyze(t *testing.T) {
	registry := NewToolRegistry(true, nil)

	// Test safe command
	safeCall := &types.ToolCall{
		ToolName:  "safe_tool",
		Arguments: map[string]interface{}{"query": "list users"},
		Context:   map[string]string{},
	}

	analysis := registry.Analyze(safeCall)
	if analysis.RequiresApproval {
		t.Error("Safe command should not require approval")
	}
	if analysis.RiskLevel != types.RiskLevelLow {
		t.Errorf("Safe command should have low risk, got %s", analysis.RiskLevel)
	}

	// Test dangerous command
	dangerousCall := &types.ToolCall{
		ToolName:  "dangerous_tool",
		Arguments: map[string]interface{}{"command": "delete all users"},
		Context:   map[string]string{},
	}

	analysis = registry.Analyze(dangerousCall)
	if !analysis.RequiresApproval {
		t.Error("Dangerous command should require approval")
	}
	if analysis.RiskLevel != types.RiskLevelMedium {
		t.Errorf("Dangerous command should have medium risk, got %s", analysis.RiskLevel)
	}
}

func TestToolRegistry_Concurrent(t *testing.T) {
	registry := NewToolRegistry(true, nil)

	// Register a tool
	registry.Register(&types.Tool{
		Name:        "concurrent_tool",
		Description: "Tool for concurrent access test",
		Source:      types.ToolSourceBuiltin,
	})

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := registry.Get("concurrent_tool")
			if err != nil {
				t.Errorf("Concurrent Get() failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent List
	for i := 0; i < 10; i++ {
		go func() {
			registry.List()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestToolRegistry_ExecuteRequiresApproval tests that dangerous tools require approval
func TestToolRegistry_ExecuteRequiresApproval(t *testing.T) {
	registry := NewToolRegistry(true, nil)

	// Register a builtin dangerous tool
	dangerousTool := &types.Tool{
		Name:        "delete_tool",
		Description: "Dangerous deletion tool",
		Source:      types.ToolSourceBuiltin,
	}
	registry.Register(dangerousTool)

	ctx := context.Background()

	// Try to execute without approval
	callWithoutApproval := &types.ToolCall{
		ToolName:  "delete_tool",
		Arguments: map[string]interface{}{"command": "delete something"},
		Context:   map[string]string{},
	}

	_, err := registry.Execute(ctx, callWithoutApproval)
	if err == nil {
		t.Error("Execute() should require approval for dangerous tool")
	}
	if _, ok := err.(ErrApprovalRequired); !ok {
		t.Errorf("Execute() should return ErrApprovalRequired, got %T: %v", err, err)
	}
}
