package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/grpcpulse/internal/ratelimit"
)

type rateLimitStatus struct {
	Target   string `json:"target"`
	Allowed  bool   `json:"allowed"`
	ResetAt  string `json:"reset_at,omitempty"`
}

func rateLimitHandler(rl *ratelimit.RateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "missing target query parameter", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			allowed := rl.Allow(target)
			status := rateLimitStatus{
				Target:  target,
				Allowed: allowed,
			}
			if !allowed {
				status.ResetAt = time.Now().Add(rl.TTL()).UTC().Format(time.RFC3339)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(status) //nolint:errcheck

		case http.MethodDelete:
			rl.Reset(target)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
