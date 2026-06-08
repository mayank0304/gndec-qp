package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) detailView() string {
	var b strings.Builder

	if m.selectedSubject == nil {
		return "Error: no subject selected"
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("Subject: %s", m.selectedSubject.Code)))
	b.WriteString("\n\n")
	b.WriteString(infoStyle.Render(fmt.Sprintf("%d papers available", len(m.selectedSubject.Papers))))
	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n\n")

	homeDir, _ := os.UserHomeDir()
	dlDir := filepath.Join(homeDir, "Downloads", "Question Papers", m.selectedSubject.Code)

	allSelected := true
	anySelected := false
	for i, paper := range m.selectedSubject.Papers {
		checked := m.selected[i]
		if !checked {
			allSelected = false
		} else {
			anySelected = true
		}

		prefix := "  "
		if checked {
			prefix = checkedBox
		} else {
			prefix = uncheckedBox
		}

		status := ""
		paperPath := filepath.Join(dlDir, fmt.Sprintf("%s_%s.pdf", m.selectedSubject.Code, paper.Session))
		if _, err := os.Stat(paperPath); err == nil {
			status = "  " + successStyle.Render("✓ cached")
		}

		b.WriteString(itemStyle.Render(fmt.Sprintf("%s%s%s", prefix, paper.Session, status)))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	if allSelected && anySelected {
		b.WriteString(infoStyle.Render("All papers selected"))
	} else {
		b.WriteString(infoStyle.Render(fmt.Sprintf("Selected: %d/%d", selectedCount(m.selected), len(m.selectedSubject.Papers))))
	}
	b.WriteString("\n\n")
	b.WriteString(separator)
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("space  toggle selection"))
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("a  select all"))
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("d  download selected"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("esc  back"))
	b.WriteString("  ")
	b.WriteString(helpStyle.Render("q  home"))
	b.WriteString("\n")

	return docStyle.Render(b.String())
}

func selectedCount(sel map[int]bool) int {
	count := 0
	for _, v := range sel {
		if v {
			count++
		}
	}
	return count
}

func (m model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case " ":
		allSelected := true
		anySelected := false
		for i := range m.selectedSubject.Papers {
			if m.selected[i] {
				anySelected = true
			} else {
				allSelected = false
			}
		}
		if allSelected && anySelected {
			for i := range m.selectedSubject.Papers {
				m.selected[i] = false
			}
		} else {
			for i := range m.selectedSubject.Papers {
				m.selected[i] = true
			}
		}
		return m, nil

	case "a":
		allSelected := true
		for i := range m.selectedSubject.Papers {
			if !m.selected[i] {
				allSelected = false
				break
			}
		}
		if allSelected {
			for i := range m.selectedSubject.Papers {
				m.selected[i] = false
			}
		} else {
			for i := range m.selectedSubject.Papers {
				m.selected[i] = true
			}
		}
		return m, nil

	case "d":
		m.queue = nil
		m.progress = make(map[string]jobProgress)

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return m, nil
		}
		targetDir := filepath.Join(homeDir, "Downloads", "Question Papers", m.selectedSubject.Code)
		os.MkdirAll(targetDir, os.ModePerm)

		for i, paper := range m.selectedSubject.Papers {
			if m.selected[i] {
				baseFileName := fmt.Sprintf("%s_%s.pdf", m.selectedSubject.Code, paper.Session)
				filePath := filepath.Join(targetDir, baseFileName)
				m.queue = append(m.queue, downloadJob{
					subjectCode: m.selectedSubject.Code,
					session:     paper.Session,
					fileID:      paper.FileID,
					filePath:    filePath,
				})
				m.progress[paper.Session] = jobProgress{
					session:  paper.Session,
					percent:  0,
					complete: false,
				}
			}
		}

		m.total = len(m.queue)
		m.completed = 0
		m.errs = nil
		m.screen = screenDownload
		m.downloading = true
		return m, func() tea.Msg {
			return startDownloadMsg{}
		}

	case "q":
		m.screen = screenHome
		m.selectedSubject = nil
		m.selected = make(map[int]bool)
		return m, nil
	}

	return m, nil
}
