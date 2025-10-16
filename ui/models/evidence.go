package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pickjonathan/sdek-cli/internal/store"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/pickjonathan/sdek-cli/ui/styles"
)

// EvidenceModel represents the evidence list screen
type EvidenceModel struct {
	state     *store.State
	width     int
	height    int
	cursor    int
	controlID string
}

// NewEvidenceModel creates a new evidence model
func NewEvidenceModel(state *store.State, controlID string) EvidenceModel {
	return EvidenceModel{
		state:     state,
		controlID: controlID,
		cursor:    0,
	}
}

// Init initializes the evidence model
func (m EvidenceModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m EvidenceModel) Update(msg tea.Msg) (EvidenceModel, tea.Cmd) {
	evidence := m.getFilteredEvidence()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(evidence)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

// View renders the evidence screen
func (m EvidenceModel) View() string {
	s := styles.TitleStyle.Render("Evidence Mappings") + "\n\n"

	evidence := m.getFilteredEvidence()

	if len(evidence) == 0 {
		s += styles.SubtleStyle.Render("No evidence found. Run 'sdek analyze' to generate evidence mappings.")
		return s
	}

	// Group by confidence
	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, ev := range evidence {
		switch ev.ConfidenceLevel {
		case "high", "High":
			highCount++
		case "medium", "Medium":
			mediumCount++
		case "low", "Low":
			lowCount++
		}
	}

	s += fmt.Sprintf("Confidence: %s %d High  %s %d Medium  %s %d Low\n\n",
		styles.StatusGreenStyle.Render("●"),
		highCount,
		styles.StatusYellowStyle.Render("●"),
		mediumCount,
		styles.StatusRedStyle.Render("●"),
		lowCount,
	)

	// Show evidence
	displayCount := 0
	maxDisplay := m.height - 10
	if maxDisplay < 5 {
		maxDisplay = 5
	}

	for i, ev := range evidence {
		if displayCount >= maxDisplay {
			s += styles.SubtleStyle.Render(fmt.Sprintf("\n... and %d more evidence items", len(evidence)-displayCount))
			break
		}

		cursor := " "
		confidenceIcon := m.getConfidenceIcon(ev.ConfidenceLevel)
		line := fmt.Sprintf("%s %s %s → %s",
			cursor,
			confidenceIcon,
			ev.EventID[:8]+"...",
			ev.ControlID,
		)

		if m.cursor == i {
			cursor = ">"
			s += styles.SelectedListItemStyle.Render(line) + "\n"
		} else {
			s += styles.ListItemStyle.Render(line) + "\n"
		}
		displayCount++
	}

	s += "\n" + styles.SubtleStyle.Render("↑/↓: Navigate | Esc: Back | q: Quit")

	return s
}

// getFilteredEvidence returns evidence filtered by control if set
func (m EvidenceModel) getFilteredEvidence() []types.Evidence {
	if m.controlID == "" {
		// Return all evidence
		evidence := make([]types.Evidence, len(m.state.Evidence))
		copy(evidence, m.state.Evidence)
		return evidence
	}

	// Filter by control
	var filtered []types.Evidence
	for _, ev := range m.state.Evidence {
		if ev.ControlID == m.controlID {
			filtered = append(filtered, ev)
		}
	}
	return filtered
}

// getConfidenceIcon returns an icon for the confidence level
func (m EvidenceModel) getConfidenceIcon(confidence string) string {
	switch confidence {
	case "high", "High":
		return styles.StatusGreenStyle.Render("✓")
	case "medium", "Medium":
		return styles.StatusYellowStyle.Render("○")
	case "low", "Low":
		return styles.StatusRedStyle.Render("△")
	default:
		return styles.SubtleStyle.Render("?")
	}
}

// SetSize updates the model dimensions
func (m *EvidenceModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
