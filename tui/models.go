package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/IshpreetSingh8264/gndec-qp/db"
)

type screen int

const (
	screenHome screen = iota
	screenSearch
	screenDetail
	screenDownload
)

type subjectEntry struct {
	Code   string
	Papers []db.Paper
}

type downloadJob struct {
	subjectCode string
	session     string
	fileID      string
	filePath    string
}

type jobProgress struct {
	session  string
	percent  float64
	complete bool
	err      error
}

type progressMsg struct {
	session  string
	percent  float64
	complete bool
	err      error
}

type downloadDoneMsg struct {
	completed int
	total     int
	errors    []string
}

type startDownloadMsg struct{}

type recentEntry struct {
	Code      string `json:"code"`
	Downloads int    `json:"downloads"`
	LastUsed  string `json:"last_used"`
}

type setProgramMsg struct {
	program *tea.Program
}
