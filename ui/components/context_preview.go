package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// ContextPreviewModel displays policy context before AI analysis
// Part of Feature 003: AI Context Injection
type ContextPreviewModel struct {
	preamble      types.ContextPreamble
	evidenceCount int
	width         int
	height        int
	confirmed     bool
	cancelled     bool
}

// NewContextPreview creates a new context preview component
func NewContextPreview(preamble types.ContextPreamble, evidenceCount int) ContextPreviewModel {
	return ContextPreviewModel{
		preamble:      preamble,
		evidenceCount: evidenceCount,
		confirmed:     false,
		cancelled:     false,
	}
}

// Init initializes the context preview
func (m ContextPreviewModel) Init() tea.Cmd {
	return nil
}

// Confirmed returns true if user confirmed to proceed
func (m ContextPreviewModel) Confirmed() bool {
	return m.confirmed
}

// Cancelled returns true if user cancelled
func (m ContextPreviewModel) Cancelled() bool {
	return m.cancelled
}

// Update handles messages
func (m ContextPreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "y":
			m.confirmed = true
			return m, tea.Quit
		case "n", "q", "esc":
			m.cancelled = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the context preview
func (m ContextPreviewModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("🔍 AI Context Preview"))
	b.WriteString("\n\n")

	// Framework info
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("69"))

	b.WriteString(labelStyle.Render("Framework: "))
	b.WriteString(fmt.Sprintf("%s %s\n", m.preamble.Framework, m.preamble.Version))

	b.WriteString(labelStyle.Render("Section:   "))
	b.WriteString(fmt.Sprintf("%s\n", m.preamble.Section))

	b.WriteString(labelStyle.Render("Evidence:  "))
	b.WriteString(fmt.Sprintf("%d events\n\n", m.evidenceCount))

	// Policy excerpt (truncate if too long)
	excerptStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1).
		Width(m.width - 4)

	excerpt := m.preamble.Excerpt
	maxLen := 500
	if len(excerpt) > maxLen {
		excerpt = excerpt[:maxLen] + "..."
	}

	b.WriteString(labelStyle.Render("Policy Excerpt:"))
	b.WriteString("\n")
	b.WriteString(excerptStyle.Render(excerpt))
	b.WriteString("\n\n")

	// Related control IDs (if any)
	if len(m.preamble.ControlIDs) > 0 {
		b.WriteString(labelStyle.Render("Related Controls: "))
		b.WriteString(strings.Join(m.preamble.ControlIDs, ", "))
		b.WriteString("\n\n")
	}

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true)

	proceedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")).
		Bold(true)

	cancelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	b.WriteString(instructionStyle.Render("This context will be sent to the AI provider for analysis."))
	b.WriteString("\n\n")
	b.WriteString(proceedStyle.Render("Press Enter or Y to proceed"))
	b.WriteString(infoStyle.Render("  |  "))
	b.WriteString(cancelStyle.Render("Press N or Q to cancel"))

	return b.String()
}

// SetSize updates the component dimensions
func (m ContextPreviewModel) SetSize(width, height int) ContextPreviewModel {
	m.width = width
	m.height = height
	return m
}
