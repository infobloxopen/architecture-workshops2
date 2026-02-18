package dep

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Run starts the dependency simulator service on :8082.
func Run() {
	port := envOr("DEP_PORT", "8082")
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/work", handleWork)
	addr := ":" + port
	log.Printf("dep: listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("dep: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func handleWork(w http.ResponseWriter, r *http.Request) {
	// Configurable sleep duration
	if s := r.URL.Query().Get("sleep"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			http.Error(w, "bad sleep param: "+err.Error(), http.StatusBadRequest)
			return
		}
		select {
		case <-time.After(d):
			// slept the full duration
		case <-r.Context().Done():
			http.Error(w, "cancelled", http.StatusServiceUnavailable)
			return
		}
	}
	// Configurable failure rate (0.0-1.0)
	if f := r.URL.Query().Get("fail"); f != "" {
		prob, err := strconv.ParseFloat(f, 64)
		if err != nil {
			http.Error(w, "bad fail param: "+err.Error(), http.StatusBadRequest)
			return
		}
		if rand.Float64() < prob {
			http.Error(w, "simulated failure", http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
