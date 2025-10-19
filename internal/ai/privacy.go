package ai

import (
	"regexp"
	"strings"
	"sync"
)

// NewPrivacyFilter creates a new PrivacyFilter with default patterns
func NewPrivacyFilter() *PrivacyFilter {
	return &PrivacyFilter{
		EmailPattern:      regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`),
		PhonePattern:      regexp.MustCompile(`\b(\+?1?[-.\s]?)?(\(?\d{3}\)?[-.\s]?)?\d{3}[-.\s]?\d{4}\b`),
		APIKeyPattern:     regexp.MustCompile(`\b(sk-[A-Za-z0-9]{32,}|ghp_[A-Za-z0-9]{36}|gho_[A-Za-z0-9]{36}|ghs_[A-Za-z0-9]{36}|AKIA[0-9A-Z]{16}|[A-Za-z0-9_-]{32,})\b`),
		CreditCardPattern: regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`),
		SSNPattern:        regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		AllowedFields:     []string{"timestamp", "log_level", "status_code", "status", "level"},
		CustomPatterns:    []*regexp.Regexp{},
	}
}

// Redact performs PII and secret redaction on the input text
func (pf *PrivacyFilter) Redact(text string) RedactionResult {
	result := RedactionResult{
		Original:   text,
		Redacted:   text,
		Redactions: []RedactionInfo{},
	}

	// Check if text contains only allowlisted fields
	lower := strings.ToLower(text)
	isAllowlisted := false
	for _, field := range pf.AllowedFields {
		if strings.Contains(lower, field) && len(text) < 100 {
			isAllowlisted = true
			break
		}
	}

	if isAllowlisted {
		return result
	}

	// Apply redaction patterns in order
	result = pf.redactPattern(result, pf.EmailPattern, "email", "<EMAIL_REDACTED>")
	result = pf.redactPattern(result, pf.PhonePattern, "phone", "<PHONE_REDACTED>")
	result = pf.redactPattern(result, pf.APIKeyPattern, "api_key", "<API_KEY_REDACTED>")
	result = pf.redactPattern(result, pf.CreditCardPattern, "credit_card", "<CARD_REDACTED>")
	result = pf.redactPattern(result, pf.SSNPattern, "ssn", "<SSN_REDACTED>")

	// Apply custom patterns
	for i, pattern := range pf.CustomPatterns {
		patternName := "custom_" + string(rune('0'+i))
		result = pf.redactPattern(result, pattern, patternName, "<REDACTED>")
	}

	return result
}

// redactPattern applies a single redaction pattern to the text
func (pf *PrivacyFilter) redactPattern(result RedactionResult, pattern *regexp.Regexp, patternName, replacement string) RedactionResult {
	matches := pattern.FindAllStringIndex(result.Redacted, -1)

	if len(matches) == 0 {
		return result
	}

	// Track statistics
	pf.IncrementRedactionCount(patternName)

	// Build new redacted string with all matches replaced
	var sb strings.Builder
	lastEnd := 0

	for _, match := range matches {
		start, end := match[0], match[1]

		// Add text before match
		sb.WriteString(result.Redacted[lastEnd:start])

		// Add replacement
		sb.WriteString(replacement)

		// Record redaction info
		result.Redactions = append(result.Redactions, RedactionInfo{
			PatternName: patternName,
			Position:    start,
			Length:      end - start,
			Replacement: replacement,
		})

		lastEnd = end
	}

	// Add remaining text
	sb.WriteString(result.Redacted[lastEnd:])

	result.Redacted = sb.String()
	return result
}

// SetAllowedFields updates the allowlist of safe fields
func (pf *PrivacyFilter) SetAllowedFields(fields []string) {
	pf.AllowedFields = fields
}

// AddCustomPattern adds a custom redaction pattern
func (pf *PrivacyFilter) AddCustomPattern(pattern *regexp.Regexp) {
	pf.CustomPatterns = append(pf.CustomPatterns, pattern)
}

// RedactEvents redacts PII from a slice of events
func (pf *PrivacyFilter) RedactEvents(events []AnalysisEvent) []AnalysisEvent {
	redacted := make([]AnalysisEvent, len(events))

	for i, event := range events {
		redacted[i] = event

		// Redact description
		descResult := pf.Redact(event.Description)
		redacted[i].Description = descResult.Redacted

		// Redact content
		contentResult := pf.Redact(event.Content)
		redacted[i].Content = contentResult.Redacted
	}

	return redacted
}

// ResetStatistics clears the redaction count statistics
func (pf *PrivacyFilter) ResetStatistics() {
	pf.redactionCount = sync.Map{}
}
