package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/grpcpulse/internal/escalation"
)

type escalationObserver interface {
	Observe(target string, healthy bool, now time.Time) escalation.Level
	Current(target string) escalation.Level
	UnhealthySince(target string) (time.Time, error)
}

type escalationResponse struct {
	Target         string `json:"target"`
	Level          string `json:"level"`
	UnhealthySince string `json:"unhealthy_since,omitempty"`
}

// escalationHandler handles GET /escalation?target=<name>
func escalationHandler(reg escalationObserver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "missing target parameter", http.StatusBadRequest)
			return
		}

		level := reg.Current(target)
		resp := escalationResponse{
			Target: target,
			Level:  level.String(),
		}

		if since, err := reg.UnhealthySince(target); err == nil {
			resp.UnhealthySince = since.UTC().Format(time.RFC3339)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}
