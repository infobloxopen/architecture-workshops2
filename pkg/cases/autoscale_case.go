package cases

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AutoscaleCase handles Case 4: CPU-bound work for HPA testing.
type AutoscaleCase struct{}

// Handle serves the /cases/autoscale endpoint.
// Generates CPU load that should trigger HPA scaling.
func (ac *AutoscaleCase) Handle(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// CPU-intensive work: repeated SHA-256 hashing
	data := []byte("workshop-autoscale-seed")
	for i := 0; i < 100000; i++ {
		h := sha256.Sum256(data)
		data = h[:]
	}

	elapsed := time.Since(start)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "ok",
		"hash":       fmt.Sprintf("%x", data[:8]),
		"elapsed_ms": elapsed.Milliseconds(),
	})
}
