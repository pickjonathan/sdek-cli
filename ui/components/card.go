package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pickjonathan/sdek-cli/ui/styles"
)

// CardConfig defines the configuration for a card component
type CardConfig struct {
	Title       string
	Content     string
	Width       int
	Height      int
	BorderColor lipgloss.Color
	TitleColor  lipgloss.Color
	ShowBorder  bool
}

// DefaultCardConfig returns a card with default styling
func DefaultCardConfig() CardConfig {
	return CardConfig{
		Width:       40,
		Height:      10,
		BorderColor: lipgloss.Color(styles.PrimaryColor),
		TitleColor:  lipgloss.Color(styles.PrimaryColor),
		ShowBorder:  true,
	}
}

// RenderCard creates a styled card with title and content
func RenderCard(config CardConfig) string {
	if config.Width <= 0 {
		config.Width = 40
	}
	if config.Height <= 0 {
		config.Height = 10
	}

	// Create title style
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.TitleColor)).
		Bold(true).
		Padding(0, 1)

	// Create content style
	contentStyle := lipgloss.NewStyle().
		Width(config.Width-4). // Account for borders and padding
		Padding(0, 1)

	// Create border style
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(config.BorderColor)).
		Width(config.Width).
		Padding(1)

	// Format title
	title := titleStyle.Render(config.Title)

	// Format content with wrapping
	content := contentStyle.Render(config.Content)

	// Combine title and content
	body := title + "\n\n" + content

	// Apply border if enabled
	if config.ShowBorder {
		return borderStyle.Render(body)
	}

	return body
}

// RenderSummaryCard creates a card optimized for summary displays (key-value pairs)
func RenderSummaryCard(title, value, description string, width int) string {
	if width <= 0 {
		width = 30
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.SecondaryColor)).
		Bold(false).
		Width(width - 4)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.PrimaryColor)).
		Bold(true).
		Width(width - 4)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.SubtleColor)).
		Width(width - 4)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(styles.BorderColor)).
		Width(width).
		Padding(1, 2)

	content := titleStyle.Render(title) + "\n" +
		valueStyle.Render(value) + "\n" +
		descStyle.Render(description)

	return borderStyle.Render(content)
}

// RenderStatusCard creates a card with status indicator (for pass/fail/warning states)
func RenderStatusCard(title, message, status string, width int) string {
	if width <= 0 {
		width = 40
	}

	// Determine colors based on status
	var statusColor lipgloss.Color
	var icon string
	switch strings.ToLower(status) {
	case "pass", "success", "green":
		statusColor = lipgloss.Color(styles.GreenColor)
		icon = "✓"
	case "warning", "caution", "yellow":
		statusColor = lipgloss.Color(styles.YellowColor)
		icon = "⚠"
	case "fail", "error", "red":
		statusColor = lipgloss.Color(styles.RedColor)
		icon = "✗"
	default:
		statusColor = lipgloss.Color(styles.SecondaryColor)
		icon = "○"
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.PrimaryColor)).
		Bold(true).
		Width(width - 6)

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor)).
		Bold(true)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextColor)).
		Width(width - 6)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(statusColor)).
		Width(width).
		Padding(1)

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		iconStyle.Render(icon+" "),
		titleStyle.Render(title),
	)

	content := header + "\n\n" + messageStyle.Render(message)

	return borderStyle.Render(content)
}

// RenderInfoCard creates a simple informational card without heavy styling
func RenderInfoCard(title, content string, width int) string {
	if width <= 0 {
		width = 40
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.AccentColor)).
		Bold(true).
		Underline(true).
		Width(width - 4)

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextColor)).
		Width(width - 4)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(styles.BorderColor)).
		Width(width).
		Padding(1, 2)

	body := titleStyle.Render(title) + "\n\n" + contentStyle.Render(content)

	return boxStyle.Render(body)
}
