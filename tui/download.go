package tui

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) downloadView() string {
	var b strings.Builder

	if m.downloading {
		b.WriteString(subtitleStyle.Render(fmt.Sprintf("Downloading %s...", m.selectedSubject.Code)))
	} else {
		b.WriteString(subtitleStyle.Render("Download Complete"))
	}
	b.WriteString("\n\n")

	for _, p := range m.progress {
		if p.err != nil {
			b.WriteString(errorStyle.Render(fmt.Sprintf("✗ %s: %s", p.session, p.err)))
			b.WriteString("\n")
		} else if p.complete {
			b.WriteString(successStyle.Render(fmt.Sprintf("✓ %s", p.session)))
			b.WriteString("\n")
		} else {
			b.WriteString(fmt.Sprintf("%s %s", m.spinner.View(), p.session))
			b.WriteString("\n")
			bar := renderProgressBar(p.percent, 30)
			b.WriteString(fmt.Sprintf("   %s %0.0f%%", bar, p.percent*100))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(separator)
	b.WriteString("\n")

	if m.downloading {
		b.WriteString(infoStyle.Render(fmt.Sprintf("Progress: %d/%d", completedCount(m.progress), len(m.progress))))
	} else {
		b.WriteString(infoStyle.Render(fmt.Sprintf("Completed: %d/%d", m.completed, m.total)))
		b.WriteString("\n")
		if len(m.errs) > 0 {
			b.WriteString(warningStyle.Render(fmt.Sprintf("%d downloads failed", len(m.errs))))
		} else {
			b.WriteString(successStyle.Render("All downloads successful!"))
		}
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc  back to detail"))
		b.WriteString("  ")
		b.WriteString(helpStyle.Render("q  home"))
	}

	b.WriteString("\n")

	return docStyle.Render(b.String())
}

func renderProgressBar(percent float64, width int) string {
	if percent > 1 {
		percent = 1
	}
	filled := int(percent * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return bar
}

func completedCount(progress map[string]jobProgress) int {
	count := 0
	for _, p := range progress {
		if p.complete {
			count++
		}
	}
	return count
}

func (m model) startDownloads() (tea.Model, tea.Cmd) {
	go func() {
		for _, job := range m.queue {
			downloadURL := fmt.Sprintf("https://docs.google.com/uc?export=download&id=%s", job.fileID)

			resp, err := http.Get(downloadURL)
			if err != nil {
				m.program.Send(progressMsg{
					session: job.session,
					percent: 0,
					err:     fmt.Errorf("network error: %w", err),
				})
				continue
			}

			out, err := os.Create(job.filePath)
			if err != nil {
				resp.Body.Close()
				m.program.Send(progressMsg{
					session: job.session,
					percent: 0,
					err:     fmt.Errorf("file error: %w", err),
				})
				continue
			}

			total := resp.ContentLength
			var written int64
			buf := make([]byte, 32*1024)
			for {
				n, readErr := resp.Body.Read(buf)
				if n > 0 {
					wn, writeErr := out.Write(buf[:n])
					if writeErr != nil {
						break
					}
					written += int64(wn)
					if total > 0 {
						m.program.Send(progressMsg{
							session: job.session,
							percent: float64(written) / float64(total),
						})
					}
				}
				if readErr == io.EOF {
					break
				}
				if readErr != nil {
					m.program.Send(progressMsg{
						session: job.session,
						percent: 0,
						err:     fmt.Errorf("read error: %w", readErr),
					})
					break
				}
			}

			out.Close()
			resp.Body.Close()

			m.program.Send(progressMsg{
				session:  job.session,
				percent:  1.0,
				complete: true,
			})
		}

		errors := []string{}
		for _, p := range m.progress {
			if p.err != nil {
				errors = append(errors, p.session)
			}
		}
		m.program.Send(downloadDoneMsg{
			completed: completedCount(m.progress),
			total:     len(m.progress),
			errors:    errors,
		})

		if m.selectedSubject != nil {
			recordRecent(m.selectedSubject.Code)
		}
	}()

	return m, nil
}

func (m model) handleDownloadKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return m, nil
}
