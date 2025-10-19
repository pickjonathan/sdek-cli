package connectors

import (
	"context"
	"fmt"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Connector is the interface that all MCP (Model Context Protocol) connectors must implement.
// Each connector is responsible for collecting evidence from a specific source (GitHub, Jira, AWS, etc.).
type Connector interface {
	// Name returns the unique identifier for this connector (e.g., "github", "jira", "aws")
	Name() string

	// Collect retrieves evidence events from the source using the provided query.
	// The query format is connector-specific (e.g., JQL for Jira, GitHub search syntax, etc.)
	//
	// Returns:
	//   - []types.EvidenceEvent: The collected evidence events normalized to the standard schema
	//   - error: Any error encountered during collection (network, auth, rate limit, etc.)
	Collect(ctx context.Context, query string) ([]types.EvidenceEvent, error)

	// Validate checks if the connector is properly configured and can connect to its source.
	// This is called during initialization to fail fast on configuration errors.
	//
	// Returns:
	//   - error: Configuration or connectivity error, nil if valid
	Validate(ctx context.Context) error
}

// Config holds the configuration for a connector instance.
type Config struct {
	// Enabled indicates if this connector should be loaded
	Enabled bool `yaml:"enabled" json:"enabled"`

	// APIKey is the authentication credential (API token, access key, etc.)
	APIKey string `yaml:"api_key" json:"api_key"`

	// Endpoint is the base URL for the service (optional, uses defaults if not specified)
	Endpoint string `yaml:"endpoint" json:"endpoint"`

	// RateLimit is the maximum number of requests per minute (0 = unlimited)
	RateLimit int `yaml:"rate_limit" json:"rate_limit"`

	// Timeout is the maximum duration for a single API call in seconds
	Timeout int `yaml:"timeout" json:"timeout"`

	// Additional connector-specific configuration
	Extra map[string]interface{} `yaml:"extra" json:"extra,omitempty"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Enabled:   true,
		RateLimit: 60, // 60 requests per minute
		Timeout:   30, // 30 seconds
		Extra:     make(map[string]interface{}),
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Enabled && c.APIKey == "" {
		return fmt.Errorf("api_key is required when connector is enabled")
	}
	if c.RateLimit < 0 {
		return fmt.Errorf("rate_limit must be >= 0")
	}
	if c.Timeout < 0 {
		return fmt.Errorf("timeout must be >= 0")
	}
	return nil
}

// Factory is a function that creates a new connector instance from configuration.
type Factory func(cfg Config) (Connector, error)

// Error types for common connector failures
var (
	ErrNotConfigured    = fmt.Errorf("connector not configured")
	ErrAuthFailed       = fmt.Errorf("authentication failed")
	ErrRateLimited      = fmt.Errorf("rate limit exceeded")
	ErrTimeout          = fmt.Errorf("request timeout")
	ErrInvalidQuery     = fmt.Errorf("invalid query syntax")
	ErrSourceNotFound   = fmt.Errorf("source not found")
	ErrPermissionDenied = fmt.Errorf("permission denied")
)
