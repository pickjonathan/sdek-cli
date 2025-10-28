package types

// Tool represents an executable capability from any source (builtin, MCP, legacy).
type Tool struct {
	// Name is the unique tool identifier
	Name string `json:"name"`

	// Description is the natural language description
	Description string `json:"description"`

	// Parameters is the input parameter schema (JSON Schema)
	Parameters map[string]interface{} `json:"parameters"`

	// Source is the tool origin: "builtin", "mcp", or "legacy"
	Source ToolSource `json:"source"`

	// ServerName is the MCP server name (if source = "mcp")
	ServerName string `json:"server_name,omitempty"`

	// SafetyTier is the safety classification
	SafetyTier SafetyTier `json:"safety_tier"`
}

// ToolSource represents the origin of a tool.
type ToolSource string

const (
	// ToolSourceBuiltin represents built-in tools
	ToolSourceBuiltin ToolSource = "builtin"

	// ToolSourceMCP represents tools from MCP servers
	ToolSourceMCP ToolSource = "mcp"

	// ToolSourceLegacy represents wrapped legacy connectors
	ToolSourceLegacy ToolSource = "legacy"
)

// SafetyTier represents the safety classification of a tool.
type SafetyTier string

const (
	// SafetyTierSafe represents read-only operations with no side effects
	SafetyTierSafe SafetyTier = "safe"

	// SafetyTierInteractive represents tools requiring terminal interaction
	SafetyTierInteractive SafetyTier = "interactive"

	// SafetyTierModifiesResource represents tools that may mutate system state
	SafetyTierModifiesResource SafetyTier = "modifies_resource"
)

// ToolCall represents a request to execute a tool.
type ToolCall struct {
	// ToolName is the name of tool to execute
	ToolName string `json:"tool_name"`

	// Arguments contains input parameters as JSON object
	Arguments map[string]interface{} `json:"arguments"`

	// Context contains additional context (user_id, session_id, etc.)
	Context map[string]string `json:"context,omitempty"`
}

// ToolCallAnalysis contains the safety analysis result for a tool call.
type ToolCallAnalysis struct {
	// IsInteractive indicates if tool requires interactive terminal
	IsInteractive bool `json:"is_interactive"`

	// ModifiesResource indicates if tool may mutate system state
	ModifiesResource bool `json:"modifies_resource"`

	// RequiresApproval indicates if user confirmation is required
	RequiresApproval bool `json:"requires_approval"`

	// RiskLevel is the risk classification
	RiskLevel RiskLevel `json:"risk_level"`

	// Rationale is the explanation of safety assessment
	Rationale string `json:"rationale"`
}

// RiskLevel represents the risk classification of a tool call.
type RiskLevel string

const (
	// RiskLevelLow represents safe operations
	RiskLevelLow RiskLevel = "low"

	// RiskLevelMedium represents operations that modify resources
	RiskLevelMedium RiskLevel = "medium"

	// RiskLevelHigh represents interactive or dangerous operations
	RiskLevelHigh RiskLevel = "high"
)

// ToolExecutionResult contains the result of a tool execution.
type ToolExecutionResult struct {
	// ToolName is the tool that was executed
	ToolName string `json:"tool_name"`

	// Success indicates whether execution succeeded
	Success bool `json:"success"`

	// Output contains the result data (format varies by tool)
	Output interface{} `json:"output,omitempty"`

	// Error is the error message if success = false
	Error string `json:"error,omitempty"`

	// LatencyMS is the execution time in milliseconds
	LatencyMS int64 `json:"latency_ms"`

	// Timestamp is when execution completed
	Timestamp string `json:"timestamp"`
}
