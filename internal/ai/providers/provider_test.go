package providers

import (
	"context"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai"
)

// TestProviderRegistration verifies that providers register themselves correctly
func TestProviderRegistration(t *testing.T) {
	tests := []struct {
		name         string
		provider     string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "OpenAI provider is registered",
			provider:    "openai",
			expectError: true, // Will fail without API key
		},
		{
			name:        "Anthropic provider is registered",
			provider:    "anthropic",
			expectError: true, // Will fail without API key
		},
		{
			name:         "Unknown provider returns error",
			provider:     "unknown",
			expectError:  true,
			errorMessage: "no factory registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ai.AIConfig{
				Provider:  tt.provider,
				Enabled:   true,
				Model:     "test-model",
				MaxTokens: 100,
				Timeout:   30,
			}

			_, err := ai.CreateProviderFromRegistry(tt.provider, cfg)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestOpenAIProvider_AnalyzeWithContext tests the OpenAI provider
func TestOpenAIProvider_AnalyzeWithContext(t *testing.T) {
	t.Skip("Skipping until valid API key is available")

	cfg := ai.AIConfig{
		Provider:    "openai",
		Enabled:     true,
		Model:       "gpt-3.5-turbo",
		MaxTokens:   100,
		Temperature: 0.3,
		Timeout:     30,
		OpenAIKey:   "test-key", // Replace with real key or use env var
	}

	engine, err := NewOpenAIEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create OpenAI engine: %v", err)
	}

	// Test AnalyzeWithContext
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	prompt := "What is 2+2? Answer in one word."
	response, err := engine.AnalyzeWithContext(ctx, prompt)
	if err != nil {
		t.Fatalf("AnalyzeWithContext failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	// Test tracking
	if engine.GetCallCount() != 1 {
		t.Errorf("Expected call count 1, got %d", engine.GetCallCount())
	}

	if engine.GetLastPrompt() != prompt {
		t.Errorf("Expected last prompt %q, got %q", prompt, engine.GetLastPrompt())
	}
}

// TestAnthropicProvider_AnalyzeWithContext tests the Anthropic provider
func TestAnthropicProvider_AnalyzeWithContext(t *testing.T) {
	t.Skip("Skipping until valid API key is available")

	cfg := ai.AIConfig{
		Provider:     "anthropic",
		Enabled:      true,
		Model:        "claude-3-haiku-20240307",
		MaxTokens:    100,
		Temperature:  0.3,
		Timeout:      30,
		AnthropicKey: "test-key", // Replace with real key or use env var
	}

	engine, err := NewAnthropicEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create Anthropic engine: %v", err)
	}

	// Test AnalyzeWithContext
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	prompt := "What is 2+2? Answer in one word."
	response, err := engine.AnalyzeWithContext(ctx, prompt)
	if err != nil {
		t.Fatalf("AnalyzeWithContext failed: %v", err)
	}

	if response == "" {
		t.Error("Expected non-empty response")
	}

	// Test tracking
	if engine.GetCallCount() != 1 {
		t.Errorf("Expected call count 1, got %d", engine.GetCallCount())
	}

	if engine.GetLastPrompt() != prompt {
		t.Errorf("Expected last prompt %q, got %q", prompt, engine.GetLastPrompt())
	}
}

// TestProvider_CallCountTracking tests that call counts increment correctly
func TestProvider_CallCountTracking(t *testing.T) {
	t.Skip("Skipping until valid API key is available")

	cfg := ai.AIConfig{
		Provider:  "openai",
		Enabled:   true,
		Model:     "gpt-3.5-turbo",
		MaxTokens: 50,
		Timeout:   30,
		OpenAIKey: "test-key", // Replace with real key
	}

	engine, err := NewOpenAIEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	ctx := context.Background()

	// Make multiple calls
	for i := 1; i <= 3; i++ {
		_, err := engine.AnalyzeWithContext(ctx, "Test prompt")
		if err != nil {
			t.Fatalf("Call %d failed: %v", i, err)
		}

		if engine.GetCallCount() != i {
			t.Errorf("After call %d, expected count %d, got %d", i, i, engine.GetCallCount())
		}
	}
}

// TestProvider_ContextCancellation tests that context cancellation is handled
func TestProvider_ContextCancellation(t *testing.T) {
	t.Skip("Skipping until valid API key is available")

	cfg := ai.AIConfig{
		Provider:  "openai",
		Enabled:   true,
		Model:     "gpt-3.5-turbo",
		MaxTokens: 100,
		Timeout:   60,
		OpenAIKey: "test-key",
	}

	engine, err := NewOpenAIEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = engine.AnalyzeWithContext(ctx, "Test prompt")
	if err == nil {
		t.Error("Expected error from cancelled context, got none")
	}
}

// TestProvider_EmptyPrompt tests error handling for empty prompts
func TestProvider_EmptyPrompt(t *testing.T) {
	cfg := ai.AIConfig{
		Provider:  "openai",
		Enabled:   true,
		Model:     "gpt-3.5-turbo",
		MaxTokens: 100,
		OpenAIKey: "test-key",
	}

	engine, err := NewOpenAIEngine(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	ctx := context.Background()
	_, err = engine.AnalyzeWithContext(ctx, "")
	if err == nil {
		t.Error("Expected error for empty prompt, got none")
	}
}

// TestProvider_FactoryPattern tests the factory registration pattern
func TestProvider_FactoryPattern(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		apiKey   string
		wantErr  bool
	}{
		{
			name:     "OpenAI with valid config",
			provider: "openai",
			apiKey:   "sk-test",
			wantErr:  false,
		},
		{
			name:     "Anthropic with valid config",
			provider: "anthropic",
			apiKey:   "sk-ant-test",
			wantErr:  false,
		},
		{
			name:     "OpenAI without API key",
			provider: "openai",
			apiKey:   "",
			wantErr:  true,
		},
		{
			name:     "Unknown provider",
			provider: "unknown",
			apiKey:   "test",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ai.AIConfig{
				Provider:  tt.provider,
				Enabled:   true,
				Model:     "test-model",
				MaxTokens: 100,
				Timeout:   30,
			}

			// Set the appropriate API key
			if tt.provider == "openai" {
				cfg.OpenAIKey = tt.apiKey
			} else if tt.provider == "anthropic" {
				cfg.AnthropicKey = tt.apiKey
			}

			_, err := ai.CreateProviderFromRegistry(tt.provider, cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProviderFromRegistry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
