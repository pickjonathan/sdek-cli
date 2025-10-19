package analyze

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/internal/policy"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Mapper maps events to framework controls using keyword matching and optional AI analysis
type Mapper struct {
	frameworks    map[string]FrameworkDefinition
	aiEngine      ai.Engine
	cache         *ai.Cache
	privacyFilter *ai.PrivacyFilter
	policyLoader  *policy.Loader
	promptGen     *ai.PromptGenerator
	aiEnabled     bool
}

// NewMapper creates a new evidence mapper with heuristic-only analysis
func NewMapper() *Mapper {
	return &Mapper{
		frameworks: GetFrameworkDefinitions(),
		aiEnabled:  false,
	}
}

// NewMapperWithAI creates a new evidence mapper with AI-enhanced analysis
func NewMapperWithAI(engine ai.Engine, cache *ai.Cache) *Mapper {
	privacyFilter := ai.NewPrivacyFilter()
	policyLoader := policy.NewLoader()

	return &Mapper{
		frameworks:    GetFrameworkDefinitions(),
		aiEngine:      engine,
		cache:         cache,
		privacyFilter: privacyFilter,
		policyLoader:  policyLoader,
		promptGen:     ai.NewPromptGenerator(),
		aiEnabled:     true,
	}
}

// MapEventsToControls maps events to controls across all frameworks
// If AI is enabled, uses AI-enhanced analysis with fallback to heuristics
func (m *Mapper) MapEventsToControls(events []types.Event) []types.Evidence {
	if m.aiEnabled {
		return m.mapEventsWithAI(context.Background(), events)
	}
	return m.mapEventsHeuristic(events)
}

// mapEventsWithAI performs AI-enhanced evidence mapping with heuristic fallback
func (m *Mapper) mapEventsWithAI(ctx context.Context, events []types.Event) []types.Evidence {
	var evidenceList []types.Evidence

	// First pass: Get heuristic mappings for all events
	heuristicEvidence := m.mapEventsHeuristic(events)

	// Group evidence by control for batch AI analysis
	controlEvidence := make(map[string][]types.Evidence)
	for _, ev := range heuristicEvidence {
		key := ev.FrameworkID + ":" + ev.ControlID
		controlEvidence[key] = append(controlEvidence[key], ev)
	}

	// Second pass: Enhance with AI analysis for each control
	for _, evidences := range controlEvidence {
		if len(evidences) == 0 {
			continue
		}

		// Get framework and control IDs
		firstEv := evidences[0]
		control := m.GetControlDefinition(firstEv.FrameworkID, firstEv.ControlID)
		if control == nil {
			evidenceList = append(evidenceList, evidences...)
			continue
		}

		// Gather all events for this control
		eventIDs := make(map[string]bool)
		for _, ev := range evidences {
			eventIDs[ev.EventID] = true
		}

		var controlEvents []types.Event
		for _, event := range events {
			if eventIDs[event.ID] {
				controlEvents = append(controlEvents, event)
			}
		}

		// Perform AI analysis
		aiEnhanced := m.analyzeControlWithAI(ctx, firstEv.FrameworkID, *control, controlEvents)
		if aiEnhanced != nil {
			evidenceList = append(evidenceList, aiEnhanced...)
		} else {
			// Fallback to heuristic if AI fails
			evidenceList = append(evidenceList, evidences...)
		}
	}

	return evidenceList
}

// mapEventsHeuristic performs traditional keyword-based mapping
func (m *Mapper) mapEventsHeuristic(events []types.Event) []types.Evidence {
	var evidenceList []types.Evidence

	for _, event := range events {
		// Check each framework
		for frameworkID, framework := range m.frameworks {
			// Check each control in the framework
			for _, control := range framework.Controls {
				// Check if event matches control keywords
				if m.matchesKeywords(event, control.Keywords) {
					confidenceScore := m.calculateConfidence(event, control)
					confidenceLevel := GetConfidenceLevel(confidenceScore)
					matchedKeywords := m.getMatchedKeywords(event, control.Keywords)

					evidence := types.Evidence{
						ID:                  uuid.New().String(),
						ControlID:           control.ID,
						FrameworkID:         frameworkID,
						EventID:             event.ID,
						MappedAt:            time.Now(),
						ConfidenceScore:     float64(confidenceScore),
						ConfidenceLevel:     strings.ToLower(confidenceLevel),
						Keywords:            matchedKeywords,
						Reasoning:           m.generateReasoning(event, control, matchedKeywords),
						HeuristicConfidence: confidenceScore,
						CombinedConfidence:  confidenceScore,
						AnalysisMethod:      "heuristic-only",
					}

					evidenceList = append(evidenceList, evidence)
				}
			}
		}
	}

	return evidenceList
}

// analyzeControlWithAI performs AI analysis for a specific control and its events
func (m *Mapper) analyzeControlWithAI(ctx context.Context, frameworkID string, control ControlDefinition, events []types.Event) []types.Evidence {
	if len(events) == 0 {
		return nil
	}

	// Construct full control ID for policy lookup (e.g., "SOC2-CC6.1")
	fullControlID := m.constructFullControlID(frameworkID, control.ID)

	// Get policy excerpt
	policyExcerpt, err := m.policyLoader.GetExcerpt(fullControlID)
	if err != nil || policyExcerpt == "" {
		// No policy available, fallback to heuristic
		// Note: This is expected for some controls that don't have policy excerpts
		return nil
	} // Convert to AnalysisEvents first
	analysisEvents := make([]ai.AnalysisEvent, len(events))
	for i, event := range events {
		analysisEvents[i] = ai.AnalysisEvent{
			EventID:     event.ID,
			EventType:   event.EventType,
			Source:      event.SourceID,
			Description: event.Title,
			Content:     event.Content,
			Timestamp:   event.Timestamp,
		}
	}

	// Then redact PII from analysis events
	redactedEvents := m.privacyFilter.RedactEvents(analysisEvents)

	// Create AI request
	req := &ai.AnalysisRequest{
		RequestID:     uuid.New().String(),
		ControlID:     control.ID,
		ControlName:   control.Title,
		Framework:     frameworkID,
		PolicyExcerpt: policyExcerpt,
		Events:        redactedEvents,
		Timestamp:     time.Now(),
	}

	// Generate cache key
	cacheKey := m.cache.GenerateKey(req)
	req.CacheKey = cacheKey

	// Check cache first
	cached, err := m.cache.Get(cacheKey)
	if err == nil && cached != nil {
		// Cache hit - use cached response
		return m.convertAIResponseToEvidence(cached.Response, events, frameworkID, control, true)
	}

	// Cache miss - call AI using backward compatibility layer
	slog.Debug("Calling AI for control analysis", "control", fullControlID, "eventCount", len(events))
	response, err := m.aiEngine.AnalyzeWithRequest(ctx, req)
	if err != nil {
		// AI failed - fallback to heuristic
		slog.Warn("AI analysis failed, falling back to heuristic", "control", fullControlID, "error", err)
		return nil
	}

	slog.Info("AI analysis successful", "control", fullControlID, "confidence", response.Confidence, "evidenceLinks", len(response.EvidenceLinks))

	// Store in cache
	cachedResult := &ai.CachedResult{
		CacheKey:     cacheKey,
		Response:     *response,
		CachedAt:     time.Now(),
		EventIDs:     extractEventIDs(events),
		ControlID:    control.ID,
		Provider:     response.Provider,
		ModelVersion: response.Model,
	}
	_ = m.cache.Set(cacheKey, cachedResult)

	// Convert AI response to evidence
	return m.convertAIResponseToEvidence(*response, events, frameworkID, control, false)
}

// convertAIResponseToEvidence converts AI analysis response to Evidence records
func (m *Mapper) convertAIResponseToEvidence(response ai.AnalysisResponse, events []types.Event, frameworkID string, control ControlDefinition, cacheHit bool) []types.Evidence {
	var evidenceList []types.Evidence

	slog.Debug("Converting AI response", "control", control.ID, "aiLinks", len(response.EvidenceLinks), "availableEvents", len(events))

	// Create evidence for each linked event
	for _, eventRef := range response.EvidenceLinks {
		// Find the actual event - handle both UUIDs and event numbers
		var matchedEvent *types.Event

		// Try parsing as event number first (e.g., "1", "2", "3")
		if eventNum, err := strconv.Atoi(eventRef); err == nil && eventNum > 0 && eventNum <= len(events) {
			// Event number (1-indexed)
			matchedEvent = &events[eventNum-1]
			slog.Debug("Matched event by number", "eventNum", eventNum, "eventID", matchedEvent.ID)
		} else {
			// Try matching by UUID
			for i, event := range events {
				if event.ID == eventRef {
					matchedEvent = &events[i]
					slog.Debug("Matched event by UUID", "eventID", eventRef)
					break
				}
			}
		}

		if matchedEvent == nil {
			// Log available event IDs for debugging
			if len(events) > 0 {
				slog.Warn("AI evidence link not found", "wantedEventRef", eventRef, "control", control.ID, "availableEvents", len(events))
			} else {
				slog.Warn("AI evidence link not found - no events available", "wantedEventRef", eventRef, "control", control.ID)
			}
			continue
		}

		// Calculate heuristic confidence for comparison
		heuristicScore := m.calculateConfidence(*matchedEvent, control)

		// Calculate weighted combined confidence (70% AI + 30% heuristic)
		combinedScore := int(float64(response.Confidence)*0.7 + float64(heuristicScore)*0.3)

		analysisMethod := "ai+heuristic"
		if cacheHit {
			analysisMethod = "ai+heuristic (cached)"
		}

		evidence := types.Evidence{
			ID:                  uuid.New().String(),
			ControlID:           control.ID,
			FrameworkID:         frameworkID,
			EventID:             matchedEvent.ID,
			MappedAt:            time.Now(),
			ConfidenceScore:     float64(combinedScore),
			ConfidenceLevel:     strings.ToLower(GetConfidenceLevel(combinedScore)),
			Keywords:            m.getMatchedKeywords(*matchedEvent, control.Keywords),
			Reasoning:           response.Justification,
			AIAnalyzed:          true,
			AIJustification:     response.Justification,
			AIConfidence:        response.Confidence,
			AIResidualRisk:      response.ResidualRisk,
			HeuristicConfidence: heuristicScore,
			CombinedConfidence:  combinedScore,
			AnalysisMethod:      analysisMethod,
		}

		slog.Debug("Created AI-enhanced evidence", "evidenceID", evidence.ID, "eventID", matchedEvent.ID, "control", control.ID, "aiAnalyzed", evidence.AIAnalyzed)
		evidenceList = append(evidenceList, evidence)
	}

	slog.Info("Converted AI response to evidence", "control", control.ID, "evidenceCount", len(evidenceList))
	return evidenceList
}

// extractEventIDs extracts event IDs from a slice of events
func extractEventIDs(events []types.Event) []string {
	ids := make([]string, len(events))
	for i, event := range events {
		ids[i] = event.ID
	}
	return ids
}

// matchesKeywords checks if an event matches any of the control keywords
func (m *Mapper) matchesKeywords(event types.Event, keywords []string) bool {
	// Combine searchable text from event
	searchText := strings.ToLower(event.Title + " " + event.Content)

	// Check if any keyword matches
	for _, keyword := range keywords {
		if strings.Contains(searchText, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

// calculateConfidence calculates the confidence score for an evidence mapping
func (m *Mapper) calculateConfidence(event types.Event, control ControlDefinition) int {
	score := 0

	// Base score for keyword match
	searchText := strings.ToLower(event.Title + " " + event.Content)
	matchCount := 0

	for _, keyword := range control.Keywords {
		if strings.Contains(searchText, strings.ToLower(keyword)) {
			matchCount++
		}
	}

	// More keyword matches = higher confidence
	if matchCount == 1 {
		score += 40
	} else if matchCount == 2 {
		score += 60
	} else if matchCount >= 3 {
		score += 80
	}

	// Recency bonus (events in last 30 days get bonus)
	daysSince := time.Since(event.Timestamp).Hours() / 24
	if daysSince <= 30 {
		score += 15
	} else if daysSince <= 60 {
		score += 10
	} else {
		score += 5
	}

	// Source type bonus (some sources are more reliable)
	switch event.SourceID {
	case string(types.SourceTypeGit):
		score += 5 // Code commits are reliable
	case string(types.SourceTypeCICD):
		score += 5 // Build/test results are reliable
	case string(types.SourceTypeDocs):
		score += 10 // Documentation is most reliable
	case string(types.SourceTypeJira):
		score += 3 // Tickets are moderately reliable
	case string(types.SourceTypeSlack):
		score += 2 // Messages are least reliable
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// GetConfidenceLevel returns the confidence level category
func GetConfidenceLevel(confidence int) string {
	if confidence <= 50 {
		return "Low"
	} else if confidence <= 75 {
		return "Medium"
	}
	return "High"
}

// MapEventToFramework maps a single event to controls in a specific framework
func (m *Mapper) MapEventToFramework(event types.Event, frameworkID string) []types.Evidence {
	var evidenceList []types.Evidence

	framework, exists := m.frameworks[frameworkID]
	if !exists {
		return evidenceList
	}

	for _, control := range framework.Controls {
		if m.matchesKeywords(event, control.Keywords) {
			confidenceScore := m.calculateConfidence(event, control)
			confidenceLevel := GetConfidenceLevel(confidenceScore)
			matchedKeywords := m.getMatchedKeywords(event, control.Keywords)

			evidence := types.Evidence{
				ID:              uuid.New().String(),
				ControlID:       control.ID,
				FrameworkID:     frameworkID,
				EventID:         event.ID,
				MappedAt:        time.Now(),
				ConfidenceScore: float64(confidenceScore),
				ConfidenceLevel: strings.ToLower(confidenceLevel),
				Keywords:        matchedKeywords,
				Reasoning:       m.generateReasoning(event, control, matchedKeywords),
			}

			evidenceList = append(evidenceList, evidence)
		}
	}

	return evidenceList
}

// GetControlDefinition retrieves a control definition by ID
func (m *Mapper) GetControlDefinition(frameworkID, controlID string) *ControlDefinition {
	framework, exists := m.frameworks[frameworkID]
	if !exists {
		return nil
	}

	for _, control := range framework.Controls {
		if control.ID == controlID {
			return &control
		}
	}

	return nil
}

// GetFramework retrieves a framework definition
func (m *Mapper) GetFramework(frameworkID string) *FrameworkDefinition {
	if framework, exists := m.frameworks[frameworkID]; exists {
		return &framework
	}
	return nil
}

// CountControlsForFramework returns the number of controls in a framework
func (m *Mapper) CountControlsForFramework(frameworkID string) int {
	if framework, exists := m.frameworks[frameworkID]; exists {
		return len(framework.Controls)
	}
	return 0
}

// getMatchedKeywords returns the list of keywords that matched in the event
func (m *Mapper) getMatchedKeywords(event types.Event, keywords []string) []string {
	searchText := strings.ToLower(event.Title + " " + event.Content)
	var matched []string

	for _, keyword := range keywords {
		if strings.Contains(searchText, strings.ToLower(keyword)) {
			matched = append(matched, keyword)
		}
	}

	return matched
}

// generateReasoning generates a human-readable reasoning for the evidence mapping
func (m *Mapper) generateReasoning(event types.Event, control ControlDefinition, matchedKeywords []string) string {
	if len(matchedKeywords) == 0 {
		return "No specific keywords matched"
	}

	if len(matchedKeywords) == 1 {
		return "Event mentions: " + matchedKeywords[0]
	}

	return "Event mentions: " + strings.Join(matchedKeywords, ", ")
}

// constructFullControlID constructs the full control ID for policy lookup
// Examples: "soc2" + "CC6.1" -> "SOC2-CC6.1", "iso27001" + "A.9.1" -> "ISO27001-A.9.1"
func (m *Mapper) constructFullControlID(frameworkID, controlID string) string {
	// Normalize framework ID to uppercase and replace underscores with hyphens
	normalized := strings.ToUpper(strings.ReplaceAll(frameworkID, "_", "-"))

	// Handle special cases
	switch normalized {
	case "SOC2":
		return "SOC2-" + controlID
	case "ISO27001":
		return "ISO27001-" + controlID
	case "PCI-DSS", "PCIDSS":
		return "PCI-DSS-" + controlID
	default:
		return normalized + "-" + controlID
	}
}
