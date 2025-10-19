package types

import (
	"strings"
	"testing"
)

// TestValidateAIConfigFeature003 tests validation for Feature 003 AI config fields
func TestValidateAIConfigFeature003(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid context mode",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:   true,
					Provider:  AIProviderAnthropic,
					Model:     "claude-3-opus",
					APIKey:    "test-key",
					Mode:      AIModeContext,
					Timeout:   60,
					RateLimit: 10,
					Concurrency: ConcurrencyLimits{
						MaxAnalyses: 25,
					},
					Budgets: BudgetLimits{
						MaxSources:  50,
						MaxAPICalls: 500,
						MaxTokens:   250000,
					},
					Redaction: RedactionConfig{
						Enabled: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid autonomous mode",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:   true,
					Provider:  AIProviderOpenAI,
					Model:     "gpt-4",
					APIKey:    "test-key",
					Mode:      AIModeAutonomous,
					Timeout:   60,
					RateLimit: 10,
					Concurrency: ConcurrencyLimits{
						MaxAnalyses: 50,
					},
					Budgets: BudgetLimits{
						MaxSources:  100,
						MaxAPICalls: 1000,
						MaxTokens:   500000,
					},
					Autonomous: AutonomousConfig{
						Enabled: true,
						AutoApprove: AutoApproveConfig{
							"github": {"auth*"},
						},
					},
					Redaction: RedactionConfig{
						Enabled: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid AI mode",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:  true,
					Provider: AIProviderOpenAI,
					Model:    "gpt-4",
					APIKey:   "test-key",
					Mode:     "invalid",
					Timeout:  60,
					Concurrency: ConcurrencyLimits{
						MaxAnalyses: 25,
					},
					Budgets: BudgetLimits{
						MaxSources:  50,
						MaxAPICalls: 500,
						MaxTokens:   250000,
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid AI mode",
		},
		{
			name: "zero max analyses",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:  true,
					Provider: AIProviderOpenAI,
					Model:    "gpt-4",
					APIKey:   "test-key",
					Mode:     AIModeContext,
					Timeout:  60,
					Concurrency: ConcurrencyLimits{
						MaxAnalyses: 0,
					},
					Budgets: BudgetLimits{
						MaxSources:  50,
						MaxAPICalls: 500,
						MaxTokens:   250000,
					},
				},
			},
			wantErr: true,
			errMsg:  "concurrency.maxAnalyses must be positive",
		},
		{
			name: "negative max analyses",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:  true,
					Provider: AIProviderOpenAI,
					Model:    "gpt-4",
					APIKey:   "test-key",
					Mode:     AIModeContext,
					Timeout:  60,
					Concurrency: ConcurrencyLimits{
						MaxAnalyses: -1,
					},
					Budgets: BudgetLimits{
						MaxSources:  50,
						MaxAPICalls: 500,
						MaxTokens:   250000,
					},
				},
			},
			wantErr: true,
			errMsg:  "concurrency.maxAnalyses must be positive",
		},
		{
			name: "zero max sources",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:  true,
					Provider: AIProviderOpenAI,
					Model:    "gpt-4",
					APIKey:   "test-key",
					Mode:     AIModeContext,
					Timeout:  60,
					Concurrency: ConcurrencyLimits{
						MaxAnalyses: 25,
					},
					Budgets: BudgetLimits{
						MaxSources:  0,
						MaxAPICalls: 500,
						MaxTokens:   250000,
					},
				},
			},
			wantErr: true,
			errMsg:  "budgets.maxSources must be positive",
		},
		{
			name: "zero max API calls",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:  true,
					Provider: AIProviderOpenAI,
					Model:    "gpt-4",
					APIKey:   "test-key",
					Mode:     AIModeContext,
					Timeout:  60,
					Concurrency: ConcurrencyLimits{
						MaxAnalyses: 25,
					},
					Budgets: BudgetLimits{
						MaxSources:  50,
						MaxAPICalls: 0,
						MaxTokens:   250000,
					},
				},
			},
			wantErr: true,
			errMsg:  "budgets.maxAPICalls must be positive",
		},
		{
			name: "zero max tokens",
			config: &Config{
				LogLevel: "info",
				Theme:    "dark",
				UserRole: RoleComplianceManager,
				Export:   ExportConfig{Format: "json"},
				AI: AIConfig{
					Enabled:  true,
					Provider: AIProviderOpenAI,
					Model:    "gpt-4",
					APIKey:   "test-key",
					Mode:     AIModeContext,
					Timeout:  60,
					Concurrency: ConcurrencyLimits{
						MaxAnalyses: 25,
					},
					Budgets: BudgetLimits{
						MaxSources:  50,
						MaxAPICalls: 500,
						MaxTokens:   0,
					},
				},
			},
			wantErr: true,
			errMsg:  "budgets.maxTokens must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateConfig() error = %v, expected to contain %q", err, tt.errMsg)
				}
			}
		})
	}
}
