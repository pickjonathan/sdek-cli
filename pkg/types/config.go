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
	AI         AIConfig         `json:"ai" mapstructure:"ai"`
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

// AIConfig contains AI analysis settings (Feature 002: AI Evidence Analysis)
type AIConfig struct {
	Enabled      bool   `json:"enabled" mapstructure:"enabled"`
	Provider     string `json:"provider" mapstructure:"provider"`
	Model        string `json:"model" mapstructure:"model"`
	OpenAIKey    string `json:"openai_key" mapstructure:"openai_key"`
	AnthropicKey string `json:"anthropic_key" mapstructure:"anthropic_key"`
	Timeout      int    `json:"timeout" mapstructure:"timeout"`       // seconds
	RateLimit    int    `json:"rate_limit" mapstructure:"rate_limit"` // requests per minute
	CacheDir     string `json:"cache_dir" mapstructure:"cache_dir"`   // cache directory path
}

// AI provider constants
const (
	AIProviderOpenAI    = "openai"
	AIProviderAnthropic = "anthropic"
)

// ValidAIProviders is the list of valid AI providers
var ValidAIProviders = []string{AIProviderOpenAI, AIProviderAnthropic}

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
		AI: AIConfig{
			Enabled:   false, // Disabled by default (opt-in)
			Provider:  AIProviderOpenAI,
			Model:     "gpt-4", // Default OpenAI model
			Timeout:   60,      // 60 seconds
			RateLimit: 10,      // 10 requests per minute
			CacheDir:  "$HOME/.sdek/cache/ai",
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

	// Validate AI config
	if c.AI.Enabled {
		// Validate provider
		valid = false
		for _, provider := range ValidAIProviders {
			if c.AI.Provider == provider {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid AI provider: %s, must be one of %v", c.AI.Provider, ValidAIProviders)
		}

		// Validate model is not empty
		if c.AI.Model == "" {
			return fmt.Errorf("AI model cannot be empty when AI is enabled")
		}

		// Validate timeout
		if c.AI.Timeout <= 0 {
			return fmt.Errorf("AI timeout must be positive, got %d", c.AI.Timeout)
		}

		// Validate rate limit
		if c.AI.RateLimit < 0 {
			return fmt.Errorf("AI rate limit cannot be negative, got %d", c.AI.RateLimit)
		}

		// Validate API keys
		if c.AI.Provider == AIProviderOpenAI && c.AI.OpenAIKey == "" {
			return fmt.Errorf("OpenAI API key required when provider is openai")
		}
		if c.AI.Provider == AIProviderAnthropic && c.AI.AnthropicKey == "" {
			return fmt.Errorf("Anthropic API key required when provider is anthropic")
		}
	}

	return nil
}
