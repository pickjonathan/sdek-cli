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

// Register OpenAI provider factory on package initialization
func init() {
	ai.RegisterProviderFactory("openai", func(config ai.AIConfig) (ai.Provider, error) {
		return NewOpenAIEngine(config)
	})
}

// OpenAIEngine implements ai.Engine and ai.Provider using OpenAI's API
type OpenAIEngine struct {
	client  *openai.Client
	config  ai.AIConfig
	limiter *RateLimiter

	// Testing/debugging fields
	callCount  int
	lastPrompt string
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
// Analyzes evidence against policy context and returns a Finding
func (e *OpenAIEngine) Analyze(ctx context.Context, preamble types.ContextPreamble, evidence types.EvidenceBundle) (*types.Finding, error) {
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

	// Define the function schema for structured output
	functionDef := openai.FunctionDefinition{
		Name:        "analyze_compliance_evidence",
		Description: "Analyze evidence events against policy context for compliance",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
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
			"required": []string{"title", "summary", "justification", "confidence_score", "mapped_controls"},
		},
	}

	// Make the API call
	chatReq := openai.ChatCompletionRequest{
		Model: e.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert compliance analyst. Analyze evidence against policy requirements and provide detailed findings.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Functions: []openai.FunctionDefinition{functionDef},
		FunctionCall: &openai.FunctionCall{
			Name: "analyze_compliance_evidence",
		},
		Temperature: float32(e.config.Temperature),
		MaxTokens:   e.config.MaxTokens,
	}

	resp, err := e.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse the function call response
	functionCall := resp.Choices[0].Message.FunctionCall
	if functionCall == nil {
		return nil, fmt.Errorf("no function call in response")
	}

	// Parse the JSON response
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

	if err := json.Unmarshal([]byte(functionCall.Arguments), &result); err != nil {
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
// buildContextAnalysisPrompt builds a prompt for Feature 003 context-based analysis
func (e *OpenAIEngine) buildContextAnalysisPrompt(preamble types.ContextPreamble, evidence types.EvidenceBundle) string {
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

// ============================================================================
// Provider Interface Implementation (for autonomous mode - Feature 003)
// ============================================================================

// AnalyzeWithContext implements ai.Provider.AnalyzeWithContext
// This is a simpler interface for autonomous mode that takes a pre-formatted prompt
// and returns the raw AI response as a string
func (e *OpenAIEngine) AnalyzeWithContext(ctx context.Context, prompt string) (string, error) {
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

	// Build request
	chatReq := openai.ChatCompletionRequest{
		Model: e.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert compliance analyst. Analyze evidence and provide detailed, policy-grounded findings.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: float32(e.config.Temperature),
		MaxTokens:   e.config.MaxTokens,
	}

	// Make API call with retry
	var resp openai.ChatCompletionResponse
	operation := func() error {
		var err error
		resp, err = e.client.CreateChatCompletion(ctx, chatReq)
		return err
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = time.Duration(e.config.Timeout) * time.Second

	err := backoff.Retry(operation, bo)
	if err != nil {
		return "", e.handleError(err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// GetCallCount implements ai.Provider.GetCallCount
func (e *OpenAIEngine) GetCallCount() int {
	return e.callCount
}

// GetLastPrompt implements ai.Provider.GetLastPrompt
func (e *OpenAIEngine) GetLastPrompt() string {
	return e.lastPrompt
}
