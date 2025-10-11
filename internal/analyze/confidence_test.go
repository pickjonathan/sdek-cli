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
		name         string
		event        types.Event
		matchCount   int
		sourceType   string
		minScore     int
		maxScore     int
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
		{4, 80}, // Capped at 80
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
