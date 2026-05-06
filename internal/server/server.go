package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server exposes Prometheus metrics over HTTP.
type Server struct {
	httpServer *http.Server
}

// Config holds configuration for the metrics HTTP server.
type Config struct {
	Port            int
	ReadTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Port:            9090,
		ReadTimeout:     5 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
}

// New creates a new metrics HTTP server.
func New(cfg Config) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &Server{
		httpServer: &http.Server{
			Addr:        fmt.Sprintf(":%d", cfg.Port),
			Handler:     mux,
			ReadTimeout: cfg.ReadTimeout,
		},
	}
}

// Start begins listening and serving HTTP requests.
// It blocks until the server encounters an error.
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Addr returns the address the server is configured to listen on.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}
