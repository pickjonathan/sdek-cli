package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pickjonathan/sdek-cli/pkg/types"
)

// PlanApprovalModel displays an evidence collection plan for user approval
// Part of Feature 003: AI Context Injection (Autonomous Mode)
type PlanApprovalModel struct {
	plan          *types.EvidencePlan
	width         int
	height        int
	selectedIndex int // Currently selected plan item
	confirmed     bool
	cancelled     bool
}

// NewPlanApproval creates a new plan approval component
func NewPlanApproval(plan *types.EvidencePlan) PlanApprovalModel {
	return PlanApprovalModel{
		plan:          plan,
		selectedIndex: 0,
		confirmed:     false,
		cancelled:     false,
	}
}

// Init initializes the plan approval
func (m PlanApprovalModel) Init() tea.Cmd {
	return nil
}

// Confirmed returns true if user confirmed the plan
func (m PlanApprovalModel) Confirmed() bool {
	return m.confirmed
}

// Cancelled returns true if user cancelled
func (m PlanApprovalModel) Cancelled() bool {
	return m.cancelled
}

// GetPlan returns the updated plan with approval status
func (m PlanApprovalModel) GetPlan() *types.EvidencePlan {
	return m.plan
}

// Update handles messages
func (m PlanApprovalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case "down", "j":
			if m.selectedIndex < len(m.plan.Items)-1 {
				m.selectedIndex++
			}
		case "space":
			// Toggle approval for selected item
			if m.selectedIndex < len(m.plan.Items) {
				item := &m.plan.Items[m.selectedIndex]
				if item.ApprovalStatus == types.ApprovalApproved || item.ApprovalStatus == types.ApprovalAutoApproved {
					item.ApprovalStatus = types.ApprovalPending
				} else {
					item.ApprovalStatus = types.ApprovalApproved
				}
			}
		case "a":
			// Approve all items
			for i := range m.plan.Items {
				m.plan.Items[i].ApprovalStatus = types.ApprovalApproved
			}
		case "r":
			// Reject all items
			for i := range m.plan.Items {
				m.plan.Items[i].ApprovalStatus = types.ApprovalDenied
			}
		case "enter", "y":
			// Confirm with current approvals
			m.confirmed = true
			return m, tea.Quit
		case "n", "q", "esc":
			// Cancel
			m.cancelled = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View renders the plan approval interface
func (m PlanApprovalModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		MarginBottom(1)

	b.WriteString(titleStyle.Render("📋 Evidence Collection Plan"))
	b.WriteString("\n\n")

	// Plan metadata
	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("69"))

	b.WriteString(labelStyle.Render("Framework: "))
	b.WriteString(fmt.Sprintf("%s\n", m.plan.Framework))

	b.WriteString(labelStyle.Render("Section:   "))
	b.WriteString(fmt.Sprintf("%s\n", m.plan.Section))

	b.WriteString(labelStyle.Render("Status:    "))
	b.WriteString(fmt.Sprintf("%s\n", m.plan.Status))

	// Budget info
	approved := m.countApproved()
	totalEvents := m.calculateTotalEvents()

	b.WriteString(labelStyle.Render("Budget:    "))
	b.WriteString(fmt.Sprintf("%d items approved, ~%d events total\n\n", approved, totalEvents))

	// Plan items
	b.WriteString(labelStyle.Render("Plan Items:"))
	b.WriteString("\n\n")

	for i, item := range m.plan.Items {
		m.renderPlanItem(&b, i, item)
	}

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true).
		MarginTop(1)

	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("↑/↓ Navigate  SPACE Toggle  A Approve All  R Reject All"))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("ENTER/Y Confirm  N/Q Cancel"))

	return b.String()
}

// renderPlanItem renders a single plan item
func (m PlanApprovalModel) renderPlanItem(b *strings.Builder, index int, item types.PlanItem) {
	// Selection indicator
	cursor := "  "
	if index == m.selectedIndex {
		cursor = "→ "
	}

	// Approval checkbox
	checkbox := "[ ]"
	checkboxStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	if item.ApprovalStatus == types.ApprovalApproved || item.ApprovalStatus == types.ApprovalAutoApproved {
		checkbox = "[✓]"
		checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	} else if item.ApprovalStatus == types.ApprovalDenied {
		checkbox = "[✗]"
		checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	}

	// Source badge
	sourceStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		Bold(true)

	// Item details
	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	estimateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Italic(true)

	b.WriteString(cursor)
	b.WriteString(checkboxStyle.Render(checkbox))
	b.WriteString(" ")
	b.WriteString(sourceStyle.Render(item.Source))
	b.WriteString(": ")
	b.WriteString(itemStyle.Render(item.Query))
	b.WriteString(" ")

	// Show signal strength as indicator
	signal := fmt.Sprintf("(%.0f%% relevant)", item.SignalStrength*100)
	b.WriteString(estimateStyle.Render(signal))
	b.WriteString("\n")
}

// countApproved counts how many items are approved
func (m PlanApprovalModel) countApproved() int {
	count := 0
	for _, item := range m.plan.Items {
		if item.ApprovalStatus == types.ApprovalApproved || item.ApprovalStatus == types.ApprovalAutoApproved {
			count++
		}
	}
	return count
}

// calculateTotalEvents returns estimated sources count for approved items
// (actual PlanItem doesn't have EstimatedEvents, using plan-level info)
func (m PlanApprovalModel) calculateTotalEvents() int {
	approvedCount := m.countApproved()
	if approvedCount == 0 {
		return 0
	}
	// Use plan-level estimates proportionally
	totalItems := len(m.plan.Items)
	if totalItems == 0 {
		return 0
	}
	return (m.plan.EstimatedSources * approvedCount) / totalItems
}

// SetSize updates the component dimensions
func (m PlanApprovalModel) SetSize(width, height int) PlanApprovalModel {
	m.width = width
	m.height = height
	return m
}
