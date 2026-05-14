package tags_test

import (
	"testing"

	"github.com/yourorg/grpcpulse/internal/tags"
)

func TestTags_String_Empty(t *testing.T) {
	var tg tags.Tags
	if got := tg.String(); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestTags_String_Deterministic(t *testing.T) {
	tg := tags.Tags{"env": "prod", "region": "eu-west"}
	expected := "env=prod,region=eu-west"
	if got := tg.String(); got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestRegistry_SetAndGet(t *testing.T) {
	r := tags.New()
	r.Set("host:50051", tags.Tags{"env": "staging"})

	got, ok := r.Get("host:50051")
	if !ok {
		t.Fatal("expected tags to be found")
	}
	if got["env"] != "staging" {
		t.Fatalf("expected env=staging, got %q", got["env"])
	}
}

func TestRegistry_Get_Unknown(t *testing.T) {
	r := tags.New()
	_, ok := r.Get("unknown:9090")
	if ok {
		t.Fatal("expected false for unknown target")
	}
}

func TestRegistry_Set_IsolatesMutation(t *testing.T) {
	r := tags.New()
	orig := tags.Tags{"env": "prod"}
	r.Set("svc:443", orig)
	orig["env"] = "mutated"

	got, _ := r.Get("svc:443")
	if got["env"] != "prod" {
		t.Fatalf("registry should store a copy; got %q", got["env"])
	}
}

func TestRegistry_Filter_MatchAll(t *testing.T) {
	r := tags.New()
	r.Set("a:1", tags.Tags{"env": "prod"})
	r.Set("b:2", tags.Tags{"env": "staging"})

	results := r.Filter(tags.Tags{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestRegistry_Filter_Selective(t *testing.T) {
	r := tags.New()
	r.Set("a:1", tags.Tags{"env": "prod", "region": "us"})
	r.Set("b:2", tags.Tags{"env": "prod", "region": "eu"})
	r.Set("c:3", tags.Tags{"env": "staging"})

	results := r.Filter(tags.Tags{"env": "prod", "region": "eu"})
	if len(results) != 1 || results[0] != "b:2" {
		t.Fatalf("expected [b:2], got %v", results)
	}
}

func TestRegistry_Delete(t *testing.T) {
	r := tags.New()
	r.Set("svc:80", tags.Tags{"env": "prod"})
	r.Delete("svc:80")

	_, ok := r.Get("svc:80")
	if ok {
		t.Fatal("expected target to be deleted")
	}
}
