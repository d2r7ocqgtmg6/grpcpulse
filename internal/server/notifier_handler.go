package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/grpcpulse/internal/notifier"
)

// notifierStateResponse represents the JSON response for a single target's alert state.
type notifierStateResponse struct {
	Target  string `json:"target"`
	State   string `json:"state"`
	Since   string `json:"since,omitempty"`
	Message string `json:"message,omitempty"`
}

// notifierHandler returns an HTTP handler that exposes the current alert/notification
// state for all tracked gRPC targets. Clients can poll this endpoint to inspect
// whether any targets have transitioned into a degraded or recovered state.
func notifierHandler(n *notifier.Notifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		targets := n.Targets()
		responses := make([]notifierStateResponse, 0, len(targets))

		for _, target := range targets {
			state := n.Current(target)
			resp := notifierStateResponse{
				Target: target,
				State:  state.String(),
			}
			responses = append(responses, resp)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(responses); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}

// notifierEventResponse represents a single state-transition event for JSON serialisation.
type notifierEventResponse struct {
	Target    string `json:"target"`
	FromState string `json:"from_state"`
	ToState   string `json:"to_state"`
	At        string `json:"at"`
}

// notifierEventsHandler returns an HTTP handler that streams recent state-transition
// events recorded by the notifier. This is useful for audit trails and alerting
// integrations that need to know when a target's health status changed.
func notifierEventsHandler(n *notifier.Notifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		events := n.Events()
		responses := make([]notifierEventResponse, 0, len(events))

		for _, e := range events {
			responses = append(responses, notifierEventResponse{
				Target:    e.Target,
				FromState: e.From.String(),
				ToState:   e.To.String(),
				At:        e.At.Format(time.RFC3339),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(responses); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}
