package report

import "time"

// RunData holds all metrics collected from a single scenario run.
type RunData struct {
	RunID      string         `json:"run_id"`
	Scenario   string         `json:"scenario"`
	StartedAt  time.Time      `json:"started_at"`
	Duration   time.Duration  `json:"duration"`
	Config     RunConfig      `json:"config"`
	Requests   int            `json:"requests"`
	Successes  int            `json:"successes"`
	Failures   int            `json:"failures"`
	Latencies  LatencyStats   `json:"latencies"`
	StatusDist map[int]int    `json:"status_dist"`
	Timeseries []TimeseriesDP `json:"timeseries"`
	DBStats    *DBStatsSnap   `json:"db_stats,omitempty"`
	HPAStats   *HPASnap       `json:"hpa_stats,omitempty"`
	BatchStats *BatchSnap     `json:"batch_stats,omitempty"`
	Score      int            `json:"score"`
	ScoreLine  string         `json:"score_line"`
}

// RunConfig stores the configuration used for a scenario run.
type RunConfig struct {
	TargetURL   string        `json:"target_url"`
	RPS         int           `json:"rps"`
	Duration    time.Duration `json:"duration"`
	Concurrency int           `json:"concurrency"`
}

// LatencyStats holds latency percentiles in milliseconds.
type LatencyStats struct {
	P50 float64 `json:"p50"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
	Max float64 `json:"max"`
	Avg float64 `json:"avg"`
}

// TimeseriesDP is a single data point in the time series.
type TimeseriesDP struct {
	Elapsed    float64 `json:"elapsed_s"`
	RPS        float64 `json:"rps"`
	LatencyP95 float64 `json:"latency_p95_ms"`
	ErrorRate  float64 `json:"error_rate"`
}

// DBStatsSnap holds a snapshot of database pool statistics.
type DBStatsSnap struct {
	MaxOpen      int    `json:"max_open"`
	Open         int    `json:"open"`
	InUse        int    `json:"in_use"`
	Idle         int    `json:"idle"`
	WaitCount    int64  `json:"wait_count"`
	WaitDuration string `json:"wait_duration"`
}

// HPASnap holds HPA autoscaler status.
type HPASnap struct {
	DesiredReplicas int `json:"desired_replicas"`
	CurrentReplicas int `json:"current_replicas"`
	MinReplicas     int `json:"min_replicas"`
	MaxReplicas     int `json:"max_replicas"`
}

// BatchSnap holds worker batch completion stats.
type BatchSnap struct {
	Total    int     `json:"total"`
	Done     int     `json:"done"`
	FastP95  float64 `json:"fast_p95_ms"`
	SlowP95  float64 `json:"slow_p95_ms"`
	Elapsed  float64 `json:"elapsed_ms"`
	Complete bool    `json:"complete"`
}
