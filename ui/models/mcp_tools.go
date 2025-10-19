package models

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pickjonathan/sdek-cli/internal/mcp"
	"github.com/pickjonathan/sdek-cli/pkg/types"
	"github.com/pickjonathan/sdek-cli/ui/styles"
)

// MCPToolsModel represents the MCP tools management screen
type MCPToolsModel struct {
	registry      *mcp.Registry
	tools         []types.MCPTool
	selectedIndex int
	width         int
	height        int
	loading       bool
	error         error
	lastRefresh   time.Time
}

// NewMCPToolsModel creates a new MCP tools model
func NewMCPToolsModel() MCPToolsModel {
	return MCPToolsModel{
		registry:      mcp.NewRegistry(),
		tools:         []types.MCPTool{},
		selectedIndex: 0,
		loading:       false,
	}
}

// mcpToolsLoadedMsg is sent when tools are loaded
type mcpToolsLoadedMsg struct {
	tools []types.MCPTool
	err   error
}

// mcpTestResultMsg is sent when a tool test completes
type mcpTestResultMsg struct {
	toolName string
	report   *types.MCPHealthReport
	err      error
}

// Init initializes the MCP tools model
func (m MCPToolsModel) Init() tea.Cmd {
	return m.loadTools()
}

// loadTools loads the list of MCP tools
func (m MCPToolsModel) loadTools() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Initialize registry
		if _, err := m.registry.Init(ctx); err != nil {
			return mcpToolsLoadedMsg{err: err}
		}

		// List tools
		tools, err := m.registry.List(ctx)
		if err != nil {
			return mcpToolsLoadedMsg{err: err}
		}

		return mcpToolsLoadedMsg{tools: tools}
	}
}

// testTool runs a health check on a specific tool
func (m MCPToolsModel) testTool(toolName string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		report, err := m.registry.Test(ctx, toolName)
		var reportPtr *types.MCPHealthReport
		if err == nil {
			reportPtr = &report
		}
		return mcpTestResultMsg{
			toolName: toolName,
			report:   reportPtr,
			err:      err,
		}
	}
}

// Update handles messages
func (m MCPToolsModel) Update(msg tea.Msg) (MCPToolsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case "down", "j":
			if m.selectedIndex < len(m.tools)-1 {
				m.selectedIndex++
			}
		case "t":
			// Test selected tool
			if len(m.tools) > 0 && m.selectedIndex < len(m.tools) {
				tool := m.tools[m.selectedIndex]
				m.loading = true
				return m, m.testTool(tool.Name)
			}
		case "r":
			// Refresh tools list
			m.loading = true
			return m, m.loadTools()
		}

	case mcpToolsLoadedMsg:
		m.loading = false
		m.lastRefresh = time.Now()
		if msg.err != nil {
			m.error = msg.err
		} else {
			m.tools = msg.tools
			m.error = nil
			// Reset selection if out of bounds
			if m.selectedIndex >= len(m.tools) {
				m.selectedIndex = 0
			}
		}

	case mcpTestResultMsg:
		m.loading = false
		if msg.err != nil {
			m.error = fmt.Errorf("test failed for %s: %w", msg.toolName, msg.err)
		} else {
			// Update tool with health report
			for i, tool := range m.tools {
				if tool.Name == msg.toolName {
					m.tools[i].LastHealthCheck = msg.report.Timestamp
					m.tools[i].Status = msg.report.Status
					m.tools[i].LastError = msg.report.LastError
					break
				}
			}
			m.error = nil
		}
	}

	return m, nil
}

// View renders the MCP tools screen
func (m MCPToolsModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var s string

	// Header
	title := styles.TitleStyle.Render("MCP Tools")
	subtitle := styles.SubtitleStyle.Render("Manage Model Context Protocol integrations")
	s += title + "\n" + subtitle + "\n\n"

	// Error display
	if m.error != nil {
		errorBox := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.RedColor)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(styles.RedColor)).
			Padding(1, 2).
			Width(m.width - 4).
			Render("⚠ Error: " + m.error.Error())
		s += errorBox + "\n\n"
	}

	// Loading indicator
	if m.loading {
		s += styles.SubtleStyle.Render("⏳ Loading...") + "\n\n"
	}

	// Tools list
	if len(m.tools) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.SubtleColor)).
			Italic(true).
			Render("No MCP tools configured. Add configs to ~/.sdek/mcp/")
		s += emptyMsg + "\n\n"
	} else {
		s += m.renderToolsTable()
	}

	// Last refresh time
	if !m.lastRefresh.IsZero() {
		refreshMsg := fmt.Sprintf("Last refreshed: %s", formatRelativeTime(m.lastRefresh))
		s += "\n" + styles.SubtleStyle.Render(refreshMsg) + "\n"
	}

	// Shortcuts
	s += "\n" + styles.SubtleStyle.Render("↑/↓: Navigate | t: Test Tool | r: Refresh | Esc: Back")

	return s
}

// renderToolsTable renders the tools table
func (m MCPToolsModel) renderToolsTable() string {
	var s string

	// Table header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(styles.PrimaryColor))

	nameWidth := 20
	statusWidth := 12
	latencyWidth := 10
	capWidth := 15
	checkWidth := 20

	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
		nameWidth, "NAME",
		statusWidth, "STATUS",
		latencyWidth, "LATENCY",
		capWidth, "CAPABILITIES",
		checkWidth, "LAST CHECK",
	)
	s += headerStyle.Render(header) + "\n"
	separator := lipgloss.NewStyle().Width(m.width - 4).Render(
		"─────────────────────────────────────────────────────────────────────────────",
	)
	s += styles.SubtleStyle.Render(separator) + "\n"

	// Tool rows
	for i, tool := range m.tools {
		rowStyle := lipgloss.NewStyle()
		if i == m.selectedIndex {
			rowStyle = rowStyle.Background(lipgloss.Color(styles.PrimaryColor)).
				Foreground(lipgloss.Color(styles.BackgroundColor))
		}

		// Status badge
		statusBadge := RenderStatusBadge(tool)

		// Latency
		latency := "─"
		if tool.Metrics.HandshakeLatency > 0 {
			latency = fmt.Sprintf("%dms", tool.Metrics.HandshakeLatency.Milliseconds())
		}

		// Capabilities count
		capCount := "─"
		if tool.Config != nil && len(tool.Config.Capabilities) > 0 {
			capCount = fmt.Sprintf("%d tools", len(tool.Config.Capabilities))
		}

		// Last check
		lastCheck := "never"
		if !tool.LastHealthCheck.IsZero() {
			lastCheck = formatRelativeTime(tool.LastHealthCheck)
		}

		row := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s",
			nameWidth, tool.Name,
			statusWidth, statusBadge,
			latencyWidth, latency,
			capWidth, capCount,
			checkWidth, lastCheck,
		)

		s += rowStyle.Render(row) + "\n"
	}

	return s
}

// SetSize updates the model dimensions
func (m *MCPToolsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// RenderStatusBadge creates a status badge for a tool (exported for use in components)
func RenderStatusBadge(tool types.MCPTool) string {
	if !tool.Enabled {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.SubtleColor)).
			Render("● DISABLED")
	}

	if tool.Status == types.ToolStatusReady {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.GreenColor)).
			Render("● ONLINE")
	} else if tool.Status == types.ToolStatusDegraded {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.YellowColor)).
			Render("● DEGRADED")
	} else {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.RedColor)).
			Render("● OFFLINE")
	}
}

// formatRelativeTime formats a timestamp as relative time
func formatRelativeTime(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return fmt.Sprintf("%ds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd ago", int(duration.Hours()/24))
	}
}
