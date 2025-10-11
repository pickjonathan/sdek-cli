package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Validator validates configuration values
type Validator struct{}

// NewValidator creates a new configuration validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates the entire configuration
func (v *Validator) Validate(config *types.Config) error {
	if err := types.ValidateConfig(config); err != nil {
		return err
	}

	// Additional validation beyond what's in ValidateConfig
	if err := v.ValidateDataDir(config.DataDir); err != nil {
		return fmt.Errorf("invalid data_dir: %w", err)
	}

	if err := v.ValidateExportPath(config.Export.DefaultPath); err != nil {
		return fmt.Errorf("invalid export.default_path: %w", err)
	}

	if err := v.ValidateSources(config.Sources.Enabled); err != nil {
		return fmt.Errorf("invalid sources.enabled: %w", err)
	}

	if err := v.ValidateFrameworks(config.Frameworks.Enabled); err != nil {
		return fmt.Errorf("invalid frameworks.enabled: %w", err)
	}

	return nil
}

// ValidateDataDir validates the data directory path
func (v *Validator) ValidateDataDir(dataDir string) error {
	if dataDir == "" {
		return fmt.Errorf("data_dir cannot be empty")
	}

	// Expand $HOME if present
	expanded := os.ExpandEnv(dataDir)

	// Check if path is absolute or starts with $HOME
	if !filepath.IsAbs(expanded) && !strings.HasPrefix(dataDir, "$HOME") {
		return fmt.Errorf("data_dir must be an absolute path or start with $HOME, got: %s", dataDir)
	}

	return nil
}

// ValidateExportPath validates the export path
func (v *Validator) ValidateExportPath(exportPath string) error {
	if exportPath == "" {
		return fmt.Errorf("export path cannot be empty")
	}

	// Expand $HOME if present
	expanded := os.ExpandEnv(exportPath)

	// Check if path is absolute or starts with $HOME or is relative
	if !filepath.IsAbs(expanded) && !strings.HasPrefix(exportPath, "$HOME") && !strings.HasPrefix(exportPath, ".") {
		return fmt.Errorf("export path must be an absolute path, start with $HOME, or be a relative path, got: %s", exportPath)
	}

	return nil
}

// ValidateSources validates the enabled sources list
func (v *Validator) ValidateSources(enabled []string) error {
	if len(enabled) == 0 {
		return fmt.Errorf("at least one source must be enabled")
	}

	// Check each enabled source is valid
	validSources := make(map[string]bool)
	for _, s := range types.ValidSourceTypes {
		validSources[s] = true
	}

	for _, source := range enabled {
		if !validSources[source] {
			return fmt.Errorf("invalid source: %s (valid sources: %v)", source, types.ValidSourceTypes)
		}
	}

	return nil
}

// ValidateFrameworks validates the enabled frameworks list
func (v *Validator) ValidateFrameworks(enabled []string) error {
	if len(enabled) == 0 {
		return fmt.Errorf("at least one framework must be enabled")
	}

	// Check each enabled framework is valid
	validFrameworks := make(map[string]bool)
	for _, f := range types.ValidFrameworkIDs {
		validFrameworks[f] = true
	}

	for _, framework := range enabled {
		if !validFrameworks[framework] {
			return fmt.Errorf("invalid framework: %s (valid frameworks: %v)", framework, types.ValidFrameworkIDs)
		}
	}

	return nil
}

// ValidateLogLevel validates the log level
func (v *Validator) ValidateLogLevel(level string) error {
	validLevels := []string{"debug", "info", "warn", "error"}
	for _, valid := range validLevels {
		if level == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid log level: %s (valid levels: %v)", level, validLevels)
}

// ValidateTheme validates the theme
func (v *Validator) ValidateTheme(theme string) error {
	validThemes := []string{"dark", "light"}
	for _, valid := range validThemes {
		if theme == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid theme: %s (valid themes: %v)", theme, validThemes)
}

// ValidateUserRole validates the user role
func (v *Validator) ValidateUserRole(role string) error {
	validRoles := []string{types.RoleComplianceManager, types.RoleEngineer}
	for _, valid := range validRoles {
		if role == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid user role: %s (valid roles: %v)", role, validRoles)
}

// ValidateExportFormat validates the export format
func (v *Validator) ValidateExportFormat(format string) error {
	validFormats := []string{"json", "yaml"}
	for _, valid := range validFormats {
		if format == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid export format: %s (valid formats: %v)", format, validFormats)
}
