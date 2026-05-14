package acknowledger

import (
	"testing"
	"time"
)

func TestAcknowledge_IsAcknowledged(t *testing.T) {
	a := New()
	a.Acknowledge("svc:50051", "investigating", "alice", time.Minute)
	if !a.IsAcknowledged("svc:50051") {
		t.Fatal("expected target to be acknowledged")
	}
}

func TestIsAcknowledged_UnknownTarget(t *testing.T) {
	a := New()
	if a.IsAcknowledged("unknown:9000") {
		t.Fatal("expected unknown target to not be acknowledged")
	}
}

func TestIsAcknowledged_Expired(t *testing.T) {
	a := New()
	a.Acknowledge("svc:50051", "reason", "bob", -time.Second)
	if a.IsAcknowledged("svc:50051") {
		t.Fatal("expected expired ack to not be acknowledged")
	}
}

func TestClear_RemovesAck(t *testing.T) {
	a := New()
	a.Acknowledge("svc:50051", "reason", "alice", time.Minute)
	a.Clear("svc:50051")
	if a.IsAcknowledged("svc:50051") {
		t.Fatal("expected ack to be cleared")
	}
}

func TestActive_ReturnsOnlyNonExpired(t *testing.T) {
	a := New()
	a.Acknowledge("live:1", "ok", "alice", time.Minute)
	a.Acknowledge("dead:2", "expired", "bob", -time.Second)

	active := a.Active()
	if len(active) != 1 {
		t.Fatalf("expected 1 active ack, got %d", len(active))
	}
	if active[0].Target != "live:1" {
		t.Errorf("unexpected target %q", active[0].Target)
	}
}

func TestAcknowledgement_IsExpired(t *testing.T) {
	now := time.Now()
	ack := Acknowledgement{ExpiresAt: now.Add(-time.Second)}
	if !ack.IsExpired(now) {
		t.Error("expected ack to be expired")
	}
	ack2 := Acknowledgement{ExpiresAt: now.Add(time.Minute)}
	if ack2.IsExpired(now) {
		t.Error("expected ack to not be expired")
	}
}
