package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

func configDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	dir := filepath.Join(home, ".config", "gndec-qp")
	os.MkdirAll(dir, 0755)
	return dir
}

func recentFile() string {
	return filepath.Join(configDir(), "recent.json")
}

func loadRecent() []recentEntry {
	data, err := os.ReadFile(recentFile())
	if err != nil {
		return nil
	}
	var entries []recentEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}
	return entries
}

func recordRecent(code string) {
	entries := loadRecent()
	if entries == nil {
		entries = []recentEntry{}
	}

	found := false
	for i, e := range entries {
		if e.Code == code {
			entries[i].Downloads++
			entries[i].LastUsed = time.Now().Format(time.RFC3339)
			found = true
			break
		}
	}

	if !found {
		entries = append([]recentEntry{{
			Code:      code,
			Downloads: 1,
			LastUsed:  time.Now().Format(time.RFC3339),
		}}, entries...)
	}

	if len(entries) > 10 {
		entries = entries[:10]
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(recentFile(), data, 0644)
}


