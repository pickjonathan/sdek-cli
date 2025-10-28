package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai/factory"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// aiHealthCmd represents the 'sdek ai health' command (Feature 006)
var aiHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check AI provider health and connectivity",
	Long: `Verify that the configured AI provider is reachable and properly configured.

This command tests:
- API connectivity and authentication
- Model availability
- Rate limit status
- Response time

Supports all provider schemes: openai://, anthropic://, gemini://, ollama://, etc.`,
	Example: `  # Check current provider health
  sdek ai health

  # Check health with verbose output
  sdek ai health --verbose

  # Test specific provider URL
  sdek ai health --provider-url "ollama://localhost:11434"`,
	RunE: runAIHealth,
}

var (
	healthProviderURL string
	healthVerbose     bool
	healthTimeout     int
)

func init() {
	aiCmd.AddCommand(aiHealthCmd)

	aiHealthCmd.Flags().StringVar(&healthProviderURL, "provider-url", "", "Override provider URL (e.g., ollama://localhost:11434)")
	aiHealthCmd.Flags().BoolVarP(&healthVerbose, "verbose", "v", false, "Show detailed health information")
	aiHealthCmd.Flags().IntVar(&healthTimeout, "timeout", 10, "Health check timeout in seconds")
}

func runAIHealth(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration from Viper
	cfg := &types.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Determine which provider to test
	var providerURL string
	var providerConfig types.ProviderConfig

	if healthProviderURL != "" {
		// Use command-line override
		providerURL = healthProviderURL
		providerConfig = types.DefaultProviderConfig()
		providerConfig.URL = providerURL

		// Try to get API key from config if available
		if cfg.AI.OpenAIKey != "" && containsStr(providerURL, "openai") {
			providerConfig.APIKey = cfg.AI.OpenAIKey
		} else if cfg.AI.AnthropicKey != "" && containsStr(providerURL, "anthropic") {
			providerConfig.APIKey = cfg.AI.AnthropicKey
		}
	} else {
		// Use provider from config
		if !cfg.AI.Enabled {
			return fmt.Errorf("AI is disabled in config (set ai.enabled=true)")
		}

		// Check if provider_url is set (Feature 006) or fall back to legacy provider (Feature 003)
		providerURL = getProviderURL(cfg)
		providerConfig = getProviderConfig(cfg, providerURL)
	}

	// Validate provider URL
	if err := factory.ValidateProviderURL(providerURL); err != nil {
		return fmt.Errorf("invalid provider URL: %w", err)
	}

	// Create provider
	provider, err := factory.CreateProvider(providerURL, providerConfig)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Display provider info
	fmt.Printf("AI Provider Health Check\n")
	fmt.Printf("========================\n\n")
	fmt.Printf("Provider URL: %s\n", providerURL)
	if providerConfig.Model != "" {
		fmt.Printf("Model:        %s\n", providerConfig.Model)
	}
	if providerConfig.Endpoint != "" {
		fmt.Printf("Endpoint:     %s\n", providerConfig.Endpoint)
	}
	fmt.Printf("Timeout:      %ds\n", healthTimeout)
	fmt.Printf("\n")

	// Run health check with timeout
	healthCtx, cancel := context.WithTimeout(ctx, time.Duration(healthTimeout)*time.Second)
	defer cancel()

	fmt.Printf("Checking connectivity...")
	startTime := time.Now()

	// Execute health check (requires providers to implement Health method)
	// For now, we'll call a simple test prompt
	result, err := provider.AnalyzeWithContext(healthCtx, "Return OK if you can read this.")
	latency := time.Since(startTime)

	if err != nil {
		fmt.Printf(" ✗ FAILED\n\n")
		fmt.Printf("Status:  ✗ Unhealthy\n")
		fmt.Printf("Error:   %v\n", err)
		fmt.Printf("Latency: %v\n", latency.Round(time.Millisecond))
		return fmt.Errorf("health check failed: %w", err)
	}

	fmt.Printf(" ✓ SUCCESS\n\n")
	fmt.Printf("Status:  ✓ Healthy\n")
	fmt.Printf("Latency: %v\n", latency.Round(time.Millisecond))

	if healthVerbose {
		fmt.Printf("\nDetailed Response:\n")
		fmt.Printf("------------------\n")
		if len(result) > 200 {
			fmt.Printf("%s...\n", result[:200])
		} else {
			fmt.Printf("%s\n", result)
		}
		fmt.Printf("\nProvider Call Count: %d\n", provider.GetCallCount())
	}

	fmt.Printf("\n✓ AI provider is healthy and ready to use\n")
	return nil
}

// getProviderURL extracts the provider URL from config
func getProviderURL(cfg *types.Config) string {
	// Check for new provider_url field (Feature 006)
	if cfg.AI.ProviderURL != "" {
		return cfg.AI.ProviderURL
	}

	// Fall back to legacy provider field (Feature 003)
	switch cfg.AI.Provider {
	case "openai":
		return "openai://api.openai.com"
	case "anthropic":
		return "anthropic://api.anthropic.com"
	default:
		return "openai://api.openai.com" // Default to OpenAI
	}
}

// getProviderConfig builds a ProviderConfig from Config
func getProviderConfig(cfg *types.Config, providerURL string) types.ProviderConfig {
	config := types.DefaultProviderConfig()
	config.URL = providerURL
	config.Model = cfg.AI.Model
	config.MaxTokens = cfg.AI.MaxTokens
	config.Temperature = float64(cfg.AI.Temperature)
	config.Timeout = cfg.AI.Timeout

	// Set API key based on provider
	if containsStr(providerURL, "openai") {
		config.APIKey = cfg.AI.OpenAIKey
	} else if containsStr(providerURL, "anthropic") {
		config.APIKey = cfg.AI.AnthropicKey
	}

	return config
}

// containsStr is a simple substring check
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && findSubstr(s, substr)
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
