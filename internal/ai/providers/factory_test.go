package providers

import (
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/ai/factory"
)

// TestProviderFactoryRegistration verifies that all providers register themselves
func TestProviderFactoryRegistration(t *testing.T) {
	// After package init, these schemes should be registered
	expectedSchemes := []string{
		"openai",
		"anthropic",
		"gemini",
		"ollama",
	}

	registeredSchemes := factory.ListRegisteredSchemes()
	schemeMap := make(map[string]bool)
	for _, scheme := range registeredSchemes {
		schemeMap[scheme] = true
	}

	for _, expected := range expectedSchemes {
		t.Run(expected, func(t *testing.T) {
			if !schemeMap[expected] {
				t.Errorf("provider scheme %q not registered (registered: %v)", expected, registeredSchemes)
			}
		})
	}

	t.Logf("Successfully registered %d provider schemes: %v", len(registeredSchemes), registeredSchemes)
}
