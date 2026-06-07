package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mayank0304/gndec-qp/db"
	"github.com/mayank0304/gndec-qp/open"
)

func Fetch(subjectCode string, auto bool) {
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
			fmt.Printf("\n	Network error while downloading this session... %s", err)
			continue
		}

		out, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("\n	Failed to create file on disk: %v\n", err)
			resp.Body.Close()
			continue
		}

		_, err = io.Copy(out, resp.Body)
		out.Close()
		resp.Body.Close()

		if err != nil {
			fmt.Printf("\n	Error writing file bytes to disk: %v\n", err)
			continue
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
