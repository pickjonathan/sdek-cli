package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	// Primary colors
	PrimaryColor   = lipgloss.Color("#7D56F4")
	SecondaryColor = lipgloss.Color("#FF8C00")
	AccentColor    = lipgloss.Color("#FF69B4")

	// Status colors
	GreenColor  = lipgloss.Color("#00FF00")
	YellowColor = lipgloss.Color("#FFFF00")
	RedColor    = lipgloss.Color("#FF0000")

	// Text colors
	TextColor   = lipgloss.Color("#FAFAFA")
	SubtleColor = lipgloss.Color("#888888")

	// Background colors
	BackgroundColor = lipgloss.Color("#1A1A1A")
	BorderColor     = lipgloss.Color("#444444")
)

// Base styles
var (
	BaseStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Background(BackgroundColor)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			MarginBottom(1)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentColor).
			Padding(0, 1)

	KeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(SecondaryColor).
			Width(12)

	ValueStyle = lipgloss.NewStyle().
			Foreground(TextColor)

	SubtleStyle = lipgloss.NewStyle().
			Foreground(SubtleColor)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(RedColor).
			Bold(true).
			Padding(1)
)

// Container styles
var (
	ContainerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(BorderColor).
			Padding(0, 1)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1).
			MarginBottom(1)
)

// List styles
var (
	ListItemStyle = lipgloss.NewStyle().
			Padding(0, 2)

	SelectedListItemStyle = lipgloss.NewStyle().
				Foreground(BackgroundColor).
				Background(PrimaryColor).
				Padding(0, 2).
				Bold(true)

	ListTitleStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true).
			Padding(0, 2).
			MarginBottom(1)
)

// Status styles
var (
	StatusGreenStyle = lipgloss.NewStyle().
				Foreground(GreenColor).
				Bold(true)

	StatusYellowStyle = lipgloss.NewStyle().
				Foreground(YellowColor).
				Bold(true)

	StatusRedStyle = lipgloss.NewStyle().
			Foreground(RedColor).
			Bold(true)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(SubtleColor).
			Padding(0, 1)
)

// GetStatusStyle returns the appropriate style for a risk status
func GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "green", "low", "Green", "Low":
		return StatusGreenStyle
	case "yellow", "medium", "Yellow", "Medium":
		return StatusYellowStyle
	case "red", "high", "critical", "Red", "High", "Critical":
		return StatusRedStyle
	default:
		return SubtleStyle
	}
}

// FormatPercentage formats a compliance percentage with appropriate color
func FormatPercentage(percent float64) string {
	var style lipgloss.Style
	if percent >= 80 {
		style = StatusGreenStyle
	} else if percent >= 50 {
		style = StatusYellowStyle
	} else {
		style = StatusRedStyle
	}
	return style.Render(lipgloss.NewStyle().Width(6).Align(lipgloss.Right).Render(
		lipgloss.NewStyle().Render(
			fmt.Sprintf("%5.1f%%", percent),
		),
	))
}

// Modal styles
var (
	// ModalOverlayStyle creates a dimmed overlay for modals
	ModalOverlayStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#000000")).
				Foreground(lipgloss.Color("#888888"))

	// ModalBoxStyle is the container for modal dialogs
	ModalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(PrimaryColor).
			Background(BackgroundColor).
			Foreground(TextColor).
			Padding(2, 4).
			Width(60)

	// ModalTitleStyle for modal dialog titles
	ModalTitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Underline(true).
			Align(lipgloss.Center)

	// ModalButtonStyle for modal action buttons
	ModalButtonStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PrimaryColor).
				Padding(0, 2).
				Foreground(TextColor)

	// ModalButtonActiveStyle for focused modal buttons
	ModalButtonActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(AccentColor).
				Background(PrimaryColor).
				Foreground(lipgloss.Color("#000000")).
				Bold(true).
				Padding(0, 2)
)

// Table styles
var (
	// TableHeaderStyle for table column headers
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true).
				Underline(true).
				Padding(0, 1).
				Align(lipgloss.Left)

	// TableCellStyle for regular table cells
	TableCellStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Padding(0, 1).
			Align(lipgloss.Left)

	// TableCellHighlightStyle for highlighted/selected cells
	TableCellHighlightStyle = lipgloss.NewStyle().
				Foreground(AccentColor).
				Bold(true).
				Padding(0, 1).
				Align(lipgloss.Left)

	// TableBorderStyle for table borders
	TableBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(BorderColor)

	// TableRowSelectedStyle for selected table rows
	TableRowSelectedStyle = lipgloss.NewStyle().
				Background(BorderColor).
				Foreground(PrimaryColor).
				Bold(true)
)

// Spinner styles
var (
	// SpinnerStyle for loading spinners
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	// SpinnerTextStyle for text next to spinners
	SpinnerTextStyle = lipgloss.NewStyle().
				Foreground(TextColor).
				Italic(true)

	// ProgressBarStyle for progress indicators
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true)

	// ProgressBarEmptyStyle for unfilled progress
	ProgressBarEmptyStyle = lipgloss.NewStyle().
				Foreground(SubtleColor)
)

// Additional utility styles
var (
	// DividerStyle for horizontal dividers
	DividerStyle = lipgloss.NewStyle().
			Foreground(BorderColor).
			Bold(false)

	// TagStyle for small tags/badges
	TagStyle = lipgloss.NewStyle().
			Background(PrimaryColor).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1).
			Bold(true)

	// CodeBlockStyle for code snippets
	CodeBlockStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2A2A2A")).
			Foreground(AccentColor).
			Padding(1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(BorderColor)

	// TooltipStyle for tooltips and hints
	TooltipStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#444444")).
			Foreground(TextColor).
			Padding(0, 1).
			Italic(true)
)
