package ingest

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// GitGenerator generates Git commit events
type GitGenerator struct {
	*BaseGenerator
}

// NewGitGenerator creates a new Git event generator
func NewGitGenerator(seed int64) *GitGenerator {
	return &GitGenerator{
		BaseGenerator: NewBaseGenerator(seed),
	}
}

// GetSourceType returns the source type
func (g *GitGenerator) GetSourceType() string {
	return string(types.SourceTypeGit)
}

// Generate creates Git commit events
func (g *GitGenerator) Generate(seed int64, count int) ([]types.Event, error) {
	// Validate event count
	if err := ValidateEventCount(count); err != nil {
		return nil, err
	}

	// Reinitialize with the provided seed for deterministic generation
	g.BaseGenerator = NewBaseGenerator(seed)

	events := make([]types.Event, count)

	for i := 0; i < count; i++ {
		// Generate commit SHA
		commitSHA := g.generateCommitSHA()

		// Select branch
		branch := g.RandomElement(BranchNames)

		// Select author
		author := g.RandomElement(AuthorNames)

		// Select files changed (1-5 files)
		filesChanged := g.RandomSubset(FilePaths, 1, 5)

		// Generate commit message
		prefix := g.RandomElement(CommitPrefixes)
		message := fmt.Sprintf("%s %s", prefix, g.generateCommitMessage())

		// Create metadata
		metadata := map[string]interface{}{
			"commit_sha":    commitSHA,
			"branch":        branch,
			"files_changed": filesChanged,
			"additions":     g.RandomInt(1, 100),
			"deletions":     g.RandomInt(1, 50),
		}

		// Determine if this is security-related
		content := message
		if g.RandomBool(0.3) {
			keyword := g.RandomElement(SecurityKeywords)
			content = fmt.Sprintf("%s - includes %s changes", message, keyword)
		}

		events[i] = types.Event{
			ID:        uuid.New().String(),
			SourceID:  string(types.SourceTypeGit),
			Timestamp: g.RandomTimestamp(),
			EventType: types.EventTypeCommit,
			Title:     fmt.Sprintf("Commit %s on %s", commitSHA[:7], branch),
			Content:   content,
			Author:    author,
			Metadata:  metadata,
		}
	}

	return events, nil
}

// generateCommitSHA creates a realistic-looking Git SHA
func (g *GitGenerator) generateCommitSHA() string {
	const hexChars = "0123456789abcdef"
	sha := make([]byte, 40)
	for i := range sha {
		sha[i] = hexChars[g.rand.Intn(len(hexChars))]
	}
	return string(sha)
}

// generateCommitMessage creates a realistic commit message
func (g *GitGenerator) generateCommitMessage() string {
	messages := []string{
		"update authentication logic",
		"add new API endpoint",
		"fix security vulnerability",
		"refactor database queries",
		"improve error handling",
		"update documentation",
		"add unit tests",
		"optimize performance",
		"fix bug in user validation",
		"implement access control",
		"update dependencies",
		"add logging",
		"fix memory leak",
		"improve code coverage",
		"update configuration",
	}
	return g.RandomElement(messages)
}
