package unit

import (
	"testing"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/stretchr/testify/assert"
)

// T007: Contract test for AutoApproveMatcher
// These tests define the contract for matching evidence plan items against auto-approve policies
// EXPECTED: These tests MUST FAIL until the AutoApproveMatcher is implemented in Phase 3.3

func TestAutoApprove_ExactMatch(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"authentication", "mfa"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.True(t, matcher.Matches("github", "authentication"), "Exact match should return true")
	assert.True(t, matcher.Matches("github", "mfa"), "Exact match should return true")
	assert.False(t, matcher.Matches("github", "payment"), "No match should return false")
}

func TestAutoApprove_WildcardPrefix(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"auth*"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.True(t, matcher.Matches("github", "authentication"), "Should match prefix wildcard")
	assert.True(t, matcher.Matches("github", "authorize"), "Should match prefix wildcard")
	assert.True(t, matcher.Matches("github", "auth"), "Should match exact prefix")
	assert.False(t, matcher.Matches("github", "unauth"), "Should not match if prefix doesn't match")
	assert.False(t, matcher.Matches("github", "payment"), "Should not match different pattern")
}

func TestAutoApprove_WildcardSuffix(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"jira": {"*-SECURITY"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.True(t, matcher.Matches("jira", "INFOSEC-SECURITY"), "Should match suffix wildcard")
	assert.True(t, matcher.Matches("jira", "APP-SECURITY"), "Should match suffix wildcard")
	assert.False(t, matcher.Matches("jira", "SECURITY-123"), "Should not match if suffix doesn't match")
}

func TestAutoApprove_WildcardMiddle(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"*login*"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.True(t, matcher.Matches("github", "user-login-flow"), "Should match middle wildcard")
	assert.True(t, matcher.Matches("github", "login"), "Should match middle wildcard with exact")
	assert.True(t, matcher.Matches("github", "oauth-login-handler"), "Should match middle wildcard")
	assert.False(t, matcher.Matches("github", "authentication"), "Should not match without 'login'")
}

func TestAutoApprove_CaseInsensitive(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"AUTH*"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.True(t, matcher.Matches("github", "authentication"), "Should be case-insensitive")
	assert.True(t, matcher.Matches("github", "AUTHENTICATION"), "Should be case-insensitive")
	assert.True(t, matcher.Matches("github", "Authentication"), "Should be case-insensitive")
	assert.True(t, matcher.Matches("GITHUB", "authentication"), "Source should be case-insensitive")
}

func TestAutoApprove_SourceNotWhitelisted(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"auth*"},
					"aws":    {"iam*"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.False(t, matcher.Matches("slack", "security-channel"), "Unlisted source should return false")
	assert.False(t, matcher.Matches("docs", "security-policy"), "Unlisted source should return false")
	assert.False(t, matcher.Matches("unknown", "anything"), "Unlisted source should return false")
}

func TestAutoApprove_MultiplePatterns(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"auth*", "*login*", "mfa*", "security/*"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.True(t, matcher.Matches("github", "authentication"), "Should match first pattern")
	assert.True(t, matcher.Matches("github", "user-login"), "Should match second pattern")
	assert.True(t, matcher.Matches("github", "mfa-setup"), "Should match third pattern")
	assert.True(t, matcher.Matches("github", "security/auth"), "Should match fourth pattern")
	assert.False(t, matcher.Matches("github", "payment"), "Should not match any pattern")
}

func TestAutoApprove_MultipleSources(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"auth*"},
					"aws":    {"iam:*", "kms:*"},
					"jira":   {"SEC-*", "INFOSEC-*"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert - GitHub
	assert.True(t, matcher.Matches("github", "authentication"))
	assert.False(t, matcher.Matches("github", "payment"))

	// Act & Assert - AWS
	assert.True(t, matcher.Matches("aws", "iam:CreateUser"))
	assert.True(t, matcher.Matches("aws", "kms:Encrypt"))
	assert.False(t, matcher.Matches("aws", "s3:PutObject"))

	// Act & Assert - Jira
	assert.True(t, matcher.Matches("jira", "SEC-123"))
	assert.True(t, matcher.Matches("jira", "INFOSEC-456"))
	assert.False(t, matcher.Matches("jira", "PROJ-789"))
}

func TestAutoApprove_PolicyDisabled(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: false, // Disabled
				AutoApprove: types.AutoApproveConfig{
					"github": {"auth*"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.False(t, matcher.Matches("github", "authentication"), "Should return false when disabled")
	assert.False(t, matcher.Matches("github", "auth"), "Should return false when disabled")
}

func TestAutoApprove_EmptyPatterns(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {}, // No patterns
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.False(t, matcher.Matches("github", "anything"), "Should return false with no patterns")
}

func TestAutoApprove_EmptyQuery(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"auth*"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.False(t, matcher.Matches("github", ""), "Should return false for empty query")
}

func TestAutoApprove_EmptySource(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"auth*"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.False(t, matcher.Matches("", "authentication"), "Should return false for empty source")
}

func TestAutoApprove_QuestionMarkWildcard(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"auth?"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.True(t, matcher.Matches("github", "auth1"), "? should match single character")
	assert.True(t, matcher.Matches("github", "authX"), "? should match single character")
	assert.False(t, matcher.Matches("github", "auth"), "? requires at least one character")
	assert.False(t, matcher.Matches("github", "auth12"), "? matches only one character")
}

func TestAutoApprove_DoubleAsteriskWildcard(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"security/**"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.True(t, matcher.Matches("github", "security/auth/login"), "** should match nested paths")
	assert.True(t, matcher.Matches("github", "security/mfa"), "** should match any depth")
	assert.True(t, matcher.Matches("github", "security/"), "** should match with trailing slash")
	assert.False(t, matcher.Matches("github", "auth/security"), "Should not match if prefix doesn't match")
}

func TestAutoApprove_PerformanceBenchmark(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"auth*", "*login*", "mfa*", "security/*"},
					"aws":    {"iam:*", "kms:*"},
					"jira":   {"SEC-*", "INFOSEC-*"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act - perform 1000 matches
	start := time.Now()
	for i := 0; i < 1000; i++ {
		matcher.Matches("github", "authentication")
		matcher.Matches("aws", "iam:CreateUser")
		matcher.Matches("jira", "SEC-123")
	}
	duration := time.Since(start)
	avgPerMatch := duration / 3000

	// Assert
	assert.Less(t, avgPerMatch, 1*time.Microsecond, "Each match should take <1μs")
	t.Logf("Average match time: %v", avgPerMatch)
}

func TestAutoApprove_SpecialCharacters(t *testing.T) {
	// Arrange
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"auth-*", "login_*"},
				},
			},
		},
	}
	matcher := ai.NewAutoApproveMatcher(cfg)

	// Act & Assert
	assert.True(t, matcher.Matches("github", "auth-handler"), "Should match with hyphen")
	assert.True(t, matcher.Matches("github", "login_flow"), "Should match with underscore")
	assert.False(t, matcher.Matches("github", "authhandler"), "Should not match without separator")
}
