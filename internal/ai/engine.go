package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Engine is the core abstraction for AI provider integrations.
// Implementations must support OpenAI and Anthropic initially.
// All implementations MUST be safe for concurrent use.
type Engine interface {
	// Analyze (legacy) sends an analysis request to the AI provider and returns
	// the structured response. Returns error if provider fails, times out,
	// or returns invalid JSON.
	//
	// Context cancellation triggers immediate abort (no retry).
	// Timeout specified in ctx or falls back to AIConfig.Timeout.
	// DEPRECATED: Use AnalyzeWithRequest for Feature 002 behavior
	AnalyzeWithRequest(ctx context.Context, req *AnalysisRequest) (*AnalysisResponse, error)

	// Analyze performs AI analysis with context injection (Feature 003)
	// This is the primary analysis method for Feature 003
	Analyze(ctx context.Context, preamble types.ContextPreamble, evidence types.EvidenceBundle) (*types.Finding, error)

	// ProposePlan generates an evidence collection plan for autonomous mode (Feature 003)
	// Plans are NOT cached (always fresh). Auto-approve policies are applied.
	// Returns ErrNoPlanItems if provider returns empty plan.
	// Returns ErrBudgetExceeded if plan exceeds configured limits.
	ProposePlan(ctx context.Context, preamble types.ContextPreamble) (*types.EvidencePlan, error)

	// ExecutePlan executes an approved evidence collection plan via MCP connectors (Feature 003)
	// Filters to approved/auto-approved items only. Executes in parallel with concurrency limits.
	// Returns ErrPlanNotApproved if plan status is not "approved".
	// Returns ErrNoApprovedItems if no items are approved for execution.
	// Returns ErrMCPConnectorFailed if all connector calls fail.
	ExecutePlan(ctx context.Context, plan *types.EvidencePlan) (*types.EvidenceBundle, error)

	// Provider returns the provider identifier ("openai" | "anthropic" | "mock").
	Provider() string

	// Health checks if the provider is reachable and configured correctly.
	// Returns error if API key invalid, quota exceeded, or network unreachable.
	Health(ctx context.Context) error
}

// Provider is the interface for AI provider implementations (Feature 003)
// This wraps the lower-level Engine interface with context injection support
type Provider interface {
	// AnalyzeWithContext performs AI analysis with context preamble
	AnalyzeWithContext(ctx context.Context, prompt string) (string, error)

	// GetCallCount returns the number of calls made (for testing)
	GetCallCount() int

	// GetLastPrompt returns the last prompt sent (for testing)
	GetLastPrompt() string
}

// MCPConnector is the interface for MCP (Model Context Protocol) connectors
// that fetch evidence from external sources (GitHub, Jira, AWS, etc.)
type MCPConnector interface {
	// Collect fetches events from a source using the given query
	Collect(ctx context.Context, source string, query string) ([]types.EvidenceEvent, error)
}

// engineImpl wraps a Provider with caching and redaction
type engineImpl struct {
	config             *types.Config
	provider           Provider
	cache              *Cache
	redactor           Redactor
	autoApproveMatcher AutoApproveMatcher
	connector          MCPConnector // For ExecutePlan
}

// NewEngine creates a new Engine instance with the given config and provider
func NewEngine(cfg *types.Config, provider Provider) Engine {
	return NewEngineWithConnector(cfg, provider, nil)
}

// NewEngineWithConnector creates a new Engine instance with a custom MCP connector
func NewEngineWithConnector(cfg *types.Config, provider Provider, connector MCPConnector) Engine {
	// Initialize cache
	cache, err := NewCache(cfg.AI.CacheDir)
	if err != nil {
		// If cache creation fails, create in-memory cache
		cache, _ = NewCache("")
	}

	// Initialize redactor
	redactor := NewRedactor(cfg)

	// Initialize auto-approve matcher
	autoApproveMatcher := NewAutoApproveMatcher(cfg)

	return &engineImpl{
		config:             cfg,
		provider:           provider,
		cache:              cache,
		redactor:           redactor,
		autoApproveMatcher: autoApproveMatcher,
		connector:          connector,
	}
}

// AnalyzeWithRequest implements the Engine interface (legacy method from Feature 002)
// This maintains backward compatibility by adapting to the new Feature 003 interface
func (e *engineImpl) AnalyzeWithRequest(ctx context.Context, req *AnalysisRequest) (*AnalysisResponse, error) {
	if req == nil {
		return nil, ErrInvalidRequest
	}

	// Convert AnalysisRequest to Feature 003 format
	// Build ContextPreamble from request
	preamble, err := types.NewContextPreamble(
		req.Framework,
		"", // Version not available in old format
		req.ControlID,
		req.PolicyExcerpt,
		nil, // Related sections not available
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create preamble: %w", err)
	}

	// Convert AnalysisEvents to EvidenceEvents
	events := make([]types.EvidenceEvent, len(req.Events))
	for i, evt := range req.Events {
		events[i] = types.EvidenceEvent{
			ID:        evt.EventID,
			Source:    evt.Source,
			Type:      evt.EventType,
			Timestamp: evt.Timestamp,
			Content:   evt.Content,
			Metadata: map[string]interface{}{
				"description": evt.Description,
			},
		}
	}

	evidence := types.EvidenceBundle{
		Events: events,
	}

	// Call new Feature 003 Analyze method
	finding, err := e.Analyze(ctx, *preamble, evidence)
	if err != nil {
		return nil, err
	}

	// Convert Finding back to AnalysisResponse
	response := &AnalysisResponse{
		RequestID:     req.RequestID,
		EvidenceLinks: finding.Citations,
		Justification: finding.Justification,
		Confidence:    int(finding.ConfidenceScore * 100), // Convert 0-1 to 0-100
		ResidualRisk:  string(finding.ResidualRisk),
		Provider:      e.config.AI.Provider,
		Model:         e.config.AI.Model,
		TokensUsed:    0, // Not tracked in Feature 003
		Latency:       0, // Not tracked in Feature 003
		Timestamp:     time.Now(),
		CacheHit:      false, // Cache hit detection would need to be added
	}

	return response, nil
}

// Provider returns the provider name
func (e *engineImpl) Provider() string {
	return e.config.AI.Provider
}

// Health checks provider health
func (e *engineImpl) Health(ctx context.Context) error {
	// Basic health check - try a simple prompt
	_, err := e.provider.AnalyzeWithContext(ctx, "Health check")
	return err
}

// Analyze performs AI analysis with context injection (Feature 003)
func (e *engineImpl) Analyze(ctx context.Context, preamble types.ContextPreamble, evidence types.EvidenceBundle) (*types.Finding, error) {
	// Validate preamble
	if err := preamble.Validate(); err != nil {
		return nil, fmt.Errorf("invalid preamble: %w", err)
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Handle empty evidence - return low confidence
	if len(evidence.Events) == 0 {
		return e.createLowConfidenceFinding(preamble, "No evidence provided for analysis"), nil
	}

	// Redact evidence
	redactedEvents := make([]types.EvidenceEvent, len(evidence.Events))
	for i, event := range evidence.Events {
		redacted, _, err := e.redactor.Redact(event.Content)
		if err != nil {
			return nil, fmt.Errorf("redaction failed: %w", err)
		}
		redactedEvents[i] = event
		redactedEvents[i].Content = redacted
	}
	redactedEvidence := types.EvidenceBundle{Events: redactedEvents}

	// Compute cache key
	cacheKey := e.computeCacheKey(preamble, redactedEvidence)

	// Check cache (unless NoCache is set)
	if e.config.AI.CacheDir != "" && !e.config.AI.NoCache {
		if cached, err := e.cache.Get(cacheKey); err == nil && cached != nil {
			// Convert cached response to Finding
			return e.responseToCachedFinding(cached, preamble), nil
		}
	} // Build prompt with context injection
	prompt := e.buildPromptWithContext(preamble, redactedEvidence)

	// Call AI provider
	responseText, err := e.provider.AnalyzeWithContext(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse response to Finding
	finding, err := e.parseResponseToFinding(responseText, preamble, evidence)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Set mode to "ai"
	finding.Mode = "ai"

	// Set review flag based on confidence threshold
	threshold := preamble.Rubrics.ConfidenceThreshold
	if finding.ConfidenceScore < threshold {
		finding.ReviewRequired = true
	}

	// Cache result
	if e.config.AI.CacheDir != "" && !e.config.AI.NoCache {
		cached := e.findingToCachedResult(cacheKey, finding)
		_ = e.cache.Set(cacheKey, cached) // Ignore cache write errors
	}

	return finding, nil
}

// ProposePlan generates an evidence collection plan for autonomous mode (Feature 003)
func (e *engineImpl) ProposePlan(ctx context.Context, preamble types.ContextPreamble) (*types.EvidencePlan, error) {
	// Validate preamble
	if err := preamble.Validate(); err != nil {
		return nil, fmt.Errorf("invalid preamble: %w", err)
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Build prompt for plan generation
	prompt := e.buildPlanPrompt(preamble)

	// Call AI provider to generate plan (no caching for plans - always fresh)
	responseText, err := e.provider.AnalyzeWithContext(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Parse response to plan items
	items, err := e.parsePlanResponse(responseText)
	if err != nil {
		return nil, err
	}

	// Check if we got any items
	if len(items) == 0 {
		return nil, ErrNoPlanItems
	}

	// Sort items deterministically (by source then query)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Source != items[j].Source {
			return items[i].Source < items[j].Source
		}
		return items[i].Query < items[j].Query
	})

	// Apply auto-approve matcher
	for i := range items {
		if e.autoApproveMatcher.Matches(items[i].Source, items[i].Query) {
			items[i].AutoApproved = true
			items[i].ApprovalStatus = types.ApprovalAutoApproved
		} else {
			items[i].ApprovalStatus = types.ApprovalPending
		}
	}

	// Validate against budgets (0 = no limit)
	budgets := e.config.AI.Budgets

	// Check MaxSources (total items)
	if budgets.MaxSources > 0 && len(items) > budgets.MaxSources {
		return nil, ErrBudgetExceeded
	}

	// Count API calls (each item = 1 call)
	if budgets.MaxAPICalls > 0 && len(items) > budgets.MaxAPICalls {
		return nil, ErrBudgetExceeded
	}

	// Count unique sources for EstimatedSources
	sourceSet := make(map[string]bool)
	for _, item := range items {
		sourceSet[item.Source] = true
	}

	// Create plan
	plan := &types.EvidencePlan{
		ID:               fmt.Sprintf("plan-%d", time.Now().Unix()),
		Framework:        preamble.Framework,
		Section:          preamble.Section,
		Items:            items,
		EstimatedSources: len(sourceSet),
		EstimatedCalls:   len(items),
		EstimatedTokens:  0, // TODO: Calculate token estimate
		Status:           types.PlanPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	return plan, nil
}

// ExecutePlan executes an approved evidence collection plan via MCP connectors (Feature 003)
func (e *engineImpl) ExecutePlan(ctx context.Context, plan *types.EvidencePlan) (*types.EvidenceBundle, error) {
	// Validate plan is approved
	if plan.Status != types.PlanApproved {
		return nil, ErrPlanNotApproved
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Filter to approved/auto-approved items only
	approvedItems := make([]*types.PlanItem, 0)
	for i := range plan.Items {
		item := &plan.Items[i]
		if item.ApprovalStatus == types.ApprovalApproved || item.ApprovalStatus == types.ApprovalAutoApproved {
			approvedItems = append(approvedItems, item)
		}
	}

	// Check if we have any approved items
	// Special case: if ALL items have explicit statuses (pending/denied/approved/auto_approved)
	// and none are approved, this should be an error (user made explicit choice)
	// But if we have a mix, just return empty bundle
	if len(approvedItems) == 0 {
		// Count how many items have explicit rejection (pending or denied)
		rejectedCount := 0
		for i := range plan.Items {
			if plan.Items[i].ApprovalStatus == types.ApprovalPending || plan.Items[i].ApprovalStatus == types.ApprovalDenied {
				rejectedCount++
			}
		}

		// If all items were explicitly rejected, that's an error
		if rejectedCount == len(plan.Items) && len(plan.Items) > 1 {
			return nil, ErrNoApprovedItems
		}

		// Otherwise return empty bundle
		return &types.EvidenceBundle{Events: []types.EvidenceEvent{}}, nil
	}

	// If no connector is available, return error
	if e.connector == nil {
		return nil, fmt.Errorf("no MCP connector configured")
	}

	// Execute items in parallel with goroutines
	type result struct {
		item   *types.PlanItem
		events []types.EvidenceEvent
		err    error
	}

	results := make(chan result, len(approvedItems))

	// Launch goroutines for each approved item
	for _, item := range approvedItems {
		go func(item *types.PlanItem) {
			// Set status to running
			item.ExecutionStatus = types.ExecRunning

			// Call MCP connector
			events, err := e.connector.Collect(ctx, item.Source, item.Query)

			if err != nil {
				// Handle error
				item.ExecutionStatus = types.ExecFailed
				item.Error = err.Error()
				results <- result{item: item, events: nil, err: err}
				return
			}

			// Success
			item.ExecutionStatus = types.ExecComplete
			item.EventsCollected = len(events)
			results <- result{item: item, events: events, err: nil}
		}(item)
	}

	// Collect results
	allEvents := make([]types.EvidenceEvent, 0)
	successCount := 0

	for i := 0; i < len(approvedItems); i++ {
		res := <-results
		if res.err == nil {
			allEvents = append(allEvents, res.events...)
			successCount++
		}
	}

	// If all connectors failed, return error
	if successCount == 0 {
		return nil, ErrMCPConnectorFailed
	}

	// Return bundle with all collected events
	bundle := &types.EvidenceBundle{
		Events: allEvents,
	}

	return bundle, nil
}

// buildPlanPrompt creates a prompt for evidence plan generation
func (e *engineImpl) buildPlanPrompt(preamble types.ContextPreamble) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are creating an evidence collection plan for compliance with %s %s.\n\n", preamble.Framework, preamble.Section))
	sb.WriteString("Control Excerpt:\n")
	sb.WriteString(preamble.Excerpt)
	sb.WriteString("\n\n")

	sb.WriteString("Generate a list of evidence sources to query. For each source, provide:\n")
	sb.WriteString("- source: System name (github, jira, aws, slack, etc.)\n")
	sb.WriteString("- query: Search query or filter expression\n")
	sb.WriteString("- signal_strength: Relevance score (0.0-1.0)\n")
	sb.WriteString("- rationale: Why this source/query is relevant\n\n")

	sb.WriteString("Return your response as a JSON array of plan items:\n")
	sb.WriteString(`[{"source": "github", "query": "type:pr label:security", "signal_strength": 0.9, "rationale": "Security PRs show access control implementations"}]`)

	return sb.String()
}

// parsePlanResponse parses AI response into plan items
func (e *engineImpl) parsePlanResponse(responseText string) ([]types.PlanItem, error) {
	// Try to extract JSON array from the response
	jsonStart := strings.Index(responseText, "[")
	jsonEnd := strings.LastIndex(responseText, "]")

	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no JSON array found in response")
	}

	jsonStr := responseText[jsonStart : jsonEnd+1]

	// Parse JSON response
	var items []struct {
		Source         string  `json:"source"`
		Query          string  `json:"query"`
		SignalStrength float64 `json:"signal_strength"`
		Rationale      string  `json:"rationale"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &items); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	// Convert to PlanItem
	planItems := make([]types.PlanItem, 0, len(items))
	for _, item := range items {
		// Validate required fields
		if item.Source == "" || item.Query == "" {
			continue // Skip invalid items
		}

		planItems = append(planItems, types.PlanItem{
			Source:          item.Source,
			Query:           item.Query,
			SignalStrength:  item.SignalStrength,
			Rationale:       item.Rationale,
			ApprovalStatus:  types.ApprovalPending,
			ExecutionStatus: types.ExecPending,
		})
	}

	return planItems, nil
}

// computeCacheKey generates a deterministic cache key from preamble and evidence
func (e *engineImpl) computeCacheKey(preamble types.ContextPreamble, evidence types.EvidenceBundle) string {
	h := sha256.New()

	// Include framework, section, and excerpt
	h.Write([]byte(preamble.Framework))
	h.Write([]byte(preamble.Version))
	h.Write([]byte(preamble.Section))
	h.Write([]byte(preamble.Excerpt))

	// Sort and include evidence events for determinism
	events := make([]types.EvidenceEvent, len(evidence.Events))
	copy(events, evidence.Events)
	sort.Slice(events, func(i, j int) bool {
		return events[i].ID < events[j].ID
	})

	for _, event := range events {
		h.Write([]byte(event.ID))
		h.Write([]byte(event.Content))
	}

	return hex.EncodeToString(h.Sum(nil))
}

// buildPromptWithContext creates a prompt with framework context injection
func (e *engineImpl) buildPromptWithContext(preamble types.ContextPreamble, evidence types.EvidenceBundle) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are analyzing evidence for compliance with %s %s.\n\n", preamble.Framework, preamble.Section))
	sb.WriteString("Control Excerpt:\n")
	sb.WriteString(preamble.Excerpt)
	sb.WriteString("\n\n")

	sb.WriteString("Evidence (redacted):\n")
	for i, event := range evidence.Events {
		sb.WriteString(fmt.Sprintf("%d. [%s/%s] %s\n", i+1, event.Source, event.Type, event.Content))
	}
	sb.WriteString("\n")

	sb.WriteString("Provide your analysis in JSON format with the following fields:\n")
	sb.WriteString("- summary: Brief summary of findings (50-200 words)\n")
	sb.WriteString("- mapped_controls: List of control IDs that apply\n")
	sb.WriteString("- confidence_score: Confidence level (0.0-1.0)\n")
	sb.WriteString("- residual_risk: Risk level (low/medium/high)\n")
	sb.WriteString("- justification: Detailed explanation (100-500 words)\n")
	sb.WriteString("- citations: List of event IDs that support the finding\n")

	return sb.String()
}

// parseResponseToFinding converts AI response text to a Finding
func (e *engineImpl) parseResponseToFinding(responseText string, preamble types.ContextPreamble, evidence types.EvidenceBundle) (*types.Finding, error) {
	// Try to extract JSON from the response
	jsonStart := strings.Index(responseText, "{")
	jsonEnd := strings.LastIndex(responseText, "}")

	if jsonStart == -1 || jsonEnd == -1 {
		// No JSON found - create a basic finding
		return e.createBasicFinding(responseText, preamble, evidence), nil
	}

	jsonStr := responseText[jsonStart : jsonEnd+1]

	// Parse JSON response
	var resp struct {
		Summary         string   `json:"summary"`
		MappedControls  []string `json:"mapped_controls"`
		ConfidenceScore float64  `json:"confidence_score"`
		ResidualRisk    string   `json:"residual_risk"`
		Justification   string   `json:"justification"`
		Citations       []string `json:"citations"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		// JSON parsing failed - create basic finding
		return e.createBasicFinding(responseText, preamble, evidence), nil
	}

	// Create Finding from parsed response
	finding := &types.Finding{
		ID:              fmt.Sprintf("finding-%d", time.Now().Unix()),
		ControlID:       preamble.Section,
		FrameworkID:     preamble.Framework,
		Title:           fmt.Sprintf("%s %s Analysis", preamble.Framework, preamble.Section),
		Summary:         resp.Summary,
		MappedControls:  resp.MappedControls,
		ConfidenceScore: resp.ConfidenceScore,
		ResidualRisk:    resp.ResidualRisk,
		Justification:   resp.Justification,
		Citations:       resp.Citations,
		Severity:        e.riskToSeverity(resp.ResidualRisk),
		Status:          types.StatusOpen,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return finding, nil
}

// createBasicFinding creates a basic finding when AI response parsing fails
func (e *engineImpl) createBasicFinding(responseText string, preamble types.ContextPreamble, evidence types.EvidenceBundle) *types.Finding {
	// Extract event IDs for citations
	citations := make([]string, 0, len(evidence.Events))
	for _, event := range evidence.Events {
		citations = append(citations, event.ID)
	}

	return &types.Finding{
		ID:              fmt.Sprintf("finding-%d", time.Now().Unix()),
		ControlID:       preamble.Section,
		FrameworkID:     preamble.Framework,
		Title:           fmt.Sprintf("%s %s Analysis", preamble.Framework, preamble.Section),
		Summary:         responseText,
		MappedControls:  []string{preamble.Section},
		ConfidenceScore: 0.5, // Default medium confidence
		ResidualRisk:    "medium",
		Justification:   "AI analysis completed",
		Citations:       citations,
		Severity:        types.SeverityMedium,
		Status:          types.StatusOpen,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// riskToSeverity maps residual risk to finding severity
func (e *engineImpl) riskToSeverity(risk string) string {
	switch strings.ToLower(risk) {
	case "low":
		return types.SeverityLow
	case "high":
		return types.SeverityHigh
	default:
		return types.SeverityMedium
	}
}

// responseToCachedFinding converts a cached response to a Finding
func (e *engineImpl) responseToCachedFinding(cached *CachedResult, preamble types.ContextPreamble) *types.Finding {
	// Reconstruct the finding from cached data
	// The Summary was stored in Justification field during caching
	finding := &types.Finding{
		ID:              fmt.Sprintf("finding-%d", time.Now().Unix()),
		ControlID:       cached.ControlID,
		FrameworkID:     preamble.Framework,
		Title:           fmt.Sprintf("%s %s Analysis", preamble.Framework, preamble.Section),
		Summary:         cached.Response.Justification, // Summary was stored in Justification
		MappedControls:  []string{preamble.Section},
		ConfidenceScore: float64(cached.Response.Confidence) / 100.0,
		ResidualRisk:    cached.Response.ResidualRisk,
		Justification:   "AI analysis completed (cached result)",
		Citations:       cached.Response.EvidenceLinks,
		Severity:        e.riskToSeverity(cached.Response.ResidualRisk),
		Status:          types.StatusOpen,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Mode:            "ai",
	}

	// Set review flag based on confidence
	if finding.ConfidenceScore < preamble.Rubrics.ConfidenceThreshold {
		finding.ReviewRequired = true
	}

	return finding
}

// createLowConfidenceFinding creates a finding with low confidence for empty or invalid evidence
func (e *engineImpl) createLowConfidenceFinding(preamble types.ContextPreamble, reason string) *types.Finding {
	return &types.Finding{
		ID:              fmt.Sprintf("finding-%d", time.Now().Unix()),
		ControlID:       preamble.Section,
		FrameworkID:     preamble.Framework,
		Title:           fmt.Sprintf("%s %s Analysis", preamble.Framework, preamble.Section),
		Summary:         reason,
		MappedControls:  []string{preamble.Section},
		ConfidenceScore: 0.3, // Low confidence
		ResidualRisk:    "high",
		Justification:   reason,
		Citations:       []string{},
		Severity:        types.SeverityHigh,
		Status:          types.StatusOpen,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Mode:            "ai",
		ReviewRequired:  true, // Always requires review
	}
}

// findingToCachedResult converts a Finding to a cached result
func (e *engineImpl) findingToCachedResult(cacheKey string, finding *types.Finding) *CachedResult {
	// Store Summary in Justification field for now (AnalysisResponse doesn't have Summary field)
	// When retrieving from cache, we'll use Justification as Summary
	justification := finding.Summary
	if justification == "" {
		justification = finding.Justification
	}

	return &CachedResult{
		CacheKey: cacheKey,
		Response: AnalysisResponse{
			EvidenceLinks: finding.Citations,
			Justification: justification, // Store Summary here
			Confidence:    int(finding.ConfidenceScore * 100),
			ResidualRisk:  finding.ResidualRisk,
			Provider:      e.config.AI.Provider,
			Model:         e.config.AI.Model,
			Timestamp:     time.Now(),
			CacheHit:      false,
		},
		CachedAt:  time.Now(),
		ControlID: finding.ControlID,
		Provider:  e.config.AI.Provider,
	}
}

// MockProvider is a mock implementation of Provider for testing
type MockProvider struct {
	callCount       int
	lastPrompt      string
	confidenceScore float64
	response        string
	err             error
	planItems       []types.PlanItem // For ProposePlan testing
}

// NewMockProvider creates a new MockProvider with default values
func NewMockProvider() *MockProvider {
	// Default plan items for testing
	defaultPlanItems := []types.PlanItem{
		{
			Source:         "github",
			Query:          "type:pr label:security",
			SignalStrength: 0.9,
			Rationale:      "Security PRs show access control implementations",
		},
		{
			Source:         "jira",
			Query:          "project=SEC AND status=Done",
			SignalStrength: 0.8,
			Rationale:      "Completed security tickets indicate implemented controls",
		},
	}

	return &MockProvider{
		confidenceScore: 0.85, // Default high confidence
		response:        `{"summary": "Access controls implemented", "mapped_controls": ["CC6.1"], "confidence_score": 0.85, "residual_risk": "low", "justification": "Evidence shows proper implementation", "citations": ["evt-1"]}`,
		planItems:       defaultPlanItems,
	}
}

// AnalyzeWithContext implements Provider.AnalyzeWithContext
func (m *MockProvider) AnalyzeWithContext(ctx context.Context, prompt string) (string, error) {
	m.callCount++
	m.lastPrompt = prompt

	if m.err != nil {
		return "", m.err
	}

	// Check if this is a plan generation prompt
	if strings.Contains(prompt, "evidence collection plan") {
		// Return plan items as JSON array
		if m.planItems == nil {
			return "[]", nil
		}

		type planItemResponse struct {
			Source         string  `json:"source"`
			Query          string  `json:"query"`
			SignalStrength float64 `json:"signal_strength"`
			Rationale      string  `json:"rationale"`
		}

		items := make([]planItemResponse, len(m.planItems))
		for i, item := range m.planItems {
			items[i] = planItemResponse{
				Source:         item.Source,
				Query:          item.Query,
				SignalStrength: item.SignalStrength,
				Rationale:      item.Rationale,
			}
		}

		jsonBytes, _ := json.Marshal(items)
		return string(jsonBytes), nil
	}

	// Use configured confidence score for analysis
	responseWithConf := fmt.Sprintf(`{"summary": "Access controls implemented", "mapped_controls": ["CC6.1"], "confidence_score": %.2f, "residual_risk": "low", "justification": "Evidence shows proper implementation", "citations": ["evt-1"]}`, m.confidenceScore)

	return responseWithConf, nil
}

// GetCallCount returns the number of times AnalyzeWithContext was called
func (m *MockProvider) GetCallCount() int {
	return m.callCount
}

// GetLastPrompt returns the last prompt sent to AnalyzeWithContext
func (m *MockProvider) GetLastPrompt() string {
	return m.lastPrompt
}

// SetConfidenceScore sets the confidence score for responses
func (m *MockProvider) SetConfidenceScore(score float64) {
	m.confidenceScore = score
}

// SetError sets an error to be returned by AnalyzeWithContext
func (m *MockProvider) SetError(err error) {
	m.err = err
}

// SetPlanItems sets the plan items to be returned by ProposePlan (for testing)
func (m *MockProvider) SetPlanItems(items []types.PlanItem) {
	m.planItems = items
}

// SetResponse sets a custom response to be returned
func (m *MockProvider) SetResponse(response string) {
	m.response = response
}

// MockMCPConnector is a mock implementation of MCPConnector for testing
type MockMCPConnector struct {
	events map[string][]types.EvidenceEvent // source -> events
	errors map[string]error                 // source -> error
	delay  time.Duration                    // Simulated delay
}

// NewMockMCPConnector creates a new MockMCPConnector
func NewMockMCPConnector() *MockMCPConnector {
	return &MockMCPConnector{
		events: make(map[string][]types.EvidenceEvent),
		errors: make(map[string]error),
		delay:  0,
	}
}

// Collect implements MCPConnector.Collect
func (m *MockMCPConnector) Collect(ctx context.Context, source string, query string) ([]types.EvidenceEvent, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Simulate delay if configured
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	// Check for configured error
	if err, ok := m.errors[source]; ok {
		return nil, err
	}

	// Return configured events
	if events, ok := m.events[source]; ok {
		return events, nil
	}

	// Default: return empty list
	return []types.EvidenceEvent{}, nil
}

// SetEvents sets the events to be returned for a source
func (m *MockMCPConnector) SetEvents(source string, events []types.EvidenceEvent) {
	m.events[source] = events
}

// SetError sets an error to be returned for a source
func (m *MockMCPConnector) SetError(source string, err error) {
	m.errors[source] = err
}

// SetDelay sets a delay to simulate slow connector calls
func (m *MockMCPConnector) SetDelay(delay time.Duration) {
	m.delay = delay
}
