package analyze

import (
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// ConfidenceCalculator calculates confidence scores for evidence mappings
type ConfidenceCalculator struct{}

// NewConfidenceCalculator creates a new confidence calculator
func NewConfidenceCalculator() *ConfidenceCalculator {
	return &ConfidenceCalculator{}
}

// Calculate computes the confidence score (0-100) for an event-control mapping
func (c *ConfidenceCalculator) Calculate(event types.Event, matchCount int, sourceType string) int {
	score := 0

	// Keyword match score (max 80 points)
	score += c.keywordMatchScore(matchCount)

	// Recency score (max 15 points)
	score += c.recencyScore(event.Timestamp)

	// Source reliability score (max 10 points)
	score += c.sourceReliabilityScore(sourceType)

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// keywordMatchScore returns score based on number of keyword matches
func (c *ConfidenceCalculator) keywordMatchScore(matchCount int) int {
	switch {
	case matchCount >= 3:
		return 80 // High confidence with 3+ matches
	case matchCount == 2:
		return 60 // Medium-high confidence with 2 matches
	case matchCount == 1:
		return 40 // Medium confidence with 1 match
	default:
		return 0 // No matches
	}
}

// recencyScore returns score based on how recent the event is
func (c *ConfidenceCalculator) recencyScore(timestamp time.Time) int {
	daysSince := time.Since(timestamp).Hours() / 24

	switch {
	case daysSince < 31:
		return 15 // Recent events (last month)
	case daysSince < 61:
		return 10 // Moderately recent (last 2 months)
	default:
		return 5 // Older events
	}
}

// sourceReliabilityScore returns score based on source type reliability
func (c *ConfidenceCalculator) sourceReliabilityScore(sourceType string) int {
	switch sourceType {
	case string(types.SourceTypeDocs):
		return 10 // Documentation is most reliable
	case string(types.SourceTypeGit):
		return 5 // Code commits are reliable
	case string(types.SourceTypeCICD):
		return 5 // Build/test results are reliable
	case string(types.SourceTypeJira):
		return 3 // Tickets are moderately reliable
	case string(types.SourceTypeSlack):
		return 2 // Messages are least reliable
	default:
		return 0
	}
}

// GetLevel converts a numeric confidence score to a level string
func (c *ConfidenceCalculator) GetLevel(score int) string {
	switch {
	case score <= 50:
		return types.ConfidenceLevelLow
	case score <= 75:
		return types.ConfidenceLevelMedium
	default:
		return types.ConfidenceLevelHigh
	}
}

// CalculateHybridConfidence computes weighted average of AI and heuristic confidence
// Uses 70% AI confidence + 30% heuristic confidence as per spec
func (c *ConfidenceCalculator) CalculateHybridConfidence(aiConfidence, heuristicConfidence int) int {
	// Weighted average: 70% AI + 30% heuristic
	weighted := float64(aiConfidence)*0.7 + float64(heuristicConfidence)*0.3
	result := int(weighted)

	// Ensure within bounds
	if result < 0 {
		return 0
	}
	if result > 100 {
		return 100
	}

	return result
}

// ValidateConfidenceScore ensures score is within valid range (0-100)
func (c *ConfidenceCalculator) ValidateConfidenceScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}
