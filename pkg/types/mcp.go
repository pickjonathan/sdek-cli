package types

// MCPConfig is the top-level MCP configuration container
// that manages MCP server connections and behavior.
type MCPConfig struct {
	// Enabled determines if MCP integration is active
	Enabled bool `yaml:"enabled" json:"enabled" mapstructure:"enabled"`

	// PreferMCP determines if MCP tools take precedence over legacy connectors
	PreferMCP bool `yaml:"prefer_mcp" json:"prefer_mcp" mapstructure:"prefer_mcp"`

	// MaxConcurrent sets the maximum concurrent MCP server connections (1-100)
	MaxConcurrent int `yaml:"max_concurrent" json:"max_concurrent" mapstructure:"max_concurrent"`

	// HealthCheckInterval sets seconds between health checks (minimum 60)
	HealthCheckInterval int `yaml:"health_check_interval" json:"health_check_interval" mapstructure:"health_check_interval"`

	// Retry contains retry behavior configuration
	Retry RetryConfig `yaml:"retry" json:"retry" mapstructure:"retry"`

	// Servers contains MCP server definitions by name
	Servers map[string]MCPServerConfig `yaml:"servers" json:"servers" mapstructure:"servers"`
}

// MCPServerConfig defines configuration for a single MCP server instance.
type MCPServerConfig struct {
	// Command is the executable path or command name (required for stdio)
	Command string `yaml:"command,omitempty" json:"command,omitempty" mapstructure:"command"`

	// Args contains command-line arguments (optional)
	Args []string `yaml:"args,omitempty" json:"args,omitempty" mapstructure:"args"`

	// URL is the HTTP endpoint URL (required for http transport)
	URL string `yaml:"url,omitempty" json:"url,omitempty" mapstructure:"url"`

	// Transport specifies the communication mechanism: "stdio" or "http"
	Transport string `yaml:"transport" json:"transport" mapstructure:"transport"`

	// Timeout is the request timeout in seconds (1-600, default 60)
	Timeout int `yaml:"timeout" json:"timeout" mapstructure:"timeout"`

	// RateLimit sets requests per minute (0 = unlimited, default 0)
	RateLimit int `yaml:"rate_limit" json:"rate_limit" mapstructure:"rate_limit"`

	// Env contains environment variables for the process
	Env map[string]string `yaml:"env,omitempty" json:"env,omitempty" mapstructure:"env"`

	// Headers contains HTTP headers (for http transport only)
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty" mapstructure:"headers"`

	// HealthURL is the health check endpoint (optional, for http transport)
	HealthURL string `yaml:"health_url,omitempty" json:"health_url,omitempty" mapstructure:"health_url"`
}

// RetryConfig defines retry behavior for MCP server failures.
type RetryConfig struct {
	// MaxAttempts is the maximum retry attempts (default 3)
	MaxAttempts int `yaml:"max_attempts" json:"max_attempts" mapstructure:"max_attempts"`

	// Backoff is the backoff strategy: "exponential", "linear", or "constant"
	Backoff string `yaml:"backoff" json:"backoff" mapstructure:"backoff"`

	// InitialDelayMS is the initial retry delay in milliseconds (default 1000)
	InitialDelayMS int `yaml:"initial_delay_ms" json:"initial_delay_ms" mapstructure:"initial_delay_ms"`

	// MaxDelayMS is the maximum retry delay in milliseconds (default 30000)
	MaxDelayMS int `yaml:"max_delay_ms" json:"max_delay_ms" mapstructure:"max_delay_ms"`
}

// DefaultMCPConfig returns a default MCP configuration.
func DefaultMCPConfig() MCPConfig {
	return MCPConfig{
		Enabled:             true,
		PreferMCP:           true,
		MaxConcurrent:       10,
		HealthCheckInterval: 300,
		Retry: RetryConfig{
			MaxAttempts:    3,
			Backoff:        "exponential",
			InitialDelayMS: 1000,
			MaxDelayMS:     30000,
		},
		Servers: make(map[string]MCPServerConfig),
	}
}

// DefaultMCPServerConfig returns a default MCP server configuration.
func DefaultMCPServerConfig() MCPServerConfig {
	return MCPServerConfig{
		Transport: "stdio",
		Timeout:   60,
		RateLimit: 0,
		Env:       make(map[string]string),
		Headers:   make(map[string]string),
	}
}
