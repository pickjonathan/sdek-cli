package factory

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// ProviderFactory is a function that creates a Provider instance from configuration.
type ProviderFactory func(config types.ProviderConfig) (ai.Provider, error)

// Global provider registry
var (
	registry   = make(map[string]ProviderFactory)
	registryMu sync.RWMutex
)

// RegisterProviderFactory registers a provider factory for the given URL scheme.
// This is typically called from provider package init() functions.
//
// Example:
//
//	func init() {
//	    factory.RegisterProviderFactory("openai", openAIFactory)
//	}
//
// Thread-safe for concurrent registration.
func RegisterProviderFactory(scheme string, factory ProviderFactory) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if factory == nil {
		panic(fmt.Sprintf("factory.RegisterProviderFactory: factory for scheme %q is nil", scheme))
	}

	if _, exists := registry[scheme]; exists {
		panic(fmt.Sprintf("factory.RegisterProviderFactory: scheme %q already registered", scheme))
	}

	registry[scheme] = factory
}

// CreateProvider creates a Provider instance from the given URL and configuration.
// The URL scheme determines which provider factory to use.
//
// Supported schemes:
//   - openai://     → OpenAI API
//   - anthropic://  → Anthropic Claude API
//   - gemini://     → Google Gemini API
//   - bedrock://    → AWS Bedrock
//   - vertexai://   → Google Vertex AI
//   - ollama://     → Ollama local inference
//   - llamacpp://   → llama.cpp local inference
//   - azopenai://   → Azure OpenAI
//
// Returns ErrUnknownScheme if the URL scheme is not registered.
// Returns ErrInvalidURL if the URL cannot be parsed.
//
// Example:
//
//	provider, err := CreateProvider("openai://api.openai.com", config)
//	if err != nil {
//	    return err
//	}
func CreateProvider(providerURL string, config types.ProviderConfig) (ai.Provider, error) {
	// Parse URL to extract scheme
	scheme, host, err := parseProviderURL(providerURL)
	if err != nil {
		return nil, err
	}

	// Get factory for scheme
	registryMu.RLock()
	factory, exists := registry[scheme]
	registryMu.RUnlock()

	if !exists {
		return nil, &ErrUnknownScheme{Scheme: scheme}
	}

	// Update config with parsed host if not already set
	if config.Endpoint == "" && host != "" {
		// Add http:// or https:// prefix based on provider type
		if host != "" && !hasScheme(host) {
			// Ollama uses HTTP (local), others use HTTPS (cloud APIs)
			if scheme == "ollama" {
				config.Endpoint = "http://" + host
			} else {
				config.Endpoint = "https://" + host
			}
		} else {
			config.Endpoint = host
		}
	}

	// Create provider using factory
	provider, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider for scheme %q: %w", scheme, err)
	}

	return provider, nil
}

// parseProviderURL extracts the scheme and host from a provider URL.
// Returns scheme, host, error.
//
// Examples:
//   - "openai://api.openai.com" → "openai", "api.openai.com", nil
//   - "ollama://localhost:11434" → "ollama", "localhost:11434", nil
//   - "bedrock://us-east-1" → "bedrock", "us-east-1", nil
func parseProviderURL(providerURL string) (string, string, error) {
	if providerURL == "" {
		return "", "", &ErrInvalidURL{URL: providerURL, Reason: "empty URL"}
	}

	parsed, err := url.Parse(providerURL)
	if err != nil {
		return "", "", &ErrInvalidURL{URL: providerURL, Reason: err.Error()}
	}

	if parsed.Scheme == "" {
		return "", "", &ErrInvalidURL{URL: providerURL, Reason: "missing scheme"}
	}

	// Extract host (handles both host:port and path-based URLs)
	host := parsed.Host
	if host == "" && parsed.Path != "" {
		// For URLs like "bedrock://us-east-1" where region is in path
		host = parsed.Path
	}

	return parsed.Scheme, host, nil
}

// ListRegisteredSchemes returns all registered provider URL schemes.
// Useful for documentation and validation.
func ListRegisteredSchemes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()

	schemes := make([]string, 0, len(registry))
	for scheme := range registry {
		schemes = append(schemes, scheme)
	}

	return schemes
}

// IsSchemeRegistered checks if a URL scheme is registered.
func IsSchemeRegistered(scheme string) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()

	_, exists := registry[scheme]
	return exists
}

// hasScheme checks if a URL string already has a scheme (http://, https://, etc.)
func hasScheme(urlStr string) bool {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return parsed.Scheme != ""
}
