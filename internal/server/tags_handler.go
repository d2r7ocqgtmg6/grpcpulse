package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/yourorg/grpcpulse/internal/tags"
)

// tagsHandler exposes the tag registry over HTTP.
//
// GET  /tags          — returns all targets and their tags.
// GET  /tags?env=prod — filters targets by the provided query parameters.
func tagsHandler(reg *tags.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		selector := make(tags.Tags)
		for k, vs := range r.URL.Query() {
			if len(vs) > 0 {
				selector[k] = strings.TrimSpace(vs[0])
			}
		}

		targets := reg.Filter(selector)

		type entry struct {
			Target string     `json:"target"`
			Tags   tags.Tags  `json:"tags"`
		}

		result := make([]entry, 0, len(targets))
		for _, target := range targets {
			t, _ := reg.Get(target)
			result = append(result, entry{Target: target, Tags: t})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	}
}
