package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/IshpreetSingh8264/gndec-qp/db"
	"github.com/IshpreetSingh8264/gndec-qp/open"
)

type ProgressFn func(session string, percent float64, err error)

func Fetch(subjectCode string, auto bool) {
	FetchWithProgress(subjectCode, auto, nil)
}

func FetchWithProgress(subjectCode string, auto bool, onProgress ProgressFn) {
	papers, exists := db.PaperRegistry[subjectCode]
	if !exists {
		fmt.Printf("There are no question papers with this code: %s", subjectCode)
		return
	}

	homeDir, ok := getUserDir()
	var targetDir string
	if !ok {
		targetDir = filepath.Join(".", subjectCode)
	} else {
		targetDir = filepath.Join(homeDir, "Downloads", subjectCode)
	}

	err := os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create the Directory %s because %s", targetDir, err)
		return
	}

	fmt.Println("\n======================================================")
	fmt.Printf("🎉 YAY! There are %d papers for %s\n", len(papers), subjectCode)
	fmt.Printf("📂 Saving downloads to: %s\n", targetDir)
	fmt.Println("======================================================")

	for _, paper := range papers {
		baseFileName := fmt.Sprintf("%s_%s.pdf", subjectCode, paper.Session)
		filePath := filepath.Join(targetDir, baseFileName)

		logStr := fmt.Sprintf("📥 Downloading %s...", baseFileName)
		fmt.Printf("%-60s", logStr)

		downloadUrl := fmt.Sprintf("https://docs.google.com/uc?export=download&id=%s", paper.FileID)

		resp, err := http.Get(downloadUrl)
		if err != nil {
			msg := fmt.Sprintf("\n\tNetwork error while downloading this session... %s", err)
			fmt.Print(msg)
			if onProgress != nil {
				onProgress(paper.Session, 0, fmt.Errorf("network error: %w", err))
			}
			continue
		}

		out, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("\n\tFailed to create file on disk: %v\n", err)
			resp.Body.Close()
			if onProgress != nil {
				onProgress(paper.Session, 0, fmt.Errorf("file error: %w", err))
			}
			continue
		}

		total := resp.ContentLength
		var written int64
		buf := make([]byte, 32*1024)
		downloadErr := error(nil)

		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				wn, writeErr := out.Write(buf[:n])
				if writeErr != nil {
					downloadErr = writeErr
					break
				}
				written += int64(wn)
				if total > 0 && onProgress != nil {
					onProgress(paper.Session, float64(written)/float64(total), nil)
				}
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				downloadErr = readErr
				break
			}
		}

		out.Close()
		resp.Body.Close()

		if downloadErr != nil {
			fmt.Printf("\n\tError writing file bytes to disk: %v\n", downloadErr)
			if onProgress != nil {
				onProgress(paper.Session, 0, downloadErr)
			}
			continue
		}

		if onProgress != nil {
			onProgress(paper.Session, 1.0, nil)
		}

		fmt.Println("	✅ Complete!")

		if auto {
			open.Open(filePath)
		}
	}
	fmt.Println("======================================================")
	fmt.Println("🚀 All operations finalized successfully!")
	if auto {
		fmt.Println("💡 Default protocol system handlers triggered automatically.")
	}
	fmt.Println("======================================================")
}

func getUserDir() (string, bool) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", false
	}
	return homeDir, true
}
