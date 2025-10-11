package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/spf13/viper"
)

// ConfigLoader handles loading configuration from multiple sources
type ConfigLoader struct {
	v *viper.Viper
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{
		v: viper.New(),
	}
}

// Load loads configuration with the following precedence:
// 1. Command-line flags (highest priority)
// 2. Environment variables (SDEK_*)
// 3. Config file ($HOME/.sdek/config.yaml)
// 4. Default values (lowest priority)
func (cl *ConfigLoader) Load() (*types.Config, error) {
	// Set default values
	cl.setDefaults()

	// Configure environment variable binding
	cl.v.SetEnvPrefix("SDEK")
	cl.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cl.v.AutomaticEnv()

	// Set config file location
	if err := cl.configureConfigFile(); err != nil {
		return nil, fmt.Errorf("failed to configure config file: %w", err)
	}

	// Try to read config file (it's okay if it doesn't exist)
	if err := cl.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found; using defaults and env vars
	}

	// Unmarshal into Config struct
	config := &types.Config{}
	if err := cl.v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// setDefaults sets default configuration values
func (cl *ConfigLoader) setDefaults() {
	cl.v.SetDefault("data_dir", "$HOME/.sdek")
	cl.v.SetDefault("log_level", "info")
	cl.v.SetDefault("theme", "dark")
	cl.v.SetDefault("user_role", types.RoleComplianceManager)

	// Export defaults
	cl.v.SetDefault("export.default_path", "$HOME/sdek/reports")
	cl.v.SetDefault("export.format", "json")

	// Sources defaults (all enabled by default)
	cl.v.SetDefault("sources.enabled", types.ValidSourceTypes)

	// Frameworks defaults (all enabled by default)
	cl.v.SetDefault("frameworks.enabled", []string{
		types.FrameworkSOC2,
		types.FrameworkISO27001,
		types.FrameworkPCIDSS,
	})
}

// configureConfigFile sets up the config file path
func (cl *ConfigLoader) configureConfigFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".sdek")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	cl.v.SetConfigName("config")
	cl.v.SetConfigType("yaml")
	cl.v.AddConfigPath(configDir)

	return nil
}

// GetConfigFilePath returns the path to the config file
func (cl *ConfigLoader) GetConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".sdek", "config.yaml"), nil
}

// Set allows setting configuration values (typically from command-line flags)
func (cl *ConfigLoader) Set(key string, value interface{}) {
	cl.v.Set(key, value)
}

// Get retrieves a configuration value
func (cl *ConfigLoader) Get(key string) interface{} {
	return cl.v.Get(key)
}

// GetString retrieves a string configuration value
func (cl *ConfigLoader) GetString(key string) string {
	return cl.v.GetString(key)
}

// GetBool retrieves a boolean configuration value
func (cl *ConfigLoader) GetBool(key string) bool {
	return cl.v.GetBool(key)
}

// GetInt retrieves an integer configuration value
func (cl *ConfigLoader) GetInt(key string) int {
	return cl.v.GetInt(key)
}

// WriteConfig writes the current configuration to the config file
func (cl *ConfigLoader) WriteConfig(config *types.Config) error {
	// Set all config values
	cl.v.Set("data_dir", config.DataDir)
	cl.v.Set("log_level", config.LogLevel)
	cl.v.Set("theme", config.Theme)
	cl.v.Set("user_role", config.UserRole)

	cl.v.Set("export.default_path", config.Export.DefaultPath)
	cl.v.Set("export.format", config.Export.Format)

	cl.v.Set("sources.enabled", config.Sources.Enabled)

	cl.v.Set("frameworks.enabled", config.Frameworks.Enabled)

	// Ensure config directory exists
	if err := cl.configureConfigFile(); err != nil {
		return fmt.Errorf("failed to configure config file: %w", err)
	}

	// Write to file
	configPath, err := cl.GetConfigFilePath()
	if err != nil {
		return fmt.Errorf("failed to get config file path: %w", err)
	}

	// Ensure the file exists or can be created
	if err := cl.v.WriteConfigAs(configPath); err != nil {
		// If the file doesn't exist, SafeWriteConfigAs will create it
		if os.IsNotExist(err) {
			if err := cl.v.SafeWriteConfigAs(configPath); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
		} else {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}

	return nil
}
