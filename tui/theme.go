package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary   = lipgloss.Color("#7C3AED")
	colorSecondary = lipgloss.Color("#3B82F6")
	colorSuccess   = lipgloss.Color("#10B981")
	colorWarning   = lipgloss.Color("#F59E0B")
	colorError     = lipgloss.Color("#EF4444")

	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSecondary).
			MarginBottom(1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			MarginBottom(1)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				MarginBottom(1).
				Foreground(colorSecondary).
				Bold(true)

	docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"}).
			MarginTop(1)

	separator = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#374151"}).
			Render("────────────────────────────────────")

	checkedBox = "[✓] "
	uncheckedBox = "[ ] "
	cursorPrefix = "▸ "
)
