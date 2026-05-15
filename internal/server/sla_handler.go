package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/grpcpulse/internal/sla"
)

type slaRegistry interface {
	Set(sla.Budget) error
	Delete(target string)
	Evaluate(target string) (sla.Status, error)
	All() []sla.Status
}

func slaHandler(reg slaRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")

		switch r.Method {
		case http.MethodGet:
			if target == "" {
				writeJSON(w, http.StatusOK, reg.All())
				return
			}
			s, err := reg.Evaluate(target)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			writeJSON(w, http.StatusOK, s)

		case http.MethodPut:
			var body struct {
				Target    string  `json:"target"`
				Objective float64 `json:"objective"`
				WindowSec int64   `json:"window_seconds"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if body.Target == "" {
				http.Error(w, "target required", http.StatusBadRequest)
				return
			}
			err := reg.Set(sla.Budget{
				Target:    body.Target,
				Objective: body.Objective,
				Window:    time.Duration(body.WindowSec) * time.Second,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if target == "" {
				http.Error(w, "target required", http.StatusBadRequest)
				return
			}
			reg.Delete(target)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
