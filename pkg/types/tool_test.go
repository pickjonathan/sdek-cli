package types

import (
	"testing"
)

func TestToolSources(t *testing.T) {
	tests := []struct {
		name   string
		source ToolSource
		want   string
	}{
		{"builtin", ToolSourceBuiltin, "builtin"},
		{"mcp", ToolSourceMCP, "mcp"},
		{"legacy", ToolSourceLegacy, "legacy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.source) != tt.want {
				t.Errorf("Expected source to be '%s', got '%s'", tt.want, tt.source)
			}
		})
	}
}

func TestSafetyTiers(t *testing.T) {
	tests := []struct {
		name string
		tier SafetyTier
		want string
	}{
		{"safe", SafetyTierSafe, "safe"},
		{"interactive", SafetyTierInteractive, "interactive"},
		{"modifies", SafetyTierModifiesResource, "modifies_resource"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.tier) != tt.want {
				t.Errorf("Expected tier to be '%s', got '%s'", tt.want, tt.tier)
			}
		})
	}
}

func TestRiskLevels(t *testing.T) {
	tests := []struct {
		name  string
		level RiskLevel
		want  string
	}{
		{"low", RiskLevelLow, "low"},
		{"medium", RiskLevelMedium, "medium"},
		{"high", RiskLevelHigh, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.level) != tt.want {
				t.Errorf("Expected level to be '%s', got '%s'", tt.want, tt.level)
			}
		})
	}
}

func TestToolCreation(t *testing.T) {
	tool := Tool{
		Name:        "call_aws",
		Description: "Execute AWS CLI commands",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type": "string",
				},
			},
		},
		Source:     ToolSourceMCP,
		ServerName: "aws-api",
		SafetyTier: SafetyTierSafe,
	}

	if tool.Name != "call_aws" {
		t.Errorf("Expected Name to be 'call_aws', got '%s'", tool.Name)
	}

	if tool.Source != ToolSourceMCP {
		t.Errorf("Expected Source to be ToolSourceMCP, got '%s'", tool.Source)
	}

	if tool.ServerName != "aws-api" {
		t.Errorf("Expected ServerName to be 'aws-api', got '%s'", tool.ServerName)
	}

	if tool.SafetyTier != SafetyTierSafe {
		t.Errorf("Expected SafetyTier to be SafetyTierSafe, got '%s'", tool.SafetyTier)
	}
}

func TestBuiltinTool(t *testing.T) {
	tool := Tool{
		Name:        "kubectl",
		Description: "Execute kubectl commands",
		Parameters:  map[string]interface{}{"type": "object"},
		Source:      ToolSourceBuiltin,
		SafetyTier:  SafetyTierModifiesResource,
	}

	if tool.Source != ToolSourceBuiltin {
		t.Errorf("Expected Source to be ToolSourceBuiltin, got '%s'", tool.Source)
	}

	if tool.ServerName != "" {
		t.Errorf("Expected ServerName to be empty for builtin tool, got '%s'", tool.ServerName)
	}

	if tool.SafetyTier != SafetyTierModifiesResource {
		t.Errorf("Expected SafetyTier to be SafetyTierModifiesResource, got '%s'", tool.SafetyTier)
	}
}

func TestLegacyTool(t *testing.T) {
	tool := Tool{
		Name:        "legacy_github",
		Description: "GitHub connector (legacy)",
		Parameters:  map[string]interface{}{"type": "object"},
		Source:      ToolSourceLegacy,
		SafetyTier:  SafetyTierSafe,
	}

	if tool.Source != ToolSourceLegacy {
		t.Errorf("Expected Source to be ToolSourceLegacy, got '%s'", tool.Source)
	}
}

func TestToolCall(t *testing.T) {
	call := ToolCall{
		ToolName: "call_aws",
		Arguments: map[string]interface{}{
			"command": "iam list-users --output json",
		},
		Context: map[string]string{
			"session_id": "test-session",
			"user_id":    "analyst@example.com",
		},
	}

	if call.ToolName != "call_aws" {
		t.Errorf("Expected ToolName to be 'call_aws', got '%s'", call.ToolName)
	}

	if call.Arguments["command"] != "iam list-users --output json" {
		t.Errorf("Expected command argument, got '%v'", call.Arguments["command"])
	}

	if call.Context["session_id"] != "test-session" {
		t.Errorf("Expected session_id context, got '%s'", call.Context["session_id"])
	}
}

func TestToolCallAnalysis(t *testing.T) {
	tests := []struct {
		name     string
		analysis ToolCallAnalysis
	}{
		{
			name: "safe operation",
			analysis: ToolCallAnalysis{
				IsInteractive:    false,
				ModifiesResource: false,
				RequiresApproval: false,
				RiskLevel:        RiskLevelLow,
				Rationale:        "Read-only operation",
			},
		},
		{
			name: "interactive operation",
			analysis: ToolCallAnalysis{
				IsInteractive:    true,
				ModifiesResource: false,
				RequiresApproval: true,
				RiskLevel:        RiskLevelHigh,
				Rationale:        "Requires interactive terminal",
			},
		},
		{
			name: "modifies resources",
			analysis: ToolCallAnalysis{
				IsInteractive:    false,
				ModifiesResource: true,
				RequiresApproval: true,
				RiskLevel:        RiskLevelMedium,
				Rationale:        "May modify system state",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.analysis.IsInteractive && tt.analysis.RiskLevel != RiskLevelHigh {
				t.Error("Expected RiskLevelHigh for interactive operations")
			}

			if tt.analysis.ModifiesResource && !tt.analysis.IsInteractive && tt.analysis.RiskLevel != RiskLevelMedium {
				t.Error("Expected RiskLevelMedium for resource-modifying operations")
			}

			if !tt.analysis.IsInteractive && !tt.analysis.ModifiesResource && tt.analysis.RiskLevel != RiskLevelLow {
				t.Error("Expected RiskLevelLow for safe operations")
			}

			if (tt.analysis.IsInteractive || tt.analysis.ModifiesResource) && !tt.analysis.RequiresApproval {
				t.Error("Expected RequiresApproval for risky operations")
			}
		})
	}
}

func TestToolExecutionResult(t *testing.T) {
	result := ToolExecutionResult{
		ToolName:  "call_aws",
		Success:   true,
		Output:    map[string]interface{}{"Users": []string{"user1", "user2"}},
		LatencyMS: 450,
		Timestamp: "2025-10-26T10:30:15Z",
	}

	if result.ToolName != "call_aws" {
		t.Errorf("Expected ToolName to be 'call_aws', got '%s'", result.ToolName)
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}

	if result.Error != "" {
		t.Errorf("Expected Error to be empty for successful execution, got '%s'", result.Error)
	}

	if result.LatencyMS != 450 {
		t.Errorf("Expected LatencyMS to be 450, got %d", result.LatencyMS)
	}
}

func TestToolExecutionResultFailure(t *testing.T) {
	result := ToolExecutionResult{
		ToolName:  "call_aws",
		Success:   false,
		Error:     "AWS MCP server timeout after 60s",
		LatencyMS: 60000,
		Timestamp: "2025-10-26T10:31:00Z",
	}

	if result.Success {
		t.Error("Expected Success to be false for failed execution")
	}

	if result.Error == "" {
		t.Error("Expected Error message for failed execution")
	}

	if result.Output != nil {
		t.Error("Expected Output to be nil for failed execution")
	}

	if result.LatencyMS != 60000 {
		t.Errorf("Expected LatencyMS to be 60000, got %d", result.LatencyMS)
	}
}

func TestToolCallWithComplexArguments(t *testing.T) {
	call := ToolCall{
		ToolName: "complex_tool",
		Arguments: map[string]interface{}{
			"string_arg":  "value",
			"int_arg":     42,
			"bool_arg":    true,
			"array_arg":   []string{"a", "b", "c"},
			"nested_arg": map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
		},
		Context: map[string]string{
			"session_id": "test",
		},
	}

	if call.Arguments["string_arg"] != "value" {
		t.Error("String argument not preserved")
	}

	if call.Arguments["int_arg"] != 42 {
		t.Error("Int argument not preserved")
	}

	if call.Arguments["bool_arg"] != true {
		t.Error("Bool argument not preserved")
	}

	arr, ok := call.Arguments["array_arg"].([]string)
	if !ok || len(arr) != 3 {
		t.Error("Array argument not preserved")
	}

	nested, ok := call.Arguments["nested_arg"].(map[string]interface{})
	if !ok || nested["key1"] != "value1" {
		t.Error("Nested argument not preserved")
	}
}

func TestToolCallAnalysisRationale(t *testing.T) {
	tests := []struct {
		name      string
		rationale string
		wantEmpty bool
	}{
		{"with rationale", "Command 'aws ec2 terminate-instances' detected", false},
		{"empty rationale", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := ToolCallAnalysis{
				Rationale: tt.rationale,
			}

			isEmpty := analysis.Rationale == ""
			if isEmpty != tt.wantEmpty {
				t.Errorf("Expected empty=%v, got empty=%v", tt.wantEmpty, isEmpty)
			}
		})
	}
}

func TestToolParameterSchema(t *testing.T) {
	tool := Tool{
		Name: "test_tool",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"required_param": map[string]interface{}{
					"type":        "string",
					"description": "A required parameter",
				},
				"optional_param": map[string]interface{}{
					"type":        "number",
					"description": "An optional parameter",
				},
			},
			"required": []string{"required_param"},
		},
	}

	params, ok := tool.Parameters["properties"].(map[string]interface{})
	if !ok {
		t.Error("Expected properties to be a map")
	}

	if _, ok := params["required_param"]; !ok {
		t.Error("Expected required_param in properties")
	}

	if _, ok := params["optional_param"]; !ok {
		t.Error("Expected optional_param in properties")
	}

	required, ok := tool.Parameters["required"].([]string)
	if !ok || len(required) != 1 || required[0] != "required_param" {
		t.Error("Expected required_param in required array")
	}
}
