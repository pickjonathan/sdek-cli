package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/sashabaranov/go-openai"
)

// OpenAIEngine implements ai.Engine using OpenAI's API
type OpenAIEngine struct {
	client  *openai.Client
	config  ai.AIConfig
	limiter *RateLimiter
}

// NewOpenAIEngine creates a new OpenAI engine
func NewOpenAIEngine(config ai.AIConfig) (*OpenAIEngine, error) {
	if config.OpenAIKey == "" {
		return nil, ai.ErrProviderAuth
	}

	client := openai.NewClient(config.OpenAIKey)

	return &OpenAIEngine{
		client:  client,
		config:  config,
		limiter: NewRateLimiter(config.RateLimit),
	}, nil
}

// AnalyzeWithRequest implements ai.Engine.AnalyzeWithRequest (Feature 002 backward compatibility)
func (e *OpenAIEngine) AnalyzeWithRequest(ctx context.Context, req *ai.AnalysisRequest) (*ai.AnalysisResponse, error) {
	// Validate request
	if err := e.validateRequest(req); err != nil {
		return nil, err
	}

	// Wait for rate limiter
	if err := e.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Set timeout from config if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(e.config.Timeout)*time.Second)
		defer cancel()
	}

	// Perform analysis with retry
	startTime := time.Now()
	response, err := e.analyzeWithRetry(ctx, req)
	latency := int(time.Since(startTime).Milliseconds())

	if err != nil {
		return nil, err
	}

	response.Latency = latency
	return response, nil
}

// Analyze implements ai.Engine.Analyze (Feature 003)
// This is a stub that returns an error - OpenAI provider needs Feature 003 implementation
func (e *OpenAIEngine) Analyze(ctx context.Context, preamble types.ContextPreamble, evidence types.EvidenceBundle) (*types.Finding, error) {
	return nil, fmt.Errorf("Feature 003 not yet implemented for OpenAI provider - use AnalyzeWithRequest for now")
}

// ProposePlan implements ai.Engine.ProposePlan (Feature 003)
// This is a stub that returns an error - OpenAI provider needs Feature 003 implementation
func (e *OpenAIEngine) ProposePlan(ctx context.Context, preamble types.ContextPreamble) (*types.EvidencePlan, error) {
	return nil, fmt.Errorf("Feature 003 not yet implemented for OpenAI provider")
}

// ExecutePlan implements ai.Engine.ExecutePlan (Feature 003)
// This is a stub that returns an error - OpenAI provider needs Feature 003 implementation
func (e *OpenAIEngine) ExecutePlan(ctx context.Context, plan *types.EvidencePlan) (*types.EvidenceBundle, error) {
	return nil, fmt.Errorf("Feature 003 not yet implemented for OpenAI provider")
}

// Provider implements ai.Engine.Provider
func (e *OpenAIEngine) Provider() string {
	return "openai"
}

// Health implements ai.Engine.Health
func (e *OpenAIEngine) Health(ctx context.Context) error {
	// Try a simple API call to verify connectivity and auth
	_, err := e.client.ListModels(ctx)
	if err != nil {
		// Check for auth errors
		if isAuthError(err) {
			return ai.ErrProviderAuth
		}
		// Check for quota errors
		if isQuotaError(err) {
			return ai.ErrProviderQuotaExceeded
		}
		return ai.ErrProviderUnavailable
	}
	return nil
}

// analyzeWithRetry performs the analysis with exponential backoff retry
func (e *OpenAIEngine) analyzeWithRetry(ctx context.Context, req *ai.AnalysisRequest) (*ai.AnalysisResponse, error) {
	var response *ai.AnalysisResponse
	var lastErr error

	operation := func() error {
		var err error
		response, err = e.performAnalysis(ctx, req)
		lastErr = err

		// Don't retry on fatal errors
		if ai.IsFatalError(err) {
			return backoff.Permanent(err)
		}

		return err
	}

	// Configure exponential backoff
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = time.Duration(e.config.Timeout) * time.Second
	bo.InitialInterval = 1 * time.Second
	bo.MaxInterval = 30 * time.Second

	// Perform retry with backoff
	err := backoff.Retry(operation, backoff.WithContext(bo, ctx))
	if err != nil {
		return nil, lastErr
	}

	return response, nil
}

// performAnalysis makes the actual API call to OpenAI
func (e *OpenAIEngine) performAnalysis(ctx context.Context, req *ai.AnalysisRequest) (*ai.AnalysisResponse, error) {
	// Build the prompt
	prompt := e.buildPrompt(req)

	// Define the function schema for structured output
	functionDef := openai.FunctionDefinition{
		Name:        "analyze_evidence",
		Description: "Analyze events for compliance control evidence",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"evidence_links": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Event IDs that support the control",
				},
				"justification": map[string]interface{}{
					"type":        "string",
					"description": "Explanation of relevance (50-500 chars)",
				},
				"confidence": map[string]interface{}{
					"type":        "integer",
					"description": "Confidence score (0-100)",
					"minimum":     0,
					"maximum":     100,
				},
				"residual_risk": map[string]interface{}{
					"type":        "string",
					"description": "Notes on gaps or concerns (0-500 chars)",
				},
			},
			"required": []string{"evidence_links", "justification", "confidence"},
		},
	}

	// Make the API call
	chatReq := openai.ChatCompletionRequest{
		Model: e.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a compliance analyst. Analyze events and map them to compliance controls.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Functions: []openai.FunctionDefinition{functionDef},
		FunctionCall: &openai.FunctionCall{
			Name: "analyze_evidence",
		},
		Temperature: float32(e.config.Temperature),
		MaxTokens:   e.config.MaxTokens,
	}

	resp, err := e.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, e.handleError(err)
	}

	// Parse the function call response
	if len(resp.Choices) == 0 {
		return nil, ai.ErrInvalidJSON
	}

	choice := resp.Choices[0]
	if choice.Message.FunctionCall == nil {
		return nil, ai.ErrInvalidJSON
	}

	// Parse the JSON arguments
	var result struct {
		EvidenceLinks []string `json:"evidence_links"`
		Justification string   `json:"justification"`
		Confidence    int      `json:"confidence"`
		ResidualRisk  string   `json:"residual_risk"`
	}

	if err := json.Unmarshal([]byte(choice.Message.FunctionCall.Arguments), &result); err != nil {
		return nil, ai.ErrInvalidJSON
	}

	// Build response
	return &ai.AnalysisResponse{
		RequestID:     req.RequestID,
		EvidenceLinks: result.EvidenceLinks,
		Justification: result.Justification,
		Confidence:    result.Confidence,
		ResidualRisk:  result.ResidualRisk,
		Provider:      "openai",
		Model:         resp.Model,
		TokensUsed:    resp.Usage.TotalTokens,
		Timestamp:     time.Now(),
		CacheHit:      false,
	}, nil
}

// buildPrompt constructs the prompt for OpenAI
func (e *OpenAIEngine) buildPrompt(req *ai.AnalysisRequest) string {
	prompt := fmt.Sprintf(`Analyze the following events for compliance with control %s (%s) in the %s framework.

Control Policy:
%s

Events to analyze:
`, req.ControlID, req.ControlName, req.Framework, req.PolicyExcerpt)

	for i, event := range req.Events {
		prompt += fmt.Sprintf("\n%d. [%s] %s - %s\n   Content: %s",
			i+1, event.Source, event.EventType, event.Description, event.Content)
	}

	prompt += "\n\nProvide your analysis including which event IDs support this control, your justification, confidence score (0-100), and any residual risks."

	return prompt
}

// validateRequest validates the analysis request
func (e *OpenAIEngine) validateRequest(req *ai.AnalysisRequest) error {
	if req == nil {
		return ai.ErrInvalidRequest
	}
	if len(req.Events) == 0 {
		return ai.ErrZeroEvents
	}
	if req.PolicyExcerpt == "" {
		return ai.ErrInvalidRequest
	}
	if req.RequestID == "" {
		return ai.ErrInvalidRequest
	}
	return nil
}

// handleError converts OpenAI errors to ai package errors
func (e *OpenAIEngine) handleError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific error types
	if isAuthError(err) {
		return ai.ErrProviderAuth
	}
	if isRateLimitError(err) {
		return ai.ErrProviderRateLimit
	}
	if isQuotaError(err) {
		return ai.ErrProviderQuotaExceeded
	}
	if isTimeoutError(err) {
		return ai.ErrProviderTimeout
	}
	if isServerError(err) {
		return ai.ErrProviderUnavailable
	}

	return fmt.Errorf("openai api error: %w", err)
}

// Error detection helpers
func isAuthError(err error) bool {
	return err != nil && (err.Error() == "401" || err.Error() == "403")
}

func isRateLimitError(err error) bool {
	return err != nil && err.Error() == "429"
}

func isQuotaError(err error) bool {
	return err != nil && (err.Error() == "429" || err.Error() == "insufficient_quota")
}

func isTimeoutError(err error) bool {
	return err != nil && err.Error() == "timeout"
}

func isServerError(err error) bool {
	return err != nil && (err.Error() == "500" || err.Error() == "502" || err.Error() == "503")
}
