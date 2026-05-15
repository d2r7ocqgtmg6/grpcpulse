package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/grpcpulse/internal/maintenance"
)

type maintenanceRegistry interface {
	Add(w maintenance.Window)
	IsUnderMaintenance(target string) bool
	Active() []maintenance.Window
	Remove(target string)
}

func maintenanceHandler(reg maintenanceRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")

		switch r.Method {
		case http.MethodGet:
			windows := reg.Active()
			type item struct {
				Target string    `json:"target"`
				Start  time.Time `json:"start"`
				End    time.Time `json:"end"`
				Reason string    `json:"reason"`
			}
			out := make([]item, 0, len(windows))
			for _, win := range windows {
				out = append(out, item{Target: win.Target, Start: win.Start, End: win.End, Reason: win.Reason})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(out) //nolint:errcheck

		case http.MethodPost:
			if target == "" {
				http.Error(w, "missing target", http.StatusBadRequest)
				return
			}
			var body struct {
				Start  time.Time `json:"start"`
				End    time.Time `json:"end"`
				Reason string    `json:"reason"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON body", http.StatusBadRequest)
				return
			}
			if body.End.IsZero() || !body.End.After(body.Start) {
				http.Error(w, "end must be after start", http.StatusBadRequest)
				return
			}
			reg.Add(maintenance.Window{
				Target: target,
				Start:  body.Start,
				End:    body.End,
				Reason: body.Reason,
			})
			w.WriteHeader(http.StatusCreated)

		case http.MethodDelete:
			if target == "" {
				http.Error(w, "missing target", http.StatusBadRequest)
				return
			}
			reg.Remove(target)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
