package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/infobloxopen/architecture-workshops2/pkg/cases"
	"github.com/infobloxopen/architecture-workshops2/pkg/depclient"
	_ "github.com/lib/pq"
)

// Server holds shared state for the API service.
type Server struct {
	DepClient *depclient.Client
	DB        *sql.DB
	Mux       *http.ServeMux
}

// Run starts the API service on :8080.
func Run() {
	port := envOr("API_PORT", "8080")
	depURL := envOr("DEP_URL", "http://dep:8082")
	srv := &Server{
		DepClient: depclient.NewClient(depURL),
		Mux:       http.NewServeMux(),
	}
	// Try to connect to postgres if DSN is provided
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("api: warning: could not open DB: %v", err)
		} else {
			db.SetMaxOpenConns(10)
			db.SetMaxIdleConns(5)
			srv.DB = db
		}
	}
	srv.Mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})
	srv.Mux.HandleFunc("/debug/dbstats", srv.handleDBStats)
	srv.RegisterCases()
	addr := ":" + port
	log.Printf("api: listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Mux); err != nil {
		log.Fatalf("api: %v", err)
	}
}

func (s *Server) handleDBStats(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		http.Error(w, "no database configured", http.StatusServiceUnavailable)
		return
	}
	stats := s.DB.Stats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"maxOpen":      stats.MaxOpenConnections,
		"open":         stats.OpenConnections,
		"inUse":        stats.InUse,
		"idle":         stats.Idle,
		"waitCount":    stats.WaitCount,
		"waitDuration": stats.WaitDuration.String(),
	})
}

// RegisterCases registers all lab case endpoints on the mux.
func (s *Server) RegisterCases() {
	tc := &cases.TimeoutCase{DepClient: s.DepClient}
	s.Mux.HandleFunc("/cases/timeouts", tc.Handle)
	txc := &cases.TxCase{DB: s.DB, DepClient: s.DepClient}
	s.Mux.HandleFunc("/cases/tx", txc.Handle)
	ac := &cases.AutoscaleCase{}
	s.Mux.HandleFunc("/cases/autoscale", ac.Handle)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
