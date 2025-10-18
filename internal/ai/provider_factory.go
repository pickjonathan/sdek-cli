package ai

import (
	"fmt"
)

// ProviderFactory is a function that creates a Provider from configuration
type ProviderFactory func(config AIConfig) (Provider, error)

// Global registry of provider factories
var providerFactories = make(map[string]ProviderFactory)

// RegisterProviderFactory registers a provider factory for a given provider name
// This is called by provider implementations in their init() functions
func RegisterProviderFactory(providerName string, factory ProviderFactory) {
	providerFactories[providerName] = factory
}

// CreateProviderFromRegistry creates a provider using the registered factory
func CreateProviderFromRegistry(providerName string, config AIConfig) (Provider, error) {
	factory, exists := providerFactories[providerName]
	if !exists {
		return nil, fmt.Errorf("no factory registered for provider: %s", providerName)
	}
	return factory(config)
}
