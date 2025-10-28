package mcp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// NormalizeToEvidenceEvent converts MCP tool results to EvidenceEvent format
func NormalizeToEvidenceEvent(serverName, toolName string, toolResult interface{}) ([]types.EvidenceEvent, error) {
	// Parse tool result based on structure
	var events []types.EvidenceEvent

	// Try to unmarshal as JSON
	resultBytes, err := json.Marshal(toolResult)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tool result: %w", err)
	}

	// Try to parse as structured data
	var resultMap map[string]interface{}
	if err := json.Unmarshal(resultBytes, &resultMap); err != nil {
		// If not structured data, create a single event with raw content
		event := types.EvidenceEvent{
			Source:    serverName,
			Type:      toolName,
			Timestamp: time.Now(),
			Content:   fmt.Sprintf("%v", toolResult),
		}
		return []types.EvidenceEvent{event}, nil
	}

	// Check if result contains an array of items
	if items, ok := resultMap["items"].([]interface{}); ok {
		// Result is a collection of items
		for _, item := range items {
			event, err := normalizeItem(serverName, toolName, item)
			if err != nil {
				// Log error but continue processing other items
				fmt.Printf("Warning: failed to normalize item: %v\n", err)
				continue
			}
			events = append(events, event)
		}
	} else if content, ok := resultMap["content"]; ok {
		// Result has a single content field
		event := types.EvidenceEvent{
			Source:    serverName,
			Type:      toolName,
			Timestamp: extractTimestamp(resultMap),
			Content:   fmt.Sprintf("%v", content),
			Metadata:  extractMetadata(resultMap),
		}
		events = append(events, event)
	} else {
		// Result is a single structured item
		event, err := normalizeItem(serverName, toolName, resultMap)
		if err != nil {
			return nil, fmt.Errorf("failed to normalize result: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// normalizeItem converts a single item to an EvidenceEvent
func normalizeItem(serverName, toolName string, item interface{}) (types.EvidenceEvent, error) {
	itemMap, ok := item.(map[string]interface{})
	if !ok {
		// Item is not a map, use as-is
		return types.EvidenceEvent{
			Source:    serverName,
			Type:      toolName,
			Timestamp: time.Now(),
			Content:   fmt.Sprintf("%v", item),
		}, nil
	}

	event := types.EvidenceEvent{
		Source:    serverName,
		Type:      toolName,
		Timestamp: extractTimestamp(itemMap),
		Content:   extractContent(itemMap),
		Metadata:  extractMetadata(itemMap),
	}

	return event, nil
}

// extractTimestamp extracts timestamp from various field names
func extractTimestamp(data map[string]interface{}) time.Time {
	// Try common timestamp field names
	timestampFields := []string{"timestamp", "created_at", "updated_at", "date", "time", "createdAt", "updatedAt"}

	for _, field := range timestampFields {
		if val, ok := data[field]; ok {
			switch v := val.(type) {
			case string:
				// Try parsing as RFC3339
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					return t
				}
				// Try parsing as other common formats
				formats := []string{
					time.RFC3339Nano,
					time.RFC1123,
					"2006-01-02T15:04:05Z",
					"2006-01-02 15:04:05",
					"2006-01-02",
				}
				for _, format := range formats {
					if t, err := time.Parse(format, v); err == nil {
						return t
					}
				}
			case float64:
				// Unix timestamp
				return time.Unix(int64(v), 0)
			case int64:
				return time.Unix(v, 0)
			}
		}
	}

	// Default to now if no timestamp found
	return time.Now()
}

// extractContent extracts the main content from the data
func extractContent(data map[string]interface{}) string {
	// Try common content field names
	contentFields := []string{"content", "message", "description", "text", "body", "summary"}

	for _, field := range contentFields {
		if val, ok := data[field]; ok {
			return fmt.Sprintf("%v", val)
		}
	}

	// If no content field found, marshal entire object as JSON
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return string(bytes)
}

// extractMetadata extracts metadata fields (excluding content and timestamp)
func extractMetadata(data map[string]interface{}) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Fields to exclude from metadata
	excludeFields := map[string]bool{
		"content":     true,
		"message":     true,
		"description": true,
		"text":        true,
		"body":        true,
		"summary":     true,
		"timestamp":   true,
		"created_at":  true,
		"updated_at":  true,
		"date":        true,
		"time":        true,
		"createdAt":   true,
		"updatedAt":   true,
	}

	for key, val := range data {
		if excludeFields[key] {
			continue
		}

		// Keep value as-is
		metadata[key] = val
	}

	return metadata
}

// NormalizeMCPError converts MCP errors to Evidence Event error format
func NormalizeMCPError(serverName, toolName string, mcpError error) types.EvidenceEvent {
	return types.EvidenceEvent{
		Source:    serverName,
		Type:      toolName,
		Timestamp: time.Now(),
		Content:   fmt.Sprintf("Error: %v", mcpError),
		Metadata: map[string]interface{}{
			"error": "true",
			"type":  "mcp_error",
		},
	}
}
