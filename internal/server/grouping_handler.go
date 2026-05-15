package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/example/grpcpulse/internal/grouping"
)

type groupRegistry interface {
	Add(group, target string)
	Remove(group, target string) error
	Members(group string) ([]string, error)
	Groups() []string
}

// groupingHandler serves /groups and /groups?group=<name> endpoints.
//
//	GET  /groups              — list all group names
//	GET  /groups?group=<name> — list members of a group
//	POST /groups?group=<name>&target=<addr> — add target to group
//	DELETE /groups?group=<name>&target=<addr> — remove target from group
func groupingHandler(reg groupRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		group := r.URL.Query().Get("group")
		target := r.URL.Query().Get("target")

		switch r.Method {
		case http.MethodGet:
			if group == "" {
				writeJSON(w, http.StatusOK, map[string][]string{"groups": reg.Groups()})
				return
			}
			members, err := reg.Members(group)
			if errors.Is(err, grouping.ErrGroupNotFound) {
				http.Error(w, "group not found", http.StatusNotFound)
				return
			}
			writeJSON(w, http.StatusOK, map[string][]string{"members": members})

		case http.MethodPost:
			if group == "" || target == "" {
				http.Error(w, "group and target are required", http.StatusBadRequest)
				return
			}
			reg.Add(group, target)
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if group == "" || target == "" {
				http.Error(w, "group and target are required", http.StatusBadRequest)
				return
			}
			if err := reg.Remove(group, target); errors.Is(err, grouping.ErrGroupNotFound) {
				http.Error(w, "group not found", http.StatusNotFound)
				return
			} else if errors.Is(err, grouping.ErrTargetNotInGroup) {
				http.Error(w, "target not in group", http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}
