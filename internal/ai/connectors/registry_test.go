package connectors

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	// Create mock connector
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.Extra = map[string]interface{}{"name": "test"}
	connector, err := NewMockConnector(cfg)
	if err != nil {
		t.Fatalf("failed to create mock connector: %v", err)
	}

	// Register connector
	if err := registry.Register(connector); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Verify it's registered
	if !registry.Has("test") {
		t.Error("expected connector to be registered")
	}

	// Verify we can retrieve it
	retrieved := registry.Get("test")
	if retrieved == nil {
		t.Error("expected to retrieve connector")
	}
	if retrieved.Name() != "test" {
		t.Errorf("expected name 'test', got '%s'", retrieved.Name())
	}
}

func TestRegistry_RegisterNil(t *testing.T) {
	registry := NewRegistry()

	err := registry.Register(nil)
	if err == nil {
		t.Error("expected error when registering nil connector")
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	registry := NewRegistry()

	connector := registry.Get("nonexistent")
	if connector != nil {
		t.Error("expected nil for nonexistent connector")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	// Register multiple connectors
	for _, name := range []string{"github", "jira", "aws"} {
		cfg := DefaultConfig()
		cfg.APIKey = "test-key"
		cfg.Extra = map[string]interface{}{"name": name}
		connector, err := NewMockConnector(cfg)
		if err != nil {
			t.Fatalf("failed to create mock connector: %v", err)
		}
		if err := registry.Register(connector); err != nil {
			t.Fatalf("Register failed: %v", err)
		}
	}

	// List connectors
	names := registry.List()
	if len(names) != 3 {
		t.Errorf("expected 3 connectors, got %d", len(names))
	}

	// Verify all names are present (order doesn't matter)
	expectedNames := map[string]bool{"github": true, "jira": true, "aws": true}
	for _, name := range names {
		if !expectedNames[name] {
			t.Errorf("unexpected connector name: %s", name)
		}
		delete(expectedNames, name)
	}
	if len(expectedNames) > 0 {
		t.Errorf("missing connector names: %v", expectedNames)
	}
}

func TestRegistry_Collect(t *testing.T) {
	registry := NewRegistry()

	// Create and register mock connector
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.Extra = map[string]interface{}{"name": "test"}
	connector, err := NewMockConnector(cfg)
	if err != nil {
		t.Fatalf("failed to create mock connector: %v", err)
	}

	// Configure mock to return specific events
	expectedEvents := []types.EvidenceEvent{
		{
			ID:        "test-1",
			Source:    "test",
			Type:      "test-type",
			Timestamp: time.Now(),
			Content:   "Test event",
		},
	}
	connector.(*MockConnector).SetEvents("test query", expectedEvents)

	if err := registry.Register(connector); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Collect events
	events, err := registry.Collect(context.Background(), "test", "test query")
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != "test-1" {
		t.Errorf("expected ID 'test-1', got '%s'", events[0].ID)
	}
}

func TestRegistry_CollectNotFound(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.Collect(context.Background(), "nonexistent", "query")
	if err == nil {
		t.Error("expected error for nonexistent connector")
	}
	// Error should wrap ErrSourceNotFound
	if err != nil && !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error message to contain 'not found', got: %v", err)
	}
}

func TestRegistry_ValidateAll(t *testing.T) {
	registry := NewRegistry()

	// Register multiple connectors
	for _, name := range []string{"connector1", "connector2"} {
		cfg := DefaultConfig()
		cfg.APIKey = "test-key"
		cfg.Extra = map[string]interface{}{"name": name}
		connector, err := NewMockConnector(cfg)
		if err != nil {
			t.Fatalf("failed to create mock connector: %v", err)
		}
		if err := registry.Register(connector); err != nil {
			t.Fatalf("Register failed: %v", err)
		}
	}

	// Validate all
	results := registry.ValidateAll(context.Background())
	if len(results) != 2 {
		t.Errorf("expected 2 validation results, got %d", len(results))
	}

	// Mock connectors always validate successfully
	for name, err := range results {
		if err != nil {
			t.Errorf("unexpected validation error for %s: %v", name, err)
		}
	}
}

func TestRegistryBuilder_Build(t *testing.T) {
	builder := NewRegistryBuilder()

	// Register factory
	builder.RegisterFactory("mock", NewMockConnector)

	// Set config
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.Enabled = true
	cfg.Extra = map[string]interface{}{"name": "test-connector"}
	builder.SetConfig("mock", cfg)

	// Build registry
	registry, err := builder.Build(context.Background())
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify connector was created and registered
	if !registry.Has("test-connector") {
		t.Error("expected connector to be registered")
	}
}

func TestRegistryBuilder_BuildDisabled(t *testing.T) {
	builder := NewRegistryBuilder()

	// Register factory
	builder.RegisterFactory("mock", NewMockConnector)

	// Set config with Enabled=false
	cfg := DefaultConfig()
	cfg.APIKey = "test-key"
	cfg.Enabled = false
	cfg.Extra = map[string]interface{}{"name": "test-connector"}
	builder.SetConfig("mock", cfg)

	// Build registry
	registry, err := builder.Build(context.Background())
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify connector was NOT registered (disabled)
	if registry.Has("test-connector") {
		t.Error("expected disabled connector to not be registered")
	}
}

func TestRegistryBuilder_BuildInvalidConfig(t *testing.T) {
	builder := NewRegistryBuilder()

	// Register factory
	builder.RegisterFactory("mock", NewMockConnector)

	// Set invalid config (no API key)
	cfg := DefaultConfig()
	cfg.APIKey = "" // Invalid
	cfg.Enabled = true
	builder.SetConfig("mock", cfg)

	// Build registry should fail
	_, err := builder.Build(context.Background())
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Enabled:   true,
				APIKey:    "test-key",
				RateLimit: 60,
				Timeout:   30,
			},
			wantErr: false,
		},
		{
			name: "missing API key when enabled",
			config: Config{
				Enabled: true,
				APIKey:  "",
			},
			wantErr: true,
		},
		{
			name: "disabled with no API key is OK",
			config: Config{
				Enabled: false,
				APIKey:  "",
			},
			wantErr: false,
		},
		{
			name: "negative rate limit",
			config: Config{
				Enabled:   true,
				APIKey:    "test-key",
				RateLimit: -1,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: Config{
				Enabled: true,
				APIKey:  "test-key",
				Timeout: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
