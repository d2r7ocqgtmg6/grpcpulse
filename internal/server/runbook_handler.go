package server

import (
	"encoding/json"
	"net/http"

	"github.com/yourorg/grpcpulse/internal/runbook"
)

// runbookHandler serves GET /runbooks and PUT/DELETE /runbooks?target=<addr>.
func runbookHandler(reg *runbook.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")

		switch r.Method {
		case http.MethodGet:
			if target == "" {
				writeJSON(w, http.StatusOK, reg.All())
				return
			}
			url, ok := reg.Get(target)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"target": target, "url": url})

		case http.MethodPut:
			if target == "" {
				http.Error(w, "target query param required", http.StatusBadRequest)
				return
			}
			var body struct {
				URL string `json:"url"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.URL == "" {
				http.Error(w, "invalid body: url required", http.StatusBadRequest)
				return
			}
			if err := reg.Set(target, body.URL); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if target == "" {
				http.Error(w, "target query param required", http.StatusBadRequest)
				return
			}
			reg.Delete(target)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
