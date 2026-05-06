package notifier_test

import (
	"sync"
	"testing"

	"github.com/yourorg/grpcpulse/internal/notifier"
)

func TestState_String(t *testing.T) {
	cases := []struct {
		s    notifier.State
		want string
	}{
		{notifier.StateHealthy, "healthy"},
		{notifier.StateUnhealthy, "unhealthy"},
		{notifier.StateUnknown, "unknown"},
	}
	for _, c := range cases {
		if got := c.s.String(); got != c.want {
			t.Errorf("State(%d).String() = %q, want %q", c.s, got, c.want)
		}
	}
}

func TestObserve_TransitionFired(t *testing.T) {
	var mu sync.Mutex
	var calls []struct{ prev, cur notifier.State }

	n := notifier.New(func(target string, prev, cur notifier.State) {
		mu.Lock()
		calls = append(calls, struct{ prev, cur notifier.State }{prev, cur})
		mu.Unlock()
	}, nil)

	n.Observe("svc-a", true)
	n.Observe("svc-a", true) // same state — no alert
	n.Observe("svc-a", false)

	mu.Lock()
	defer mu.Unlock()

	if len(calls) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(calls))
	}
	if calls[0].prev != notifier.StateUnknown || calls[0].cur != notifier.StateHealthy {
		t.Errorf("first transition wrong: %+v", calls[0])
	}
	if calls[1].prev != notifier.StateHealthy || calls[1].cur != notifier.StateUnhealthy {
		t.Errorf("second transition wrong: %+v", calls[1])
	}
}

func TestObserve_NoAlertOnSameState(t *testing.T) {
	count := 0
	n := notifier.New(func(_ string, _, _ notifier.State) { count++ }, nil)

	for i := 0; i < 5; i++ {
		n.Observe("svc-b", false)
	}

	if count != 1 {
		t.Errorf("expected 1 alert for repeated unhealthy, got %d", count)
	}
}

func TestCurrent_ReturnsLastState(t *testing.T) {
	n := notifier.New(nil, nil)

	if got := n.Current("svc-c"); got != notifier.StateUnknown {
		t.Errorf("expected StateUnknown for unseen target, got %v", got)
	}

	n.Observe("svc-c", true)
	if got := n.Current("svc-c"); got != notifier.StateHealthy {
		t.Errorf("expected StateHealthy, got %v", got)
	}

	n.Observe("svc-c", false)
	if got := n.Current("svc-c"); got != notifier.StateUnhealthy {
		t.Errorf("expected StateUnhealthy, got %v", got)
	}
}

func TestObserve_MultipleTargets(t *testing.T) {
	n := notifier.New(nil, nil)
	n.Observe("alpha", true)
	n.Observe("beta", false)

	if n.Current("alpha") != notifier.StateHealthy {
		t.Error("alpha should be healthy")
	}
	if n.Current("beta") != notifier.StateUnhealthy {
		t.Error("beta should be unhealthy")
	}
}
