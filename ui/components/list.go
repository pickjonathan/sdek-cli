package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pickjonathan/sdek-cli/ui/styles"
)

// ListConfig defines configuration for list rendering
type ListConfig struct {
	Items            []string
	SelectedIndex    int
	Width            int
	MaxDisplay       int
	ShowNumbers      bool
	ShowCursor       bool
	CursorSymbol     string
	NumberedSymbol   string
	SelectedPrefix   string
	UnselectedPrefix string
}

// DefaultListConfig returns a list with default styling
func DefaultListConfig() ListConfig {
	return ListConfig{
		Width:            60,
		MaxDisplay:       10,
		ShowNumbers:      false,
		ShowCursor:       true,
		CursorSymbol:     "▶",
		NumberedSymbol:   ".",
		SelectedPrefix:   "▶ ",
		UnselectedPrefix: "  ",
	}
}

// RenderList creates a styled list with cursor or numbers
func RenderList(config ListConfig) string {
	if len(config.Items) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.SubtleColor)).
			Italic(true)
		return emptyStyle.Render("No items to display")
	}

	// Determine visible range with pagination
	start, end := getPaginationRange(config.SelectedIndex, len(config.Items), config.MaxDisplay)

	var lines []string
	for i := start; i < end; i++ {
		item := config.Items[i]
		isSelected := i == config.SelectedIndex

		// Choose style
		var line string
		if config.ShowNumbers {
			line = renderNumberedItem(i+1, item, isSelected, config)
		} else if config.ShowCursor {
			line = renderCursorItem(item, isSelected, config)
		} else {
			line = renderPlainItem(item, isSelected, config)
		}

		lines = append(lines, line)
	}

	// Add pagination indicator if needed
	if end < len(config.Items) {
		moreStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.SubtleColor)).
			Italic(true)
		lines = append(lines, moreStyle.Render(fmt.Sprintf("  ... %d more items", len(config.Items)-end)))
	}

	return strings.Join(lines, "\n")
}

// renderCursorItem creates a list item with cursor indicator
func renderCursorItem(text string, isSelected bool, config ListConfig) string {
	prefix := config.UnselectedPrefix
	if isSelected {
		prefix = config.SelectedPrefix
	}

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextColor)).
		Width(config.Width - len(prefix))

	if isSelected {
		itemStyle = itemStyle.
			Foreground(lipgloss.Color(styles.PrimaryColor)).
			Bold(true)
	}

	return prefix + itemStyle.Render(text)
}

// renderNumberedItem creates a numbered list item
func renderNumberedItem(number int, text string, isSelected bool, config ListConfig) string {
	numberStr := fmt.Sprintf("%d%s ", number, config.NumberedSymbol)

	numberStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.SubtleColor)).
		Width(5)

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextColor)).
		Width(config.Width - 5)

	if isSelected {
		numberStyle = numberStyle.
			Foreground(lipgloss.Color(styles.PrimaryColor)).
			Bold(true)
		itemStyle = itemStyle.
			Foreground(lipgloss.Color(styles.PrimaryColor)).
			Bold(true)
	}

	return numberStyle.Render(numberStr) + itemStyle.Render(text)
}

// renderPlainItem creates a plain list item without cursor or numbers
func renderPlainItem(text string, isSelected bool, config ListConfig) string {
	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextColor)).
		Width(config.Width).
		PaddingLeft(2)

	if isSelected {
		itemStyle = itemStyle.
			Foreground(lipgloss.Color(styles.PrimaryColor)).
			Bold(true).
			Background(lipgloss.Color(styles.BorderColor))
	}

	return itemStyle.Render(text)
}

// RenderKeyValueList creates a list of key-value pairs
func RenderKeyValueList(items map[string]string, width int) string {
	if len(items) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.SubtleColor)).
			Italic(true)
		return emptyStyle.Render("No data available")
	}

	if width <= 0 {
		width = 60
	}

	keyWidth := width / 3
	valueWidth := width - keyWidth - 3

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.SecondaryColor)).
		Bold(true).
		Width(keyWidth)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextColor)).
		Width(valueWidth)

	var lines []string
	for key, value := range items {
		line := keyStyle.Render(key+":") + " " + valueStyle.Render(value)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// RenderTableRow creates a formatted table row with columns
func RenderTableRow(columns []string, widths []int, isHeader bool) string {
	if len(columns) != len(widths) {
		return "Error: Column count mismatch"
	}

	var cells []string
	for i, col := range columns {
		cellStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextColor)).
			Width(widths[i]).
			PaddingLeft(1).
			PaddingRight(1)

		if isHeader {
			cellStyle = cellStyle.
				Foreground(lipgloss.Color(styles.PrimaryColor)).
				Bold(true).
				Underline(true)
		}

		cells = append(cells, cellStyle.Render(col))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, cells...)
}

// getPaginationRange calculates the visible range for pagination
func getPaginationRange(selectedIndex, totalItems, maxDisplay int) (start, end int) {
	if totalItems <= maxDisplay {
		return 0, totalItems
	}

	// Center the selected item when possible
	start = selectedIndex - maxDisplay/2
	if start < 0 {
		start = 0
	}

	end = start + maxDisplay
	if end > totalItems {
		end = totalItems
		start = end - maxDisplay
		if start < 0 {
			start = 0
		}
	}

	return start, end
}

// RenderPaginationInfo creates pagination information text
func RenderPaginationInfo(currentIndex, totalItems int) string {
	if totalItems == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.SubtleColor)).
			Render("0 of 0")
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.SecondaryColor))

	return style.Render(fmt.Sprintf("%d of %d", currentIndex+1, totalItems))
}
