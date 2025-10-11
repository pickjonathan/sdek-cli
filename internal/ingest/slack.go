package ingest

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// SlackGenerator generates Slack message events
type SlackGenerator struct {
	*BaseGenerator
}

// NewSlackGenerator creates a new Slack event generator
func NewSlackGenerator(seed int64) *SlackGenerator {
	return &SlackGenerator{
		BaseGenerator: NewBaseGenerator(seed),
	}
}

// GetSourceType returns the source type
func (s *SlackGenerator) GetSourceType() string {
	return string(types.SourceTypeSlack)
}

// Generate creates Slack message events
func (s *SlackGenerator) Generate(seed int64, count int) ([]types.Event, error) {
	// Validate event count
	if err := ValidateEventCount(count); err != nil {
		return nil, err
	}

	// Reinitialize with the provided seed for deterministic generation
	s.BaseGenerator = NewBaseGenerator(seed)

	events := make([]types.Event, count)

	// Slack-specific data
	channels := []string{
		"#security",
		"#compliance",
		"#engineering",
		"#devops",
		"#incidents",
		"#audit-logs",
	}

	reactions := []string{
		"thumbsup",
		"eyes",
		"white_check_mark",
		"warning",
		"lock",
		"shield",
	}

	for i := 0; i < count; i++ {
		// Select attributes
		channel := s.RandomElement(channels)
		author := s.RandomElement(AuthorNames)

		// Generate message
		message := s.generateMessage()
		if s.RandomBool(0.3) {
			keyword := s.RandomElement(SecurityKeywords)
			message = fmt.Sprintf("%s - discussed %s", message, keyword)
		}

		// Randomly create threads
		var threadID string
		var replyCount int
		if s.RandomBool(0.4) {
			threadID = uuid.New().String()
			replyCount = s.RandomInt(1, 10)
		}

		// Create metadata
		metadata := map[string]interface{}{
			"channel":      channel,
			"thread_id":    threadID,
			"reply_count":  replyCount,
			"reactions":    s.RandomSubset(reactions, 0, 3),
			"has_mentions": s.RandomBool(0.5),
		}

		events[i] = types.Event{
			ID:        uuid.New().String(),
			SourceID:  string(types.SourceTypeSlack),
			Timestamp: s.RandomTimestamp(),
			EventType: types.EventTypeMessage,
			Title:     fmt.Sprintf("Message in %s", channel),
			Content:   message,
			Author:    author,
			Metadata:  metadata,
		}
	}

	return events, nil
}

// generateMessage creates a realistic Slack message
func (s *SlackGenerator) generateMessage() string {
	messages := []string{
		"Security patch deployed successfully",
		"Need review on access control implementation",
		"Audit log analysis shows anomaly",
		"Compliance checklist updated for Q4",
		"Security training scheduled for next week",
		"Found potential vulnerability in API",
		"Code review completed - security approved",
		"Backup verification passed all checks",
		"Incident response plan updated",
		"SSL certificate renewal reminder",
		"MFA enrollment now mandatory",
		"Security scan results are clean",
		"Policy update requires acknowledgment",
		"Access request approved for new team member",
		"Encryption keys rotated successfully",
	}
	return s.RandomElement(messages)
}
