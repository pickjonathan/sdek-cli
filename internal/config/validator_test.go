package config

import (
	"testing"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Fatal("NewValidator returned nil")
	}
}

func TestValidateDataDir(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		dataDir string
		wantErr bool
	}{
		{
			name:    "empty data dir",
			dataDir: "",
			wantErr: true,
		},
		{
			name:    "absolute path",
			dataDir: "/home/user/.sdek",
			wantErr: false,
		},
		{
			name:    "$HOME prefix",
			dataDir: "$HOME/.sdek",
			wantErr: false,
		},
		{
			name:    "relative path",
			dataDir: "./data",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDataDir(tt.dataDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDataDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateExportPath(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name       string
		exportPath string
		wantErr    bool
	}{
		{
			name:       "empty export path",
			exportPath: "",
			wantErr:    true,
		},
		{
			name:       "absolute path",
			exportPath: "/home/user/exports",
			wantErr:    false,
		},
		{
			name:       "$HOME prefix",
			exportPath: "$HOME/exports",
			wantErr:    false,
		},
		{
			name:       "relative path with dot",
			exportPath: "./exports",
			wantErr:    false,
		},
		{
			name:       "relative path with double dot",
			exportPath: "../exports",
			wantErr:    false,
		},
		{
			name:       "relative path without dot",
			exportPath: "exports",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateExportPath(tt.exportPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExportPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSources(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		sources []string
		wantErr bool
	}{
		{
			name:    "empty sources list",
			sources: []string{},
			wantErr: true,
		},
		{
			name:    "valid single source",
			sources: []string{types.SourceTypeGit},
			wantErr: false,
		},
		{
			name:    "valid multiple sources",
			sources: []string{types.SourceTypeGit, types.SourceTypeJira, types.SourceTypeSlack},
			wantErr: false,
		},
		{
			name:    "invalid source",
			sources: []string{"invalid"},
			wantErr: true,
		},
		{
			name:    "mix of valid and invalid",
			sources: []string{types.SourceTypeGit, "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSources(tt.sources)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSources() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFrameworks(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name       string
		frameworks []string
		wantErr    bool
	}{
		{
			name:       "empty frameworks list",
			frameworks: []string{},
			wantErr:    true,
		},
		{
			name:       "valid single framework",
			frameworks: []string{types.FrameworkSOC2},
			wantErr:    false,
		},
		{
			name:       "valid multiple frameworks",
			frameworks: []string{types.FrameworkSOC2, types.FrameworkISO27001, types.FrameworkPCIDSS},
			wantErr:    false,
		},
		{
			name:       "invalid framework",
			frameworks: []string{"invalid"},
			wantErr:    true,
		},
		{
			name:       "mix of valid and invalid",
			frameworks: []string{types.FrameworkSOC2, "invalid"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFrameworks(tt.frameworks)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFrameworks() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLogLevel(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		logLevel string
		wantErr  bool
	}{
		{
			name:     "valid debug",
			logLevel: "debug",
			wantErr:  false,
		},
		{
			name:     "valid info",
			logLevel: "info",
			wantErr:  false,
		},
		{
			name:     "valid warn",
			logLevel: "warn",
			wantErr:  false,
		},
		{
			name:     "valid error",
			logLevel: "error",
			wantErr:  false,
		},
		{
			name:     "invalid level",
			logLevel: "invalid",
			wantErr:  true,
		},
		{
			name:     "empty level",
			logLevel: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateLogLevel(tt.logLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLogLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTheme(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		theme   string
		wantErr bool
	}{
		{
			name:    "valid dark",
			theme:   "dark",
			wantErr: false,
		},
		{
			name:    "valid light",
			theme:   "light",
			wantErr: false,
		},
		{
			name:    "invalid theme",
			theme:   "blue",
			wantErr: true,
		},
		{
			name:    "empty theme",
			theme:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTheme(tt.theme)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTheme() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUserRole(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{
			name:    "valid compliance_manager",
			role:    types.RoleComplianceManager,
			wantErr: false,
		},
		{
			name:    "valid engineer",
			role:    types.RoleEngineer,
			wantErr: false,
		},
		{
			name:    "invalid role",
			role:    "admin",
			wantErr: true,
		},
		{
			name:    "empty role",
			role:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateUserRole(tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUserRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateExportFormat(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "valid json",
			format:  "json",
			wantErr: false,
		},
		{
			name:    "valid yaml",
			format:  "yaml",
			wantErr: false,
		},
		{
			name:    "invalid format",
			format:  "xml",
			wantErr: true,
		},
		{
			name:    "empty format",
			format:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateExportFormat(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExportFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFullConfig(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		config  *types.Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  types.DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: &types.Config{
				DataDir:  "$HOME/.sdek",
				LogLevel: "invalid",
				Theme:    "dark",
				UserRole: types.RoleComplianceManager,
				Export: types.ExportConfig{
					DefaultPath: "$HOME/exports",
					Format:      "json",
				},
				Sources: types.SourcesConfig{
					Enabled: []string{types.SourceTypeGit},
				},
				Frameworks: types.FrameworksConfig{
					Enabled: []string{types.FrameworkSOC2},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid theme",
			config: &types.Config{
				DataDir:  "$HOME/.sdek",
				LogLevel: "info",
				Theme:    "invalid",
				UserRole: types.RoleComplianceManager,
				Export: types.ExportConfig{
					DefaultPath: "$HOME/exports",
					Format:      "json",
				},
				Sources: types.SourcesConfig{
					Enabled: []string{types.SourceTypeGit},
				},
				Frameworks: types.FrameworksConfig{
					Enabled: []string{types.FrameworkSOC2},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid user role",
			config: &types.Config{
				DataDir:  "$HOME/.sdek",
				LogLevel: "info",
				Theme:    "dark",
				UserRole: "invalid",
				Export: types.ExportConfig{
					DefaultPath: "$HOME/exports",
					Format:      "json",
				},
				Sources: types.SourcesConfig{
					Enabled: []string{types.SourceTypeGit},
				},
				Frameworks: types.FrameworksConfig{
					Enabled: []string{types.FrameworkSOC2},
				},
			},
			wantErr: true,
		},
		{
			name: "empty sources",
			config: &types.Config{
				DataDir:  "$HOME/.sdek",
				LogLevel: "info",
				Theme:    "dark",
				UserRole: types.RoleComplianceManager,
				Export: types.ExportConfig{
					DefaultPath: "$HOME/exports",
					Format:      "json",
				},
				Sources: types.SourcesConfig{
					Enabled: []string{},
				},
				Frameworks: types.FrameworksConfig{
					Enabled: []string{types.FrameworkSOC2},
				},
			},
			wantErr: true,
		},
		{
			name: "empty frameworks",
			config: &types.Config{
				DataDir:  "$HOME/.sdek",
				LogLevel: "info",
				Theme:    "dark",
				UserRole: types.RoleComplianceManager,
				Export: types.ExportConfig{
					DefaultPath: "$HOME/exports",
					Format:      "json",
				},
				Sources: types.SourcesConfig{
					Enabled: []string{types.SourceTypeGit},
				},
				Frameworks: types.FrameworksConfig{
					Enabled: []string{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
