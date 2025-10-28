package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/internal/ai/factory"
	"github.com/pickjonathan/sdek-cli/internal/analyze"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/pickjonathan/sdek-cli/ui/components"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// aiPlanCmd represents the 'sdek ai plan' command for autonomous evidence collection
var aiPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generate autonomous evidence collection plan",
	Long: `Generate and execute an AI-driven autonomous evidence collection plan.

This command uses AI to generate a strategic plan for collecting compliance
evidence from multiple sources (GitHub, Jira, AWS, Slack, CI/CD, Documentation),
then executes the plan to gather relevant evidence and perform analysis.

Workflow:
1. AI analyzes the policy excerpt and generates collection queries
2. Plan is reviewed (TUI) or auto-approved (--approve-all) or previewed (--dry-run)
3. Approved queries are executed across configured sources
4. Collected evidence is analyzed with context injection
5. Finding is exported with full provenance tracking

This enables truly autonomous compliance evidence collection where the AI
determines what data to collect and from where, based on policy requirements.

Connectors:
The command requires at least one enabled connector in config.yaml (ai.connectors).
Supported connectors: github, jira, aws, slack
Configure connectors with API keys, endpoints, and rate limits as needed.`,
	Example: `  # Generate plan with interactive approval
  sdek ai plan --framework SOC2 --section CC6.1 \
      --excerpts-file ./policies/soc2_excerpts.json

  # Preview plan without execution (dry-run)
  sdek ai plan --framework ISO27001 --section A.9.4.2 \
      --excerpts-file ./policies/iso_excerpts.json \
      --dry-run

  # Auto-approve all plan items and execute
  sdek ai plan --framework PCI-DSS --section 8.2.4 \
      --excerpts-file ./policies/pci_excerpts.json \
      --approve-all

  # Specify custom output file for finding results
  sdek ai plan --framework SOC2 --section CC6.1 \
      --excerpts-file ./policies/soc2_excerpts.json \
      --output ./findings/autonomous_cc61.json

Note: Budget limits (max sources, API calls, tokens) are configured in config.yaml
      Auto-approve patterns can be configured under ai.autonomous.autoApprove`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		framework, _ := cmd.Flags().GetString("framework")
		section, _ := cmd.Flags().GetString("section")
		excerptsFile, _ := cmd.Flags().GetString("excerpts-file")

		if framework == "" {
			return fmt.Errorf("--framework is required")
		}
		if section == "" {
			return fmt.Errorf("--section is required")
		}
		if excerptsFile == "" {
			return fmt.Errorf("--excerpts-file is required")
		}

		// Check excerpts file exists
		if _, err := os.Stat(excerptsFile); os.IsNotExist(err) {
			return fmt.Errorf("excerpts file not found: %s", excerptsFile)
		}

		// Check AI is enabled in config
		if !viper.GetBool("ai.enabled") {
			return fmt.Errorf("AI features are disabled. Enable in config.yaml (ai.enabled: true)")
		}

		// Check autonomous mode is enabled
		if !viper.GetBool("ai.autonomous.enabled") {
			return fmt.Errorf("autonomous mode is disabled. Enable in config.yaml (ai.autonomous.enabled: true)")
		}

		// Verify AI provider configuration
		provider := viper.GetString("ai.provider")
		providerURL := viper.GetString("ai.provider_url")

		// Check if using provider_url (Feature 006) or legacy provider (Feature 003)
		if providerURL == "" && provider == "" {
			return fmt.Errorf("AI provider not configured (set ai.provider or ai.provider_url)")
		}

		// Check API key is configured (not required for local providers like Ollama)
		requiresAPIKey := provider != "ollama" && !strings.Contains(strings.ToLower(providerURL), "ollama://")
		if requiresAPIKey {
			apiKey := viper.GetString("ai.apiKey")
			if apiKey == "" {
				// Try provider-specific keys as fallback
				if provider == "openai" {
					apiKey = viper.GetString("ai.openai_key")
				} else if provider == "anthropic" {
					apiKey = viper.GetString("ai.anthropic_key")
				}
				if apiKey == "" {
					return fmt.Errorf("AI API key not configured (ai.apiKey or ai.%s_key)", provider)
				}
			}
		}

		// Validate connector configuration
		connectors := viper.GetStringMap("ai.connectors")
		if len(connectors) == 0 {
			slog.Warn("No connectors configured - autonomous mode requires at least one enabled connector")
		} else {
			// Check if at least one connector is enabled
			hasEnabled := false
			for name := range connectors {
				if viper.GetBool(fmt.Sprintf("ai.connectors.%s.enabled", name)) {
					hasEnabled = true
					break
				}
			}
			if !hasEnabled {
				return fmt.Errorf("no connectors enabled - autonomous mode requires at least one enabled connector (github, jira, aws, or slack)")
			}
		}

		return nil
	},
	RunE: runAIPlan,
}

func init() {
	aiCmd.AddCommand(aiPlanCmd)

	aiPlanCmd.Flags().String("framework", "", "Framework name (e.g., SOC2, ISO27001, PCI-DSS)")
	aiPlanCmd.Flags().String("section", "", "Section ID (e.g., CC6.1, A.9.4.2)")
	aiPlanCmd.Flags().String("excerpts-file", "", "Path to policy excerpts JSON file")
	aiPlanCmd.Flags().Bool("dry-run", false, "Preview plan without execution")
	aiPlanCmd.Flags().Bool("approve-all", false, "Auto-approve all plan items without TUI")
	aiPlanCmd.Flags().String("output", "findings.json", "Output file path for finding results")

	aiPlanCmd.MarkFlagRequired("framework")
	aiPlanCmd.MarkFlagRequired("section")
	aiPlanCmd.MarkFlagRequired("excerpts-file")
}

func runAIPlan(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// Get flags
	framework, _ := cmd.Flags().GetString("framework")
	section, _ := cmd.Flags().GetString("section")
	excerptsFile, _ := cmd.Flags().GetString("excerpts-file")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	approveAll, _ := cmd.Flags().GetBool("approve-all")
	outputFile, _ := cmd.Flags().GetString("output")

	slog.Info("Starting AI plan generation", "framework", framework, "section", section, "dryRun", dryRun, "approveAll", approveAll)

	// Step 1: Load config
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Step 2: Load policy excerpts
	slog.Info("Loading policy excerpts", "file", excerptsFile)
	excerpts, err := loadExcerpts(excerptsFile)
	if err != nil {
		return fmt.Errorf("failed to load excerpts: %w", err)
	}

	// Find the specific excerpt for this framework and section
	excerpt, found := findExcerpt(excerpts, framework, section)
	if !found {
		return fmt.Errorf("no excerpt found for framework=%s section=%s in %s", framework, section, excerptsFile)
	}

	// Step 3: Build context preamble
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

	// Step 4: Initialize AI provider and engine
	slog.Info("Initializing AI engine", "provider", cfg.AI.Provider)

	// Build provider configuration
	providerConfig := types.ProviderConfig{
		APIKey:      cfg.AI.APIKey,
		Model:       cfg.AI.Model,
		MaxTokens:   cfg.AI.MaxTokens,
		Temperature: float64(cfg.AI.Temperature),
		Timeout:     cfg.AI.Timeout,
		MaxRetries:  3, // Default retries
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
		// Construct URL from provider name
		switch cfg.AI.Provider {
		case types.AIProviderOpenAI:
			providerURL = "openai://api.openai.com"
			if providerConfig.APIKey == "" {
				providerConfig.APIKey = cfg.AI.OpenAIKey
			}
		case types.AIProviderAnthropic:
			providerURL = "anthropic://api.anthropic.com"
			if providerConfig.APIKey == "" {
				providerConfig.APIKey = cfg.AI.AnthropicKey
			}
		case "ollama":
			providerURL = "ollama://localhost:11434"
		case "gemini":
			providerURL = "gemini://generativelanguage.googleapis.com"
		default:
			return fmt.Errorf("unsupported AI provider: %s (use provider_url config instead)", cfg.AI.Provider)
		}
	}

	// Validate API key (not required for local providers like Ollama)
	requiresAPIKey := !strings.Contains(strings.ToLower(providerURL), "ollama://")
	if requiresAPIKey && providerConfig.APIKey == "" {
		return fmt.Errorf("API key not configured for provider: %s", cfg.AI.Provider)
	}

	// Create provider
	provider, err := factory.CreateProvider(providerURL, providerConfig)
	if err != nil {
		return fmt.Errorf("failed to create AI provider: %w", err)
	}

	// Create engine with MCP support (Feature 006)
	engine, err := ai.NewEngineWithMCP(cmd.Context(), cfg, provider)
	if err != nil {
		return fmt.Errorf("failed to create AI engine: %w", err)
	}

	// Log enabled connectors
	if len(cfg.AI.Connectors) > 0 {
		enabledConnectors := []string{}
		for name, conn := range cfg.AI.Connectors {
			if conn.Enabled {
				enabledConnectors = append(enabledConnectors, name)
			}
		}
		if len(enabledConnectors) > 0 {
			slog.Info("Connectors enabled", "connectors", enabledConnectors)
		}
	}

	// Step 5: Generate plan
	slog.Info("Generating evidence collection plan")
	plan, err := engine.ProposePlan(cmd.Context(), *preamble)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	slog.Info("Plan generated", "items", len(plan.Items), "autoApproved", countAutoApproved(plan))

	// Step 6: Handle dry-run
	if dryRun {
		fmt.Println("\n=== Evidence Collection Plan (Dry Run) ===")
		fmt.Printf("Framework: %s %s\n", plan.Framework, plan.Section)
		fmt.Printf("Generated: %s\n", plan.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Total Items: %d\n", len(plan.Items))
		fmt.Printf("Auto-Approved: %d\n", countAutoApproved(plan))
		fmt.Printf("Budget Estimate: %d sources, %d API calls, %d tokens\n", plan.EstimatedSources, plan.EstimatedCalls, plan.EstimatedTokens)
		fmt.Println("\nPlan Items:")
		for i, item := range plan.Items {
			status := "pending"
			if item.ApprovalStatus == types.ApprovalAutoApproved {
				status = "auto-approved ✓"
			}
			fmt.Printf("  %d. [%s] %s: %s\n", i+1, status, item.Source, item.Query)
			fmt.Printf("     Signal: %.2f, Rationale: %s\n", item.SignalStrength, item.Rationale)
		}
		fmt.Println("\n[Dry run - no execution performed]")
		return nil
	}

	// Step 7: Get approval (TUI or auto-approve)
	if approveAll {
		// Auto-approve all items
		for i := range plan.Items {
			plan.Items[i].ApprovalStatus = types.ApprovalApproved
		}
		plan.Status = types.PlanApproved
		slog.Info("Auto-approved all plan items", "count", len(plan.Items))
	} else {
		// Launch TUI for interactive approval
		model := components.NewPlanApproval(plan)
		p := tea.NewProgram(model)
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}

		// Extract approved plan
		approvedModel := finalModel.(components.PlanApprovalModel)
		if approvedModel.Cancelled() {
			return fmt.Errorf("plan cancelled by user")
		}
		plan = approvedModel.GetPlan()

		approvedCount := countApproved(plan)
		slog.Info("Plan approved", "approved", approvedCount, "total", len(plan.Items))
	}

	// Step 8: Execute plan
	slog.Info("Executing evidence collection plan")
	bundle, err := engine.ExecutePlan(cmd.Context(), plan)
	if err != nil {
		return fmt.Errorf("failed to execute plan: %w", err)
	}

	slog.Info("Evidence collected", "events", len(bundle.Events))

	// Step 9: Analyze collected evidence with context injection
	slog.Info("Analyzing collected evidence")
	finding, err := engine.Analyze(cmd.Context(), *preamble, *bundle)
	if err != nil {
		return fmt.Errorf("failed to analyze evidence: %w", err)
	}

	// Step 10: Flag low confidence if needed
	confidenceThreshold := preamble.Rubrics.ConfidenceThreshold
	analyze.FlagLowConfidence(finding, confidenceThreshold)

	// Set mode to autonomous
	finding.Mode = "autonomous"

	// Step 11: Export finding
	if err := exportFinding(finding, outputFile); err != nil {
		return fmt.Errorf("failed to export finding: %w", err)
	}

	// Step 12: Display summary
	duration := time.Since(startTime)
	fmt.Println("\n✓ Autonomous evidence collection and analysis complete!")
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Framework:     %s %s\n", finding.FrameworkID, finding.ControlID)
	fmt.Printf("  Confidence:    %.1f%% (%s)\n", finding.ConfidenceScore*100, confidenceLevel(finding.ConfidenceScore))
	fmt.Printf("  Severity:      %s\n", finding.Severity)
	fmt.Printf("  Review:        %v\n", finding.ReviewRequired)
	fmt.Printf("  Evidence:      %d events collected\n", len(bundle.Events))
	fmt.Printf("  Plan Items:    %d approved, %d executed\n", countApproved(plan), countExecuted(plan))
	fmt.Printf("  Duration:      %s\n", duration.Round(time.Millisecond))
	fmt.Printf("  Output:        %s\n", outputFile)

	if finding.ReviewRequired {
		fmt.Println("\n⚠ Low confidence detected - manual review recommended")
	}

	fmt.Println("\nNext steps:")
	fmt.Println("  - Review the finding in", outputFile)
	fmt.Println("  - Run 'sdek report' to generate compliance report")
	fmt.Println("  - Run 'sdek tui' to explore evidence interactively")

	return nil
}

// Helper functions

func countAutoApproved(plan *types.EvidencePlan) int {
	count := 0
	for _, item := range plan.Items {
		if item.ApprovalStatus == types.ApprovalAutoApproved {
			count++
		}
	}
	return count
}

func countApproved(plan *types.EvidencePlan) int {
	count := 0
	for _, item := range plan.Items {
		if item.ApprovalStatus == types.ApprovalApproved ||
			item.ApprovalStatus == types.ApprovalAutoApproved {
			count++
		}
	}
	return count
}

func countExecuted(plan *types.EvidencePlan) int {
	count := 0
	for _, item := range plan.Items {
		if item.ExecutionStatus == types.ExecComplete {
			count++
		}
	}
	return count
}

func confidenceLevel(score float64) string {
	if score >= 0.8 {
		return "high"
	} else if score >= 0.6 {
		return "medium"
	}
	return "low"
}
