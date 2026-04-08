package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/infobloxopen/architecture-workshops2/pkg/driver"
	"github.com/infobloxopen/architecture-workshops2/pkg/report"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: driver run <scenario>")
			fmt.Fprintf(os.Stderr, "Available: %s\n", strings.Join(driver.ListScenarios(), ", "))
			os.Exit(1)
		}
		runScenario(os.Args[2])
	case "list":
		for _, s := range driver.ListScenarios() {
			sc := driver.Registry[s]
			fmt.Printf("  %-12s %s\n", s, sc.Description)
		}
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: driver <command>")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  run <scenario>   Run a scenario and generate report")
	fmt.Fprintln(os.Stderr, "  list             List available scenarios")
}

func runScenario(name string) {
	scenario, ok := driver.Registry[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown scenario: %s\n", name)
		fmt.Fprintf(os.Stderr, "Available: %s\n", strings.Join(driver.ListScenarios(), ", "))
		os.Exit(1)
	}
	fmt.Printf("==> Running scenario: %s\n", scenario.Name)
	fmt.Printf("    %s\n", scenario.Description)
	fmt.Printf("    Target: %s | RPS: %d | Duration: %s\n",
		scenario.TargetURL, scenario.RPS, scenario.Duration)
	fmt.Println()

	runner := driver.NewRunner(driver.RunConfig{
		TargetURL:   scenario.TargetURL,
		Method:      scenario.Method,
		Body:        scenario.Body,
		RPS:         scenario.RPS,
		Duration:    scenario.Duration,
		Concurrency: scenario.Concurrency,
		FaultFunc:   scenario.FaultFunc,
		FaultDelay:  scenario.FaultDelay,
	})
	ctx := context.Background()
	data := runner.Run(ctx)
	data.Scenario = scenario.Name

	// Cleanup after fault injection (e.g. uncordon drained nodes).
	if scenario.FaultFunc != nil {
		driver.UncordonAllNodes()
	}

	score, scoreLine := driver.Score(data, scenario)
	data.Score = score
	data.ScoreLine = scoreLine
	fmt.Println(scoreLine)
	fmt.Println()

	reportsDir := "reports"
	reportPath, err := report.Generate(data, reportsDir)
	if err != nil {
		log.Fatalf("Failed to generate report: %v", err)
	}
	fmt.Printf("==> Report: %s\n", reportPath)
	report.OpenReport(reportPath)
}
