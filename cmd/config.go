package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage sdek configuration",
	Long: `Manage sdek configuration files and settings.

The config command provides subcommands to:
- Initialize a new configuration file
- Get configuration values
- Set configuration values
- List all configuration values
- Validate the configuration

Configuration precedence (highest to lowest):
1. Command-line flags
2. Environment variables (SDEK_*)
3. Configuration file (~/.sdek/config.yaml)
4. Default values`,
	Example: `  # Initialize a new config file
  sdek config init

  # Get a configuration value
  sdek config get export.enabled

  # Set a configuration value
  sdek config set export.enabled true

  # List all configuration values
  sdek config list

  # Validate configuration
  sdek config validate`,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configValidateCmd)
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new configuration file",
	Long:  `Create a new configuration file with default values at ~/.sdek/config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		configPath := filepath.Join(homeDir, ".sdek", "config.yaml")

		// Check if config already exists
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("configuration file already exists at %s", configPath)
		}

		// Create .sdek directory if it doesn't exist
		sdekDir := filepath.Join(homeDir, ".sdek")
		if err := os.MkdirAll(sdekDir, 0755); err != nil {
			return fmt.Errorf("failed to create .sdek directory: %w", err)
		}

		// Default configuration
		defaultConfig := `# sdek Configuration File
# Documentation: https://github.com/yourorg/sdek-cli

# Logging configuration
log-level: info
verbose: false

# Data directory
data-dir: ~/.sdek

# Export settings
export:
  enabled: true
  format: json
  path: ~/.sdek/reports

# Enabled frameworks
frameworks:
  soc2: true
  iso27001: true
  pci_dss: true

# Enabled sources
sources:
  git: true
  jira: true
  slack: true
  cicd: true
  docs: true

# UI settings
ui:
  theme: dark
  refresh-interval: 5s
`

		// Write default configuration
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			return fmt.Errorf("failed to write configuration file: %w", err)
		}

		fmt.Printf("✓ Configuration file created at: %s\n", configPath)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long:  `Get the value of a specific configuration key`,
	Example: `  sdek config get log-level
  sdek config get export.enabled
  sdek config get frameworks.soc2`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		
		if !viper.IsSet(key) {
			return fmt.Errorf("configuration key '%s' not found", key)
		}

		value := viper.Get(key)
		fmt.Printf("%s = %v\n", key, value)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  `Set a configuration value and save it to the config file`,
	Example: `  sdek config set log-level debug
  sdek config set export.enabled true
  sdek config set frameworks.soc2 false`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		viper.Set(key, value)

		// Write configuration
		if err := viper.WriteConfig(); err != nil {
			// If config doesn't exist, create it
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				homeDir, _ := os.UserHomeDir()
				configPath := filepath.Join(homeDir, ".sdek", "config.yaml")
				
				// Create directory
				if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
					return fmt.Errorf("failed to create config directory: %w", err)
				}
				
				// Set config file
				viper.SetConfigFile(configPath)
				
				// Write config
				if err := viper.WriteConfig(); err != nil {
					return fmt.Errorf("failed to write config: %w", err)
				}
			} else {
				return fmt.Errorf("failed to write configuration: %w", err)
			}
		}

		fmt.Printf("✓ Set %s = %s\n", key, value)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  `Display all configuration keys and their current values`,
	RunE: func(cmd *cobra.Command, args []string) error {
		settings := viper.AllSettings()
		
		if len(settings) == 0 {
			fmt.Println("No configuration values found")
			return nil
		}

		fmt.Println("Current configuration:")
		fmt.Println()
		
		printSettings(settings, "")
		
		return nil
	},
}

func printSettings(settings map[string]interface{}, prefix string) {
	for key, value := range settings {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		
		switch v := value.(type) {
		case map[string]interface{}:
			printSettings(v, fullKey)
		default:
			fmt.Printf("  %-30s %v\n", fullKey, v)
		}
	}
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  `Check if the current configuration is valid`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check log level
		logLevel := viper.GetString("log-level")
		validLevels := []string{"debug", "info", "warn", "error"}
		valid := false
		for _, level := range validLevels {
			if logLevel == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log-level '%s', must be one of: debug, info, warn, error", logLevel)
		}

		// Check data directory exists or can be created
		dataDir := viper.GetString("data-dir")
		if dataDir != "" {
			// Expand home directory
			if dataDir[:2] == "~/" {
				homeDir, _ := os.UserHomeDir()
				dataDir = filepath.Join(homeDir, dataDir[2:])
			}
			
			// Check if directory exists or can be created
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				return fmt.Errorf("invalid data-dir '%s': %w", dataDir, err)
			}
		}

		fmt.Println("✓ Configuration is valid")
		return nil
	},
}
