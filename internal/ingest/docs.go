package ingest

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// DocsGenerator generates documentation change events
type DocsGenerator struct {
	*BaseGenerator
}

// NewDocsGenerator creates a new documentation event generator
func NewDocsGenerator(seed int64) *DocsGenerator {
	return &DocsGenerator{
		BaseGenerator: NewBaseGenerator(seed),
	}
}

// GetSourceType returns the source type
func (d *DocsGenerator) GetSourceType() string {
	return string(types.SourceTypeDocs)
}

// Generate creates documentation change events
func (d *DocsGenerator) Generate(seed int64, count int) ([]types.Event, error) {
	// Validate event count
	if err := ValidateEventCount(count); err != nil {
		return nil, err
	}

	// Reinitialize with the provided seed for deterministic generation
	d.BaseGenerator = NewBaseGenerator(seed)

	events := make([]types.Event, count)

	// Docs-specific data
	docPaths := []string{
		"docs/security/authentication.md",
		"docs/security/encryption.md",
		"docs/compliance/soc2.md",
		"docs/compliance/iso27001.md",
		"docs/policies/access-control.md",
		"docs/policies/data-retention.md",
		"docs/runbooks/incident-response.md",
		"docs/guides/security-best-practices.md",
	}

	changeTypes := []string{"created", "updated", "reviewed", "approved", "archived"}

	for i := 0; i < count; i++ {
		// Generate document attributes
		docPath := d.RandomElement(docPaths)
		changeType := d.RandomElement(changeTypes)
		author := d.RandomElement(AuthorNames)
		reviewer := d.RandomElement(AuthorNames)

		// Ensure reviewer is different from author
		for reviewer == author {
			reviewer = d.RandomElement(AuthorNames)
		}

		// Generate title and content
		title := fmt.Sprintf("%s: %s", changeType, docPath)
		content := d.generateDocContent(changeType)

		if d.RandomBool(0.5) {
			keyword := d.RandomElement(SecurityKeywords)
			content = fmt.Sprintf("%s. Covers %s requirements.", content, keyword)
		}

		// Create metadata
		metadata := map[string]interface{}{
			"file_path":   docPath,
			"change_type": changeType,
			"reviewer":    reviewer,
			"version":     fmt.Sprintf("v%d.%d", d.RandomInt(1, 5), d.RandomInt(0, 9)),
			"word_count":  d.RandomInt(100, 5000),
			"is_policy":   d.RandomBool(0.4),
		}

		events[i] = types.Event{
			ID:        uuid.New().String(),
			SourceID:  string(types.SourceTypeDocs),
			Timestamp: d.RandomTimestamp(),
			EventType: types.EventTypeDocumentChange,
			Title:     title,
			Content:   content,
			Author:    author,
			Metadata:  metadata,
		}
	}

	return events, nil
}

// generateDocContent creates realistic documentation content
func (d *DocsGenerator) generateDocContent(changeType string) string {
	createdMessages := []string{
		"New security policy document created",
		"Added compliance framework documentation",
		"Created incident response runbook",
		"New authentication guide published",
	}

	updatedMessages := []string{
		"Updated security controls documentation",
		"Revised access control policies",
		"Refreshed compliance requirements",
		"Updated encryption standards",
		"Modified data retention policies",
	}

	reviewedMessages := []string{
		"Security documentation reviewed for accuracy",
		"Compliance docs reviewed by legal team",
		"Policy documentation peer reviewed",
		"Technical review completed",
	}

	approvedMessages := []string{
		"Security policy approved by CISO",
		"Compliance documentation approved",
		"Runbook approved for production use",
		"Policy changes formally approved",
	}

	switch changeType {
	case "created":
		return d.RandomElement(createdMessages)
	case "updated":
		return d.RandomElement(updatedMessages)
	case "reviewed":
		return d.RandomElement(reviewedMessages)
	case "approved":
		return d.RandomElement(approvedMessages)
	default:
		return "Documentation change"
	}
}
