package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) homeView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("GNDEC Question Paper Downloader"))
	b.WriteString("\n\n")

	totalPapers := 0
	for _, s := range m.allSubjects {
		totalPapers += len(s.Papers)
	}

	b.WriteString(infoStyle.Render(fmt.Sprintf("Subjects available: %d", len(m.allSubjects))))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render(fmt.Sprintf("Total papers indexed: %d", totalPapers)))
	b.WriteString("\n\n")
	b.WriteString(separator)
	b.WriteString("\n\n")

	if len(m.recent) > 0 {
		b.WriteString(subtitleStyle.Render("Recently Used"))
		b.WriteString("\n")
		for i, r := range m.recent {
			if i >= 5 {
				break
			}
			b.WriteString(itemStyle.Render(fmt.Sprintf("%s  %s  (%d papers)", cursorPrefix, r.Code, r.Downloads)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("enter  search subjects"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("q / esc  quit"))
	b.WriteString("\n")

	return docStyle.Render(b.String())
}

func (m model) handleHomeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.screen = screenSearch
		m.searchInput.Focus()
		m.searchMode = true
		m.searchInput.SetValue("")
		m.cursor = 0
		m.applyFilter()
		return m, nil
	}
	return m, nil
}
