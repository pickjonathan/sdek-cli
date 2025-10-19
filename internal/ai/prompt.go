package ai

import (
	"fmt"
	"strings"
)

// PromptGenerator creates system and user prompts for AI analysis
type PromptGenerator struct{}

// NewPromptGenerator creates a new prompt generator
func NewPromptGenerator() *PromptGenerator {
	return &PromptGenerator{}
}

// GenerateSystemPrompt creates the system message for AI context
func (pg *PromptGenerator) GenerateSystemPrompt() string {
	return `You are an expert compliance analyst specializing in mapping technical evidence to security and compliance controls.

Your task is to analyze development and operational events (commits, tickets, builds, messages, documentation) and determine if they provide evidence of compliance with specific controls from frameworks like SOC 2, ISO 27001, and PCI DSS.

Guidelines:
1. Evidence quality: Strong evidence directly demonstrates the control is implemented and operating effectively. Weak evidence is tangential or incomplete.
2. Confidence scoring: 
   - 90-100%: Multiple strong pieces of evidence with clear implementation
   - 70-89%: Good evidence but some gaps or assumptions
   - 50-69%: Moderate evidence with significant assumptions
   - 30-49%: Weak or indirect evidence
   - 0-29%: Little to no relevant evidence
3. Justification: Explain clearly how the evidence supports the control, referencing specific events.
4. Residual risks: Note any gaps, concerns, or areas needing additional evidence.

Be objective and conservative in your assessments. It's better to express uncertainty than to overstate evidence quality.`
}

// GenerateUserPrompt creates the user message with control and events
func (pg *PromptGenerator) GenerateUserPrompt(req *AnalysisRequest) string {
	var sb strings.Builder

	// Control header
	sb.WriteString(fmt.Sprintf("# Control Analysis Request\n\n"))
	sb.WriteString(fmt.Sprintf("**Control ID:** %s\n", req.ControlID))
	sb.WriteString(fmt.Sprintf("**Control Name:** %s\n", req.ControlName))
	sb.WriteString(fmt.Sprintf("**Framework:** %s\n\n", req.Framework))

	// Policy excerpt
	sb.WriteString("## Control Policy\n\n")
	sb.WriteString(req.PolicyExcerpt)
	sb.WriteString("\n\n")

	// Events to analyze
	sb.WriteString("## Events to Analyze\n\n")
	if len(req.Events) == 0 {
		sb.WriteString("*No events provided*\n")
	} else {
		for i, event := range req.Events {
			sb.WriteString(fmt.Sprintf("### Event %d (ID: %s)\n", i+1, event.EventID))
			sb.WriteString(fmt.Sprintf("- **Source:** %s\n", event.Source))
			sb.WriteString(fmt.Sprintf("- **Type:** %s\n", event.EventType))
			sb.WriteString(fmt.Sprintf("- **Description:** %s\n", event.Description))
			sb.WriteString(fmt.Sprintf("- **Timestamp:** %s\n", event.Timestamp.Format("2006-01-02 15:04:05")))
			sb.WriteString(fmt.Sprintf("- **Content:**\n```\n%s\n```\n\n", event.Content))
		}
	}

	// Analysis request
	sb.WriteString("## Analysis Required\n\n")
	sb.WriteString("Please provide your analysis in JSON format with:\n\n")
	sb.WriteString("1. **evidence_links**: Array of event IDs (UUID strings) that support this control. Use the EXACT UUID from the event heading (e.g., if the heading shows \"Event 1 (ID: abc-123-xyz)\", use \"abc-123-xyz\").\n")
	sb.WriteString("2. **justification**: 50-500 character explanation of how the evidence supports the control\n")
	sb.WriteString("3. **confidence**: Integer score 0-100 representing confidence in the evidence\n")
	sb.WriteString("4. **residual_risk**: (optional) 0-500 character notes on gaps or concerns\n\n")
	sb.WriteString("IMPORTANT: For evidence_links, copy the full UUID exactly as it appears in parentheses after \"Event N (ID: ...)\". Do not invent event IDs or use descriptions.\n")

	return sb.String()
}

// GeneratePrompts is a convenience method that generates both system and user prompts
func (pg *PromptGenerator) GeneratePrompts(req *AnalysisRequest) (system, user string) {
	return pg.GenerateSystemPrompt(), pg.GenerateUserPrompt(req)
}
