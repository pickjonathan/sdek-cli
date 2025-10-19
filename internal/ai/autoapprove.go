package ai

import (
	"strings"

	"github.com/gobwas/glob"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// AutoApproveMatcher interface defines the contract for matching evidence plan items
// against auto-approve policies.
type AutoApproveMatcher interface {
	Matches(source, query string) bool
}

// autoApproveMatcher implements the AutoApproveMatcher interface.
type autoApproveMatcher struct {
	policy  map[string][]string    // Rules from config
	enabled bool                   // Autonomous.Enabled flag
	globs   map[string][]glob.Glob // Pre-compiled glob patterns per source
}

// NewAutoApproveMatcher creates a new AutoApproveMatcher instance.
func NewAutoApproveMatcher(cfg *types.Config) AutoApproveMatcher {
	// Use cfg.AI.Autonomous.AutoApprove as the direct map (backward compatibility)
	// Or cfg.AI.Autonomous.AutoApprove.Rules if it's the nested structure
	policy := cfg.AI.Autonomous.AutoApprove

	matcher := &autoApproveMatcher{
		policy:  policy,
		enabled: cfg.AI.Autonomous.Enabled,
		globs:   make(map[string][]glob.Glob),
	}
	matcher.compilePatterns()
	return matcher
}

// compilePatterns pre-compiles all glob patterns for performance.
func (m *autoApproveMatcher) compilePatterns() {
	for source, patterns := range m.policy {
		sourceLower := strings.ToLower(source)
		m.globs[sourceLower] = make([]glob.Glob, 0, len(patterns))

		for _, pattern := range patterns {
			// Compile pattern as case-insensitive
			patternLower := strings.ToLower(pattern)
			g, err := glob.Compile(patternLower)
			if err != nil {
				// Skip invalid patterns (should be validated at config load time)
				continue
			}
			m.globs[sourceLower] = append(m.globs[sourceLower], g)
		}
	}
}

// Matches checks if a source/query combination matches any auto-approve pattern.
func (m *autoApproveMatcher) Matches(source, query string) bool {
	// Policy must be enabled
	if !m.enabled {
		return false
	}

	// Empty inputs don't match
	if source == "" || query == "" {
		return false
	}

	// Case-insensitive lookup
	sourceLower := strings.ToLower(source)
	queryLower := strings.ToLower(query)

	// Check if source exists in policy
	patterns, exists := m.globs[sourceLower]
	if !exists {
		return false // Source not whitelisted
	}

	// Empty patterns for source
	if len(patterns) == 0 {
		return false
	}

	// Try each pattern
	for _, pattern := range patterns {
		if pattern.Match(queryLower) {
			return true
		}
	}

	return false
}
