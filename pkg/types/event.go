package types

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Event represents a discrete signal from a source
type Event struct {
	ID        string                 `json:"id"`
	SourceID  string                 `json:"source_id"`
	Timestamp time.Time              `json:"timestamp"`
	EventType string                 `json:"event_type"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Author    string                 `json:"author"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Event type constants
const (
	EventTypeCommit         = "commit"
	EventTypeTicket         = "ticket"
	EventTypeMessage        = "message"
	EventTypeBuild          = "build"
	EventTypeDocumentChange = "document_change"
)

// ValidateEvent checks if an Event meets all validation rules
func ValidateEvent(e *Event) error {
	if e == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Validate ID is valid UUID
	if _, err := uuid.Parse(e.ID); err != nil {
		return fmt.Errorf("invalid event ID: must be valid UUID, got %s", e.ID)
	}

	// Validate source ID is not empty
	if e.SourceID == "" {
		return fmt.Errorf("source_id cannot be empty")
	}

	// Validate timestamp is within 90 days
	if time.Since(e.Timestamp) > 90*24*time.Hour {
		return fmt.Errorf("timestamp must be within 90 days")
	}

	// Validate event type matches source type
	validTypes := []string{EventTypeCommit, EventTypeTicket, EventTypeMessage, EventTypeBuild, EventTypeDocumentChange}
	valid := false
	for _, t := range validTypes {
		if e.EventType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid event type: %s", e.EventType)
	}

	// Validate title
	if e.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if len(e.Title) > 200 {
		return fmt.Errorf("title cannot exceed 200 characters, got %d", len(e.Title))
	}

	// Validate content length if present
	if len(e.Content) > 10000 {
		return fmt.Errorf("content cannot exceed 10000 characters, got %d", len(e.Content))
	}

	return nil
}

// NewEvent creates a new Event with generated UUID
func NewEvent(sourceID, eventType, title, author string) *Event {
	return &Event{
		ID:        uuid.New().String(),
		SourceID:  sourceID,
		Timestamp: time.Now(),
		EventType: eventType,
		Title:     title,
		Author:    author,
		Content:   "",
		Metadata:  make(map[string]interface{}),
	}
}

// AddGitMetadata adds Git-specific metadata
func (e *Event) AddGitMetadata(commitSHA, branch string, filesChanged int) {
	e.Metadata["commit_sha"] = commitSHA
	e.Metadata["branch"] = branch
	e.Metadata["files_changed"] = filesChanged
}

// AddJiraMetadata adds Jira-specific metadata
func (e *Event) AddJiraMetadata(ticketID, status, priority string) {
	e.Metadata["ticket_id"] = ticketID
	e.Metadata["status"] = status
	e.Metadata["priority"] = priority
}

// AddSlackMetadata adds Slack-specific metadata
func (e *Event) AddSlackMetadata(channel, threadID string, reactions int) {
	e.Metadata["channel"] = channel
	e.Metadata["thread_id"] = threadID
	e.Metadata["reactions"] = reactions
}

// AddCICDMetadata adds CI/CD-specific metadata
func (e *Event) AddCICDMetadata(pipelineID, status string, duration int) {
	e.Metadata["pipeline_id"] = pipelineID
	e.Metadata["status"] = status
	e.Metadata["duration"] = duration
}

// AddDocsMetadata adds documentation-specific metadata
func (e *Event) AddDocsMetadata(filePath, changeType, reviewer string) {
	e.Metadata["file_path"] = filePath
	e.Metadata["change_type"] = changeType
	e.Metadata["reviewer"] = reviewer
}
