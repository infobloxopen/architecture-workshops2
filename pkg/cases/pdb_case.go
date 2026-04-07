package cases

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// PDBCase handles Case 5: PDB & rolling updates with CloudNativePG.
type PDBCase struct {
	DB *sql.DB
}

// Handle serves the /cases/pdb endpoint.
// Inserts a row and reads it back to verify write availability.
func (pc *PDBCase) Handle(w http.ResponseWriter, r *http.Request) {
	if pc.DB == nil {
		http.Error(w, "cnpg database not configured", http.StatusServiceUnavailable)
		return
	}

	start := time.Now()

	value := fmt.Sprintf("write-%d", start.UnixNano())
	var id int
	err := pc.DB.QueryRowContext(r.Context(),
		"INSERT INTO pdb_writes (value) VALUES ($1) RETURNING id", value,
	).Scan(&id)
	if err != nil {
		log.Printf("pdb: insert error: %v", err)
		http.Error(w, "insert failed: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	var readBack string
	err = pc.DB.QueryRowContext(r.Context(),
		"SELECT value FROM pdb_writes WHERE id = $1", id,
	).Scan(&readBack)
	if err != nil {
		log.Printf("pdb: read-back error: %v", err)
		http.Error(w, "read-back failed: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	elapsed := time.Since(start)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "ok",
		"id":         id,
		"value":      readBack,
		"elapsed_ms": elapsed.Milliseconds(),
	})
}
