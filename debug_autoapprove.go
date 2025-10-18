package main

import (
	"fmt"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func main() {
	cfg := &types.Config{
		AI: types.AIConfig{
			Autonomous: types.AutonomousConfig{
				Enabled: true,
				AutoApprove: types.AutoApproveConfig{
					"github": {"auth*", "*login*"},
					"aws":    {"iam:*"},
				},
			},
		},
	}

	matcher := ai.NewAutoApproveMatcher(cfg)

	// Test cases
	tests := []struct {
		source string
		query  string
	}{
		{"github", "authentication"},
		{"github", "payment"},
		{"aws", "iam:CreateUser"},
	}

	for _, test := range tests {
		result := matcher.Matches(test.source, test.query)
		fmt.Printf("%s/%s: %v\n", test.source, test.query, result)
	}
}
