package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var noColor bool

func init() {
	_, noColor = os.LookupEnv("NO_COLOR")
	if !noColor {
		_, noColor = os.LookupEnv("NOCOLOR")
	}
}

func ifNoColor(s lipgloss.Style) lipgloss.Style {
	if noColor {
		return s.Foreground(lipgloss.NoColor{}).Background(lipgloss.NoColor{})
	}
	return s
}

var (
	primary   = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"}
	secondary = lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#60A5FA"}
	green     = lipgloss.AdaptiveColor{Light: "#059669", Dark: "#34D399"}
	red       = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#F87171"}
	muted     = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}
	muted2    = lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"}
	bgColor   = lipgloss.AdaptiveColor{Light: "#EEF2FF", Dark: "#1E1B4B"}
	borderC    = lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#374151"}
)

var (
	titleStyle = ifNoColor(lipgloss.NewStyle().Bold(true).Foreground(primary))

	successStyle = ifNoColor(lipgloss.NewStyle().Foreground(green))

	errorStyle = ifNoColor(lipgloss.NewStyle().Foreground(red))

	infoStyle = ifNoColor(lipgloss.NewStyle().Foreground(muted))

	dimStyle = ifNoColor(lipgloss.NewStyle().Foreground(muted2))

	helpStyle = ifNoColor(lipgloss.NewStyle().Foreground(muted2))

	selectedRow = ifNoColor(lipgloss.NewStyle().
				Background(bgColor).
				Foreground(secondary).
				Bold(true))

	normalRow = lipgloss.NewStyle()

	panelBorder = ifNoColor(lipgloss.NewStyle().Foreground(borderC))

	pad1 = lipgloss.NewStyle().PaddingLeft(1)
)

const (
	checkedBox   = "[✓] "
	uncheckedBox = "[ ] "
)

func bar(width int, pct float64) string {
	if pct > 1 {
		pct = 1
	}
	if pct < 0 {
		pct = 0
	}
	f := int(pct * float64(width))
	return barFilledStyle.Render(strings.Repeat("█", f)) +
		barEmptyStyle.Render(strings.Repeat("░", width-f))
}

var (
	barFilledStyle = ifNoColor(lipgloss.NewStyle().Foreground(green))
	barEmptyStyle  = ifNoColor(lipgloss.NewStyle().Foreground(muted2))
)

func hLine(width int) string {
	return panelBorder.Render("─" + strings.Repeat("─", width))
}

func borderTop(width int) string {
	return panelBorder.Render("╭" + strings.Repeat("─", width) + "╮")
}

func borderBottom(width int) string {
	return panelBorder.Render("╰" + strings.Repeat("─", width) + "╯")
}

func borderSides(line string) string {
	return panelBorder.Render("│") + line + panelBorder.Render("│")
}

func formatSession(raw string) string {
	s := raw
	makeup := false
	if strings.Contains(s, "(Makeup)") || strings.Contains(s, "(Makeup") {
		makeup = true
		s = strings.ReplaceAll(s, "(Makeup)", "")
		s = strings.ReplaceAll(s, "(Makeup", "")
	}

	parts := strings.Fields(s)
	if len(parts) == 0 {
		return raw
	}
	year := parts[len(parts)-1]

	if makeup {
		return year + " (makeup)"
	}
	return year
}
