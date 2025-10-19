package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func TestNewConfigLoader(t *testing.T) {
	loader := NewConfigLoader()
	if loader == nil {
		t.Fatal("NewConfigLoader returned nil")
	}
	if loader.v == nil {
		t.Error("ConfigLoader viper instance is nil")
	}
}

func TestLoadDefaults(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	loader := NewConfigLoader()
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config == nil {
		t.Fatal("Loaded config is nil")
	}

	// Check default values
	if config.LogLevel != "info" {
		t.Errorf("Expected log level 'info', got '%s'", config.LogLevel)
	}

	if config.Theme != "dark" {
		t.Errorf("Expected theme 'dark', got '%s'", config.Theme)
	}

	if config.UserRole != types.RoleComplianceManager {
		t.Errorf("Expected user role '%s', got '%s'", types.RoleComplianceManager, config.UserRole)
	}

	if config.Export.Format != "json" {
		t.Errorf("Expected export format 'json', got '%s'", config.Export.Format)
	}
}

func TestLoadFromConfigFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config directory
	configDir := filepath.Join(tmpDir, ".sdek")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Write a config file
	configContent := `
data_dir: /custom/data
log_level: debug
theme: light
user_role: engineer
export:
  default_path: /custom/exports
  format: yaml
sources:
  enabled:
    - git
    - jira
frameworks:
  enabled:
    - soc2
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	loader := NewConfigLoader()
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values from file
	if config.DataDir != "/custom/data" {
		t.Errorf("Expected data_dir '/custom/data', got '%s'", config.DataDir)
	}

	if config.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", config.LogLevel)
	}

	if config.Theme != "light" {
		t.Errorf("Expected theme 'light', got '%s'", config.Theme)
	}

	if config.UserRole != types.RoleEngineer {
		t.Errorf("Expected user role '%s', got '%s'", types.RoleEngineer, config.UserRole)
	}

	if config.Export.Format != "yaml" {
		t.Errorf("Expected export format 'yaml', got '%s'", config.Export.Format)
	}

	if len(config.Sources.Enabled) != 2 {
		t.Errorf("Expected 2 enabled sources, got %d", len(config.Sources.Enabled))
	}

	if len(config.Frameworks.Enabled) != 1 {
		t.Errorf("Expected 1 enabled framework, got %d", len(config.Frameworks.Enabled))
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Set environment variables
	os.Setenv("SDEK_LOG_LEVEL", "warn")
	os.Setenv("SDEK_THEME", "light")
	os.Setenv("SDEK_USER_ROLE", "engineer")
	defer func() {
		os.Unsetenv("SDEK_LOG_LEVEL")
		os.Unsetenv("SDEK_THEME")
		os.Unsetenv("SDEK_USER_ROLE")
	}()

	loader := NewConfigLoader()
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify environment variables override defaults
	if config.LogLevel != "warn" {
		t.Errorf("Expected log level 'warn' from env, got '%s'", config.LogLevel)
	}

	if config.Theme != "light" {
		t.Errorf("Expected theme 'light' from env, got '%s'", config.Theme)
	}

	if config.UserRole != types.RoleEngineer {
		t.Errorf("Expected user role '%s' from env, got '%s'", types.RoleEngineer, config.UserRole)
	}
}

func TestPrecedenceOrder(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config directory and file
	configDir := filepath.Join(tmpDir, ".sdek")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configContent := `
log_level: debug
theme: dark
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variable (should override config file)
	os.Setenv("SDEK_LOG_LEVEL", "error")
	defer os.Unsetenv("SDEK_LOG_LEVEL")

	loader := NewConfigLoader()

	// Set via code (should have highest priority)
	loader.Set("log_level", "warn")

	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify precedence: Set() > Env > Config File > Default
	if config.LogLevel != "warn" {
		t.Errorf("Expected log level 'warn' from Set(), got '%s'", config.LogLevel)
	}

	// Theme should come from config file (no env or Set override)
	if config.Theme != "dark" {
		t.Errorf("Expected theme 'dark' from config file, got '%s'", config.Theme)
	}
}

func TestWriteConfig(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	loader := NewConfigLoader()

	// Create a custom config
	config := &types.Config{
		DataDir:  "/test/data",
		LogLevel: "debug",
		Theme:    "light",
		UserRole: types.RoleEngineer,
		Export: types.ExportConfig{
			DefaultPath: "/test/exports",
			Format:      "yaml",
		},
		Sources: types.SourcesConfig{
			Enabled: []string{types.SourceTypeGit},
		},
		Frameworks: types.FrameworksConfig{
			Enabled: []string{types.FrameworkSOC2},
		},
	}

	// Write config
	if err := loader.WriteConfig(config); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Verify file exists
	configPath, err := loader.GetConfigFilePath()
	if err != nil {
		t.Fatalf("Failed to get config file path: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config again and verify
	newLoader := NewConfigLoader()
	loadedConfig, err := newLoader.Load()
	if err != nil {
		t.Fatalf("Failed to load written config: %v", err)
	}

	if loadedConfig.DataDir != "/test/data" {
		t.Errorf("Expected data_dir '/test/data', got '%s'", loadedConfig.DataDir)
	}

	if loadedConfig.LogLevel != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", loadedConfig.LogLevel)
	}
}

func TestGetConfigFilePath(t *testing.T) {
	loader := NewConfigLoader()
	path, err := loader.GetConfigFilePath()
	if err != nil {
		t.Fatalf("Failed to get config file path: %v", err)
	}

	if path == "" {
		t.Error("Config file path is empty")
	}

	if filepath.Base(path) != "config.yaml" {
		t.Errorf("Expected config.yaml, got %s", filepath.Base(path))
	}

	if filepath.Base(filepath.Dir(path)) != ".sdek" {
		t.Errorf("Expected .sdek directory, got %s", filepath.Base(filepath.Dir(path)))
	}
}

func TestGetSetMethods(t *testing.T) {
	loader := NewConfigLoader()

	// Test Set and Get
	loader.Set("test_key", "test_value")
	value := loader.Get("test_key")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%v'", value)
	}

	// Test GetString
	loader.Set("string_key", "string_value")
	strValue := loader.GetString("string_key")
	if strValue != "string_value" {
		t.Errorf("Expected 'string_value', got '%s'", strValue)
	}

	// Test GetBool
	loader.Set("bool_key", true)
	boolValue := loader.GetBool("bool_key")
	if !boolValue {
		t.Error("Expected true, got false")
	}

	// Test GetInt
	loader.Set("int_key", 42)
	intValue := loader.GetInt("int_key")
	if intValue != 42 {
		t.Errorf("Expected 42, got %d", intValue)
	}
}
