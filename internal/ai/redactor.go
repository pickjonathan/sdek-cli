package ai

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Redactor interface defines the contract for redacting PII and secrets from text.
type Redactor interface {
	Redact(text string) (redacted string, redactionMap *types.RedactionMap, err error)
}

// redactor implements the Redactor interface.
type redactor struct {
	config   *types.Config
	patterns map[string]*regexp.Regexp
}

// Redaction pattern types
const (
	patternEmail      = "email"
	patternIPv4       = "ipv4"
	patternIPv6       = "ipv6"
	patternPhone      = "phone"
	patternAWSKey     = "awskey"
	patternGenericKey = "generickey"
)

// NewRedactor creates a new Redactor instance.
func NewRedactor(cfg *types.Config) Redactor {
	r := &redactor{
		config:   cfg,
		patterns: make(map[string]*regexp.Regexp),
	}
	r.compilePatterns()
	return r
}

// compilePatterns compiles all regex patterns used for redaction.
func (r *redactor) compilePatterns() {
	// Email pattern
	r.patterns[patternEmail] = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)

	// IPv4 pattern
	r.patterns[patternIPv4] = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)

	// IPv6 pattern (simplified - matches common formats)
	r.patterns[patternIPv6] = regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b`)

	// Phone pattern (matches various formats: 555-123-4567, (555) 123-4567, 5551234567)
	r.patterns[patternPhone] = regexp.MustCompile(`\b(?:\(?\d{3}\)?[-.\s]?)?\d{3}[-.\s]?\d{4}\b`)

	// AWS Access Key pattern
	r.patterns[patternAWSKey] = regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)

	// Generic API key pattern (32+ character alphanumeric strings)
	r.patterns[patternGenericKey] = regexp.MustCompile(`\b[a-zA-Z0-9]{32,}\b`)
}

// Redact redacts PII and secrets from the input text.
func (r *redactor) Redact(text string) (string, *types.RedactionMap, error) {
	// If redaction is disabled, return original text
	if !r.config.AI.Redaction.Enabled {
		return text, types.NewRedactionMap(), nil
	}

	redactionMap := types.NewRedactionMap()
	result := text
	originalLength := len(text)
	position := 0

	// Apply redactions in order: denylist → emails → IPs → phones → keys

	// 1. Denylist (case-insensitive)
	for _, denyItem := range r.config.AI.Redaction.Denylist {
		pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(denyItem) + `\b`)
		matches := pattern.FindAllStringIndex(result, -1)
		for _, match := range matches {
			originalText := result[match[0]:match[1]]
			hash := hashString(originalText)
			placeholder := "[REDACTED:SECRET]"

			redactionMap.AddEntry(hash, types.RedactionEntry{
				OriginalHash: hash,
				Placeholder:  placeholder,
				Type:         types.RedactionSecret,
				Position:     position + match[0],
				Timestamp:    time.Now(),
			})

			result = result[:match[0]] + placeholder + result[match[1]:]
			position += len(placeholder) - (match[1] - match[0])
		}
	}

	// 2. Emails
	result = r.redactPattern(result, patternEmail, "[REDACTED:PII:EMAIL]", types.RedactionPII, redactionMap)

	// 3. IPv6 (must come before IPv4 to avoid partial matches)
	result = r.redactPattern(result, patternIPv6, "[REDACTED:PII:IP]", types.RedactionPII, redactionMap)

	// 4. IPv4
	result = r.redactPattern(result, patternIPv4, "[REDACTED:PII:IP]", types.RedactionPII, redactionMap)

	// 5. Phone numbers
	result = r.redactPattern(result, patternPhone, "[REDACTED:PII:PHONE]", types.RedactionPII, redactionMap)

	// 6. AWS Keys
	result = r.redactPattern(result, patternAWSKey, "[REDACTED:SECRET]", types.RedactionSecret, redactionMap)

	// 7. Generic API keys
	result = r.redactPattern(result, patternGenericKey, "[REDACTED:SECRET]", types.RedactionSecret, redactionMap)

	// Check if redaction percentage exceeds threshold (>40%)
	redactedLength := originalLength - len(result) + (redactionMap.TotalRedactions * len("[REDACTED:XXX:XXX]"))
	if float64(redactedLength)/float64(originalLength) > 0.4 {
		// Warning: High redaction percentage
		// Implementation note: Could add a warning field to RedactionMap if needed
	}

	return result, redactionMap, nil
}

// redactPattern applies a regex pattern and redacts all matches.
func (r *redactor) redactPattern(text, patternName, placeholder string, redactionType types.RedactionType, rm *types.RedactionMap) string {
	pattern := r.patterns[patternName]
	if pattern == nil {
		return text
	}

	matches := pattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return text
	}

	// Process matches in reverse order to avoid index shifting
	result := text
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		originalText := text[match[0]:match[1]]

		// Skip if already redacted
		if strings.Contains(originalText, "[REDACTED:") {
			continue
		}

		hash := hashString(originalText)

		rm.AddEntry(hash, types.RedactionEntry{
			OriginalHash: hash,
			Placeholder:  placeholder,
			Type:         redactionType,
			Position:     match[0],
			Timestamp:    time.Now(),
		})

		result = result[:match[0]] + placeholder + result[match[1]:]
	}

	return result
}

// hashString creates a SHA256 hash of the input string.
func hashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
