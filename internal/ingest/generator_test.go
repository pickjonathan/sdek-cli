package ingest

import (
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// TestGitGenerator_Deterministic verifies that the same seed produces the same events
func TestGitGenerator_Deterministic(t *testing.T) {
	gen := NewGitGenerator(12345)

	events1, err1 := gen.Generate(12345, 10)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	events2, err2 := gen.Generate(12345, 10)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	if len(events1) != len(events2) {
		t.Errorf("Event count mismatch: %d vs %d", len(events1), len(events2))
	}

	for i := range events1 {
		// IDs will be different (UUIDs), but other fields should match
		if events1[i].SourceID != events2[i].SourceID {
			t.Errorf("Event %d: SourceID mismatch", i)
		}
		if events1[i].EventType != events2[i].EventType {
			t.Errorf("Event %d: EventType mismatch", i)
		}
		if events1[i].Author != events2[i].Author {
			t.Errorf("Event %d: Author mismatch: %s vs %s", i, events1[i].Author, events2[i].Author)
		}
		if events1[i].Title != events2[i].Title {
			t.Errorf("Event %d: Title mismatch", i)
		}
	}
}

// TestGitGenerator_EventCount verifies event count boundaries
func TestGitGenerator_EventCount(t *testing.T) {
	gen := NewGitGenerator(12345)

	tests := []struct {
		name      string
		count     int
		expectErr bool
	}{
		{"Too Few", 5, true},
		{"Min Valid", 10, false},
		{"Mid Range", 30, false},
		{"Max Valid", 50, false},
		{"Too Many", 51, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := gen.Generate(12345, tt.count)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error for count %d, got nil", tt.count)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(events) != tt.count {
					t.Errorf("Expected %d events, got %d", tt.count, len(events))
				}
			}
		})
	}
}

// TestGitGenerator_TimestampRange verifies timestamps are within 90 days
func TestGitGenerator_TimestampRange(t *testing.T) {
	gen := NewGitGenerator(12345)
	events, err := gen.Generate(12345, 20)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	now := time.Now()
	ninetyDaysAgo := now.AddDate(0, 0, -90)

	for i, event := range events {
		if event.Timestamp.After(now) {
			t.Errorf("Event %d: timestamp %v is in the future", i, event.Timestamp)
		}
		if event.Timestamp.Before(ninetyDaysAgo) {
			t.Errorf("Event %d: timestamp %v is more than 90 days old", i, event.Timestamp)
		}
	}
}

// TestGitGenerator_Metadata verifies Git-specific metadata
func TestGitGenerator_Metadata(t *testing.T) {
	gen := NewGitGenerator(12345)
	events, err := gen.Generate(12345, 10)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	for i, event := range events {
		if event.SourceID != string(types.SourceTypeGit) {
			t.Errorf("Event %d: wrong source ID: %s", i, event.SourceID)
		}

		if event.EventType != types.EventTypeCommit {
			t.Errorf("Event %d: wrong event type: %s", i, event.EventType)
		}

		// Check metadata fields
		metadata := event.Metadata
		if metadata == nil {
			t.Errorf("Event %d: metadata is nil", i)
			continue
		}

		requiredFields := []string{"commit_sha", "branch", "files_changed", "additions", "deletions"}
		for _, field := range requiredFields {
			if _, ok := metadata[field]; !ok {
				t.Errorf("Event %d: missing metadata field: %s", i, field)
			}
		}

		// Verify commit SHA format (40 hex characters)
		if sha, ok := metadata["commit_sha"].(string); ok {
			if len(sha) != 40 {
				t.Errorf("Event %d: invalid commit SHA length: %d", i, len(sha))
			}
		}

		// Verify author is set
		if event.Author == "" {
			t.Errorf("Event %d: author is empty", i)
		}
	}
}

// TestGitGenerator_GetSourceType verifies source type
func TestGitGenerator_GetSourceType(t *testing.T) {
	gen := NewGitGenerator(12345)
	sourceType := gen.GetSourceType()

	if sourceType != string(types.SourceTypeGit) {
		t.Errorf("Expected source type %s, got %s", types.SourceTypeGit, sourceType)
	}
}

// TestJiraGenerator_Deterministic verifies that the same seed produces the same events
func TestJiraGenerator_Deterministic(t *testing.T) {
	gen := NewJiraGenerator(54321)

	events1, err1 := gen.Generate(54321, 10)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	events2, err2 := gen.Generate(54321, 10)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	if len(events1) != len(events2) {
		t.Errorf("Event count mismatch: %d vs %d", len(events1), len(events2))
	}

	for i := range events1 {
		if events1[i].Author != events2[i].Author {
			t.Errorf("Event %d: Author mismatch: %s vs %s", i, events1[i].Author, events2[i].Author)
		}
	}
}

// TestJiraGenerator_Metadata verifies Jira-specific metadata
func TestJiraGenerator_Metadata(t *testing.T) {
	gen := NewJiraGenerator(54321)
	events, err := gen.Generate(54321, 10)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	for i, event := range events {
		if event.SourceID != string(types.SourceTypeJira) {
			t.Errorf("Event %d: wrong source ID: %s", i, event.SourceID)
		}

		if event.EventType != types.EventTypeTicket {
			t.Errorf("Event %d: wrong event type: %s", i, event.EventType)
		}

		metadata := event.Metadata
		if metadata == nil {
			t.Errorf("Event %d: metadata is nil", i)
			continue
		}

		requiredFields := []string{"ticket_id", "type", "status", "priority", "assignee", "labels"}
		for _, field := range requiredFields {
			if _, ok := metadata[field]; !ok {
				t.Errorf("Event %d: missing metadata field: %s", i, field)
			}
		}
	}
}

// TestSlackGenerator_Metadata verifies Slack-specific metadata
func TestSlackGenerator_Metadata(t *testing.T) {
	gen := NewSlackGenerator(11111)
	events, err := gen.Generate(11111, 10)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	for i, event := range events {
		if event.SourceID != string(types.SourceTypeSlack) {
			t.Errorf("Event %d: wrong source ID: %s", i, event.SourceID)
		}

		if event.EventType != types.EventTypeMessage {
			t.Errorf("Event %d: wrong event type: %s", i, event.EventType)
		}

		metadata := event.Metadata
		if metadata == nil {
			t.Errorf("Event %d: metadata is nil", i)
			continue
		}

		requiredFields := []string{"channel", "thread_id", "reply_count", "reactions", "has_mentions"}
		for _, field := range requiredFields {
			if _, ok := metadata[field]; !ok {
				t.Errorf("Event %d: missing metadata field: %s", i, field)
			}
		}
	}
}

// TestCICDGenerator_Metadata verifies CI/CD-specific metadata
func TestCICDGenerator_Metadata(t *testing.T) {
	gen := NewCICDGenerator(22222)
	events, err := gen.Generate(22222, 10)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	for i, event := range events {
		if event.SourceID != string(types.SourceTypeCICD) {
			t.Errorf("Event %d: wrong source ID: %s", i, event.SourceID)
		}

		if event.EventType != types.EventTypeBuild {
			t.Errorf("Event %d: wrong event type: %s", i, event.EventType)
		}

		metadata := event.Metadata
		if metadata == nil {
			t.Errorf("Event %d: metadata is nil", i)
			continue
		}

		requiredFields := []string{"pipeline_id", "pipeline_name", "status", "stage", "duration_secs", "branch"}
		for _, field := range requiredFields {
			if _, ok := metadata[field]; !ok {
				t.Errorf("Event %d: missing metadata field: %s", i, field)
			}
		}

		// Verify duration is reasonable (30-1800 seconds)
		if duration, ok := metadata["duration_secs"].(int); ok {
			if duration < 30 || duration > 1800 {
				t.Errorf("Event %d: duration %d out of range [30, 1800]", i, duration)
			}
		}
	}
}

// TestDocsGenerator_Metadata verifies Docs-specific metadata
func TestDocsGenerator_Metadata(t *testing.T) {
	gen := NewDocsGenerator(33333)
	events, err := gen.Generate(33333, 10)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	for i, event := range events {
		if event.SourceID != string(types.SourceTypeDocs) {
			t.Errorf("Event %d: wrong source ID: %s", i, event.SourceID)
		}

		if event.EventType != types.EventTypeDocumentChange {
			t.Errorf("Event %d: wrong event type: %s", i, event.EventType)
		}

		metadata := event.Metadata
		if metadata == nil {
			t.Errorf("Event %d: metadata is nil", i)
			continue
		}

		requiredFields := []string{"file_path", "change_type", "reviewer", "version", "word_count", "is_policy"}
		for _, field := range requiredFields {
			if _, ok := metadata[field]; !ok {
				t.Errorf("Event %d: missing metadata field: %s", i, field)
			}
		}

		// Verify reviewer is different from author
		if reviewer, ok := metadata["reviewer"].(string); ok {
			if reviewer == event.Author {
				t.Errorf("Event %d: reviewer should differ from author", i)
			}
		}
	}
}

// TestBaseGenerator_RandomTimestamp verifies timestamp generation
func TestBaseGenerator_RandomTimestamp(t *testing.T) {
	bg := NewBaseGenerator(99999)

	now := time.Now()
	ninetyDaysAgo := now.AddDate(0, 0, -90)

	// Generate multiple timestamps
	for i := 0; i < 100; i++ {
		ts := bg.RandomTimestamp()

		if ts.After(now) {
			t.Errorf("Timestamp %d is in the future: %v", i, ts)
		}
		if ts.Before(ninetyDaysAgo) {
			t.Errorf("Timestamp %d is too old: %v", i, ts)
		}
	}
}

// TestBaseGenerator_RandomInt verifies random integer generation
func TestBaseGenerator_RandomInt(t *testing.T) {
	bg := NewBaseGenerator(88888)

	tests := []struct {
		name string
		min  int
		max  int
	}{
		{"Small Range", 1, 10},
		{"Large Range", 100, 1000},
		{"Same Value", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 50; i++ {
				val := bg.RandomInt(tt.min, tt.max)
				if val < tt.min || val > tt.max {
					t.Errorf("Value %d out of range [%d, %d]", val, tt.min, tt.max)
				}
			}
		})
	}
}

// TestBaseGenerator_RandomString verifies random string selection
func TestBaseGenerator_RandomString(t *testing.T) {
	bg := NewBaseGenerator(77777)

	options := []string{"a", "b", "c"}
	seen := make(map[string]bool)

	for i := 0; i < 50; i++ {
		result := bg.RandomString(options)
		seen[result] = true

		found := false
		for _, opt := range options {
			if result == opt {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Result '%s' not in options", result)
		}
	}

	// With enough iterations, we should see multiple options
	if len(seen) < 2 {
		t.Errorf("Expected to see multiple options, only saw: %v", seen)
	}
}

// TestBaseGenerator_RandomSubset verifies subset generation
func TestBaseGenerator_RandomSubset(t *testing.T) {
	bg := NewBaseGenerator(66666)

	slice := []string{"a", "b", "c", "d", "e"}

	subset := bg.RandomSubset(slice, 2, 4)

	if len(subset) < 2 || len(subset) > 4 {
		t.Errorf("Subset size %d out of range [2, 4]", len(subset))
	}

	// Verify all elements are from original slice
	for _, elem := range subset {
		found := false
		for _, orig := range slice {
			if elem == orig {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Element '%s' not in original slice", elem)
		}
	}
}

// TestValidateEventCount verifies event count validation
func TestValidateEventCount(t *testing.T) {
	tests := []struct {
		name      string
		count     int
		expectErr bool
	}{
		{"Too Low", 5, true},
		{"Min Valid", 10, false},
		{"Mid Range", 30, false},
		{"Max Valid", 50, false},
		{"Too High", 51, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEventCount(tt.count)

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for count %d", tt.count)
			}

			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error for count %d: %v", tt.count, err)
			}
		})
	}
}
