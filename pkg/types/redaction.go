package types

import (
	"time"
)

// RedactionMap represents redacted content mapping.
// WARNING: Stored in-memory only, NEVER persisted to disk or sent to AI providers.
type RedactionMap struct {
	// entries maps original content hash to redaction metadata
	// This field is intentionally unexported and never serialized
	entries map[string]RedactionEntry

	// Statistics (safe to log/export)
	TotalRedactions int             `json:"total_redactions"`
	RedactionTypes  []RedactionType `json:"redaction_types"`
}

// RedactionEntry represents a single redacted piece of content.
type RedactionEntry struct {
	OriginalHash string        `json:"-"` // SHA256 of original text (NEVER exported)
	Placeholder  string        `json:"placeholder"`
	Type         RedactionType `json:"type"`
	Position     int           `json:"position"` // Character offset in original text
	Timestamp    time.Time     `json:"timestamp"`
}

// RedactionType categorizes the type of content that was redacted.
type RedactionType string

const (
	RedactionPII    RedactionType = "pii"    // Email, phone, IP
	RedactionSecret RedactionType = "secret" // API keys, tokens
)

// NewRedactionMap creates a new empty redaction map.
func NewRedactionMap() *RedactionMap {
	return &RedactionMap{
		entries:         make(map[string]RedactionEntry),
		TotalRedactions: 0,
		RedactionTypes:  []RedactionType{},
	}
}

// AddEntry adds a redaction entry to the map.
func (rm *RedactionMap) AddEntry(hash string, entry RedactionEntry) {
	rm.entries[hash] = entry
	rm.TotalRedactions++

	// Add type if not already present
	typeExists := false
	for _, t := range rm.RedactionTypes {
		if t == entry.Type {
			typeExists = true
			break
		}
	}
	if !typeExists {
		rm.RedactionTypes = append(rm.RedactionTypes, entry.Type)
	}
}

// GetEntry retrieves a redaction entry by hash (for internal use only).
func (rm *RedactionMap) GetEntry(hash string) (RedactionEntry, bool) {
	entry, exists := rm.entries[hash]
	return entry, exists
}

// HasType checks if a specific redaction type was used.
func (rm *RedactionMap) HasType(redactionType RedactionType) bool {
	for _, t := range rm.RedactionTypes {
		if t == redactionType {
			return true
		}
	}
	return false
}
