package cmd

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/pickjonathan/sdek-cli/internal/ingest"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/spf13/cobra"
)

var (
	ingestSource string
	ingestEvents int
	ingestSeed   int64
)

// ingestCmd represents the ingest command
var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest events from a specific data source",
	Long: `Ingest simulated events from a specific data source (Git, Jira, Slack, CI/CD, or Docs).

The ingest command generates events from the specified source and adds them to
the application state. This is useful for:
- Adding more events to an existing dataset
- Testing individual source integrations
- Controlling the number of events per source

Events are generated deterministically when using the --seed flag, allowing
for reproducible test scenarios.`,
	Example: `  # Ingest 50 Git commit events
  sdek ingest --source git --events 50

  # Ingest Jira tickets with a specific seed
  sdek ingest --source jira --events 30 --seed 12345

  # Ingest CI/CD pipeline events
  sdek ingest --source cicd --events 40`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate source
		validSources := []string{"git", "jira", "slack", "cicd", "docs"}
		valid := false
		for _, s := range validSources {
			if ingestSource == s {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid source '%s', must be one of: git, jira, slack, cicd, docs", ingestSource)
		}

		// Validate event count
		if ingestEvents < 10 || ingestEvents > 50 {
			return fmt.Errorf("event count must be between 10 and 50, got %d", ingestEvents)
		}

		return nil
	},
	RunE: runIngest,
}

func init() {
	rootCmd.AddCommand(ingestCmd)

	ingestCmd.Flags().StringVar(&ingestSource, "source", "", "Data source to ingest from (git, jira, slack, cicd, docs) [required]")
	ingestCmd.Flags().IntVar(&ingestEvents, "events", 25, "Number of events to generate (10-50)")
	ingestCmd.Flags().Int64Var(&ingestSeed, "seed", time.Now().UnixNano(), "Random seed for reproducible generation")

	ingestCmd.MarkFlagRequired("source")
}

func runIngest(cmd *cobra.Command, args []string) error {
	slog.Info("Starting ingest command", "source", ingestSource, "events", ingestEvents, "seed", ingestSeed)

	// Load existing state
	state, err := store.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Generate events for the specified source
	slog.Info("Generating events", "source", ingestSource, "count", ingestEvents)
	
	var gen ingest.Generator
	var events []types.Event

	switch ingestSource {
	case "git":
		gen = ingest.NewGitGenerator(ingestSeed)
	case "jira":
		gen = ingest.NewJiraGenerator(ingestSeed)
	case "slack":
		gen = ingest.NewSlackGenerator(ingestSeed)
	case "cicd":
		gen = ingest.NewCICDGenerator(ingestSeed)
	case "docs":
		gen = ingest.NewDocsGenerator(ingestSeed)
	}

	events, err = gen.Generate(ingestSeed, ingestEvents)
	if err != nil {
		return fmt.Errorf("failed to generate events: %w", err)
	}

	// Add events to state
	state.Events = append(state.Events, events...)

	// Update source event count
	sourceFound := false
	for i := range state.Sources {
		if state.Sources[i].ID == ingestSource {
			state.Sources[i].EventCount += len(events)
			state.Sources[i].LastSync = time.Now()
			sourceFound = true
			break
		}
	}

	// If source doesn't exist, create it
	if !sourceFound {
		source := types.Source{
			ID:         ingestSource,
			Name:       fmt.Sprintf("%s Events", capitalize(ingestSource)),
			Type:       ingestSource,
			Status:     "simulated",
			LastSync:   time.Now(),
			EventCount: len(events),
			Enabled:    true,
		}
		state.Sources = append(state.Sources, source)
	}

	// Save updated state
	slog.Info("Saving state")
	if err := state.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Print summary
	fmt.Printf("✓ Ingested %d events from %s\n", len(events), ingestSource)
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Total events: %d\n", len(state.Events))
	fmt.Printf("  Source: %s\n", capitalize(ingestSource))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  - Run 'sdek analyze' to map new events to controls")
	fmt.Println("  - Run 'sdek tui' to explore the updated data")

	slog.Info("Ingest command completed successfully")
	return nil
}

// capitalize returns a string with the first letter capitalized
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
