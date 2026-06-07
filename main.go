package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mayank0304/gndec-qp/downloader"
)

func main() {
	var codeFlag string
	flag.StringVar(&codeFlag, "code", "", "The Subject Code to search for (e.g. PCIT-114)")

	var auto bool
	flag.BoolVar(&auto, "auto", false, "To open automatically in the browser / system viewer")

	flag.Parse()

	subjectCode := strings.ToUpper(strings.TrimSpace(codeFlag))

	if subjectCode == "" {
		fmt.Println("Error: Please provide a subject code using the --code flag.")
		fmt.Println("Example usage: qp --code=PCIT-114")
		return
	}

	downloader.Fetch(subjectCode, auto)
}
