package types

import (
	"time"
)

// EvidenceBundle represents a collection of evidence events from various sources.
// This is the normalized schema for evidence collected via MCP connectors.
type EvidenceBundle struct {
	Events []EvidenceEvent `json:"events"`
}

// EvidenceEvent represents a single piece of evidence from a source.
// This is a normalized format that MCP connector outputs are converted to.
type EvidenceEvent struct {
	ID        string                 `json:"id"`
	Source    string                 `json:"source"`    // "github", "jira", "aws", etc.
	Type      string                 `json:"type"`      // "commit", "ticket", "log", etc.
	Timestamp time.Time              `json:"timestamp"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
