package runbook_test

import (
	"testing"

	"github.com/yourorg/grpcpulse/internal/runbook"
)

func TestSet_And_Get(t *testing.T) {
	r := runbook.New()
	if err := r.Set("svc:50051", "https://wiki.example.com/runbook-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := r.Get("svc:50051")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if v != "https://wiki.example.com/runbook-1" {
		t.Fatalf("unexpected URL: %s", v)
	}
}

func TestGet_Unknown(t *testing.T) {
	r := runbook.New()
	_, ok := r.Get("missing:9000")
	if ok {
		t.Fatal("expected no entry for unknown target")
	}
}

func TestSet_InvalidURL(t *testing.T) {
	r := runbook.New()
	for _, bad := range []string{"", "not-a-url", "/relative/path"} {
		if err := r.Set("svc:1", bad); err == nil {
			t.Fatalf("expected error for URL %q", bad)
		}
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	r := runbook.New()
	_ = r.Set("svc:50051", "https://wiki.example.com/rb")
	r.Delete("svc:50051")
	_, ok := r.Get("svc:50051")
	if ok {
		t.Fatal("entry should have been deleted")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	r := runbook.New()
	_ = r.Set("a:1", "https://example.com/a")
	_ = r.Set("b:2", "https://example.com/b")
	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating the copy must not affect the registry.
	delete(all, "a:1")
	if _, ok := r.Get("a:1"); !ok {
		t.Fatal("original registry should not be mutated")
	}
}
