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
	if config.AI.Autonomous.AutoApprove.Enabled {
		t.Error("expected auto-approve to be disabled by default")
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
		{
			name: "zero rate limit - valid (unlimited)",
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
					RateLimit: 0,
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
