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
	"github.com/charmbracelet/lipgloss"
	"github.com/IshpreetSingh8264/gndec-qp/db"
)

var globalProg *tea.Program

func Run() {
	m := initialModel()
	m.state = stateSearch
	p := tea.NewProgram(m, tea.WithAltScreen())
	globalProg = p
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
	globalProg = p
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// ─── model ────────────────────────────────────────────────

type model struct {
	state    state
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
		m.setInputWidth()
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

// ─── rendering helpers ────────────────────────────────────

func (m model) innerW() int { return m.width - 6 }

func (m model) lineCount(s string) int {
	return strings.Count(s, "\n")
}

func (m model) fit(heightBudget int) int {
	n := m.height - heightBudget
	if n < 1 {
		return 1
	}
	return n
}

func (m model) row(left, right string, highlight bool) string {
	// Build a single row that exactly fills m.innerW() width.
	// Format: " left_part                      right_part"
	padW := m.innerW() - 2 - lipgloss.Width(left) - lipgloss.Width(right)
	if padW < 1 {
		left = trunc(left, m.innerW()-2-lipgloss.Width(right)-1)
		padW = m.innerW() - 2 - lipgloss.Width(left) - lipgloss.Width(right)
	}
	if padW < 0 {
		padW = 0
	}
	content := " " + left + strings.Repeat(" ", padW) + right
	if highlight {
		return selectedRow.Render(content)
	}
	return content
}

func trunc(s string, maxW int) string {
	runes := []rune(s)
	if len(runes) <= maxW {
		return s
	}
	return string(runes[:maxW-1]) + "…"
}

func (m *model) setInputWidth() {
	m.searchInput.Width = m.innerW() - 4
	if m.searchInput.Width < 10 {
		m.searchInput.Width = 10
	}
}

func (m model) footer(status string, keys ...string) string {
	w := m.innerW() - 2
	status = trunc(status, w)
	keyLine := trunc(strings.Join(keys, " · "), w)
	return "\n" + hLine(m.innerW()) + "\n" +
		pad1.Render(status) + "\n" +
		pad1.Render(keyLine)
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
		maxVis := m.fit(17)
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
			code := s.Code
			tag := ""
			if recMap[s.Code] {
				tag = dimStyle.Render(" recent")
			}
			right := fmt.Sprintf("%d papers", len(s.Papers))
			b.WriteString(m.row("▸ "+code+tag, right, i == m.cursor))
			b.WriteString("\n")
		}

		if len(m.subjects) > maxVis {
			if start > 0 {
				b.WriteString(infoStyle.Render(" ▲ more above\n"))
			}
			b.WriteString(infoStyle.Render(fmt.Sprintf(" %d – %d of %d", start+1, end, len(m.subjects))))
			if end < len(m.subjects) {
				b.WriteString(infoStyle.Render(" · ▼ more below"))
			}
			b.WriteString("\n")
		}
	}

	hlp := fmt.Sprintf("%d matches", len(m.subjects))
	f := m.footer(hlp, "↑↓ navigate", "enter to select", "esc to quit")
	b.WriteString(f)
	return borderTop(m.innerW()) + "\n" + m.wrapLines(b.String()) + "\n" + borderBottom(m.innerW())
}

// ─── select view ──────────────────────────────────────────

func (m model) renderSelect() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.selSubject.Code))
	b.WriteString("\n\n")

	homeDir, _ := os.UserHomeDir()
	dlBase := filepathJoin(homeDir, "Downloads", "Question Papers", m.selSubject.Code)

	totalPapers := len(m.selSubject.Papers)
	maxVis := m.fit(13)
	selStart := (m.detailCursor / maxVis) * maxVis
	if selStart > totalPapers-maxVis {
		selStart = totalPapers - maxVis
	}
	if selStart < 0 {
		selStart = 0
	}
	selEnd := selStart + maxVis
	if selEnd > totalPapers {
		selEnd = totalPapers
	}

	for i := selStart; i < selEnd; i++ {
		paper := m.selSubject.Papers[i]
		prefix := uncheckedBox
		if m.selected[i] {
			prefix = checkedBox
		}

		year := formatSession(paper.Session)
		cached := ""
		paperPath := filepathJoin(dlBase, fmt.Sprintf("%s_%s.pdf", m.selSubject.Code, paper.Session))
		if _, err := os.Stat(paperPath); err == nil {
			cached = successStyle.Render("cached")
		}

		b.WriteString(m.row(prefix+year, cached, i == m.detailCursor))
		b.WriteString("\n")
	}

	selCount := 0
	for _, v := range m.selected {
		if v {
			selCount++
		}
	}

	if totalPapers > maxVis {
		if selStart > 0 {
			b.WriteString(infoStyle.Render(" ▲ more above\n"))
		}
		b.WriteString(infoStyle.Render(fmt.Sprintf(" %d – %d of %d", selStart+1, selEnd, totalPapers)))
		if selEnd < totalPapers {
			b.WriteString(infoStyle.Render(" · ▼ more below"))
		}
		b.WriteString("\n")
	}

	f := m.footer(fmt.Sprintf("%d/%d selected", selCount, totalPapers), "↑↓ navigate", "space toggle", "a all/none", "enter to download", "esc back")
	b.WriteString(f)
	return borderTop(m.innerW()) + "\n" + m.wrapLines(b.String()) + "\n" + borderBottom(m.innerW())
}

// ─── download view ────────────────────────────────────────

func (m model) renderDownload() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Downloading " + m.selSubject.Code))
	b.WriteString("\n\n")

	availW := m.innerW() - 6
	barW := availW / 5
	if barW > 30 {
		barW = 30
	}
	if barW < 8 {
		barW = 8
	}

	doneCount := 0
	for i, paper := range m.selSubject.Papers {
		if !m.selected[i] {
			continue
		}

		year := formatSession(paper.Session)
		session := trunc(year, 12)

		if err, ok := m.failed[paper.Session]; ok {
			b.WriteString(m.row(errorStyle.Render("✗ "+session), dimStyle.Render(trunc(err.Error(), 20)), false))
		} else if m.done[paper.Session] {
			doneCount++
			b.WriteString(m.row(successStyle.Render("✓ "+session), "", false))
		} else if pct, ok := m.progress[paper.Session]; ok {
			pb := bar(barW, pct)
			status := fmt.Sprintf("%s %s %s %0.0f%%", m.spinner.View(), session, pb, pct*100)
			b.WriteString(m.row(status, "", false))
		} else {
			b.WriteString(m.row(dimStyle.Render("⏳ "+session), "", false))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	total := len(m.jobs)
	if m.finished {
		f := m.footer(fmt.Sprintf("%d downloaded, %d failed", doneCount, m.errors), "esc to search", "q to quit")
		b.WriteString(f)
	} else {
		f := m.footer(fmt.Sprintf("%d/%d complete", doneCount, total), "ctrl+c to cancel")
		b.WriteString(f)
	}
	return borderTop(m.innerW()) + "\n" + m.wrapLines(b.String()) + "\n" + borderBottom(m.innerW())
}

func (m model) wrapLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = borderSides(l)
	}
	return strings.Join(lines, "\n")
}

// ─── helpers ──────────────────────────────────────────────

func filepathJoin(elems ...string) string {
	p := elems[0]
	for _, e := range elems[1:] {
		p += string(os.PathSeparator) + e
	}
	return p
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
				globalProg.Send(progressMsg{session: job.session, err: fmt.Errorf("network: %w", err)})
				continue
			}

			out, err := os.Create(job.filePath)
			if err != nil {
				resp.Body.Close()
				globalProg.Send(progressMsg{session: job.session, err: fmt.Errorf("file: %w", err)})
				continue
			}

			total := resp.ContentLength
			var written int64
			buf := make([]byte, 32*1024)

			for {
				n, rerr := resp.Body.Read(buf)
				if n > 0 {
					if _, werr := out.Write(buf[:n]); werr != nil {
						globalProg.Send(progressMsg{session: job.session, err: werr})
						break
					}
					written += int64(n)
					if total > 0 {
						globalProg.Send(progressMsg{session: job.session, percent: float64(written) / float64(total)})
					}
				}
				if rerr != nil {
					if rerr != io.EOF {
						globalProg.Send(progressMsg{session: job.session, err: rerr})
					}
					break
				}
			}
			out.Close()
			resp.Body.Close()
			globalProg.Send(progressMsg{session: job.session, percent: 1.0, complete: true})
		}
		globalProg.Send(downloadDoneMsg{})
		recordRecent(code)
	}()
}
