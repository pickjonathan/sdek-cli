package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// TestLoadAIConfigFeature003_Defaults tests Feature 003 AI config defaults
func TestLoadAIConfigFeature003_Defaults(t *testing.T) {
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

	// Verify Feature 003 defaults
	if config.AI.Mode != types.AIModeDisabled {
		t.Errorf("Expected AI mode '%s', got '%s'", types.AIModeDisabled, config.AI.Mode)
	}

	if config.AI.Concurrency.MaxAnalyses != 25 {
		t.Errorf("Expected maxAnalyses 25, got %d", config.AI.Concurrency.MaxAnalyses)
	}

	if config.AI.Budgets.MaxSources != 50 {
		t.Errorf("Expected maxSources 50, got %d", config.AI.Budgets.MaxSources)
	}

	if config.AI.Budgets.MaxAPICalls != 500 {
		t.Errorf("Expected maxAPICalls 500, got %d", config.AI.Budgets.MaxAPICalls)
	}

	if config.AI.Budgets.MaxTokens != 250000 {
		t.Errorf("Expected maxTokens 250000, got %d", config.AI.Budgets.MaxTokens)
	}

	if config.AI.Autonomous.Enabled {
		t.Error("Expected autonomous.enabled false by default")
	}

	if config.AI.Autonomous.AutoApprove.Enabled {
		t.Error("Expected autoApprove.enabled false by default")
	}

	if !config.AI.Redaction.Enabled {
		t.Error("Expected redaction.enabled true by default")
	}

	if len(config.AI.Redaction.Denylist) != 0 {
		t.Errorf("Expected empty denylist by default, got %d items", len(config.AI.Redaction.Denylist))
	}
}

// TestLoadAIConfigFeature003_FromFile tests Feature 003 AI config from YAML file
func TestLoadAIConfigFeature003_FromFile(t *testing.T) {
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

	// Write a config file with Feature 003 settings
	configContent := `
ai:
  enabled: true
  provider: anthropic
  model: claude-3-opus
  mode: context
  apiKey: test-key-123
  
  concurrency:
    maxAnalyses: 50
  
  budgets:
    maxSources: 100
    maxAPICalls: 1000
    maxTokens: 500000
  
  autonomous:
    enabled: true
    autoApprove:
      enabled: true
      rules:
        github: ["auth*", "*login*", "mfa*"]
        aws: ["iam*", "security*"]
        jira: ["INFOSEC-*"]
  
  redaction:
    enabled: true
    denylist:
      - "password123"
      - "secret-token"
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

	// Verify Feature 003 values from file
	if config.AI.Mode != types.AIModeContext {
		t.Errorf("Expected AI mode '%s', got '%s'", types.AIModeContext, config.AI.Mode)
	}

	if config.AI.APIKey != "test-key-123" {
		t.Errorf("Expected apiKey 'test-key-123', got '%s'", config.AI.APIKey)
	}

	if config.AI.Concurrency.MaxAnalyses != 50 {
		t.Errorf("Expected maxAnalyses 50, got %d", config.AI.Concurrency.MaxAnalyses)
	}

	if config.AI.Budgets.MaxSources != 100 {
		t.Errorf("Expected maxSources 100, got %d", config.AI.Budgets.MaxSources)
	}

	if config.AI.Budgets.MaxAPICalls != 1000 {
		t.Errorf("Expected maxAPICalls 1000, got %d", config.AI.Budgets.MaxAPICalls)
	}

	if config.AI.Budgets.MaxTokens != 500000 {
		t.Errorf("Expected maxTokens 500000, got %d", config.AI.Budgets.MaxTokens)
	}

	if !config.AI.Autonomous.Enabled {
		t.Error("Expected autonomous.enabled true")
	}

	if !config.AI.Autonomous.AutoApprove.Enabled {
		t.Error("Expected autoApprove.enabled true")
	}

	// Verify auto-approve rules
	if len(config.AI.Autonomous.AutoApprove.Rules) != 3 {
		t.Errorf("Expected 3 auto-approve rule sets, got %d", len(config.AI.Autonomous.AutoApprove.Rules))
	}

	githubRules, ok := config.AI.Autonomous.AutoApprove.Rules["github"]
	if !ok {
		t.Error("Expected github rules to exist")
	} else if len(githubRules) != 3 {
		t.Errorf("Expected 3 github rules, got %d", len(githubRules))
	}

	awsRules, ok := config.AI.Autonomous.AutoApprove.Rules["aws"]
	if !ok {
		t.Error("Expected aws rules to exist")
	} else if len(awsRules) != 2 {
		t.Errorf("Expected 2 aws rules, got %d", len(awsRules))
	}

	jiraRules, ok := config.AI.Autonomous.AutoApprove.Rules["jira"]
	if !ok {
		t.Error("Expected jira rules to exist")
	} else if len(jiraRules) != 1 {
		t.Errorf("Expected 1 jira rule, got %d", len(jiraRules))
	}

	// Verify redaction
	if !config.AI.Redaction.Enabled {
		t.Error("Expected redaction.enabled true")
	}

	if len(config.AI.Redaction.Denylist) != 2 {
		t.Errorf("Expected 2 denylist items, got %d", len(config.AI.Redaction.Denylist))
	}

	if config.AI.Redaction.Denylist[0] != "password123" {
		t.Errorf("Expected first denylist item 'password123', got '%s'", config.AI.Redaction.Denylist[0])
	}

	if config.AI.Redaction.Denylist[1] != "secret-token" {
		t.Errorf("Expected second denylist item 'secret-token', got '%s'", config.AI.Redaction.Denylist[1])
	}
}

// TestLoadAIConfigFeature003_FromEnvironment tests Feature 003 AI config from environment variables
func TestLoadAIConfigFeature003_FromEnvironment(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Set environment variables for Feature 003
	os.Setenv("SDEK_AI_MODE", types.AIModeAutonomous)
	os.Setenv("SDEK_AI_CONCURRENCY_MAXANALYSES", "75")
	os.Setenv("SDEK_AI_BUDGETS_MAXSOURCES", "200")
	os.Setenv("SDEK_AI_BUDGETS_MAXAPICALLS", "2000")
	os.Setenv("SDEK_AI_BUDGETS_MAXTOKENS", "1000000")
	os.Setenv("SDEK_AI_AUTONOMOUS_ENABLED", "true")
	os.Setenv("SDEK_AI_REDACTION_ENABLED", "false")

	defer func() {
		os.Unsetenv("SDEK_AI_MODE")
		os.Unsetenv("SDEK_AI_CONCURRENCY_MAXANALYSES")
		os.Unsetenv("SDEK_AI_BUDGETS_MAXSOURCES")
		os.Unsetenv("SDEK_AI_BUDGETS_MAXAPICALLS")
		os.Unsetenv("SDEK_AI_BUDGETS_MAXTOKENS")
		os.Unsetenv("SDEK_AI_AUTONOMOUS_ENABLED")
		os.Unsetenv("SDEK_AI_REDACTION_ENABLED")
	}()

	loader := NewConfigLoader()
	config, err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify environment variables override defaults
	if config.AI.Mode != types.AIModeAutonomous {
		t.Errorf("Expected AI mode '%s' from env, got '%s'", types.AIModeAutonomous, config.AI.Mode)
	}

	if config.AI.Concurrency.MaxAnalyses != 75 {
		t.Errorf("Expected maxAnalyses 75 from env, got %d", config.AI.Concurrency.MaxAnalyses)
	}

	if config.AI.Budgets.MaxSources != 200 {
		t.Errorf("Expected maxSources 200 from env, got %d", config.AI.Budgets.MaxSources)
	}

	if config.AI.Budgets.MaxAPICalls != 2000 {
		t.Errorf("Expected maxAPICalls 2000 from env, got %d", config.AI.Budgets.MaxAPICalls)
	}

	if config.AI.Budgets.MaxTokens != 1000000 {
		t.Errorf("Expected maxTokens 1000000 from env, got %d", config.AI.Budgets.MaxTokens)
	}

	if !config.AI.Autonomous.Enabled {
		t.Error("Expected autonomous.enabled true from env")
	}

	if config.AI.Redaction.Enabled {
		t.Error("Expected redaction.enabled false from env")
	}
}

// TestWriteAIConfigFeature003 tests writing Feature 003 AI config to file
func TestWriteAIConfigFeature003(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	loader := NewConfigLoader()

	// Create a config with Feature 003 settings
	config := types.DefaultConfig()
	config.AI.Enabled = true
	config.AI.Mode = types.AIModeContext
	config.AI.Provider = types.AIProviderAnthropic
	config.AI.APIKey = "test-api-key"
	config.AI.Concurrency.MaxAnalyses = 100
	config.AI.Budgets.MaxSources = 150
	config.AI.Budgets.MaxAPICalls = 3000
	config.AI.Budgets.MaxTokens = 750000
	config.AI.Autonomous.Enabled = true
	config.AI.Autonomous.AutoApprove.Enabled = true
	config.AI.Autonomous.AutoApprove.Rules = map[string][]string{
		"github": {"auth*", "security*"},
		"aws":    {"iam*"},
	}
	config.AI.Redaction.Enabled = true
	config.AI.Redaction.Denylist = []string{"test-secret", "test-password"}

	// Write config
	if err := loader.WriteConfig(config); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config again and verify Feature 003 fields
	newLoader := NewConfigLoader()
	loadedConfig, err := newLoader.Load()
	if err != nil {
		t.Fatalf("Failed to load written config: %v", err)
	}

	if loadedConfig.AI.Mode != types.AIModeContext {
		t.Errorf("Expected AI mode '%s', got '%s'", types.AIModeContext, loadedConfig.AI.Mode)
	}

	if loadedConfig.AI.APIKey != "test-api-key" {
		t.Errorf("Expected apiKey 'test-api-key', got '%s'", loadedConfig.AI.APIKey)
	}

	if loadedConfig.AI.Concurrency.MaxAnalyses != 100 {
		t.Errorf("Expected maxAnalyses 100, got %d", loadedConfig.AI.Concurrency.MaxAnalyses)
	}

	if loadedConfig.AI.Budgets.MaxSources != 150 {
		t.Errorf("Expected maxSources 150, got %d", loadedConfig.AI.Budgets.MaxSources)
	}

	if loadedConfig.AI.Budgets.MaxAPICalls != 3000 {
		t.Errorf("Expected maxAPICalls 3000, got %d", loadedConfig.AI.Budgets.MaxAPICalls)
	}

	if loadedConfig.AI.Budgets.MaxTokens != 750000 {
		t.Errorf("Expected maxTokens 750000, got %d", loadedConfig.AI.Budgets.MaxTokens)
	}

	if !loadedConfig.AI.Autonomous.Enabled {
		t.Error("Expected autonomous.enabled true")
	}

	if !loadedConfig.AI.Autonomous.AutoApprove.Enabled {
		t.Error("Expected autoApprove.enabled true")
	}

	if len(loadedConfig.AI.Autonomous.AutoApprove.Rules) != 2 {
		t.Errorf("Expected 2 auto-approve rule sets, got %d", len(loadedConfig.AI.Autonomous.AutoApprove.Rules))
	}

	if !loadedConfig.AI.Redaction.Enabled {
		t.Error("Expected redaction.enabled true")
	}

	if len(loadedConfig.AI.Redaction.Denylist) != 2 {
		t.Errorf("Expected 2 denylist items, got %d", len(loadedConfig.AI.Redaction.Denylist))
	}
}
