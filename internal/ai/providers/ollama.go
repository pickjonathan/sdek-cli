package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/internal/ai/factory"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Register Ollama provider factory on package initialization
func init() {
	// Register with new URL scheme-based factory (Feature 006)
	factory.RegisterProviderFactory("ollama", func(config types.ProviderConfig) (ai.Provider, error) {
		return NewOllamaProvider(config)
	})
}

// OllamaProvider implements ai.Provider using Ollama's local inference API
type OllamaProvider struct {
	baseURL    string
	modelName  string
	config     types.ProviderConfig
	client     *http.Client
	limiter    *RateLimiter
	callCount  int
	lastPrompt string
}

// OllamaGenerateRequest represents the request format for Ollama's generate API
type OllamaGenerateRequest struct {
	Model  string                 `json:"model"`
	Prompt string                 `json:"prompt"`
	Stream bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// OllamaGenerateResponse represents the response format from Ollama
type OllamaGenerateResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
	Context   []int     `json:"context,omitempty"`
}

// NewOllamaProvider creates a new Ollama provider from ProviderConfig
func NewOllamaProvider(config types.ProviderConfig) (*OllamaProvider, error) {
	// Set defaults if not provided
	if config.Timeout == 0 {
		config.Timeout = 120 // Longer timeout for local models
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}
	if config.Model == "" {
		config.Model = "gemma2:2b" // Default to small, fast model
	}

	// Determine base URL
	baseURL := "http://localhost:11434"
	if config.Endpoint != "" {
		baseURL = config.Endpoint
		// Ensure URL has http:// or https:// prefix
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			baseURL = "http://" + baseURL
		}
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	return &OllamaProvider{
		baseURL:   baseURL,
		modelName: config.Model,
		config:    config,
		client:    client,
		limiter:   NewRateLimiter(0), // No rate limiting for local models
	}, nil
}

// AnalyzeWithContext implements ai.Provider.AnalyzeWithContext
func (p *OllamaProvider) AnalyzeWithContext(ctx context.Context, prompt string) (string, error) {
	// Validate prompt
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	// Wait for rate limiter
	if err := p.limiter.Wait(ctx); err != nil {
		return "", err
	}

	// Track for testing
	p.callCount++
	p.lastPrompt = prompt

	// Build request
	reqBody := OllamaGenerateRequest{
		Model:  p.modelName,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": p.config.Temperature,
			"num_predict": p.config.MaxTokens,
		},
	}

	// Add extra options from config
	for k, v := range p.config.Extra {
		reqBody.Options[k] = v
	}

	// Marshal request
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := p.client.Do(req)
	if err != nil {
		return "", p.handleError(err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp OllamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if ollamaResp.Response == "" {
		return "", fmt.Errorf("empty response from Ollama")
	}

	return ollamaResp.Response, nil
}

// Health implements ai.Provider.Health (via Engine interface)
func (p *OllamaProvider) Health(ctx context.Context) error {
	// Set timeout
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	// Check /api/tags endpoint to verify server is running
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return ai.ErrProviderUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ai.ErrProviderUnavailable
	}

	// Verify our model is available
	var tagsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return fmt.Errorf("failed to decode tags response: %w", err)
	}

	// Check if our model exists
	modelFound := false
	for _, model := range tagsResp.Models {
		if model.Name == p.modelName {
			modelFound = true
			break
		}
	}

	if !modelFound {
		return fmt.Errorf("model %q not found in Ollama (available models: %d)", p.modelName, len(tagsResp.Models))
	}

	return nil
}

// GetCallCount implements ai.Provider.GetCallCount
func (p *OllamaProvider) GetCallCount() int {
	return p.callCount
}

// GetLastPrompt implements ai.Provider.GetLastPrompt
func (p *OllamaProvider) GetLastPrompt() string {
	return p.lastPrompt
}

// handleError converts Ollama errors to ai package errors
func (p *OllamaProvider) handleError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for specific error types
	switch {
	case contains(errStr, "connection refused"), contains(errStr, "no such host"):
		return ai.ErrProviderUnavailable
	case contains(errStr, "timeout"), contains(errStr, "deadline"):
		return ai.ErrProviderTimeout
	case contains(errStr, "model not found"), contains(errStr, "404"):
		return fmt.Errorf("model %q not found: ensure it's pulled with 'ollama pull %s'", p.modelName, p.modelName)
	default:
		return fmt.Errorf("ollama error: %w", err)
	}
}
