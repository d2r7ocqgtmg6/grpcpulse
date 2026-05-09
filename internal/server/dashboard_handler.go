package server

import (
	"encoding/json"
	"net/http"

	"github.com/grpcpulse/internal/dashboard"
)

// dashboardHandler returns an http.HandlerFunc that serialises a fresh
// [dashboard.Snapshot] as JSON on every request.
func dashboardHandler(b *dashboard.Builder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := b.Build()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(snap); err != nil {
			http.Error(w, "failed to encode snapshot", http.StatusInternalServerError)
		}
	}
}
