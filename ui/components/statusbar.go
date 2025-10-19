package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pickjonathan/sdek-cli/ui/styles"
)

// StatusBarConfig defines configuration for the status bar
type StatusBarConfig struct {
	Width      int
	LeftText   string
	CenterText string
	RightText  string
	Shortcuts  []Shortcut
	ShowHelp   bool
}

// Shortcut represents a keyboard shortcut with key and description
type Shortcut struct {
	Key         string
	Description string
}

// RenderStatusBar creates a bottom status bar with information and shortcuts
func RenderStatusBar(config StatusBarConfig) string {
	if config.Width <= 0 {
		config.Width = 80
	}

	// Create style for status bar background
	barStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(styles.BorderColor)).
		Foreground(lipgloss.Color(styles.TextColor)).
		Width(config.Width).
		Padding(0, 1)

	// Create content sections
	leftStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.PrimaryColor)).
		Bold(true)

	centerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextColor))

	rightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.SubtleColor))

	// Build shortcuts text if enabled
	var content string
	if config.ShowHelp && len(config.Shortcuts) > 0 {
		shortcuts := renderShortcuts(config.Shortcuts)
		content = shortcuts
	} else {
		// Calculate widths for left, center, right sections
		sectionWidth := (config.Width - 6) / 3 // Account for padding

		left := leftStyle.Render(truncate(config.LeftText, sectionWidth))
		center := centerStyle.Render(truncate(config.CenterText, sectionWidth))
		right := rightStyle.Render(truncate(config.RightText, sectionWidth))

		// Join sections with spacing
		content = lipgloss.JoinHorizontal(
			lipgloss.Left,
			left,
			strings.Repeat(" ", sectionWidth-lipgloss.Width(config.LeftText)),
			center,
			strings.Repeat(" ", sectionWidth-lipgloss.Width(config.CenterText)),
			right,
		)
	}

	return barStyle.Render(content)
}

// renderShortcuts creates a formatted shortcuts string
func renderShortcuts(shortcuts []Shortcut) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.AccentColor)).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextColor))

	var parts []string
	for _, sc := range shortcuts {
		part := keyStyle.Render(sc.Key) + " " + descStyle.Render(sc.Description)
		parts = append(parts, part)
	}

	return strings.Join(parts, "  ")
}

// RenderSimpleStatusBar creates a simple status bar with just text
func RenderSimpleStatusBar(text string, width int) string {
	if width <= 0 {
		width = 80
	}

	style := lipgloss.NewStyle().
		Background(lipgloss.Color(styles.BorderColor)).
		Foreground(lipgloss.Color(styles.TextColor)).
		Width(width).
		Padding(0, 1)

	return style.Render(truncate(text, width-2))
}

// RenderHelpBar creates a help bar showing available commands
func RenderHelpBar(shortcuts []Shortcut, width int) string {
	if width <= 0 {
		width = 80
	}

	config := StatusBarConfig{
		Width:     width,
		Shortcuts: shortcuts,
		ShowHelp:  true,
	}

	return RenderStatusBar(config)
}

// RenderNavigationBar creates a navigation status bar
func RenderNavigationBar(currentScreen, totalScreens int, screenName string, width int) string {
	if width <= 0 {
		width = 80
	}

	leftText := fmt.Sprintf("Screen %d/%d", currentScreen, totalScreens)
	centerText := screenName
	rightText := "? help • q quit"

	config := StatusBarConfig{
		Width:      width,
		LeftText:   leftText,
		CenterText: centerText,
		RightText:  rightText,
		ShowHelp:   false,
	}

	return RenderStatusBar(config)
}

// RenderLoadingBar creates a loading/progress status bar
func RenderLoadingBar(message string, percent float64, width int) string {
	if width <= 0 {
		width = 80
	}

	// Create progress bar
	barWidth := width - len(message) - 10 // Leave space for message and percentage
	filledWidth := int(float64(barWidth) * percent / 100.0)
	emptyWidth := barWidth - filledWidth

	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.PrimaryColor)).
		Bold(true)

	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.SubtleColor))

	bar := progressStyle.Render(strings.Repeat("█", filledWidth)) +
		emptyStyle.Render(strings.Repeat("░", emptyWidth))

	percentText := fmt.Sprintf("%.0f%%", percent)

	// Combine everything
	content := fmt.Sprintf("%s %s %s", message, bar, percentText)

	barStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(styles.BorderColor)).
		Foreground(lipgloss.Color(styles.TextColor)).
		Width(width).
		Padding(0, 1)

	return barStyle.Render(content)
}

// RenderErrorBar creates an error status bar
func RenderErrorBar(errorMsg string, width int) string {
	if width <= 0 {
		width = 80
	}

	errorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(styles.RedColor)).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Width(width).
		Padding(0, 1)

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	content := iconStyle.Render("✗ ") + truncate(errorMsg, width-4)

	return errorStyle.Render(content)
}

// RenderSuccessBar creates a success status bar
func RenderSuccessBar(message string, width int) string {
	if width <= 0 {
		width = 80
	}

	successStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(styles.GreenColor)).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Width(width).
		Padding(0, 1)

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Bold(true)

	content := iconStyle.Render("✓ ") + truncate(message, width-4)

	return successStyle.Render(content)
}

// truncate truncates text to fit within a given width
func truncate(text string, width int) string {
	if len(text) <= width {
		return text
	}
	if width <= 3 {
		return text[:width]
	}
	return text[:width-3] + "..."
}
