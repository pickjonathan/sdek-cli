package ingest

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// CICDGenerator generates CI/CD pipeline events
type CICDGenerator struct {
	*BaseGenerator
}

// NewCICDGenerator creates a new CI/CD event generator
func NewCICDGenerator(seed int64) *CICDGenerator {
	return &CICDGenerator{
		BaseGenerator: NewBaseGenerator(seed),
	}
}

// GetSourceType returns the source type
func (c *CICDGenerator) GetSourceType() string {
	return string(types.SourceTypeCICD)
}

// Generate creates CI/CD pipeline events
func (c *CICDGenerator) Generate(seed int64, count int) ([]types.Event, error) {
	// Validate event count
	if err := ValidateEventCount(count); err != nil {
		return nil, err
	}

	// Reinitialize with the provided seed for deterministic generation
	c.BaseGenerator = NewBaseGenerator(seed)

	events := make([]types.Event, count)

	// CI/CD-specific data
	pipelineNames := []string{
		"security-scan",
		"build-and-test",
		"deploy-production",
		"compliance-check",
		"integration-tests",
		"vulnerability-scan",
	}

	statuses := []string{"success", "failure", "in_progress", "cancelled"}
	stages := []string{"build", "test", "security", "deploy"}

	for i := 0; i < count; i++ {
		// Generate pipeline attributes
		pipelineName := c.RandomElement(pipelineNames)
		pipelineID := fmt.Sprintf("pipeline-%d", c.RandomInt(10000, 99999))
		status := c.RandomElement(statuses)
		stage := c.RandomElement(stages)
		author := c.RandomElement(AuthorNames)

		// Generate duration (30 seconds to 30 minutes)
		durationSecs := c.RandomInt(30, 1800)

		// Generate title and content
		title := fmt.Sprintf("%s #%d - %s", pipelineName, c.RandomInt(100, 999), status)
		content := c.generateBuildContent(pipelineName, status)

		if c.RandomBool(0.3) {
			keyword := c.RandomElement(SecurityKeywords)
			content = fmt.Sprintf("%s. Includes %s checks.", content, keyword)
		}

		// Create metadata
		metadata := map[string]interface{}{
			"pipeline_id":   pipelineID,
			"pipeline_name": pipelineName,
			"status":        status,
			"stage":         stage,
			"duration_secs": durationSecs,
			"branch":        c.RandomElement(BranchNames),
			"tests_passed":  c.RandomInt(0, 100),
			"tests_failed":  c.RandomInt(0, 10),
		}

		events[i] = types.Event{
			ID:        uuid.New().String(),
			SourceID:  string(types.SourceTypeCICD),
			Timestamp: c.RandomTimestamp(),
			EventType: types.EventTypeBuild,
			Title:     title,
			Content:   content,
			Author:    author,
			Metadata:  metadata,
		}
	}

	return events, nil
}

// generateBuildContent creates realistic build content
func (c *CICDGenerator) generateBuildContent(pipelineName, status string) string {
	successMessages := []string{
		"All security checks passed",
		"Build completed successfully with no vulnerabilities",
		"Deployment successful to production environment",
		"Integration tests passed with 100% coverage",
		"Compliance checks satisfied all requirements",
		"Security scan found no critical issues",
	}

	failureMessages := []string{
		"Security scan detected vulnerabilities",
		"Tests failed - authentication module errors",
		"Deployment blocked by compliance checks",
		"Build failed - dependency security issues",
		"Code quality gate failed",
		"Security policy violations detected",
	}

	if status == "success" {
		return c.RandomElement(successMessages)
	}
	return c.RandomElement(failureMessages)
}
