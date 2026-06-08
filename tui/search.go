package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) searchView() string {
	var b strings.Builder

	b.WriteString(subtitleStyle.Render("Search Subjects"))
	b.WriteString("\n\n")
	b.WriteString(m.searchInput.View())
	b.WriteString("\n\n")

	if len(m.subjects) == 0 {
		b.WriteString(infoStyle.Render("No subjects found matching your query."))
		b.WriteString("\n")
	} else {
		b.WriteString(infoStyle.Render(fmt.Sprintf("%d subjects found", len(m.subjects))))
		b.WriteString("\n\n")

		maxVisible := m.height - 10
		if maxVisible < 5 {
			maxVisible = 5
		}
		start := 0
		if m.cursor > maxVisible-1 {
			start = m.cursor - maxVisible + 1
		}
		end := start + maxVisible
		if end > len(m.subjects) {
			end = len(m.subjects)
		}

		for i, s := range m.subjects[start:end] {
			actualIdx := start + i
			prefix := "  "
			if actualIdx == m.cursor {
				prefix = "▸ "
			}

			code := s.Code
			if actualIdx == m.cursor {
				b.WriteString(selectedItemStyle.Render(fmt.Sprintf("%s%s", prefix, code)))
			} else {
				b.WriteString(itemStyle.Render(fmt.Sprintf("%s%s", prefix, code)))
			}
			b.WriteString("\n")
			countStr := infoStyle.Render(fmt.Sprintf("   %d papers", len(s.Papers)))
			if actualIdx == m.cursor {
				b.WriteString(countStr)
			} else {
				b.WriteString(countStr)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑↓  navigate"))
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("enter  select"))
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("esc  back"))
	b.WriteString("\n")

	return docStyle.Render(b.String())
}

func (m model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
		if m.cursor < len(m.subjects)-1 {
			m.cursor++
		}
		return m, nil

	case "enter":
		if len(m.subjects) == 0 {
			return m, nil
		}
		m.selectedSubject = &m.subjects[m.cursor]
		m.screen = screenDetail
		m.selected = make(map[int]bool)
		for i := range m.selectedSubject.Papers {
			m.selected[i] = true
		}
		m.queue = nil
		return m, nil

	case "esc":
		m.screen = screenHome
		m.cursor = 0
		m.searchInput.SetValue("")
		return m, nil

	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.applyFilter()
		return m, cmd
	}
}

func (m *model) applyFilter() {
	query := strings.ToUpper(strings.TrimSpace(m.searchInput.Value()))
	if query == "" {
		m.subjects = m.allSubjects
	} else {
		var filtered []subjectEntry
		for _, s := range m.allSubjects {
			if strings.Contains(strings.ToUpper(s.Code), query) {
				filtered = append(filtered, s)
			}
		}
		m.subjects = filtered
	}
	if m.cursor >= len(m.subjects) {
		m.cursor = 0
	}
}
