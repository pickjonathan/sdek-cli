package types

import (
	"testing"
)

func TestDefaultMCPConfig(t *testing.T) {
	cfg := DefaultMCPConfig()

	// Test default values
	if !cfg.Enabled {
		t.Error("Expected Enabled to be true by default")
	}

	if !cfg.PreferMCP {
		t.Error("Expected PreferMCP to be true by default")
	}

	if cfg.MaxConcurrent != 10 {
		t.Errorf("Expected MaxConcurrent to be 10, got %d", cfg.MaxConcurrent)
	}

	if cfg.HealthCheckInterval != 300 {
		t.Errorf("Expected HealthCheckInterval to be 300, got %d", cfg.HealthCheckInterval)
	}

	// Test retry config defaults
	if cfg.Retry.MaxAttempts != 3 {
		t.Errorf("Expected Retry.MaxAttempts to be 3, got %d", cfg.Retry.MaxAttempts)
	}

	if cfg.Retry.Backoff != "exponential" {
		t.Errorf("Expected Retry.Backoff to be 'exponential', got '%s'", cfg.Retry.Backoff)
	}

	if cfg.Retry.InitialDelayMS != 1000 {
		t.Errorf("Expected Retry.InitialDelayMS to be 1000, got %d", cfg.Retry.InitialDelayMS)
	}

	if cfg.Retry.MaxDelayMS != 30000 {
		t.Errorf("Expected Retry.MaxDelayMS to be 30000, got %d", cfg.Retry.MaxDelayMS)
	}

	// Test servers map is initialized
	if cfg.Servers == nil {
		t.Error("Expected Servers map to be initialized")
	}
}

func TestDefaultMCPServerConfig(t *testing.T) {
	cfg := DefaultMCPServerConfig()

	if cfg.Transport != "stdio" {
		t.Errorf("Expected Transport to be 'stdio', got '%s'", cfg.Transport)
	}

	if cfg.Timeout != 60 {
		t.Errorf("Expected Timeout to be 60, got %d", cfg.Timeout)
	}

	if cfg.RateLimit != 0 {
		t.Errorf("Expected RateLimit to be 0, got %d", cfg.RateLimit)
	}

	if cfg.Env == nil {
		t.Error("Expected Env map to be initialized")
	}

	if cfg.Headers == nil {
		t.Error("Expected Headers map to be initialized")
	}
}

func TestMCPConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *MCPConfig
		wantError bool
		errorMsg  string
	}{
		{
			name:      "nil config is valid",
			config:    nil,
			wantError: false,
		},
		{
			name: "valid config",
			config: &MCPConfig{
				Enabled:             true,
				PreferMCP:           true,
				MaxConcurrent:       10,
				HealthCheckInterval: 300,
				Servers: map[string]MCPServerConfig{
					"test-server": {
						Command:   "test",
						Transport: "stdio",
						Timeout:   60,
					},
				},
			},
			wantError: false,
		},
		{
			name: "max_concurrent too low",
			config: &MCPConfig{
				MaxConcurrent:       0,
				HealthCheckInterval: 300,
				Servers:             map[string]MCPServerConfig{},
			},
			wantError: true,
			errorMsg:  "max_concurrent must be between 1 and 100",
		},
		{
			name: "max_concurrent too high",
			config: &MCPConfig{
				MaxConcurrent:       101,
				HealthCheckInterval: 300,
				Servers:             map[string]MCPServerConfig{},
			},
			wantError: true,
			errorMsg:  "max_concurrent must be between 1 and 100",
		},
		{
			name: "health_check_interval too low",
			config: &MCPConfig{
				MaxConcurrent:       10,
				HealthCheckInterval: 30,
				Servers:             map[string]MCPServerConfig{},
			},
			wantError: true,
			errorMsg:  "health_check_interval must be >= 60 seconds",
		},
		{
			name: "invalid transport type",
			config: &MCPConfig{
				MaxConcurrent:       10,
				HealthCheckInterval: 300,
				Servers: map[string]MCPServerConfig{
					"test": {
						Transport: "invalid",
						Timeout:   60,
					},
				},
			},
			wantError: true,
			errorMsg:  "transport must be 'stdio' or 'http'",
		},
		{
			name: "stdio missing command",
			config: &MCPConfig{
				MaxConcurrent:       10,
				HealthCheckInterval: 300,
				Servers: map[string]MCPServerConfig{
					"test": {
						Transport: "stdio",
						Timeout:   60,
					},
				},
			},
			wantError: true,
			errorMsg:  "command is required for stdio transport",
		},
		{
			name: "stdio with url should fail",
			config: &MCPConfig{
				MaxConcurrent:       10,
				HealthCheckInterval: 300,
				Servers: map[string]MCPServerConfig{
					"test": {
						Command:   "test",
						URL:       "http://example.com",
						Transport: "stdio",
						Timeout:   60,
					},
				},
			},
			wantError: true,
			errorMsg:  "url must be empty for stdio transport",
		},
		{
			name: "http missing url",
			config: &MCPConfig{
				MaxConcurrent:       10,
				HealthCheckInterval: 300,
				Servers: map[string]MCPServerConfig{
					"test": {
						Transport: "http",
						Timeout:   60,
					},
				},
			},
			wantError: true,
			errorMsg:  "url is required for http transport",
		},
		{
			name: "http with command should fail",
			config: &MCPConfig{
				MaxConcurrent:       10,
				HealthCheckInterval: 300,
				Servers: map[string]MCPServerConfig{
					"test": {
						Command:   "test",
						URL:       "http://example.com",
						Transport: "http",
						Timeout:   60,
					},
				},
			},
			wantError: true,
			errorMsg:  "command must be empty for http transport",
		},
		{
			name: "timeout too low",
			config: &MCPConfig{
				MaxConcurrent:       10,
				HealthCheckInterval: 300,
				Servers: map[string]MCPServerConfig{
					"test": {
						Command:   "test",
						Transport: "stdio",
						Timeout:   0,
					},
				},
			},
			wantError: true,
			errorMsg:  "timeout must be between 1 and 600 seconds",
		},
		{
			name: "timeout too high",
			config: &MCPConfig{
				MaxConcurrent:       10,
				HealthCheckInterval: 300,
				Servers: map[string]MCPServerConfig{
					"test": {
						Command:   "test",
						Transport: "stdio",
						Timeout:   601,
					},
				},
			},
			wantError: true,
			errorMsg:  "timeout must be between 1 and 600 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMCPConfig(tt.config)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestMCPServerConfigStdio(t *testing.T) {
	cfg := MCPServerConfig{
		Command:   "uvx",
		Args:      []string{"aws-api-mcp-server"},
		Transport: "stdio",
		Timeout:   60,
		Env: map[string]string{
			"AWS_PROFILE": "readonly",
		},
	}

	if cfg.Command != "uvx" {
		t.Errorf("Expected Command to be 'uvx', got '%s'", cfg.Command)
	}

	if len(cfg.Args) != 1 || cfg.Args[0] != "aws-api-mcp-server" {
		t.Errorf("Expected Args to be ['aws-api-mcp-server'], got %v", cfg.Args)
	}

	if cfg.Transport != "stdio" {
		t.Errorf("Expected Transport to be 'stdio', got '%s'", cfg.Transport)
	}

	if cfg.Env["AWS_PROFILE"] != "readonly" {
		t.Errorf("Expected AWS_PROFILE to be 'readonly', got '%s'", cfg.Env["AWS_PROFILE"])
	}
}

func TestMCPServerConfigHTTP(t *testing.T) {
	cfg := MCPServerConfig{
		URL:       "https://mcp.example.com/api",
		Transport: "http",
		Timeout:   30,
		Headers: map[string]string{
			"Authorization": "Bearer token123",
		},
		HealthURL: "https://mcp.example.com/health",
	}

	if cfg.URL != "https://mcp.example.com/api" {
		t.Errorf("Expected URL to be 'https://mcp.example.com/api', got '%s'", cfg.URL)
	}

	if cfg.Transport != "http" {
		t.Errorf("Expected Transport to be 'http', got '%s'", cfg.Transport)
	}

	if cfg.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("Expected Authorization header, got '%s'", cfg.Headers["Authorization"])
	}

	if cfg.HealthURL != "https://mcp.example.com/health" {
		t.Errorf("Expected HealthURL to be 'https://mcp.example.com/health', got '%s'", cfg.HealthURL)
	}
}

func TestRetryConfig(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:    5,
		Backoff:        "linear",
		InitialDelayMS: 500,
		MaxDelayMS:     15000,
	}

	if cfg.MaxAttempts != 5 {
		t.Errorf("Expected MaxAttempts to be 5, got %d", cfg.MaxAttempts)
	}

	if cfg.Backoff != "linear" {
		t.Errorf("Expected Backoff to be 'linear', got '%s'", cfg.Backoff)
	}

	if cfg.InitialDelayMS != 500 {
		t.Errorf("Expected InitialDelayMS to be 500, got %d", cfg.InitialDelayMS)
	}

	if cfg.MaxDelayMS != 15000 {
		t.Errorf("Expected MaxDelayMS to be 15000, got %d", cfg.MaxDelayMS)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
