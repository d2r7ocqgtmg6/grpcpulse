package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/grpcpulse/internal/silencer"
)

type silenceRequest struct {
	Target   string `json:"target"`
	Reason   string `json:"reason"`
	Duration string `json:"duration"`
}

type silenceItem struct {
	Target    string `json:"target"`
	Reason    string `json:"reason"`
	ExpiresAt string `json:"expires_at"`
}

func silencerHandler(reg *silencer.Registry) http.Handler {
	mux := http.NewServeMux()

	// GET /silences — list active silences
	mux.HandleFunc("GET /silences", func(w http.ResponseWriter, r *http.Request) {
		active := reg.Active()
		items := make([]silenceItem, 0, len(active))
		for _, s := range active {
			items = append(items, silenceItem{
				Target:    s.Target,
				Reason:    s.Reason,
				ExpiresAt: s.ExpiresAt.UTC().Format(time.RFC3339),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items) //nolint:errcheck
	})

	// POST /silences — add a silence
	mux.HandleFunc("POST /silences", func(w http.ResponseWriter, r *http.Request) {
		var req silenceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if req.Target == "" {
			http.Error(w, "target is required", http.StatusBadRequest)
			return
		}
		dur, err := time.ParseDuration(req.Duration)
		if err != nil || dur <= 0 {
			http.Error(w, "invalid duration", http.StatusBadRequest)
			return
		}
		reg.Silence(req.Target, req.Reason, dur)
		w.WriteHeader(http.StatusCreated)
	})

	// DELETE /silences?target=... — lift a silence
	mux.HandleFunc("DELETE /silences", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "target query param required", http.StatusBadRequest)
			return
		}
		reg.Lift(target)
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}
