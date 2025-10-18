package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/spf13/cobra"
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
	Example: `  # Analyze evidence for SOC2 CC6.1 (Access Controls)
  sdek ai analyze --framework SOC2 --section CC6.1 \
      --excerpts-file ./policies/soc2_excerpts.json \
      --evidence-path ./evidence/github_*.json \
      --evidence-path ./evidence/jira_*.json

  # Bypass cache for fresh analysis
  sdek ai analyze --framework SOC2 --section CC6.1 \
      --excerpts-file ./policies/soc2_excerpts.json \
      --evidence-path ./evidence/*.json \
      --no-cache

  # Specify custom output file
  sdek ai analyze --framework ISO27001 --section A.9.4.2 \
      --excerpts-file ./policies/iso_excerpts.json \
      --evidence-path ./evidence/*.json \
      --output ./findings/iso_a942_finding.json`,
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

		// Step 3: Build ContextPreamble (validate it works)
		slog.Info("Building context preamble", "framework", framework, "section", section)
		_, err = types.NewContextPreamble(
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

		// TODO: Full implementation requires:
		// 1. Load config and check AI settings
		// 2. Initialize AI Engine with real provider
		// 3. Call Engine.Analyze()
		// 4. Flag low confidence
		// 5. Export finding to output file
		// 6. Display summary
		//
		// For now, just validate inputs and data loading works
		fmt.Println("✓ Command validation successful!")
		fmt.Printf("  Framework: %s %s\n", framework, excerpt.Version)
		fmt.Printf("  Section: %s\n", section)
		fmt.Printf("  Excerpt length: %d chars\n", len(excerpt.Text))
		fmt.Printf("  Evidence events: %d\n", len(evidence.Events))
		fmt.Println("\nNote: Full AI analysis implementation pending (requires provider setup)")

		return nil
	},
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
