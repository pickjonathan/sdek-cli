package ingest

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// JiraGenerator generates Jira ticket events
type JiraGenerator struct {
	*BaseGenerator
}

// NewJiraGenerator creates a new Jira event generator
func NewJiraGenerator(seed int64) *JiraGenerator {
	return &JiraGenerator{
		BaseGenerator: NewBaseGenerator(seed),
	}
}

// GetSourceType returns the source type
func (j *JiraGenerator) GetSourceType() string {
	return string(types.SourceTypeJira)
}

// Generate creates Jira ticket events
func (j *JiraGenerator) Generate(seed int64, count int) ([]types.Event, error) {
	// Validate event count
	if err := ValidateEventCount(count); err != nil {
		return nil, err
	}

	// Reinitialize with the provided seed for deterministic generation
	j.BaseGenerator = NewBaseGenerator(seed)

	events := make([]types.Event, count)

	// Jira-specific data
	ticketTypes := []string{"Bug", "Task", "Story", "Epic"}
	statuses := []string{"Open", "In Progress", "In Review", "Resolved", "Closed"}
	priorities := []string{"Low", "Medium", "High", "Critical"}
	projects := []string{"SEC", "COMP", "AUD", "DEV"}

	for i := 0; i < count; i++ {
		// Generate ticket ID
		project := j.RandomElement(projects)
		ticketID := fmt.Sprintf("%s-%d", project, j.RandomInt(1000, 9999))

		// Select attributes
		ticketType := j.RandomElement(ticketTypes)
		status := j.RandomElement(statuses)
		priority := j.RandomElement(priorities)
		author := j.RandomElement(AuthorNames)
		assignee := j.RandomElement(AuthorNames)

		// Generate title
		title := j.generateTicketTitle(ticketType)

		// Generate content
		content := j.generateTicketContent(ticketType)
		if j.RandomBool(0.4) {
			keyword := j.RandomElement(SecurityKeywords)
			content = fmt.Sprintf("%s. Related to %s.", content, keyword)
		}

		// Create metadata
		metadata := map[string]interface{}{
			"ticket_id": ticketID,
			"type":      ticketType,
			"status":    status,
			"priority":  priority,
			"assignee":  assignee,
			"labels":    j.RandomSubset(SecurityKeywords, 0, 3),
		}

		events[i] = types.Event{
			ID:        uuid.New().String(),
			SourceID:  string(types.SourceTypeJira),
			Timestamp: j.RandomTimestamp(),
			EventType: types.EventTypeTicket,
			Title:     fmt.Sprintf("[%s] %s", ticketID, title),
			Content:   content,
			Author:    author,
			Metadata:  metadata,
		}
	}

	return events, nil
}

// generateTicketTitle creates a realistic ticket title
func (j *JiraGenerator) generateTicketTitle(ticketType string) string {
	bugTitles := []string{
		"Authentication fails for OAuth users",
		"Memory leak in batch processor",
		"XSS vulnerability in user input",
		"Race condition in concurrent requests",
		"SQL injection in search endpoint",
	}

	taskTitles := []string{
		"Update security documentation",
		"Configure audit logging",
		"Implement rate limiting",
		"Add encryption for sensitive data",
		"Setup automated backups",
	}

	storyTitles := []string{
		"As a user, I need two-factor authentication",
		"As an admin, I need to view audit logs",
		"As a developer, I need API access control",
		"As a compliance officer, I need compliance reports",
		"As a security engineer, I need threat detection",
	}

	switch ticketType {
	case "Bug":
		return j.RandomElement(bugTitles)
	case "Task":
		return j.RandomElement(taskTitles)
	case "Story":
		return j.RandomElement(storyTitles)
	default:
		return "General ticket"
	}
}

// generateTicketContent creates realistic ticket content
func (j *JiraGenerator) generateTicketContent(ticketType string) string {
	contents := []string{
		"This issue requires immediate attention for security compliance",
		"Investigation shows this affects user authentication flow",
		"Implementation needed to meet SOC2 requirements",
		"Discovered during security audit review",
		"Customer reported this security concern",
		"Automated scan detected this vulnerability",
		"Required for ISO 27001 certification",
		"Part of quarterly security assessment",
	}
	return j.RandomElement(contents)
}
