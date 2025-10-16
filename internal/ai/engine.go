package ai

import (
	"context"
)

// Engine is the core abstraction for AI provider integrations.
// Implementations must support OpenAI and Anthropic initially.
// All implementations MUST be safe for concurrent use.
type Engine interface {
	// Analyze sends an analysis request to the AI provider and returns
	// the structured response. Returns error if provider fails, times out,
	// or returns invalid JSON.
	//
	// Context cancellation triggers immediate abort (no retry).
	// Timeout specified in ctx or falls back to AIConfig.Timeout.
	Analyze(ctx context.Context, req *AnalysisRequest) (*AnalysisResponse, error)

	// Provider returns the provider identifier ("openai" | "anthropic" | "mock").
	Provider() string

	// Health checks if the provider is reachable and configured correctly.
	// Returns error if API key invalid, quota exceeded, or network unreachable.
	Health(ctx context.Context) error
}
