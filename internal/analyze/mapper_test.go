package analyze

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// TestNewMapper verifies mapper initialization
func TestNewMapper(t *testing.T) {
	mapper := NewMapper()

	if mapper == nil {
		t.Fatal("Mapper should not be nil")
	}

	if mapper.frameworks == nil {
		t.Fatal("Frameworks should be initialized")
	}

	// Verify all three frameworks are loaded
	if len(mapper.frameworks) != 3 {
		t.Errorf("Expected 3 frameworks, got %d", len(mapper.frameworks))
	}
}

// TestMapEventsToControls verifies event mapping
func TestMapEventsToControls(t *testing.T) {
	mapper := NewMapper()

	// Create test events
	events := []types.Event{
		{
			ID:        "event-1",
			SourceID:  string(types.SourceTypeGit),
			Timestamp: time.Now().AddDate(0, 0, -10),
			EventType: types.EventTypeCommit,
			Title:     "Add authentication system",
			Content:   "Implement OAuth authentication with multi-factor support",
			Author:    "Alice",
			Metadata:  map[string]interface{}{},
		},
		{
			ID:        "event-2",
			SourceID:  string(types.SourceTypeDocs),
			Timestamp: time.Now().AddDate(0, 0, -5),
			EventType: types.EventTypeDocumentChange,
			Title:     "Update encryption policy",
			Content:   "Document new TLS 1.3 encryption requirements",
			Author:    "Bob",
			Metadata:  map[string]interface{}{},
		},
	}

	evidence := mapper.MapEventsToControls(events)

	if len(evidence) == 0 {
		t.Fatal("Expected evidence to be generated")
	}

	// Verify evidence has required fields
	for i, ev := range evidence {
		if ev.ID == "" {
			t.Errorf("Evidence %d: ID is empty", i)
		}
		if ev.EventID == "" {
			t.Errorf("Evidence %d: EventID is empty", i)
		}
		if ev.ControlID == "" {
			t.Errorf("Evidence %d: ControlID is empty", i)
		}
		if ev.FrameworkID == "" {
			t.Errorf("Evidence %d: FrameworkID is empty", i)
		}
		if ev.ConfidenceScore <= 0 {
			t.Errorf("Evidence %d: ConfidenceScore should be positive", i)
		}
		if ev.ConfidenceLevel == "" {
			t.Errorf("Evidence %d: ConfidenceLevel is empty", i)
		}
	}
}

// TestMatchesKeywords verifies keyword matching
func TestMatchesKeywords(t *testing.T) {
	mapper := NewMapper()

	tests := []struct {
		name     string
		event    types.Event
		keywords []string
		expected bool
	}{
		{
			name: "Single keyword match in title",
			event: types.Event{
				Title:   "Update authentication system",
				Content: "Some content",
			},
			keywords: []string{"authentication", "authorization"},
			expected: true,
		},
		{
			name: "Keyword match in content",
			event: types.Event{
				Title:   "System update",
				Content: "Added encryption for sensitive data",
			},
			keywords: []string{"encryption", "crypto"},
			expected: true,
		},
		{
			name: "No keyword match",
			event: types.Event{
				Title:   "Update documentation",
				Content: "Fixed typos in README",
			},
			keywords: []string{"authentication", "encryption"},
			expected: false,
		},
		{
			name: "Case insensitive match",
			event: types.Event{
				Title:   "IMPLEMENT FIREWALL",
				Content: "Network security update",
			},
			keywords: []string{"firewall"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.matchesKeywords(tt.event, tt.keywords)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCalculateConfidence verifies confidence calculation
func TestCalculateConfidence(t *testing.T) {
	mapper := NewMapper()

	control := ControlDefinition{
		Keywords: []string{"authentication", "authorization", "access control"},
	}

	tests := []struct {
		name          string
		event         types.Event
		minConfidence int
		maxConfidence int
	}{
		{
			name: "Recent event with multiple keywords",
			event: types.Event{
				SourceID:  string(types.SourceTypeDocs),
				Timestamp: time.Now().AddDate(0, 0, -10),
				Title:     "Authentication and authorization system",
				Content:   "Implement access control with MFA",
			},
			minConfidence: 80,
			maxConfidence: 100,
		},
		{
			name: "Old event with single keyword",
			event: types.Event{
				SourceID:  string(types.SourceTypeSlack),
				Timestamp: time.Now().AddDate(0, 0, -80),
				Title:     "Discuss authentication",
				Content:   "Let's review the auth system",
			},
			minConfidence: 40,
			maxConfidence: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := mapper.calculateConfidence(tt.event, control)

			if confidence < tt.minConfidence || confidence > tt.maxConfidence {
				t.Errorf("Expected confidence between %d and %d, got %d",
					tt.minConfidence, tt.maxConfidence, confidence)
			}
		})
	}
}

// TestGetConfidenceLevel verifies confidence level categorization
func TestGetConfidenceLevel(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{0, "Low"},
		{25, "Low"},
		{50, "Low"},
		{51, "Medium"},
		{65, "Medium"},
		{75, "Medium"},
		{76, "High"},
		{90, "High"},
		{100, "High"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			level := GetConfidenceLevel(tt.score)
			if level != tt.expected {
				t.Errorf("Score %d: expected %s, got %s", tt.score, tt.expected, level)
			}
		})
	}
}

// TestMapEventToFramework verifies single framework mapping
func TestMapEventToFramework(t *testing.T) {
	mapper := NewMapper()

	event := types.Event{
		ID:        "event-1",
		SourceID:  string(types.SourceTypeGit),
		Timestamp: time.Now().AddDate(0, 0, -10),
		Title:     "Add firewall configuration",
		Content:   "Configure network security firewall",
	}

	evidence := mapper.MapEventToFramework(event, string(types.FrameworkPCIDSS))

	if len(evidence) == 0 {
		t.Fatal("Expected evidence for PCI DSS firewall control")
	}

	// Verify all evidence is for PCI DSS
	for _, ev := range evidence {
		if ev.FrameworkID != string(types.FrameworkPCIDSS) {
			t.Errorf("Expected framework %s, got %s", types.FrameworkPCIDSS, ev.FrameworkID)
		}
	}
}

// TestGetControlDefinition verifies control retrieval
func TestGetControlDefinition(t *testing.T) {
	mapper := NewMapper()

	// Test valid control
	control := mapper.GetControlDefinition(string(types.FrameworkSOC2), "CC6.1")
	if control == nil {
		t.Fatal("Expected to find SOC2 control CC6.1")
	}
	if control.ID != "CC6.1" {
		t.Errorf("Expected control ID CC6.1, got %s", control.ID)
	}

	// Test invalid framework
	control = mapper.GetControlDefinition("invalid-framework", "CC6.1")
	if control != nil {
		t.Error("Expected nil for invalid framework")
	}

	// Test invalid control
	control = mapper.GetControlDefinition(string(types.FrameworkSOC2), "INVALID")
	if control != nil {
		t.Error("Expected nil for invalid control")
	}
}

// TestGetFramework verifies framework retrieval
func TestGetFramework(t *testing.T) {
	mapper := NewMapper()

	// Test valid framework
	framework := mapper.GetFramework(string(types.FrameworkSOC2))
	if framework == nil {
		t.Fatal("Expected to find SOC2 framework")
	}
	if framework.ID != string(types.FrameworkSOC2) {
		t.Errorf("Expected framework ID %s, got %s", types.FrameworkSOC2, framework.ID)
	}

	// Test invalid framework
	framework = mapper.GetFramework("invalid-framework")
	if framework != nil {
		t.Error("Expected nil for invalid framework")
	}
}

// TestCountControlsForFramework verifies control counting
func TestCountControlsForFramework(t *testing.T) {
	mapper := NewMapper()

	tests := []struct {
		framework string
		expected  int
	}{
		{string(types.FrameworkSOC2), 45},
		{string(types.FrameworkISO27001), 64},
		{string(types.FrameworkPCIDSS), 15},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.framework, func(t *testing.T) {
			count := mapper.CountControlsForFramework(tt.framework)
			if count != tt.expected {
				t.Errorf("Expected %d controls, got %d", tt.expected, count)
			}
		})
	}
}

// TestGetMatchedKeywords verifies keyword extraction
func TestGetMatchedKeywords(t *testing.T) {
	mapper := NewMapper()

	event := types.Event{
		Title:   "Implement authentication and encryption",
		Content: "Add OAuth authentication with TLS encryption",
	}

	keywords := []string{"authentication", "encryption", "authorization", "firewall"}
	matched := mapper.getMatchedKeywords(event, keywords)

	if len(matched) != 2 {
		t.Errorf("Expected 2 matched keywords, got %d", len(matched))
	}

	// Verify the matched keywords
	expectedMatches := map[string]bool{"authentication": true, "encryption": true}
	for _, keyword := range matched {
		if !expectedMatches[keyword] {
			t.Errorf("Unexpected matched keyword: %s", keyword)
		}
	}
}

// TestGenerateReasoning verifies reasoning generation
func TestGenerateReasoning(t *testing.T) {
	mapper := NewMapper()

	event := types.Event{Title: "Test", Content: "Test"}
	control := ControlDefinition{ID: "TEST-1"}

	tests := []struct {
		name     string
		keywords []string
		contains string
	}{
		{
			name:     "No keywords",
			keywords: []string{},
			contains: "No specific keywords",
		},
		{
			name:     "Single keyword",
			keywords: []string{"authentication"},
			contains: "Event mentions: authentication",
		},
		{
			name:     "Multiple keywords",
			keywords: []string{"authentication", "encryption"},
			contains: "Event mentions:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reasoning := mapper.generateReasoning(event, control, tt.keywords)
			if reasoning == "" {
				t.Error("Reasoning should not be empty")
			}
		})
	}
}

// TestNewMapperWithAI verifies AI-enhanced mapper initialization
func TestNewMapperWithAI(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	// Create mock AI engine
	mockEngine := &mockAIEngine{}
	cache, err := newTestCache()
	if err != nil {
		t.Fatalf("Failed to create test cache: %v", err)
	}

	mapper := NewMapperWithAI(mockEngine, cache)

	if mapper == nil {
		t.Fatal("Mapper should not be nil")
	}

	if !mapper.aiEnabled {
		t.Error("AI should be enabled")
	}

	if mapper.aiEngine == nil {
		t.Error("AI engine should be set")
	}

	if mapper.cache == nil {
		t.Error("Cache should be set")
	}

	if mapper.privacyFilter == nil {
		t.Error("Privacy filter should be set")
	}

	if mapper.policyLoader == nil {
		t.Error("Policy loader should be set")
	}
}

// TestMapEventsWithAI_Success verifies AI-enhanced mapping with successful AI response
func TestMapEventsWithAI_Success(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	mockEngine := &mockAIEngine{
		response: &mockAIResponse{
			evidenceLinks: []string{"event-1"},
			justification: "Event shows authentication implementation",
			confidence:    85,
			residualRisk:  "No multi-region support yet",
		},
	}

	cache, _ := newTestCache()
	mapper := NewMapperWithAI(mockEngine, cache)

	events := []types.Event{
		{
			ID:        "event-1",
			SourceID:  string(types.SourceTypeGit),
			Timestamp: time.Now(),
			EventType: types.EventTypeCommit,
			Title:     "Add authentication",
			Content:   "Implement OAuth with MFA",
		},
	}

	evidence := mapper.MapEventsToControls(events)

	if len(evidence) == 0 {
		t.Fatal("Expected evidence to be mapped")
	}

	// Check AI fields are populated
	if !evidence[0].AIAnalyzed {
		t.Error("Expected AIAnalyzed to be true")
	}

	if evidence[0].AIConfidence != 85 {
		t.Errorf("Expected AI confidence 85, got %d", evidence[0].AIConfidence)
	}

	if evidence[0].AIJustification == "" {
		t.Error("Expected AI justification to be set")
	}

	if evidence[0].AnalysisMethod != "ai+heuristic" {
		t.Errorf("Expected analysis method 'ai+heuristic', got %s", evidence[0].AnalysisMethod)
	}

	// Check hybrid confidence (70% AI + 30% heuristic)
	if evidence[0].CombinedConfidence == 0 {
		t.Error("Expected combined confidence to be calculated")
	}
}

// TestMapEventsWithAI_Fallback verifies fallback to heuristics when AI fails
func TestMapEventsWithAI_Fallback(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	mockEngine := &mockAIEngine{
		shouldError: true,
	}

	cache, _ := newTestCache()
	mapper := NewMapperWithAI(mockEngine, cache)

	events := []types.Event{
		{
			ID:        "event-1",
			SourceID:  string(types.SourceTypeGit),
			Timestamp: time.Now(),
			EventType: types.EventTypeCommit,
			Title:     "Add authentication",
			Content:   "Implement OAuth with MFA",
		},
	}

	evidence := mapper.MapEventsToControls(events)

	if len(evidence) == 0 {
		t.Fatal("Expected fallback to heuristic mapping")
	}

	// Check that heuristic method was used
	if evidence[0].AnalysisMethod != "heuristic-only" {
		t.Errorf("Expected analysis method 'heuristic-only', got %s", evidence[0].AnalysisMethod)
	}

	if evidence[0].AIAnalyzed {
		t.Error("Expected AIAnalyzed to be false on fallback")
	}
}

// TestMapEventsWithAI_CacheHit verifies cache functionality
func TestMapEventsWithAI_CacheHit(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	mockEngine := &mockAIEngine{
		callCount: 0,
	}

	cache, _ := newTestCache()
	mapper := NewMapperWithAI(mockEngine, cache)

	events := []types.Event{
		{
			ID:        "event-1",
			SourceID:  string(types.SourceTypeGit),
			Timestamp: time.Now(),
			EventType: types.EventTypeCommit,
			Title:     "Add authentication",
			Content:   "Implement OAuth",
		},
	}

	// First call - should hit AI
	evidence1 := mapper.MapEventsToControls(events)
	firstCallCount := mockEngine.callCount

	// Second call - should hit cache
	evidence2 := mapper.MapEventsToControls(events)
	secondCallCount := mockEngine.callCount

	if firstCallCount == 0 {
		t.Error("Expected AI engine to be called on first request")
	}

	if secondCallCount != firstCallCount {
		t.Error("Expected AI engine to NOT be called on second request (cache hit)")
	}

	if len(evidence1) != len(evidence2) {
		t.Error("Expected same evidence from cache")
	}

	if evidence2[0].AnalysisMethod != "ai+heuristic (cached)" {
		t.Errorf("Expected cached analysis method, got %s", evidence2[0].AnalysisMethod)
	}
}

// TestMapEventsWithAI_PrivacyRedaction verifies PII redaction before AI
func TestMapEventsWithAI_PrivacyRedaction(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	mockEngine := &mockAIEngine{
		captureRequest: true,
	}

	cache, _ := newTestCache()
	mapper := NewMapperWithAI(mockEngine, cache)

	events := []types.Event{
		{
			ID:        "event-1",
			SourceID:  string(types.SourceTypeGit),
			Timestamp: time.Now(),
			EventType: types.EventTypeCommit,
			Title:     "Add user auth",
			Content:   "Email: user@example.com, API Key: sk-abc123def456",
		},
	}

	mapper.MapEventsToControls(events)

	// Check that captured request had redacted content
	if mockEngine.lastRequest == nil {
		t.Fatal("Expected AI engine to capture request")
	}

	// Verify email and API key were redacted
	for _, event := range mockEngine.lastRequest.Events {
		if event.Content == events[0].Content {
			t.Error("Expected event content to be redacted")
		}
		if !containsRedactionMarker(event.Content) {
			t.Error("Expected redaction markers in content")
		}
	}
}

// Helper functions for tests
func newTestCache() (*ai.Cache, error) {
	// Create temporary cache directory
	return ai.NewCache("")
}

type mockAIEngine struct {
	response       *mockAIResponse
	shouldError    bool
	callCount      int
	captureRequest bool
	lastRequest    *ai.AnalysisRequest
}

type mockAIResponse struct {
	evidenceLinks []string
	justification string
	confidence    int
	residualRisk  string
}

func (m *mockAIEngine) Analyze(ctx context.Context, req *ai.AnalysisRequest) (*ai.AnalysisResponse, error) {
	m.callCount++

	if m.captureRequest {
		m.lastRequest = req
	}

	if m.shouldError {
		return nil, ai.ErrProviderUnavailable
	}

	if m.response == nil {
		m.response = &mockAIResponse{
			evidenceLinks: []string{req.Events[0].EventID},
			justification: "Mock analysis",
			confidence:    85,
			residualRisk:  "",
		}
	}

	return &ai.AnalysisResponse{
		RequestID:     req.RequestID,
		EvidenceLinks: m.response.evidenceLinks,
		Justification: m.response.justification,
		Confidence:    m.response.confidence,
		ResidualRisk:  m.response.residualRisk,
		Provider:      "mock",
		Model:         "mock-model",
		TokensUsed:    100,
		Timestamp:     time.Now(),
		CacheHit:      false,
	}, nil
}

func (m *mockAIEngine) Provider() string {
	return "mock"
}

func (m *mockAIEngine) Health(ctx context.Context) error {
	return nil
}

func containsRedactionMarker(s string) bool {
	return strings.Contains(s, "[REDACTED") || strings.Contains(s, "[EMAIL") || strings.Contains(s, "[API_KEY")
}
