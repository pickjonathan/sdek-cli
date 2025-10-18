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

// AIConfig contains AI analysis settings (Feature 002 + 003: AI Evidence Analysis + Context Injection)
type AIConfig struct {
	Enabled      bool              `json:"enabled" mapstructure:"enabled"`
	Provider     string            `json:"provider" mapstructure:"provider"` // anthropic|openai
	Model        string            `json:"model" mapstructure:"model"`
	OpenAIKey    string            `json:"openai_key" mapstructure:"openai_key"`
	AnthropicKey string            `json:"anthropic_key" mapstructure:"anthropic_key"`
	APIKey       string            `json:"apiKey" mapstructure:"apiKey"`           // Feature 003: Unified API key field
	Mode         string            `json:"mode" mapstructure:"mode"`               // Feature 003: disabled|context|autonomous
	Timeout      int               `json:"timeout" mapstructure:"timeout"`         // seconds
	RateLimit    int               `json:"rate_limit" mapstructure:"rate_limit"`   // requests per minute
	CacheDir     string            `json:"cache_dir" mapstructure:"cache_dir"`     // cache directory path
	NoCache      bool              `json:"no_cache" mapstructure:"no_cache"`       // Feature 003: Disable caching
	Concurrency  ConcurrencyLimits `json:"concurrency" mapstructure:"concurrency"` // Feature 003: Concurrency limits
	Budgets      BudgetLimits      `json:"budgets" mapstructure:"budgets"`         // Feature 003: Budget limits
	Autonomous   AutonomousConfig  `json:"autonomous" mapstructure:"autonomous"`   // Feature 003: Autonomous mode config
	Redaction    RedactionConfig   `json:"redaction" mapstructure:"redaction"`     // Feature 003: Redaction settings
}

// ConcurrencyLimits defines concurrency constraints for AI operations (Feature 003)
type ConcurrencyLimits struct {
	MaxAnalyses int `json:"maxAnalyses" mapstructure:"maxAnalyses"` // Default: 25
}

// BudgetLimits defines resource constraints for AI operations (Feature 003)
type BudgetLimits struct {
	MaxSources  int `json:"maxSources" mapstructure:"maxSources"`   // Default: 50
	MaxAPICalls int `json:"maxAPICalls" mapstructure:"maxAPICalls"` // Default: 500
	MaxTokens   int `json:"maxTokens" mapstructure:"maxTokens"`     // Default: 250000
}

// AutonomousConfig defines autonomous evidence collection settings (Feature 003)
type AutonomousConfig struct {
	Enabled     bool              `json:"enabled" mapstructure:"enabled"`
	AutoApprove AutoApproveConfig `json:"autoApprove" mapstructure:"autoApprove"`
}

// AutoApproveConfig defines auto-approval policy for evidence plans (Feature 003)
// It's a map of source name to list of glob patterns
type AutoApproveConfig map[string][]string // source -> patterns

// RedactionConfig defines PII/secret redaction settings (Feature 003)
type RedactionConfig struct {
	Enabled  bool     `json:"enabled" mapstructure:"enabled"`   // Default: true
	Denylist []string `json:"denylist" mapstructure:"denylist"` // Exact match strings
}

// AI provider constants
const (
	AIProviderOpenAI    = "openai"
	AIProviderAnthropic = "anthropic"
)

// ValidAIProviders is the list of valid AI providers
var ValidAIProviders = []string{AIProviderOpenAI, AIProviderAnthropic}

// AI mode constants (Feature 003)
const (
	AIModeDisabled   = "disabled"
	AIModeContext    = "context"
	AIModeAutonomous = "autonomous"
)

// ValidAIModes is the list of valid AI modes
var ValidAIModes = []string{AIModeDisabled, AIModeContext, AIModeAutonomous}

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
			Mode:      AIModeDisabled,
			Timeout:   60, // 60 seconds
			RateLimit: 10, // 10 requests per minute
			CacheDir:  "$HOME/.sdek/cache/ai",
			Concurrency: ConcurrencyLimits{
				MaxAnalyses: 25,
			},
			Budgets: BudgetLimits{
				MaxSources:  50,
				MaxAPICalls: 500,
				MaxTokens:   250000,
			},
			Autonomous: AutonomousConfig{
				Enabled:     false,
				AutoApprove: make(AutoApproveConfig),
			},
			Redaction: RedactionConfig{
				Enabled:  true,
				Denylist: []string{},
			},
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

		// Validate mode (Feature 003)
		valid = false
		for _, mode := range ValidAIModes {
			if c.AI.Mode == mode {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid AI mode: %s, must be one of %v", c.AI.Mode, ValidAIModes)
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
		if c.AI.Provider == AIProviderOpenAI && c.AI.OpenAIKey == "" && c.AI.APIKey == "" {
			return fmt.Errorf("OpenAI API key required when provider is openai")
		}
		if c.AI.Provider == AIProviderAnthropic && c.AI.AnthropicKey == "" && c.AI.APIKey == "" {
			return fmt.Errorf("Anthropic API key required when provider is anthropic")
		}

		// Validate concurrency limits (Feature 003)
		if c.AI.Concurrency.MaxAnalyses <= 0 {
			return fmt.Errorf("AI concurrency.maxAnalyses must be positive, got %d", c.AI.Concurrency.MaxAnalyses)
		}

		// Validate budget limits (Feature 003)
		if c.AI.Budgets.MaxSources <= 0 {
			return fmt.Errorf("AI budgets.maxSources must be positive, got %d", c.AI.Budgets.MaxSources)
		}
		if c.AI.Budgets.MaxAPICalls <= 0 {
			return fmt.Errorf("AI budgets.maxAPICalls must be positive, got %d", c.AI.Budgets.MaxAPICalls)
		}
		if c.AI.Budgets.MaxTokens <= 0 {
			return fmt.Errorf("AI budgets.maxTokens must be positive, got %d", c.AI.Budgets.MaxTokens)
		}
	}

	return nil
}
