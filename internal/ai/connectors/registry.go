package connectors

import (
	"context"
	"fmt"
	"sync"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Registry manages all configured connectors and routes collection requests to the appropriate connector.
type Registry struct {
	connectors map[string]Connector
	mu         sync.RWMutex
}

// NewRegistry creates a new empty connector registry.
func NewRegistry() *Registry {
	return &Registry{
		connectors: make(map[string]Connector),
	}
}

// Register adds a connector to the registry.
// If a connector with the same name already exists, it will be replaced.
func (r *Registry) Register(connector Connector) error {
	if connector == nil {
		return fmt.Errorf("connector cannot be nil")
	}

	name := connector.Name()
	if name == "" {
		return fmt.Errorf("connector name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.connectors[name] = connector
	return nil
}

// Get retrieves a connector by name.
// Returns nil if the connector is not registered.
func (r *Registry) Get(name string) Connector {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.connectors[name]
}

// Has checks if a connector with the given name is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.connectors[name]
	return ok
}

// List returns the names of all registered connectors.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.connectors))
	for name := range r.connectors {
		names = append(names, name)
	}
	return names
}

// Collect routes a collection request to the appropriate connector.
// This implements the ai.MCPConnector interface, allowing the registry to be used
// directly with the existing Engine.ExecutePlan implementation.
func (r *Registry) Collect(ctx context.Context, source string, query string) ([]types.EvidenceEvent, error) {
	connector := r.Get(source)
	if connector == nil {
		return nil, fmt.Errorf("%w: %s", ErrSourceNotFound, source)
	}

	// Delegate to the specific connector
	events, err := connector.Collect(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("connector %s failed: %w", source, err)
	}

	return events, nil
}

// ValidateAll validates all registered connectors.
// Returns a map of connector name to validation error (nil if valid).
func (r *Registry) ValidateAll(ctx context.Context) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)
	for name, connector := range r.connectors {
		results[name] = connector.Validate(ctx)
	}
	return results
}

// Close gracefully shuts down all connectors (for future use with connection pools, etc.).
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Clear the registry
	r.connectors = make(map[string]Connector)
	return nil
}

// RegistryBuilder helps construct a Registry from configuration.
type RegistryBuilder struct {
	factories map[string]Factory
	configs   map[string]Config
}

// NewRegistryBuilder creates a new builder for constructing a Registry.
func NewRegistryBuilder() *RegistryBuilder {
	return &RegistryBuilder{
		factories: make(map[string]Factory),
		configs:   make(map[string]Config),
	}
}

// RegisterFactory registers a factory function for creating connectors of a given type.
// Example: builder.RegisterFactory("github", NewGitHubConnector)
func (b *RegistryBuilder) RegisterFactory(name string, factory Factory) *RegistryBuilder {
	b.factories[name] = factory
	return b
}

// SetConfig sets the configuration for a connector type.
func (b *RegistryBuilder) SetConfig(name string, cfg Config) *RegistryBuilder {
	b.configs[name] = cfg
	return b
}

// Build constructs the Registry by creating and validating all configured connectors.
// Only enabled connectors with valid configurations will be included.
func (b *RegistryBuilder) Build(ctx context.Context) (*Registry, error) {
	registry := NewRegistry()

	for name, factory := range b.factories {
		cfg, ok := b.configs[name]
		if !ok {
			// No config provided, skip this connector
			continue
		}

		if !cfg.Enabled {
			// Connector disabled, skip
			continue
		}

		// Validate config
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("invalid config for %s: %w", name, err)
		}

		// Create connector instance
		connector, err := factory(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s connector: %w", name, err)
		}

		// Validate connector
		if err := connector.Validate(ctx); err != nil {
			return nil, fmt.Errorf("validation failed for %s: %w", name, err)
		}

		// Register connector
		if err := registry.Register(connector); err != nil {
			return nil, fmt.Errorf("failed to register %s: %w", name, err)
		}
	}

	return registry, nil
}
