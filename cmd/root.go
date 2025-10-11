package cmd
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	dataDir  string
	logLevel string
	verbose  bool
	version  = "dev"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sdek",
	Short: "sdek - Compliance Evidence Mapping Tool",
	Long: `sdek is a CLI and terminal UI tool that reduces audit preparation time
by automating compliance evidence mapping.

It ingests data from multiple sources (Git, Jira, Slack, CI/CD, Docs),
maps them to compliance frameworks (SOC2, ISO 27001, PCI DSS), and
provides interactive visualization with export capabilities.`,
	Example: `  # Start the terminal UI
  sdek tui

  # Generate demo data
  sdek seed --demo

  # Ingest from specific source
  sdek ingest --source git --events 50

  # Analyze evidence and calculate risk scores
  sdek analyze

  # Export compliance report
  sdek report --output ~/reports/compliance.json

  # Manage configuration
  sdek config get export.enabled`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration
		if err := initConfig(); err != nil {
			return fmt.Errorf("failed to initialize configuration: %w", err)
		}

		// Initialize logging
		if err := initLogging(); err != nil {
			return fmt.Errorf("failed to initialize logging: %w", err)
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sdek/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", "", "data directory (default is $HOME/.sdek)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("sdek version %s\n", version)
		},
	})

	// Bind flags to viper
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("data-dir", rootCmd.PersistentFlags().Lookup("data-dir"))
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		// Search config in home directory with name ".sdek/config" (without extension)
		sdekDir := filepath.Join(home, ".sdek")
		viper.AddConfigPath(sdekDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		// Create .sdek directory if it doesn't exist
		if err := os.MkdirAll(sdekDir, 0755); err != nil {
			return fmt.Errorf("failed to create .sdek directory: %w", err)
		}
	}

	// Environment variables
	viper.SetEnvPrefix("SDEK")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}

	// Set data directory default
	if dataDir == "" && !viper.IsSet("data-dir") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		dataDir = filepath.Join(home, ".sdek")
		viper.Set("data-dir", dataDir)
	}

	return nil
}

// initLogging sets up structured logging based on configuration
func initLogging() error {
	// Determine log level
	level := logLevel
	if viper.IsSet("log-level") {
		level = viper.GetString("log-level")
	}
	if verbose {
		level = "debug"
	}

	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", level)
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	// Create JSON handler for structured logging to stderr
	handler := slog.NewJSONHandler(os.Stderr, opts)
	logger := slog.New(handler)

	// Set as default logger
	slog.SetDefault(logger)

	if verbose {
		slog.Debug("Logging initialized", "level", level)
	}

	return nil
}

// GetVersion returns the current version
func GetVersion() string {
	return version
}

// SetVersion sets the version (called from main for ldflags injection)
func SetVersion(v string) {
	version = v
}
