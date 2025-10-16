package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/internal/ai/providers"
	"github.com/pickjonathan/sdek-cli/internal/analyze"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/spf13/cobra"
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze events and map them to compliance controls",
	Long: `Analyze ingested events and map them to compliance framework controls.

The analyze command performs evidence mapping by:
1. Loading events from the state
2. Matching events to controls using keyword-based heuristics
3. Calculating confidence scores for each evidence mapping
4. Computing risk scores for each control
5. Generating findings for controls with insufficient evidence or high risk

This command should be run after ingesting new events to update the
compliance posture analysis.`,
	Example: `  # Analyze all events and update control mappings
  sdek analyze

  # Run analysis with verbose logging
  sdek analyze --verbose`,
	RunE: runAnalyze,
}

var (
	aiEnabled  bool
	aiProvider string
	aiModel    string
	aiCacheDir string
	aiTimeout  int
	noAI       bool
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// AI-related flags (Feature 002: AI Evidence Analysis)
	analyzeCmd.Flags().BoolVar(&aiEnabled, "ai", false, "Enable AI-enhanced analysis")
	analyzeCmd.Flags().StringVar(&aiProvider, "ai-provider", "openai", "AI provider: openai, anthropic")
	analyzeCmd.Flags().StringVar(&aiModel, "ai-model", "", "AI model name (overrides config)")
	analyzeCmd.Flags().StringVar(&aiCacheDir, "cache-dir", "", "AI cache directory (overrides config)")
	analyzeCmd.Flags().IntVar(&aiTimeout, "ai-timeout", 0, "AI request timeout in seconds (overrides config)")
	analyzeCmd.Flags().BoolVar(&noAI, "no-ai", false, "Disable AI analysis (force heuristic-only)")
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	slog.Info("Starting analyze command")

	// Load existing state
	state, err := store.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Check if we have events to analyze
	if len(state.Events) == 0 {
		return fmt.Errorf("no events found to analyze, run 'sdek seed --demo' or 'sdek ingest' first")
	}

	// Check if we have controls
	if len(state.Controls) == 0 {
		slog.Info("No controls found, initializing frameworks")
		frameworks, controls := initializeFrameworks()
		state.Frameworks = frameworks
		state.Controls = controls
	}

	// Map events to controls (evidence generation)
	slog.Info("Mapping events to controls", "eventCount", len(state.Events), "controlCount", len(state.Controls))

	// Determine if AI should be used
	useAI := shouldUseAI(state.Config)

	var mapper *analyze.Mapper
	if useAI {
		slog.Info("Initializing AI-enhanced mapper", "provider", getAIProvider(state.Config))
		aiMapper, err := initializeAIMapper(state.Config)
		if err != nil {
			slog.Warn("Failed to initialize AI mapper, falling back to heuristic-only", "error", err)
			mapper = analyze.NewMapper()
		} else {
			mapper = aiMapper
		}
	} else {
		slog.Info("Using heuristic-only mapper")
		mapper = analyze.NewMapper()
	}

	evidence := mapper.MapEventsToControls(state.Events)
	state.Evidence = evidence

	slog.Info("Generated evidence mappings", "count", len(evidence))

	// Calculate risk scores and update control statuses
	slog.Info("Calculating risk scores")
	scorer := analyze.NewRiskScorer()

	// Clear existing findings
	state.Findings = []types.Finding{}

	// Update each control with its risk status
	for i := range state.Controls {
		control := &state.Controls[i]

		// Get evidence for this control
		controlEvidence := filterEvidenceByControl(evidence, control.ID)
		control.EvidenceCount = len(controlEvidence)

		// Generate findings for this control
		controlFindings := scorer.GenerateFindingsForControl(*control, controlEvidence)

		// Calculate risk status
		control.RiskStatus = scorer.CalculateControlRisk(controlFindings, len(controlEvidence))

		// Add findings to state
		state.Findings = append(state.Findings, controlFindings...)
	}

	// Update framework compliance percentages
	slog.Info("Updating framework compliance")
	for i := range state.Frameworks {
		framework := &state.Frameworks[i]
		frameworkControls := filterControlsByFramework(state.Controls, framework.ID)

		greenCount := 0
		yellowCount := 0
		redCount := 0

		for _, ctrl := range frameworkControls {
			switch ctrl.RiskStatus {
			case types.RiskStatusGreen:
				greenCount++
			case types.RiskStatusYellow:
				yellowCount++
			case types.RiskStatusRed:
				redCount++
			}
		}

		if len(frameworkControls) > 0 {
			framework.CompliancePercentage = float64(greenCount) / float64(len(frameworkControls)) * 100
		}

		slog.Info("Framework status",
			"framework", framework.Name,
			"green", greenCount,
			"yellow", yellowCount,
			"red", redCount,
			"compliance", fmt.Sprintf("%.1f%%", framework.CompliancePercentage))
	}

	// Save updated state
	slog.Info("Saving state")
	if err := state.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Print summary
	fmt.Println("✓ Analysis completed successfully!")
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Events analyzed:    %d\n", len(state.Events))
	fmt.Printf("  Evidence generated: %d\n", len(state.Evidence))
	fmt.Printf("  Findings:           %d\n", len(state.Findings))
	fmt.Println()
	fmt.Println("Compliance Status:")
	for _, fw := range state.Frameworks {
		fmt.Printf("  %-15s %.1f%% compliant\n", fw.Name, fw.CompliancePercentage)
	}
	fmt.Println()

	// Show risk distribution
	greenCount := 0
	yellowCount := 0
	redCount := 0
	for _, ctrl := range state.Controls {
		switch ctrl.RiskStatus {
		case types.RiskStatusGreen:
			greenCount++
		case types.RiskStatusYellow:
			yellowCount++
		case types.RiskStatusRed:
			redCount++
		}
	}

	fmt.Println("Risk Distribution:")
	fmt.Printf("  Green (Low Risk):     %d controls\n", greenCount)
	fmt.Printf("  Yellow (Medium Risk): %d controls\n", yellowCount)
	fmt.Printf("  Red (High Risk):      %d controls\n", redCount)
	fmt.Println()

	fmt.Println("Next steps:")
	fmt.Println("  - Run 'sdek tui' to explore the analysis interactively")
	fmt.Println("  - Run 'sdek report' to export a detailed compliance report")

	slog.Info("Analyze command completed successfully")
	return nil
}

// shouldUseAI determines if AI should be enabled based on flags and config
func shouldUseAI(config *types.Config) bool {
	// --no-ai flag takes highest priority
	if noAI {
		return false
	}

	// --ai flag enables AI (overrides config)
	if aiEnabled {
		return true
	}

	// Fall back to config setting
	return config != nil && config.AI.Enabled
}

// getAIProvider returns the AI provider to use based on flags and config
func getAIProvider(config *types.Config) string {
	// Command-line flag overrides config
	if aiProvider != "" && aiProvider != "openai" {
		return aiProvider
	}

	// Fall back to config
	if config != nil && config.AI.Provider != "" {
		return config.AI.Provider
	}

	return types.AIProviderOpenAI
}

// getAIModel returns the AI model to use based on flags and config
func getAIModel(config *types.Config, provider string) string {
	// Command-line flag overrides config
	if aiModel != "" {
		return aiModel
	}

	// Fall back to config
	if config != nil && config.AI.Model != "" {
		return config.AI.Model
	}

	// Provider-specific defaults
	if provider == types.AIProviderAnthropic {
		return "claude-3-opus"
	}
	return "gpt-4"
}

// getAITimeout returns the AI timeout to use based on flags and config
func getAITimeout(config *types.Config) int {
	// Command-line flag overrides config
	if aiTimeout > 0 {
		return aiTimeout
	}

	// Fall back to config
	if config != nil && config.AI.Timeout > 0 {
		return config.AI.Timeout
	}

	return 60 // Default 60 seconds
}

// getAICacheDir returns the AI cache directory to use based on flags and config
func getAICacheDir(config *types.Config) string {
	// Command-line flag overrides config
	if aiCacheDir != "" {
		return aiCacheDir
	}

	// Fall back to config
	if config != nil && config.AI.CacheDir != "" {
		return config.AI.CacheDir
	}

	return "" // Empty string means use default
}

// initializeAIMapper creates an AI-enhanced mapper with the configured provider
func initializeAIMapper(config *types.Config) (*analyze.Mapper, error) {
	provider := getAIProvider(config)
	model := getAIModel(config, provider)
	timeout := getAITimeout(config)
	cacheDir := getAICacheDir(config)

	slog.Info("AI configuration",
		"provider", provider,
		"model", model,
		"timeout", timeout,
		"cacheDir", cacheDir)

	// Build AIConfig from settings
	aiConfig := ai.AIConfig{
		Provider:     string(provider),
		Enabled:      true,
		Model:        model,
		Timeout:      timeout,
		OpenAIKey:    os.Getenv("SDEK_OPENAI_KEY"),
		AnthropicKey: os.Getenv("SDEK_ANTHROPIC_KEY"),
	}

	// Override with config values if available
	if config != nil {
		if aiConfig.OpenAIKey == "" {
			aiConfig.OpenAIKey = config.AI.OpenAIKey
		}
		if aiConfig.AnthropicKey == "" {
			aiConfig.AnthropicKey = config.AI.AnthropicKey
		}
		if config.AI.RateLimit > 0 {
			aiConfig.RateLimit = config.AI.RateLimit
		} else {
			aiConfig.RateLimit = 10 // Default 10 requests per minute
		}
	} else {
		aiConfig.RateLimit = 10
	}

	// Set defaults for fields not in types.AIConfig
	aiConfig.MaxTokens = 4096  // Default token limit
	aiConfig.Temperature = 0.3 // Default temperature for deterministic output

	// Validate API key for the selected provider
	if provider == types.AIProviderOpenAI && aiConfig.OpenAIKey == "" {
		return nil, fmt.Errorf("OpenAI API key required - set SDEK_OPENAI_KEY environment variable or configure in config.yaml")
	}
	if provider == types.AIProviderAnthropic && aiConfig.AnthropicKey == "" {
		return nil, fmt.Errorf("Anthropic API key required - set SDEK_ANTHROPIC_KEY environment variable or configure in config.yaml")
	}

	// Create AI engine based on provider
	var engine ai.Engine
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	switch provider {
	case types.AIProviderOpenAI:
		slog.Info("Initializing OpenAI engine", "model", model, "rateLimit", aiConfig.RateLimit)
		engine, err = providers.NewOpenAIEngine(aiConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI engine: %w", err)
		}

	case types.AIProviderAnthropic:
		slog.Info("Initializing Anthropic engine", "model", model, "rateLimit", aiConfig.RateLimit)
		engine, err = providers.NewAnthropicEngine(aiConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Anthropic engine: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}

	// Test AI engine health
	slog.Info("Testing AI engine health")
	if err := engine.Health(ctx); err != nil {
		slog.Warn("AI engine health check failed", "error", err)
		return nil, fmt.Errorf("AI engine health check failed: %w", err)
	}
	slog.Info("AI engine health check passed")

	// Create cache
	cache, err := ai.NewCache(cacheDir)
	if err != nil {
		slog.Warn("Failed to create AI cache, continuing without cache", "error", err)
		cache = nil
	} else {
		slog.Info("AI cache initialized", "dir", cacheDir)
	}

	// Create AI-enhanced mapper
	slog.Info("Creating AI-enhanced mapper")
	mapper := analyze.NewMapperWithAI(engine, cache)

	return mapper, nil
}
