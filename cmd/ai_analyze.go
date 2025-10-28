package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/internal/ai/factory"
	"github.com/pickjonathan/sdek-cli/internal/analyze"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/pickjonathan/sdek-cli/ui/components"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// aiAnalyzeCmd represents the 'sdek ai analyze' command for context injection analysis
var aiAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze evidence with AI context injection",
	Long: `Analyze evidence with AI context injection using policy excerpts.

This command performs AI-enhanced compliance analysis by injecting policy
context (framework excerpts, control descriptions) into the AI prompt. This
provides more accurate, policy-grounded analysis compared to generic AI analysis.

Key features:
- Context injection: Policy excerpts guide AI analysis
- PII/secret redaction before sending to AI provider
- Response caching for performance
- Confidence scoring with low-confidence flagging
- Detailed findings with citations and residual risk assessment

This is different from 'sdek analyze --ai' which enhances the standard
event-to-control mapping workflow. This command provides specialized
context-grounded analysis for specific policy sections.`,
	Example: `  # Basic context mode analysis for SOC2 CC6.1 (Access Controls)
  sdek ai analyze --framework SOC2 --section CC6.1 \
      --excerpts-file ./policies/soc2_excerpts.json \
      --evidence-path ./evidence/github_*.json \
      --evidence-path ./evidence/jira_*.json

  # Analyze ISO 27001 section with single evidence source
  sdek ai analyze --framework ISO27001 --section A.9.4.2 \
      --excerpts-file ./policies/iso_excerpts.json \
      --evidence-path ./evidence/audit_logs.json

  # Bypass cache for fresh analysis (useful for testing policy changes)
  sdek ai analyze --framework SOC2 --section CC6.1 \
      --excerpts-file ./policies/soc2_excerpts.json \
      --evidence-path ./evidence/*.json \
      --no-cache

  # Multiple evidence paths from different sources
  sdek ai analyze --framework PCI-DSS --section 8.2.4 \
      --excerpts-file ./policies/pci_excerpts.json \
      --evidence-path ./evidence/github/*.json \
      --evidence-path ./evidence/jira/*.json \
      --evidence-path ./evidence/slack/*.json

  # Specify custom output file for finding results
  sdek ai analyze --framework ISO27001 --section A.9.4.2 \
      --excerpts-file ./policies/iso_excerpts.json \
      --evidence-path ./evidence/*.json \
      --output ./findings/iso_a942_finding.json

Note: Confidence thresholds are configured in config.yaml under ai.context_injection.confidence_threshold
      PII/secrets are automatically redacted before sending to AI providers`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		framework, _ := cmd.Flags().GetString("framework")
		section, _ := cmd.Flags().GetString("section")
		excerptsFile, _ := cmd.Flags().GetString("excerpts-file")
		evidencePaths, _ := cmd.Flags().GetStringSlice("evidence-path")

		if framework == "" {
			return fmt.Errorf("--framework is required")
		}
		if section == "" {
			return fmt.Errorf("--section is required")
		}
		if excerptsFile == "" {
			return fmt.Errorf("--excerpts-file is required")
		}
		if len(evidencePaths) == 0 {
			return fmt.Errorf("--evidence-path is required (at least one path)")
		}

		// Check excerpts file exists
		if _, err := os.Stat(excerptsFile); os.IsNotExist(err) {
			return fmt.Errorf("excerpts file not found: %s", excerptsFile)
		}

		// Validate evidence paths exist
		for _, path := range evidencePaths {
			// Support glob patterns
			matches, err := filepath.Glob(path)
			if err != nil {
				return fmt.Errorf("invalid evidence path pattern: %s: %w", path, err)
			}
			if len(matches) == 0 {
				return fmt.Errorf("no files match evidence path: %s", path)
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		framework, _ := cmd.Flags().GetString("framework")
		section, _ := cmd.Flags().GetString("section")
		excerptsFile, _ := cmd.Flags().GetString("excerpts-file")
		evidencePaths, _ := cmd.Flags().GetStringSlice("evidence-path")

		slog.Info("Starting AI context mode analysis",
			"framework", framework,
			"section", section,
			"excerpts_file", excerptsFile,
			"evidence_paths", len(evidencePaths))

		// Step 2: Load policy excerpts
		slog.Info("Loading policy excerpts", "file", excerptsFile)
		excerpts, err := loadExcerpts(excerptsFile)
		if err != nil {
			return fmt.Errorf("failed to load excerpts: %w", err)
		}

		// Find the excerpt for this framework/section
		excerpt, found := findExcerpt(excerpts, framework, section)
		if !found {
			return fmt.Errorf("excerpt not found for %s %s in %s", framework, section, excerptsFile)
		}

		// Step 3: Build ContextPreamble
		slog.Info("Building context preamble", "framework", framework, "section", section)
		preamble, err := types.NewContextPreamble(
			framework,
			excerpt.Version,
			section,
			excerpt.Text,
			excerpt.RelatedSections,
		)
		if err != nil {
			return fmt.Errorf("failed to create context preamble: %w", err)
		}

		// Step 4: Load evidence from paths
		slog.Info("Loading evidence files", "paths", len(evidencePaths))
		evidence, err := loadEvidenceFromPaths(evidencePaths)
		if err != nil {
			return fmt.Errorf("failed to load evidence: %w", err)
		}

		slog.Info("Evidence loaded", "event_count", len(evidence.Events))

		if len(evidence.Events) == 0 {
			return fmt.Errorf("no evidence events found in specified paths")
		}

		// Step 5: Show interactive context preview (Feature 003) unless --yes flag is set
		skipPreview, _ := cmd.Flags().GetBool("yes")
		if !skipPreview {
			if err := showContextPreview(preamble, len(evidence.Events)); err != nil {
				return fmt.Errorf("preview cancelled or failed: %w", err)
			}
		} else {
			slog.Info("Skipping interactive preview (--yes flag set)")
		}

		// Step 6: Load configuration
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Step 7: Check if AI is enabled
		if !cfg.AI.Enabled {
			return fmt.Errorf("AI analysis is disabled in config. Set ai.enabled=true to use this command")
		}

		// Step 8: Initialize AI engine
		slog.Info("Initializing AI engine", "provider", cfg.AI.Provider)
		engine, err := initializeAIEngine(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize AI engine: %w", err)
		}

		// Step 9: Perform AI analysis
		fmt.Println("\n🤖 Analyzing evidence with AI context injection...")
		finding, err := engine.Analyze(cmd.Context(), *preamble, *evidence)
		if err != nil {
			return fmt.Errorf("AI analysis failed: %w", err)
		}

		slog.Info("AI analysis complete",
			"confidence", finding.ConfidenceScore,
			"controls", len(finding.MappedControls),
			"citations", len(finding.Citations))

		// Step 10: Flag low confidence findings
		confidenceThreshold := preamble.Rubrics.ConfidenceThreshold
		analyze.FlagLowConfidence(finding, confidenceThreshold)

		if finding.ReviewRequired {
			slog.Warn("Low confidence finding flagged for review",
				"confidence", finding.ConfidenceScore,
				"threshold", confidenceThreshold)
		}

		// Step 11: Export finding to output file
		outputFile, _ := cmd.Flags().GetString("output")
		if err := exportFinding(finding, outputFile); err != nil {
			return fmt.Errorf("failed to export finding: %w", err)
		}

		// Step 12: Display summary
		displayFindingSummary(finding, outputFile)

		return nil
	},
}

// showContextPreview displays an interactive preview of the analysis context
func showContextPreview(preamble *types.ContextPreamble, evidenceCount int) error {
	model := components.NewContextPreview(*preamble, evidenceCount)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run preview: %w", err)
	}

	// Check if user confirmed
	previewModel, ok := finalModel.(components.ContextPreviewModel)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	if previewModel.Cancelled() {
		return fmt.Errorf("user cancelled")
	}

	if !previewModel.Confirmed() {
		return fmt.Errorf("preview not confirmed")
	}

	return nil
}

// loadConfig loads the configuration from Viper (which is already initialized by root.go)
func loadConfig() (*types.Config, error) {
	cfg := &types.Config{}

	// Unmarshal viper config into types.Config
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

// initializeAIEngine creates an AI engine based on the config
func initializeAIEngine(cfg *types.Config) (ai.Engine, error) {
	provider := cfg.AI.Provider
	if provider == "" {
		provider = "openai" // Default
	}

	model := cfg.AI.Model
	if model == "" {
		if provider == "openai" {
			model = "gpt-4"
		} else if provider == "anthropic" {
			model = "claude-3-opus-20240229"
		}
	}

	// Build provider configuration
	providerConfig := types.ProviderConfig{
		APIKey:      cfg.AI.APIKey,
		Model:       model,
		MaxTokens:   cfg.AI.MaxTokens,
		Temperature: float64(cfg.AI.Temperature),
		Timeout:     cfg.AI.Timeout,
		MaxRetries:  3,
	}

	// Override with environment variables if not set
	if providerConfig.APIKey == "" {
		if provider == "openai" {
			providerConfig.APIKey = os.Getenv("SDEK_OPENAI_KEY")
			if providerConfig.APIKey == "" {
				providerConfig.APIKey = cfg.AI.OpenAIKey
			}
		} else if provider == "anthropic" {
			providerConfig.APIKey = os.Getenv("SDEK_ANTHROPIC_KEY")
			if providerConfig.APIKey == "" {
				providerConfig.APIKey = cfg.AI.AnthropicKey
			}
		}
	}

	// Set defaults
	if providerConfig.Timeout == 0 {
		providerConfig.Timeout = 60
	}
	if providerConfig.MaxTokens == 0 {
		providerConfig.MaxTokens = 4096
	}
	if providerConfig.Temperature == 0 {
		providerConfig.Temperature = 0.3
	}

	// Determine provider URL
	providerURL := cfg.AI.ProviderURL
	if providerURL == "" {
		switch provider {
		case "openai":
			providerURL = "openai://api.openai.com"
		case "anthropic":
			providerURL = "anthropic://api.anthropic.com"
		case "ollama":
			providerURL = "ollama://localhost:11434"
		case "gemini":
			providerURL = "gemini://generativelanguage.googleapis.com"
		default:
			return nil, fmt.Errorf("unsupported AI provider: %s (use provider_url config)", provider)
		}
	}

	// Validate API key (not required for local providers like Ollama)
	requiresAPIKey := !strings.Contains(strings.ToLower(providerURL), "ollama://")
	if requiresAPIKey && providerConfig.APIKey == "" {
		return nil, fmt.Errorf("API key required for %s - set SDEK_%s_KEY or configure in config.yaml", provider, strings.ToUpper(provider))
	}

	// Create provider
	aiProvider, err := factory.CreateProvider(providerURL, providerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	// Create engine
	engine := ai.NewEngine(cfg, aiProvider)
	return engine, nil
}

// exportFinding saves the finding to a JSON file
func exportFinding(finding *types.Finding, outputPath string) error {
	data, err := json.MarshalIndent(finding, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal finding: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// displayFindingSummary shows a summary of the finding to the user
func displayFindingSummary(finding *types.Finding, outputFile string) {
	fmt.Println("\n✅ Analysis Complete!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Framework:       %s\n", finding.FrameworkID)
	fmt.Printf("Control:         %s\n", finding.ControlID)
	fmt.Printf("Confidence:      %.1f%%\n", finding.ConfidenceScore*100)
	fmt.Printf("Residual Risk:   %s\n", finding.ResidualRisk)

	if finding.ReviewRequired {
		fmt.Println("⚠️  Review Required: Low confidence score")
	}

	fmt.Printf("\nMapped Controls: %d\n", len(finding.MappedControls))
	if len(finding.MappedControls) > 0 {
		for _, ctrl := range finding.MappedControls {
			fmt.Printf("  - %s\n", ctrl)
		}
	}

	fmt.Printf("\nCitations:       %d\n", len(finding.Citations))
	if len(finding.Citations) > 0 && len(finding.Citations) <= 5 {
		for _, cite := range finding.Citations {
			fmt.Printf("  - %s\n", cite)
		}
	} else if len(finding.Citations) > 5 {
		fmt.Printf("  (showing first 5 of %d)\n", len(finding.Citations))
		for i := 0; i < 5; i++ {
			fmt.Printf("  - %s\n", finding.Citations[i])
		}
	}

	fmt.Printf("\nJustification:\n%s\n", finding.Justification)

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📄 Finding saved to: %s\n", outputFile)
}

func init() {
	aiCmd.AddCommand(aiAnalyzeCmd)

	// Required flags
	aiAnalyzeCmd.Flags().String("framework", "", "Framework name (e.g., SOC2, ISO27001, PCI-DSS)")
	aiAnalyzeCmd.Flags().String("section", "", "Section ID (e.g., CC6.1, A.9.4.2)")
	aiAnalyzeCmd.Flags().String("excerpts-file", "", "Path to policy excerpts JSON file")
	aiAnalyzeCmd.Flags().StringSlice("evidence-path", []string{}, "Evidence file paths (supports globs, can be specified multiple times)")

	// Optional flags
	aiAnalyzeCmd.Flags().Bool("no-cache", false, "Bypass cache and perform fresh analysis")
	aiAnalyzeCmd.Flags().String("output", "findings.json", "Output file for finding results")
	aiAnalyzeCmd.Flags().BoolP("yes", "y", false, "Skip interactive preview and auto-approve analysis")

	aiAnalyzeCmd.MarkFlagRequired("framework")
	aiAnalyzeCmd.MarkFlagRequired("section")
	aiAnalyzeCmd.MarkFlagRequired("excerpts-file")
	aiAnalyzeCmd.MarkFlagRequired("evidence-path")
}

// Excerpt represents a policy excerpt from the excerpts file
type Excerpt struct {
	Framework       string   `json:"framework"`
	Version         string   `json:"version"`
	Section         string   `json:"section"`
	Text            string   `json:"text"`
	RelatedSections []string `json:"related_sections,omitempty"`
}

// loadExcerpts loads policy excerpts from a JSON file
// Supports both array format and map format (legacy)
func loadExcerpts(filepath string) ([]Excerpt, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Try array format first (new format)
	var excerpts []Excerpt
	if err := json.Unmarshal(data, &excerpts); err == nil {
		return excerpts, nil
	}

	// Fall back to map format (legacy format from testdata)
	var excerptMap map[string]struct {
		ControlID string `json:"control_id"`
		Title     string `json:"title"`
		Excerpt   string `json:"excerpt"`
	}
	if err := json.Unmarshal(data, &excerptMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON (tried array and map formats): %w", err)
	}

	// Convert map to array format
	excerpts = make([]Excerpt, 0, len(excerptMap))
	for section, e := range excerptMap {
		excerpts = append(excerpts, Excerpt{
			Framework: "",     // Will be filled from command flag
			Version:   "2023", // Default version for legacy format
			Section:   section,
			Text:      e.Excerpt,
		})
	}

	return excerpts, nil
}

// findExcerpt finds an excerpt matching framework and section
// If framework is empty in excerpt (legacy map format), match on section only
func findExcerpt(excerpts []Excerpt, framework, section string) (Excerpt, bool) {
	for _, e := range excerpts {
		// Match section first
		if e.Section != section {
			continue
		}
		// If excerpt has no framework (legacy format), accept it
		if e.Framework == "" {
			e.Framework = framework // Fill in the framework from request
			return e, true
		}
		// Otherwise require exact framework match
		if e.Framework == framework {
			return e, true
		}
	}
	return Excerpt{}, false
}

// loadEvidenceFromPaths loads evidence events from file paths (supports globs)
func loadEvidenceFromPaths(paths []string) (*types.EvidenceBundle, error) {
	bundle := &types.EvidenceBundle{
		Events: []types.EvidenceEvent{},
	}

	for _, pattern := range paths {
		// Expand glob pattern
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %s: %w", pattern, err)
		}

		// Load each matched file
		for _, path := range matches {
			events, err := loadEventsFromFile(path)
			if err != nil {
				slog.Warn("Failed to load evidence file", "path", path, "error", err)
				continue
			}
			bundle.Events = append(bundle.Events, events...)
		}
	}

	return bundle, nil
}

// loadEventsFromFile loads events from a single JSON file
func loadEventsFromFile(filepath string) ([]types.EvidenceEvent, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var events []types.EvidenceEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return events, nil
}
