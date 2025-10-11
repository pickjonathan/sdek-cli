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
}
