package unit

import (
	"testing"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Helper function to create config with auto-approve rules
func createAutoApproveConfig(rules map[string][]string) *types.Config {
	return &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled:     true,
				AutoApprove: rules,
			},
		},
	}
}

// BenchmarkAutoApproveMatches tests pattern matching performance
// Target: <1µs per match
func BenchmarkAutoApproveMatches(b *testing.B) {
	config := createAutoApproveConfig(map[string][]string{
		"github": {"auth*", "*login*", "mfa*"},
		"aws":    {"iam*", "security*"},
		"jira":   {"INFOSEC-*"},
	})

	matcher := ai.NewAutoApproveMatcher(config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = matcher.Matches("github", "authentication")
	}
}

// BenchmarkAutoApproveMatches_Miss tests performance when pattern doesn't match
func BenchmarkAutoApproveMatches_Miss(b *testing.B) {
	config := createAutoApproveConfig(map[string][]string{
		"github": {"auth*", "*login*", "mfa*"},
		"aws":    {"iam*", "security*"},
		"jira":   {"INFOSEC-*"},
	})

	matcher := ai.NewAutoApproveMatcher(config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = matcher.Matches("github", "payment-processing")
	}
}

// BenchmarkAutoApproveMatches_UnknownSource tests performance when source is not in rules
func BenchmarkAutoApproveMatches_UnknownSource(b *testing.B) {
	config := createAutoApproveConfig(map[string][]string{
		"github": {"auth*", "*login*", "mfa*"},
		"aws":    {"iam*", "security*"},
		"jira":   {"INFOSEC-*"},
	})

	matcher := ai.NewAutoApproveMatcher(config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = matcher.Matches("slack", "security-channel")
	}
}

// BenchmarkAutoApproveMatches_MultiplePatterns tests with many patterns
func BenchmarkAutoApproveMatches_MultiplePatterns(b *testing.B) {
	config := createAutoApproveConfig(map[string][]string{
		"github": {
			"auth*", "*login*", "mfa*", "security*", "*password*",
			"oauth*", "*session*", "token*", "*credential*", "access*",
		},
	})

	matcher := ai.NewAutoApproveMatcher(config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = matcher.Matches("github", "authentication")
	}
}

// BenchmarkAutoApproveMatches_MiddleWildcard tests performance with middle wildcards
func BenchmarkAutoApproveMatches_MiddleWildcard(b *testing.B) {
	config := createAutoApproveConfig(map[string][]string{
		"github": {"*auth*", "*login*", "*security*"},
	})

	matcher := ai.NewAutoApproveMatcher(config)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = matcher.Matches("github", "user-authentication-service")
	}
}

// BenchmarkAutoApproveMatches_LongQuery tests performance with long query strings
func BenchmarkAutoApproveMatches_LongQuery(b *testing.B) {
	config := createAutoApproveConfig(map[string][]string{
		"github": {"auth*", "*login*", "mfa*"},
	})

	matcher := ai.NewAutoApproveMatcher(config)
	longQuery := "this-is-a-very-long-query-string-that-contains-authentication-logic-and-other-security-related-features"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = matcher.Matches("github", longQuery)
	}
}

// BenchmarkAutoApproveMatcher_Creation tests the cost of creating a new matcher
func BenchmarkAutoApproveMatcher_Creation(b *testing.B) {
	config := createAutoApproveConfig(map[string][]string{
		"github": {"auth*", "*login*", "mfa*"},
		"aws":    {"iam*", "security*"},
		"jira":   {"INFOSEC-*"},
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ai.NewAutoApproveMatcher(config)
	}
}

// BenchmarkAutoApproveMatches_Sequential tests sequential pattern matching
func BenchmarkAutoApproveMatches_Sequential(b *testing.B) {
	config := createAutoApproveConfig(map[string][]string{
		"github": {"auth*", "*login*", "mfa*"},
		"aws":    {"iam*", "security*"},
		"jira":   {"INFOSEC-*"},
	})

	matcher := ai.NewAutoApproveMatcher(config)

	queries := []struct {
		source string
		query  string
	}{
		{"github", "authentication"},
		{"aws", "iam-policy"},
		{"jira", "INFOSEC-123"},
		{"github", "user-login"},
		{"aws", "security-group"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		q := queries[i%len(queries)]
		_ = matcher.Matches(q.source, q.query)
	}
}
