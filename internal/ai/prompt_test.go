package ai

import (
	"strings"
	"testing"
	"time"
)

func TestPromptGenerator_GenerateSystemPrompt(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	pg := NewPromptGenerator()
	prompt := pg.GenerateSystemPrompt()

	// Check key content
	if !strings.Contains(prompt, "compliance analyst") {
		t.Error("Expected 'compliance analyst' in system prompt")
	}
	if !strings.Contains(prompt, "SOC 2") {
		t.Error("Expected 'SOC 2' in system prompt")
	}
	if !strings.Contains(prompt, "Confidence scoring") {
		t.Error("Expected 'Confidence scoring' in system prompt")
	}

	// Check not empty
	if len(prompt) < 100 {
		t.Errorf("Expected substantial system prompt, got %d chars", len(prompt))
	}
}

func TestPromptGenerator_GenerateUserPrompt(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	pg := NewPromptGenerator()

	req := &AnalysisRequest{
		ControlID:     "SOC2-CC1.1",
		ControlName:   "COSO Principle 1",
		Framework:     "SOC2",
		PolicyExcerpt: "The entity demonstrates a commitment to integrity and ethical values.",
		Events: []AnalysisEvent{
			{
				EventID:     "event-1",
				Source:      "git",
				EventType:   "commit",
				Description: "Added code of conduct",
				Content:     "feat: Add CODE_OF_CONDUCT.md with ethics guidelines",
				Timestamp:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			},
			{
				EventID:     "event-2",
				Source:      "jira",
				EventType:   "ticket",
				Description: "Ethics training ticket",
				Content:     "PROJ-123: Implement mandatory ethics training for all staff",
				Timestamp:   time.Date(2024, 1, 20, 14, 30, 0, 0, time.UTC),
			},
		},
	}

	prompt := pg.GenerateUserPrompt(req)

	// Check control info
	if !strings.Contains(prompt, "SOC2-CC1.1") {
		t.Error("Expected control ID in prompt")
	}
	if !strings.Contains(prompt, "COSO Principle 1") {
		t.Error("Expected control name in prompt")
	}
	if !strings.Contains(prompt, "SOC2") {
		t.Error("Expected framework in prompt")
	}

	// Check policy
	if !strings.Contains(prompt, "integrity and ethical values") {
		t.Error("Expected policy excerpt in prompt")
	}

	// Check events
	if !strings.Contains(prompt, "event-1") {
		t.Error("Expected event-1 in prompt")
	}
	if !strings.Contains(prompt, "event-2") {
		t.Error("Expected event-2 in prompt")
	}
	if !strings.Contains(prompt, "CODE_OF_CONDUCT.md") {
		t.Error("Expected event content in prompt")
	}

	// Check structure
	if !strings.Contains(prompt, "## Control Policy") {
		t.Error("Expected '## Control Policy' section")
	}
	if !strings.Contains(prompt, "## Events to Analyze") {
		t.Error("Expected '## Events to Analyze' section")
	}
	if !strings.Contains(prompt, "## Analysis Required") {
		t.Error("Expected '## Analysis Required' section")
	}
}

func TestPromptGenerator_GenerateUserPrompt_NoEvents(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	pg := NewPromptGenerator()

	req := &AnalysisRequest{
		ControlID:     "SOC2-CC1.1",
		ControlName:   "COSO Principle 1",
		Framework:     "SOC2",
		PolicyExcerpt: "Test policy",
		Events:        []AnalysisEvent{}, // Empty
	}

	prompt := pg.GenerateUserPrompt(req)

	// Should handle empty events gracefully
	if !strings.Contains(prompt, "No events provided") || !strings.Contains(prompt, "Events to Analyze") {
		t.Error("Expected 'No events' message when events are empty")
	}
}

func TestPromptGenerator_GeneratePrompts(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	pg := NewPromptGenerator()

	req := &AnalysisRequest{
		ControlID:     "SOC2-CC1.1",
		ControlName:   "Test Control",
		Framework:     "SOC2",
		PolicyExcerpt: "Test policy",
		Events: []AnalysisEvent{
			{
				EventID:   "event-1",
				Source:    "git",
				EventType: "commit",
				Content:   "test",
				Timestamp: time.Now(),
			},
		},
	}

	system, user := pg.GeneratePrompts(req)

	// Check both prompts generated
	if len(system) == 0 {
		t.Error("Expected non-empty system prompt")
	}
	if len(user) == 0 {
		t.Error("Expected non-empty user prompt")
	}

	// Check they're different
	if system == user {
		t.Error("Expected system and user prompts to be different")
	}

	// Check system prompt characteristics
	if !strings.Contains(system, "compliance analyst") {
		t.Error("Expected system prompt to have analyst context")
	}

	// Check user prompt has control info
	if !strings.Contains(user, "SOC2-CC1.1") {
		t.Error("Expected user prompt to have control ID")
	}
}

func TestPromptGenerator_EventFormatting(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	pg := NewPromptGenerator()

	req := &AnalysisRequest{
		ControlID:     "TEST-1",
		ControlName:   "Test",
		Framework:     "TEST",
		PolicyExcerpt: "Test",
		Events: []AnalysisEvent{
			{
				EventID:     "evt-123",
				Source:      "github",
				EventType:   "pr",
				Description: "Pull request merged",
				Content:     "PR #42: Add security headers\n\nImplemented CSP and HSTS headers",
				Timestamp:   time.Date(2024, 3, 15, 9, 30, 0, 0, time.UTC),
			},
		},
	}

	prompt := pg.GenerateUserPrompt(req)

	// Check event details formatted correctly
	if !strings.Contains(prompt, "evt-123") {
		t.Error("Expected event ID")
	}
	if !strings.Contains(prompt, "github") {
		t.Error("Expected source")
	}
	if !strings.Contains(prompt, "pr") {
		t.Error("Expected event type")
	}
	if !strings.Contains(prompt, "Pull request merged") {
		t.Error("Expected description")
	}
	if !strings.Contains(prompt, "2024-03-15") {
		t.Error("Expected formatted timestamp")
	}
	if !strings.Contains(prompt, "CSP and HSTS") {
		t.Error("Expected content")
	}

	// Check code block formatting
	if !strings.Contains(prompt, "```") {
		t.Error("Expected code block markers for content")
	}
}

func TestPromptGenerator_MultipleEvents(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	pg := NewPromptGenerator()

	// Create request with 5 events
	events := make([]AnalysisEvent, 5)
	for i := 0; i < 5; i++ {
		events[i] = AnalysisEvent{
			EventID:   string(rune('a' + i)),
			Source:    "test",
			EventType: "test",
			Content:   "test content",
			Timestamp: time.Now(),
		}
	}

	req := &AnalysisRequest{
		ControlID:     "TEST-1",
		ControlName:   "Test",
		Framework:     "TEST",
		PolicyExcerpt: "Test",
		Events:        events,
	}

	prompt := pg.GenerateUserPrompt(req)

	// Check all events included
	for i := 0; i < 5; i++ {
		eventID := string(rune('a' + i))
		if !strings.Contains(prompt, eventID) {
			t.Errorf("Expected event %s in prompt", eventID)
		}
	}

	// Check event numbering
	if !strings.Contains(prompt, "Event 1:") {
		t.Error("Expected 'Event 1:' in prompt")
	}
	if !strings.Contains(prompt, "Event 5:") {
		t.Error("Expected 'Event 5:' in prompt")
	}
}

func TestPromptGenerator_SpecialCharacters(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	pg := NewPromptGenerator()

	req := &AnalysisRequest{
		ControlID:     "TEST-1",
		ControlName:   "Test with \"quotes\" and <tags>",
		Framework:     "TEST",
		PolicyExcerpt: "Policy with special chars: & | $ #",
		Events: []AnalysisEvent{
			{
				EventID:     "event-1",
				Source:      "test",
				EventType:   "test",
				Description: "Contains <script>alert('xss')</script>",
				Content:     "Content with\nnewlines\nand\ttabs",
				Timestamp:   time.Now(),
			},
		},
	}

	prompt := pg.GenerateUserPrompt(req)

	// Check special characters preserved (not escaped/removed)
	if !strings.Contains(prompt, "\"quotes\"") {
		t.Error("Expected quotes preserved")
	}
	if !strings.Contains(prompt, "<tags>") {
		t.Error("Expected tags preserved")
	}
	if !strings.Contains(prompt, "& | $ #") {
		t.Error("Expected special chars preserved")
	}
	if !strings.Contains(prompt, "newlines") && strings.Contains(prompt, "tabs") {
		t.Error("Expected newlines and tabs preserved in content")
	}
}
