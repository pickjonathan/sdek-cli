package cmd

import (
	"fmt"
	"log/slog"

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

func init() {
	rootCmd.AddCommand(analyzeCmd)
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
	mapper := analyze.NewMapper()
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
