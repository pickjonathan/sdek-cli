package analyze

import (
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// TestNewConfidenceCalculator verifies calculator initialization
func TestNewConfidenceCalculator(t *testing.T) {
	calc := NewConfidenceCalculator()
	if calc == nil {
		t.Fatal("Calculator should not be nil")
	}
}

// TestCalculate verifies overall confidence calculation
func TestCalculate(t *testing.T) {
	calc := NewConfidenceCalculator()

	tests := []struct {
		name       string
		event      types.Event
		matchCount int
		sourceType string
		minScore   int
		maxScore   int
	}{
		{
			name: "High confidence - recent docs with multiple matches",
			event: types.Event{
				Timestamp: time.Now().AddDate(0, 0, -10),
			},
			matchCount: 3,
			sourceType: string(types.SourceTypeDocs),
			minScore:   95,
			maxScore:   100,
		},
		{
			name: "Medium confidence - moderate recency and matches",
			event: types.Event{
				Timestamp: time.Now().AddDate(0, 0, -45),
			},
			matchCount: 2,
			sourceType: string(types.SourceTypeGit),
			minScore:   70,
			maxScore:   80,
		},
		{
			name: "Low confidence - old slack message with single match",
			event: types.Event{
				Timestamp: time.Now().AddDate(0, 0, -85),
			},
			matchCount: 1,
			sourceType: string(types.SourceTypeSlack),
			minScore:   40,
			maxScore:   50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calc.Calculate(tt.event, tt.matchCount, tt.sourceType)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("Expected score between %d and %d, got %d",
					tt.minScore, tt.maxScore, score)
			}
		})
	}
}

// TestKeywordMatchScore verifies keyword match scoring
func TestKeywordMatchScore(t *testing.T) {
	calc := NewConfidenceCalculator()

	tests := []struct {
		matchCount int
		expected   int
	}{
		{0, 0},
		{1, 40},
		{2, 60},
		{3, 80},
		{4, 80},  // Capped at 80
		{10, 80}, // Capped at 80
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			score := calc.keywordMatchScore(tt.matchCount)
			if score != tt.expected {
				t.Errorf("Match count %d: expected %d, got %d",
					tt.matchCount, tt.expected, score)
			}
		})
	}
}

// TestRecencyScore verifies recency scoring
func TestRecencyScore(t *testing.T) {
	calc := NewConfidenceCalculator()

	tests := []struct {
		name      string
		timestamp time.Time
		expected  int
	}{
		{
			name:      "Very recent (10 days)",
			timestamp: time.Now().AddDate(0, 0, -10),
			expected:  15,
		},
		{
			name:      "Recent (45 days)",
			timestamp: time.Now().AddDate(0, 0, -45),
			expected:  10,
		},
		{
			name:      "Old (80 days)",
			timestamp: time.Now().AddDate(0, 0, -80),
			expected:  5,
		},
		{
			name:      "At 30 day boundary",
			timestamp: time.Now().AddDate(0, 0, -30),
			expected:  15,
		},
		{
			name:      "At 60 day boundary",
			timestamp: time.Now().AddDate(0, 0, -60),
			expected:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calc.recencyScore(tt.timestamp)
			if score != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, score)
			}
		})
	}
}

// TestSourceReliabilityScore verifies source reliability scoring
func TestSourceReliabilityScore(t *testing.T) {
	calc := NewConfidenceCalculator()

	tests := []struct {
		sourceType string
		expected   int
	}{
		{string(types.SourceTypeDocs), 10},
		{string(types.SourceTypeGit), 5},
		{string(types.SourceTypeCICD), 5},
		{string(types.SourceTypeJira), 3},
		{string(types.SourceTypeSlack), 2},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.sourceType, func(t *testing.T) {
			score := calc.sourceReliabilityScore(tt.sourceType)
			if score != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, score)
			}
		})
	}
}

// TestGetLevel verifies confidence level conversion
func TestGetLevel(t *testing.T) {
	calc := NewConfidenceCalculator()

	tests := []struct {
		score    int
		expected string
	}{
		{0, types.ConfidenceLevelLow},
		{25, types.ConfidenceLevelLow},
		{50, types.ConfidenceLevelLow},
		{51, types.ConfidenceLevelMedium},
		{65, types.ConfidenceLevelMedium},
		{75, types.ConfidenceLevelMedium},
		{76, types.ConfidenceLevelHigh},
		{90, types.ConfidenceLevelHigh},
		{100, types.ConfidenceLevelHigh},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			level := calc.GetLevel(tt.score)
			if level != tt.expected {
				t.Errorf("Score %d: expected %s, got %s",
					tt.score, tt.expected, level)
			}
		})
	}
}

// TestCalculateWithCapping verifies score is capped at 100
func TestCalculateWithCapping(t *testing.T) {
	calc := NewConfidenceCalculator()

	// Create scenario that could exceed 100 (80 + 15 + 10 = 105)
	event := types.Event{
		Timestamp: time.Now().AddDate(0, 0, -10), // 15 points
	}

	score := calc.Calculate(event, 5, string(types.SourceTypeDocs)) // 80 + 10 points

	if score > 100 {
		t.Errorf("Score should be capped at 100, got %d", score)
	}

	if score != 100 {
		t.Errorf("Expected score to be capped at 100, got %d", score)
	}
}

// TestCalculateHybridConfidence verifies 70% AI + 30% heuristic weighting
func TestCalculateHybridConfidence(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	calc := NewConfidenceCalculator()

	tests := []struct {
		name                string
		aiConfidence        int
		heuristicConfidence int
		expected            int
	}{
		{
			name:                "Equal scores",
			aiConfidence:        80,
			heuristicConfidence: 80,
			expected:            80, // (80*0.7 + 80*0.3) = 80
		},
		{
			name:                "AI higher than heuristic",
			aiConfidence:        90,
			heuristicConfidence: 50,
			expected:            78, // (90*0.7 + 50*0.3) = 63 + 15 = 78
		},
		{
			name:                "Heuristic higher than AI",
			aiConfidence:        60,
			heuristicConfidence: 90,
			expected:            69, // (60*0.7 + 90*0.3) = 42 + 27 = 69
		},
		{
			name:                "Maximum scores",
			aiConfidence:        100,
			heuristicConfidence: 100,
			expected:            100,
		},
		{
			name:                "Minimum scores",
			aiConfidence:        0,
			heuristicConfidence: 0,
			expected:            0,
		},
		{
			name:                "AI high, heuristic zero",
			aiConfidence:        100,
			heuristicConfidence: 0,
			expected:            70, // (100*0.7 + 0*0.3) = 70
		},
		{
			name:                "AI zero, heuristic high",
			aiConfidence:        0,
			heuristicConfidence: 100,
			expected:            30, // (0*0.7 + 100*0.3) = 30
		},
		{
			name:                "Mid-range scores",
			aiConfidence:        75,
			heuristicConfidence: 65,
			expected:            72, // (75*0.7 + 65*0.3) = 52.5 + 19.5 = 72
		},
		{
			name:                "Low AI, mid heuristic",
			aiConfidence:        30,
			heuristicConfidence: 70,
			expected:            42, // (30*0.7 + 70*0.3) = 21 + 21 = 42
		},
		{
			name:                "High AI, low heuristic",
			aiConfidence:        95,
			heuristicConfidence: 20,
			expected:            72, // (95*0.7 + 20*0.3) = 66.5 + 6 = 72.5 → 72
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateHybridConfidence(tt.aiConfidence, tt.heuristicConfidence)
			if result != tt.expected {
				t.Errorf("CalculateHybridConfidence(%d, %d) = %d, expected %d",
					tt.aiConfidence, tt.heuristicConfidence, result, tt.expected)
			}
		})
	}
}

// TestCalculateHybridConfidence_Boundaries verifies boundary conditions
func TestCalculateHybridConfidence_Boundaries(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	calc := NewConfidenceCalculator()

	tests := []struct {
		name                string
		aiConfidence        int
		heuristicConfidence int
		description         string
	}{
		{
			name:                "Negative AI score clamped to zero",
			aiConfidence:        -10,
			heuristicConfidence: 50,
			description:         "Should treat negative as 0",
		},
		{
			name:                "Negative heuristic score clamped to zero",
			aiConfidence:        50,
			heuristicConfidence: -10,
			description:         "Should treat negative as 0",
		},
		{
			name:                "AI over 100 clamped",
			aiConfidence:        150,
			heuristicConfidence: 50,
			description:         "Should cap at 100",
		},
		{
			name:                "Heuristic over 100 clamped",
			aiConfidence:        50,
			heuristicConfidence: 150,
			description:         "Should cap at 100",
		},
		{
			name:                "Both negative clamped to zero",
			aiConfidence:        -50,
			heuristicConfidence: -30,
			description:         "Should result in 0",
		},
		{
			name:                "Both over 100 clamped",
			aiConfidence:        200,
			heuristicConfidence: 150,
			description:         "Should result in 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.CalculateHybridConfidence(tt.aiConfidence, tt.heuristicConfidence)

			// Result should always be 0-100
			if result < 0 || result > 100 {
				t.Errorf("CalculateHybridConfidence(%d, %d) = %d, outside valid range [0, 100]",
					tt.aiConfidence, tt.heuristicConfidence, result)
			}
		})
	}
}

// TestValidateConfidenceScore verifies score validation
func TestValidateConfidenceScore(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	calc := NewConfidenceCalculator()

	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{
			name:     "Valid score in range",
			input:    75,
			expected: 75,
		},
		{
			name:     "Zero score",
			input:    0,
			expected: 0,
		},
		{
			name:     "Maximum score",
			input:    100,
			expected: 100,
		},
		{
			name:     "Negative score clamped to zero",
			input:    -50,
			expected: 0,
		},
		{
			name:     "Above maximum clamped to 100",
			input:    150,
			expected: 100,
		},
		{
			name:     "Slightly above maximum",
			input:    101,
			expected: 100,
		},
		{
			name:     "Slightly below minimum",
			input:    -1,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.ValidateConfidenceScore(tt.input)
			if result != tt.expected {
				t.Errorf("ValidateConfidenceScore(%d) = %d, expected %d",
					tt.input, result, tt.expected)
			}
		})
	}
}

// TestHybridConfidence_Integration verifies end-to-end hybrid confidence
func TestHybridConfidence_Integration(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	calc := NewConfidenceCalculator()

	// Simulate a scenario where AI has high confidence but heuristic is lower
	event := types.Event{
		ID:        "test-event",
		SourceID:  string(types.SourceTypeGit),
		Timestamp: time.Now().Add(-45 * 24 * time.Hour), // 45 days ago
		EventType: types.EventTypeCommit,
		Title:     "Security update",
		Content:   "Implemented authentication changes",
	}

	// Calculate heuristic confidence: 1 match (40) + recency (10) + git source (5) = 55
	heuristicScore := calc.Calculate(event, 1, string(types.SourceTypeGit))

	if heuristicScore != 55 {
		t.Errorf("Heuristic score = %d, expected 55", heuristicScore)
	}

	// Simulate AI confidence (AI analyzed code and found strong evidence)
	aiScore := 92

	// Calculate hybrid: (92*0.7 + 55*0.3) = 64.4 + 16.5 = 80.9 → 80
	hybridScore := calc.CalculateHybridConfidence(aiScore, heuristicScore)

	if hybridScore != 80 {
		t.Errorf("Hybrid confidence = %d, expected 80", hybridScore)
	}

	// Verify the hybrid score is weighted toward AI
	if hybridScore < heuristicScore {
		t.Error("Hybrid score should be higher than heuristic when AI has high confidence")
	}
}
