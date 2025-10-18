package cmd

import (
	"github.com/spf13/cobra"
)

// aiCmd represents the ai parent command for AI-powered features (Feature 003)
var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI-powered analysis and evidence collection",
	Long: `AI-powered compliance analysis with context injection and autonomous evidence collection.

This command provides two modes:
- Context Mode: Analyze evidence against specific policy excerpts with AI context injection
- Autonomous Mode: Generate evidence collection plans and execute them automatically

These are distinct from the existing 'sdek analyze --ai' flag, which enhances
the standard analysis workflow. The 'sdek ai' commands provide specialized
AI workflows focused on policy context and autonomous collection.`,
	Example: `  # Analyze evidence with policy context injection
  sdek ai analyze --framework SOC2 --section CC6.1 \
      --excerpts-file policies.json \
      --evidence-path evidence/*.json

  # Generate and execute evidence collection plan
  sdek ai plan --framework ISO27001 --section A.9.4.2 \
      --excerpts-file policies.json

  # Dry-run mode (preview plan without execution)
  sdek ai plan --framework SOC2 --section CC6.1 \
      --excerpts-file policies.json --dry-run`,
}

func init() {
	rootCmd.AddCommand(aiCmd)
}
