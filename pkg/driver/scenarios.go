package driver

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// Scenario defines a load test configuration for a specific lab case.
type Scenario struct {
	Name        string
	Description string
	TargetURL   string
	Method      string
	Body        string
	RPS         int
	Duration    time.Duration
	Concurrency int
	MaxP95Ms    float64
	MaxErrRate  float64
	DBStatsURL  string
	HPAStatsURL string
	BatchURL    string
	// FaultFunc is called mid-run to inject a failure (e.g. kubectl drain).
	// If nil, no fault is injected.
	FaultFunc func(ctx context.Context) error
	// FaultDelay is how long after the run starts before FaultFunc fires.
	FaultDelay time.Duration
}

// Registry maps scenario names to their configs.
var Registry = map[string]*Scenario{
	"timeouts": {
		Name:        "timeouts",
		Description: "Case 1: Timeout patterns — calls to slow dependency without deadline",
		TargetURL:   "http://localhost:8080/cases/timeouts",
		Method:      "GET",
		RPS:         10,
		Duration:    30 * time.Second,
		Concurrency: 20,
		MaxP95Ms:    2500,
		MaxErrRate:  0.1,
	},
	"tx": {
		Name:        "tx",
		Description: "Case 2: DB transaction scope — holding TX across network calls",
		TargetURL:   "http://localhost:8080/cases/tx",
		Method:      "GET",
		RPS:         10,
		Duration:    30 * time.Second,
		Concurrency: 20,
		MaxP95Ms:    3000,
		MaxErrRate:  0.1,
		DBStatsURL:  "http://localhost:8080/debug/dbstats",
	},
	"bulkheads": {
		Name:        "bulkheads",
		Description: "Case 3: Bulkhead pattern — shared pool starvation",
		TargetURL:   "http://localhost:8081/batches",
		Method:      "POST",
		Body:        `{"fast": 100, "slow": 20}`,
		RPS:         1,
		Duration:    10 * time.Second,
		Concurrency: 5,
		MaxP95Ms:    500,
		MaxErrRate:  0.05,
		BatchURL:    "http://localhost:8081/batches",
	},
	"autoscale": {
		Name:        "autoscale",
		Description: "Case 4: Autoscaling — CPU-bound without HPA",
		TargetURL:   "http://localhost:8080/cases/autoscale",
		Method:      "GET",
		RPS:         20,
		Duration:    60 * time.Second,
		Concurrency: 30,
		MaxP95Ms:    5000,
		MaxErrRate:  0.1,
	},
	"pdb": {
		Name:        "pdb",
		Description: "Case 5: PDB & failover — CNPG primary killed mid-run",
		TargetURL:   "http://localhost:8080/cases/pdb",
		Method:      "GET",
		RPS:         10,
		Duration:    60 * time.Second,
		Concurrency: 10,
		MaxP95Ms:    2000,
		MaxErrRate:  0.05,
		FaultDelay:  15 * time.Second,
		FaultFunc:   drainCNPGPrimaryNode,
	},
}

// ListScenarios returns all scenario names.
func ListScenarios() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	return names
}

// drainCNPGPrimaryNode finds the CNPG primary pod's node and drains it.
func drainCNPGPrimaryNode(ctx context.Context) error {
	// Find the primary pod's node.
	out, err := exec.CommandContext(ctx, "kubectl", "get", "pods",
		"-l", "cnpg.io/cluster=workshop-pg,role=primary",
		"-o", "jsonpath={.items[0].spec.nodeName}",
	).Output()
	if err != nil {
		return fmt.Errorf("find primary node: %w", err)
	}
	node := strings.TrimSpace(string(out))
	if node == "" {
		return fmt.Errorf("no primary node found")
	}

	log.Printf("fault: draining node %s (CNPG primary)", node)
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "kubectl", "drain", node,
		"--ignore-daemonsets", "--delete-emptydir-data", "--force",
		"--grace-period=5", "--timeout=30s",
	)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("drain node %s: %w: %s", node, err, stderr.String())
	}

	log.Printf("fault: node %s drained", node)
	return nil
}

// UncordonAllNodes uncordons all nodes so the cluster recovers.
func UncordonAllNodes() {
	out, err := exec.Command("kubectl", "get", "nodes",
		"-o", "jsonpath={.items[*].metadata.name}",
	).Output()
	if err != nil {
		log.Printf("fault: uncordon list nodes: %v", err)
		return
	}
	for _, node := range strings.Fields(string(out)) {
		if err := exec.Command("kubectl", "uncordon", node).Run(); err != nil {
			log.Printf("fault: uncordon %s: %v", node, err)
		} else {
			log.Printf("fault: uncordoned %s", node)
		}
	}
}
