package tui

import (
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/IshpreetSingh8264/gndec-qp/db"
)

func Run() {
	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	m.program = p
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func RunWithCode(code string) {
	m := initialModel()
	m.searchInput.SetValue(code)
	m.screen = screenSearch
	m.applyFilter()
	if len(m.subjects) == 1 {
		m.selectedSubject = &m.subjects[0]
		m.screen = screenDetail
		m.selected = make(map[int]bool)
		for i := range m.subjects[0].Papers {
			m.selected[i] = false
		}
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	m.program = p
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type a subject code (e.g. PCIT-114)..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	s := spinner.New()
	s.Style = subtitleStyle
	s.Spinner = spinner.Dot

	pb := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	subjects := loadSubjects()

	return model{
		screen:        screenHome,
		allSubjects:   subjects,
		subjects:      subjects,
		searchInput:   ti,
		cursor:        0,
		selected:      make(map[int]bool),
		spinner:       s,
		progBar:       pb,
		progress:      make(map[string]jobProgress),
		recent:        loadRecent(),
		searchMode:    false,
	}
}

func loadSubjects() []subjectEntry {
	codes := make([]string, 0, len(db.PaperRegistry))
	for code := range db.PaperRegistry {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	entries := make([]subjectEntry, len(codes))
	for i, code := range codes {
		entries[i] = subjectEntry{
			Code:   code,
			Papers: db.PaperRegistry[code],
		}
	}
	return entries
}

type model struct {
	screen  screen
	program *tea.Program

	allSubjects []subjectEntry
	subjects    []subjectEntry

	searchInput textinput.Model
	cursor      int
	searchMode  bool

	selectedSubject *subjectEntry
	selected        map[int]bool

	queue       []downloadJob
	progress    map[string]jobProgress
	spinner     spinner.Model
	progBar     progress.Model
	downloading bool
	completed   int
	total       int
	errs        []string

	recent []recentEntry

	width, height int
	ready         bool
	quitting      bool
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, textinput.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ready = true
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.screen == screenHome {
				m.quitting = true
				return m, tea.Quit
			}
			if m.screen == screenDownload && m.downloading {
				return m, nil
			}
			m.screen = screenHome
			m.selectedSubject = nil
			m.selected = make(map[int]bool)
			m.queue = nil
			m.errs = nil
			m.completed = 0
			m.total = 0
			return m, nil

		case "esc":
			if m.screen == screenSearch {
				m.screen = screenHome
				m.cursor = 0
				return m, nil
			}
			if m.screen == screenDetail {
				m.screen = screenSearch
				m.selectedSubject = nil
				m.selected = make(map[int]bool)
				return m, nil
			}
			if m.screen == screenDownload && !m.downloading {
				m.screen = screenDetail
				return m, nil
			}
			if m.screen == screenHome {
				m.quitting = true
				return m, tea.Quit
			}
		}

		switch m.screen {
		case screenHome:
			return m.handleHomeKey(msg)
		case screenSearch:
			return m.handleSearchKey(msg)
		case screenDetail:
			return m.handleDetailKey(msg)
		case screenDownload:
			return m.handleDownloadKey(msg)
		}

	case startDownloadMsg:
		return m.startDownloads()

	case progressMsg:
		p := m.progress[msg.session]
		p.percent = msg.percent
		if msg.err != nil {
			p.err = msg.err
			p.complete = true
		}
		m.progress[msg.session] = p
		return m, nil

	case downloadDoneMsg:
		m.downloading = false
		m.completed = msg.completed
		m.total = msg.total
		m.errs = msg.errors
		if len(m.errs) == 0 {
			recordRecent(m.selectedSubject.Code)
			m.recent = loadRecent()
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}
	if m.quitting {
		return ""
	}

	switch m.screen {
	case screenHome:
		return m.homeView()
	case screenSearch:
		return m.searchView()
	case screenDetail:
		return m.detailView()
	case screenDownload:
		return m.downloadView()
	}
	return ""
}


