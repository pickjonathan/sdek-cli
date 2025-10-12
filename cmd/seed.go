package cmd

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/analyze"
	"github.com/pickjonathan/sdek-cli/internal/ingest"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/spf13/cobra"
)

var (
	seedDemo  bool
	seedValue int64
	seedReset bool
)

// seedCmd represents the seed command
var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Generate demo data for testing and development",
	Long: `Generate simulated compliance data including sources, events, frameworks,
controls, evidence, and findings.

The seed command creates realistic demo data that can be used for:
- Testing the CLI and TUI functionality
- Demonstrating compliance evidence mapping
- Developing and debugging features

By default, it generates:
- 5 data sources (Git, Jira, Slack, CI/CD, Docs)
- 10-50 events per source (~130 total events)
- 3 compliance frameworks (SOC2, ISO 27001, PCI DSS)
- 120 controls across all frameworks
- ~245 evidence mappings
- ~18 findings (risks identified)`,
	Example: `  # Generate demo data
  sdek seed --demo

  # Generate data with specific seed for reproducibility
  sdek seed --demo --seed 12345

  # Reset state and generate fresh data
  sdek seed --demo --reset`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !seedDemo {
			return fmt.Errorf("--demo flag is required")
		}
		return nil
	},
	RunE: runSeed,
}

func init() {
	rootCmd.AddCommand(seedCmd)

	seedCmd.Flags().BoolVar(&seedDemo, "demo", false, "Generate demo data (required)")
	seedCmd.Flags().Int64Var(&seedValue, "seed", time.Now().UnixNano(), "Random seed for reproducible data generation")
	seedCmd.Flags().BoolVar(&seedReset, "reset", false, "Reset state before generating new data")
}

func runSeed(cmd *cobra.Command, args []string) error {
	slog.Info("Starting seed command", "demo", seedDemo, "seed", seedValue, "reset", seedReset)

	// Load or create state
	var state *store.State
	var err error

	if seedReset {
		slog.Info("Resetting state")
		state = store.NewState()
	} else {
		state, err = store.Load()
		if err != nil {
			slog.Warn("Could not load existing state, creating new", "error", err)
			state = store.NewState()
		}
	}

	// Generate sources
	slog.Info("Generating sources")
	sources := []types.Source{
		{
			ID:         "git",
			Name:       "Git Commits",
			Type:       "git",
			Status:     "simulated",
			LastSync:   time.Now(),
			EventCount: 0,
			Enabled:    true,
		},
		{
			ID:         "jira",
			Name:       "Jira Tickets",
			Type:       "jira",
			Status:     "simulated",
			LastSync:   time.Now(),
			EventCount: 0,
			Enabled:    true,
		},
		{
			ID:         "slack",
			Name:       "Slack Messages",
			Type:       "slack",
			Status:     "simulated",
			LastSync:   time.Now(),
			EventCount: 0,
			Enabled:    true,
		},
		{
			ID:         "cicd",
			Name:       "CI/CD Pipelines",
			Type:       "cicd",
			Status:     "simulated",
			LastSync:   time.Now(),
			EventCount: 0,
			Enabled:    true,
		},
		{
			ID:         "docs",
			Name:       "Documentation",
			Type:       "docs",
			Status:     "simulated",
			LastSync:   time.Now(),
			EventCount: 0,
			Enabled:    true,
		},
	}
	state.Sources = sources

	// Generate events for each source
	slog.Info("Generating events")
	var allEvents []types.Event

	for _, source := range sources {
		var events []types.Event
		var gen ingest.Generator
		
		switch source.Type {
		case "git":
			gen = ingest.NewGitGenerator(seedValue)
			events, err = gen.Generate(seedValue, 25)
		case "jira":
			gen = ingest.NewJiraGenerator(seedValue)
			events, err = gen.Generate(seedValue, 30)
		case "slack":
			gen = ingest.NewSlackGenerator(seedValue)
			events, err = gen.Generate(seedValue, 20)
		case "cicd":
			gen = ingest.NewCICDGenerator(seedValue)
			events, err = gen.Generate(seedValue, 30)
		case "docs":
			gen = ingest.NewDocsGenerator(seedValue)
			events, err = gen.Generate(seedValue, 25)
		}
		
		if err != nil {
			return fmt.Errorf("failed to generate events for %s: %w", source.Name, err)
		}

		// Update source event count
		for i := range state.Sources {
			if state.Sources[i].ID == source.ID {
				state.Sources[i].EventCount = len(events)
				break
			}
		}

		allEvents = append(allEvents, events...)
		slog.Info("Generated events for source", "source", source.Name, "count", len(events))
	}
	state.Events = allEvents

	// Initialize frameworks with controls
	slog.Info("Initializing compliance frameworks")
	frameworks, controls := initializeFrameworks()
	state.Frameworks = frameworks
	state.Controls = controls

	// Map events to controls (evidence generation)
	slog.Info("Mapping events to controls")
	mapper := analyze.NewMapper()
	evidence := mapper.MapEventsToControls(allEvents)
	state.Evidence = evidence

	// Calculate risk scores and update control statuses
	slog.Info("Calculating risk scores")
	scorer := analyze.NewRiskScorer()
	
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
	for i := range state.Frameworks {
		framework := &state.Frameworks[i]
		frameworkControls := filterControlsByFramework(state.Controls, framework.ID)
		greenCount := 0
		for _, ctrl := range frameworkControls {
			if ctrl.RiskStatus == types.RiskStatusGreen {
				greenCount++
			}
		}
		if len(frameworkControls) > 0 {
			framework.CompliancePercentage = float64(greenCount) / float64(len(frameworkControls)) * 100
		}
	}

	// Save state
	slog.Info("Saving state")
	if err := state.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Print summary
	fmt.Println("✓ Demo data generated successfully!")
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Sources:    %d\n", len(state.Sources))
	fmt.Printf("  Events:     %d\n", len(state.Events))
	fmt.Printf("  Frameworks: %d\n", len(state.Frameworks))
	fmt.Printf("  Controls:   %d\n", len(state.Controls))
	fmt.Printf("  Evidence:   %d\n", len(state.Evidence))
	fmt.Printf("  Findings:   %d\n", len(state.Findings))
	fmt.Println()
	fmt.Println("Compliance Status:")
	for _, fw := range state.Frameworks {
		fmt.Printf("  %-15s %.1f%% compliant\n", fw.Name, fw.CompliancePercentage)
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  - Run 'sdek tui' to explore the data interactively")
	fmt.Println("  - Run 'sdek report' to export a compliance report")
	fmt.Println("  - Run 'sdek analyze' to recalculate evidence mapping")

	slog.Info("Seed command completed successfully")
	return nil
}

// initializeFrameworks creates frameworks and controls from definitions
func initializeFrameworks() ([]types.Framework, []types.Control) {
	var frameworks []types.Framework
	var controls []types.Control

	// Get framework definitions
	defs := analyze.GetFrameworkDefinitions()

	for fwID, def := range defs {
		// Create framework
		framework := types.Framework{
			ID:                   fwID,
			Name:                 def.Name,
			Version:              "1.0",
			ControlCount:         len(def.Controls),
			CompliancePercentage: 0.0,
			Description:          def.Description,
			Category:             "compliance",
		}
		frameworks = append(frameworks, framework)

		// Create controls for this framework
		for _, ctrlDef := range def.Controls {
			control := types.Control{
				ID:               ctrlDef.ID,
				FrameworkID:      fwID,
				Title:            ctrlDef.Title,
				Description:      ctrlDef.Description,
				Category:         ctrlDef.Category,
				RiskStatus:       types.RiskStatusYellow, // Default to yellow until calculated
				RiskScore:        0.0,
				EvidenceCount:    0,
				ConfidenceLevel:  0.0,
				Keywords:         ctrlDef.Keywords,
				RequiredEvidence: 3,
			}
			controls = append(controls, control)
		}
	}

	return frameworks, controls
}

// filterControlsByFramework returns controls belonging to a specific framework
func filterControlsByFramework(controls []types.Control, frameworkID string) []types.Control {
	var result []types.Control
	for _, ctrl := range controls {
		if ctrl.FrameworkID == frameworkID {
			result = append(result, ctrl)
		}
	}
	return result
}

// filterEvidenceByControl returns evidence linked to a specific control
func filterEvidenceByControl(evidence []types.Evidence, controlID string) []types.Evidence {
	var result []types.Evidence
	for _, ev := range evidence {
		if ev.ControlID == controlID {
			result = append(result, ev)
		}
	}
	return result
}
