package tui

import tea "github.com/charmbracelet/bubbletea"

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateSearch:
		return m.searchKeys(msg)
	case stateSelect:
		return m.selectKeys(msg)
	case stateDownload:
		return m.downloadKeys(msg)
	}
	return m, nil
}

func (m model) searchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit

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
		m.selectSubject(m.subjects[m.cursor].Code)
		m.state = stateSelect
		return m, nil

	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.applyFilter()
		return m, cmd
	}
}

func (m model) selectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "esc":
		m.state = stateSearch
		m.selSubject = nil
		return m, nil

	case "up", "k":
		if m.detailCursor > 0 {
			m.detailCursor--
		}
		return m, nil

	case "down", "j":
		if m.detailCursor < len(m.selSubject.Papers)-1 {
			m.detailCursor++
		}
		return m, nil

	case " ":
		m.selected[m.detailCursor] = !m.selected[m.detailCursor]
		return m, nil

	case "a":
		allOn := true
		for _, v := range m.selected {
			if !v {
				allOn = false
				break
			}
		}
		for i := range m.selSubject.Papers {
			m.selected[i] = !allOn
		}
		return m, nil

	case "enter":
		m.buildJobs()
		if len(m.jobs) == 0 {
			return m, nil
		}
		m.state = stateDownload
		m.startDownloads()
		return m, m.spinner.Tick
	}

	return m, nil
}

func (m model) downloadKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.state = stateSelect
		m.jobs = nil
		return m, nil

	case "esc", "q":
		if !m.finished {
			return m, nil
		}
		m.state = stateSearch
		m.selSubject = nil
		m.jobs = nil
		return m, nil
	}
	return m, nil
}
