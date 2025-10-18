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
