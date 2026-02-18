package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Generate writes a report for the given run data to the reports directory.
func Generate(data *RunData, reportsDir string) (string, error) {
	runDir := filepath.Join(reportsDir, data.Scenario, data.RunID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return "", fmt.Errorf("creating report dir: %w", err)
	}

	dataPath := filepath.Join(runDir, "data.json")
	f, err := os.Create(dataPath)
	if err != nil {
		return "", fmt.Errorf("creating data.json: %w", err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		f.Close()
		return "", fmt.Errorf("writing data.json: %w", err)
	}
	f.Close()

	reportPath := filepath.Join(runDir, "report.html")
	tmpl, err := template.New("report").Parse(reportHTMLTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing report template: %w", err)
	}
	rf, err := os.Create(reportPath)
	if err != nil {
		return "", fmt.Errorf("creating report.html: %w", err)
	}
	if err := tmpl.Execute(rf, data); err != nil {
		rf.Close()
		return "", fmt.Errorf("rendering report.html: %w", err)
	}
	rf.Close()

	if err := updateIndex(reportsDir, data); err != nil {
		log.Printf("warning: could not update index: %v", err)
	}

	return reportPath, nil
}

// OpenReport opens a report in the default browser (macOS only).
func OpenReport(path string) {
	if runtime.GOOS == "darwin" {
		exec.Command("open", path).Start()
	}
}

func updateIndex(reportsDir string, latest *RunData) error {
	indexPath := filepath.Join(reportsDir, "index.html")
	type entry struct {
		Scenario string
		RunID    string
		Time     string
		Score    int
		Link     string
	}
	entries := []entry{
		{
			Scenario: latest.Scenario,
			RunID:    latest.RunID,
			Time:     latest.StartedAt.Format(time.RFC3339),
			Score:    latest.Score,
			Link:     filepath.Join(latest.Scenario, latest.RunID, "report.html"),
		},
	}
	tmpl, err := template.New("index").Parse(indexHTMLTemplate)
	if err != nil {
		return err
	}
	fi, err := os.Create(indexPath)
	if err != nil {
		return err
	}
	defer fi.Close()
	return tmpl.Execute(fi, entries)
}
