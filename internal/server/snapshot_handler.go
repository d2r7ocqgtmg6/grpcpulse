package server

import (
	"encoding/json"
	"net/http"
)

// SnapshotProvider can capture and return the latest snapshot.
type SnapshotProvider interface {
	Capture() interface{}
	Latest() interface{}
}

type snapshotCapturer interface {
	Capture() snapshotResult
	Latest() *snapshotResult
}

type snapshotResult interface{}

// snapshotRegistry is the minimal interface required by snapshotHandler.
type snapshotRegistry interface {
	CaptureRaw() interface{}
	LatestRaw() interface{}
}

// snapshotter is the interface the handler depends on.
type snapshotter interface {
	CaptureJSON() interface{}
	LatestJSON() interface{}
}

// snapshotSource is the concrete interface used by snapshotHandler.
type snapshotSource interface {
	Capture() interface{}
	Latest() interface{}
}

func snapshotHandler(src interface {
	Capture() interface{}
	Latest() interface{}
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var result interface{}
		if r.Method == http.MethodPost {
			result = src.Capture()
		} else {
			result = src.Latest()
			if result == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, "encode error", http.StatusInternalServerError)
		}
	}
}
