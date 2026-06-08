package tui

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/IshpreetSingh8264/gndec-qp/db"
)

func Run() {
	m := initialModel()
	m.state = stateSearch
	p := tea.NewProgram(m, tea.WithAltScreen())
	p.Send(setProgramMsg{program: p})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func RunWithCode(code string) {
	m := initialModel()
	m.state = stateSelect
	m.searchInput.SetValue(code)
	m.selectSubject(code)
	p := tea.NewProgram(m, tea.WithAltScreen())
	p.Send(setProgramMsg{program: p})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// ─── model ────────────────────────────────────────────────

type model struct {
	state    state
	program  *tea.Program
	quitting bool

	allSubjects []subjectEntry
	subjects    []subjectEntry
	cursor      int

	searchInput textinput.Model

	selSubject   *subjectEntry
	selected     map[int]bool
	detailCursor int

	jobs     []downloadJob
	progress map[string]float64
	done     map[string]bool
	failed   map[string]error
	spinner  spinner.Model
	finished bool
	errors   int

	recent []recentEntry

	width, height int
	ready         bool
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type a subject code..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Dot

	codes := make([]string, 0, len(db.PaperRegistry))
	for code := range db.PaperRegistry {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	entries := make([]subjectEntry, len(codes))
	for i, code := range codes {
		entries[i] = subjectEntry{Code: code, Papers: db.PaperRegistry[code]}
	}

	return model{
		allSubjects: entries,
		subjects:    entries,
		searchInput: ti,
		selected:    make(map[int]bool),
		progress:    make(map[string]float64),
		done:        make(map[string]bool),
		failed:      make(map[string]error),
		spinner:     s,
		recent:      loadRecent(),
	}
}

// ─── bubbletea interface ─────────────────────────────────

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, textinput.Blink)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case setProgramMsg:
		m.program = msg.program.(*tea.Program)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case progressMsg:
		if msg.err != nil {
			m.failed[msg.session] = msg.err
			m.errors++
		} else if msg.complete {
			m.done[msg.session] = true
		} else {
			m.progress[msg.session] = msg.percent
		}
		return m, nil

	case downloadDoneMsg:
		m.finished = true
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.state == stateSearch {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.applyFilter()
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}
	if m.quitting {
		return ""
	}
	switch m.state {
	case stateSearch:
		return m.renderSearch()
	case stateSelect:
		return m.renderSelect()
	case stateDownload:
		return m.renderDownload()
	}
	return ""
}

// ─── search view ──────────────────────────────────────────

func (m model) renderSearch() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Find Subject"))
	b.WriteString("\n\n")
	b.WriteString(m.searchInput.View())
	b.WriteString("\n\n")

	if len(m.subjects) == 0 {
		b.WriteString(infoStyle.Render("No subjects match."))
	} else {
		maxVis := max(5, m.height-10)
		start := (m.cursor / maxVis) * maxVis
		if start > len(m.subjects)-maxVis {
			start = len(m.subjects) - maxVis
		}
		if start < 0 {
			start = 0
		}
		end := start + maxVis
		if end > len(m.subjects) {
			end = len(m.subjects)
		}

		recMap := map[string]bool{}
		for _, r := range m.recent {
			recMap[r.Code] = true
		}

		for i := start; i < end; i++ {
			s := m.subjects[i]
			prefix := "  "
			hl := i == m.cursor

			line := fmt.Sprintf("%s%s", prefix, s.Code)
			tag := ""
			if recMap[s.Code] {
				tag = " " + dimStyle.Render("(recent)")
			}
			paperInfo := fmt.Sprintf("   %d papers", len(s.Papers))

			if hl {
				b.WriteString(selectedItemStyle.Render("▸ " + s.Code + tag))
			} else {
				b.WriteString(itemStyle.Render(line + tag))
			}
			b.WriteString("\n")
			b.WriteString(infoStyle.Render(paperInfo))
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑↓  navigate"))
	b.WriteString("    ")
	b.WriteString(helpStyle.Render("enter  select"))
	b.WriteString("    ")
	b.WriteString(helpStyle.Render("esc  quit"))
	return docStyle.Render(b.String())
}

// ─── select view ──────────────────────────────────────────

func (m model) renderSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.selSubject.Code))
	b.WriteString("\n\n")

	homeDir, _ := os.UserHomeDir()
	dlBase := filepathJoin(homeDir, "Downloads", "Question Papers", m.selSubject.Code)

	for i, paper := range m.selSubject.Papers {
		prefix := uncheckedBox
		if m.selected[i] {
			prefix = checkedBox
		}
		cacheMark := ""
		if _, err := os.Stat(filepathJoin(dlBase, fmt.Sprintf("%s_%s.pdf", m.selSubject.Code, paper.Session))); err == nil {
			cacheMark = " " + successStyle.Render("(cached)")
		}

		if i == m.detailCursor {
			b.WriteString(selectedItemStyle.Render(prefix + paper.Session + cacheMark))
		} else {
			b.WriteString(itemStyle.Render(prefix + paper.Session + cacheMark))
		}
		b.WriteString("\n")
	}

	selCount := 0
	for _, v := range m.selected {
		if v {
			selCount++
		}
	}

	b.WriteString("\n")
	b.WriteString(infoStyle.Render(fmt.Sprintf("%d/%d selected", selCount, len(m.selSubject.Papers))))
	b.WriteString("\n\n")
	b.WriteString(separator)
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑↓  navigate"))
	b.WriteString("    ")
	b.WriteString(helpStyle.Render("space  toggle"))
	b.WriteString("    ")
	b.WriteString(helpStyle.Render("a  all/none"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter  download"))
	b.WriteString("    ")
	b.WriteString(helpStyle.Render("esc  back"))
	return docStyle.Render(b.String())
}

// ─── download view ────────────────────────────────────────

func (m model) renderDownload() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Downloading " + m.selSubject.Code))
	b.WriteString("\n\n")

	doneCount := 0
	for i, paper := range m.selSubject.Papers {
		if !m.selected[i] {
			continue
		}

		line := ""
		if err, ok := m.failed[paper.Session]; ok {
			line = errorStyle.Render(fmt.Sprintf("✗  %s — %s", paper.Session, err))
		} else if m.done[paper.Session] {
			line = successStyle.Render(fmt.Sprintf("✓  %s", paper.Session))
			doneCount++
		} else if pct, ok := m.progress[paper.Session]; ok {
			bar := renderBar(pct, 20)
			line = fmt.Sprintf("%s  %s  %s %0.0f%%", m.spinner.View(), paper.Session, bar, pct*100)
		} else {
			line = dimStyle.Render(fmt.Sprintf("⏳  %s", paper.Session))
		}
		b.WriteString("  " + line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	total := len(m.jobs)
	if m.finished {
		b.WriteString(successStyle.Render(fmt.Sprintf("Done! %d downloaded, %d failed.", doneCount, m.errors)))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc  back"))
		b.WriteString("    ")
		b.WriteString(helpStyle.Render("q  quit"))
	} else {
		b.WriteString(infoStyle.Render(fmt.Sprintf("%d/%d complete", doneCount, total)))
	}
	return docStyle.Render(b.String())
}

// ─── helpers ──────────────────────────────────────────────

func filepathJoin(elems ...string) string {
	p := elems[0]
	for _, e := range elems[1:] {
		p += string(os.PathSeparator) + e
	}
	return p
}

func renderBar(pct float64, w int) string {
	if pct > 1 {
		pct = 1
	}
	f := int(pct * float64(w))
	return strings.Repeat("█", f) + strings.Repeat("░", w-f)
}

func (m *model) applyFilter() {
	q := strings.ToUpper(strings.TrimSpace(m.searchInput.Value()))
	if q == "" {
		m.subjects = m.allSubjects
	} else {
		var filtered []subjectEntry
		for _, s := range m.allSubjects {
			if strings.Contains(strings.ToUpper(s.Code), q) {
				filtered = append(filtered, s)
			}
		}
		m.subjects = filtered
	}
	if m.cursor >= len(m.subjects) {
		m.cursor = 0
	}
}

func (m *model) selectSubject(code string) {
	code = strings.ToUpper(strings.TrimSpace(code))
	for _, s := range m.allSubjects {
		if strings.EqualFold(s.Code, code) {
			sCopy := s
			m.selSubject = &sCopy
			m.selected = make(map[int]bool)
			for i := range sCopy.Papers {
				m.selected[i] = true
			}
			m.detailCursor = 0
			return
		}
	}
	if len(m.subjects) > 0 {
		sCopy := m.subjects[m.cursor]
		m.selSubject = &sCopy
		m.selected = make(map[int]bool)
		for i := range sCopy.Papers {
			m.selected[i] = true
		}
		m.detailCursor = 0
	}
}

func (m *model) buildJobs() {
	m.jobs = nil
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	targetDir := filepathJoin(homeDir, "Downloads", "Question Papers", m.selSubject.Code)
	os.MkdirAll(targetDir, os.ModePerm)
	m.progress = make(map[string]float64)
	m.done = make(map[string]bool)
	m.failed = make(map[string]error)
	m.finished = false
	m.errors = 0

	for i, paper := range m.selSubject.Papers {
		if m.selected[i] {
			target := filepathJoin(targetDir, fmt.Sprintf("%s_%s.pdf", m.selSubject.Code, paper.Session))
			m.jobs = append(m.jobs, downloadJob{
				session:  paper.Session,
				fileID:   paper.FileID,
				filePath: target,
			})
		}
	}
}

func (m *model) startDownloads() {
	code := m.selSubject.Code
	go func() {
		for _, job := range m.jobs {
			url := fmt.Sprintf("https://docs.google.com/uc?export=download&id=%s", job.fileID)
			resp, err := http.Get(url)
			if err != nil {
				m.program.Send(progressMsg{session: job.session, err: fmt.Errorf("network: %w", err)})
				continue
			}

			out, err := os.Create(job.filePath)
			if err != nil {
				resp.Body.Close()
				m.program.Send(progressMsg{session: job.session, err: fmt.Errorf("file: %w", err)})
				continue
			}

			total := resp.ContentLength
			var written int64
			buf := make([]byte, 32*1024)

			for {
				n, rerr := resp.Body.Read(buf)
				if n > 0 {
					if _, werr := out.Write(buf[:n]); werr != nil {
						m.program.Send(progressMsg{session: job.session, err: werr})
						break
					}
					written += int64(n)
					if total > 0 {
						m.program.Send(progressMsg{session: job.session, percent: float64(written) / float64(total)})
					}
				}
				if rerr != nil {
					if rerr != io.EOF {
						m.program.Send(progressMsg{session: job.session, err: rerr})
					}
					break
				}
			}
			out.Close()
			resp.Body.Close()
			m.program.Send(progressMsg{session: job.session, percent: 1.0, complete: true})
		}
		m.program.Send(downloadDoneMsg{})
		recordRecent(code)
	}()
}
