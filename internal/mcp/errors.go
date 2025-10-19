package mcp

import "errors"

var (
	// ErrToolNotFound is returned when a requested tool doesn't exist in the registry.
	ErrToolNotFound = errors.New("mcp: tool not found")

	// ErrToolDisabled is returned when attempting to invoke an administratively disabled tool.
	ErrToolDisabled = errors.New("mcp: tool is disabled")

	// ErrInvalidConfig is returned when a configuration fails validation.
	ErrInvalidConfig = errors.New("mcp: invalid configuration")

	// ErrHandshakeFailed is returned when the initial handshake with an MCP server fails.
	ErrHandshakeFailed = errors.New("mcp: handshake failed")

	// ErrPermissionDenied is returned when an agent lacks required capabilities.
	ErrPermissionDenied = errors.New("mcp: permission denied")

	// ErrCircuitOpen is returned when a circuit breaker is open (tool offline).
	ErrCircuitOpen = errors.New("mcp: circuit breaker open")

	// ErrRateLimited is returned when a rate limit is exceeded.
	ErrRateLimited = errors.New("mcp: rate limit exceeded")

	// ErrNoConfigDirs is returned when no config directories are found or accessible.
	ErrNoConfigDirs = errors.New("mcp: no config directories found")

	// ErrSchemaLoadFailed is returned when the JSON schema cannot be loaded.
	ErrSchemaLoadFailed = errors.New("mcp: failed to load schema")
)
