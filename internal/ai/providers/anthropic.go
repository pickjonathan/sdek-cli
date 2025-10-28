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
	"github.com/pickjonathan/sdek-cli/internal/ai/factory"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Register Anthropic provider factory on package initialization
func init() {
	factory.RegisterProviderFactory("anthropic", func(config types.ProviderConfig) (ai.Provider, error) {
		return NewAnthropicEngine(config)
	})
}

// AnthropicEngine implements ai.Engine and ai.Provider using Anthropic's API
type AnthropicEngine struct {
	client  *anthropic.Client
	config  ai.AIConfig
	limiter *RateLimiter

	// Testing/debugging fields
	callCount  int
	lastPrompt string
}

// NewAnthropicEngine creates a new Anthropic engine from ProviderConfig
func NewAnthropicEngine(config types.ProviderConfig) (*AnthropicEngine, error) {
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

	// Create Anthropic client with options
	options := []option.RequestOption{
		option.WithAPIKey(config.APIKey),
	}

	// Override base URL if custom endpoint provided
	if config.Endpoint != "" {
		options = append(options, option.WithBaseURL(config.Endpoint))
	}

	client := anthropic.NewClient(options...)

	// Convert ProviderConfig to legacy AIConfig for internal use
	// This maintains compatibility with existing methods
	legacyConfig := ai.AIConfig{
		Provider:     "anthropic",
		Enabled:      true,
		Model:        config.Model,
		MaxTokens:    config.MaxTokens,
		Temperature:  float32(config.Temperature),
		Timeout:      config.Timeout,
		RateLimit:    0, // Rate limiting handled by provider if configured
		OpenAIKey:    "",
		AnthropicKey: config.APIKey,
	}

	return &AnthropicEngine{
		client:  &client,
		config:  legacyConfig,
		limiter: NewRateLimiter(0), // No rate limiting by default for new system
	}, nil
}

// AnalyzeWithRequest implements ai.Engine.AnalyzeWithRequest (Feature 002 backward compatibility)
func (e *AnthropicEngine) AnalyzeWithRequest(ctx context.Context, req *ai.AnalysisRequest) (*ai.AnalysisResponse, error) {
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
// Analyzes evidence against policy context and returns a Finding
func (e *AnthropicEngine) Analyze(ctx context.Context, preamble types.ContextPreamble, evidence types.EvidenceBundle) (*types.Finding, error) {
	// Validate inputs
	if preamble.Framework == "" {
		return nil, fmt.Errorf("framework is required in context preamble")
	}
	if preamble.Section == "" {
		return nil, fmt.Errorf("section is required in context preamble")
	}
	if len(evidence.Events) == 0 {
		return nil, fmt.Errorf("no evidence events provided")
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

	// Build the analysis prompt
	prompt := e.buildContextAnalysisPrompt(preamble, evidence)

	// Define the tool schema for structured output
	toolParam := anthropic.ToolParam{
		Name:        "analyze_compliance_evidence",
		Description: anthropic.String("Analyze evidence events against policy context for compliance"),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Brief title summarizing the finding (20-100 chars)",
				},
				"summary": map[string]interface{}{
					"type":        "string",
					"description": "Summary of analysis and what was found (100-500 chars)",
				},
				"justification": map[string]interface{}{
					"type":        "string",
					"description": "Explanation of how evidence maps to policy requirements (100-1000 chars)",
				},
				"confidence_score": map[string]interface{}{
					"type":        "number",
					"description": "Confidence score (0.0-1.0)",
					"minimum":     0,
					"maximum":     1,
				},
				"residual_risk": map[string]interface{}{
					"type":        "string",
					"description": "Any gaps, concerns, or remaining risks (0-500 chars)",
				},
				"mapped_controls": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "List of control IDs that this evidence supports",
				},
				"citations": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Event IDs or sources cited in the analysis",
				},
				"severity": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"low", "medium", "high", "critical"},
					"description": "Severity level based on gaps and risks",
				},
			},
			Required: []string{"title", "summary", "justification", "confidence_score", "mapped_controls"},
		},
	}

	// Make the API call
	msg, err := e.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:       anthropic.Model(e.config.Model),
		MaxTokens:   int64(e.config.MaxTokens),
		Temperature: anthropic.Float(float64(e.config.Temperature)),
		System: []anthropic.TextBlockParam{
			{
				Text: "You are an expert compliance analyst. Analyze evidence against policy requirements and provide detailed findings.",
			},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Tools: []anthropic.ToolUnionParam{{OfTool: &toolParam}},
	})

	if err != nil {
		return nil, fmt.Errorf("Anthropic API call failed: %w", err)
	}

	// Parse the tool use response
	if len(msg.Content) == 0 {
		return nil, fmt.Errorf("no response from Anthropic")
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
		return nil, fmt.Errorf("no tool use in response")
	}

	// Parse the JSON input
	inputJSON, err := json.Marshal(toolUse.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool input: %w", err)
	}

	var result struct {
		Title           string   `json:"title"`
		Summary         string   `json:"summary"`
		Justification   string   `json:"justification"`
		ConfidenceScore float64  `json:"confidence_score"`
		ResidualRisk    string   `json:"residual_risk"`
		MappedControls  []string `json:"mapped_controls"`
		Citations       []string `json:"citations"`
		Severity        string   `json:"severity"`
	}

	if err := json.Unmarshal(inputJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Build the Finding
	now := time.Now()
	finding := &types.Finding{
		ID:              fmt.Sprintf("finding-%s-%d", preamble.Section, now.Unix()),
		ControlID:       preamble.Section,
		FrameworkID:     preamble.Framework,
		Title:           result.Title,
		Description:     result.Summary,
		Summary:         result.Summary,
		Severity:        result.Severity,
		Status:          types.StatusOpen,
		CreatedAt:       now,
		UpdatedAt:       now,
		MappedControls:  result.MappedControls,
		ConfidenceScore: result.ConfidenceScore,
		ResidualRisk:    result.ResidualRisk,
		Justification:   result.Justification,
		Citations:       result.Citations,
		ReviewRequired:  result.ConfidenceScore < 0.7, // Flag for review if confidence < 70%
		Mode:            "ai",
	}

	// Build provenance from evidence sources
	sourceCount := make(map[string]int)
	for _, event := range evidence.Events {
		sourceCount[event.Source]++
	}
	for source, count := range sourceCount {
		finding.Provenance = append(finding.Provenance, types.ProvenanceEntry{
			Source:     source,
			Query:      "", // Not tracked at event level
			EventsUsed: count,
		})
	}

	return finding, nil
}

// ProposePlan implements ai.Engine.ProposePlan (Feature 003)
// This is a stub that returns an error - Anthropic provider needs Feature 003 implementation
func (e *AnthropicEngine) ProposePlan(ctx context.Context, preamble types.ContextPreamble) (*types.EvidencePlan, error) {
	return nil, fmt.Errorf("Feature 003 not yet implemented for Anthropic provider")
}

// ExecutePlan implements ai.Engine.ExecutePlan (Feature 003)
// This is a stub that returns an error - Anthropic provider needs Feature 003 implementation
func (e *AnthropicEngine) ExecutePlan(ctx context.Context, plan *types.EvidencePlan) (*types.EvidenceBundle, error) {
	return nil, fmt.Errorf("Feature 003 not yet implemented for Anthropic provider")
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
// buildContextAnalysisPrompt builds a prompt for Feature 003 context-based analysis
func (e *AnthropicEngine) buildContextAnalysisPrompt(preamble types.ContextPreamble, evidence types.EvidenceBundle) string {
	prompt := fmt.Sprintf(`Analyze the following evidence against policy requirements for %s %s.

Framework: %s
Section: %s
Policy Excerpt:
%s
`, preamble.Framework, preamble.Section, preamble.Framework, preamble.Section, preamble.Excerpt)

	// Add related controls if present
	if len(preamble.ControlIDs) > 0 {
		prompt += fmt.Sprintf("\nRelated Controls: %v\n", preamble.ControlIDs)
	}

	prompt += "\nEvidence Events:\n"
	for i, event := range evidence.Events {
		prompt += fmt.Sprintf("\n%d. [%s/%s] %s\n   ID: %s\n   Content: %s\n",
			i+1, event.Source, event.Type, event.Timestamp.Format(time.RFC3339),
			event.ID, event.Content)

		// Add relevant metadata
		if len(event.Metadata) > 0 {
			prompt += "   Metadata: "
			for k, v := range event.Metadata {
				prompt += fmt.Sprintf("%s=%v ", k, v)
			}
			prompt += "\n"
		}
	}

	prompt += `

Provide a comprehensive compliance analysis including:
1. Title: Brief summary of findings
2. Summary: What evidence was found and how it relates to the policy
3. Justification: Detailed explanation of compliance status
4. Confidence Score: 0.0-1.0 based on evidence quality and coverage
5. Residual Risk: Any gaps, concerns, or remaining risks
6. Mapped Controls: Control IDs that this evidence supports
7. Citations: Specific event IDs referenced in your analysis
8. Severity: Overall risk level (low, medium, high, critical)
`

	return prompt
}

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

// ============================================================================
// Provider Interface Implementation (for autonomous mode - Feature 003)
// ============================================================================

// AnalyzeWithContext implements ai.Provider.AnalyzeWithContext
// This is a simpler interface for autonomous mode that takes a pre-formatted prompt
// and returns the raw AI response as a string
func (e *AnthropicEngine) AnalyzeWithContext(ctx context.Context, prompt string) (string, error) {
	// Validate prompt
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	// Wait for rate limiter
	if err := e.limiter.Wait(ctx); err != nil {
		return "", err
	}

	// Set timeout from config if not already set
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(e.config.Timeout)*time.Second)
		defer cancel()
	}

	// Track for testing
	e.callCount++
	e.lastPrompt = prompt

	// Make API call with retry
	var resp *anthropic.Message
	operation := func() error {
		var err error
		resp, err = e.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:       anthropic.Model(e.config.Model),
			MaxTokens:   int64(e.config.MaxTokens),
			Temperature: anthropic.Float(float64(e.config.Temperature)),
			System: []anthropic.TextBlockParam{
				{
					Text: "You are an expert compliance analyst. Analyze evidence and provide detailed, policy-grounded findings.",
				},
			},
			Messages: []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
			},
		})
		return err
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = time.Duration(e.config.Timeout) * time.Second

	err := backoff.Retry(operation, bo)
	if err != nil {
		return "", e.handleError(err)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("no content in Anthropic response")
	}

	// Extract text from the first content block
	if block := resp.Content[0].AsAny(); block != nil {
		if textBlock, ok := block.(anthropic.TextBlock); ok {
			return textBlock.Text, nil
		}
	}

	return "", fmt.Errorf("unexpected content type in Anthropic response")
}

// GetCallCount implements ai.Provider.GetCallCount
func (e *AnthropicEngine) GetCallCount() int {
	return e.callCount
}

// GetLastPrompt implements ai.Provider.GetLastPrompt
func (e *AnthropicEngine) GetLastPrompt() string {
	return e.lastPrompt
}
