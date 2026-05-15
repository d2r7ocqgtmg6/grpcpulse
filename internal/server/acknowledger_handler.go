package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/grpcpulse/internal/acknowledger"
)

type acknowledgerStore interface {
	Acknowledge(target string, duration time.Duration) error
	IsAcknowledged(target string) bool
	Clear(target string)
	Active() []acknowledger.Acknowledgement
}

func acknowledgerHandler(store acknowledgerStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			acks := store.Active()
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(acks); err != nil {
				http.Error(w, "encode error", http.StatusInternalServerError)
			}

		case http.MethodPost:
			var req struct {
				Target   string `json:"target"`
				Duration string `json:"duration"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
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
			if err := store.Acknowledge(req.Target, dur); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)

		case http.MethodDelete:
			target := r.URL.Query().Get("target")
			if target == "" {
				http.Error(w, "target query param required", http.StatusBadRequest)
				return
			}
			store.Clear(target)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
