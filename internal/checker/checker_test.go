package checker_test

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/yourorg/grpcpulse/internal/checker"
)

func startFakeGRPCServer(t *testing.T) (addr string, stop func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	go srv.Serve(lis) //nolint:errcheck
	return lis.Addr().String(), srv.Stop
}

func TestChecker_HealthyServer(t *testing.T) {
	addr, stop := startFakeGRPCServer(t)
	defer stop()

	c := checker.New(3 * time.Second)
	res := c.Check(context.Background(), addr)

	if res.Status != checker.StatusHealthy {
		t.Errorf("expected healthy, got %s (err: %v)", res.Status, res.Err)
	}
	if res.Latency <= 0 {
		t.Error("expected positive latency")
	}
	if res.Address != addr {
		t.Errorf("expected address %q, got %q", addr, res.Address)
	}
}

func TestChecker_UnreachableServer(t *testing.T) {
	c := checker.New(500 * time.Millisecond)
	res := c.Check(context.Background(), "127.0.0.1:19999")

	if res.Status != checker.StatusUnhealthy {
		t.Errorf("expected unhealthy, got %s", res.Status)
	}
	if res.Err == nil {
		t.Error("expected non-nil error for unreachable server")
	}
}

func TestChecker_DefaultTimeout(t *testing.T) {
	// A zero timeout should fall back to the default (5s).
	c := checker.New(0)
	if c == nil {
		t.Fatal("expected non-nil checker")
	}
}

func TestStatus_String(t *testing.T) {
	cases := []struct {
		status checker.Status
		want   string
	}{
		{checker.StatusHealthy, "healthy"},
		{checker.StatusUnhealthy, "unhealthy"},
		{checker.StatusUnknown, "unknown"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.want {
			t.Errorf("Status(%d).String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}
