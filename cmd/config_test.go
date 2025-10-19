package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestConfigCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "help flag",
			args:        []string{"config", "--help"},
			expectError: false,
		},
		{
			name:        "config without subcommand",
			args:        []string{"config"},
			expectError: false, // Shows help
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
		})
	}
}

func TestConfigInitCommand(t *testing.T) {
	// Create temporary directory for test config
	tmpDir := t.TempDir()
	
	// Set HOME to temp directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Run config init command
	rootCmd.SetArgs([]string{"config", "init"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("config init command failed: %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(tmpDir, ".sdek", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("expected config file to be created at %s", configPath)
	}

	// Verify config file is not empty
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("expected config file to have content, got 0 bytes")
	}
}

func TestConfigGetCommand(t *testing.T) {
	// Create temporary directory for test config
	tmpDir := t.TempDir()
	
	// Set HOME to temp directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// First, init config
	rootCmd.SetArgs([]string{"config", "init"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	_ = rootCmd.Execute()

	// Reset viper to reload config
	viper.Reset()

	tests := []struct {
		name        string
		key         string
		expectError bool
	}{
		{
			name:        "get log level",
			key:         "log.level",
			expectError: false,
		},
		{
			name:        "get export enabled",
			key:         "export.enabled",
			expectError: false,
		},
		{
			name:        "get data dir",
			key:         "data.dir",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd.SetArgs([]string{"config", "get", tt.key})
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			err := rootCmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify output contains the key
			output := buf.String()
			if !tt.expectError && !bytes.Contains([]byte(output), []byte(tt.key)) {
				t.Errorf("expected output to contain key %q, but it didn't", tt.key)
			}
		})
	}
}

func TestConfigSetCommand(t *testing.T) {
	// Create temporary directory for test config
	tmpDir := t.TempDir()
	
	// Set HOME to temp directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// First, init config
	rootCmd.SetArgs([]string{"config", "init"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	_ = rootCmd.Execute()

	// Reset viper to reload config
	viper.Reset()

	tests := []struct {
		name        string
		key         string
		value       string
		expectError bool
	}{
		{
			name:        "set log level",
			key:         "log.level",
			value:       "debug",
			expectError: false,
		},
		{
			name:        "set export enabled",
			key:         "export.enabled",
			value:       "true",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd.SetArgs([]string{"config", "set", tt.key, tt.value})
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			err := rootCmd.Execute()

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConfigListCommand(t *testing.T) {
	// Create temporary directory for test config
	tmpDir := t.TempDir()
	
	// Set HOME to temp directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// First, init config
	rootCmd.SetArgs([]string{"config", "init"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	_ = rootCmd.Execute()

	// Reset viper to reload config
	viper.Reset()

	// Run config list command
	rootCmd.SetArgs([]string{"config", "list"})
	buf = new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("config list command failed: %v", err)
	}

	output := buf.String()

	// Verify output contains expected configuration keys
	expectedKeys := []string{
		"log",
		"export",
		"data",
	}

	for _, key := range expectedKeys {
		if !bytes.Contains([]byte(output), []byte(key)) {
			t.Errorf("expected output to contain key %q, but it didn't", key)
		}
	}
}

func TestConfigValidateCommand(t *testing.T) {
	// Create temporary directory for test config
	tmpDir := t.TempDir()
	
	// Set HOME to temp directory for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// First, init config
	rootCmd.SetArgs([]string{"config", "init"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	_ = rootCmd.Execute()

	// Reset viper to reload config
	viper.Reset()

	// Run config validate command
	rootCmd.SetArgs([]string{"config", "validate"})
	buf = new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("config validate command failed: %v", err)
	}

	output := buf.String()

	// Verify output indicates validation success
	if !bytes.Contains([]byte(output), []byte("valid")) {
		t.Errorf("expected output to indicate validation success, got: %s", output)
	}
}
