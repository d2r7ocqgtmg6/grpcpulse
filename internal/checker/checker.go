package checker

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// Status represents the health status of a gRPC endpoint.
type Status int

const (
	StatusUnknown Status = iota
	StatusHealthy
	StatusUnhealthy
)

func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// Result holds the outcome of a single health check.
type Result struct {
	Address  string
	Status   Status
	Latency  time.Duration
	Err      error
	CheckedAt time.Time
}

// Checker performs gRPC connectivity health checks.
type Checker struct {
	timeout time.Duration
}

// New creates a new Checker with the given dial timeout.
func New(timeout time.Duration) *Checker {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Checker{timeout: timeout}
}

// Check dials the given gRPC address and returns a Result.
func (c *Checker) Check(ctx context.Context, address string) Result {
	start := time.Now()
	result := Result{
		Address:   address,
		CheckedAt: start,
	}

	dialCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	conn, err := grpc.DialContext(
		dialCtx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	result.Latency = time.Since(start)

	if err != nil {
		result.Status = StatusUnhealthy
		result.Err = err
		return result
	}
	defer conn.Close()

	if conn.GetState() == connectivity.Ready {
		result.Status = StatusHealthy
	} else {
		result.Status = StatusUnhealthy
	}
	return result
}
