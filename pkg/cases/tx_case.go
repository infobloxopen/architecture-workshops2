package cases

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/infobloxopen/architecture-workshops2/pkg/depclient"
)

// TxCase handles Case 2: DB transaction scope anti-pattern.
type TxCase struct {
	DB        *sql.DB
	DepClient *depclient.Client
}

// Handle serves the /cases/tx endpoint.
// BASELINE (broken): Holds a database transaction open while making
// a slow network call to the dep service. This causes connection pool
// exhaustion under load.
func (tc *TxCase) Handle(w http.ResponseWriter, r *http.Request) {
	if tc.DB == nil {
		http.Error(w, "database not configured", http.StatusServiceUnavailable)
		return
	}

	start := time.Now()

	// SOLUTION: Make the network call OUTSIDE the transaction first.
	_, depErr := depclient.Call(r.Context(), tc.DepClient, "2s", "0.0")
	if depErr != nil {
		log.Printf("tx: dep call error: %v", depErr)
		http.Error(w, "dep call failed: "+depErr.Error(), http.StatusBadGateway)
		return
	}

	// Now do a short transaction for just the DB work.
	tx, err := tc.DB.Begin()
	if err != nil {
		log.Printf("tx: begin error: %v", err)
		http.Error(w, "tx begin failed", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var balance int
	err = tx.QueryRow("SELECT balance FROM accounts WHERE name = $1 FOR UPDATE", "alice").Scan(&balance)
	if err != nil {
		log.Printf("tx: query error: %v", err)
		http.Error(w, "query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("UPDATE accounts SET balance = balance - 1, updated_at = NOW() WHERE name = $1", "alice")
	if err != nil {
		log.Printf("tx: update error: %v", err)
		http.Error(w, "update failed", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("tx: commit error: %v", err)
		http.Error(w, "commit failed", http.StatusInternalServerError)
		return
	}

	elapsed := time.Since(start)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "ok",
		"balance":    balance,
		"elapsed_ms": elapsed.Milliseconds(),
	})
}
