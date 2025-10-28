package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/internal/ai/factory"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"google.golang.org/api/option"
)

// Register Gemini provider factory on package initialization
func init() {
	// Register with new URL scheme-based factory (Feature 006)
	factory.RegisterProviderFactory("gemini", func(config types.ProviderConfig) (ai.Provider, error) {
		return NewGeminiProvider(config)
	})
}

// GeminiProvider implements ai.Provider using Google's Gemini API
type GeminiProvider struct {
	client     *genai.Client
	config     types.ProviderConfig
	modelName  string
	limiter    *RateLimiter
	callCount  int
	lastPrompt string
}

// NewGeminiProvider creates a new Gemini provider from ProviderConfig
func NewGeminiProvider(config types.ProviderConfig) (*GeminiProvider, error) {
	// Validate API key
	if config.APIKey == "" {
		return nil, ai.ErrProviderAuth
	}

	// Set defaults if not provided
	if config.Timeout == 0 {
		config.Timeout = 60
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}
	if config.Model == "" {
		config.Model = "gemini-2.0-flash-exp" // Default to latest flash model
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiProvider{
		client:    client,
		config:    config,
		modelName: config.Model,
		limiter:   NewRateLimiter(0), // No rate limiting by default
	}, nil
}

// AnalyzeWithContext implements ai.Provider.AnalyzeWithContext
func (p *GeminiProvider) AnalyzeWithContext(ctx context.Context, prompt string) (string, error) {
	// Validate prompt
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	// Wait for rate limiter
	if err := p.limiter.Wait(ctx); err != nil {
		return "", err
	}

	// Set timeout from config if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(p.config.Timeout)*time.Second)
		defer cancel()
	}

	// Track for testing
	p.callCount++
	p.lastPrompt = prompt

	// Get model
	model := p.client.GenerativeModel(p.modelName)

	// Configure model parameters
	model.SetTemperature(float32(p.config.Temperature))
	model.SetMaxOutputTokens(int32(p.config.MaxTokens))

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", p.handleError(err)
	}

	// Extract text from response
	if resp == nil || len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	// Concatenate all text parts
	var result string
	for _, part := range candidate.Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result += string(text)
		}
	}

	if result == "" {
		return "", fmt.Errorf("no text content in Gemini response")
	}

	return result, nil
}

// Health implements ai.Provider.Health (via Engine interface)
func (p *GeminiProvider) Health(ctx context.Context) error {
	// Set timeout
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	// Try a simple API call to verify connectivity
	model := p.client.GenerativeModel(p.modelName)
	resp, err := model.GenerateContent(ctx, genai.Text("Hello"))
	if err != nil {
		return p.handleError(err)
	}

	if resp == nil {
		return ai.ErrProviderUnavailable
	}

	return nil
}

// GetCallCount implements ai.Provider.GetCallCount
func (p *GeminiProvider) GetCallCount() int {
	return p.callCount
}

// GetLastPrompt implements ai.Provider.GetLastPrompt
func (p *GeminiProvider) GetLastPrompt() string {
	return p.lastPrompt
}

// Close cleans up the provider resources
func (p *GeminiProvider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// handleError converts Gemini errors to ai package errors
func (p *GeminiProvider) handleError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for specific error types
	switch {
	case contains(errStr, "API key"):
		return ai.ErrProviderAuth
	case contains(errStr, "quota"):
		return ai.ErrProviderQuotaExceeded
	case contains(errStr, "rate limit"):
		return ai.ErrProviderRateLimit
	case contains(errStr, "timeout"), contains(errStr, "deadline"):
		return ai.ErrProviderTimeout
	case contains(errStr, "unavailable"), contains(errStr, "503"):
		return ai.ErrProviderUnavailable
	default:
		return fmt.Errorf("gemini api error: %w", err)
	}
}

// contains is a helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
