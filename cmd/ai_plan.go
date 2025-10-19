package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pickjonathan/sdek-cli/internal/ai"
	"github.com/pickjonathan/sdek-cli/internal/analyze"
	"github.com/pickjonathan/sdek-cli/internal/mcp"
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
		if provider != "openai" && provider != "anthropic" {
			return fmt.Errorf("unsupported AI provider: %s (must be 'openai' or 'anthropic')", provider)
		}

		// Check API key is configured
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

	// Step 4: Initialize MCP registry and show available tools
	slog.Info("Initializing MCP registry")
	mcpRegistry, err := initializeMCPRegistry(cmd.Context(), cfg)
	if err != nil {
		slog.Warn("MCP registry initialization failed", "error", err)
		mcpRegistry = nil // Ensure it's nil on error
	}

	// Display available MCP tools
	if mcpRegistry != nil {
		tools, err := mcpRegistry.List(cmd.Context())
		if err != nil {
			slog.Warn("Failed to list MCP tools", "error", err)
		} else {
			slog.Info("MCP tools registered", "count", len(tools))
			for _, tool := range tools {
				slog.Info("MCP tool available",
					"name", tool.Name,
					"status", tool.Status,
					"enabled", tool.Enabled)
			}
		}
	}

	// Step 4a: Initialize AI engine with MCP connector
	slog.Info("Initializing AI engine", "provider", cfg.AI.Provider)

	var engine ai.Engine
	if mcpRegistry != nil {
		// Create pass-through enforcer for autonomous mode (no RBAC restrictions)
		enforcer := &passThroughEnforcer{}
		redactor := &passThroughRedactor{}

		// Create MCP invoker
		mcpInvoker := mcp.NewInvoker(
			mcpRegistry,
			enforcer,
			redactor,
			nil, // cache - will be skipped if nil
		)

		// Create MCP adapter that implements ai.MCPConnector interface
		mcpConnector := &mcpRegistryAdapter{
			registry: mcpRegistry,
			invoker:  mcpInvoker,
		}

		slog.Info("Using new MCP registry for evidence collection (Feature 004)")

		// Create AI provider configuration
		aiConfig := ai.AIConfig{
			Provider:     cfg.AI.Provider,
			Enabled:      cfg.AI.Enabled,
			Model:        cfg.AI.Model,
			MaxTokens:    4096,
			Temperature:  0.3,
			Timeout:      cfg.AI.Timeout,
			RateLimit:    cfg.AI.RateLimit,
			OpenAIKey:    cfg.AI.OpenAIKey,
			AnthropicKey: cfg.AI.AnthropicKey,
		}

		// Override with APIKey if provider-specific keys are empty
		if cfg.AI.Provider == types.AIProviderOpenAI && aiConfig.OpenAIKey == "" {
			aiConfig.OpenAIKey = cfg.AI.APIKey
		}
		if cfg.AI.Provider == types.AIProviderAnthropic && aiConfig.AnthropicKey == "" {
			aiConfig.AnthropicKey = cfg.AI.APIKey
		}

		// Create provider
		provider, err := ai.CreateProviderFromRegistry(cfg.AI.Provider, aiConfig)
		if err != nil {
			return fmt.Errorf("failed to create AI provider: %w", err)
		}

		// Create engine with MCP connector
		engine = ai.NewEngineWithConnector(cfg, provider, mcpConnector)
	} else {
		// Fall back to legacy connector system
		slog.Warn("MCP registry unavailable, using legacy connector system")
		engine, err = ai.NewEngineFromConfig(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize AI engine: %w", err)
		}
	}

	// Log connector status
	if len(cfg.AI.Connectors) > 0 {
		enabledConnectors := []string{}
		for name, conn := range cfg.AI.Connectors {
			if conn.Enabled {
				enabledConnectors = append(enabledConnectors, name)
			}
		}
		if len(enabledConnectors) > 0 && mcpRegistry == nil {
			slog.Info("Legacy connectors enabled", "connectors", enabledConnectors)
			slog.Warn("Note: Legacy connector system - only GitHub implemented in connectors")
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
		plan.Status = types.PlanApproved // Mark the overall plan as approved
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

// initializeMCPRegistry initializes the MCP registry if MCP is enabled
func initializeMCPRegistry(ctx context.Context, cfg *types.Config) (*mcp.Registry, error) {
	// Check if MCP feature is enabled (via viper config)
	if !viper.GetBool("mcp.enabled") {
		slog.Debug("MCP is disabled in config")
		return nil, nil
	}

	// Create registry
	registry := mcp.NewRegistry()

	// Initialize registry (it will discover configs from standard paths)
	_, err := registry.Init(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MCP registry: %w", err)
	}

	return registry, nil
}

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

// mcpRegistryAdapter adapts the new MCP registry (Feature 004) to the legacy MCPConnector interface.
// This allows ExecutePlan to use MCP tools while maintaining backward compatibility.
type mcpRegistryAdapter struct {
	registry *mcp.Registry
	invoker  mcp.AgentInvoker
}

// Collect fetches evidence from an MCP tool based on the source and query.
func (m *mcpRegistryAdapter) Collect(ctx context.Context, source string, query string) ([]types.EvidenceEvent, error) {
	slog.Debug("MCP adapter collecting evidence", "source", source, "query", query)

	// Map source to MCP tool name
	toolName, method := mapSourceToMCPTool(source)
	if toolName == "" {
		return nil, fmt.Errorf("no MCP tool available for source: %s", source)
	}

	// Check if tool is available and ready
	tool, err := m.registry.Get(ctx, toolName)
	if err != nil {
		return nil, fmt.Errorf("MCP tool %s not found: %w", toolName, err)
	}

	if tool.Status != "ready" {
		return nil, fmt.Errorf("MCP tool %s is not ready (status: %s)", toolName, tool.Status)
	}

	slog.Info("Using MCP tool for evidence collection", "tool", toolName, "method", method, "source", source)

	// Parse query into MCP tool arguments
	args, err := parseQueryToArgs(source, query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	slog.Debug("Invoking MCP tool", "tool", toolName, "method", method, "args", args)

	// MCP protocol requires tools/call method with name and arguments
	mcpParams := map[string]interface{}{
		"name":      method, // "call_aws" or "suggest_aws_commands"
		"arguments": args,   // The actual tool arguments
	}

	// Invoke MCP tool using tools/call method
	evidence, err := m.invoker.InvokeTool(ctx, "autonomous-agent", toolName, "tools/call", mcpParams)
	if err != nil {
		slog.Error("MCP tool invocation failed", "tool", toolName, "error", err)
		return nil, fmt.Errorf("MCP tool invocation failed: %w", err)
	}

	slog.Info("MCP tool invocation successful", "tool", toolName, "evidence_id", evidence.ID)

	// Log the MCP evidence details for debugging
	if slog.Default().Enabled(ctx, slog.LevelDebug) {
		slog.Debug("MCP evidence details",
			"tool", toolName,
			"evidence_id", evidence.ID,
			"reasoning", evidence.Reasoning,
			"keywords", evidence.Keywords,
			"confidence_score", evidence.ConfidenceScore,
			"analysis_method", evidence.AnalysisMethod)
	}

	// Convert Evidence to EvidenceEvent
	events := convertEvidenceToEvents(evidence, source, query)
	return events, nil
}

// mapSourceToMCPTool maps plan item source names to MCP tool names and methods.
func mapSourceToMCPTool(source string) (toolName string, method string) {
	// Check if source contains AWS-related keywords
	if contains(source, "AWS") || contains(source, "CloudTrail") || contains(source, "IAM") || contains(source, "EC2") || contains(source, "S3") {
		return "aws-api", "tools/call"
	}

	// Check for GitHub
	if contains(source, "Github") || contains(source, "GitHub") || contains(source, "GITHUB") {
		return "github-api", "tools/call"
	}

	// Check for Jira
	if contains(source, "Jira") || contains(source, "JIRA") {
		return "jira-api", "tools/call"
	}

	// Check for Slack
	if contains(source, "Slack") || contains(source, "SLACK") {
		return "slack-api", "tools/call"
	}

	return "", ""
} // parseQueryToArgs converts a natural language query into MCP tool arguments.
func parseQueryToArgs(source, query string) (map[string]interface{}, error) {
	// Check if source is AWS-related
	if contains(source, "AWS") || contains(source, "CloudTrail") || contains(source, "IAM") || contains(source, "EC2") || contains(source, "S3") {
		return parseAWSQuery(query)
	}

	if contains(source, "Github") || contains(source, "GitHub") {
		return parseGitHubQuery(query)
	}

	if contains(source, "Jira") {
		return parseJiraQuery(query)
	}

	if contains(source, "Slack") {
		return parseSlackQuery(query)
	}

	return nil, fmt.Errorf("unsupported source: %s", source)
}

// parseAWSQuery converts a natural language AWS query into AWS CLI command arguments.
func parseAWSQuery(query string) (map[string]interface{}, error) {
	// Simple keyword-based parsing - can be enhanced with AI later
	queryLower := query

	var cliCommand string

	// Detect CloudTrail queries
	if contains(queryLower, "cloudtrail") || contains(queryLower, "CloudTrail") {
		cliCommand = "aws cloudtrail lookup-events --max-results 50"
	} else if contains(queryLower, "iam") || contains(queryLower, "IAM") {
		// Detect IAM queries
		if contains(queryLower, "user") {
			cliCommand = "aws iam list-users"
		} else if contains(queryLower, "policy") || contains(queryLower, "policies") {
			cliCommand = "aws iam list-policies --scope Local --max-items 50"
		} else {
			cliCommand = "aws iam list-users"
		}
	} else {
		// Default: return simple command
		cliCommand = "aws sts get-caller-identity"
	}

	// Return in MCP tools/call format
	return map[string]interface{}{
		"name": "call_aws",
		"arguments": map[string]interface{}{
			"cli_command": cliCommand,
		},
	}, nil
} // parseGitHubQuery converts GitHub queries (placeholder).
func parseGitHubQuery(query string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"query": query,
	}, nil
}

// parseJiraQuery converts Jira queries (placeholder).
func parseJiraQuery(query string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"jql": query,
	}, nil
}

// parseSlackQuery converts Slack queries (placeholder).
func parseSlackQuery(query string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"query": query,
	}, nil
}

// convertEvidenceToEvents converts an MCP Evidence to EvidenceEvents.
func convertEvidenceToEvents(evidence *types.Evidence, source, query string) []types.EvidenceEvent {
	if evidence == nil {
		return []types.EvidenceEvent{}
	}

	event := types.EvidenceEvent{
		ID:        evidence.ID,
		Source:    source,
		Type:      "mcp-evidence",
		Timestamp: evidence.MappedAt,
		Content:   evidence.Reasoning,
		Metadata: map[string]interface{}{
			"evidence_id":          evidence.ID,
			"confidence_score":     evidence.ConfidenceScore,
			"confidence_level":     evidence.ConfidenceLevel,
			"analysis_method":      evidence.AnalysisMethod,
			"keywords":             evidence.Keywords,
			"ai_analyzed":          evidence.AIAnalyzed,
			"heuristic_confidence": evidence.HeuristicConfidence,
			"combined_confidence":  evidence.CombinedConfidence,
			"query":                query,
		},
	}

	return []types.EvidenceEvent{event}
}

// contains is a case-insensitive string contains check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || len(s) > len(substr)+1 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// passThroughEnforcer is a no-op enforcer that allows all operations.
// Used for autonomous mode where RBAC restrictions are not needed.
type passThroughEnforcer struct{}

func (p *passThroughEnforcer) CheckPermission(ctx context.Context, agentRole string, capability string) (bool, error) {
	return true, nil // Always allow
}

func (p *passThroughEnforcer) GetCapabilities(ctx context.Context, agentRole string) ([]types.AgentCapability, error) {
	return []types.AgentCapability{}, nil
}

func (p *passThroughEnforcer) ApplyBudget(ctx context.Context, toolName string, budget *types.ToolBudget) error {
	return nil // No budget enforcement
}

func (p *passThroughEnforcer) RecordInvocation(ctx context.Context, log *types.MCPInvocationLog) error {
	return nil // No logging
}

// passThroughRedactor is a no-op redactor that doesn't redact anything.
type passThroughRedactor struct{}

func (p *passThroughRedactor) Redact(text string) (string, *types.RedactionMap, error) {
	return text, &types.RedactionMap{}, nil
}
