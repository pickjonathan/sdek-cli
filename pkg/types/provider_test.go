package types

import (
	"testing"
)

func TestDefaultProviderConfig(t *testing.T) {
	cfg := DefaultProviderConfig()

	if cfg.Timeout != 60 {
		t.Errorf("Expected Timeout to be 60, got %d", cfg.Timeout)
	}

	if cfg.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", cfg.MaxRetries)
	}

	if cfg.Temperature != 0.0 {
		t.Errorf("Expected Temperature to be 0.0, got %f", cfg.Temperature)
	}

	if cfg.MaxTokens != 4096 {
		t.Errorf("Expected MaxTokens to be 4096, got %d", cfg.MaxTokens)
	}

	if cfg.Extra == nil {
		t.Error("Expected Extra map to be initialized")
	}
}

func TestProviderConfigURLSchemes(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"OpenAI", "openai://api.openai.com", "openai://api.openai.com"},
		{"Anthropic", "anthropic://api.anthropic.com", "anthropic://api.anthropic.com"},
		{"Gemini", "gemini://generativelanguage.googleapis.com", "gemini://generativelanguage.googleapis.com"},
		{"Bedrock", "bedrock://us-east-1", "bedrock://us-east-1"},
		{"Vertex AI", "vertexai://us-central1", "vertexai://us-central1"},
		{"Ollama", "ollama://localhost:11434", "ollama://localhost:11434"},
		{"LlamaCpp", "llamacpp://localhost:8080", "llamacpp://localhost:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ProviderConfig{
				URL:   tt.url,
				Model: "test-model",
			}

			if cfg.URL != tt.want {
				t.Errorf("Expected URL to be '%s', got '%s'", tt.want, cfg.URL)
			}
		})
	}
}

func TestProviderConfigWithAPIKey(t *testing.T) {
	cfg := ProviderConfig{
		URL:    "openai://api.openai.com",
		APIKey: "${OPENAI_API_KEY}",
		Model:  "gpt-4o",
	}

	if cfg.APIKey != "${OPENAI_API_KEY}" {
		t.Errorf("Expected APIKey to be '${OPENAI_API_KEY}', got '%s'", cfg.APIKey)
	}
}

func TestProviderConfigTemperatureRange(t *testing.T) {
	tests := []struct {
		name        string
		temperature float64
	}{
		{"minimum", 0.0},
		{"low", 0.5},
		{"medium", 1.0},
		{"high", 1.5},
		{"maximum", 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ProviderConfig{
				Temperature: tt.temperature,
			}

			if cfg.Temperature != tt.temperature {
				t.Errorf("Expected Temperature to be %f, got %f", tt.temperature, cfg.Temperature)
			}
		})
	}
}

func TestChatSession(t *testing.T) {
	session := ChatSession{
		ID:       "test-session-id",
		Messages: []Message{},
		Functions: []FunctionDefinition{
			{
				Name:        "test_function",
				Description: "A test function",
				Parameters:  map[string]interface{}{"type": "object"},
			},
		},
		Config: SessionConfig{
			Temperature: 0.0,
			MaxTokens:   4096,
		},
		Metadata: map[string]string{
			"framework":  "soc2",
			"control_id": "CC6.1",
		},
	}

	if session.ID != "test-session-id" {
		t.Errorf("Expected ID to be 'test-session-id', got '%s'", session.ID)
	}

	if len(session.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(session.Functions))
	}

	if session.Functions[0].Name != "test_function" {
		t.Errorf("Expected function name to be 'test_function', got '%s'", session.Functions[0].Name)
	}

	if session.Config.Temperature != 0.0 {
		t.Errorf("Expected Temperature to be 0.0, got %f", session.Config.Temperature)
	}

	if session.Metadata["framework"] != "soc2" {
		t.Errorf("Expected framework metadata to be 'soc2', got '%s'", session.Metadata["framework"])
	}
}

func TestMessageRoles(t *testing.T) {
	tests := []struct {
		name string
		role string
	}{
		{"system", "system"},
		{"user", "user"},
		{"assistant", "assistant"},
		{"function", "function"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message{
				Role:      tt.role,
				Content:   "test content",
				Timestamp: "2025-10-26T10:00:00Z",
			}

			if msg.Role != tt.role {
				t.Errorf("Expected Role to be '%s', got '%s'", tt.role, msg.Role)
			}
		})
	}
}

func TestMessageWithFunctionCall(t *testing.T) {
	msg := Message{
		Role:    "assistant",
		Content: "I'll call the function",
		FunctionCall: &FunctionCall{
			Name: "call_aws",
			Arguments: map[string]interface{}{
				"command": "iam list-users",
			},
		},
		Timestamp: "2025-10-26T10:00:00Z",
	}

	if msg.FunctionCall == nil {
		t.Error("Expected FunctionCall to be set")
	}

	if msg.FunctionCall.Name != "call_aws" {
		t.Errorf("Expected FunctionCall.Name to be 'call_aws', got '%s'", msg.FunctionCall.Name)
	}

	if msg.FunctionCall.Arguments["command"] != "iam list-users" {
		t.Errorf("Expected command argument to be 'iam list-users', got '%v'", msg.FunctionCall.Arguments["command"])
	}
}

func TestMessageWithFunctionResult(t *testing.T) {
	msg := Message{
		Role:    "function",
		Content: "",
		FunctionResult: map[string]interface{}{
			"Users": []string{"user1", "user2"},
		},
		Timestamp: "2025-10-26T10:00:01Z",
	}

	if msg.FunctionResult == nil {
		t.Error("Expected FunctionResult to be set")
	}

	result, ok := msg.FunctionResult.(map[string]interface{})
	if !ok {
		t.Error("Expected FunctionResult to be a map")
	}

	users, ok := result["Users"].([]string)
	if !ok || len(users) != 2 {
		t.Errorf("Expected 2 users in result, got %v", result["Users"])
	}
}

func TestFunctionDefinition(t *testing.T) {
	fn := FunctionDefinition{
		Name:        "call_aws",
		Description: "Execute AWS CLI commands",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "AWS CLI command",
				},
			},
			"required": []string{"command"},
		},
	}

	if fn.Name != "call_aws" {
		t.Errorf("Expected Name to be 'call_aws', got '%s'", fn.Name)
	}

	if fn.Description != "Execute AWS CLI commands" {
		t.Errorf("Expected Description, got '%s'", fn.Description)
	}

	if fn.Parameters["type"] != "object" {
		t.Errorf("Expected Parameters type to be 'object', got '%v'", fn.Parameters["type"])
	}
}

func TestSessionConfig(t *testing.T) {
	config := SessionConfig{
		Temperature: 0.7,
		MaxTokens:   2048,
	}

	if config.Temperature != 0.7 {
		t.Errorf("Expected Temperature to be 0.7, got %f", config.Temperature)
	}

	if config.MaxTokens != 2048 {
		t.Errorf("Expected MaxTokens to be 2048, got %d", config.MaxTokens)
	}
}

func TestProviderConfigWithExtra(t *testing.T) {
	cfg := ProviderConfig{
		URL:   "ollama://localhost:11434",
		Model: "gemma3:12b",
		Extra: map[string]string{
			"num_ctx":      "8192",
			"num_predict":  "512",
			"temperature":  "0.0",
		},
	}

	if cfg.Extra["num_ctx"] != "8192" {
		t.Errorf("Expected num_ctx to be '8192', got '%s'", cfg.Extra["num_ctx"])
	}

	if cfg.Extra["num_predict"] != "512" {
		t.Errorf("Expected num_predict to be '512', got '%s'", cfg.Extra["num_predict"])
	}
}

func TestFunctionCallArguments(t *testing.T) {
	call := FunctionCall{
		Name: "test_tool",
		Arguments: map[string]interface{}{
			"arg1": "value1",
			"arg2": 42,
			"arg3": true,
			"arg4": []string{"a", "b", "c"},
		},
	}

	if call.Name != "test_tool" {
		t.Errorf("Expected Name to be 'test_tool', got '%s'", call.Name)
	}

	if call.Arguments["arg1"] != "value1" {
		t.Errorf("Expected arg1 to be 'value1', got '%v'", call.Arguments["arg1"])
	}

	if call.Arguments["arg2"] != 42 {
		t.Errorf("Expected arg2 to be 42, got '%v'", call.Arguments["arg2"])
	}

	if call.Arguments["arg3"] != true {
		t.Errorf("Expected arg3 to be true, got '%v'", call.Arguments["arg3"])
	}

	arr, ok := call.Arguments["arg4"].([]string)
	if !ok || len(arr) != 3 {
		t.Errorf("Expected arg4 to be []string with 3 elements, got '%v'", call.Arguments["arg4"])
	}
}
