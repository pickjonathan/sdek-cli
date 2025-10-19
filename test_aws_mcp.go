package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

func main() {
	// Load the AWS MCP configuration
	configPath := filepath.Join(os.Getenv("HOME"), ".sdek", "mcp", "aws.json")

	fmt.Printf("Loading configuration from: %s\n", configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config types.MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	fmt.Printf("\n✓ Configuration loaded successfully\n")
	fmt.Printf("  Name: %s\n", config.Name)
	fmt.Printf("  Command: %s %v\n", config.Command, config.Args)
	fmt.Printf("  Transport: %s\n", config.Transport)
	fmt.Printf("  Capabilities: %v\n", config.Capabilities)
	fmt.Printf("  Timeout: %s\n", config.Timeout)
	fmt.Printf("  Schema Version: %s\n", config.SchemaVersion)

	// Validate the configuration
	fmt.Printf("\nValidating configuration...\n")
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}
	fmt.Printf("✓ Configuration is valid\n")

	// Create validator
	fmt.Printf("\nInitializing JSON Schema validator...\n")
	validator := mcp.NewValidator()
	fmt.Printf("✓ Validator created successfully\n")

	// Validate against JSON Schema
	fmt.Printf("\nValidating against JSON Schema...\n")
	if err := validator.ValidateConfig(&config); err != nil {
		log.Fatalf("Schema validation failed: %v", err)
	}
	fmt.Printf("✓ Schema validation passed\n")

	// Test if uvx is available
	fmt.Printf("\nChecking if uvx is installed...\n")
	cmd := exec.Command("which", "uvx")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("⚠ uvx not found in PATH\n")
		fmt.Printf("  Install with: curl -LsSf https://astral.sh/uv/install.sh | sh\n")
	} else {
		fmt.Printf("✓ uvx found at %s\n", strings.TrimSpace(string(output)))
	}

	// Test if AWS credentials are configured
	fmt.Printf("\nChecking AWS credentials...\n")
	awsConfigPath := filepath.Join(os.Getenv("HOME"), ".aws", "credentials")
	if _, err := os.Stat(awsConfigPath); err == nil {
		fmt.Printf("✓ AWS credentials file found\n")
	} else {
		fmt.Printf("⚠ AWS credentials file not found at %s\n", awsConfigPath)
		fmt.Printf("  Configure with: aws configure\n")
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("✓ AWS MCP Configuration Test Complete\n")
	fmt.Printf("%s\n", strings.Repeat("=", 60))
	fmt.Printf("\nConfiguration is valid and ready to use!\n")
	fmt.Printf("\nTo test actual connection, the Registry implementation will:\n")
	fmt.Printf("1. Spawn the MCP server process using the stdio transport\n")
	fmt.Printf("2. Perform JSON-RPC handshake\n")
	fmt.Printf("3. Discover available tools/resources\n")
	fmt.Printf("4. Monitor health status\n")
}
