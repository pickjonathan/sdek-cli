package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/cenkalti/backoff/v4"
	"github.com/pickjonathan/sdek-cli/internal/ai"
)

// AnthropicEngine implements ai.Engine using Anthropic's API
type AnthropicEngine struct {
	client  *anthropic.Client
	config  ai.AIConfig
	limiter *RateLimiter
}

// NewAnthropicEngine creates a new Anthropic engine
func NewAnthropicEngine(config ai.AIConfig) (*AnthropicEngine, error) {
	if config.AnthropicKey == "" {
		return nil, ai.ErrProviderAuth
	}

	client := anthropic.NewClient(
		option.WithAPIKey(config.AnthropicKey),
	)

	return &AnthropicEngine{
		client:  &client,
		config:  config,
		limiter: NewRateLimiter(config.RateLimit),
	}, nil
}

// Analyze implements ai.Engine.Analyze
func (e *AnthropicEngine) Analyze(ctx context.Context, req *ai.AnalysisRequest) (*ai.AnalysisResponse, error) {
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

// Provider implements ai.Engine.Provider
func (e *AnthropicEngine) Provider() string {
	return "anthropic"
}

// Health implements ai.Engine.Health
func (e *AnthropicEngine) Health(ctx context.Context) error {
	// Try a simple API call to verify connectivity and auth
	// Anthropic doesn't have a list models endpoint, so we'll do a minimal completion
	_, err := e.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: 10,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("test")),
		},
	})

	if err != nil {
		// Check for auth errors
		if isAnthropicAuthError(err) {
			return ai.ErrProviderAuth
		}
		// Check for quota errors
		if isAnthropicQuotaError(err) {
			return ai.ErrProviderQuotaExceeded
		}
		return ai.ErrProviderUnavailable
	}
	return nil
}

// analyzeWithRetry performs the analysis with exponential backoff retry
func (e *AnthropicEngine) analyzeWithRetry(ctx context.Context, req *ai.AnalysisRequest) (*ai.AnalysisResponse, error) {
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

// performAnalysis makes the actual API call to Anthropic
func (e *AnthropicEngine) performAnalysis(ctx context.Context, req *ai.AnalysisRequest) (*ai.AnalysisResponse, error) {
	// Build the prompt
	systemPrompt, userPrompt := e.buildPrompt(req)

	// Define the tool schema for structured output
	toolParam := anthropic.ToolParam{
		Name:        "analyze_evidence",
		Description: anthropic.String("Analyze events for compliance control evidence"),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
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
			Required: []string{"evidence_links", "justification", "confidence"},
		},
	}

	// Make the API call
	msg, err := e.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:       anthropic.Model(e.config.Model),
		MaxTokens:   int64(e.config.MaxTokens),
		Temperature: anthropic.Float(float64(e.config.Temperature)),
		System: []anthropic.TextBlockParam{
			{
				Text: systemPrompt,
			},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
		Tools: []anthropic.ToolUnionParam{{OfTool: &toolParam}},
	})

	if err != nil {
		return nil, e.handleError(err)
	}

	// Parse the tool use response
	if len(msg.Content) == 0 {
		return nil, ai.ErrInvalidJSON
	}

	var toolUse *anthropic.ToolUseBlock
	for _, content := range msg.Content {
		block := content.AsAny()
		if block, ok := block.(anthropic.ToolUseBlock); ok {
			toolUse = &block
			break
		}
	}

	if toolUse == nil {
		return nil, ai.ErrInvalidJSON
	}

	// Parse the JSON input
	var result struct {
		EvidenceLinks []string `json:"evidence_links"`
		Justification string   `json:"justification"`
		Confidence    int      `json:"confidence"`
		ResidualRisk  string   `json:"residual_risk"`
	}

	inputJSON, err := json.Marshal(toolUse.Input)
	if err != nil {
		return nil, ai.ErrInvalidJSON
	}

	if err := json.Unmarshal(inputJSON, &result); err != nil {
		return nil, ai.ErrInvalidJSON
	}

	// Build response
	return &ai.AnalysisResponse{
		RequestID:     req.RequestID,
		EvidenceLinks: result.EvidenceLinks,
		Justification: result.Justification,
		Confidence:    result.Confidence,
		ResidualRisk:  result.ResidualRisk,
		Provider:      "anthropic",
		Model:         string(msg.Model),
		TokensUsed:    int(msg.Usage.InputTokens + msg.Usage.OutputTokens),
		Timestamp:     time.Now(),
		CacheHit:      false,
	}, nil
}

// buildPrompt constructs the prompt for Anthropic
func (e *AnthropicEngine) buildPrompt(req *ai.AnalysisRequest) (system, user string) {
	system = "You are a compliance analyst. Analyze events and map them to compliance controls."

	user = fmt.Sprintf(`Analyze the following events for compliance with control %s (%s) in the %s framework.

Control Policy:
%s

Events to analyze:
`, req.ControlID, req.ControlName, req.Framework, req.PolicyExcerpt)

	for i, event := range req.Events {
		user += fmt.Sprintf("\n%d. [%s] %s - %s\n   Content: %s",
			i+1, event.Source, event.EventType, event.Description, event.Content)
	}

	user += "\n\nProvide your analysis including which event IDs support this control, your justification, confidence score (0-100), and any residual risks."

	return system, user
}

// validateRequest validates the analysis request
func (e *AnthropicEngine) validateRequest(req *ai.AnalysisRequest) error {
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

// handleError converts Anthropic errors to ai package errors
func (e *AnthropicEngine) handleError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific error types
	if isAnthropicAuthError(err) {
		return ai.ErrProviderAuth
	}
	if isAnthropicRateLimitError(err) {
		return ai.ErrProviderRateLimit
	}
	if isAnthropicQuotaError(err) {
		return ai.ErrProviderQuotaExceeded
	}
	if isAnthropicTimeoutError(err) {
		return ai.ErrProviderTimeout
	}
	if isAnthropicServerError(err) {
		return ai.ErrProviderUnavailable
	}

	return fmt.Errorf("anthropic api error: %w", err)
}

// Error detection helpers
func isAnthropicAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errStr == "401" || errStr == "403" || errStr == "authentication_error" || errStr == "permission_error"
}

func isAnthropicRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errStr == "429" || errStr == "rate_limit_error"
}

func isAnthropicQuotaError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errStr == "429" || errStr == "insufficient_quota" || errStr == "overloaded_error"
}

func isAnthropicTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "timeout"
}

func isAnthropicServerError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errStr == "500" || errStr == "502" || errStr == "503" || errStr == "api_error" || errStr == "internal_server_error"
}
