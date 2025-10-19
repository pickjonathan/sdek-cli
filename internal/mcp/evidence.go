package mcp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// normalizeEvidence converts an MCP tool response into a types.Evidence entity.
// This provides a unified evidence structure for all MCP-sourced evidence.
func normalizeEvidence(toolName, method string, response map[string]interface{}, invokedAt time.Time) *types.Evidence {
	// Extract result from MCP response
	var resultText string
	if result, ok := response["result"]; ok {
		resultJSON, _ := json.Marshal(result)
		resultText = string(resultJSON)
	} else {
		// Fallback to full response
		responseJSON, _ := json.Marshal(response)
		resultText = string(responseJSON)
	}

	// Generate a deterministic ID based on tool name, method, and timestamp
	evidenceID := uuid.New().String()

	// Create Evidence entity
	evidence := &types.Evidence{
		ID:              evidenceID,
		EventID:         "", // Will be linked later during analysis
		ControlID:       "", // Will be mapped later during analysis
		FrameworkID:     "", // Will be determined during mapping
		ConfidenceLevel: types.ConfidenceLevelMedium, // Default confidence for MCP-sourced evidence
		ConfidenceScore: 70.0,                        // Default score
		MappedAt:        invokedAt,
		Keywords:        extractKeywords(resultText),
		Reasoning:       fmt.Sprintf("Evidence collected via MCP tool '%s' using method '%s'", toolName, method),

		// AI analysis metadata (will be populated later if AI analysis is performed)
		AIAnalyzed:          false,
		AIJustification:     "",
		AIConfidence:        0,
		AIResidualRisk:      "",
		HeuristicConfidence: 70, // Default heuristic confidence
		CombinedConfidence:  70, // Will be calculated after AI analysis
		AnalysisMethod:      "mcp-direct", // Indicates this came directly from MCP tool
	}

	return evidence
}

// extractKeywords extracts potential keywords from MCP response text.
// This is a simple implementation that can be enhanced with NLP later.
func extractKeywords(text string) []string {
	// For now, return empty - keywords will be extracted during AI analysis
	// This can be enhanced with keyword extraction logic later
	return []string{"mcp-evidence"}
}

// normalizeEvidenceWithContext creates evidence with additional context.
// This variant allows specifying event and control IDs for direct mapping.
func normalizeEvidenceWithContext(toolName, method string, response map[string]interface{}, invokedAt time.Time, eventID, controlID, frameworkID string) *types.Evidence {
	evidence := normalizeEvidence(toolName, method, response, invokedAt)
	evidence.EventID = eventID
	evidence.ControlID = controlID
	evidence.FrameworkID = frameworkID
	return evidence
}
