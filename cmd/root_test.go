package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedErrMsg string
		expectError    bool
	}{
		{
			name:        "help flag",
			args:        []string{"--help"},
			expectError: false,
		},
		{
			name:        "version command",
			args:        []string{"version"},
			expectError: false,
		},
		{
			name:        "no arguments shows help",
			args:        []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset root command
			rootCmd.SetArgs(tt.args)

			// Capture output
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			// Execute command
			err := rootCmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectedErrMsg != "" && (err == nil || err.Error() != tt.expectedErrMsg) {
				t.Errorf("expected error message %q, got %q", tt.expectedErrMsg, err)
			}
		})
	}
}

func TestGlobalFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		checkFn func(*testing.T)
	}{
		{
			name: "config flag",
			args: []string{"--config", "/custom/config.yaml", "version"},
			checkFn: func(t *testing.T) {
				if cfgFile != "/custom/config.yaml" {
					t.Errorf("expected cfgFile to be /custom/config.yaml, got %s", cfgFile)
				}
			},
		},
		{
			name: "data-dir flag",
			args: []string{"--data-dir", "/custom/data", "version"},
			checkFn: func(t *testing.T) {
				if dataDir != "/custom/data" {
					t.Errorf("expected dataDir to be /custom/data, got %s", dataDir)
				}
			},
		},
		{
			name: "log-level flag",
			args: []string{"--log-level", "debug", "version"},
			checkFn: func(t *testing.T) {
				if logLevel != "debug" {
					t.Errorf("expected logLevel to be debug, got %s", logLevel)
				}
			},
		},
		{
			name: "verbose flag",
			args: []string{"--verbose", "version"},
			checkFn: func(t *testing.T) {
				if !verbose {
					t.Errorf("expected verbose to be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			cfgFile = ""
			dataDir = ""
			logLevel = "info"
			verbose = false

			rootCmd.SetArgs(tt.args)
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			_ = rootCmd.Execute()

			if tt.checkFn != nil {
				tt.checkFn(t)
			}
		})
	}
}

func TestInitConfig(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Set config file path
	cfgFile = filepath.Join(tmpDir, "config.yaml")

	// Create a test config file
	configContent := `export:
  enabled: true
  path: /tmp/reports
log:
  level: debug`

	err := os.WriteFile(cfgFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test config file: %v", err)
	}

	// Initialize config
	err = initConfig()
	if err != nil {
		t.Fatalf("initConfig failed: %v", err)
	}

	// Verify config values
	if !viper.GetBool("export.enabled") {
		t.Errorf("expected export.enabled to be true")
	}
	if viper.GetString("export.path") != "/tmp/reports" {
		t.Errorf("expected export.path to be /tmp/reports, got %s", viper.GetString("export.path"))
	}
	if viper.GetString("log.level") != "debug" {
		t.Errorf("expected log.level to be debug, got %s", viper.GetString("log.level"))
	}
}

func TestInitLogging(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		verbose   bool
		expectErr bool
	}{
		{
			name:      "valid debug level",
			logLevel:  "debug",
			verbose:   false,
			expectErr: false,
		},
		{
			name:      "valid info level",
			logLevel:  "info",
			verbose:   false,
			expectErr: false,
		},
		{
			name:      "valid warn level",
			logLevel:  "warn",
			verbose:   false,
			expectErr: false,
		},
		{
			name:      "valid error level",
			logLevel:  "error",
			verbose:   false,
			expectErr: false,
		},
		{
			name:      "invalid level",
			logLevel:  "invalid",
			verbose:   false,
			expectErr: true,
		},
		{
			name:      "verbose flag overrides to debug",
			logLevel:  "info",
			verbose:   true,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logLevel = tt.logLevel
			verbose = tt.verbose

			err := initLogging()

			if tt.expectErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	// Test default version
	if GetVersion() == "" {
		t.Errorf("GetVersion should return non-empty string")
	}

	// Test setting version
	SetVersion("1.2.3")
	if GetVersion() != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %s", GetVersion())
	}
}

func TestSetVersion(t *testing.T) {
	tests := []string{
		"1.0.0",
		"v2.3.4",
		"dev",
		"0.0.1-alpha",
	}

	for _, v := range tests {
		SetVersion(v)
		if version != v {
			t.Errorf("expected version %s, got %s", v, version)
		}
	}
}
