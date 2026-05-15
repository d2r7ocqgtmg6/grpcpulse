package server

import (
	"encoding/json"
	"net/http"
)

type dependencyRegistry interface {
	Add(from, to string) error
	Remove(from, to string)
	DependsOn(target string) []string
	Dependents(target string) []string
}

func dependencyHandler(reg dependencyRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "missing target query param", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			payload := map[string][]string{
				"depends_on": reg.DependsOn(target),
				"dependents": reg.Dependents(target),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(payload)

		case http.MethodPost:
			dep := r.URL.Query().Get("dep")
			if dep == "" {
				http.Error(w, "missing dep query param", http.StatusBadRequest)
				return
			}
			if err := reg.Add(target, dep); err != nil {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusCreated)

		case http.MethodDelete:
			dep := r.URL.Query().Get("dep")
			if dep == "" {
				http.Error(w, "missing dep query param", http.StatusBadRequest)
				return
			}
			reg.Remove(target, dep)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
