package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/IshpreetSingh8264/gndec-qp/downloader"
	"github.com/IshpreetSingh8264/gndec-qp/tui"
)

func main() {
	var codeFlag string
	flag.StringVar(&codeFlag, "code", "", "The Subject Code to search for (e.g. PCIT-114)")

	var auto bool
	flag.BoolVar(&auto, "auto", false, "To open automatically in the browser / system viewer")

	var tuiFlag bool
	flag.BoolVar(&tuiFlag, "tui", false, "Launch the interactive TUI")

	flag.Parse()

	subjectCode := strings.ToUpper(strings.TrimSpace(codeFlag))

	if tuiFlag || subjectCode == "" {
		if subjectCode != "" {
			tui.RunWithCode(subjectCode)
		} else {
			tui.Run()
		}
		return
	}

	if subjectCode == "" {
		fmt.Println("Error: Please provide a subject code using the --code flag.")
		fmt.Println("Example usage: qp --code=PCIT-114")
		return
	}

	downloader.Fetch(subjectCode, auto)
}
