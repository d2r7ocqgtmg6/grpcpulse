package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/grpcpulse/internal/history"
)

// entryJSON is the JSON-serialisable form of a history.Entry.
type entryJSON struct {
	Timestamp time.Time     `json:"timestamp"`
	Healthy   bool          `json:"healthy"`
	LatencyMs float64       `json:"latency_ms"`
}

// historyResponse is the top-level response for /history.
type historyResponse struct {
	Target  string      `json:"target"`
	Uptime  float64     `json:"uptime_percent"`
	Entries []entryJSON `json:"entries"`
}

// historyHandler returns an http.HandlerFunc that serves recent check history
// for all known targets.
func historyHandler(h *history.History) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targets := h.Targets()
		results := make([]historyResponse, 0, len(targets))

		for _, t := range targets {
			entries := h.Get(t)
			jEntries := make([]entryJSON, len(entries))
			for i, e := range entries {
				jEntries[i] = entryJSON{
					Timestamp: e.Timestamp,
					Healthy:   e.Healthy,
					LatencyMs: float64(e.Latency.Microseconds()) / 1000.0,
				}
			}
			results = append(results, historyResponse{
				Target:  t,
				Uptime:  h.UptimePercent(t),
				Entries: jEntries,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(results); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
