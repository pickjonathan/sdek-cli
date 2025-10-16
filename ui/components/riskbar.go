package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pickjonathan/sdek-cli/ui/styles"
)

// RiskDistribution represents the counts of each risk level
type RiskDistribution struct {
	High   int
	Medium int
	Low    int
	Total  int
}

// RenderRiskBar creates a visual representation of risk distribution
// Shows colored segments proportional to risk levels
func RenderRiskBar(dist RiskDistribution, width int) string {
	if width <= 0 {
		width = 50
	}

	if dist.Total == 0 {
		// Empty bar
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.SubtleColor))
		return emptyStyle.Render(strings.Repeat("░", width))
	}

	// Calculate proportions
	highWidth := int(float64(dist.High) / float64(dist.Total) * float64(width))
	mediumWidth := int(float64(dist.Medium) / float64(dist.Total) * float64(width))
	lowWidth := width - highWidth - mediumWidth // Remaining space

	// Create styled segments
	highStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.RedColor)).
		Bold(true)

	mediumStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.YellowColor)).
		Bold(true)

	lowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.GreenColor)).
		Bold(true)

	// Build bar
	highBar := highStyle.Render(strings.Repeat("█", highWidth))
	mediumBar := mediumStyle.Render(strings.Repeat("█", mediumWidth))
	lowBar := lowStyle.Render(strings.Repeat("█", lowWidth))

	return highBar + mediumBar + lowBar
}

// RenderRiskBarWithLegend creates a risk bar with count labels
func RenderRiskBarWithLegend(dist RiskDistribution, width int) string {
	if width <= 0 {
		width = 50
	}

	// Create the bar
	bar := RenderRiskBar(dist, width-20) // Reserve space for legend

	// Create legend
	legendStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextColor))

	highStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.RedColor)).
		Bold(true)

	mediumStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.YellowColor)).
		Bold(true)

	lowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.GreenColor)).
		Bold(true)

	legend := fmt.Sprintf("%s %d  %s %d  %s %d",
		highStyle.Render("●"),
		dist.High,
		mediumStyle.Render("●"),
		dist.Medium,
		lowStyle.Render("●"),
		dist.Low,
	)

	return legendStyle.Render(bar + "  " + legend)
}

// RenderRiskSummary creates a detailed risk summary box
func RenderRiskSummary(dist RiskDistribution, title string, width int) string {
	if width <= 0 {
		width = 60
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.PrimaryColor)).
		Bold(true).
		Width(width - 4)

	// Calculate percentages
	highPercent := 0.0
	mediumPercent := 0.0
	lowPercent := 0.0
	if dist.Total > 0 {
		highPercent = float64(dist.High) / float64(dist.Total) * 100
		mediumPercent = float64(dist.Medium) / float64(dist.Total) * 100
		lowPercent = float64(dist.Low) / float64(dist.Total) * 100
	}

	// Create risk level rows
	highRow := renderRiskRow("High Risk", dist.High, highPercent, styles.RedColor, width-4)
	mediumRow := renderRiskRow("Medium Risk", dist.Medium, mediumPercent, styles.YellowColor, width-4)
	lowRow := renderRiskRow("Low Risk", dist.Low, lowPercent, styles.GreenColor, width-4)

	// Create total row
	totalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextColor)).
		Bold(true).
		Width(width - 4)
	totalRow := totalStyle.Render(fmt.Sprintf("Total: %d", dist.Total))

	// Combine everything
	content := titleStyle.Render(title) + "\n\n" +
		highRow + "\n" +
		mediumRow + "\n" +
		lowRow + "\n\n" +
		totalRow

	// Add border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(styles.BorderColor)).
		Width(width).
		Padding(1)

	return boxStyle.Render(content)
}

// renderRiskRow creates a single row showing risk level, count, and percentage
func renderRiskRow(label string, count int, percent float64, color lipgloss.Color, width int) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true).
		Width(15)

	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextColor)).
		Width(10)

	percentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.SubtleColor)).
		Width(10)

	barWidth := width - 35
	bar := RenderRiskBar(RiskDistribution{
		High:   count,
		Medium: 0,
		Low:    0,
		Total:  count,
	}, barWidth)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		labelStyle.Render(label),
		countStyle.Render(fmt.Sprintf("%d", count)),
		percentStyle.Render(fmt.Sprintf("%.1f%%", percent)),
		bar,
	)
}

// RenderCompactRiskIndicator creates a small risk indicator with icon
func RenderCompactRiskIndicator(level string) string {
	var icon string
	var color lipgloss.Color

	switch strings.ToLower(level) {
	case "high", "red":
		icon = "✗"
		color = lipgloss.Color(styles.RedColor)
	case "medium", "yellow":
		icon = "⚠"
		color = lipgloss.Color(styles.YellowColor)
	case "low", "green":
		icon = "✓"
		color = lipgloss.Color(styles.GreenColor)
	default:
		icon = "○"
		color = lipgloss.Color(styles.SubtleColor)
	}

	iconStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true)

	return iconStyle.Render(icon)
}
