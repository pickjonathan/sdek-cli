package types

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateEvent(t *testing.T) {
	tests := []struct {
		name    string
		event   *Event
		wantErr bool
	}{
		{
			name:    "nil event",
			event:   nil,
			wantErr: true,
		},
		{
			name: "valid commit event",
			event: &Event{
				ID:        uuid.New().String(),
				SourceID:  SourceTypeGit,
				Timestamp: time.Now(),
				EventType: EventTypeCommit,
				Title:     "Add authentication middleware",
				Content:   "Implement JWT-based authentication",
				Author:    "Bob Martinez",
				Metadata:  make(map[string]interface{}),
			},
			wantErr: false,
		},
		{
			name: "invalid UUID",
			event: &Event{
				ID:        "invalid-uuid",
				SourceID:  SourceTypeGit,
				Timestamp: time.Now(),
				EventType: EventTypeCommit,
				Title:     "Test",
				Author:    "Test",
				Metadata:  make(map[string]interface{}),
			},
			wantErr: true,
		},
		{
			name: "empty source ID",
			event: &Event{
				ID:        uuid.New().String(),
				SourceID:  "",
				Timestamp: time.Now(),
				EventType: EventTypeCommit,
				Title:     "Test",
				Author:    "Test",
				Metadata:  make(map[string]interface{}),
			},
			wantErr: true,
		},
		{
			name: "timestamp too old",
			event: &Event{
				ID:        uuid.New().String(),
				SourceID:  SourceTypeGit,
				Timestamp: time.Now().AddDate(0, 0, -100),
				EventType: EventTypeCommit,
				Title:     "Test",
				Author:    "Test",
				Metadata:  make(map[string]interface{}),
			},
			wantErr: true,
		},
		{
			name: "invalid event type",
			event: &Event{
				ID:        uuid.New().String(),
				SourceID:  SourceTypeGit,
				Timestamp: time.Now(),
				EventType: "invalid",
				Title:     "Test",
				Author:    "Test",
				Metadata:  make(map[string]interface{}),
			},
			wantErr: true,
		},
		{
			name: "empty title",
			event: &Event{
				ID:        uuid.New().String(),
				SourceID:  SourceTypeGit,
				Timestamp: time.Now(),
				EventType: EventTypeCommit,
				Title:     "",
				Author:    "Test",
				Metadata:  make(map[string]interface{}),
			},
			wantErr: true,
		},
		{
			name: "title too long",
			event: &Event{
				ID:        uuid.New().String(),
				SourceID:  SourceTypeGit,
				Timestamp: time.Now(),
				EventType: EventTypeCommit,
				Title:     string(make([]byte, 201)),
				Author:    "Test",
				Metadata:  make(map[string]interface{}),
			},
			wantErr: true,
		},
		{
			name: "content too long",
			event: &Event{
				ID:        uuid.New().String(),
				SourceID:  SourceTypeGit,
				Timestamp: time.Now(),
				EventType: EventTypeCommit,
				Title:     "Test",
				Content:   string(make([]byte, 10001)),
				Author:    "Test",
				Metadata:  make(map[string]interface{}),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEvent(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewEvent(t *testing.T) {
	event := NewEvent(SourceTypeGit, EventTypeCommit, "Test commit", "Alice")

	// Validate UUID format
	if _, err := uuid.Parse(event.ID); err != nil {
		t.Errorf("expected valid UUID, got error: %v", err)
	}

	if event.SourceID != SourceTypeGit {
		t.Errorf("expected SourceID %s, got %s", SourceTypeGit, event.SourceID)
	}

	if event.EventType != EventTypeCommit {
		t.Errorf("expected EventType %s, got %s", EventTypeCommit, event.EventType)
	}

	if event.Title != "Test commit" {
		t.Errorf("expected Title 'Test commit', got %s", event.Title)
	}

	if event.Author != "Alice" {
		t.Errorf("expected Author 'Alice', got %s", event.Author)
	}

	if event.Metadata == nil {
		t.Error("expected Metadata to be initialized")
	}
}

func TestEventMetadataHelpers(t *testing.T) {
	event := NewEvent(SourceTypeGit, EventTypeCommit, "Test", "Alice")

	// Test Git metadata
	event.AddGitMetadata("abc123", "main", 5)
	if event.Metadata["commit_sha"] != "abc123" {
		t.Errorf("expected commit_sha 'abc123', got %v", event.Metadata["commit_sha"])
	}
	if event.Metadata["branch"] != "main" {
		t.Errorf("expected branch 'main', got %v", event.Metadata["branch"])
	}
	if event.Metadata["files_changed"] != 5 {
		t.Errorf("expected files_changed 5, got %v", event.Metadata["files_changed"])
	}

	// Test Jira metadata
	event = NewEvent(SourceTypeJira, EventTypeTicket, "Test", "Bob")
	event.AddJiraMetadata("PROJ-123", "Done", "High")
	if event.Metadata["ticket_id"] != "PROJ-123" {
		t.Errorf("expected ticket_id 'PROJ-123', got %v", event.Metadata["ticket_id"])
	}

	// Test Slack metadata
	event = NewEvent(SourceTypeSlack, EventTypeMessage, "Test", "Carol")
	event.AddSlackMetadata("#general", "1234.5678", 10)
	if event.Metadata["channel"] != "#general" {
		t.Errorf("expected channel '#general', got %v", event.Metadata["channel"])
	}

	// Test CICD metadata
	event = NewEvent(SourceTypeCICD, EventTypeBuild, "Test", "System")
	event.AddCICDMetadata("pipeline-001", "success", 120)
	if event.Metadata["pipeline_id"] != "pipeline-001" {
		t.Errorf("expected pipeline_id 'pipeline-001', got %v", event.Metadata["pipeline_id"])
	}

	// Test Docs metadata
	event = NewEvent(SourceTypeDocs, EventTypeDocumentChange, "Test", "Alice")
	event.AddDocsMetadata("/docs/api.md", "update", "Bob")
	if event.Metadata["file_path"] != "/docs/api.md" {
		t.Errorf("expected file_path '/docs/api.md', got %v", event.Metadata["file_path"])
	}
}
