package providers

import (
	"context"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai"
)

// MockEngine is a mock implementation of ai.Engine for testing
type MockEngine struct {
	provider  string
	response  *ai.AnalysisResponse
	err       error
	healthErr error
}

// NewMockEngine creates a new mock engine
func NewMockEngine(provider string) *MockEngine {
	return &MockEngine{
		provider: provider,
	}
}

// Analyze implements ai.Engine.Analyze
func (m *MockEngine) Analyze(ctx context.Context, req *ai.AnalysisRequest) (*ai.AnalysisResponse, error) {
	// Check for context cancellation/timeout
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Return configured error if set
	if m.err != nil {
		return nil, m.err
	}

	// Return configured response if set
	if m.response != nil {
		return m.response, nil
	}

	// Return default response
	return &ai.AnalysisResponse{
		RequestID:     req.RequestID,
		EvidenceLinks: []string{req.Events[0].EventID},
		Justification: "Mock analysis justification",
		Confidence:    85,
		ResidualRisk:  "No residual risk identified",
		Provider:      m.provider,
		Model:         "mock-model",
		TokensUsed:    100,
		Latency:       10,
		Timestamp:     time.Now(),
		CacheHit:      false,
	}, nil
}

// Provider implements ai.Engine.Provider
func (m *MockEngine) Provider() string {
	return m.provider
}

// Health implements ai.Engine.Health
func (m *MockEngine) Health(ctx context.Context) error {
	// Check for context cancellation/timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Return configured health error if set
	if m.healthErr != nil {
		return m.healthErr
	}

	return nil
}

// SetResponse configures the response to return from Analyze
func (m *MockEngine) SetResponse(response *ai.AnalysisResponse) {
	m.response = response
}

// SetError configures the error to return from Analyze
func (m *MockEngine) SetError(err error) {
	m.err = err
}

// SetHealthError configures the error to return from Health
func (m *MockEngine) SetHealthError(err error) {
	m.healthErr = err
}
