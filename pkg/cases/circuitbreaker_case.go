package cases

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// CircuitBreakerCase handles Case 6: Circuit breaker for database calls.
type CircuitBreakerCase struct {
	DB       *sql.DB
	breaker  Breaker
	initOnce sync.Once
}

// NewCircuitBreakerCase creates a case with a disabled breaker (the broken state).
func NewCircuitBreakerCase(db *sql.DB) *CircuitBreakerCase {
	return &CircuitBreakerCase{
		DB: db,
		breaker: Breaker{
			// LAB: STEP6 TODO — The breaker is disabled (Threshold=0).
			// Every request will attempt a DB call even when the database is
			// down, blocking on connection timeouts for up to 30 seconds.
			// Fix: set Threshold to 5 and Timeout to 5 seconds.
			Threshold: 0,
			Timeout:   0,
		},
	}
}

func (cb *CircuitBreakerCase) ensureTable() {
	cb.initOnce.Do(func() {
		_, err := cb.DB.Exec(`CREATE TABLE IF NOT EXISTS cb_writes (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL,
			written_at TIMESTAMP DEFAULT NOW()
		)`)
		if err != nil {
			log.Printf("circuitbreaker: auto-create table error: %v", err)
		}
	})
}

// Handle serves the /cases/circuitbreaker endpoint.
func (cb *CircuitBreakerCase) Handle(w http.ResponseWriter, r *http.Request) {
	if cb.DB == nil {
		http.Error(w, "cnpg database not configured", http.StatusServiceUnavailable)
		return
	}

	cb.ensureTable()

	start := time.Now()

	// Check circuit breaker before attempting DB call.
	if !cb.breaker.Allow() {
		elapsed := time.Since(start)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "circuit_open",
			"breaker":    cb.breaker.State().String(),
			"failures":   cb.breaker.Failures(),
			"elapsed_ms": elapsed.Milliseconds(),
		})
		return
	}

	value := fmt.Sprintf("cb-write-%d", start.UnixNano())
	var id int
	err := cb.DB.QueryRowContext(r.Context(),
		"INSERT INTO cb_writes (value) VALUES ($1) RETURNING id", value,
	).Scan(&id)
	if err != nil {
		cb.breaker.RecordFailure()
		elapsed := time.Since(start)
		log.Printf("circuitbreaker: insert error (breaker=%s, failures=%d): %v",
			cb.breaker.State(), cb.breaker.Failures(), err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "error",
			"error":      err.Error(),
			"breaker":    cb.breaker.State().String(),
			"failures":   cb.breaker.Failures(),
			"elapsed_ms": elapsed.Milliseconds(),
		})
		return
	}

	cb.breaker.RecordSuccess()
	elapsed := time.Since(start)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "ok",
		"id":         id,
		"breaker":    cb.breaker.State().String(),
		"elapsed_ms": elapsed.Milliseconds(),
	})
}
