package server

import (
	"encoding/json"
	"net/http"
	"time"
)

// oncallProvider is the interface required by oncallHandler.
type oncallProvider interface {
	CurrentResponder(target string, at time.Time) (string, bool)
	All() map[string]interface{}
}

// oncallHandler serves GET /oncall to query the current on-call responder
// for a given target, or list all rotations.
func oncallHandler(oc oncallProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		target := r.URL.Query().Get("target")
		if target == "" {
			// Return all rotations.
			all := oc.All()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(all)
			return
		}

		responder, ok := oc.CurrentResponder(target, time.Now())
		if !ok {
			http.Error(w, "no rotation found for target", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"target":    target,
			"responder": responder,
		})
	}
}
