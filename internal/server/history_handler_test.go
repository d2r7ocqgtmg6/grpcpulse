package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/grpcpulse/internal/history"
	"github.com/yourorg/grpcpulse/internal/server"
)

func newServerWithHistory(t *testing.T, h *history.History) *httptest.Server {
	t.Helper()
	srv, err := server.New(server.Config{
		Addr:    "127.0.0.1:0",
		History: h,
	})
	if err != nil {
		t.Fatalf("server.New: %v", err)
	}
	return httptest.NewServer(srv.Handler())
}

func TestHistoryEndpoint_Empty(t *testing.T) {
	h := history.New(10)
	ts := newServerWithHistory(t, h)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/history")
	if err != nil {
		t.Fatalf("GET /history: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty array, got %d items", len(result))
	}
}

func TestHistoryEndpoint_WithData(t *testing.T) {
	h := history.New(10)
	h.Record("svc:50051", true, 4*time.Millisecond)
	h.Record("svc:50051", false, 8*time.Millisecond)

	ts := newServerWithHistory(t, h)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/history")
	if err != nil {
		t.Fatalf("GET /history: %v", err)
	}
	defer resp.Body.Close()

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 target, got %d", len(result))
	}
	entries, ok := result[0]["entries"].([]interface{})
	if !ok || len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %v", result[0]["entries"])
	}
	uptime, ok := result[0]["uptime_percent"].(float64)
	if !ok || uptime < 49 || uptime > 51 {
		t.Fatalf("expected ~50%% uptime, got %v", result[0]["uptime_percent"])
	}
}
