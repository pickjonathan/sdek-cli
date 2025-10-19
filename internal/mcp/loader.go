package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// Loader discovers and loads MCP configurations from multiple sources.
type Loader struct {
	globalDir  string
	projectDir string
}

// NewLoader creates a new config loader.
// globalDir is typically the user's home directory.
// projectDir is typically the current working directory.
func NewLoader(globalDir, projectDir string) *Loader {
	if globalDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			globalDir = home
		}
	}
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}

	return &Loader{
		globalDir:  globalDir,
		projectDir: projectDir,
	}
}

// LoadConfigs discovers and loads all MCP configurations.
// Precedence (highest to lowest):
// 1. $SDEK_MCP_PATH (colon-separated paths)
// 2. Project: ./.sdek/mcp/*.json
// 3. Global: ~/.sdek/mcp/*.json
func (l *Loader) LoadConfigs() ([]*types.MCPConfig, error) {
	configMap := make(map[string]*types.MCPConfig) // name -> config

	// Load from global (lowest precedence)
	if l.globalDir != "" {
		globalPath := filepath.Join(l.globalDir, ".sdek", "mcp")
		l.loadFromDirectory(globalPath, configMap)
	}

	// Load from project (overrides global)
	if l.projectDir != "" {
		projectPath := filepath.Join(l.projectDir, ".sdek", "mcp")
		l.loadFromDirectory(projectPath, configMap)
	}

	// Load from SDEK_MCP_PATH (highest precedence)
	if mcpPath := os.Getenv("SDEK_MCP_PATH"); mcpPath != "" {
		paths := strings.Split(mcpPath, ":")
		for _, path := range paths {
			l.loadFromDirectory(strings.TrimSpace(path), configMap)
		}
	}

	// Convert map to slice
	configs := make([]*types.MCPConfig, 0, len(configMap))
	for _, config := range configMap {
		configs = append(configs, config)
	}

	return configs, nil
}

// loadFromDirectory loads all JSON configs from a directory.
func (l *Loader) loadFromDirectory(dir string, configMap map[string]*types.MCPConfig) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		// Directory doesn't exist or not accessible - not an error
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		configPath := filepath.Join(dir, entry.Name())
		config, err := l.loadConfig(configPath)
		if err != nil {
			// Log error but continue with other configs
			fmt.Fprintf(os.Stderr, "warning: failed to load %s: %v\n", configPath, err)
			continue
		}

		// Store in map (will override if name already exists)
		configMap[config.Name] = config
	}
}

// loadConfig loads a single config file with environment variable expansion.
func (l *Loader) loadConfig(path string) (*types.MCPConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var config types.MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	// Expand environment variables in env map
	if config.Env != nil {
		for key, value := range config.Env {
			config.Env[key] = os.ExpandEnv(value)
		}
	}

	// Validate the config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &config, nil
}
