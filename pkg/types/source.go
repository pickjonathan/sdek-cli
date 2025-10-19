package types

import (
	"fmt"
	"time"
)

// Source represents a data integration point (Git, Jira, Slack, CI/CD, Docs).
type Source struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	LastSync   time.Time `json:"last_sync"`
	EventCount int       `json:"event_count"`
	Enabled    bool      `json:"enabled"`
}

// Source type constants
const (
	SourceTypeGit   = "git"
	SourceTypeJira  = "jira"
	SourceTypeSlack = "slack"
	SourceTypeCICD  = "cicd"
	SourceTypeDocs  = "docs"
)

// ValidSourceTypes contains all valid source type identifiers
var ValidSourceTypes = []string{
	SourceTypeGit,
	SourceTypeJira,
	SourceTypeSlack,
	SourceTypeCICD,
	SourceTypeDocs,
}

// ValidateSource checks if a Source meets all validation rules
func ValidateSource(s *Source) error {
	if s == nil {
		return fmt.Errorf("source cannot be nil")
	}

	// Validate ID is one of valid types
	valid := false
	for _, t := range ValidSourceTypes {
		if s.ID == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid source ID: %s, must be one of %v", s.ID, ValidSourceTypes)
	}

	// Validate event count
	if s.EventCount < 10 || s.EventCount > 50 {
		return fmt.Errorf("event count must be between 10 and 50, got %d", s.EventCount)
	}

	// Validate last sync is within 90 days
	if time.Since(s.LastSync) > 90*24*time.Hour {
		return fmt.Errorf("last sync must be within 90 days")
	}

	return nil
}

// NewSource creates a new Source with default values
func NewSource(id, name, sourceType string) *Source {
	return &Source{
		ID:         id,
		Name:       name,
		Type:       sourceType,
		Status:     "simulated",
		LastSync:   time.Now(),
		EventCount: 0,
		Enabled:    true,
	}
}
