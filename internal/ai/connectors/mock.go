package connectors

import (
	"context"
	"fmt"
	"time"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// MockConnector is a simple connector implementation for testing.
// It returns predefined events based on the query.
type MockConnector struct {
	name   string
	events map[string][]types.EvidenceEvent // query -> events
	errors map[string]error                 // query -> error
}

// NewMockConnector creates a new mock connector with the given name.
func NewMockConnector(cfg Config) (Connector, error) {
	name, ok := cfg.Extra["name"].(string)
	if !ok || name == "" {
		name = "mock"
	}

	return &MockConnector{
		name:   name,
		events: make(map[string][]types.EvidenceEvent),
		errors: make(map[string]error),
	}, nil
}

// Name returns the connector identifier.
func (m *MockConnector) Name() string {
	return m.name
}

// Collect retrieves pre-configured events for the given query.
func (m *MockConnector) Collect(ctx context.Context, query string) ([]types.EvidenceEvent, error) {
	// Check for configured error
	if err, ok := m.errors[query]; ok {
		return nil, err
	}

	// Return configured events
	if events, ok := m.events[query]; ok {
		return events, nil
	}

	// Default: return a single event with the query embedded
	return []types.EvidenceEvent{
		{
			ID:        fmt.Sprintf("%s-mock-1", m.name),
			Source:    m.name,
			Type:      "mock",
			Timestamp: time.Now(),
			Content:   fmt.Sprintf("Mock event for query: %s", query),
			Metadata: map[string]interface{}{
				"query": query,
			},
		},
	}, nil
}

// Validate always returns nil for mock connectors.
func (m *MockConnector) Validate(ctx context.Context) error {
	return nil
}

// SetEvents configures the events to return for a specific query.
func (m *MockConnector) SetEvents(query string, events []types.EvidenceEvent) {
	m.events[query] = events
}

// SetError configures an error to return for a specific query.
func (m *MockConnector) SetError(query string, err error) {
	m.errors[query] = err
}
