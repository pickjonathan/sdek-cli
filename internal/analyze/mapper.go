package analyze

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Mapper maps events to framework controls using keyword matching
type Mapper struct {
	frameworks map[string]FrameworkDefinition
}

// NewMapper creates a new evidence mapper
func NewMapper() *Mapper {
	return &Mapper{
		frameworks: GetFrameworkDefinitions(),
	}
}

// MapEventsToControls maps events to controls across all frameworks
func (m *Mapper) MapEventsToControls(events []types.Event) []types.Evidence {
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
		}
	}

	return evidenceList
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
