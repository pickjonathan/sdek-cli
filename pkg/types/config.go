package types

import "fmt"

// Config represents the application configuration
type Config struct {
	DataDir    string           `json:"data_dir" mapstructure:"data_dir"`
	LogLevel   string           `json:"log_level" mapstructure:"log_level"`
	Theme      string           `json:"theme" mapstructure:"theme"`
	UserRole   string           `json:"user_role" mapstructure:"user_role"`
	Export     ExportConfig     `json:"export" mapstructure:"export"`
	Frameworks FrameworksConfig `json:"frameworks" mapstructure:"frameworks"`
	Sources    SourcesConfig    `json:"sources" mapstructure:"sources"`
}

// ExportConfig contains export-related settings
type ExportConfig struct {
	DefaultPath string `json:"default_path" mapstructure:"default_path"`
	Format      string `json:"format" mapstructure:"format"`
}

// FrameworksConfig contains framework-related settings
type FrameworksConfig struct {
	Enabled []string `json:"enabled" mapstructure:"enabled"`
}

// SourcesConfig contains source-related settings
type SourcesConfig struct {
	Enabled []string `json:"enabled" mapstructure:"enabled"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		DataDir:  "$HOME/.sdek",
		LogLevel: "info",
		Theme:    "dark",
		UserRole: RoleComplianceManager,
		Export: ExportConfig{
			DefaultPath: "$HOME/sdek/reports",
			Format:      "json",
		},
		Frameworks: FrameworksConfig{
			Enabled: []string{FrameworkSOC2, FrameworkISO27001, FrameworkPCIDSS},
		},
		Sources: SourcesConfig{
			Enabled: ValidSourceTypes,
		},
	}
}

// ValidateConfig checks if a Config meets all validation rules
func ValidateConfig(c *Config) error {
	if c == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	valid := false
	for _, level := range validLogLevels {
		if c.LogLevel == level {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid log level: %s, must be one of %v", c.LogLevel, validLogLevels)
	}

	// Validate theme
	validThemes := []string{"dark", "light"}
	valid = false
	for _, theme := range validThemes {
		if c.Theme == theme {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid theme: %s, must be one of %v", c.Theme, validThemes)
	}

	// Validate user role
	validRoles := []string{RoleComplianceManager, RoleEngineer}
	valid = false
	for _, role := range validRoles {
		if c.UserRole == role {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid user role: %s, must be one of %v", c.UserRole, validRoles)
	}

	// Validate export format
	validFormats := []string{"json"}
	valid = false
	for _, format := range validFormats {
		if c.Export.Format == format {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid export format: %s, must be one of %v", c.Export.Format, validFormats)
	}

	// Validate enabled frameworks
	for _, fw := range c.Frameworks.Enabled {
		valid = false
		for _, validFW := range ValidFrameworkIDs {
			if fw == validFW {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid framework: %s, must be one of %v", fw, ValidFrameworkIDs)
		}
	}

	// Validate enabled sources
	for _, src := range c.Sources.Enabled {
		valid = false
		for _, validSrc := range ValidSourceTypes {
			if src == validSrc {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid source: %s, must be one of %v", src, ValidSourceTypes)
		}
	}

	return nil
}
