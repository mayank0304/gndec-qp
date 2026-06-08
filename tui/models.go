package tui

import "github.com/IshpreetSingh8264/gndec-qp/db"

type state int

const (
	stateSearch state = iota
	stateSelect
	stateDownload
)

type subjectEntry struct {
	Code   string
	Papers []db.Paper
}

type downloadJob struct {
	session  string
	fileID   string
	filePath string
}

type progressMsg struct {
	session  string
	percent  float64
	complete bool
	err      error
}

type downloadDoneMsg struct{}

type recentEntry struct {
	Code      string `json:"code"`
	Downloads int    `json:"downloads"`
	LastUsed  string `json:"last_used"`
}
