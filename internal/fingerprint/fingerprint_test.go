package fingerprint_test

import (
	"testing"

	"github.com/yourorg/grpcpulse/internal/fingerprint"
)

func TestOf_Deterministic(t *testing.T) {
	a := fingerprint.Of("localhost:50051")
	b := fingerprint.Of("localhost:50051")
	if a != b {
		t.Fatalf("expected same fingerprint, got %q and %q", a, b)
	}
}

func TestOf_LabelOrderIndependent(t *testing.T) {
	a := fingerprint.Of("host:9090", "env=prod", "region=us")
	b := fingerprint.Of("host:9090", "region=us", "env=prod")
	if a != b {
		t.Fatalf("label order should not affect fingerprint: %q vs %q", a, b)
	}
}

func TestOf_DifferentAddresses(t *testing.T) {
	a := fingerprint.Of("host-a:50051")
	b := fingerprint.Of("host-b:50051")
	if a == b {
		t.Fatal("different addresses must produce different fingerprints")
	}
}

func TestOf_LabelsChangeFingerprint(t *testing.T) {
	base := fingerprint.Of("host:50051")
	withLabel := fingerprint.Of("host:50051", "env=prod")
	if base == withLabel {
		t.Fatal("adding a label must change the fingerprint")
	}
}

func TestOf_Length(t *testing.T) {
	fp := fingerprint.Of("localhost:50051")
	if len(fp.String()) != 16 {
		t.Fatalf("expected 16-char fingerprint, got %d: %q", len(fp.String()), fp)
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := fingerprint.New()
	fp := r.Register("localhost:50051", "env=staging")

	got, ok := r.Get("localhost:50051")
	if !ok {
		t.Fatal("expected fingerprint to be found")
	}
	if got != fp {
		t.Fatalf("expected %q, got %q", fp, got)
	}
}

func TestRegistry_Get_Unknown(t *testing.T) {
	r := fingerprint.New()
	_, ok := r.Get("unknown:9999")
	if ok {
		t.Fatal("expected not found for unregistered address")
	}
}

func TestRegistry_All_ReturnsCopy(t *testing.T) {
	r := fingerprint.New()
	r.Register("a:1")
	r.Register("b:2")

	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}

	// Mutating the returned map must not affect the registry.
	delete(all, "a:1")
	if _, ok := r.Get("a:1"); !ok {
		t.Fatal("registry must not be affected by mutation of All() result")
	}
}
