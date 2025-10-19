package types

import (
	"fmt"
	"regexp"
	"time"
)

// MCPConfig represents an MCP tool configuration loaded from a JSON file.
// Compatible with VS Code/Cursor MCP configuration structure.
type MCPConfig struct {
	Name          string            `json:"name"`
	Command       string            `json:"command"`
	Args          []string          `json:"args,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Transport     string            `json:"transport"` // "stdio" or "http"
	BaseURL       string            `json:"baseURL,omitempty"`
	Capabilities  []string          `json:"capabilities"`
	Timeout       string            `json:"timeout,omitempty"` // Go duration format
	SchemaVersion string            `json:"schemaVersion"`
}

// Validate validates the MCPConfig according to the schema rules.
func (c *MCPConfig) Validate() error {
	// Name validation
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	namePattern := regexp.MustCompile(`^[a-z0-9-]+$`)
	if !namePattern.MatchString(c.Name) {
		return fmt.Errorf("name must be alphanumeric with hyphens only: %s", c.Name)
	}
	if len(c.Name) > 64 {
		return fmt.Errorf("name must be at most 64 characters: %s", c.Name)
	}

	// Command validation
	if c.Command == "" && c.Transport != "http" {
		return fmt.Errorf("command is required for non-HTTP transports")
	}

	// Transport validation
	if c.Transport != "stdio" && c.Transport != "http" {
		return fmt.Errorf("transport must be 'stdio' or 'http': %s", c.Transport)
	}

	// BaseURL validation for HTTP transport
	if c.Transport == "http" && c.BaseURL == "" {
		return fmt.Errorf("baseURL is required for HTTP transport")
	}

	// Capabilities validation
	if len(c.Capabilities) == 0 {
		return fmt.Errorf("capabilities array must not be empty")
	}
	capPattern := regexp.MustCompile(`^[a-z0-9-]+(\.[a-z0-9-]+)*$`)
	for _, cap := range c.Capabilities {
		if !capPattern.MatchString(cap) {
			return fmt.Errorf("invalid capability format: %s", cap)
		}
	}

	// SchemaVersion validation
	versionPattern := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	if !versionPattern.MatchString(c.SchemaVersion) {
		return fmt.Errorf("schemaVersion must be semantic version: %s", c.SchemaVersion)
	}

	// Timeout validation (if specified)
	if c.Timeout != "" {
		if _, err := time.ParseDuration(c.Timeout); err != nil {
			return fmt.Errorf("invalid timeout duration: %w", err)
		}
	}

	return nil
}

// ToolStatus represents the health status of an MCP tool.
type ToolStatus string

const (
	ToolStatusReady    ToolStatus = "ready"
	ToolStatusDegraded ToolStatus = "degraded"
	ToolStatusOffline  ToolStatus = "offline"
)

// MCPTool represents the runtime state of an active MCP tool connection.
type MCPTool struct {
	Name            string
	Config          *MCPConfig
	Status          ToolStatus
	CircuitBreaker  *CircuitBreaker
	Metrics         ToolMetrics
	LastHealthCheck time.Time
	LastError       error
	Enabled         bool
}

// ToolMetrics contains runtime metrics for an MCP tool.
type ToolMetrics struct {
	HandshakeLatency    time.Duration
	InvocationCount     int64
	SuccessCount        int64
	ErrorCount          int64
	LastInvocationTime  time.Time
	AverageLatency      time.Duration
}

// MCPInvocationLog is an audit record of an MCP tool invocation by an agent.
type MCPInvocationLog struct {
	ID                string
	Timestamp         time.Time
	RunID             string
	AgentID           string
	AgentRole         string
	ToolName          string
	Method            string
	ArgsHash          string // SHA256 hash
	RedactionApplied  bool
	Duration          time.Duration
	Status            string // "success", "error", "permission_denied", "rate_limited"
	ErrorMessage      string
}

// AgentCapability defines which RBAC capabilities an agent role possesses.
type AgentCapability struct {
	Role         string
	Capabilities []string
}

// ToolBudget defines rate limits, concurrency limits, and timeout constraints for an MCP tool.
type ToolBudget struct {
	ToolName         string
	RateLimit        RateLimit
	ConcurrencyLimit int
	Timeout          time.Duration
	DailyQuota       int // 0 = unlimited
}

// RateLimit defines rate limiting parameters.
type RateLimit struct {
	RequestsPerSecond float64
	BurstSize         int
}

// MCPHealthReport contains diagnostic information from a tool health check.
type MCPHealthReport struct {
	ToolName         string
	Status           ToolStatus
	HandshakeLatency time.Duration
	Capabilities     []string
	LastError        error
	Timestamp        time.Time
}

// CircuitBreakerState represents the state of a circuit breaker.
type CircuitBreakerState string

const (
	CircuitBreakerClosed   CircuitBreakerState = "closed"
	CircuitBreakerOpen     CircuitBreakerState = "open"
	CircuitBreakerHalfOpen CircuitBreakerState = "half-open"
)

// CircuitBreaker manages failure handling for MCP tools.
type CircuitBreaker struct {
	State        CircuitBreakerState
	Failures     int
	LastFailTime time.Time
	Successes    int // Used in half-open state
}

// SchemaError represents a validation error with file/line/property context.
type SchemaError struct {
	FilePath string
	Line     int
	Column   int
	JSONPath string
	Message  string
}

func (e SchemaError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s: %s", e.FilePath, e.Line, e.Column, e.JSONPath, e.Message)
}
