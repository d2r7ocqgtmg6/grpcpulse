package server

import (
	"encoding/json"
	"net/http"

	"github.com/yourorg/grpcpulse/internal/audit"
)

// auditHandler serves the audit log over HTTP.
// GET /audit          — returns all events
// GET /audit?target=x — returns events filtered by target
func auditHandler(log *audit.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var events []audit.Event
		if target := r.URL.Query().Get("target"); target != "" {
			events = log.FilterByTarget(target)
		} else {
			events = log.All()
		}

		if events == nil {
			events = []audit.Event{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(events); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
