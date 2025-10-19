package ai

import (
	"context"
	"strings"
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/ai/connectors"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// TestNewEngineFromConfig tests engine creation from configuration
func TestNewEngineFromConfig(t *testing.T) {
	t.Run("nil config returns error", func(t *testing.T) {
		engine, err := NewEngineFromConfig(nil)
		if err == nil {
			t.Fatal("expected error for nil config")
		}
		if engine != nil {
			t.Error("expected nil engine for nil config")
		}
	})

	t.Run("missing provider returns error", func(t *testing.T) {
		cfg := types.DefaultConfig()
		cfg.AI.Enabled = true
		cfg.AI.Provider = "invalid"

		engine, err := NewEngineFromConfig(cfg)
		if err == nil {
			t.Fatal("expected error for invalid provider")
		}
		if engine != nil {
			t.Error("expected nil engine for invalid provider")
		}
	})

	t.Run("OpenAI without API key returns error", func(t *testing.T) {
		cfg := types.DefaultConfig()
		cfg.AI.Enabled = true
		cfg.AI.Provider = types.AIProviderOpenAI
		cfg.AI.OpenAIKey = ""
		cfg.AI.APIKey = ""

		engine, err := NewEngineFromConfig(cfg)
		if err == nil {
			t.Fatal("expected error for missing OpenAI API key")
		}
		if engine != nil {
			t.Error("expected nil engine for missing API key")
		}
	})

	t.Run("Anthropic without API key returns error", func(t *testing.T) {
		cfg := types.DefaultConfig()
		cfg.AI.Enabled = true
		cfg.AI.Provider = types.AIProviderAnthropic
		cfg.AI.AnthropicKey = ""
		cfg.AI.APIKey = ""

		engine, err := NewEngineFromConfig(cfg)
		if err == nil {
			t.Fatal("expected error for missing Anthropic API key")
		}
		if engine != nil {
			t.Error("expected nil engine for missing API key")
		}
	})

	// Note: Provider implementation tests will be added once providers are implemented
	// For now, we expect errors since providers are not yet implemented
}

// TestBuildConnectorRegistry tests connector registry building
func TestBuildConnectorRegistry(t *testing.T) {
	t.Run("empty config returns nil", func(t *testing.T) {
		registry, err := buildConnectorRegistry(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if registry != nil {
			t.Error("expected nil registry for nil config")
		}

		registry, err = buildConnectorRegistry(map[string]types.ConnectorConfig{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if registry != nil {
			t.Error("expected nil registry for empty config")
		}
	})

	t.Run("disabled connectors are skipped", func(t *testing.T) {
		configs := map[string]types.ConnectorConfig{
			"github": {
				Enabled: false, // Disabled
				APIKey:  "ghp_test_token",
			},
		}

		registry, err := buildConnectorRegistry(configs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Registry is created but empty (no enabled connectors)
		if registry == nil {
			t.Error("expected non-nil registry (even if empty)")
		}
	})

	t.Run("invalid connector name fails validation", func(t *testing.T) {
		configs := map[string]types.ConnectorConfig{
			"invalid_connector": {
				Enabled: true,
			},
		}

		registry, err := buildConnectorRegistry(configs)
		// Invalid connectors are skipped (no factory registered), so no error
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if registry == nil {
			t.Error("expected non-nil registry")
		}
	})

	// Note: Tests that require actual connector validation (with API keys)
	// are tested in the integration tests with real or mock connectors.
	// The buildConnectorRegistry function creates connectors that validate
	// themselves, so we can't easily mock that here without restructuring.
}

// TestCreateProvider tests AI provider creation
func TestCreateProvider(t *testing.T) {
	t.Run("OpenAI with API key configured", func(t *testing.T) {
		cfg := types.DefaultConfig()
		cfg.AI.Provider = types.AIProviderOpenAI
		cfg.AI.OpenAIKey = "sk-test-key"

		provider, err := createProvider(cfg)
		// For now, we expect an error since providers aren't implemented yet
		if err == nil {
			t.Fatal("expected error (provider not yet implemented)")
		}
		if provider != nil {
			t.Error("expected nil provider until implementation complete")
		}
	})

	t.Run("OpenAI with unified API key", func(t *testing.T) {
		cfg := types.DefaultConfig()
		cfg.AI.Provider = types.AIProviderOpenAI
		cfg.AI.APIKey = "sk-test-key"

		provider, err := createProvider(cfg)
		// For now, we expect an error since providers aren't implemented yet
		if err == nil {
			t.Fatal("expected error (provider not yet implemented)")
		}
		if provider != nil {
			t.Error("expected nil provider until implementation complete")
		}
	})

	t.Run("OpenAI without API key returns error", func(t *testing.T) {
		cfg := types.DefaultConfig()
		cfg.AI.Provider = types.AIProviderOpenAI
		cfg.AI.OpenAIKey = ""
		cfg.AI.APIKey = ""

		provider, err := createProvider(cfg)
		if err == nil {
			t.Fatal("expected error for missing API key")
		}
		if provider != nil {
			t.Error("expected nil provider for missing API key")
		}
	})

	t.Run("Anthropic with API key configured", func(t *testing.T) {
		cfg := types.DefaultConfig()
		cfg.AI.Provider = types.AIProviderAnthropic
		cfg.AI.AnthropicKey = "sk-ant-test-key"

		provider, err := createProvider(cfg)
		// For now, we expect an error since providers aren't implemented yet
		if err == nil {
			t.Fatal("expected error (provider not yet implemented)")
		}
		if provider != nil {
			t.Error("expected nil provider until implementation complete")
		}
	})

	t.Run("Anthropic with unified API key", func(t *testing.T) {
		cfg := types.DefaultConfig()
		cfg.AI.Provider = types.AIProviderAnthropic
		cfg.AI.APIKey = "sk-ant-test-key"

		provider, err := createProvider(cfg)
		// For now, we expect an error since providers aren't implemented yet
		if err == nil {
			t.Fatal("expected error (provider not yet implemented)")
		}
		if provider != nil {
			t.Error("expected nil provider until implementation complete")
		}
	})

	t.Run("Anthropic without API key returns error", func(t *testing.T) {
		cfg := types.DefaultConfig()
		cfg.AI.Provider = types.AIProviderAnthropic
		cfg.AI.AnthropicKey = ""
		cfg.AI.APIKey = ""

		provider, err := createProvider(cfg)
		if err == nil {
			t.Fatal("expected error for missing API key")
		}
		if provider != nil {
			t.Error("expected nil provider for missing API key")
		}
	})

	t.Run("unsupported provider returns error", func(t *testing.T) {
		cfg := types.DefaultConfig()
		cfg.AI.Provider = "unsupported"

		provider, err := createProvider(cfg)
		if err == nil {
			t.Fatal("expected error for unsupported provider")
		}
		if provider != nil {
			t.Error("expected nil provider for unsupported provider")
		}
	})
}

// TestNewEngineFromConfigIntegration tests full engine creation with mock connectors
func TestNewEngineFromConfigIntegration(t *testing.T) {
	t.Skip("Skipping until AI providers are implemented")

	t.Run("engine with github connector", func(t *testing.T) {
		cfg := types.DefaultConfig()
		cfg.AI.Enabled = true
		cfg.AI.Provider = types.AIProviderOpenAI
		cfg.AI.OpenAIKey = "sk-test-key"
		cfg.AI.Mode = types.AIModeAutonomous
		cfg.AI.Connectors = map[string]types.ConnectorConfig{
			"github": {
				Enabled:   true,
				APIKey:    "ghp_test",
				RateLimit: 60,
				Timeout:   30,
			},
		}

		engine, err := NewEngineFromConfig(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if engine == nil {
			t.Fatal("expected non-nil engine")
		}

		// Verify engine provider
		if engine.Provider() != types.AIProviderOpenAI {
			t.Errorf("expected provider %s, got %s", types.AIProviderOpenAI, engine.Provider())
		}
	})
}

// TestConnectorRegistryIntegration tests registry functionality with mock connectors
func TestConnectorRegistryIntegration(t *testing.T) {
	t.Run("collect from github mock", func(t *testing.T) {
		// Create mock connector with name in Extra config
		cfg := connectors.Config{
			Enabled: true,
			Extra: map[string]interface{}{
				"name": "github-mock",
			},
		}
		connector, err := connectors.NewMockConnector(cfg)
		if err != nil {
			t.Fatalf("failed to create mock connector: %v", err)
		}

		// Type assert to *MockConnector to access SetEvents
		mockConnector, ok := connector.(*connectors.MockConnector)
		if !ok {
			t.Fatal("expected *connectors.MockConnector")
		}

		mockConnector.SetEvents("test query", []types.EvidenceEvent{
			{
				ID:     "event-1",
				Source: "github",
				Type:   types.EventTypeCommit,
			},
		})

		// Test collection
		events, err := connector.Collect(context.Background(), "test query")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 1 {
			t.Errorf("expected 1 event, got %d", len(events))
		}
	})

	t.Run("registry routes to correct connector", func(t *testing.T) {
		// Build registry with github connector
		configs := map[string]types.ConnectorConfig{
			"github": {
				Enabled: true,
				APIKey:  "test-invalid-key",
			},
		}

		// buildConnectorRegistry validates connectors during Build()
		// With an invalid API key, we expect validation to fail
		_, err := buildConnectorRegistry(configs)
		if err == nil {
			t.Fatal("expected validation error with invalid API key, got nil")
		}

		// Verify the error message mentions validation failure
		errMsg := err.Error()
		if !strings.Contains(errMsg, "validation failed") && !strings.Contains(errMsg, "authentication") {
			t.Errorf("expected validation/authentication error, got: %v", err)
		}

		// Note: Actual collection testing requires the connector to be configured
		// with a real API key or mock. The registry tests in connectors package
		// cover this functionality more thoroughly.
	})
}
