package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// ToolRegistry maintains a unified catalog of all available tools from different sources.
// It combines builtin tools, MCP tools, and legacy connector tools into a single registry
// with consistent interfaces and safety validation.
type ToolRegistry struct {
	mu sync.RWMutex

	// Tool maps by source
	builtinTools map[string]*types.Tool // Built-in tools (kubectl, bash, etc.)
	mcpTools     map[string]*types.Tool // Tools from MCP servers
	legacyTools  map[string]*types.Tool // Wrapped legacy connectors

	// Safety validator for tool call analysis
	safetyValidator *SafetyValidator

	// MCP Manager for executing MCP tools
	mcpManager *mcp.MCPManager

	// Configuration
	preferMCP bool // If true, MCP tools shadow builtin/legacy tools with same name
}

// NewToolRegistry creates a new tool registry with the given configuration.
func NewToolRegistry(preferMCP bool, mcpManager *mcp.MCPManager) *ToolRegistry {
	return &ToolRegistry{
		builtinTools:    make(map[string]*types.Tool),
		mcpTools:        make(map[string]*types.Tool),
		legacyTools:     make(map[string]*types.Tool),
		safetyValidator: NewSafetyValidator(),
		mcpManager:      mcpManager,
		preferMCP:       preferMCP,
	}
}

// Register adds a tool to the appropriate registry based on its source.
// If a tool with the same name already exists in the target source, it is replaced.
func (r *ToolRegistry) Register(tool *types.Tool) error {
	if tool == nil {
		return fmt.Errorf("cannot register nil tool")
	}

	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	switch tool.Source {
	case types.ToolSourceBuiltin:
		r.builtinTools[tool.Name] = tool
	case types.ToolSourceMCP:
		r.mcpTools[tool.Name] = tool
	case types.ToolSourceLegacy:
		r.legacyTools[tool.Name] = tool
	default:
		return fmt.Errorf("unknown tool source: %s", tool.Source)
	}

	return nil
}

// List returns all registered tools, respecting the preference order:
// - If preferMCP is true: MCP > builtin > legacy
// - If preferMCP is false: builtin > MCP > legacy
// Tools with duplicate names are resolved according to preference.
func (r *ToolRegistry) List() []*types.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Use a map to handle deduplication with preference
	merged := make(map[string]*types.Tool)

	// Apply tools in reverse preference order (later overwrites earlier)
	if r.preferMCP {
		// Legacy < Builtin < MCP
		for name, tool := range r.legacyTools {
			merged[name] = tool
		}
		for name, tool := range r.builtinTools {
			merged[name] = tool
		}
		for name, tool := range r.mcpTools {
			merged[name] = tool
		}
	} else {
		// Legacy < MCP < Builtin
		for name, tool := range r.legacyTools {
			merged[name] = tool
		}
		for name, tool := range r.mcpTools {
			merged[name] = tool
		}
		for name, tool := range r.builtinTools {
			merged[name] = tool
		}
	}

	// Convert map to slice
	tools := make([]*types.Tool, 0, len(merged))
	for _, tool := range merged {
		tools = append(tools, tool)
	}

	return tools
}

// Get retrieves a tool by name, respecting preference order.
// Returns ErrToolNotFound if the tool doesn't exist.
func (r *ToolRegistry) Get(name string) (*types.Tool, error) {
	if name == "" {
		return nil, fmt.Errorf("tool name cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Search in preference order
	if r.preferMCP {
		// MCP > Builtin > Legacy
		if tool, ok := r.mcpTools[name]; ok {
			return tool, nil
		}
		if tool, ok := r.builtinTools[name]; ok {
			return tool, nil
		}
		if tool, ok := r.legacyTools[name]; ok {
			return tool, nil
		}
	} else {
		// Builtin > MCP > Legacy
		if tool, ok := r.builtinTools[name]; ok {
			return tool, nil
		}
		if tool, ok := r.mcpTools[name]; ok {
			return tool, nil
		}
		if tool, ok := r.legacyTools[name]; ok {
			return tool, nil
		}
	}

	return nil, ErrToolNotFound{Name: name}
}

// Execute runs a tool with the given call parameters.
// It performs safety validation before execution and routes to the appropriate executor.
func (r *ToolRegistry) Execute(ctx context.Context, call *types.ToolCall) (*types.ToolExecutionResult, error) {
	if call == nil {
		return nil, fmt.Errorf("tool call cannot be nil")
	}

	// Get the tool
	tool, err := r.Get(call.ToolName)
	if err != nil {
		return nil, err
	}

	// Perform safety analysis
	analysis := r.Analyze(call)

	// If requires approval and not already approved in context, return error
	if analysis.RequiresApproval {
		approved, ok := call.Context["approved"]
		if !ok || approved != "true" {
			return nil, ErrApprovalRequired{
				ToolName:  call.ToolName,
				RiskLevel: string(analysis.RiskLevel),
				Rationale: analysis.Rationale,
			}
		}
	}

	// Route execution based on tool source
	switch tool.Source {
	case types.ToolSourceMCP:
		// Execute via MCP Manager
		if r.mcpManager == nil {
			return nil, fmt.Errorf("MCP manager not available for tool %s", tool.Name)
		}
		return r.executeMCPTool(ctx, tool, call)

	case types.ToolSourceBuiltin, types.ToolSourceLegacy:
		// Builtin and legacy tools not yet implemented in Phase 5
		return nil, fmt.Errorf("execution not implemented for source: %s", tool.Source)

	default:
		return nil, fmt.Errorf("unknown tool source: %s", tool.Source)
	}
}

// executeMCPTool executes a tool via the MCP Manager.
func (r *ToolRegistry) executeMCPTool(ctx context.Context, tool *types.Tool, call *types.ToolCall) (*types.ToolExecutionResult, error) {
	// Execute via MCP manager
	result, err := r.mcpManager.ExecuteTool(ctx, tool.ServerName, tool.Name, call.Arguments)
	if err != nil {
		return &types.ToolExecutionResult{
			ToolName:  tool.Name,
			Success:   false,
			Error:     err.Error(),
			Timestamp: fmt.Sprintf("%d", ctx.Value("start_time")),
		}, err
	}

	// Normalize to ToolExecutionResult
	// For now, return a simple success result
	return &types.ToolExecutionResult{
		ToolName:  tool.Name,
		Success:   true,
		Output:    result,
		Timestamp: fmt.Sprintf("%d", ctx.Value("start_time")),
	}, nil
}

// Analyze performs safety analysis on a tool call.
// It delegates to the SafetyValidator to determine if the call is safe,
// interactive, or modifies resources.
func (r *ToolRegistry) Analyze(call *types.ToolCall) *types.ToolCallAnalysis {
	return r.safetyValidator.Analyze(call)
}

// Count returns the number of tools in each source category.
func (r *ToolRegistry) Count() (builtin, mcp, legacy int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.builtinTools), len(r.mcpTools), len(r.legacyTools)
}

// Clear removes all tools from the registry.
func (r *ToolRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.builtinTools = make(map[string]*types.Tool)
	r.mcpTools = make(map[string]*types.Tool)
	r.legacyTools = make(map[string]*types.Tool)
}

// ErrToolNotFound is returned when a tool is not found in the registry.
type ErrToolNotFound struct {
	Name string
}

func (e ErrToolNotFound) Error() string {
	return fmt.Sprintf("tool not found: %s", e.Name)
}

// ErrApprovalRequired is returned when a tool call requires user approval.
type ErrApprovalRequired struct {
	ToolName  string
	RiskLevel string
	Rationale string
}

func (e ErrApprovalRequired) Error() string {
	return fmt.Sprintf("tool %s requires approval (risk: %s): %s", e.ToolName, e.RiskLevel, e.Rationale)
}
