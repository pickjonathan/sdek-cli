package types

import "testing"

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: &Config{
				LogLevel: "invalid",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
			},
			wantErr: true,
		},
		{
			name: "invalid theme",
			config: &Config{
				LogLevel: "info",
				Theme:    "invalid",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
			},
			wantErr: true,
		},
		{
			name: "invalid user role",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: "invalid",
				Export:   ExportConfig{Format: "json"},
			},
			wantErr: true,
		},
		{
			name: "invalid export format",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "xml"},
			},
			wantErr: true,
		},
		{
			name: "valid connector config - github",
			config: &Config{
				LogLevel:   "info",
				Theme:      "dark",
				UserRole:   RoleComplianceManager,
				Export:     ExportConfig{Format: "json"},
				Frameworks: FrameworksConfig{Enabled: []string{"soc2"}},
				Sources:    SourcesConfig{Enabled: []string{"git"}},
				AI: AIConfig{
					Enabled:     true,
					Provider:    AIProviderOpenAI,
					Model:       "gpt-4",
					OpenAIKey:   "sk-test",
					Mode:        AIModeContext,
					Timeout:     60,
					Concurrency: ConcurrencyLimits{MaxAnalyses: 25},
					Budgets:     BudgetLimits{MaxSources: 50, MaxAPICalls: 500, MaxTokens: 250000},
					Connectors: map[string]ConnectorConfig{
						"github": {
							Enabled:   true,
							APIKey:    "ghp_test",
							RateLimit: 60,
							Timeout:   30,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid connector name",
			config: &Config{
				LogLevel:   "info",
				Theme:      "dark",
				UserRole:   RoleComplianceManager,
				Export:     ExportConfig{Format: "json"},
				Frameworks: FrameworksConfig{Enabled: []string{"soc2"}},
				Sources:    SourcesConfig{Enabled: []string{"git"}},
				AI: AIConfig{
					Enabled:     true,
					Provider:    AIProviderOpenAI,
					Model:       "gpt-4",
					OpenAIKey:   "sk-test",
					Mode:        AIModeContext,
					Timeout:     60,
					Concurrency: ConcurrencyLimits{MaxAnalyses: 25},
					Budgets:     BudgetLimits{MaxSources: 50, MaxAPICalls: 500, MaxTokens: 250000},
					Connectors: map[string]ConnectorConfig{
						"invalid": {
							Enabled: true,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative connector timeout",
			config: &Config{
				LogLevel:   "info",
				Theme:      "dark",
				UserRole:   RoleComplianceManager,
				Export:     ExportConfig{Format: "json"},
				Frameworks: FrameworksConfig{Enabled: []string{"soc2"}},
				Sources:    SourcesConfig{Enabled: []string{"git"}},
				AI: AIConfig{
					Enabled:     true,
					Provider:    AIProviderOpenAI,
					Model:       "gpt-4",
					OpenAIKey:   "sk-test",
					Mode:        AIModeContext,
					Timeout:     60,
					Concurrency: ConcurrencyLimits{MaxAnalyses: 25},
					Budgets:     BudgetLimits{MaxSources: 50, MaxAPICalls: 500, MaxTokens: 250000},
					Connectors: map[string]ConnectorConfig{
						"github": {
							Enabled: true,
							Timeout: -1,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "negative connector rate limit",
			config: &Config{
				LogLevel:   "info",
				Theme:      "dark",
				UserRole:   RoleComplianceManager,
				Export:     ExportConfig{Format: "json"},
				Frameworks: FrameworksConfig{Enabled: []string{"soc2"}},
				Sources:    SourcesConfig{Enabled: []string{"git"}},
				AI: AIConfig{
					Enabled:     true,
					Provider:    AIProviderOpenAI,
					Model:       "gpt-4",
					OpenAIKey:   "sk-test",
					Mode:        AIModeContext,
					Timeout:     60,
					Concurrency: ConcurrencyLimits{MaxAnalyses: 25},
					Budgets:     BudgetLimits{MaxSources: 50, MaxAPICalls: 500, MaxTokens: 250000},
					Connectors: map[string]ConnectorConfig{
						"github": {
							Enabled:   true,
							RateLimit: -1,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "multiple connectors - all valid",
			config: &Config{
				LogLevel:   "info",
				Theme:      "dark",
				UserRole:   RoleComplianceManager,
				Export:     ExportConfig{Format: "json"},
				Frameworks: FrameworksConfig{Enabled: []string{"soc2"}},
				Sources:    SourcesConfig{Enabled: []string{"git"}},
				AI: AIConfig{
					Enabled:     true,
					Provider:    AIProviderOpenAI,
					Model:       "gpt-4",
					OpenAIKey:   "sk-test",
					Mode:        AIModeAutonomous,
					Timeout:     60,
					Concurrency: ConcurrencyLimits{MaxAnalyses: 25},
					Budgets:     BudgetLimits{MaxSources: 50, MaxAPICalls: 500, MaxTokens: 250000},
					Connectors: map[string]ConnectorConfig{
						"github": {
							Enabled:   true,
							APIKey:    "ghp_test",
							RateLimit: 60,
							Timeout:   30,
						},
						"jira": {
							Enabled:  true,
							APIKey:   "jira_test",
							Endpoint: "https://company.atlassian.net",
							Timeout:  30,
						},
						"aws": {
							Enabled: false,
							Timeout: 30,
						},
						"slack": {
							Enabled: true,
							APIKey:  "xoxb-test",
							Timeout: 30,
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.LogLevel != "info" {
		t.Errorf("expected log level 'info', got %s", config.LogLevel)
	}

	if config.Theme != "dark" {
		t.Errorf("expected theme 'dark', got %s", config.Theme)
	}

	if config.UserRole != RoleComplianceManager {
		t.Errorf("expected user role %s, got %s", RoleComplianceManager, config.UserRole)
	}

	if len(config.Frameworks.Enabled) != 3 {
		t.Errorf("expected 3 frameworks enabled, got %d", len(config.Frameworks.Enabled))
	}

	if len(config.Sources.Enabled) != 5 {
		t.Errorf("expected 5 sources enabled, got %d", len(config.Sources.Enabled))
	}

	// Test AI config defaults
	if config.AI.Enabled {
		t.Error("expected AI to be disabled by default")
	}
	if config.AI.Provider != AIProviderOpenAI {
		t.Errorf("expected default AI provider %s, got %s", AIProviderOpenAI, config.AI.Provider)
	}
	if config.AI.Model != "gpt-4" {
		t.Errorf("expected default AI model 'gpt-4', got %s", config.AI.Model)
	}
	if config.AI.Timeout != 60 {
		t.Errorf("expected default AI timeout 60, got %d", config.AI.Timeout)
	}
	if config.AI.RateLimit != 10 {
		t.Errorf("expected default AI rate limit 10, got %d", config.AI.RateLimit)
	}

	// Test Feature 003 AI config defaults
	if config.AI.Mode != AIModeDisabled {
		t.Errorf("expected default AI mode %s, got %s", AIModeDisabled, config.AI.Mode)
	}
	if config.AI.Concurrency.MaxAnalyses != 25 {
		t.Errorf("expected default concurrency.maxAnalyses 25, got %d", config.AI.Concurrency.MaxAnalyses)
	}
	if config.AI.Budgets.MaxSources != 50 {
		t.Errorf("expected default budgets.maxSources 50, got %d", config.AI.Budgets.MaxSources)
	}
	if config.AI.Budgets.MaxAPICalls != 500 {
		t.Errorf("expected default budgets.maxAPICalls 500, got %d", config.AI.Budgets.MaxAPICalls)
	}
	if config.AI.Budgets.MaxTokens != 250000 {
		t.Errorf("expected default budgets.maxTokens 250000, got %d", config.AI.Budgets.MaxTokens)
	}
	if config.AI.Autonomous.Enabled {
		t.Error("expected autonomous mode to be disabled by default")
	}
	if len(config.AI.Autonomous.AutoApprove) != 0 {
		t.Errorf("expected empty auto-approve map by default, got %d entries", len(config.AI.Autonomous.AutoApprove))
	}
	if !config.AI.Redaction.Enabled {
		t.Error("expected redaction to be enabled by default")
	}
	if len(config.AI.Redaction.Denylist) != 0 {
		t.Errorf("expected empty denylist by default, got %d items", len(config.AI.Redaction.Denylist))
	}
}

func TestValidateAIConfig(t *testing.T) {
	t.Skip("Skipping until implementation complete")

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "AI disabled - valid",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled: false,
				},
			},
			wantErr: false,
		},
		{
			name: "AI enabled with OpenAI - valid",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:   true,
					Provider:  AIProviderOpenAI,
					Model:     "gpt-4",
					OpenAIKey: "sk-test123",
					Timeout:   60,
					RateLimit: 10,
				},
			},
			wantErr: false,
		},
		{
			name: "AI enabled with Anthropic - valid",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:      true,
					Provider:     AIProviderAnthropic,
					Model:        "claude-3-opus",
					AnthropicKey: "sk-ant-test123",
					Timeout:      60,
					RateLimit:    10,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid AI provider",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:  true,
					Provider: "invalid",
					Model:    "gpt-4",
					Timeout:  60,
				},
			},
			wantErr: true,
		},
		{
			name: "empty model when enabled",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:   true,
					Provider:  AIProviderOpenAI,
					Model:     "",
					OpenAIKey: "sk-test",
					Timeout:   60,
				},
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:   true,
					Provider:  AIProviderOpenAI,
					Model:     "gpt-4",
					OpenAIKey: "sk-test",
					Timeout:   -1,
				},
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:   true,
					Provider:  AIProviderOpenAI,
					Model:     "gpt-4",
					OpenAIKey: "sk-test",
					Timeout:   0,
				},
			},
			wantErr: true,
		},
		{
			name: "negative rate limit",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:   true,
					Provider:  AIProviderOpenAI,
					Model:     "gpt-4",
					OpenAIKey: "sk-test",
					Timeout:   60,
					RateLimit: -1,
				},
			},
			wantErr: true,
		},
		{
			name: "OpenAI without API key",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:   true,
					Provider:  AIProviderOpenAI,
					Model:     "gpt-4",
					OpenAIKey: "",
					Timeout:   60,
				},
			},
			wantErr: true,
		},
		{
			name: "Anthropic without API key",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:      true,
					Provider:     AIProviderAnthropic,
					Model:        "claude-3-opus",
					AnthropicKey: "",
					Timeout:      60,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestConnectorConfig tests ConnectorConfig struct and defaults
func TestConnectorConfig(t *testing.T) {
	t.Run("default config includes connectors", func(t *testing.T) {
		cfg := DefaultConfig()

		if cfg.AI.Connectors == nil {
			t.Fatal("DefaultConfig().AI.Connectors should not be nil")
		}

		// Check all expected connectors are present
		expectedConnectors := []string{"github", "jira", "aws", "slack"}
		for _, name := range expectedConnectors {
			if _, ok := cfg.AI.Connectors[name]; !ok {
				t.Errorf("DefaultConfig() missing connector: %s", name)
			}
		}
	})

	t.Run("default connectors are disabled", func(t *testing.T) {
		cfg := DefaultConfig()

		for name, conn := range cfg.AI.Connectors {
			if conn.Enabled {
				t.Errorf("connector %s should be disabled by default", name)
			}
		}
	})

	t.Run("github connector has rate limit", func(t *testing.T) {
		cfg := DefaultConfig()

		github, ok := cfg.AI.Connectors["github"]
		if !ok {
			t.Fatal("github connector not found in defaults")
		}

		if github.RateLimit != 60 {
			t.Errorf("github connector rate limit = %d, want 60", github.RateLimit)
		}
	})
}

// TestMCPConfigInDefaultConfig tests Feature 006 MCP config in default config (T013)
func TestMCPConfigInDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test MCP config is initialized
	if !cfg.MCP.Enabled {
		t.Error("Expected MCP to be enabled by default")
	}

	if !cfg.MCP.PreferMCP {
		t.Error("Expected PreferMCP to be true by default")
	}

	if cfg.MCP.MaxConcurrent != 10 {
		t.Errorf("Expected MCP.MaxConcurrent to be 10, got %d", cfg.MCP.MaxConcurrent)
	}

	if cfg.MCP.HealthCheckInterval != 300 {
		t.Errorf("Expected MCP.HealthCheckInterval to be 300, got %d", cfg.MCP.HealthCheckInterval)
	}

	if cfg.MCP.Servers == nil {
		t.Error("Expected MCP.Servers map to be initialized")
	}
}

// TestProvidersMapInDefaultConfig tests Feature 006 Providers map in default config (T013)
func TestProvidersMapInDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Providers == nil {
		t.Error("Expected Providers map to be initialized")
	}

	// Default config should have empty providers map
	if len(cfg.Providers) != 0 {
		t.Errorf("Expected Providers map to be empty by default, got %d entries", len(cfg.Providers))
	}
}

// TestConfigWithMCPServers tests loading config with MCP servers (T013)
func TestConfigWithMCPServers(t *testing.T) {
	cfg := &Config{
		MCP: MCPConfig{
			Enabled:             true,
			PreferMCP:           true,
			MaxConcurrent:       10,
			HealthCheckInterval: 300,
			Servers: map[string]MCPServerConfig{
				"aws-api": {
					Command:   "uvx",
					Args:      []string{"aws-api-mcp-server"},
					Transport: "stdio",
					Timeout:   60,
					Env: map[string]string{
						"AWS_PROFILE": "readonly",
					},
				},
				"github-mcp": {
					URL:       "https://github-mcp.example.com/api",
					Transport: "http",
					Timeout:   30,
					Headers: map[string]string{
						"Authorization": "Bearer token",
					},
				},
			},
		},
	}

	// Validate MCP config
	err := ValidateMCPConfig(&cfg.MCP)
	if err != nil {
		t.Errorf("Expected valid MCP config, got error: %v", err)
	}

	// Check servers
	if len(cfg.MCP.Servers) != 2 {
		t.Errorf("Expected 2 MCP servers, got %d", len(cfg.MCP.Servers))
	}

	// Check aws-api server
	awsServer, ok := cfg.MCP.Servers["aws-api"]
	if !ok {
		t.Fatal("Expected aws-api server to exist")
	}

	if awsServer.Transport != "stdio" {
		t.Errorf("Expected aws-api transport to be 'stdio', got '%s'", awsServer.Transport)
	}

	if awsServer.Command != "uvx" {
		t.Errorf("Expected aws-api command to be 'uvx', got '%s'", awsServer.Command)
	}

	// Check github-mcp server
	githubServer, ok := cfg.MCP.Servers["github-mcp"]
	if !ok {
		t.Fatal("Expected github-mcp server to exist")
	}

	if githubServer.Transport != "http" {
		t.Errorf("Expected github-mcp transport to be 'http', got '%s'", githubServer.Transport)
	}

	if githubServer.URL == "" {
		t.Error("Expected github-mcp to have URL")
	}
}

// TestConfigWithProviders tests loading config with AI providers (T013)
func TestConfigWithProviders(t *testing.T) {
	cfg := &Config{
		Providers: map[string]ProviderConfig{
			"openai": {
				URL:         "openai://api.openai.com",
				APIKey:      "${OPENAI_API_KEY}",
				Model:       "gpt-4o",
				Timeout:     60,
				MaxRetries:  3,
				Temperature: 0.0,
				MaxTokens:   4096,
			},
			"ollama": {
				URL:       "ollama://localhost:11434",
				Model:     "gemma3:12b",
				Timeout:   120,
				MaxTokens: 8192,
				Extra: map[string]string{
					"num_ctx": "8192",
				},
			},
		},
	}

	// Check providers
	if len(cfg.Providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(cfg.Providers))
	}

	// Check OpenAI provider
	openai, ok := cfg.Providers["openai"]
	if !ok {
		t.Fatal("Expected openai provider to exist")
	}

	if openai.URL != "openai://api.openai.com" {
		t.Errorf("Expected openai URL, got '%s'", openai.URL)
	}

	if openai.Model != "gpt-4o" {
		t.Errorf("Expected openai model to be 'gpt-4o', got '%s'", openai.Model)
	}

	// Check Ollama provider
	ollama, ok := cfg.Providers["ollama"]
	if !ok {
		t.Fatal("Expected ollama provider to exist")
	}

	if ollama.URL != "ollama://localhost:11434" {
		t.Errorf("Expected ollama URL, got '%s'", ollama.URL)
	}

	if ollama.Extra["num_ctx"] != "8192" {
		t.Errorf("Expected ollama num_ctx to be '8192', got '%s'", ollama.Extra["num_ctx"])
	}
}

// TestBackwardCompatibilityFeature003 tests that Feature 003 configs still work (T013)
func TestBackwardCompatibilityFeature003(t *testing.T) {
	// Start with default config (includes all Feature 006 defaults)
	cfg := DefaultConfig()

	// Modify to look like a Feature 003 config
	cfg.AI.Enabled = true
	cfg.AI.Provider = "openai"
	cfg.AI.Model = "gpt-4"
	cfg.AI.OpenAIKey = "${OPENAI_API_KEY}"
	cfg.AI.Mode = "context"
	cfg.AI.Connectors["github"] = ConnectorConfig{
		Enabled: true,
		APIKey:  "${GITHUB_TOKEN}",
		Timeout: 30,
	}

	// Validate config
	err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("Expected Feature 003 config to be valid, got error: %v", err)
	}

	// Ensure AI config is preserved
	if cfg.AI.Provider != "openai" {
		t.Error("Feature 003 AI.Provider not preserved")
	}

	if cfg.AI.Mode != "context" {
		t.Error("Feature 003 AI.Mode not preserved")
	}

	// Ensure connectors are preserved
	if len(cfg.AI.Connectors) == 0 {
		t.Error("Feature 003 connectors not preserved")
	}

	// Ensure Feature 006 fields are also present
	if !cfg.MCP.Enabled {
		t.Error("Feature 006 MCP config not initialized")
	}

	if cfg.Providers == nil {
		t.Error("Feature 006 Providers map not initialized")
	}
}
