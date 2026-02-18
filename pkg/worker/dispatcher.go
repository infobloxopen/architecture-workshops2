package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Batch represents a submitted batch of work items.
type Batch struct {
	ID        string      `json:"id"`
	Fast      int         `json:"fast"`
	Slow      int         `json:"slow"`
	StartedAt time.Time   `json:"started_at"`
	Results   []JobResult `json:"-"`
	Done      atomic.Int32
	Total     int `json:"total"`
	mu        sync.Mutex
}

// JobResult records completion of a single job.
type JobResult struct {
	Type     string        `json:"type"`
	Duration time.Duration `json:"duration"`
}

var (
	batches   = map[string]*Batch{}
	batchesMu sync.RWMutex
	batchSeq  atomic.Int64
)

// Run starts the worker service on :8081.
func Run() {
	port := envOr("WORKER_PORT", "8081")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("POST /batches", handleSubmitBatch)
	mux.HandleFunc("GET /batches/{id}", handleBatchStatus)
	addr := ":" + port
	log.Printf("worker: listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("worker: %v", err)
	}
}

func handleSubmitBatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Fast int `json:"fast"`
		Slow int `json:"slow"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Fast <= 0 && req.Slow <= 0 {
		http.Error(w, "must specify fast or slow > 0", http.StatusBadRequest)
		return
	}
	id := strconv.FormatInt(batchSeq.Add(1), 10)
	b := &Batch{
		ID:        id,
		Fast:      req.Fast,
		Slow:      req.Slow,
		Total:     req.Fast + req.Slow,
		StartedAt: time.Now(),
	}
	batchesMu.Lock()
	batches[id] = b
	batchesMu.Unlock()
	// LAB: STEP3 TODO - Currently all jobs run in a single shared pool.
	// Slow jobs (simulated ~1s each) block fast jobs (simulated ~10ms each).
	// Participants should:
	//   1. Create separate goroutine pools for fast and slow jobs
	//   2. Cap slow concurrency (e.g., max 5 slow workers) so it cannot starve fast
	//   3. Keep fast pool large enough to process fast jobs without delay
	go processBatch(b)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"batch_id": id})
}

func processBatch(b *Batch) {
	// LAB: STEP3 FIXED - Bulkhead pattern: separate pools for fast and slow jobs.
	// Fast jobs get their own pool so slow jobs cannot starve them.
	fastPoolSize := 10 // Fast jobs get dedicated workers
	slowPoolSize := 5  // Slow jobs are capped to prevent resource hogging

	fastSem := make(chan struct{}, fastPoolSize)
	slowSem := make(chan struct{}, slowPoolSize)

	var wg sync.WaitGroup
	for i := 0; i < b.Fast; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fastSem <- struct{}{}
			defer func() { <-fastSem }()
			start := time.Now()
			time.Sleep(10 * time.Millisecond)
			b.recordResult("fast", time.Since(start))
		}()
	}
	for i := 0; i < b.Slow; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			slowSem <- struct{}{}
			defer func() { <-slowSem }()
			start := time.Now()
			time.Sleep(1 * time.Second)
			b.recordResult("slow", time.Since(start))
		}()
	}
	wg.Wait()
}

func (b *Batch) recordResult(jobType string, d time.Duration) {
	b.mu.Lock()
	b.Results = append(b.Results, JobResult{Type: jobType, Duration: d})
	b.mu.Unlock()
	b.Done.Add(1)
}

func handleBatchStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	batchesMu.RLock()
	b, ok := batches[id]
	batchesMu.RUnlock()
	if !ok {
		http.Error(w, "batch not found", http.StatusNotFound)
		return
	}
	done := int(b.Done.Load())
	b.mu.Lock()
	results := make([]JobResult, len(b.Results))
	copy(results, b.Results)
	b.mu.Unlock()
	fastP95 := percentile(results, "fast", 0.95)
	slowP95 := percentile(results, "slow", 0.95)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"batch_id":    b.ID,
		"total":       b.Total,
		"done":        done,
		"complete":    done >= b.Total,
		"fast_p95_ms": fastP95,
		"slow_p95_ms": slowP95,
		"elapsed_ms":  time.Since(b.StartedAt).Milliseconds(),
	})
}

func percentile(results []JobResult, jobType string, pct float64) float64 {
	var durations []float64
	for _, r := range results {
		if r.Type == jobType {
			durations = append(durations, float64(r.Duration.Milliseconds()))
		}
	}
	if len(durations) == 0 {
		return 0
	}
	sort.Float64s(durations)
	idx := int(math.Ceil(pct*float64(len(durations)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(durations) {
		idx = len(durations) - 1
	}
	return durations[idx]
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
