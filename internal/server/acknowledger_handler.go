package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/grpcpulse/internal/acknowledger"
)

type acknowledgerHandler struct {
	ack *acknowledger.Acknowledger
}

func (h *acknowledgerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w)
	case http.MethodPost:
		h.add(w, r)
	case http.MethodDelete:
		h.remove(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *acknowledgerHandler) list(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.ack.Active())
}

func (h *acknowledgerHandler) add(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target  string `json:"target"`
		Reason  string `json:"reason"`
		AckedBy string `json:"acked_by"`
		TTL     string `json:"ttl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Target == "" {
		http.Error(w, "target is required", http.StatusBadRequest)
		return
	}
	ttl := 30 * time.Minute
	if req.TTL != "" {
		if d, err := time.ParseDuration(req.TTL); err == nil {
			ttl = d
		}
	}
	h.ack.Acknowledge(req.Target, req.Reason, req.AckedBy, ttl)
	w.WriteHeader(http.StatusCreated)
}

func (h *acknowledgerHandler) remove(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "target query param required", http.StatusBadRequest)
		return
	}
	h.ack.Clear(target)
	w.WriteHeader(http.StatusNoContent)
}
